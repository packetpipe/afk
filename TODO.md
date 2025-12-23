# TODO

## Pending Upstream Fixes

### Claude Code `$'...'` Pattern Matching Bug

**Issue:** Claude Code's allowed tools pattern matching (e.g., `Bash(afk:*)`) doesn't recognize ANSI-C quoting syntax (`$'...'`). Commands using this syntax still prompt for permission.

**Workaround implemented:** `afk` now converts `\!` back to `!` in messages, allowing Claude Code users to use double quotes without seeing escaped exclamation marks in their messages.

**Upstream issue:** https://github.com/anthropics/claude-code/issues/XXXX (update when filed)

**When fixed:**
- [ ] Update README.md to recommend `$'...'` syntax for Claude Code
- [ ] Update CLAUDE.md template in README to use `$'...'` syntax
- [ ] Consider removing `\!` unescaping (or keep for backwards compatibility)
