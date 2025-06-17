package wordbomb

type sessionInfo struct {
	Room struct {
		Clients   int    `json:"clients"`
		Locked    bool   `json:"locked"`
		Private   bool   `json:"private"`
		Unlisted  bool   `json:"unlisted"`
		Name      string `json:"name"`
		ProcessID string `json:"processId"`
		RoomID    string `json:"roomId"`
	} `json:"room"`
	SessionID string `json:"sessionId"`
}
