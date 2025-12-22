package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/davedotdev/afk/internal/config"
)

// Formatter handles output formatting based on the configured format
type Formatter struct {
	format config.OutputFormat
	quiet  bool
}

// New creates a new Formatter
func New(format config.OutputFormat, quiet bool) *Formatter {
	return &Formatter{
		format: format,
		quiet:  quiet,
	}
}

// MessageSent outputs the message sent confirmation
func (f *Formatter) MessageSent(channel, sessionID, messageID string, msgLen int, timeout time.Duration, waiting bool) {
	if f.quiet {
		fmt.Printf("Session: %s\n", sessionID)
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event":              "message_sent",
			"status":             200,
			"channel":            channel,
			"session":            sessionID,
			"message_id":         messageID,
			"message_length":     msgLen,
			"timeout":            timeout.String(),
			"sent_at":            time.Now().UTC().Format(time.RFC3339),
			"awaiting_response":  waiting,
		})
	case config.FormatHuman:
		fmt.Printf("200 OK - Message sent via %s\n", channel)
		fmt.Printf("Session: %s\n", sessionID)
	default: // FormatLLM
		fmt.Println("═══ AFK MESSAGE SENT ═══")
		fmt.Println("Status: 200 OK")
		fmt.Printf("Channel: %s\n", channel)
		fmt.Printf("Session: %s\n", sessionID)
		fmt.Printf("Sent: %s\n", time.Now().UTC().Format(time.RFC3339))
		if waiting {
			fmt.Printf("Timeout: %s\n", formatDuration(timeout))
		}
		fmt.Printf("Message-Length: %d chars\n", msgLen)
		if waiting {
			fmt.Println()
			fmt.Println("Awaiting response...")
		}
	}
}

// Waiting outputs the waiting/reminder message
func (f *Formatter) Waiting(sessionID string, elapsed, remaining time.Duration) {
	if f.quiet {
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event":     "waiting",
			"session":   sessionID,
			"elapsed":   elapsed.String(),
			"remaining": remaining.String(),
		})
	case config.FormatHuman:
		fmt.Printf("\n[%s elapsed, %s remaining] Still waiting for response...\n",
			formatDuration(elapsed), formatDuration(remaining))
	default: // FormatLLM
		fmt.Println()
		fmt.Println("═══ AFK WAITING ═══")
		fmt.Printf("Session: %s\n", sessionID)
		fmt.Printf("Elapsed: %s | Remaining: %s\n", formatDuration(elapsed), formatDuration(remaining))
		fmt.Println("Status: No response yet")
		fmt.Println()
		fmt.Println("<instruction>")
		fmt.Println("You are waiting for human input. You may continue other work")
		fmt.Println("if possible, or wait for the response.")
		fmt.Println("</instruction>")
	}
}

// Response outputs the received response
func (f *Formatter) Response(sessionID, from, channel, content string, waitTime time.Duration) {
	if f.quiet {
		fmt.Println(content)
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event":       "response",
			"session":     sessionID,
			"from":        from,
			"channel":     channel,
			"content":     content,
			"wait_time":   waitTime.String(),
			"received_at": time.Now().UTC().Format(time.RFC3339),
		})
	case config.FormatHuman:
		fmt.Println()
		fmt.Println("────────────────────────────────────────")
		fmt.Printf("[%s] Response from %s:\n", time.Now().Format("15:04:05"), from)
		fmt.Println()
		fmt.Println(content)
		fmt.Println()
		fmt.Println("────────────────────────────────────────")
		fmt.Println()
		fmt.Println("Response received. Exiting.")
	default: // FormatLLM
		fmt.Println()
		fmt.Println("═══ AFK RESPONSE ═══")
		fmt.Printf("Session: %s\n", sessionID)
		fmt.Printf("From: %s\n", from)
		fmt.Printf("Channel: %s\n", channel)
		fmt.Printf("Received: %s\n", time.Now().UTC().Format(time.RFC3339))
		fmt.Printf("Wait-Time: %s\n", formatDuration(waitTime))
		fmt.Println()
		fmt.Println("<response>")
		fmt.Println(content)
		fmt.Println("</response>")
	}
}

