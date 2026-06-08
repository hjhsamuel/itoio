package core

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hjhsamuel/itoio/internal/entities"
	"github.com/hjhsamuel/itoio/pkg"
)

const (
	websocketWriteWait = time.Second * 2
)

type ConnMessage struct {
	Info  *ConnBase
	Event Event
}

type ConnBase struct {
	ID    string
	Write chan entities.Message
}

type WebSocketConn struct {
	conn *websocket.Conn
	once pkg.Once
	read chan *ConnMessage
	info *ConnBase
}

func (c *WebSocketConn) Start() {
	go c.doReading()
	go c.doWriting()
}

func (c *WebSocketConn) closeNormally(code int, msg string) {
	c.once.Do(func() {
		c.close(code, msg)
	})
}

func (c *WebSocketConn) closeWithError(code int, msg string) {
	c.once.Do(func() {
		go func() {
			c.read <- &ConnMessage{
				Info: c.info,
				Event: &EventDisconnect{
					Code: code,
					Msg:  msg,
				},
			}
		}()
		c.close(code, msg)
	})
}

func (c *WebSocketConn) close(code int, msg string) {
	message := websocket.FormatCloseMessage(code, msg)
	_ = c.conn.WriteControl(websocket.CloseMessage, message, time.Now().Add(websocketWriteWait))
	c.conn.Close()
}

func (c *WebSocketConn) doReading() {
	for {
		t, m, err := c.conn.ReadMessage()
		if err != nil {
			c.closeWithError(websocket.CloseNormalClosure, fmt.Sprintf("read error: %v", err))
			return
		}
		if t == websocket.BinaryMessage {
			c.closeWithError(websocket.CloseUnsupportedData, "unsupported binary message")
			return
		}

		event, err := c.switchEvent(m)
		if err != nil {
			c.closeWithError(websocket.CloseUnsupportedData, err.Error())
			return
		}
		c.read <- &ConnMessage{
			Info:  c.info,
			Event: event,
		}
	}
}

func (c *WebSocketConn) switchEvent(m []byte) (Event, error) {
	var req *Request
	if err := json.Unmarshal(m, &req); err != nil {
		return nil, err
	}
	var event Event
	switch req.Type {
	case EventTypeOffer, EventTypeAnswer, EventTypeCandidate, EventTypeStopShare:
		var obj *EventSignaling
		if err := json.Unmarshal(req.Data, &obj); err != nil {
			return nil, err
		}
		obj.Typ = req.Type
		event = obj
	case EventTypeCreate:
		var obj *EventCreate
		if err := json.Unmarshal(req.Data, &obj); err != nil {
			return nil, err
		}
		event = obj
	case EventTypeJoin:
		var obj *EventJoin
		if err := json.Unmarshal(req.Data, &obj); err != nil {
			return nil, err
		}
		event = obj
	case EventTypeLeave:
		event = &EventLeave{}
	default:
		return nil, fmt.Errorf("unknown event type: %s", req.Type)
	}
	return event, nil
}

func (c *WebSocketConn) doWriting() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			err := c.conn.WriteMessage(websocket.PingMessage, nil)
			if err != nil {
				c.closeWithError(websocket.CloseNormalClosure, "ping timeout")
				return
			}
		case data := <-c.info.Write:
			if data.Type() == entities.MessageTypeCloseConn {
				c.closeNormally(websocket.CloseNormalClosure, "close connection")
				return
			}

			content, err := json.Marshal(data)
			if err != nil {
				c.closeWithError(websocket.CloseNormalClosure, fmt.Sprintf("marshal error: %v", err))
				return
			}
			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			err = c.conn.WriteJSON(&Request{
				Type: data.Type(),
				Data: content,
			})
			if err != nil {
				c.closeWithError(websocket.CloseNormalClosure, fmt.Sprintf("write error: %v", err))
				return
			}
		}
	}
}

func NewWebSocketConn(id string, conn *websocket.Conn, read chan *ConnMessage) *WebSocketConn {
	return &WebSocketConn{
		info: &ConnBase{
			ID:    id,
			Write: make(chan entities.Message, 1),
		},
		conn: conn,
		read: read,
	}
}

func Write(ch chan entities.Message, msg entities.Message) {
	select {
	case <-time.After(time.Second * 2):
	case ch <- msg:
	}
}
