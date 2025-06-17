package wordbomb

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) Connect() error {
	h := http.Header{}
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.1148 Chrome/134.0.6998.205 Electron/35.3.0 Safari/537.36")
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://1266394578702041119.discordsays.com/.proxy/colyseus/server/%s/%s?sessionId=%s", c.ProcessID, c.RoomID, c.SessionID), h)
	if err != nil {
		return err
	}

	c.Conn = conn
	c.SendQueue = make(chan []byte, 999999)
	return nil
}

func (c *Client) Disconnect() error {
	err := c.Conn.Close()
	if err != nil {
		return err
	}
	c.Conn = nil
	return nil
}

func (c *Client) Send(message []byte) error {
	if c.Conn == nil {
		return fmt.Errorf("connection not established")
	}

	c.SendQueue <- message
	return nil
}

func (c *Client) handleSend() {
	if c.HandleSendBound {
		return
	}
	c.HandleSendBound = true
	for msg := range c.SendQueue {
		c.Conn.WriteMessage(websocket.BinaryMessage, msg)
	}
}

func (c *Client) HandleConnection() {
	go c.handleSend()
	for {
		if c.Conn == nil {
			continue
		}
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			if err == websocket.ErrCloseSent {
				c.Conn = nil
				continue
			}
			continue
		}
		switch getPacketID(msg) {
		case "JOIN_ROOM":
			c.handleJoinRoomPacket(msg)
			fmt.Printf("[JOIN_ROOM] reconn token: %s\n", c.ReconnectionToken)
		case "ROOM_DATA":
			payload, err := c.handleRoomDataPacket(msg)
			if err != nil {
				fmt.Println(msg)
				fmt.Println(err, "rd")
				break
			}
			event := payload.MessageType.(string)
			if event == "+" {
				roomInfo := c.AddRoom(payload.MessagePayload)
				fmt.Printf("[ROOM_DATA] New Room: %+v\n", roomInfo)
				break
			}
			if event == "-" {
				fmt.Printf("[ROOM_DATA] Removed Room: %+v\n", payload.MessagePayload)
				c.RemoveRoom(payload.MessagePayload)
				break
			}
			if event == "rooms" {
				c.InitRooms(payload.MessagePayload)
				fmt.Printf("[ROOM_DATA] Init Rooms: %+v\n", c.Rooms[0])
				c.NotificationChannel <- 1
				break
			}
			if event == "correct" {
				fmt.Printf("[ROOM_DATA] Correct: %+v\n", payload.MessagePayload)
				c.UsedWords = append(c.UsedWords, strings.ToLower(payload.MessagePayload.(map[string]interface{})["word"].(string)))
				break
			}
			if event == "hurt" {
				fmt.Printf("[ROOM_DATA] Hurt: %+v\n", payload.MessagePayload)
				break
			}
			if event == "turn" {
				fmt.Printf("[ROOM_DATA] Turn: %+v\n", payload.MessagePayload)
				c.CurrentTurnLetters = payload.MessagePayload.(map[string]interface{})["letter"].(string)
				if payload.MessagePayload.(map[string]interface{})["turn"].(string) == c.SelfID {
					c.NotificationChannel <- 3
				}
				break
			}
			if event == "winner" {
				c.NotificationChannel <- -1
				fmt.Printf("[ROOM_DATA] Winner: %+v\n", payload.MessagePayload)
				break
			}
			if event == "user-letter-healths" {
				fmt.Printf("[ROOM_DATA] User Letter Healths: %+v\n", payload.MessagePayload)
				break
			}
			if event == "countdown" {
				fmt.Printf("[ROOM_DATA] Countdown: %+v\n", payload.MessagePayload)
				break
			}
			if event == "chat" {
				chat := payload.MessagePayload.(map[string]interface{})
				if chat["type"] == "system" && !c.ObtainedSelfID {
					c.MyDiscordUserID = chat["id"].(string)
					c.ObtainedSelfID = true
					for ID, discordID := range c.DiscordUserIDMap {
						if discordID == c.MyDiscordUserID {
							c.SelfID = ID
							break
						}
					}
					if c.SelfID == "" {
						panic("Failed to get self ID")
					}
					c.NotificationChannel <- 2
				}
				fmt.Printf("[ROOM_DATA] Chat: %+v\n", payload.MessagePayload)
				break
			}
			fmt.Printf("[ROOM_DATA] Unknown event: %s\n", event)
		case "ROOM_STATE_PATCH":
			payload, err := c.handleRoomDataPacket(msg)
			if err != nil {
				fmt.Println(err, "rsp")
				break
			}
			event := payload.MessageType.(int)
			if event == -127 {
				fmt.Println("Game started! Current turn:", payload.MessagePayload)
				if payload.MessagePayload.(string) == c.SelfID {
					time.Sleep(3 * time.Second)
					c.NotificationChannel <- 3
					continue
				}
				continue
			}
			if event == -125 {
				c.TurnTimer = payload.MessagePayload.(int)
				continue
			}
			fmt.Printf("[ROOM_STATE_PATCH] EVENT: %+v\n", payload.MessageType)
			fmt.Printf("[ROOM_STATE_PATCH] PAYLOAD: %+v\n", payload.MessagePayload)
		case "ROOM_STATE":
			payload, err := c.handleRoomDataPacket(msg)
			if err != nil {
				fmt.Println(err, "rs")
				break
			}
			event := payload.MessageType.(int)
			if event == -128 {
				c.DiscordUserIDMap = parseUserIDMap(msg)
				if !c.InRoom {
					c.InRoom = true
				}
				break
			}
			fmt.Printf("[ROOM_STATE] Event: %+v\n", event)
			fmt.Printf("[ROOM_STATE] Payload: %+v\n", payload.MessagePayload)
		default:
			fmt.Printf("[%s] %+v\n", getPacketID(msg), msg[1:])
		}
	}
}
