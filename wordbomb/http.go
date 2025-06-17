package wordbomb

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	http "github.com/bogdanfinn/fhttp"
)

func (c *Client) GetSessionInfo() error {
	req, err := http.NewRequest("POST", "https://1266394578702041119.discordsays.com/.proxy/colyseus/server/matchmake/joinOrCreate/lobby", strings.NewReader("{}"))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.1148 Chrome/134.0.6998.205 Electron/35.3.0 Safari/537.36")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get session info: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var sessionInfo sessionInfo
	err = json.Unmarshal(b, &sessionInfo)
	if err != nil {
		return err
	}

	c.RoomID = sessionInfo.Room.RoomID
	c.SessionID = sessionInfo.SessionID
	c.ProcessID = sessionInfo.Room.ProcessID

	if c.RoomID == "" || c.SessionID == "" || c.ProcessID == "" {
		return fmt.Errorf("failed to get session info: %v", sessionInfo)
	}

	return nil
}

func (c *Client) GetNewRoomSessionID(r *Room) error {
	req, err := http.NewRequest("POST", "https://1266394578702041119.discordsays.com/.proxy/colyseus/server/matchmake/joinById/"+r.RoomID, strings.NewReader(`{"spectate":false,"wbPlayer":false,"locale":"en-US"}`))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", c.AuthToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) discord/1.0.1148 Chrome/134.0.6998.205 Electron/35.3.0 Safari/537.36")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to get new room session id: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var sessionInfo sessionInfo
	err = json.Unmarshal(b, &sessionInfo)
	if err != nil {
		return err
	}

	c.RoomID = sessionInfo.Room.RoomID
	c.SessionID = sessionInfo.SessionID
	c.ProcessID = sessionInfo.Room.ProcessID

	if c.RoomID == "" || c.SessionID == "" || c.ProcessID == "" {
		return fmt.Errorf("failed to get session info: %v", sessionInfo)
	}
	return nil
}
