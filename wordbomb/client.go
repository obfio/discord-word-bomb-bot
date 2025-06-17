package wordbomb

import (
	"sync"

	"github.com/gorilla/websocket"

	tlsclient "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

type Client struct {
	Conn              *websocket.Conn
	RoomID            string
	SessionID         string
	ProcessID         string
	AuthToken         string
	HTTPClient        tlsclient.HttpClient
	SendQueue         chan []byte
	HandshakeOffset   int
	ReconnectionToken string

	RWMutex sync.RWMutex

	Rooms []*Room

	NotificationChannel chan int

	HandleSendBound bool

	InRoom bool

	SelfID string

	CurrentTurnLetters string

	TurnTimer int

	UsedWords []string

	DiscordUserIDMap map[string]string
	MyDiscordUserID  string
	ObtainedSelfID   bool

	SolveMode bool
}

func NewClient(token string) *Client {
	if token[:2] == "ey" {
		token = "Bearer " + token
	}
	tlsProfile := profiles.Chrome_133

	opts := []tlsclient.HttpClientOption{
		tlsclient.WithTimeoutSeconds(60),
		tlsclient.WithInsecureSkipVerify(),
		tlsclient.WithClientProfile(tlsProfile),
		tlsclient.WithNotFollowRedirects(),
		tlsclient.WithForceHttp1(),
	}
	httpClient, err := tlsclient.NewHttpClient(nil, opts...)
	if err != nil {
		panic(err)
	}

	return &Client{
		HTTPClient: httpClient,
		AuthToken:  token,
	}
}
