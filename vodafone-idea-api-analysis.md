# Vodafone Idea Integration: Razorpay API Availability Analysis

**Document Version:** 1.0  
**Date:** January 2025  
**Project:** Razorpay MCP Server - Vodafone Idea Integration  

---

## Executive Summary

This document analyzes the availability of Razorpay APIs required to fulfill Vodafone Idea's agentic integration requirements. Based on comprehensive research of Razorpay's public API documentation, **only 3 out of 9 critical missing tools have available APIs**, creating significant implementation challenges for VI's use cases.

## Vodafone Idea Use Cases Overview

### Use Case 1: First-time user (specific payment method specified)
- User evaluates VI's popular recharge plans
- User selects plan and pays with saved card
- VI displays all saved cards on user's number
- User selects card, receives OTP, writes OTP to authenticate
- Payment success message displayed

### Use Case 2: First-time user (no payment method specified)
- User evaluates VI's popular recharge plans
- User selects plan without mentioning payment method
- VI nudges user to create SBMD mandate with amount block options
- User creates mandate with favorite TPAP & redirected to VI chat
- VI proceeds with SBMD payment after user confirmation

### Use Case 3: Repeat user (no payment method specified)
- User wants to recharge expired plan
- User selects plan without specifying payment method
- VI displays SBMD payment option and waits for confirmation
- Payment success message displayed

### Use Case 4: Repeat user changes payment amount (payment method specified later)
- User wants to recharge expired plan
- User initially chooses SBMD, then switches to card payment
- VI displays saved cards, user selects card and enters OTP
- Payment success message displayed

---

## Current MCP Server Status

### Available Tools (34 total)
The Razorpay MCP server currently includes:
- **Payment Management:** 5 tools
- **Order Management:** 5 tools  
- **Payment Links:** 6 tools
- **QR Code Management:** 6 tools
- **Refund Management:** 5 tools
- **Settlement Management:** 5 tools
- **Payout Management:** 2 tools

### Missing Tools for VI Integration (9 total)
1. `create_customer` - Create customer with basic details
2. `fetch_customer` - Fetch customer details by ID
3. `fetch_customer_tokens` - Fetch saved payment methods (cards, UPI)
4. `create_s2s_payment` - Initiate server-to-server payments
5. `submit_otp_payment` - Submit OTP for payment authentication
6. `create_mandate` - Create SBMD mandate for recurring debits
7. `fetch_mandate_balance` - Check available balance in mandate
8. `debit_mandate` - Execute debit from mandate balance
9. `create_order_with_transfers` - Create order with automatic transfers

---

## API Availability Analysis

### ✅ APIs with Available cURL Commands

#### 1. Create Customer ✅
**API Documentation:** https://razorpay.com/docs/api/customers/  
**Status:** Available

```bash
curl -u [YOUR_KEY_ID]:[YOUR_KEY_SECRET] \
-X POST https://api.razorpay.com/v1/customers \
-H "Content-Type: application/json" \
-d '{
    "name": "John Smith",
    "contact": "+11234567890",
    "email": "john.smith@example.com",
    "fail_existing": "0",
    "notes": {
      "notes_key_1": "Additional customer info"
    }
}'
```

**Response:**
```json
{
  "id": "cust_1Aa00000000003",
  "entity": "customer",
  "name": "John Smith",
  "email": "john.smith@example.com",
  "contact": "+11234567890",
  "gstin": null,
  "notes": {
    "notes_key_1": "Additional customer info"
  },
  "created_at": 1582033731
}
```

#### 2. Fetch Customer ✅
**API Documentation:** https://razorpay.com/docs/api/customers/  
**Status:** Available

```bash
curl -u [YOUR_KEY_ID]:[YOUR_KEY_SECRET] \
-X GET https://api.razorpay.com/v1/customers/{customer_id}
```

#### 3. Create Order with Transfers ✅
**API Documentation:** https://razorpay.com/docs/api/orders/  
**Status:** Available

```bash
curl -u [YOUR_KEY_ID]:[YOUR_KEY_SECRET] \
-X POST https://api.razorpay.com/v1/orders \
-H "Content-Type: application/json" \
-d '{
    "amount": 50000,
    "currency": "INR",
    "receipt": "receipt#1",
    "transfers": [
        {
            "account": "acc_7jO4N6LScw5CEG",
            "amount": 50000,
            "currency": "INR",
            "on_hold": 0
        }
    ]
}'
```

### ❌ APIs NOT Available (Critical Gaps)

