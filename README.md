# afk - Away From Keyboard Messenger

Send messages to developers via WhatsApp or SMS when you need their input, then wait for their response. Designed for AI coding agents (Claude Code, Codex, Amp, etc.) to communicate with developers who are AFK.

## Installation

### Option 1: Download Binary (Recommended)

Download the latest binary for your platform from the [releases page](https://github.com/packetpipe/afk/releases).

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/packetpipe/afk/releases/latest/download/afk-darwin-arm64 -o afk
chmod +x afk
sudo mv afk /usr/local/bin/
```

**macOS (Intel):**
```bash
curl -L https://github.com/packetpipe/afk/releases/latest/download/afk-darwin-amd64 -o afk
chmod +x afk
sudo mv afk /usr/local/bin/
```

**Linux (x64):**
```bash
curl -L https://github.com/packetpipe/afk/releases/latest/download/afk-linux-amd64 -o afk
chmod +x afk
sudo mv afk /usr/local/bin/
```

**Linux (ARM64):**
```bash
curl -L https://github.com/packetpipe/afk/releases/latest/download/afk-linux-arm64 -o afk
chmod +x afk
sudo mv afk /usr/local/bin/
```

**Windows (x64):**
```powershell
# Download to current directory
Invoke-WebRequest -Uri "https://github.com/packetpipe/afk/releases/latest/download/afk-windows-amd64.exe" -OutFile "afk.exe"

# Move to a directory in your PATH, or add current directory to PATH
```

**Windows (ARM64):**
```powershell
Invoke-WebRequest -Uri "https://github.com/packetpipe/afk/releases/latest/download/afk-windows-arm64.exe" -OutFile "afk.exe"
```

### Option 2: Build from Source

Requires Go 1.21 or later.

```bash
# Clone the repository
git clone https://github.com/packetpipe/afk.git
cd afk

# Build
go build -o afk ./cmd/main.go

# Install to PATH
sudo mv afk /usr/local/bin/
```

### Verify Installation

```bash
afk -v
```

## Quick Start

### 1. Sign up at ChatBridge

Visit [chatbridge.net](https://chatbridge.net) to create an account and get your API key.

### 2. Login

```bash
afk login
```

Enter your API key and optionally set a system name (e.g., "Claude Code") that appears in WhatsApp messages.

### 3. Send a Message

```bash
# Send via WhatsApp (recommended)
afk --whatsapp --msg "Should I use Redis or PostgreSQL for caching?"

# Send via SMS
afk --sms --msg "Build complete. Deploy to staging?"
```

### 4. Wait for Response

afk will wait (default: 1 hour) for the developer to reply. When they do, the response is printed and afk exits.

## Usage

```
afk --whatsapp --msg "Your message"    # Send via WhatsApp
afk --sms --msg "Your message"         # Send via SMS
afk --whatsapp --msg "Done" --no-wait  # Send without waiting for reply
afk --sms --msg "Done" --no-wait --no-hint  # No wait, no hint (saves SMS chars)
afk --whatsapp --msg "Question?" --timeout 30m  # Custom timeout
afk status                              # Check connection
afk logout                              # Remove credentials
```

### Flags

| Flag | Description |
|------|-------------|
| `--sms` | Send message via SMS |
| `--whatsapp` | Send message via WhatsApp |
| `--msg` | Message content (required) |
| `--session` | Session ID for grouping messages (auto-generated if not set) |
| `--no-wait` | Send message and exit without waiting for response |
| `--no-hint` | Don't append '[No reply expected]' hint (saves 22 chars for SMS) |
| `--timeout` | How long to wait for response (default: 1h, e.g., 30m, 2h) |
| `--reminder` | Reminder interval while waiting (default: 15m, 0 to disable) |
| `--format` | Output format: llm (default), human, json |
| `--quiet` | Minimal output (just response content) |

## Setting Up for Claude Code

To let Claude Code use afk to contact you, add the following to your `~/.claude/CLAUDE.md` file:

````markdown
# AFK Tool - Contact Developer When Away

When you need developer input and they may be away from their keyboard, use the `afk` command-line tool to reach them via WhatsApp or SMS.

## When to Use afk

Use afk when:
- You need a decision only the developer can make
- The task is blocked until you get human input
- You want to notify the developer of something important
- You've completed a significant task and want confirmation

## How to Use

Send a message and wait for response:
```bash
afk --whatsapp --msg "I found 3 approaches to implement caching:
1. Redis (fast, needs infrastructure)
2. In-memory (simple, loses data on restart)
3. SQLite (persistent, slower)
Which should I use?"
```

Send a notification without waiting:
```bash
afk --whatsapp --msg "Build completed. Tests: 142 passed, 0 failed." --no-wait
```

Send SMS notification without the '[No reply expected]' hint (saves characters):
```bash
afk --sms --msg "Build complete. 142 passed." --no-wait --no-hint
```

## Output Format

Responses are wrapped in `<response>` tags for easy parsing:
```
<response>
Use Redis, we already have it set up for sessions
</response>
```

## Tips

- Keep questions clear and concise
- Provide options when asking for decisions
- Use --no-wait for notifications that don't need a reply
- Use --no-hint with --no-wait for SMS to save 22 characters
- SMS messages over 255 characters become web links automatically
````

## Other AI Agents

### Codex / OpenAI

Add to your system prompt or instructions:

```
You have access to the `afk` command-line tool for contacting the developer when they're away.

Usage:
- afk --whatsapp --msg "Your question" - Send WhatsApp message and wait for reply
- afk --sms --msg "Your message" - Send SMS and wait for reply
- Add --no-wait to send without waiting for a response
- Add --no-hint with --no-wait to skip '[No reply expected]' suffix (saves SMS chars)
- Add --timeout 30m to set custom timeout (default: 1 hour)

Use afk when you need developer input and they may be away from their keyboard.
The response will be printed when the developer replies.
```

### Amp

Add to your Amp configuration or custom instructions:

```yaml
tools:
  - name: afk
    description: Contact developer via WhatsApp/SMS when they're AFK
    usage: |
      Send message and wait for reply:
        afk --whatsapp --msg "Your question here"

      Send notification without waiting:
        afk --whatsapp --msg "Task complete" --no-wait

      SMS notification (no hint to save chars):
        afk --sms --msg "Done" --no-wait --no-hint

      Use when you need human input and the developer may be away.
```

### Generic Integration

For any AI agent that can execute shell commands:

1. Ensure `afk` is in the system PATH
2. Run `afk login` to configure credentials
3. Instruct the agent to use `afk --whatsapp --msg "..."` when it needs developer input

## Configuration

Config is stored in `~/.afk/config.json`:

```json
{
  "api_key": "cb_live_...",
  "api_url": "https://chatbridge.net",
  "sys_name": "Claude Code",
  "reminder_interval": "15m",
  "format": "llm"
}
```

Options:
- `sys_name`: Identifies the AI agent in WhatsApp messages
- `reminder_interval`: How often to show waiting reminders (default: 15m, set to "0" to disable)
- `format`: Output format - "llm" (default), "human", or "json"

## Exit Codes

- `0` - Success (message sent, response received if waiting)
- `1` - Invalid arguments or configuration error
- `2` - API connection failed
- `3` - Timeout waiting for response
- `4` - Message send failed

## Links

- **GitHub:** [github.com/packetpipe/afk](https://github.com/packetpipe/afk)
- **ChatBridge:** [chatbridge.net](https://chatbridge.net)
- **Support:** support@chatbridge.net

## License

MIT License - See LICENSE file for details.
