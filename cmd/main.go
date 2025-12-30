package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/davedotdev/afk/internal/api"
	"github.com/davedotdev/afk/internal/config"
	"github.com/davedotdev/afk/internal/output"
	"github.com/davedotdev/afk/internal/sse"
)

// Version information
const version = "1.0.0"

// Exit codes
const (
	exitSuccess    = 0
	exitBadArgs    = 1
	exitAPIError   = 2
	exitTimeout    = 3
	exitSendFailed = 4
)

func main() {
	os.Exit(run())
}

func run() int {
	// Handle no arguments
	if len(os.Args) == 1 {
		printHelp()
		return exitSuccess
	}

	// Check for subcommands first
	switch os.Args[1] {
	case "login":
		return cmdLogin()
	case "logout":
		return cmdLogout()
	case "status":
		return cmdStatus()
	case "-h", "--help", "help":
		printHelp()
		return exitSuccess
	case "-v", "--version", "version":
		fmt.Printf("afk version %s\n", version)
		return exitSuccess
	}

	// Parse flags for message sending
	smsFlag := flag.Bool("sms", false, "Send message via SMS")
	whatsappFlag := flag.Bool("whatsapp", false, "Send message via WhatsApp")
	msgFlag := flag.String("msg", "", "Message content")
	sessionFlag := flag.String("session", "", "Session ID (auto-generated if not set)")
	noWaitFlag := flag.Bool("no-wait", false, "Send message and exit without waiting")
	noHintFlag := flag.Bool("no-hint", false, "Don't append '[No reply expected]' hint (use with --no-wait)")
	timeoutFlag := flag.Duration("timeout", time.Hour, "Timeout for waiting (default: 1h)")
	reminderFlag := flag.String("reminder", "", "Reminder interval (e.g., 15m, 0 to disable)")
	formatFlag := flag.String("format", "", "Output format: llm, human, json")
	quietFlag := flag.Bool("quiet", false, "Minimal output (just response content)")
	helpFlag := flag.Bool("h", false, "Show help")
	versionFlag := flag.Bool("v", false, "Show version")

	flag.Parse()

	if *helpFlag {
		printHelp()
		return exitSuccess
	}

	if *versionFlag {
		fmt.Printf("afk version %s\n", version)
		return exitSuccess
	}

	// Validate flags
	if !*smsFlag && !*whatsappFlag {
		fmt.Fprintln(os.Stderr, "400 Bad Request: Must specify --sms or --whatsapp")
		fmt.Fprintln(os.Stderr, "Run 'afk -h' for usage")
		return exitBadArgs
	}

	if *smsFlag && *whatsappFlag {
		fmt.Fprintln(os.Stderr, "400 Bad Request: Cannot use both --sms and --whatsapp")
		return exitBadArgs
	}

	if *msgFlag == "" {
		fmt.Fprintln(os.Stderr, "400 Bad Request: --msg is required")
		return exitBadArgs
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "401 Unauthorized: %v\n", err)
		return exitBadArgs
	}

	// Override config with flags if provided
	if *formatFlag != "" {
		cfg.Format = config.OutputFormat(*formatFlag)
	}
	if *reminderFlag != "" {
		cfg.ReminderInterval = *reminderFlag
	}

	// Parse reminder interval
	var reminderInterval time.Duration
	if cfg.ReminderInterval != "0" && cfg.ReminderInterval != "" {
		reminderInterval, err = time.ParseDuration(cfg.ReminderInterval)
		if err != nil {
			fmt.Fprintf(os.Stderr, "400 Bad Request: Invalid reminder interval: %v\n", err)
			return exitBadArgs
		}
	}

	// Create output formatter
	out := output.New(cfg.Format, *quietFlag)

	// Session ID is now generated server-side for security
	// Client-provided session IDs (via --session flag) are ignored by the server
	// but we keep the flag for backwards compatibility (it has no effect)
	_ = *sessionFlag // Explicitly ignore - server generates session IDs

	// Create API client
	client := api.NewClient(cfg.APIURL, cfg.APIKey)

	// Prepare message
	message := *msgFlag

	// Unescape common shell escapes that Claude Code may add in double-quoted strings
	// These appear when shell history expansion or other escaping is applied
	message = strings.ReplaceAll(message, "\\!", "!")  // History expansion
	message = strings.ReplaceAll(message, "\\?", "?")  // Glob pattern
	message = strings.ReplaceAll(message, "\\*", "*")  // Glob pattern
	message = strings.ReplaceAll(message, "\\[", "[")  // Glob pattern
	message = strings.ReplaceAll(message, "\\]", "]")  // Glob pattern

	// Append no-reply notice if --no-wait is set (unless --no-hint)
	if *noWaitFlag && !*noHintFlag {
		message = message + "\n\n[No reply expected]"
	}

	// Send message
	var msgType string
	var sendErr error
	var resp *api.SendMessageResponse

	if *smsFlag {
		msgType = "SMS"
		resp, sendErr = client.SendSMS(message, "")
	} else {
		msgType = "WhatsApp"
		resp, sendErr = client.SendWhatsApp(message, "", cfg.SysName)
	}

	if sendErr != nil {
		out.Error(500, sendErr.Error(), "")
		return exitSendFailed
	}

	// Get message ID and session ID from server response
	// Server always generates and returns the session ID for security
	messageID := ""
	sessionID := ""
	if resp != nil {
		messageID = resp.MessageID
		sessionID = resp.SessionID
	}

	if sessionID == "" {
		out.Error(500, "Server did not return session ID", "")
		return exitSendFailed
	}

	out.MessageSent(msgType, sessionID, messageID, len(*msgFlag), *timeoutFlag, !*noWaitFlag)

	// If no-wait, we're done
	if *noWaitFlag {
		return exitSuccess
	}

	// Wait for response (human format shows extra waiting message)
	out.WaitingStart(*timeoutFlag)

	// Setup signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		out.Cancelled()
		cancel()
	}()

	// Track start time for wait duration
	startTime := time.Now()

	// Listen for response with reminders
	listener := sse.NewListener(cfg.APIURL, cfg.APIKey)

	event, err := listener.ListenWithOptions(ctx, sessionID, sse.ListenOptions{
		Timeout:          *timeoutFlag,
		ReminderInterval: reminderInterval,
		OnEvent: func(e *sse.Event) {
			// Calculate wait time
			waitTime := time.Since(startTime)
			// Determine channel from event source
			channel := "SMS"
			if e.From == "web" {
				channel = "Web"
			}
			out.Response(sessionID, e.From, channel, e.Content, waitTime)
		},
		OnReminder: func(elapsed, remaining time.Duration) {
			out.Waiting(sessionID, elapsed, remaining)
		},
	})

	if err != nil {
		if strings.Contains(err.Error(), "timeout") {
			out.Timeout(sessionID, *timeoutFlag)
			return exitTimeout
		}
		if strings.Contains(err.Error(), "cancelled") {
			return exitSuccess
		}
		out.Error(503, err.Error(), sessionID)
		return exitAPIError
	}

	_ = event // Response already printed by OnEvent callback
	return exitSuccess
}