#### 1. Fetch Customer Tokens/Saved Payment Methods ❌
**Status:** **NO DIRECT API FOUND**  
**What exists:** Customer Fund Account API (bank accounts only)  
**What's missing:** API to fetch saved cards, UPI tokens, wallet tokens  
**VI Impact:** Cannot display saved payment methods in Use Cases 1 & 4  

**Current Alternative:** Customer Fund Account API only supports bank accounts:
```bash
curl -u [YOUR_KEY_ID]:[YOUR_KEY_SECRET] \
-X GET https://api.razorpay.com/v1/fund_accounts?customer_id=cust_Aa000000000001
```

#### 2. Server-to-Server Payment Creation ❌
**Status:** **NO S2S JSON API FOUND**  
**What exists:** Hosted checkout integration only  
**What's missing:** Direct payment creation API with OTP handling  
**VI Impact:** Cannot handle OTP-based card payments without redirect  

**Current Alternative:** Only hosted checkout available:
```bash
# Only creates hosted checkout session, not direct payment
curl -X POST https://api.razorpay.com/v1/checkout/embedded
```

#### 3. OTP Generation/Submission APIs ❌
**Status:** **NO DIRECT OTP APIs FOUND**  
**What exists:** OTP handling within hosted checkout only  
**What's missing:** APIs to generate and submit OTPs independently  
**VI Impact:** Cannot implement Use Cases 1 & 4 OTP flows  

#### 4. SBMD (Standing Block Mandate Debit) Management ❌
**Status:** **NO SBMD APIs FOUND**  
**What exists:** General subscription/recurring payment APIs  
**What's missing:**
- Create SBMD mandate API
- Check SBMD balance API  
- Debit from SBMD API

**VI Impact:** Cannot implement Use Cases 2 & 3 SBMD flows

**Current Alternative:** Basic recurring payments API exists but lacks SBMD-specific features:
```bash
curl -u [YOUR_KEY_ID]:[YOUR_KEY_SECRET] \
-X POST https://api.razorpay.com/v1/subscriptions
```

---

## Implementation Impact Matrix

| **Missing Tool** | **API Available** | **cURL Exists** | **VI Use Cases Affected** | **Implementation Status** |
|------------------|-------------------|-----------------|---------------------------|---------------------------|
| ✅ Create Customer | Yes | Yes | All | ✅ **Can implement** |
| ✅ Fetch Customer | Yes | Yes | All | ✅ **Can implement** |
| ❌ Fetch Customer Tokens | No | No | UC1, UC4 | ❌ **BLOCKING** |
| ❌ Create S2S Payment | No | No | UC1, UC4 | ❌ **BLOCKING** |
| ❌ Submit OTP Payment | No | No | UC1, UC4 | ❌ **BLOCKING** |
| ❌ Create SBMD Mandate | No | No | UC2 | ❌ **BLOCKING** |
| ❌ Fetch SBMD Balance | No | No | UC2, UC3 | ❌ **BLOCKING** |
| ❌ Debit SBMD | No | No | UC2, UC3 | ❌ **BLOCKING** |
| ✅ Create Order w/ Transfers | Yes | Yes | All | ✅ **Can implement** |

### Use Case Impact Summary
- **Use Case 1:** ❌ **BLOCKED** - Missing customer tokens and S2S payment APIs
- **Use Case 2:** ❌ **BLOCKED** - Missing SBMD mandate management APIs
- **Use Case 3:** ❌ **BLOCKED** - Missing SBMD balance and debit APIs  
- **Use Case 4:** ❌ **BLOCKED** - Missing customer tokens and S2S payment APIs

---

## Critical Findings

### ✅ Available for Implementation (3/9 tools)
1. **Create Customer** - Full API support with cURL examples
2. **Fetch Customer** - Full API support with cURL examples
3. **Create Order with Transfers** - Full API support with cURL examples

### ❌ Major API Gaps (6/9 tools)
1. **Fetch Customer Tokens** - No API for saved payment methods
2. **Server-to-Server Payment Creation** - Only hosted checkout available
3. **OTP Generation/Submission** - No independent OTP APIs
4. **SBMD Mandate Creation** - No SBMD-specific APIs
5. **SBMD Balance Check** - No balance inquiry APIs
6. **SBMD Debit Execution** - No direct debit APIs

### Impact on VI Integration
- **67% of required tools lack API support**
- **All 4 use cases are currently blocked**
- **Critical payment flows cannot be implemented**

---

## Recommendations

