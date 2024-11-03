package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/websocket"
)

type Event struct {
	Kind    int
	Message []byte
}

func (e Event) String() string {
	return fmt.Sprintf("{ Kind: %v, Message: %v }", e.Kind, string(e.Message))
}

type Client struct {
	conn   *websocket.Conn
	recvCh chan Event
	sendCh chan Event
}

func NewClient(url string) (Client, error) {
	log.Println("Connection to url:", url)
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return Client{}, err
	}
	if resp != nil && resp.StatusCode == 101 {
		log.Println("Websocket connection established with", conn.RemoteAddr())
	}
	return Client{
		conn:   conn,
		recvCh: make(chan Event),
		sendCh: make(chan Event),
	}, nil
}

func (c *Client) Close() {
	close(c.recvCh)
	close(c.sendCh)
	c.conn.Close()
}

func (c *Client) HandleUserInput() {
	sc := bufio.NewScanner(os.Stdin)

	for {
		if sc.Scan() {
			c.sendCh <- Event{
				Kind:    websocket.TextMessage,
				Message: sc.Bytes(),
			}
		}
		if err := sc.Err(); err != nil {
			log.Println("Scan error: ", err)
			return
		}
	}
}

func (c *Client) Run() {
	for {
		select {
		case e := <-c.recvCh:
			fmt.Println(e)
		case e := <-c.sendCh:
			c.conn.WriteMessage(e.Kind, e.Message)
		}
	}
}
func (c *Client) HandleServerEvents() {
	for {
		kind, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("WS error: ", err)
			return
		}
		c.recvCh <- Event{
			Kind:    kind,
			Message: message,
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Websocket URL required")
	}
	url := os.Args[1]

	client, err := NewClient(url)
	if err != nil {
		log.Fatal("connection error: ", err)
	}

	defer client.Close()

	go client.HandleServerEvents()
	go client.HandleUserInput()

	client.Run()
}
