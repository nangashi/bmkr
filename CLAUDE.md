# bmkr

## Development Environment

Docker container with AI development tools (Claude Code, Codex CLI, TAKT).

### Setup

1. Configure 1Password CLI references in `.mise.toml` `[env]` section to match your vault structure
2. `mise run build` — Build the container
3. `mise run up` — Start the container (1Password resolves API keys)
4. `mise run claude` / `mise run codex` / `mise run takt` — Launch AI tools

### Available Tasks

- `mise run build` — Build dev container
- `mise run up` — Start dev container (resolves 1Password secrets)
- `mise run down` — Stop dev container
- `mise run shell` — Open bash shell
- `mise run claude` — Start Claude Code
- `mise run codex` — Start Codex CLI
- `mise run takt` — Start TAKT
