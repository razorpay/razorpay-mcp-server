# MCP Configuration for SCA Remediation

DevRev MCP server configuration validated through production use. Only DevRev MCP is required — DevRev SCA tickets already contain all Semgrep finding data.

---

## Recommended .mcp.json

Place this file in the project root (or `~/.config/claude/mcp.json` for user-level):

```json
{
  "mcpServers": {
    "devrev": {
      "type": "streamable-http",
      "url": "https://api.devrev.ai/mcp/v1",
      "headers": {
        "Authorization": "Bearer ${DEVREV_PAT}"
      }
    }
  }
}
```

Replace `${DEVREV_PAT}` with the actual DevRev Personal Access Token.

---

## Why streamable-http (Not npx)

The `npx @anthropic-ai/devrev-mcp-server` local command frequently fails to connect. The remote `streamable-http` transport is more reliable:

| Transport | Reliability | Setup |
|-----------|------------|-------|
| `npx` (local command) | Unreliable, frequent connection failures | `command` + `args` + `env` |
| `streamable-http` (remote) | Reliable, production-tested | `type` + `url` + `headers` |

### Incorrect Configuration (Do Not Use)

```json
{
  "devrev": {
    "command": "npx",
    "args": ["-y", "@anthropic-ai/devrev-mcp-server"],
    "env": {
      "DEVREV_API_TOKEN": "..."
    }
  }
}
```

This starts a local process that may fail silently. Use the `streamable-http` configuration above.

---

## DevRev PAT Token

### Obtaining a Token

1. Log in to DevRev at `https://app.devrev.ai`
2. Navigate to Settings > API Tokens
3. Create a Personal Access Token (PAT) with read/write access to work items

### Token Placement

**NEVER commit tokens to version control.** Options:

1. **Environment variable** (recommended):
   ```bash
   export DEVREV_PAT="eyJhbG..."
   ```
   Then reference in `.mcp.json`:
   ```json
   "Authorization": "Bearer ${DEVREV_PAT}"
   ```

2. **Direct in .mcp.json** (for personal use only):
   ```json
   "Authorization": "Bearer eyJhbG..."
   ```
   Ensure `.mcp.json` is in `.gitignore`.

---

## Verifying Connectivity

### Via DevRev MCP

Use a simple works list query through MCP tools.

### Via Direct API (Fallback)

If DevRev MCP is unavailable, all operations can be performed via direct `curl` calls:

```bash
# List works
curl -X POST "https://api.devrev.ai/works.list" \
  -H "Authorization: Bearer ${DEVREV_PAT}" \
  -H "Content-Type: application/json" \
  -d '{"limit": 1, "type": ["issue"]}'
```

Expected: Returns a JSON response with at least one work item.

See [devrev-ticket-fetching.md](devrev-ticket-fetching.md) for complete API usage.

---

## .gitignore Entry

Always ensure `.mcp.json` is ignored if it contains tokens:

```gitignore
# MCP server configuration (may contain tokens)
.mcp.json
```
