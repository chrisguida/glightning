package jrpc2

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync/atomic"
	"encoding/json"
	"log"
	"time"
)

// a client needs to be able to ...
// - 'call' a method which is really...
// - fire off a request 
// - receive a result back (& match that result to outbound request)
// bonus round: 
//    - send and receive in batches 
type Client struct {
	requestQueue chan *Request
	pendingReq map[string]chan *RawResponse
	requestCounter int64
	shutdown bool
	timeout time.Duration
}

func NewClient() *Client {
	client := &Client{}
	client.requestQueue = make(chan *Request)
	client.pendingReq = make(map[string]chan *RawResponse)
	client.timeout = time.Duration(20)
	return client
}

func (c *Client) SetTimeout(secs uint) {
	c.timeout = time.Duration(secs)
}

func (c *Client) StartUp(in, out *os.File) {
	c.shutdown = false
	go c.setupWriteQueue(out)
	c.readQueue(in)
}

func (c *Client) Shutdown() {
	c.shutdown = true
	close(c.requestQueue)
	for _, v := range c.pendingReq {
		close(v)
	}
	c.pendingReq = make(map[string]chan *RawResponse)
	c.requestQueue = make(chan *Request)
}

func (c *Client) setupWriteQueue(outW io.Writer) {
	out := bufio.NewWriter(outW)
	defer out.Flush()
	twoNewlines := []byte("\n\n")
	for request := range c.requestQueue {
		data, err := json.Marshal(request)
		if err != nil {
			log.Println(err.Error())
			continue
		}
		data = append(data, twoNewlines...)
		out.Write(data)
		out.Flush()
	}
}

func (c *Client) readQueue(in io.Reader) {
	scanner := bufio.NewScanner(in)
	scanner.Split(scanDoubleNewline)
	for scanner.Scan() && !c.shutdown {
		msg := scanner.Bytes()
		go processResponse(c, msg)
	}
}

func processResponse(c *Client, msg []byte) {
	var rawResp *RawResponse
	err := json.Unmarshal(msg, &rawResp)
	if err != nil {
		log.Printf("Error parsing response %s", err.Error())
		return
	}

	// the response should have an ID
	if rawResp.Id == nil || rawResp.Id.Val() == "" {
		// no id means there's no one listening
		// for this to come back through ...
		log.Printf("No Id provided %v", rawResp)
		return
	}

	// look up 'reply channel' via the
	// client (should have a registry of
	// resonses that are waiting...)
	c.sendResponse(rawResp.Id.Val(), rawResp)
}

func (c *Client) sendResponse(id string, resp *RawResponse) {
	respChan, exists := c.pendingReq[id]
	if !exists {
		log.Printf("No return channel found for response with id %s", id)
		return
	}
	respChan <- resp
	delete(c.pendingReq, id)
}

// Isses an RPC call. Is blocking. Times out after {timeout}
// seconds (set on client).
func (c *Client) Request(m Method, resp interface{}) (error) {
	if c.shutdown {
		return fmt.Errorf("Client is shutdown")
	}
	id := c.NextId()
	// set up to get a response back
	replyChan := make(chan *RawResponse, 1)
	c.pendingReq[id.Val()] = replyChan

	// send the request out
	req := &Request{id, m}
	c.requestQueue <- req

	select {
	case rawResp := <-replyChan:
		return handleReply(rawResp, resp)
	case <- time.After(c.timeout * time.Second):
		delete(c.pendingReq, id.Val())
		return fmt.Errorf("Request timed out")
	}
}

func handleReply(rawResp *RawResponse, resp interface{}) error {
	if rawResp == nil {
		return fmt.Errorf("Pipe closed unexpectedly, nil result")
	}

	// when the response comes back, it will either have an error,
	// that we should parse into an 'error' (depending on the code?)
	if rawResp.Error != nil {
		return rawResp.Error.ToErr()
	}

	// or a raw response, that we should json map into the 
	// provided resp (interface)
	return json.Unmarshal(rawResp.Raw, resp)
}

// for now, use a counter as the id for requests
func (c *Client) NextId() *Id {
	val := atomic.AddInt64(&c.requestCounter, 1)
	return NewIdAsInt(val)
}
