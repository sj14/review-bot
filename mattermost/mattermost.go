package mattermost

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// Send text to mattermost channel.
func Send(channel, text, webhook string) {
	payload := []byte(fmt.Sprintf(`{"channel": "%s", "username": "Review Bot 🧐", "text": "%s"}`, channel, text))

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
