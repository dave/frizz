package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/dave/frizz/config"
	"github.com/dave/frizz/server/messages"
	"github.com/dave/services"
	"github.com/gorilla/websocket"
)

func (h *Handler) Command(ctx context.Context, req *http.Request, send func(message services.Message), receive chan services.Message) error {
	select {
	case m := <-receive:
		switch m := m.(type) {
		case messages.Types:
			return h.Types(ctx, m, req, send, receive)
		default:
			return fmt.Errorf("invalid init message %T", m)
		}
	case <-time.After(config.WebsocketInstructionTimeout):
		return errors.New("timed out waiting for instruction from client")
	}
}

func (h *Handler) Socket(w http.ResponseWriter, req *http.Request) {

	h.Waitgroup.Add(1)
	defer func() {
		h.Waitgroup.Done()
	}()

	ctx, cancel := context.WithTimeout(req.Context(), config.WebsocketTimeout)
	defer func() {
		cancel()
	}()

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		h.storeError(ctx, fmt.Errorf("upgrading request to websocket: %v", err), req)
		return
	}

	var sendWg sync.WaitGroup
	sendChan := make(chan services.Message, 256)
	receive := make(chan services.Message, 256)
	var finished bool

	send := func(message services.Message) {
		if finished {
			return // prevent more messages from being sent after we want to finish
		}
		sendWg.Add(1)
		sendChan <- message
	}

	defer func() {
		finished = true // we won't be adding any more messages to the send channel
		sendWg.Wait()   // wait for in-flight sends to finish
		close(sendChan) // close the sendChan, so the send loop will exit
		conn.Close()    // finally close the websocket
	}()

	// Recover from any panic and log the error.
	defer func() {
		if r := recover(); r != nil {
			h.sendAndStoreError(ctx, send, "", errors.New(fmt.Sprintf("panic recovered: %s", r)), req)
		}
	}()

	// Set up a ticker to ping the client regularly
	go func() {
		ticker := time.NewTicker(config.WebsocketPingPeriod)
		defer func() {
			ticker.Stop()
			cancel()
		}()
		for {
			select {
			case message, ok := <-sendChan:
				if !ok {
					// the send channel was closed - exit immediately
					conn.WriteMessage(websocket.CloseMessage, []byte{})
					return
				}
				func() {
					defer sendWg.Done()
					b, err := messages.Marshal(message)
					if err != nil {
						fmt.Println(err)
						return
					}
					conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
					conn.WriteMessage(websocket.BinaryMessage, b)
				}()
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(config.WebsocketWriteTimeout))
				conn.WriteMessage(websocket.PingMessage, nil)
			}
		}
	}()

	// React to pongs from the client
	go func() {
		defer func() {
			cancel()
		}()
		conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(config.WebsocketPongTimeout))
			return nil
		})
		for {
			messageType, messageBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
					// Don't bother storing an error if the client disconnects gracefully
					break
				}
				h.storeError(ctx, err, req)
				break
			}
			if messageType == websocket.CloseMessage {
				break
			}
			message, err := messages.Unmarshal(messageBytes)
			if err != nil {
				h.storeError(ctx, err, req)
				break
			}
			select {
			case receive <- message:
			default:
			}
		}
	}()

	// React to the server shutdown signal
	go func() {
		select {
		case <-h.shutdown:
			h.sendAndStoreError(ctx, send, "", errors.New("server shut down"), req)
			cancel()
		case <-ctx.Done():
		}
	}()

	// Request a slot in the queue...
	start, end, err := h.Queue.Slot(func(position int) {
		send(messages.Queueing{Position: position})
	})
	if err != nil {
		h.sendAndStoreError(ctx, send, "", err, req)
		return
	}

	// Signal to the queue that processing has finished.
	defer func() {
		close(end)
	}()

	// Wait for the slot to become available.
	select {
	case <-start:
		// continue
	case <-ctx.Done():
		return
	}

	// Send a message to the client that queue step has finished.
	send(messages.Queueing{Done: true})

	if err := h.Command(ctx, req, send, receive); err != nil {
		h.sendAndStoreError(ctx, send, "", err, req)
		return
	}

	return
}
