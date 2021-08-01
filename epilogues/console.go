package epilogues

import (
	"encoding/json"

	"github.com/Gimulator/protobuf/go/api"
	"github.com/sirupsen/logrus"
)

type Console struct {
	log   *logrus.Entry
}

func NewConsole() (*Console, error) {
	c := &Console{
		log: logrus.WithField("component", "console"),
	}

	return c, nil
}

func (c *Console) Test() error {
	return nil  // TODO
}

func (c *Console) Write(result *api.Result) error {
	c.log.Info("starting to send result")

	c.log.Info("starting to marshal result")
	data, err := json.Marshal(result)
	if err != nil {
		c.log.WithError(err).Error("could not marshal result")
		return err
	}
	s := string(data)

	c.log.Info("starting to print message")
	c.log.Info(s)

	return nil
}
