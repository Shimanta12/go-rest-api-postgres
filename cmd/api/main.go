package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shimanta12/go-rest-api-postgres/internal/handler"
	"github.com/Shimanta12/go-rest-api-postgres/internal/middleware"
	"github.com/Shimanta12/go-rest-api-postgres/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
)

func main(){
	godotenv.Load()

	db, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil{
		log.Fatalf("failed to open database: %v\n", err)
	}
	defer db.Close()
	
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)
	
	if err := db.Ping(); err != nil{
		log.Fatalf("failed to ping database: %v\n", err)
	}

	fmt.Println("Database Connected")

	// dependencies
	userStore := store.NewUserStore(db)
	userHandler := handler.NewUserHandler(userStore)

	
	port := os.Getenv("PORT")
	if port == ""{
		port = "8080"
	}
	
	mux := http.NewServeMux()
	userHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr: ":" + port,
		Handler: middleware.LoggerMiddleware(mux),
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout: 2 * time.Minute,
	}

	go func(){
		fmt.Printf("server running on http://localhost:%v\n", port)
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed{
			log.Fatalf("server error: %v\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	fmt.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second * 10)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil{
		log.Fatalf("shutdown failed: %v", err)
	}
	log.Println("server stopped cleanly")
}