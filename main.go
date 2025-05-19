package main

import (
	"crawler-index/controllers"
	pgdb "crawler-index/db/postgres"
	redisdb "crawler-index/db/redis"
	"fmt"
)

func main() {
	redisdb.InitRedis()
	pgdb.InitPostgres()
	controllers.RemoveDuplicationFromRedis()
	controllers.SetCertified()
	// controllers.ValidateCategory()
	// controllers.StoreTrainingData()
	// controllers.CleanTrainingData()
	controllers.StoreIndexData()
	controllers.StoreSellerDistribution()
	controllers.StoreBrandLeaderboard()
	controllers.StoreDiscountProduct()
	fmt.Printf("Cleaning Complete")
}
