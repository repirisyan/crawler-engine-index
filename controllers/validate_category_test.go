package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	Crawler "crawler-index/models/postgres/crawler"
)

func mlStub(t *testing.T, byTitle map[string]any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			ProductTitle string `json:"product_title"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		resp, ok := byTitle[body.ProductTitle]
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestPredictKeyword(t *testing.T) {
	srv := mlStub(t, map[string]any{
		"matte lipstick": map[string]any{"keyword_id": 42, "status": "ok"},
		"mystery object": map[string]any{"keyword_id": nil, "status": "unknown"},
		"zero id":        map[string]any{"keyword_id": 0, "status": "ok"},
		"not ok":         map[string]any{"keyword_id": 9, "status": "unknown"},
	})
	defer srv.Close()

	cases := []struct {
		title  string
		wantID uint64
		wantOK bool
	}{
		{"matte lipstick", 42, true},
		{"mystery object", 0, false},
		{"zero id", 0, false},
		{"not ok", 0, false},
		{"never seen", 0, false}, // stub 500s
	}
	for _, c := range cases {
		gotID, gotOK := predictKeyword(srv.Client(), srv.URL, c.title)
		if gotID != c.wantID || gotOK != c.wantOK {
			t.Errorf("predictKeyword(%q) = (%d, %v), want (%d, %v)", c.title, gotID, gotOK, c.wantID, c.wantOK)
		}
	}
}

func TestPredictPage(t *testing.T) {
	srv := mlStub(t, map[string]any{
		"keep":            map[string]any{"keyword_id": 5, "status": "ok"},  // == current, no change
		"move to known":   map[string]any{"keyword_id": 7, "status": "ok"},  // change, keyword exists
		"move to unknown": map[string]any{"keyword_id": 99, "status": "ok"}, // change, keyword missing -> skip
	})
	defer srv.Close()

	rows := []Crawler.CategoryCandidate{
		{ID: 1, Title: "keep", KeywordID: 5},
		{ID: 2, Title: "move to known", KeywordID: 5},
		{ID: 3, Title: "move to unknown", KeywordID: 5},
		{ID: 4, Title: "never seen", KeywordID: 5}, // stub 500 -> keep
	}
	kwComodity := map[uint64]uint64{5: 50, 7: 70}

	updates := predictPage(srv.Client(), srv.URL, rows, kwComodity, 4)

	if len(updates) != 1 {
		t.Fatalf("expected 1 update, got %d: %+v", len(updates), updates)
	}
	u := updates[0]
	if u.ID != 2 || u.KeywordID != 7 || u.ComodityID != 70 {
		t.Errorf("unexpected update: %+v", u)
	}
}
