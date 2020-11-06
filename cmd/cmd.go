package cmd

import (
	"flag"
	"os"
)

var (
	RabbitHost     = ""
	RabbitUsername = ""
	RabbitPassword = ""
	RabbitQueue    = ""
	ConfigDir      = ""
	Host           = ""
)

func ParseFlags() {

	flag.StringVar(&RabbitHost, "rabbit-url", "", "the host of rabbitMQ, Gimulator will use this address to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitUsername, "rabbit-username", "", "the username of rabbitMQ, Gimulator will use this username to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitPassword, "rabbit-password", "", "the password of rabbitMQ, Gimulator will use this password to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitQueue, "rabbit-result-queue", "", "the queue of rabbitMQ where Gimulator will use to send the result of room")
	flag.StringVar(&ConfigDir, "config-dir", "", "the direction of the Gimulator's configuration, this directory should contain two rules.yaml and credentials.yaml files")
	flag.StringVar(&Host, "host", "", "the host of Gimulator, where Gimulator listens on")
	flag.Parse()

	if RabbitHost == "" {
		RabbitHost = os.Getenv("GIMULATOR_RABBIT_URL")
	}
	if RabbitUsername == "" {
		RabbitUsername = os.Getenv("GIMULATOR_RABBIT_URL")
	}
	if RabbitPassword == "" {
		RabbitPassword = os.Getenv("GIMULATOR_RABBIT_URL")
	}
	if RabbitQueue == "" {
		RabbitQueue = os.Getenv("GIMULATOR_RABBIT_RESULT_QUEUE")
	}
	if ConfigDir == "" {
		ConfigDir = os.Getenv("GIMULATOR_CONFIG_DIR")
	}
	if Host == "" {
		Host = os.Getenv("GIMULATOR_HOST")
	}

	if RabbitHost == "" || RabbitUsername == "" || RabbitPassword == "" || RabbitQueue == "" || ConfigDir == "" || Host == "" {
		println("please set the needed flags")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
