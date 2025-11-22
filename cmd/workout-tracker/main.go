package main

import (
	"WorkoutTracker/internal/storage/postgresql"

	_ "github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

func main() {

	storage, err := postgresql.New()
	if err != nil {
		panic("failed to initialize storage: " + err.Error())
	}

	_ = storage

	// TODO: init logger

	// TODO: init router

	// TODO: run server
}
