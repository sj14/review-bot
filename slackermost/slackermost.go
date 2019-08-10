package slackermost

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type message struct {
	User    string `json:"username,omitempty"` // only mattermost(?)
	Channel string `json:"channel,omitempty"`  // only mattermost(?)
	Text    string `json:"text"`
}

// Send text to Slack or Mattermost channel.
func Send(channel, text, webhook string) error {
	payload, err := json.Marshal(message{User: "Review Bot 🧐", Channel: channel, Text: text})
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v\n", err)
	}

	defer func() {
		if resp == nil || resp.Body == nil {
			return
		}
		if err := resp.Body.Close(); err != nil {
			log.Printf("failed to close slackermost client: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("response status: %v; header: %v; body: %v", resp.Status, resp.Header, string(body))
	}
	return nil
}
