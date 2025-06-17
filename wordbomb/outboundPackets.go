package wordbomb

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) JoinRoom(room *Room) error {
	c.Disconnect()
	err := c.GetNewRoomSessionID(room)
	if err != nil {
		return err
	}
	h := http.Header{}
	h.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.1148 Chrome/134.0.6998.205 Electron/35.3.0 Safari/537.36")
	conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("wss://1266394578702041119.discordsays.com/.proxy/colyseus/server/%s/%s?sessionId=%s", room.ProcessID, room.RoomID, c.SessionID), h)
	if err != nil {
		return err
	}

	c.Conn = conn
	return nil
}

var (
	firstLetterPacket = []byte{0xd, 0xa4, 0x77, 0x6f, 0x72, 0x64, 0x82, 0xa1, 0x74, 0xa1, 0x72, 0xa1, 0x77, 0xa1}
	otherLetterPacket = []byte{0xd, 0xa4, 0x77, 0x6f, 0x72, 0x64, 0x82, 0xa1, 0x74, 0xa1, 0x61, 0xa1, 0x61, 0xa1}

	submitWordPacket = []byte{0x0d, 0xa6, 0x73, 0x75, 0x62, 0x6d, 0x69, 0x74, 0x81, 0xa4, 0x77, 0x6f, 0x72, 0x64}
)

func (c *Client) SendWord(word string) error {
	if len(word) == 0 {
		return errors.New("word is empty")
	}
	word = strings.ToUpper(word)
	// first letter is different
	// sleep for 1/10th of the turn timer
	time.Sleep(time.Duration(c.TurnTimer/10) * time.Millisecond)
	if !c.SolveMode {
		time.Sleep(time.Duration(c.TurnTimer/10) * time.Second)
	}
	packet := []byte{}
	packet = append(packet, firstLetterPacket...)
	packet = append(packet, byte(word[0]))
	c.Send(packet)
	for i := 1; i < len(word); i++ {
		// sleep for (7/10ths of the turn timer)/len(word) ms
		time.Sleep(time.Duration(c.TurnTimer*7/10/len(word)) * time.Millisecond)
		if !c.SolveMode {
			time.Sleep(time.Duration(c.TurnTimer*7/10/len(word)) * time.Second)
		}
		packet = []byte{}
		packet = append(packet, otherLetterPacket...)
		packet = append(packet, byte(word[i]))
		c.Send(packet)
	}
	// submit word packet
	// sleep for 1/10th of the turn timer
	time.Sleep(time.Duration(c.TurnTimer/10) * time.Millisecond)
	if !c.SolveMode {
		time.Sleep(time.Duration(c.TurnTimer/10) * time.Second)
	}
	word = formatWord(word)
	packet = []byte{}
	packet = append(packet, submitWordPacket...)
	msgpackPrefix := 0xA0 | byte(len(word))
	packet = append(packet, msgpackPrefix)
	packet = append(packet, []byte(word)...)
	fmt.Println(string(packet))
	c.Send(packet)
	return nil
}

/*
	function Pe(e, t=234) {
	    const n = e.split("")
	      , a = function(e) {
	        return () => (e ^= e << 13,
	        e ^= e >> 17,
	        ((e ^= e << 5) < 0 ? 1 + ~e : e) % 1e5 / 1e5)
	    }(t);
	    for (let i = n.length - 1; i > 0; i--) {
	        const e = Math.floor(a() * (i + 1));
	        [n[i],n[e]] = [n[e], n[i]]
	    }
	    return n.join("")
	}
*/
func formatWord(word string) string {
	n := strings.Split(word, "")
	t := int32(234)
	for i := len(n) - 1; i > 0; i-- {
		fmt.Println(i)
		t1, asd := a(t)
		t = t1
		// fmt.Println(asd)
		e := int(math.Floor(asd * float64(i+1)))
		n[i], n[e] = n[e], n[i]
	}
	return strings.Join(n, "")
}

func a(e int32) (int32, float64) {
	e ^= e << 13
	e ^= e >> 17
	e ^= e << 5
	tmp := e
	if e < 0 {
		tmp = 1 + ^e
	}
	return e, float64(tmp%1e5) / 1e5
}

var (
	chatPacket = []byte{0x0d, 0xa4, 0x63, 0x68, 0x61, 0x74, 0x81, 0xa7, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65}
)

func (c *Client) SendChat(message string) {
	packet := []byte{}
	packet = append(packet, chatPacket...)
	msgpackPrefix := 0xA0 | byte(len(message))
	packet = append(packet, msgpackPrefix)
	packet = append(packet, []byte(message)...)
	c.Send(packet)
}
