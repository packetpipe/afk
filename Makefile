.PHONY: build clean install install-system test setup-claude tidy

# Binary name
BINARY=afk

# Build the binary
build:
	@echo "Building afk..."
	@go build -o $(BINARY) ./cmd/main.go
	@echo "✓ Built: ./$(BINARY)"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f $(BINARY)
	@echo "✓ Clean complete"

# Install to user's local bin directory
install: build
	@echo "Installing afk to ~/bin..."
	@mkdir -p ~/bin
	@cp $(BINARY) ~/bin/$(BINARY)
	@chmod +x ~/bin/$(BINARY)
	@echo "✓ Installed to ~/bin/$(BINARY)"
	@echo ""
	@echo "Ensure ~/bin is in your PATH:"
	@echo '  export PATH="$$HOME/bin:$$PATH"'

# Install system-wide (requires sudo)
install-system: build
	@echo "Installing afk system-wide..."
	@sudo cp $(BINARY) /usr/local/bin/$(BINARY)
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@echo "✓ Installed to /usr/local/bin/$(BINARY)"

# Run go mod tidy
tidy:
	@echo "Running go mod tidy..."
	@go mod tidy
	@echo "✓ Dependencies tidied"

# Test the binary
test: build
	@echo "Testing afk..."
	@./$(BINARY) -h
	@echo ""
	@echo "✓ Test completed"

# Setup Claude Code integration
setup-claude:
	@echo "Setting up Claude Code integration..."
	@mkdir -p ~/.claude
	@if [ -f ~/.claude/CLAUDE.md ]; then \
		if ! grep -q "# afk - Away From Keyboard" ~/.claude/CLAUDE.md 2>/dev/null; then \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "# afk - Away From Keyboard Messenger" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "You have access to \`afk\`, a tool for contacting developers when they're away." >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "## When to Use" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "Use \`afk\` when:" >> ~/.claude/CLAUDE.md; \
			echo "- You need a decision that only the human can make" >> ~/.claude/CLAUDE.md; \
			echo "- The task is blocked waiting for human input" >> ~/.claude/CLAUDE.md; \
			echo "- You want to notify the developer of something important" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "## Usage" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "\`\`\`bash" >> ~/.claude/CLAUDE.md; \
			echo "# Send SMS and wait for response (up to 1 hour)" >> ~/.claude/CLAUDE.md; \
			echo 'afk --sms --msg "Your question here"' >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "# Send WhatsApp and wait for response" >> ~/.claude/CLAUDE.md; \
			echo 'afk --whatsapp --msg "Your question here"' >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "# Send without waiting for response" >> ~/.claude/CLAUDE.md; \
			echo 'afk --sms --msg "Notification message" --no-wait' >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "# Check status" >> ~/.claude/CLAUDE.md; \
			echo "afk status" >> ~/.claude/CLAUDE.md; \
			echo "\`\`\`" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "## Example" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "\`\`\`bash" >> ~/.claude/CLAUDE.md; \
			echo 'afk --sms --msg "Should I use JWT or session-based auth for this API?"' >> ~/.claude/CLAUDE.md; \
			echo "\`\`\`" >> ~/.claude/CLAUDE.md; \
			echo "" >> ~/.claude/CLAUDE.md; \
			echo "Then wait for the response before continuing." >> ~/.claude/CLAUDE.md; \
			echo "✓ Added afk documentation to ~/.claude/CLAUDE.md"; \
		else \
			echo "✓ afk documentation already in ~/.claude/CLAUDE.md"; \
		fi \
	else \
		echo "# afk - Away From Keyboard Messenger" > ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "You have access to \`afk\`, a tool for contacting developers when they're away." >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "## When to Use" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "Use \`afk\` when:" >> ~/.claude/CLAUDE.md; \
		echo "- You need a decision that only the human can make" >> ~/.claude/CLAUDE.md; \
		echo "- The task is blocked waiting for human input" >> ~/.claude/CLAUDE.md; \
		echo "- You want to notify the developer of something important" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "## Usage" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "\`\`\`bash" >> ~/.claude/CLAUDE.md; \
		echo "# Send SMS and wait for response (up to 1 hour)" >> ~/.claude/CLAUDE.md; \
		echo 'afk --sms --msg "Your question here"' >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "# Send WhatsApp and wait for response" >> ~/.claude/CLAUDE.md; \
		echo 'afk --whatsapp --msg "Your question here"' >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "# Send without waiting for response" >> ~/.claude/CLAUDE.md; \
		echo 'afk --sms --msg "Notification message" --no-wait' >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "# Check status" >> ~/.claude/CLAUDE.md; \
		echo "afk status" >> ~/.claude/CLAUDE.md; \
		echo "\`\`\`" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "## Example" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "\`\`\`bash" >> ~/.claude/CLAUDE.md; \
		echo 'afk --sms --msg "Should I use JWT or session-based auth for this API?"' >> ~/.claude/CLAUDE.md; \
		echo "\`\`\`" >> ~/.claude/CLAUDE.md; \
		echo "" >> ~/.claude/CLAUDE.md; \
		echo "Then wait for the response before continuing." >> ~/.claude/CLAUDE.md; \
		echo "✓ Created ~/.claude/CLAUDE.md with afk documentation"; \
	fi
	@if [ -f ~/.claude/settings.json ]; then \
		if ! grep -q 'Bash(afk:\*)' ~/.claude/settings.json 2>/dev/null; then \
			echo "⚠ Warning: Add \"Bash(afk:*)\" to the allow list in ~/.claude/settings.json"; \
		else \
			echo "✓ afk permission already in ~/.claude/settings.json"; \
		fi \
	else \
		echo '{"permissions":{"allow":["Bash(afk:*)"],"deny":[]}}' > ~/.claude/settings.json; \
		echo "✓ Created ~/.claude/settings.json with afk permissions"; \
	fi
	@echo ""
	@echo "Setup complete! Run 'afk login' to configure your API key."
