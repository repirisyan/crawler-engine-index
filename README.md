# Crawler Indexing App (crawler-cleaning-app)

Drains the crawl results from Redis, cleans/dedupes them into PostgreSQL, and
builds the derived tables (supervisions, product index, the summary boards).
Invoked by `crawler-engine-cleaning` as the Go binary `goapp/cleaning-app`.

## Requirements

- Go 1.25+
- PostgreSQL, Redis

## Run

```sh
cp .env.example .env   # then edit
go run main.go          # dev
go build -o cleaning-app && ./cleaning-app   # prod (place in ../goapp/)
go test ./...
```

## Pipeline (`main.go`)

Each step is wrapped so one failure can't abort the rest:

| Step | Does |
| --- | --- |
| `RemoveDuplicationFromRedis` | drain Redis set `products:crawler` → `crawlers` (dedupe via `ON CONFLICT`) |
| `SetCertified` | detect BPOM / SNI / halal / distribution-permit from the description |
| `ValidateCategory` | **opt-in** — re-predict each new product's category via crawler-ml-category and fix `keyword_id` / `comodity_id`; no-op unless `ML_CATEGORY_HOST` is set |
| `StoreSupervisionData` | flag products matching `supervision_lists`, copy them to `supervisions` |
| `StoreIndexData` | copy `crawlers` → `products` |
| `StoreSellerDistribution` / `StoreBrandLeaderboard` / `StoreDiscountProduct` / `StoreRegionBrand` | rebuild the summary tables for the current (year, month) |
| `FlushAllDB` | `FLUSHALL` Redis |

### ValidateCategory

Pages `crawlers` rows updated in the last `ML_CATEGORY_WINDOW_HOURS` (i.e. this
run's inserts), POSTs each title to `ML_CATEGORY_HOST` with
`ML_CATEGORY_CONCURRENCY` workers, and when the model returns a different, valid
`keyword_id` it updates the row (and the matching `comodity_id` from `keywords`).

> This is a feedback loop: corrected rows land in `products`, which is what
> `crawler-ml-category/training/training.py` trains the next model on. Only enable
> it with a model you trust.
