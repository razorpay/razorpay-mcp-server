---
name: razorpay-api-review
description: Comprehensive API review skill for Razorpay API council submissions. Reads Google Docs URLs, validates API design compliance, checks REST principles, naming conventions, error handling, URL structures, and assesses merchant-facing API documentation. Ensures adherence to Razorpay API Design Guide v1.0 standards with mandatory reviewer approval checks.
---

# Razorpay API Council Review Skill

## Overview
This skill validates API council documents against Razorpay's **API Design Guide v1.0**, ensuring compliance with REST principles, naming conventions, error handling, and documentation quality.

## Document Access & Reading

### Reading Google Docs
API Council documents are typically shared as Google Docs. Follow these steps to access them:

#### Step 1: Load MCP Tools
**MANDATORY**: Before reading any Google Doc, you MUST first load the MCP tool:

```
Use MCPSearch tool with:
query: "select:mcp__plugin_compass_google-workspace__get_doc_content"
max_results: 5
```

#### Step 2: Extract Document ID from URL
Google Docs URLs follow this pattern:
```
https://docs.google.com/document/d/{DOCUMENT_ID}/edit...
```

**Examples:**
- URL: `https://docs.google.com/document/d/1txEcOX2rlAldDCTwV_MvKCGCCFzZ-vlKZuXtdx2Clqw/edit?tab=t.0`
- Document ID: `1txEcOX2rlAldDCTwV_MvKCGCCFzZ-vlKZuXtdx2Clqw`

- URL: `https://docs.google.com/document/d/1ABC123xyz/edit`
- Document ID: `1ABC123xyz`

#### Step 3: Read Document Content
Use the MCP tool to fetch the document:

```
mcp__plugin_compass_google-workspace__get_doc_content
Parameters:
  document_id: "{EXTRACTED_DOCUMENT_ID}"
```

#### Step 4: Proceed with Review
Once document content is loaded, proceed with the comprehensive API review using the validation framework below.

**Important Notes:**
- Always load MCP tools BEFORE attempting to read Google Docs
- If document has multiple tabs, all content will be retrieved
- Document metadata (title, link, type) will be included in response
- If access fails, verify document sharing permissions

**Workflow Example:**
```
1. User shares: "Review this doc: https://docs.google.com/document/d/1ABC123/edit"
2. Load MCP tool via MCPSearch
3. Extract document ID: "1ABC123"
4. Call mcp__plugin_compass_google-workspace__get_doc_content
5. Analyze content against API Design Guide
6. Provide comprehensive review
```

### Handling Different Input Formats

**Google Docs URL** (Preferred):
- Direct URL: `https://docs.google.com/document/d/{ID}/edit`
- Extract ID and use MCP tool to read content

**Markdown Files** (Local):
- Use Read tool to access local .md files
- File path must be provided by user

**Other Formats**:
- If user provides a different format, request a Google Docs link or file path
- API Council documents should be in Google Docs for collaboration

### Troubleshooting Document Access

**Common Issues:**

1. **"Access Denied" Error**:
   - Document may not be shared with the appropriate Google account
   - Ask user to verify sharing permissions
   - Document should be accessible to anyone with link (at minimum)

2. **"Document Not Found"**:
   - Verify document ID is correctly extracted
   - Check if URL is complete and valid
   - Ensure document hasn't been deleted

3. **Empty Content**:
   - Document may be genuinely empty
   - Check if user has edit/comment access
   - May indicate permissions issue

4. **Multiple Tabs**:
   - Google Docs with multiple tabs will return all content
   - Content will be organized by tab name
   - Look for section headers like "--- TAB: [Tab Name] ---"

## Core Validation Framework

### 1. REST Principles Compliance

#### HTTP Methods (Section 3.2)
**Correct Usage:**
- `GET` - Fetch/read resources (safe & idempotent)
- `POST` - Create resources or perform actions
- `PATCH` - Partial updates
- `DELETE` - Remove resources
- ❌ **NEVER use `PUT`** - Razorpay policy: no wholesale replacement

**Examples:**
```
✅ GET /payments
✅ GET /payments/{payment_id}
✅ POST /refunds
✅ PATCH /subscriptions/{id}
✅ POST /subscriptions/{id}/cancel

❌ GET /getPayments (verb in URL)
❌ PUT /payments/{id} (PUT not allowed)
❌ /payment (should be plural)
```

#### HTTP Status Codes (Section 3.3)
- **2xx** - Success (200, 201, 204)
- **4xx** - Client errors (400, 401, 403, 404, 422)
- **5xx** - Server errors (500, 502, 503)

