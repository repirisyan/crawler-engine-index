package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	Crawler "crawler-index/models/postgres/crawler"
	Keyword "crawler-index/models/postgres/keyword"

	"github.com/joho/godotenv"
)

// ValidateCategory re-predicts each freshly-crawled product's category with the
// ML service (crawler-ml-category) and, when the model disagrees with the search
// keyword it was crawled under, rewrites keyword_id / comodity_id on the crawler
// row. Downstream steps (indexing, the summaries) then see the corrected values.
//
// Opt-in: does nothing unless ML_CATEGORY_HOST is set (full URL to /predict).
// NOTE: this feeds a loop - corrected rows become the next model's training data
// (training/training.py reads `products`). Only enable with a model you trust.
func ValidateCategory() {
	_ = godotenv.Load()

	host := os.Getenv("ML_CATEGORY_HOST")
	if host == "" {
		log.Println("ValidateCategory skipped: ML_CATEGORY_HOST not set")
		return
	}

	windowHours := envInt("ML_CATEGORY_WINDOW_HOURS", 24)
	concurrency := envInt("ML_CATEGORY_CONCURRENCY", 8)
	timeout := time.Duration(envInt("ML_CATEGORY_TIMEOUT_MS", 5000)) * time.Millisecond

	kwComodity, err := Keyword.GetKeywordComodityMap()
	if err != nil {
		log.Printf("ValidateCategory aborted: cannot load keyword map: %v", err)
		return
	}

	client := &http.Client{Timeout: timeout}
	const pageSize = 1000
	offset := 0
	checked, updated := 0, 0

	for {
		rows, err := Crawler.GetProductsForCategoryValidation(windowHours, offset, pageSize)
		if err != nil {
			log.Printf("ValidateCategory: fetch page failed: %v", err)
			break
		}
		if len(rows) == 0 {
			break
		}

		updates := predictPage(client, host, rows, kwComodity, concurrency)
		if err := Crawler.UpdateProductCategories(updates); err != nil {
			log.Printf("ValidateCategory: batch update failed: %v", err)
		} else {
			updated += len(updates)
		}

		checked += len(rows)
		if len(rows) < pageSize {
			break
		}
		offset += pageSize
	}

	log.Printf("ValidateCategory: checked %d, corrected %d", checked, updated)
}

// predictPage fans the ML calls for one page out across `concurrency` workers and
// returns the category corrections to apply.
func predictPage(
	client *http.Client,
	host string,
	rows []Crawler.CategoryCandidate,
	kwComodity map[uint64]uint64,
	concurrency int,
) []Crawler.CategoryUpdate {
	if concurrency < 1 {
		concurrency = 1
	}

	jobs := make(chan Crawler.CategoryCandidate)
	results := make(chan Crawler.CategoryUpdate)
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for row := range jobs {
				predicted, ok := predictKeyword(client, host, row.Title)
				if !ok || predicted == 0 || predicted == row.KeywordID {
					continue
				}
				comodityID, known := kwComodity[predicted]
				if !known {
					continue // model returned an id that is not a real keyword
				}
				results <- Crawler.CategoryUpdate{
					ID:         row.ID,
					KeywordID:  predicted,
					ComodityID: comodityID,
				}
			}
		}()
	}

	go func() {
		for _, row := range rows {
			jobs <- row
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	var updates []Crawler.CategoryUpdate
	for u := range results {
		updates = append(updates, u)
	}
	return updates
}

type predictResponse struct {
	KeywordID *int64 `json:"keyword_id"`
	Status    string `json:"status"`
}

// predictKeyword POSTs one title to the ML service. Returns (0, false) on any
// error or an inconclusive result - the caller then keeps the original category.
func predictKeyword(client *http.Client, host, title string) (uint64, bool) {
	body, err := json.Marshal(map[string]string{"product_title": title})
	if err != nil {
		return 0, false
	}

	resp, err := client.Post(host, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, false
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil || resp.StatusCode != http.StatusOK {
		return 0, false
	}

	var parsed predictResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return 0, false
	}
	if parsed.Status != "ok" || parsed.KeywordID == nil || *parsed.KeywordID <= 0 {
		return 0, false
	}
	return uint64(*parsed.KeywordID), true
}

func envInt(name string, def int) int {
	if v := os.Getenv(name); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
