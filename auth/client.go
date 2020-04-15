package auth

import (
	"fmt"

	"github.com/gorilla/websocket"
	"gitlab.com/Syfract/Xerac/gimulator/object"
)

type Credential struct {
	Username string
	Password string
	Role     string
}

type Client struct {
	cred  Credential
	ch    chan object.Object
	conn  *websocket.Conn
	token string
}

func NewClient(cred Credential, token string) *Client {
	return &Client{
		cred:  cred,
		token: token,
		ch:    make(chan object.Object, 128),
	}
}

func (c *Client) GetChan() chan object.Object {
	return c.ch
}

func (c *Client) GetToken() string {
	return c.token
}

func (c *Client) SetConn(conn *websocket.Conn) {
	c.conn = conn
	go c.write()
}

func (c *Client) write() {
	var (
		err error
		obj object.Object
	)

	for {
		obj = <-c.ch
		err = c.conn.WriteMessage(websocket.PingMessage, []byte{})
		if err != nil {
			fmt.Println("client-write ", err)
			return
		}

		err = c.conn.WriteJSON(obj)
		if err != nil {
			fmt.Println("client-write ", err)
			return
		}
	}
}
