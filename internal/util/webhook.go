package util

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
)

const maxWebhookRetries = 3

func SendJSONWebhook(webhookURL string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	var lastErr error
	for attempt := range maxWebhookRetries + 1 {
		if attempt > 0 {
			backoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second
			time.Sleep(backoff)
		}

		resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(data))
		if err != nil {
			lastErr = fmt.Errorf("failed to send POST request: %w", err)
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("webhook returned status code: %d", resp.StatusCode)
			resp.Body.Close()
			continue
		}

		resp.Body.Close()
		return nil
	}

	return fmt.Errorf("webhook failed after %d retries: %w", maxWebhookRetries, lastErr)
}