### 2. URL Design Validation

#### Entity URLs (Section 4.4.1)
**Rules:**
1. Always use plurals: `/payments` not `/payment`
2. No verbs in URLs
3. Use `{entity_id}` for path parameters

```
✅ GET /payments
✅ GET /payments/{payment_id}
✅ GET /refunds?payment_id={id}

❌ /payment (not plural)
❌ /getPayments (has verb)
❌ /payments/:id/refunds (incorrect nesting)
```

#### Action URLs (Section 4.4.2)
**Format:** `/entity/{entity_id}/action`

- Always use POST method
- Single-word verb only
- No compound verbs

```
✅ POST /subscriptions/{id}/cancel
✅ POST /payments/{id}/capture
✅ POST /virtual_accounts/{id}/close

❌ POST /subscriptions/{id}/cancelled (wrong tense)
❌ POST /subscriptions/{id}/pleaseCancel (too verbose)
```

#### Related Entities (Section 4.4.4)
Use query parameters for filtering, not nested paths:
```
✅ GET /refunds?payment_id={payment_id}
❌ GET /payments/{payment_id}/refunds
```

### 3. Naming Conventions

#### Attribute Naming (Section 4.3)
**Principles:**
1. Simple, clear names
2. Full words - no abbreviations
3. snake_case with underscores
4. Never use "razorpay" or "rzp" prefix

```
✅ transactions (not txnList, allTxns)
✅ payment_link (not pl)
✅ customer_id (not cust_id)
✅ reference_id (not ref)

❌ txn, txns → use 'transaction', 'transactions'
❌ pmnt → use 'payment'
❌ rzp_payment_id → use 'payment_id'
```

#### Acronyms (Section 4.3.3)
Only publicly accepted industry terms:
```
✅ ifsc, iin, upi (industry standard)
❌ rzp, Custom internal acronyms
```

### 4. Data Types & Formats

#### Timestamps (Section 6.1)
- Always Unix epoch integers
- Never ISO 8601 strings

```
✅ "created_at": 1609459200
❌ "created_at": "2021-01-01T00:00:00Z"
```

#### Monetary Amounts (Section 6.2)
- Always smallest currency unit (paise/cents)
- Integer only, no decimals

```
✅ "amount": 10050 (₹100.50)
❌ "amount": "100.50"
❌ "amount": 100.50
```

#### Booleans (Section 6.3)
**Input:** `true`, `false`, `0`, `1`
**Output:** Always `true` or `false`

```
✅ Response: "captured": true
❌ Response: "captured": 1
```

#### Nulls vs Empty Strings (Section 6.5)
- `null` - Not applicable
- Empty string - Empty but valid value
- Omit field - Unknown/not provided

```
✅ "middle_name": null (no middle name)
✅ "description": "" (empty but valid)
Omit "optional_field" entirely if not provided
```

### 5. Response Structure

#### Common Attributes (Section 7)
**Every entity must have:**
- `id` - Unique identifier with prefix (e.g., `pay_`, `order_`)
- `entity` - Entity type (e.g., "payment", "order")
- `created_at` - Unix timestamp

#### Arrays & Keys (Section 6.6)
Never use numbered keys:
```
❌ Bad:
{
  "item_1": {},
  "item_2": {}
}

✅ Good:
{
  "items": [{}, {}]
}
```

### 6. Error Responses

#### Structure (Section 9)
```json
{
  "error": {
    "code": "BAD_REQUEST_ERROR",
    "description": "The amount must be at least INR 1.00",
    "source": "business",
    "step": "payment_initiation",
    "reason": "amount_too_small",
    "metadata": {},
    "field": "amount"
  }
}
```

**Required fields:**
- `code` - Error type (e.g., "BAD_REQUEST_ERROR")
- `description` - Human-readable, actionable message
- `source` - "business" or "internal"
- `reason` - Machine-readable error identifier

**Internal errors:**
- Always use `"reason": "server_error"`
- Never expose internal details

### 7. Breaking Changes (Section 11)

#### Non-Breaking (✅ Safe):
- Adding new optional fields
- Adding new endpoints
- Adding new enum values (with caution)

#### Breaking (❌ Dangerous):
- Removing fields
- Renaming fields
- Changing data types
- Changing behavior
- Making optional → required

**Mitigation strategies:**
- Versioning
- Deprecation notices
- Parallel endpoints
- Backward-compatible defaults

### 8. Documentation Requirements

