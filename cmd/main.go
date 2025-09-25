package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/prospera/internals/configs"
	"github.com/prospera/internals/routers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := configs.InitDB()
	if err != nil {
		log.Printf("DB ERROR: %s", err.Error())
	}

	if err := configs.PingDB(db); err != nil {
		log.Printf("DB ERROR: %s", err.Error())
	}
	log.Println("pg connected")

	router := routers.InitRouter(db)
	router.Run(":8080")
}
