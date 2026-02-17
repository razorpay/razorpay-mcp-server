//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

import (
	"context"
	"encoding/json"
	"fmt"

	rzpsdk "github.com/razorpay/razorpay-go"

	"github.com/razorpay/razorpay-mcp-server/pkg/contextkey"
)

// getClientFromContextOrDefault returns either the provided default
// client or gets one from context.
func getClientFromContextOrDefault(
	ctx context.Context,
	defaultClient *rzpsdk.Client,
) (*rzpsdk.Client, error) {
	if defaultClient != nil {
		return defaultClient, nil
	}

	clientInterface := contextkey.ClientFromContext(ctx)
	if clientInterface == nil {
		return nil, fmt.Errorf("no client found in context")
	}

	client, ok := clientInterface.(*rzpsdk.Client)
	if !ok {
		return nil, fmt.Errorf("invalid client type in context")
	}

	return client, nil
}

// Helper to get keys or placeholders
func getKeysOrPlaceholders(creds Credentials) (string, string) {
	keyID := creds.KeyID
	keySecret := creds.KeySecret
	if keyID == "" {
		keyID = "rzp_test_YOUR_KEY_ID"
	}
	if keySecret == "" {
		keySecret = "YOUR_KEY_SECRET"
	}
	return keyID, keySecret
}

// Common wire_payment action for all backends
func getWirePaymentAction() FileAction {
	return FileAction{
		Action:      "wire_payment",
		Path:        "DISCOVER",
		Description: "CRITICAL: Discover and modify the actual checkout flow - DO NOT assume file names",
		Code: `⚠️⚠️⚠️ ACTION REQUIRED - DO NOT JUST ANALYZE ⚠️⚠️⚠️

YOU MUST ACTUALLY MODIFY THE CODE - NOT JUST REPORT WHAT'S WRONG!

If you find that checkout.js has placeholder comments or COD-only code:
- DO NOT just say "Found the issue - Razorpay is missing"
- DO NOT just list what files exist
- YOU MUST USE Edit/StrReplace TO WRITE THE FIX

Your job is to WRITE CODE, not write reports. After finding the checkout file,
IMMEDIATELY edit it to add the initiateRazorpayPayment() call.

═══════════════════════════════════════════════════════════════════════════════

STEP-BY-STEP DISCOVERY PROCESS:

1. FIND THE CHECKOUT/PAYMENT PAGE:
   - Look for: checkout.html, cart.html, payment.html, or checkout route/component
   - For SPAs: find the checkout component/page
   - For templates: find the template with the checkout form

2. FIND WHICH JS/COMPONENT HANDLES CHECKOUT:
   - Check <script> tags or imports
   - Common names: checkout.js, cart.js, payment.js, Checkout.jsx/vue/svelte
   - DO NOT assume app.js or main.js

3. ADD RAZORPAY SCRIPT/IMPORT:
   - For vanilla JS: add <script src="/js/razorpay.js"></script> to the correct HTML
   - For React/Vue/etc: import the Razorpay component in the checkout file

4. FIND THE PAYMENT/CHECKOUT FUNCTION:
   - Search for: initiatePayment, handleCheckout, checkout, placeOrder, processPayment, submitOrder
   - Look for: paymentMethod: 'cod', payment placeholders, TODO comments

5. MODIFY THAT FUNCTION to use Razorpay:

   CRITICAL: Before calling initiateRazorpayPayment(), you MUST:
   a) Collect all order data (cart items, customer info, shipping address, etc.)
   b) Save it to localStorage/state so it's available in the success callback

   Example pattern:
   async function existingCheckoutFunction() {
     // 1. GET the payment amount
     const total = calculateTotal(); // or get from existing code

     // 2. SAVE order data BEFORE payment (so success callback can access it)
     const pendingOrder = {
       items: getCartItems(),
       customerInfo: {
         name: document.getElementById('name-field').value,
         email: document.getElementById('email-field').value,
         // ... other fields from the form
       },
       shippingAddress: { /* ... */ },
       // Include whatever data the original order creation needed
     };
     localStorage.setItem('pendingOrder', JSON.stringify(pendingOrder));

     // 3. THEN call Razorpay payment
     initiateRazorpayPayment(
       total,
       async (paymentResponse) => {
         // 4. On SUCCESS: retrieve saved data and create order
         const orderData = JSON.parse(localStorage.getItem('pendingOrder'));
         // Add payment info
         orderData.paymentMethod = 'razorpay';
         orderData.paymentId = paymentResponse.paymentId;
         // Call existing order creation API with orderData
         localStorage.removeItem('pendingOrder'); // cleanup
       },
       (error) => {
         alert('Payment failed: ' + error.message);
         localStorage.removeItem('pendingOrder'); // cleanup on failure
       }
     );  // <-- closes initiateRazorpayPayment()
   }  // <-- CRITICAL: This } closes the function! DO NOT FORGET IT!

   ⚠️ BRACKET CHECK: Count your braces! The function MUST have a closing }
   If you see ");// next function" without a } between them, you forgot the closing brace!

   ⚠️ STRING REPLACEMENT WARNING: When using StrReplace/Edit to modify a function:
   - If your old_string includes part of the function body
   - Your new_string MUST include the closing } of the function
   - Don't replace "function body" with "new body" - replace "function body }" with "new body }"
   - The closing brace is part of what you're replacing!

   ⚠️ CRITICAL: DO NOT ESCAPE CHARACTERS IN JAVASCRIPT!
   When writing JavaScript code, use NORMAL characters - NO BACKSLASHES:
   ✅ CORRECT: document.getElementById('name')
   ❌ WRONG:   document.getElementById(BACKSLASH'name BACKSLASH')
   ✅ CORRECT: localStorage.setItem('key', value)
   ❌ WRONG:   localStorage.setItem(BACKSLASH'key BACKSLASH', value)
   For template literals, use normal backtick and $ - NOT backslash-backtick or backslash-$

   DO NOT add backslashes before quotes, backticks, or dollar signs in JavaScript!

COMMON MISTAKES TO AVOID:
- Modifying wrong file (e.g., app.js when checkout.html uses checkout.js)
- Adding script to index.html when checkout is in checkout.html
- Creating new functions instead of modifying existing checkout flow
- Leaving COD/placeholder payment code active
- NOT SAVING order data before payment (causes "order data not found" errors)
- Trying to access form fields in success callback (form may be gone/reset)
- MISSING CLOSING BRACE: The function must end with } after the initiateRazorpayPayment() call:
     );
   }  // <-- THIS CLOSING BRACE IS REQUIRED - don't forget it!

⚠️ CRITICAL: JUST ANALYZING = FAILURE ⚠️
If your output says "Found the issue" or "What's missing" without ACTUALLY EDITING the file,
YOU HAVE FAILED. Use Edit/StrReplace to modify checkout.js NOW.`,
	}
}

