package main

import (
	"crawler-index/controllers"
	"crawler-index/db/mysql"
	"fmt"
)

func main() {
	mysql.Init()
	defer mysql.DB.Close()
	controllers.RemoveDuplicationData("temp_items")
	controllers.SetCertified()
	controllers.ValidateCategory()
	controllers.StoreTrainingData()
	controllers.RemoveDuplicationData("training_data")
	controllers.StoreSupervision()
	controllers.RemoveDuplicationData("supervisions")
	controllers.StoreIndexData()
	controllers.RemoveDuplicationData("products")
	fmt.Printf("Cleaning Complete")
}