func cmdLogin() int {
	fmt.Println("ChatBridge Login")
	fmt.Println("================")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	// Get API key
	fmt.Print("API Key: ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: API key is required")
		return exitBadArgs
	}

	// Validate API key format
	if !strings.HasPrefix(apiKey, "cb_live_") && !strings.HasPrefix(apiKey, "cb_test_") {
		fmt.Fprintln(os.Stderr, "Error: Invalid API key format (should start with cb_live_ or cb_test_)")
		return exitBadArgs
	}

	// Get API URL (with default)
	fmt.Printf("API URL [default: %s]: ", config.DefaultAPIURL)
	apiURL, _ := reader.ReadString('\n')
	apiURL = strings.TrimSpace(apiURL)
	if apiURL == "" {
		apiURL = config.DefaultAPIURL
	}

	// Get system name for WhatsApp messages
	fmt.Print("System Name (for WhatsApp, e.g., 'Claude Code') [default: AI Agent]: ")
	sysName, _ := reader.ReadString('\n')
	sysName = strings.TrimSpace(sysName)
	if sysName == "" {
		sysName = "AI Agent"
	}

	// Test connection
	fmt.Print("\nTesting connection... ")
	client := api.NewClient(apiURL, apiKey)

	_, err := client.Health()
	if err != nil {
		fmt.Printf("✗ Failed\n")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitAPIError
	}
	fmt.Println("✓ Connected")

	// Validate API key
	fmt.Print("Validating API key... ")
	if err := client.ValidateKey(); err != nil {
		fmt.Printf("✗ Failed\n")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitAPIError
	}
	fmt.Println("✓ Valid")

	// Save config
	cfg := &config.Config{
		APIKey:  apiKey,
		APIURL:  apiURL,
		SysName: sysName,
	}

	if err := config.Save(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
		return exitBadArgs
	}

	fmt.Println()
	fmt.Printf("Credentials saved to %s\n", config.Path())
	fmt.Println()
	fmt.Println("You can now use:")
	fmt.Println("  afk --sms --msg \"Your message\"")
	fmt.Println("  afk --whatsapp --msg \"Your message\"")

	return exitSuccess
}

