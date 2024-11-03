package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"

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
	conn *websocket.Conn
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
		conn: conn,
	}, nil
}

func (c *Client) HandleSend() {
	sc := bufio.NewScanner(os.Stdin)

	for {
		if sc.Scan() {
			c.conn.WriteMessage(websocket.TextMessage, sc.Bytes())
		}
		if err := sc.Err(); err != nil {
			color.Red("Scan error: %v", err)
			return
		}
	}
}

func (c *Client) HandleRecv() {
	for {
		kind, msg, err := c.conn.ReadMessage()
		if err != nil {
			color.Red("Websocket error: %v", err)
			return
		}
		color.Blue("%v", Event{Kind: kind, Message: msg})
	}
}

func (c *Client) Close() {
	color.Green("Terminating websocket connection from %v", c.conn.RemoteAddr())
	c.conn.Close()
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

	recvCh := make(chan os.Signal, 1)
	signal.Notify(recvCh, os.Interrupt)

	go client.HandleRecv()
	go client.HandleSend()

	// wait for client to terminate the application
	// needed to terminate websocket connection gracefully
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}
