# Claude Code Features to Implement

## CLI Commands

### Session Management
- [x] `agentflow` - Start interactive session
- [ ] `agentflow "query"` - Start with initial prompt
- [x] `agentflow run "query"` - Non-interactive mode
- [ ] `cat file | agentflow -p "query"` - Process piped content
- [ ] `agentflow -c` / `--continue` - Continue most recent conversation
- [ ] `agentflow -r "<session>"` / `--resume` - Resume session by ID/name
- [ ] `agentflow --fork-session` - Fork when resuming

### CLI Flags
- [ ] `--model` - Set model for session
- [ ] `--verbose` - Verbose logging
- [ ] `--max-turns` - Limit agentic turns
- [ ] `--max-budget-usd` - Budget limit
- [ ] `--system-prompt` - Replace system prompt
- [ ] `--append-system-prompt` - Append to system prompt
- [ ] `--tools` - Restrict available tools
- [ ] `--allowed-tools` - Whitelist tools
- [ ] `--disallowed-tools` - Blacklist tools
- [ ] `--agents` - Define subagents via JSON
- [ ] `--permission-mode` - Plan/Auto-accept/Normal
- [ ] `--worktree` / `-w` - Git worktree isolation
- [ ] `--output-format` - text/json/stream-json
- [ ] `--json-schema` - Structured output

### Auth & Config
- [ ] `agentflow auth login`
- [ ] `agentflow auth logout`
- [ ] `agentflow auth status`
- [ ] `agentflow update`
- [ ] `agentflow agents` - List subagents
- [ ] `agentflow mcp` - MCP config

## Slash Commands (Interactive)

### Session
- [x] `/help` - Show help
- [x] `/exit`, `/quit` - Exit
- [x] `/clear` - Clear history
- [ ] `/compact [instructions]` - Compact conversation
- [ ] `/resume [session]` - Resume session
- [ ] `/rename [name]` - Rename session
- [ ] `/rewind` - Rewind conversation/code
- [ ] `/export [filename]` - Export conversation

### Display & Config
- [x] `/model [name]` - Show/change model
- [x] `/provider [name]` - Show/change provider
- [x] `/status` - Session stats
- [ ] `/config` - Open settings UI
- [ ] `/context` - Visualize context usage (colored grid)
- [ ] `/cost` - Token usage stats
- [ ] `/theme` - Change color theme
- [ ] `/vim` - Enable vim mode

### Tools & Tasks
- [x] `/skills` - List skills
- [ ] `/commands` - List all commands
- [ ] `/tasks` - Background tasks
- [ ] `/todos` - TODO list
- [ ] `/doctor` - Health check

### Project
- [ ] `/init` - Initialize AGENTFLOW.md
- [ ] `/memory` - Edit memory files
- [ ] `/permissions` - View/update permissions
- [ ] `/plan` - Enter plan mode

### Clipboard
- [ ] `/copy` - Copy last response

## Keyboard Shortcuts

### General
- [x] `Ctrl+C` - Cancel/Exit
- [ ] `Ctrl+D` - Exit (EOF)
- [x] `Ctrl+L` - Clear screen
- [ ] `Ctrl+O` - Toggle verbose output
- [ ] `Ctrl+R` - Reverse search history
- [ ] `Ctrl+B` - Background running tasks
- [ ] `Ctrl+T` - Toggle task list
- [ ] `Ctrl+G` - Open in text editor

### Navigation
- [x] `Up/Down` - Navigate history
- [x] `PgUp/PgDown` - Scroll viewport
- [ ] `Left/Right` - Cycle dialog tabs
- [ ] `Esc+Esc` - Rewind/summarize

### Mode Switching
- [ ] `Shift+Tab` - Toggle permission modes
- [ ] `Option+P` - Switch model
- [ ] `Option+T` - Toggle extended thinking

### Input
- [ ] `Option+Enter` - Multiline input
- [ ] `Shift+Enter` - Multiline (iTerm2/WezTerm)
- [ ] `\+Enter` - Multiline (universal)
- [ ] `Ctrl+V` - Paste image

### Vim Mode
- [ ] Full vim keybindings (hjkl, w, b, dd, yy, etc.)
- [ ] Text objects (iw, aw, i", a", etc.)
- [ ] Mode indicators (NORMAL, INSERT)

## Features

### Session Persistence
- [ ] Auto-save sessions to disk
- [ ] Resume by ID or name
- [ ] Session picker UI
- [ ] Fork sessions
- [ ] Per-directory session history

### Background Tasks
- [ ] Run commands in background
- [ ] Task list with progress
- [ ] Task output retrieval
- [ ] Kill background tasks
- [ ] Persist task state

### Context Management
- [ ] Context visualization (colored grid)
- [ ] Token counting
- [ ] Auto-compaction
- [ ] Manual compaction with focus
- [ ] Cost tracking

### Permission Modes
- [ ] Normal mode (ask for each action)
- [ ] Plan mode (show plan, no execution)
- [ ] Auto-accept mode (yolo)
- [ ] Per-tool permissions

### Project Integration
- [ ] AGENTFLOW.md initialization
- [ ] Memory files
- [ ] Git worktree support
- [ ] PR review status display
- [ ] Branch detection

### Input Features
- [ ] Command history with persistence
- [ ] Reverse search (Ctrl+R)
- [ ] Autocomplete for file paths (@)
- [ ] Autocomplete for commands (/)
- [ ] Bash mode (! prefix)
- [ ] Multiline input
- [ ] Image paste from clipboard
- [ ] Prompt suggestions

### Output Features
- [ ] Syntax highlighting
- [ ] Markdown rendering
- [ ] Streaming responses
- [ ] Spinner/progress indicators
- [ ] Verbose mode toggle
- [ ] Export to file/clipboard

### Themes
- [ ] Multiple color themes
- [ ] Syntax highlighting toggle
- [ ] Customizable colors

### MCP Integration
- [ ] MCP server configuration
- [ ] MCP prompts as commands
- [ ] OAuth authentication

### Subagents
- [ ] Define via CLI flags
- [ ] Define via config
- [ ] Per-agent tools/permissions
- [ ] Agent teams

## Status Bar Components
- [x] Provider/Model display
- [x] Session duration
- [x] Message count
- [x] Active skill indicator
- [x] Streaming indicator
- [ ] Token count
- [ ] Cost estimate
- [ ] PR status (if applicable)
- [ ] Permission mode indicator
- [ ] Background tasks count

## Priority Implementation Order

### Phase 1: Core UX (Week 1)
1. Session persistence (save/resume)
2. Command history with persistence
3. Multiline input
4. Continue flag (-c)
5. Resume flag (-r)
6. /compact command
7. /export command

### Phase 2: Context & Costs (Week 2)
1. Token counting
2. /cost command
3. /context visualization
4. Auto-compaction
5. Max budget flag

### Phase 3: Productivity (Week 3)
1. Vim mode
2. Reverse search (Ctrl+R)
3. Autocomplete (files, commands)
4. Prompt suggestions
5. /copy command
6. Themes

### Phase 4: Advanced (Week 4)
1. Background tasks
2. Permission modes
3. Git worktree support
4. /init and memory files
5. MCP integration

### Phase 5: Polish (Week 5)
1. /doctor health check
2. PR status display
3. Image paste
4. Subagent management
5. Full documentation
