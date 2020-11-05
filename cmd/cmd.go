package cmd

import (
	"flag"
	"os"
)

var (
	RabbitURL         = ""
	RabbitResultQueue = ""
	ConfigDir         = ""
	Host              = ""
)

func ParseFlags() {

	flag.StringVar(&RabbitURL, "rabbit-url", "", "the url of rabbitMQ, Gimulator will use this url to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitResultQueue, "rabbit-result-queue", "", "the queue of rabbitMQ where Gimulator will use to send the result of room")
	flag.StringVar(&ConfigDir, "config-dir", "", "the direction of the Gimulator's configuration, this directory should contain two rules.yaml and credentials.yaml files")
	flag.StringVar(&Host, "host", "", "the host of Gimulator, where Gimulator listens on")
	flag.Parse()

	if RabbitURL == "" {
		RabbitURL = os.Getenv("GIMULATOR_RABBIT_URL")
	}
	if RabbitResultQueue == "" {
		RabbitResultQueue = os.Getenv("GIMULATOR_RABBIT_RESULT_QUEUE")
	}
	if ConfigDir == "" {
		ConfigDir = os.Getenv("GIMULATOR_CONFIG_DIR")
	}
	if Host == "" {
		Host = os.Getenv("GIMULATOR_HOST")
	}

	if RabbitURL == "" || RabbitResultQueue == "" || ConfigDir == "" || Host == "" {
		println("please set the needed flags")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
