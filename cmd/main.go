package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/prospera/internals/configs"
	"github.com/prospera/internals/routers"
)

func main() {
	// Load ENV
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Init Database
	db, err := configs.InitDB()
	if err != nil {
		log.Printf("DB ERROR: %s", err.Error())
	}

	// Ping Database
	if err := configs.PingDB(db); err != nil {
		log.Printf("DB ERROR: %s", err.Error())
	}
	log.Println("database connected")

	//Init Redis
	rdb := configs.InitRedis()

	router := routers.InitRouter(db, rdb)
	router.Run(":8080")
}
