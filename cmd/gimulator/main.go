package main

import (
	"flag"

	"github.com/Gimulator/Gimulator/api"
	"github.com/Gimulator/Gimulator/auth"
	"github.com/Gimulator/Gimulator/simulator"
	"github.com/Gimulator/Gimulator/storage"
)

func main() {
	ip := flag.String("ip", "localhost:3030", "ip is for listening and serving")
	configFile := flag.String("config-file", "", "this is a config file for auth")
	flag.Parse()

	if *configFile == "" {
		flag.PrintDefaults()
		return
	}

	storage := storage.NewMemory()
	simulator := simulator.NewSimulator(storage)
	auth, err := auth.NewAuth(*configFile)
	if err != nil {
		panic(err.Error())
	}

	api := api.NewManager(simulator, auth)
	api.ListenAndServe(*ip)
}
