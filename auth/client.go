package auth

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"gitlab.com/Syfract/Xerac/gimulator/object"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 5 * time.Second

	// Send pings to peer with this period.
	pingPeriod = time.Second * 3
)

type Credential struct {
	Username string
	Password string
	Role     string
}

type Client struct {
	cred  Credential
	ch    chan object.Object
	token string
	log   *logrus.Entry
}

func NewClient(cred Credential, token string) *Client {
	return &Client{
		cred:  cred,
		token: token,
		ch:    make(chan object.Object, 128),
		log:   logrus.WithField("Entity", "client"),
	}
}

func (c *Client) GetChan() chan object.Object {
	return c.ch
}

func (c *Client) GetToken() string {
	return c.token
}

func (c *Client) Reconcile(conn *websocket.Conn) {
	c.log.Info("Start to write")
	defer c.log.Debug("End of writing to the connection")

	var err error
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		conn.Close()
	}()
	for {
		select {
		case obj, ok := <-c.ch:
			if !ok {
				conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			conn.SetWriteDeadline(time.Now().Add(writeWait))
			err = conn.WriteJSON(obj)
			if err != nil {
				c.log.WithError(err).Error("Can not write json to connection")
				fmt.Println("client-write ", err)
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
