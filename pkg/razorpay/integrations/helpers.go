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

// getKeysOrPlaceholders always returns placeholder values for security.
// Real credentials should never be exposed to the AI agent.
func getKeysOrPlaceholders(_ Credentials) (string, string) {
	return "rzp_test_YOUR_KEY_ID", "YOUR_KEY_SECRET"
}

func getNextStepsFile() FileAction {
	return FileAction{
		Action:      "create",
		Path:        "NEXT_STEPS.md",
		Description: "Setup guide - add your Razorpay credentials and test the integration",
		Code: `# Next Steps - Razorpay Integration Setup

## 1. Add Your Razorpay API Credentials

Open the ` + "`.env`" + ` file in your project root and replace the placeholder values:

` + "```" + `
RAZORPAY_KEY_ID=your_actual_key_id_here
RAZORPAY_KEY_SECRET=your_actual_key_secret_here
` + "```" + `

### Where to find your credentials

1. Log in to the [Razorpay Dashboard](https://dashboard.razorpay.com/)
2. Navigate to **Settings** → **API Keys**
3. Generate or copy your **Key ID** and **Key Secret**
4. For testing, use **Test Mode** keys (they start with ` + "`rzp_test_`" + `)


## 2. Test the Integration

After adding your credentials, start your development server and test with these credentials:

### Test Card
| Field       | Value                |
|-------------|----------------------|
| Card Number | 4111 1111 1111 1111  |
| Expiry      | Any future date      |
| CVV         | Any 3-digit number   |

### Test UPI
| Field | Value              |
|-------|--------------------|
| VPA   | success@razorpay   |

## 3. Go Live

When you are ready for production:

1. Switch to **Live Mode** on the Razorpay Dashboard
2. Generate **Live** API keys
3. Replace the test keys in your ` + "`.env`" + ` with the live keys
4. Ensure your server is running over HTTPS
`,
	}
}

func getNextStepsAIInstructions() string {
	return `

**AFTER APPLYING ALL FILES - YOU MUST DO THIS:**
1. Create the NEXT_STEPS.md file in the project root
2. OPEN the NEXT_STEPS.md file in the user's IDE/editor so they can see it immediately
3. Briefly tell the user that the integration is complete and they need to add their Razorpay credentials as shown in the file
4. DO NOT attempt to fill in real API credentials yourself - the user must do this manually`
}

// Common wire_payment action for all backends
func getWirePaymentAction() FileAction {
	return FileAction{
		Action:      "wire_payment",
		Path:        "DISCOVER",
		Description: "CRITICAL: Discover and modify the actual checkout flow - DO NOT assume file names",
		Code: `═══════════════════════════════════════════════════════════════════════════════
RULE #1 - CLOSING BRACES (READ THIS FIRST - MOST COMMON BUG)
═══════════════════════════════════════════════════════════════════════════════

When you add initiateRazorpayPayment() inside a function, the function MUST
still have its closing }. The pattern MUST be:

    initiateRazorpayPayment(
      amount,
      async (paymentResponse) => { ... },
      (error) => { ... }
    );
  }   <-- THIS CLOSING BRACE IS REQUIRED

If your edit ends with ");  " and no "}" after it, YOU HAVE A BUG.
After EVERY edit, verify: does the function still have its closing }?

When using StrReplace to modify a function:
- Your old_string MUST include the function's closing }
- Your new_string MUST include the function's closing }
- Example: replace "old body }" with "new body }" - NOT just "old body" with "new body"

AFTER editing any JS/TS file, run: node --check <filepath>
If it says "Unexpected end of input" you forgot a closing brace. FIX IT.

═══════════════════════════════════════════════════════════════════════════════
RULE #2 - YOU MUST EDIT CODE, NOT JUST ANALYZE
═══════════════════════════════════════════════════════════════════════════════

YOU MUST USE Edit/StrReplace TO WRITE THE FIX. Do NOT just report what's wrong.
If you say "Found the issue" without editing the file, YOU HAVE FAILED.

═══════════════════════════════════════════════════════════════════════════════
STEP-BY-STEP DISCOVERY PROCESS
═══════════════════════════════════════════════════════════════════════════════

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
     const total = calculateTotal();
     const pendingOrder = {
       items: getCartItems(),
       customerInfo: { name: document.getElementById('name-field').value },
     };
     localStorage.setItem('pendingOrder', JSON.stringify(pendingOrder));

     initiateRazorpayPayment(
       total,
       async (paymentResponse) => {
         const orderData = JSON.parse(localStorage.getItem('pendingOrder'));
         orderData.paymentMethod = 'razorpay';
         orderData.paymentId = paymentResponse.paymentId;
         localStorage.removeItem('pendingOrder');
       },
       (error) => {
         alert('Payment failed: ' + error.message);
         localStorage.removeItem('pendingOrder');
       }
     );
   }  <-- CLOSING BRACE IS HERE. DO NOT FORGET IT.

6. VERIFY SYNTAX: Run node --check on every JS/TS file you edited.
   If it fails, you likely forgot a closing brace. Fix it immediately.

DO NOT escape characters in JavaScript:
   ✅ CORRECT: document.getElementById('name')
   ❌ WRONG:   document.getElementById(BACKSLASH'name BACKSLASH')

COMMON MISTAKES:
- Modifying wrong file (e.g., app.js when checkout.html uses checkout.js)
- Adding script to index.html when checkout is in checkout.html
- Creating new functions instead of modifying existing checkout flow
- Leaving COD/placeholder payment code active
- NOT SAVING order data before payment (causes "order data not found" errors)
- Trying to access form fields in success callback (form may be gone/reset)
- FORGETTING THE CLOSING BRACE after initiateRazorpayPayment() call`,
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
✅ Did you run "node --check" on every JS/TS file you edited?
✅ Did you ADD env vars to .env file?
✅ Did you use NORMAL quotes (not escaped \')?

❌ DO NOT give "Next Steps" - integration must be COMPLETE
❌ DO NOT tell user to "wire up" anything - YOU must do it
❌ DO NOT just ANALYZE and say "Found issue" - USE Edit TO FIX IT
❌ DO NOT escape quotes/backticks in JavaScript (\' is WRONG, ' is CORRECT)

If "node --check" reports an error, FIX IT before finishing.
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
