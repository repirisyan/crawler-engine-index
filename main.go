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
	controllers.StoreSupervisionData()
	controllers.StoreIndexData()
	controllers.StoreSellerDistribution()
	controllers.StoreBrandLeaderboard()
	controllers.StoreDiscountProduct()
	redisdb.FlushAllDB()
	fmt.Printf("Cleaning Complete")
}