func cmdLogout() int {
	if !config.Exists() {
		fmt.Println("Already logged out")
		return exitSuccess
	}

	if err := config.Delete(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return exitBadArgs
	}

	fmt.Println("Logged out. Credentials removed.")
	return exitSuccess
}

func cmdStatus() int {
	fmt.Println("ChatBridge Status")
	fmt.Println("=================")
	fmt.Println()

	// Check config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Credentials: %s ✗ Not configured\n", config.Path())
		fmt.Println()
		fmt.Println("Run 'afk login' to configure")
		return exitBadArgs
	}

	fmt.Printf("Credentials: %s ✓\n", config.Path())
	fmt.Printf("API: %s ", cfg.APIURL)

	// Check API health
	client := api.NewClient(cfg.APIURL, cfg.APIKey)
	_, err = client.Health()
	if err != nil {
		fmt.Println("✗ Offline")
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		return exitAPIError
	}
	fmt.Println("✓ Online")

	// Validate API key
	fmt.Print("API Key: ")
	if err := client.ValidateKey(); err != nil {
		fmt.Println("✗ Invalid")
		fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
		return exitAPIError
	}
	fmt.Printf("✓ Valid (%s...)\n", cfg.APIKey[:16])

	fmt.Println()
	fmt.Println("Ready to send messages.")

	return exitSuccess
}


