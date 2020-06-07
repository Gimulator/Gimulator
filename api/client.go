package api

import (
	"time"

	"github.com/Gimulator/Gimulator/object"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Send pings to peer with this period.
	pingPeriod = time.Second * 3

	channelBufSize = 128
)

type client struct {
	id      string
	ch      chan *object.Object
	token   string
	log     *logrus.Entry
	lastObj *object.Object
}

func NewClient(id string, token string) *client {
	return &client{
		id:      id,
		token:   token,
		ch:      make(chan *object.Object, channelBufSize),
		log:     logrus.WithField("entity", "client"),
		lastObj: nil,
	}
}

func (c *client) GetChan() chan *object.Object {
	return c.ch
}

func (c *client) GetToken() string {
	return c.token
}

func (c *client) Reconcile(conn *websocket.Conn) {
	log := c.log.WithField("client-id", c.id)
	log.Info("starting to reconcile connection")
	defer log.Info("end of reconciling connection")

	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()

	if c.lastObj != nil {
		if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
			log.WithError(err).Error("could not set write deadline for connection")
		}

		log.WithField("object", c.lastObj.String()).Debug("starting to write an object to the connection")
		if err := conn.WriteJSON(c.lastObj); err != nil {
			log.WithError(err).Error("could not write json to the connection")
			return
		}

		c.lastObj = nil
	}

	for {
		select {
		case obj, ok := <-c.ch:
			if !ok {
				log.Debug("the channel of objects is closed")
				if err := conn.WriteMessage(websocket.CloseMessage, []byte{}); err != nil {
					log.WithError(err).Error("could not write the close message to connection")
				}
				return
			}

			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.WithError(err).Error("could not set write deadline for connection")
			}

			log.WithField("object", obj.String()).Debug("starting to write an object to the connection")
			if err := conn.WriteJSON(obj); err != nil {
				log.WithError(err).Error("could not write json to the connection")
				c.lastObj = obj
				return
			}
			c.lastObj = nil
		case <-ticker.C:
			if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.WithError(err).Error("could not set write deadline for connection")
			}

			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.WithError(err).Error("could not write the ping message to the connection")
				return
			}
		}
	}
}
