package main

import (
	"WorkoutTracker/cmd/internal/storage/postgresql"

	_ "github.com/lib/pq"
)

func main() {

	storage, err := postgresql.New()
	if err != nil {
		panic("failed to initialize storage: " + err.Error())
	}

	_ = storage

	// TODO: init logger

	// TODO: init storage

	// TODO: init router

	// TODO: run server
}
