package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// YiCameraRequest request for yi camera
type YiCameraRequest struct {
	ID    int         `json:"msg_id"`
	Param interface{} `json:"param,omitempty"`
	Token int         `json:"token"`
}

// YiCameraResponse response for request
type YiCameraResponse struct {
	Rval  int         `json:"rval"`
	ID    int         `json:"msg_id"`
	Param interface{} `json:"param"`
	Type  string      `json:"type"`
}

// Listener msg listener
type Listener struct {
	Channel chan YiCameraResponse
	MsgIDs  []int
}

// Want check if listener want message
func (c *Listener) Want(MsgID int) bool {
	for _, msgID := range c.MsgIDs {
		if msgID == MsgID {
			return true
		}
	}
	return false
}

// YiCameraClient Yi Camera Client
type YiCameraClient struct {
	Address   string
	conn      net.Conn
	listeners []Listener
	token     int
	active    bool
}

// Connect make connection to camera
func (c *YiCameraClient) Connect() {
	var err error
	c.conn, err = net.Dial("tcp", c.Address)
	if err != nil {
		panic(err)
	}
	c.conn.(*net.TCPConn).SetKeepAlive(true)
	c.conn.(*net.TCPConn).SetKeepAlivePeriod(time.Second * 30)
	c.active = true
	go c.dispatcher()
}

// Disconnect disconnect
func (c *YiCameraClient) Disconnect() {
	c.active = true
}

// IsAuthenticated check if client is authenticated
func (c *YiCameraClient) IsAuthenticated() bool {
	if c.token != 0 {
		return true
	}
	return false
}

// IsActive check if active
func (c *YiCameraClient) IsActive() bool {
	if c.active == true {
		return true
	}
	return false
}

// Authenticate authenticate client
func (c *YiCameraClient) Authenticate() bool {
	fmt.Println("Authentication")
	json, _ := json.Marshal(YiCameraRequest{ID: 257, Token: 0})
	fmt.Println(string(json))
	fmt.Fprintf(c.conn, string(json))
	channel := c.waitMsg([]int{257})
	for i := 0; i < 10; i++ {
		select {
		case msg, ok := <-channel:
			if !ok {
				break
			}
			if msg.Param == nil {
				continue
			}
			fmt.Println(msg.Param)
			c.token = int(msg.Param.(float64))
			c.active = true
			return true
		default:
			time.Sleep(1000 * time.Millisecond)
		}
	}
	return false
}

// Stream enable live stream
func (c *YiCameraClient) Stream(wait int) {
	fmt.Println("Trying to start stream")
	if c.IsAuthenticated() == false {
		return
	}
	fmt.Println("Starting stream")
	json, _ := json.Marshal(YiCameraRequest{ID: 259, Token: c.token, Param: "none_force"})
	fmt.Fprintf(c.conn, string(json))
	channel := c.waitMsg([]int{})
	if wait == 0 {
		<-channel
	} else {
		for i := 0; i < wait; i++ {
			if !c.IsActive() {
				break
			}
			time.Sleep(1000 * time.Millisecond)
		}
	}
	fmt.Println("Stream stopped")
}

func (c *YiCameraClient) dispatcher() {
	fmt.Println("Starting dispatcher")
	c.listeners = []Listener{}
	d := json.NewDecoder(c.conn)
	for c.active {
		var msg YiCameraResponse
		err := d.Decode(&msg)
		if err != nil {
			break
		}
		for _, listener := range c.listeners {
			if listener.Want(msg.ID) {
				listener.Channel <- msg
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	for _, listener := range c.listeners {
		close(listener.Channel)
	}
	c.listeners = nil
	c.active = false
	c.token = 0
}

func (c *YiCameraClient) waitMsg(MsgIDs []int) chan YiCameraResponse {
	var listener Listener = Listener{MsgIDs: MsgIDs, Channel: make(chan YiCameraResponse, 10)}
	c.listeners = append(c.listeners, listener)
	return listener.Channel
}
