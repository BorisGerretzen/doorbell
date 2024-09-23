package common

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type TelegramChat struct {
	Result struct {
		ID string `json:"id"`
	} `json:"result"`
}

type TelegramUser struct {
	ChatId string `json:"chat_id"`
	User   string `json:"user"`
}

type TelegramLogin struct {
	AuthDate  int    `json:"auth_date"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	ID        int    `json:"id"`
	Username  string `json:"username,omitempty"`
	Hash      string `json:"hash"`
}

func sendTelegramMessage(token, chatID string, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)

	// Create the payload
	payload := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	// Convert the payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Send the request
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send message, status code: %d", resp.StatusCode)
	}

	return nil
}

func SendTelegramDoorbell(key string, message string, targets []string) error {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	for _, target := range targets {
		wg.Add(1)

		go func(chatID string) {
			defer wg.Done()
			err := sendTelegramMessage(key, fmt.Sprintf("%s", chatID), message)
			if err != nil {
				mu.Lock()
				errors = append(errors, fmt.Errorf("failed to send message to %s: %w", chatID, err))
				mu.Unlock()
			}
		}(target)
	}

	wg.Wait()
	if len(errors) > 0 {
		return fmt.Errorf("encountered errors: %v", errors)
	}

	return nil
}

func (t *TelegramLogin) Validate(key string) bool {
	var fields []string
	if t.AuthDate != 0 {
		fields = append(fields, fmt.Sprintf("auth_date=%d", t.AuthDate))
	}
	if t.FirstName != "" {
		fields = append(fields, fmt.Sprintf("first_name=%s", t.FirstName))
	}
	if t.ID != 0 {
		fields = append(fields, fmt.Sprintf("id=%d", t.ID))
	}
	if t.LastName != "" {
		fields = append(fields, fmt.Sprintf("last_name=%s", t.LastName))
	}
	if t.Username != "" {
		fields = append(fields, fmt.Sprintf("username=%s", t.Username))
	}

	sort.Strings(fields)

	// join with newline
	data := []byte(strings.Join(fields, "\n"))

	// Get hash from key to use as HMAC secret
	h := sha256.New()
	h.Write([]byte(key))
	bs := h.Sum(nil)

	hash := HmacSha256(data, bs)
	return hash == t.Hash
}
