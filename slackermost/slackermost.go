package slackermost

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

type message struct {
	User    string `json:"username,omitempty"` // only mattermost(?)
	Channel string `json:"channel,omitempty"`  // only mattermost(?)
	Text    string `json:"text"`
}

// Send text to Slack or Mattermost channel.
func Send(channel, text, webhook string) {
	payload, err := json.Marshal(message{User: "Review Bot 🧐", Channel: channel, Text: text})
	if err != nil {
		log.Fatalf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(payload))
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %v\n", err)
	}

	defer func() {
		if resp == nil || resp.Body == nil {
			return
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close mattermost client: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Println("response Status:", resp.Status)
		log.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		log.Println("response Body:", string(body))
		os.Exit(1)
	}
}
