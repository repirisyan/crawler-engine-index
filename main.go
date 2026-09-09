package main

import (
	"crawler-index/controllers"
	pgdb "crawler-index/db/postgres"
	redisdb "crawler-index/db/redis"
	"fmt"
	"log"
)

// safeRun executes a cleaning step, recovering from any panic so that a single
// failing step does not abort the whole pipeline (in particular the final
// Redis flush).
func safeRun(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️ step %q panicked, continuing: %v", name, r)
		}
	}()
	log.Printf("▶️ %s", name)
	fn()
}

func main() {
	redisdb.InitRedis()
	pgdb.InitPostgres()

	steps := []struct {
		name string
		fn   func()
	}{
		{"RemoveDuplicationFromRedis", controllers.RemoveDuplicationFromRedis},
		{"SetCertified", controllers.SetCertified},
		// ValidateCategory is opt-in: no-op unless ML_CATEGORY_HOST is set.
		{"ValidateCategory", controllers.ValidateCategory},
		{"StoreSupervisionData", controllers.StoreSupervisionData},
		{"StoreIndexData", controllers.StoreIndexData},
		{"StoreSellerDistribution", controllers.StoreSellerDistribution},
		{"StoreBrandLeaderboard", controllers.StoreBrandLeaderboard},
		{"StoreDiscountProduct", controllers.StoreDiscountProduct},
		{"StoreRegionBrand", controllers.StoreRegionBrand},
	}

	for _, s := range steps {
		safeRun(s.name, s.fn)
	}

	safeRun("FlushAllDB", redisdb.FlushAllDB)
	fmt.Println("Cleaning Complete")
}