func printHelp() {
	help := `afk - Away From Keyboard messenger for AI agents

Send messages to developers via WhatsApp or SMS when you need their input,
then wait for their response. Designed for AI coding agents (Claude Code,
Codex, Amp, etc.) to communicate with developers who are AFK.

USAGE:
  afk login                    # Store API credentials (run once)
  afk logout                   # Remove stored credentials
  afk status                   # Check connection and quota
  afk --sms --msg "text"       # Send SMS and wait for response
  afk --whatsapp --msg "text"  # Send WhatsApp and wait for response
  afk -v                       # Show version
  afk -h                       # Show this help

MESSAGE FLAGS:
  --sms          Send message via SMS
  --whatsapp     Send message via WhatsApp
  --msg          Message content (required with --sms or --whatsapp)
  --session      Session ID for grouping messages (auto-generated if not set)
  --no-wait      Send message and exit without waiting for response
  --no-hint      Don't append '[No reply expected]' (saves chars for SMS)
  --timeout      How long to wait for response (default: 1h, e.g., 30m, 2h)
  --reminder     Reminder interval while waiting (default: 15m, 0 to disable)
  --format       Output format: llm (default), human, json
  --quiet        Minimal output (just response content)

FOR AI AGENTS (Claude Code, Codex, Amp, etc.):
  ═══════════════════════════════════════════════════════════════════════
  USE AFK WHEN YOU NEED DEVELOPER INPUT AND THEY MAY BE AWAY
  ═══════════════════════════════════════════════════════════════════════

  IMPORTANT - CLAUDE CODE PERMISSIONS:
    To use afk without permission prompts, add to ~/.claude/settings.json:
    {
      "permissions": {
        "allow": ["Bash(afk:*)", "Bash(afk --sms:*)", "Bash(afk --whatsapp:*)"]
      }
    }
    Verify with /permissions command in Claude Code.

  SHELL QUOTING:

    Use double quotes for all messages:
      afk --whatsapp --msg "Hello! How are you?"
      afk --whatsapp --msg "Don't forget to check the logs!"

    The afk tool automatically unescapes shell escape sequences
    (\! → !, \? → ?, \* → *, \[ → [, \] → ]) so your message
    arrives correctly regardless of which shell you use.

  When to use afk:
    - You need a decision that only the developer can make
    - The task is blocked until you get human input
    - You want to notify the developer of something important

  Workflow:
    1. Send your question via SMS or WhatsApp
    2. afk will wait (up to 1 hour by default) for a response
    3. When the developer replies, you'll see their response
    4. Continue your work based on their answer

  Example - Asking for a decision:
    afk --sms --msg "I found 3 approaches to implement caching:
    1. Redis (fast, needs infrastructure)
    2. In-memory (simple, loses data on restart)
    3. SQLite (persistent, slower)
    Which should I use?"

  Example - Long message (auto-creates web link for SMS > 255 chars):
    afk --sms --msg "I need your input on the database schema. Here are the options:

    Option A: Normalized (3NF)
    - Pros: Data integrity, less duplication
    - Cons: Complex joins, slower reads

    Option B: Denormalized
    - Pros: Fast reads, simple queries
    - Cons: Data duplication, update anomalies

    Which approach fits better for this use case?"

  Example - Quick notification without waiting:
    afk --whatsapp --msg "Build completed! Tests: 142 passed." --no-wait

  Tips:
    - SMS messages > 255 chars automatically become web links
    - WhatsApp is usually faster and supports richer formatting
    - Use --session to group related messages in a conversation
    - The developer can reply via the messaging app or web interface
    - Reminders output every 15m by default while waiting

  Output Format (LLM-optimized by default):
    Response content is wrapped in <response>...</response> tags
    Instructions are wrapped in <instruction>...</instruction> tags
    Use --format=json for pure JSON output if preferred

CONFIGURATION:
  Credentials stored in: ~/.afk/config.json
  Run 'afk login' to configure your API key

  Config file options:
    {
      "api_key": "cb_test_...",
      "api_url": "https://chatbridge.net",
      "sys_name": "Claude Code",
      "reminder_interval": "15m",
      "format": "llm"
    }

  sys_name: Identifies the AI agent in WhatsApp messages (e.g., "Claude Code")

EXIT CODES:
  0 - Success (message sent, response received if waiting)
  1 - Invalid arguments or configuration error
  2 - API connection failed
  3 - Timeout waiting for response
  4 - Message send failed

EXAMPLES:
  # First-time setup
  afk login

  # Send message and wait for response
  afk --sms --msg "Should I deploy to staging or production?"
  afk --whatsapp --msg "Need approval to merge PR #123" --timeout 30m

  # Send notification without waiting
  afk --whatsapp --msg "Build complete! Tests passed." --no-wait
  afk --sms --msg "Task completed: Database migration finished!" --no-wait

  # Check your connection
  afk status`

	fmt.Println(help)
}
