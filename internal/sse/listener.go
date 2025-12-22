package sse

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Event represents an SSE event from ChatBridge
type Event struct {
	Type      string `json:"type"`
	SessionID string `json:"session_id"`
	From      string `json:"from"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// Listener handles SSE connections
type Listener struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// NewListener creates a new SSE listener
func NewListener(baseURL, apiKey string) *Listener {
	return &Listener{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client: &http.Client{
			// No timeout - SSE is long-lived
		},
	}
}

// ResponseHandler is called when a response is received
type ResponseHandler func(event *Event)

// ReminderHandler is called at reminder intervals
type ReminderHandler func(elapsed, remaining time.Duration)

// ListenOptions configures the listener behavior
type ListenOptions struct {
	Timeout          time.Duration
	ReminderInterval time.Duration
	OnEvent          ResponseHandler
	OnReminder       ReminderHandler
}

// Listen connects to the SSE endpoint and waits for responses
// It returns when a message is received, timeout occurs, or context is cancelled
func (l *Listener) Listen(ctx context.Context, sessionID string, timeout time.Duration, onEvent ResponseHandler) (*Event, error) {
	return l.ListenWithOptions(ctx, sessionID, ListenOptions{
		Timeout: timeout,
		OnEvent: onEvent,
	})
}

// ListenWithOptions connects to the SSE endpoint with full configuration
func (l *Listener) ListenWithOptions(ctx context.Context, sessionID string, opts ListenOptions) (*Event, error) {
	url := fmt.Sprintf("%s/api/events/%s", l.BaseURL, sessionID)
	startTime := time.Now()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("X-API-Key", l.APIKey)
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := l.Client.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout waiting for response")
		}
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("cancelled")
		}
		return nil, fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, fmt.Errorf("unauthorized: invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	// Set up reminder ticker if configured
	var reminderTicker *time.Ticker
	var reminderChan <-chan time.Time
	if opts.ReminderInterval > 0 && opts.OnReminder != nil {
		reminderTicker = time.NewTicker(opts.ReminderInterval)
		reminderChan = reminderTicker.C
		defer reminderTicker.Stop()
	}

	// Create channels for SSE events
	eventChan := make(chan *Event, 1)
	errChan := make(chan error, 1)

	// Read SSE in goroutine
	go func() {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()

			// Skip empty lines and comments
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}

			// Handle event type
			if strings.HasPrefix(line, "event:") {
				eventType := strings.TrimSpace(strings.TrimPrefix(line, "event:"))
				if eventType == "connected" {
					continue
				}
			}

			// Handle data
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))

				var event Event
				if err := json.Unmarshal([]byte(data), &event); err != nil {
					continue
				}

				if event.Type == "message" && event.Content != "" {
					eventChan <- &event
					return
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		} else {
			errChan <- fmt.Errorf("connection closed without response")
		}
	}()

	// Wait for event, reminder, or context done
	for {
		select {
		case event := <-eventChan:
			if opts.OnEvent != nil {
				opts.OnEvent(event)
			}
			return event, nil

		case err := <-errChan:
			return nil, err

		case <-reminderChan:
			elapsed := time.Since(startTime)
			remaining := opts.Timeout - elapsed
			if remaining < 0 {
				remaining = 0
			}
			opts.OnReminder(elapsed, remaining)

		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return nil, fmt.Errorf("timeout waiting for response")
			}
			return nil, fmt.Errorf("cancelled")
		}
	}
}

// FormatTimestamp formats a Unix timestamp for display
func FormatTimestamp(ts int64) string {
	if ts == 0 {
		return time.Now().Format("15:04:05")
	}
	return time.Unix(ts, 0).Format("15:04:05")
}
