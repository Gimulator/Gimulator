package mq

import (
	"encoding/json"

	"github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
)

type Rabbit struct {
	url       string
	queueName string
	log       *logrus.Entry
}

func NewRabbit(url string, queueName string) (*Rabbit, error) {
	r := &Rabbit{
		url:       url,
		queueName: queueName,
		log:       logrus.WithField("component", "rabbit"),
	}

	if err := r.test(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Rabbit) test() error {
	conn, err := amqp.Dial(r.url)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	return nil
}

func (r *Rabbit) Send(result *api.Result) error {
	r.log.Info("starting to send result")

	r.log.Info("starting to marshal result")
	data, err := json.Marshal(result)
	if err != nil {
		r.log.WithError(err).Error("could not marshal result")
		return err
	}
	body := string(data)

	r.log.Info("starting to open a new connection")
	conn, err := amqp.Dial(r.url)
	if err != nil {
		r.log.WithError(err).Error("could not open new connection")
		return err
	}
	defer conn.Close()

	r.log.Info("starting to create a new channel from connection")
	ch, err := conn.Channel()
	if err != nil {
		r.log.WithError(err).Error("could not create new channel from connection")
		return err
	}
	defer ch.Close()

	r.log.Info("starting to declare queue")
	queue, err := ch.QueueDeclare(
		r.queueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		r.log.WithError(err).Error("could not declare queue")
		return err
	}

	r.log.Info("starting to publish message")
	if err := ch.Publish(
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
