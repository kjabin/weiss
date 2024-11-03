package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/fatih/color"
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
	color.Yellow("Connecting to url: %v", url)
	conn, resp, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return Client{}, err
	}
	if resp != nil && resp.StatusCode == 101 {
		color.Yellow("Websocket connection established with %v", conn.RemoteAddr())
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
			color.Red("Scan error: %v", err)
			return
		}
	}
}

func (c *Client) Run() {
	for {
		select {
		case e := <-c.recvCh:
			color.Blue("%v", e)
		case e := <-c.sendCh:
			c.conn.WriteMessage(e.Kind, e.Message)
		}
	}
}
func (c *Client) HandleServerEvents() {
	for {
		kind, message, err := c.conn.ReadMessage()
		if err != nil {
			color.Red("Websocket error: %v", err)
			os.Exit(1)
		}
		c.recvCh <- Event{
			Kind:    kind,
			Message: message,
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		color.Red("Websocket URL required")
		os.Exit(1)
	}
	url := os.Args[1]

	client, err := NewClient(url)
	if err != nil {
		color.Red("Connection error: %v", err)
		os.Exit(1)
	}

	defer client.Close()

	go client.HandleServerEvents()
	go client.HandleUserInput()

	client.Run()
}