### Phase 1: Implement Available APIs (Immediate)
**Timeline:** 1-2 weeks
- ✅ Build `create_customer` tool
- ✅ Build `fetch_customer` tool  
- ✅ Build `create_order_with_transfers` tool

**Benefits:**
- Provides foundational customer and order management
- Demonstrates MCP server capabilities
- Sets up infrastructure for future tools

### Phase 2: Escalate Missing APIs (Critical)
**Timeline:** 2-4 weeks (depends on Razorpay response)
- 📧 Contact Razorpay product team
- 📋 Request specific APIs for VI use cases:
  - Customer tokens/saved payment methods API
  - Server-to-server payment APIs with OTP handling
  - SBMD mandate management APIs
- 📄 Provide VI requirements document as justification

### Phase 3: Alternative Solutions (Interim)
**Timeline:** 2-3 weeks
- 🔄 Use Razorpay's hosted checkout for payment flows
- 📡 Implement webhook-based workarounds where possible
- 🔗 Use existing payment links as interim solution for simple flows
- 🏗️ Build framework to easily add missing tools once APIs become available

### Phase 4: Custom Integration (If needed)
**Timeline:** 4-6 weeks
- 🤝 Explore Razorpay partnership for custom API access
- 🔧 Consider building proxy layer for missing functionality
- 📊 Implement analytics to track VI integration success

---

## Technical Requirements for Missing APIs

### Fetch Customer Tokens API (Required)
```bash
# Proposed API endpoint
GET /v1/customers/{customer_id}/tokens

# Expected response
{
  "entity": "collection",
  "count": 2,
  "items": [
    {
      "id": "token_EhYXXXXXXXXX",
      "entity": "token",
      "method": "card",
      "card": {
        "last4": "1111",
        "network": "Visa",
        "type": "credit"
      },
      "created_at": 1234567890
    },
    {
      "id": "token_UpiXXXXXXXX",
      "entity": "token", 
      "method": "upi",
      "vpa": "user@upi",
      "created_at": 1234567890
    }
  ]
}
```

### Server-to-Server Payment API (Required)
```bash
# Proposed API endpoint
POST /v1/payments/s2s

# Request payload
{
  "order_id": "order_XXXXXXXX",
  "method": "card",
  "token": "token_XXXXXXXX",
  "otp_required": true
}

# Expected response
{
  "id": "pay_XXXXXXXX",
  "status": "otp_required",
  "otp_reference": "otp_ref_XXXXX"
}
```

### SBMD Mandate APIs (Required)
```bash
# Create mandate
POST /v1/mandates/sbmd
{
  "customer_id": "cust_XXXXX",
  "amount_block": 100000,
  "frequency": "as_required"
}

# Check balance
GET /v1/mandates/{mandate_id}/balance

# Execute debit
POST /v1/mandates/{mandate_id}/debit
{
  "amount": 50000,
  "description": "Recharge payment"
}
```

---

## Next Steps

### Immediate Actions (This Week)
1. ✅ **Implement 3 available tools** in MCP server
2. 📧 **Contact Razorpay** regarding missing APIs
3. 📋 **Document VI requirements** for Razorpay team
4. 🔄 **Set up project tracking** for API availability

### Short Term (2-4 weeks)
1. 📞 **Schedule calls** with Razorpay product team
2. 🏗️ **Build framework** for missing tools
3. 🔗 **Implement workarounds** using existing APIs
4. 📊 **Create monitoring** for API updates

### Medium Term (1-2 months)
1. 🚀 **Deploy available tools** to production
2. 🤝 **Negotiate custom API access** if needed
3. 🧪 **Test integration** with VI team
4. 📈 **Measure success metrics**

---

## Conclusion

While Razorpay provides a comprehensive payment platform, **significant API gaps exist** for advanced use cases like Vodafone Idea's agentic integration requirements. The current public APIs support only **33% (3/9) of the required tools**, making it impossible to fully implement VI's use cases without additional API development from Razorpay.

**Success depends on:**
1. Razorpay developing missing APIs for customer tokens and SBMD management
2. Implementation of server-to-server payment flows with OTP handling
3. Close collaboration between VI, Razorpay, and the MCP server development team

**Recommendation:** Proceed with implementing available tools while actively engaging Razorpay for the missing critical APIs.

---

**Document Prepared By:** AI Assistant  
**Review Required By:** Technical Team, VI Integration Team, Razorpay Partnership Team  
**Next Review Date:** 2 weeks from document creation