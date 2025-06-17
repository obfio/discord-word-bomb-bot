package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"slices"
	"strings"

	"github.com/obfio/discord-word-bomb-cheats/wordbomb"
)

type command struct {
	Command int    `json:"command"`
	Message string `json:"message"`
}

type config struct {
	Token         string `json:"token"`
	HostDiscordID string `json:"hostDiscordID"`
	Fast          bool   `json:"fast"`
}

var (
	r1    = regexp.MustCompile(`^[a-zA-Z]+$`)
	words = []string{}

	httpCommandChannel = make(chan *command)

	c = &config{}
)

func init() {
	f, err := os.ReadFile("words.txt")
	if err != nil {
		panic(err)
	}
	for _, word := range strings.Split(strings.ReplaceAll(string(f), "\r", ""), "\n") {
		// only append the word if it only contains letters, no numbers or special characters
		if r1.MatchString(word) {
			words = append(words, word)
		}
	}

	f, err = os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(f, &c)
	if err != nil {
		panic(err)
	}
}

func main() {
	client := wordbomb.NewClient(c.Token)
	client.SolveMode = c.Fast
	err := client.GetSessionInfo()
	if err != nil {
		panic(err)
	}

	notiChan := make(chan int)
	client.NotificationChannel = notiChan

	err = client.Connect()
	if err != nil {
		panic(err)
	}

	go client.HandleConnection()

	go func() {
		http.HandleFunc("/command", sendCommand)
		http.ListenAndServe(":8080", nil)
	}()

	go func() {
		for command := range httpCommandChannel {
			fmt.Println("Received command:", command)
			switch command.Command {
			case 1:
				fmt.Println("Sending chat message:", command.Message)
				client.SendChat(command.Message)
			}
		}
	}()

	for {
		noti := <-notiChan
		if noti == 1 {
			client.RWMutex.Lock()
			r := &wordbomb.Room{}
			r = nil
			for _, room := range client.Rooms {
				if room.Metadata.AuthID == c.HostDiscordID {
					r = room
					break
				}
			}
			client.RWMutex.Unlock()
			if r == nil {
				fmt.Println("Room not found")
				continue
			}
			fmt.Println("Attempting to join room", r.Metadata.Name)
			client.JoinRoom(r)
			continue
		}
		if noti == 2 {
			fmt.Println("Room joined")
			fmt.Println("Got self ID:", client.SelfID)
			continue
		}
		if noti == 3 {
			fmt.Println("My turn, sending word")
			word := getWord(client.CurrentTurnLetters, client.UsedWords)
			fmt.Printf("=====================\nSending Word: %s\nUsed Words: %+v\n=====================\n", word, client.UsedWords)
			client.SendWord(word)
			client.UsedWords = append(client.UsedWords, strings.ToLower(word))
			continue
		}
		if noti == -1 {
			fmt.Println("Game over!")
			client.UsedWords = []string{}
			continue
		}
	}
}

func getWord(letters string, usedWords []string) string {
	letters = strings.ToLower(letters)
	longestWord := ""
	longestWordLength := 0
	// get longest word that contains the 2 letters in it and hasn't been used yet
	for _, word := range words {
		if strings.Contains(word, letters) && !slices.Contains(usedWords, word) {
			if len(word) > longestWordLength {
				longestWord = word
				longestWordLength = len(word)
			}
		}
	}
	return longestWord
}
