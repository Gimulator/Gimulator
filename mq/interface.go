package mq

import "github.com/Gimulator/protobuf/go/api"

type MessageQueue interface {
	Send(result *api.Result) error
}
