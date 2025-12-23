package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communication with the ChatBridge API
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

// NewClient creates a new API client
func NewClient(baseURL, apiKey string) *Client {
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendMessageRequest is the request body for sending messages
type SendMessageRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
	SysName   string `json:"sys_name,omitempty"` // Name of the AI agent/system (WhatsApp only)
}

// SendMessageResponse is the response from sending messages
type SendMessageResponse struct {
	Success   bool   `json:"success"`
	MessageID string `json:"message_id,omitempty"`
	SessionID string `json:"session_id,omitempty"` // Actual session ID to use (may differ from request)
	Error     string `json:"error,omitempty"`
}

// HealthResponse is the response from the health endpoint
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// AccountResponse is the response from account verification
type AccountResponse struct {
	Email              string `json:"email"`
	Phone              string `json:"phone"`
	SubscriptionTier   string `json:"subscription_tier"`
	SubscriptionStatus string `json:"subscription_status"`
	WAMessagesUsed     int    `json:"wa_messages_used"`
	WAMessagesLimit    int    `json:"wa_messages_limit"`
	SMSMessagesUsed    int    `json:"sms_messages_used"`
	SMSMessagesLimit   int    `json:"sms_messages_limit"`
}

// SendSMS sends an SMS message
func (c *Client) SendSMS(message, sessionID string) (*SendMessageResponse, error) {
	return c.sendMessage("/api/sendsms", message, sessionID, "")
}

// SendWhatsApp sends a WhatsApp message
func (c *Client) SendWhatsApp(message, sessionID, sysName string) (*SendMessageResponse, error) {
	return c.sendMessage("/api/sendwhatsapp", message, sessionID, sysName)
}

func (c *Client) sendMessage(endpoint, message, sessionID, sysName string) (*SendMessageResponse, error) {
	req := SendMessageRequest{
		Message:   message,
		SessionID: sessionID,
		SysName:   sysName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("%s", errResp.Error)
		}
		return nil, fmt.Errorf("API error: %s", resp.Status)
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		if result.Error != "" {
			return nil, fmt.Errorf("%s", result.Error)
		}
		return nil, fmt.Errorf("message send failed")
	}

	return &result, nil
}

// Health checks if the API is reachable
func (c *Client) Health() (*HealthResponse, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/health")
	if err != nil {
		return nil, fmt.Errorf("API unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned: %s", resp.Status)
	}

	var result HealthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// ValidateKey checks if the API key is valid by making an authenticated request
// This is a lightweight check - we'll use the SSE endpoint which requires auth
func (c *Client) ValidateKey() error {
	// Try to connect to SSE with a unique session - it will validate the key
	// Session must match "afk.*" pattern for NATS stream
	// Use unique ID to avoid consumer conflicts with other validation attempts
	sessionID := fmt.Sprintf("afk-validate-%d", time.Now().UnixNano())
	req, err := http.NewRequest("GET", c.BaseURL+"/api/events/"+sessionID, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-API-Key", c.APIKey)

	// Use a short timeout client for validation
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer resp.Body.Close()

	// 401 means invalid key, anything else (including 200 for SSE) means valid
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("invalid API key")
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("subscription not active")
	}

	return nil
}