#### Merchant-Facing Docs Must Include:
1. **Overview** - What the API does
2. **Use Cases** - When to use it
3. **Request Format** - Complete examples
4. **Response Format** - All fields explained
5. **Error Scenarios** - Common errors & fixes
6. **Code Examples** - In multiple languages
7. **Best Practices** - How to use correctly

### 9. Review Output Format

Provide review in this structure:

**🔍 Reviewer Approval Check (MANDATORY):**
First check if the document has proper reviewer signoffs:
- **AD/D of Product** for the Business Unit - Review date and approval status filled
- **AD/D or Staff Engineer** for the Business Unit - Review date and approval status filled
- Minimum 2 council members signed off
- If reviewer information is incomplete (missing dates/status), this is a **BLOCKING ISSUE**

**✅ Strengths:**
- List what follows guidelines correctly

**❌ Issues (Priority: Critical/High/Medium/Low):**
- Issue description
- Guideline violated
- Impact on merchants
- Recommended fix

**💡 Recommendations:**
- Suggestions for improvement
- Alternative approaches

**🚨 Blocking Issues (Must Fix):**
- Missing or incomplete reviewer approvals (AD/D Product + AD/D/Staff Eng)
- Breaking changes without mitigation
- Core REST principle violations
- Naming convention violations
- Missing required fields

**📋 Approval Checklist:**
- [ ] **AD/D of Product (BU) reviewed and approved (with date)**
- [ ] **AD/D or Staff Engineer (BU) reviewed and approved (with date)**
- [ ] All blocking issues resolved
- [ ] Pre-submission docs complete
- [ ] 24-hour notice on #api_council
- [ ] Payment APIs: #payments_api_council pre-approval
- [ ] 2+ council members signed off
- [ ] Merchant docs ready

### 10. Common Mistakes

**Naming:**
```
❌ txnList → ✅ transactions
❌ payment_link_id → ✅ payment_link (if referring to ID)
❌ pl → ✅ payment_link
❌ ref → ✅ reference_id
```

**URLs:**
```
❌ /getPayments → ✅ GET /payments
❌ /payment/:id → ✅ GET /payments/{payment_id}
❌ /subscriptions/:id/cancelled → ✅ POST /subscriptions/{id}/cancel
```

**Data Types:**
```
❌ "amount": "100.50" → ✅ "amount": 10050
❌ "captured": 1 → ✅ "captured": true
❌ "created_at": "2021-01-01" → ✅ "created_at": 1609459200
```

### 11. Quick Reference

**Design Principles:**
1. Move slow and deliberate - APIs are long-lived
2. Readability first - parseable in one read
3. Backward compatibility - never break integrations
4. Merchant-centric - think from customer perspective
5. Simplicity over complexity
6. Consistency with existing patterns

**Key Rules:**
- URLs: Plurals, no verbs, path parameters
- Naming: snake_case, full words, no "razorpay"
- Data: Unix timestamps, paise amounts, true/false booleans
- Errors: Complete structure, actionable descriptions
- Breaking: Addition OK, removal/rename breaking

### 12. Process Integration

**Reviewer Approval Requirements:**
Before scheduling council review, ensure the document has been reviewed and approved by:
1. **AD/D of Product** for the Business Unit
2. **AD/D or Staff Engineer** for the Business Unit

The reviewer table in the document MUST include:
- Reviewer name and email
- Review completion date
- Approval status (Approved/Pending/Rejected)

**Example Reviewer Table:**
```markdown
| Reviewer Name | Role | Reviewed Date | Status |
|--------------|------|---------------|--------|
| John Doe | AD/D Product (Payments BU) | 2026-01-15 | Approved |
| Jane Smith | Staff Engineer (Payments BU) | 2026-01-16 | Approved |
```

❌ **Incomplete reviewer information is a blocking issue that prevents council review.**

**Meeting Preparation:**
- Calendar: https://calendar.app.google/KNs4RLmZCZ444JYy9
- One slot per week, one API per slot
- Share docs 24 hours in advance

**Required Documents:**
1. Product spec/Concept Note
2. Research & Validation notes
3. Internal council notes
4. Merchant-facing API documentation

**Approval Flow:**
1. **PM prepares documentation**
2. **AD/D of Product (BU) reviews and approves**
3. **AD/D or Staff Engineer (BU) reviews and approves**
4. Payment APIs: Pre-approval in #payments_api_council
5. Share on #api_council (24h notice)
6. Council review
7. Minimum 2 members sign-off
8. Address action items
9. Final approval

---

**Note:** Use this alongside human expertise. Consider specific product context, merchant needs, and technical constraints.
