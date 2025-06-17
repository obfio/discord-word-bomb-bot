package wordbomb

import (
	"encoding/json"
	"fmt"
	"time"
)

type Room struct {
	Clients    int       `json:"clients"`
	CreatedAt  time.Time `json:"createdAt"`
	Locked     bool      `json:"locked"`
	MaxClients int       `json:"maxClients"`
	Metadata   Metadata  `json:"metadata"`
	Name       string    `json:"name"`
	Private    bool      `json:"private"`
	ProcessID  string    `json:"processId"`
	RoomID     string    `json:"roomId"`
	Unlisted   bool      `json:"unlisted"`
}

type Metadata struct {
	AuthID   string `json:"auth_id"`
	Avatar   string `json:"avatar"`
	Clients  int    `json:"clients"`
	Diff     int    `json:"diff"`
	Language string `json:"language"`
	MLang    bool   `json:"mlang"`
	Mode     string `json:"mode"`
	Name     string `json:"name"`
	SC       int    `json:"sc"`
	Started  bool   `json:"started"`
	Type     int    `json:"type"`
	WPP      int    `json:"wpp"`
}

func (c *Client) InitRooms(data interface{}) {
	var rooms []*Room

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = json.Unmarshal(jsonData, &rooms)
	if err != nil {
		fmt.Println(err)
	}

	c.RWMutex.Lock()
	c.Rooms = rooms
	c.RWMutex.Unlock()
}

func (c *Client) AddRoom(data interface{}) *Room {
	room := &Room{}

	d := data.([]interface{})
	jsonData, err := json.Marshal(d[1])
	if err != nil {
		fmt.Println(err)
		return nil
	}

	err = json.Unmarshal(jsonData, &room)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	c.RWMutex.Lock()
	c.Rooms = append(c.Rooms, room)
	c.RWMutex.Unlock()

	return room
}

func (c *Client) RemoveRoom(data interface{}) {
	roomID := data.(string)

	c.RWMutex.Lock()
	for i, r := range c.Rooms {
		if r.RoomID == roomID {
			c.Rooms = append(c.Rooms[:i], c.Rooms[i+1:]...)
		}
	}
	c.RWMutex.Unlock()
}
