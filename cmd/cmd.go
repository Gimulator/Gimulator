package cmd

import (
	"flag"
	"os"
)

var (
	EpilogueType   = ""

	RabbitHost     = ""
	RabbitUsername = ""
	RabbitPassword = ""
	RabbitQueue    = ""
	ConfigDir      = ""
	Host           = ""
	Id             = ""
)

func ParseFlags() {
	flag.StringVar(&EpilogueType, "epilogue-type", "", "The epilogue component which Gimulator will write the result to it. Choices are: console, rabbitmq. Note: If you choose rabbitmq, you need to set the corresponding flags too.")

	flag.StringVar(&RabbitHost, "rabbit-url", "", "the host of rabbitMQ, Gimulator will use this address to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitUsername, "rabbit-username", "", "the username of rabbitMQ, Gimulator will use this username to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitPassword, "rabbit-password", "", "the password of rabbitMQ, Gimulator will use this password to connect to rabbitMQ for sending the result of the room")
	flag.StringVar(&RabbitQueue, "rabbit-result-queue", "", "the queue of rabbitMQ where Gimulator will use to send the result of room")
	flag.StringVar(&ConfigDir, "config-dir", "", "the direction of the Gimulator's configuration, this directory should contain two rules.yaml and credentials.yaml files")
	flag.StringVar(&Host, "host", "", "the host of Gimulator, where Gimulator listens on")
	flag.StringVar(&Id, "id", "", "the id of Gimulator, which distinguishes each gimulator instance from others")
	flag.Parse()

	if EpilogueType == "" {
		if EpilogueType = os.Getenv("GIMULATOR_EPILOGUE_TYPE"); EpilogueType == "" {
			EpilogueType = "console"
		}
	}

	if RabbitHost == "" {
		RabbitHost = os.Getenv("GIMULATOR_RABBIT_HOST")
	}
	if RabbitUsername == "" {
		RabbitUsername = os.Getenv("GIMULATOR_RABBIT_USERNAME")
	}
	if RabbitPassword == "" {
		RabbitPassword = os.Getenv("GIMULATOR_RABBIT_PASSWORD")
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
	if Id == "" {
		Id = os.Getenv("GIMULATOR_ID")
	}

	if ((EpilogueType == "rabbitmq") && (RabbitHost == "" || RabbitUsername == "" || RabbitPassword == "" || RabbitQueue == "")) || ConfigDir == "" || Host == "" || Id == "" {
		println("Please set the needed flags.")
		flag.PrintDefaults()
		os.Exit(1)
	}
}