// Timeout outputs the timeout message
func (f *Formatter) Timeout(sessionID string, elapsed time.Duration) {
	if f.quiet {
		fmt.Fprintln(os.Stderr, "TIMEOUT")
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event":   "timeout",
			"session": sessionID,
			"elapsed": elapsed.String(),
		})
	case config.FormatHuman:
		fmt.Fprintf(os.Stderr, "\n408 Timeout: No response within %s\n", formatDuration(elapsed))
	default: // FormatLLM
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "═══ AFK TIMEOUT ═══")
		fmt.Fprintf(os.Stderr, "Session: %s\n", sessionID)
		fmt.Fprintf(os.Stderr, "Elapsed: %s\n", formatDuration(elapsed))
		fmt.Fprintln(os.Stderr, "Status: No response")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "<instruction>")
		fmt.Fprintln(os.Stderr, "Developer did not respond. Consider: retry with different")
		fmt.Fprintln(os.Stderr, "wording, proceed with your best judgment, or ask again later.")
		fmt.Fprintln(os.Stderr, "</instruction>")
	}
}

// Error outputs an error message
func (f *Formatter) Error(statusCode int, errMsg string, sessionID string) {
	if f.quiet {
		fmt.Fprintf(os.Stderr, "%d %s\n", statusCode, errMsg)
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event":   "error",
			"status":  statusCode,
			"error":   errMsg,
			"session": sessionID,
		})
	case config.FormatHuman:
		fmt.Fprintf(os.Stderr, "%d Error: %s\n", statusCode, errMsg)
	default: // FormatLLM
		fmt.Fprintln(os.Stderr, "═══ AFK ERROR ═══")
		fmt.Fprintf(os.Stderr, "Status: %d\n", statusCode)
		fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		if sessionID != "" {
			fmt.Fprintf(os.Stderr, "Session: %s\n", sessionID)
		}
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "<instruction>")
		fmt.Fprintln(os.Stderr, "Message failed to send. Run 'afk status' to diagnose.")
		fmt.Fprintln(os.Stderr, "</instruction>")
	}
}

// Cancelled outputs the cancellation message
func (f *Formatter) Cancelled() {
	if f.quiet {
		return
	}

	switch f.format {
	case config.FormatJSON:
		f.jsonOutput(map[string]interface{}{
			"event": "cancelled",
		})
	case config.FormatHuman:
		fmt.Println("\nCancelled. Exiting.")
	default: // FormatLLM
		fmt.Println()
		fmt.Println("═══ AFK CANCELLED ═══")
		fmt.Println("Status: User cancelled")
	}
}

// WaitingStart outputs the initial waiting message (before SSE connects)
func (f *Formatter) WaitingStart(timeout time.Duration) {
	if f.quiet {
		return
	}

	switch f.format {
	case config.FormatJSON:
		// No output for JSON - wait for actual events
	case config.FormatHuman:
		fmt.Println()
		fmt.Printf("Waiting for response (timeout: %s)...\n", formatDuration(timeout))
		fmt.Println("Press Ctrl+C to cancel")
		fmt.Println()
	default: // FormatLLM
		// Already included in MessageSent for LLM format
	}
}

func (f *Formatter) jsonOutput(data map[string]interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

func formatDuration(d time.Duration) string {
	if d >= time.Hour {
		hours := d / time.Hour
		mins := (d % time.Hour) / time.Minute
		if mins > 0 {
			return fmt.Sprintf("%dh%dm", hours, mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	if d >= time.Minute {
		mins := d / time.Minute
		secs := (d % time.Minute) / time.Second
		if secs > 0 {
			return fmt.Sprintf("%dm%ds", mins, secs)
		}
		return fmt.Sprintf("%dm", mins)
	}
	return fmt.Sprintf("%ds", d/time.Second)
}
