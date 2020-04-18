package main

import (
	"os"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/storage"
)

func main() {
	if len(os.Args) < 1 {
		panic("Usage: " + os.Args[0] + " path-to-rules.yaml")
	}
	path := os.Args[1]

	storage := storage.NewMemory()
	simulator := simulator.NewSimulator(storage)
	auth, err := auth.NewAuth(path)
	if err != nil {
		panic(err.Error())
	}

	api := api.NewManager(simulator, auth)
	api.ListenAndServe("localhost:5000")
}
