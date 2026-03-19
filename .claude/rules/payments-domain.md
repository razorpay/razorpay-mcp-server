---
description: "Guardrails for the payments domain (payments, tokens)"
globs: ["pkg/razorpay/payments.go", "pkg/razorpay/tokens.go"]
---

- `capture_payment` is only valid on payments with `status=authorized` — document this in tool description; any other status returns 400 from Razorpay API.
- Capture amount must equal the authorized amount — partial capture is not supported.
- `initiate_payment` uses S2S JSON v1 flow (`client.Payment.CreatePaymentJson()`) — not redirect-based; the `next` array in response contains follow-up actions.
- S2S initiate sub-flows require different params: saved method needs `token`; UPI collect needs `vpa`; UPI intent needs `upi_intent=true` — wrong combination produces API error, not local validation error.
- OTP flow is strictly sequential: `initiate_payment` → `resend_otp` (optional) → `submit_otp` — there is no OTP without a payment in flight.
- When `initiate_payment` detects `otp_generate` in the `next` array, `payments.go:sendOtp()` fires automatically — agent only needs to collect OTP from user and call `submit_otp`.
- Non-INR recurring payments use `client.Payment.CreateRecurringPayment()` instead of `CreatePaymentJson()` — switch happens inside `payments.go:createPaymentWithParams()`.
- `fetch_tokens` looks up by phone number (contact), NOT customer_id — internally resolves phone → customer_id via `Customer.Create` with `fail_existing=0`.
- `revoke_token` is IRREVERSIBLE — add explicit irreversibility warning in tool description.
- New token-related tools belong in `tokens.go` and must register in the `payments` toolset, NOT a separate tokens toolset.
- When `email` is omitted, `payments.go:addContactAndEmailToPaymentData()` synthesizes `{contact}@mcp.razorpay.com` — not a real email address.
- `fetch_payment_card_details` only works for card-method payments — document restriction for UPI/netbanking/wallet.
