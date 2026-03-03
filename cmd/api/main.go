package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Shimanta12/go-rest-api-postgres/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main(){
	if err := godotenv.Load(); err != nil{
		log.Fatalln("failed to load .env")
	}
	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil{
		log.Fatalf("failed to open database: %v\n", err)
	}
	defer db.Close()
	
	if err := db.Ping(); err != nil{
		log.Fatalf("failed to ping database: %v\n", err)
	}
	fmt.Println("Database Connected")

	userStore := store.CreateUserStore(db)
	fmt.Println(userStore)
}