// Common AI instructions suffix for frontend wiring
func getFrontendWiringInstructions(frontend FrontendIntegration) string {
	return `

**CRITICAL - FRONTEND WIRING (DO NOT SKIP):**

1. DISCOVER the project structure:
   - Find where the checkout page/component is
   - Find which file handles the checkout logic
   - DO NOT assume file names like app.js or index.html

2. ADD the Razorpay frontend code:
   - ` + frontend.ScriptTag + `
   - Make sure it's loaded BEFORE the checkout code

3. WIRE UP the payment:
   - Find the existing checkout/payment function
   - Modify it to call the Razorpay payment function
   - Order creation should happen ONLY after payment succeeds
   - Payment failures should show error without creating order

THE INTEGRATION IS **NOT COMPLETE** IF:
- You modified the wrong JS file (one not used by checkout page)
- The checkout still bypasses payment (COD, direct order creation)
- You tell the user to "wire up payment" as a next step

DO NOT give "Next Steps" - complete EVERYTHING including frontend wiring.

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER - CHECK BEFORE FINISHING:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you FIND the actual checkout page/file?
✅ Did you ADD the Razorpay script to that page?
✅ Did you MODIFY the checkout function to call Razorpay payment?
✅ Did you include closing } braces for ALL functions?
✅ Did you ADD env vars to .env file?
✅ Did you use NORMAL quotes (not escaped \')?

❌ DO NOT give "Next Steps" - integration must be COMPLETE
❌ DO NOT tell user to "wire up" anything - YOU must do it
❌ DO NOT just ANALYZE and say "Found issue" - USE Edit TO FIX IT
❌ DO NOT escape quotes/backticks in JavaScript (\' is WRONG, ' is CORRECT)

If you said "Found the issue" but didn't edit the file, GO BACK AND EDIT IT NOW.`
}

// Helper functions
func containsSuffix(files []string, suffix string) bool {
	for _, f := range files {
		if len(f) >= len(suffix) && f[len(f)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}

func containsPath(files []string, path string) bool {
	for _, f := range files {
		if f == path || contains(f, path) {
			return true
		}
	}
	return false
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure json import is used
var _ = json.Marshal
