package epilogues

import (
	"encoding/json"
	"fmt"

	"github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type RabbitMQ struct {
	uri   string
	queue string
	log   *logrus.Entry
	ch    *amqp.Channel
}

func NewRabbitMQ(host, username, password, queueName string) (*RabbitMQ, error) {
	uri := fmt.Sprintf("amqps://%v:%v@%v:5671", username, password, host)
	r := &RabbitMQ{
		uri:   uri,
		queue: queueName,
		log:   logrus.WithField("component", "rabbitmq"),
	}

	conn, err := amqp.Dial(r.uri)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	r.ch = ch

	return r, nil
}

func (r *RabbitMQ) Test() error {
	return nil  // TODO
}

func (r *RabbitMQ) Write(result *api.Result) error {
	r.log.Info("starting to send result")

	r.log.Info("starting to marshal result")
	data, err := json.Marshal(result)
	if err != nil {
		r.log.WithError(err).Error("could not marshal result")
		return err
	}
	body := string(data)

	r.log.Info("starting to declare queue")
	queue, err := r.ch.QueueDeclare(
		r.queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		r.log.WithError(err).Error("could not declare queue")
		return err
	}

	r.log.Info("starting to publish message")
	if err := r.ch.Publish(
		"",         // exchange
		queue.Name, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "application/x-yaml",
			Body:        []byte(body),
		},
	); err != nil {
		r.log.WithError(err).Error("could not publish message")
		return err
	}

	return nil
}
