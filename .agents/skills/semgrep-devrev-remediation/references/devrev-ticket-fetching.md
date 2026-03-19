# DevRev Ticket Fetching for SCA Remediation

How to retrieve, parse, and group DevRev SCA security tickets for dependency remediation.

---

## Prerequisites

- DevRev MCP server configured OR DevRev PAT token for direct API calls (see [mcp-config.md](mcp-config.md))
- Knowledge of the DevRev "part" (capability/product) associated with the target repository

## Step 1: Identify the DevRev Part for the Repository

DevRev organizes work items under "parts" (Products > Capabilities). Each repository is typically associated with a capability.

### Using DevRev API

```bash
# Search for parts matching the repository name
curl -X POST "https://api.devrev.ai/parts.list" \
  -H "Authorization: Bearer ${DEVREV_PAT}" \
  -H "Content-Type: application/json" \
  -d '{"limit": 50}' | \
  python3 -c "
import sys, json
data = json.load(sys.stdin)
for part in data.get('parts', []):
    name = part.get('name', '')
    if 'KEYWORD' in name.lower():
        print(f'{part[\"display_id\"]}: {name} (type: {part[\"type\"]})')
"
```

Replace `KEYWORD` with the repository name or a related term. Note the `display_id` (e.g., `CAPL-41`) and the full DON ID (e.g., `don:core:dvrv-in-1:devo/2sRI6Hepzz:capability/41`).

### Example: razorpay/indie

```
Repository: razorpay/indie
Capability: "Loyalty" (CAPL-41)
Product: "Engage" (PROD-11)
DON ID: don:core:dvrv-in-1:devo/2sRI6Hepzz:capability/41
```

## Step 2: Fetch Tickets with SemgrepIssues Tag

SCA tickets created by Semgrep integration carry a `SemgrepIssues` tag.

```bash
curl -X POST "https://api.devrev.ai/works.list" \
  -H "Authorization: Bearer ${DEVREV_PAT}" \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 50,
    "type": ["issue"],
    "applies_to_part": ["don:core:dvrv-in-1:devo/2sRI6Hepzz:capability/41"],
    "tags": ["don:core:dvrv-in-1:devo/2sRI6Hepzz:tag/3458"]
  }'
```

### Pagination

If the response has `next_cursor`, fetch the next page:

```bash
curl -X POST "https://api.devrev.ai/works.list" \
  -H "Authorization: Bearer ${DEVREV_PAT}" \
  -H "Content-Type: application/json" \
  -d '{
    "limit": 50,
    "type": ["issue"],
    "applies_to_part": ["..."],
    "cursor": "NEXT_CURSOR_VALUE"
  }'
```

## Step 3: Extract Ticket Details

Each ticket contains critical fields for remediation:

### Key Fields

| Field Path | Content | Example |
|------------|---------|---------|
| `display_id` | Ticket ID | `ISS-1568953` |
| `title` | Vulnerability summary | `sca_vuln: @babel/traverse...` |
| `body` | Full description with remediation | Markdown body with CVE, fix version |
| `custom_fields.ctype__issue_url` | Semgrep finding URL | `https://semgrep.dev/...` |
| `custom_fields.ctype__triggered_rule_name` | Semgrep rule | `ssc-...` |
| `custom_fields.ctype__remediation_guide` | Fix guidance | Version to upgrade to |
| `priority` | P0/P1/P2/P3 | `p0` |
| `severity` | Critical/High/Med/Low | `critical` |

### Fetching a Single Ticket with Full Body

```bash
curl -X GET "https://api.devrev.ai/works.get?id=don:core:dvrv-in-1:devo/2sRI6Hepzz:issue/1568953" \
  -H "Authorization: Bearer ${DEVREV_PAT}" | \
  python3 -c "
import sys, json
data = json.load(sys.stdin)
work = data.get('work', {})
print(f'Title: {work.get(\"title\", \"\")}')
print(f'Body: {work.get(\"body\", \"\")[:500]}')
cf = work.get('custom_fields', {})
for k, v in cf.items():
    print(f'  {k}: {v}')
"
```

## Step 4: Group Tickets by Vulnerable Dependency

### Parsing Logic

Extract the package name from the ticket title. SCA ticket titles follow a pattern:

```
sca_vuln: {package_name} reachable from {parent_package} [via {path}]
```

### Grouping Script

```python
import json
from collections import defaultdict

# Assume `tickets` is the list of fetched work items
groups = defaultdict(list)

for ticket in tickets:
    title = ticket.get('title', '')
    # Extract package name - typically the second token after "sca_vuln:"
    if 'sca_vuln:' in title:
        parts = title.split('sca_vuln:')[1].strip().split(' ')
        pkg_name = parts[0]
        # Normalize: group by the root vulnerable package
        groups[pkg_name].append({
            'id': ticket['display_id'],
            'title': title,
            'priority': ticket.get('priority', 'unknown'),
            'body': ticket.get('body', ''),
            'custom_fields': ticket.get('custom_fields', {})
        })

# Output grouped summary
for pkg, items in sorted(groups.items(), key=lambda x: len(x[1]), reverse=True):
    print(f"\n{pkg}: {len(items)} tickets")
    for item in items:
        print(f"  - {item['id']}: {item['title'][:80]}")
```

## Step 5: Build the Remediation Queue

After grouping, prioritize by:

1. **Priority** (P0 first)
2. **Ticket count** (more tickets = wider impact)
3. **Vulnerability type** (RCE > XSS > DoS > Info)

### Example Output

```
Group A: @babel/traverse (6 tickets, P0 Critical)
  - ISS-1568953: sca_vuln: @babel/traverse reachable from babel-plugin-polyfill-regenerator
  - ISS-1568952: sca_vuln: @babel/traverse reachable from babel-plugin-polyfill-corejs2
  - ISS-1568951: sca_vuln: @babel/traverse reachable from babel-plugin-polyfill-corejs3
  - ISS-1568950: sca_vuln: @babel/traverse reachable from @babel/preset-env
  - ISS-1568949: sca_vuln: @babel/traverse reachable from @babel/helper-define-polyfill-provider
  - ISS-1568948: sca_vuln: @babel/traverse reachable from @babel/plugin-transform-runtime

Group B: @angular/compiler (124 tickets, P1 High)
  ...
```

## Troubleshooting

### DevRev MCP Connection Fails

The `npx @anthropic-ai/devrev-mcp-server` command may fail to connect. Use the remote `streamable-http` transport instead:

```json
{
  "devrev": {
    "type": "streamable-http",
    "url": "https://api.devrev.ai/mcp/v1",
    "headers": {
      "Authorization": "Bearer ${DEVREV_PAT}"
    }
  }
}
```

See [mcp-config.md](mcp-config.md) for the complete configuration.

### Finding the SemgrepIssues Tag ID

If you do not know the tag DON ID:

```bash
curl -X POST "https://api.devrev.ai/tags.list" \
  -H "Authorization: Bearer ${DEVREV_PAT}" \
  -H "Content-Type: application/json" \
  -d '{"limit": 100}' | \
  python3 -c "
import sys, json
for t in json.load(sys.stdin).get('tags', []):
    if 'semgrep' in t.get('name', '').lower():
        print(f'{t[\"id\"]}: {t[\"name\"]}')"
```

