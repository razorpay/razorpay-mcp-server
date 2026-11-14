//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

// =============================================================================
// EXPRESS + VANILLA JS INTEGRATION
// =============================================================================

func getExpressVanillaIntegration(language string, frontend FrontendIntegration) IntegrateCheckoutOutput {
	ext := "js"
	if language == "typescript" {
		ext = "ts"
	}

	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	razorpayRoutesCode := `const express = require('express');
const Razorpay = require('razorpay');
const crypto = require('crypto');

const router = express.Router();

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

async function getOrderAmount(req) {
  // Return { amountInSubunits, currency, reference } for the cart this
  // request belongs to. The price MUST come from the server: read it from
  // the cart or order record this project already has. Never read it from
  // the request body. The Razorpay signature covers the order id and the
  // payment id, not the price, so a shopper who chooses the amount can pay
  // 1 paise for any cart and still produce a signature that verifies.
  // Replace this body with a lookup, for example:
  //   const cart = await Cart.findOne({ user: req.user.id });
  //   return { amountInSubunits: Math.round(cart.total * 100), currency: 'INR', reference: String(cart.id) };
  throw new Error('getOrderAmount must read the amount from the server-side cart or order record. See the integration instructions.');
}

// Create Razorpay Order
router.post('/order', async (req, res) => {
  try {
    const { amountInSubunits, currency, reference } = await getOrderAmount(req);

    if (!amountInSubunits || amountInSubunits <= 0) {
      return res.status(400).json({ success: false, error: 'Invalid amount' });
    }

    const order = await razorpay.orders.create({
      amount: amountInSubunits,
      currency,
      receipt: reference,
    });

    res.json({
      success: true,
      orderId: order.id,
      amount: order.amount,
      currency: order.currency,
      keyId: process.env.RAZORPAY_KEY_ID,
    });
  } catch (error) {
    console.error('Razorpay order creation failed:', error);
    res.status(500).json({ success: false, error: 'Failed to create payment order' });
  }
});

// Verify Payment Signature
router.post('/verify', async (req, res) => {
  try {
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = req.body;

    if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
      return res.status(400).json({ success: false, error: 'Missing payment details' });
    }

    const expectedSignature = crypto
      .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
      .update(razorpay_order_id + '|' + razorpay_payment_id)
      .digest('hex');

    const expectedBuf = Buffer.from(expectedSignature, 'utf8');
    const receivedBuf = Buffer.from(razorpay_signature || '', 'utf8');
    const isValidSignature =
      expectedBuf.length === receivedBuf.length &&
      crypto.timingSafeEqual(expectedBuf, receivedBuf);

    if (isValidSignature) {
      const payment = await razorpay.payments.fetch(razorpay_payment_id);
      if (payment.order_id !== razorpay_order_id) {
        return res.status(400).json({ success: false, error: 'Payment does not match the order' });
      }
      if (!['authorized', 'captured'].includes(payment.status)) {
        return res.status(400).json({ success: false, error: 'Payment does not match the order' });
      }
      const order = await razorpay.orders.fetch(razorpay_order_id);
      if (Number(payment.amount) !== Number(order.amount)) {
        return res.status(400).json({ success: false, error: 'Payment does not match the order' });
      }
      if (payment.currency !== order.currency) {
        return res.status(400).json({ success: false, error: 'Payment does not match the order' });
      }

      res.json({
        success: true,
        message: 'Payment verified successfully',
        paymentId: razorpay_payment_id,
        orderId: razorpay_order_id,
        reference: order.receipt,
        amount: order.amount,
        currency: order.currency,
      });
    } else {
      res.status(400).json({ success: false, error: 'Invalid payment signature' });
    }
  } catch (error) {
    console.error('Payment verification failed:', error);
    res.status(500).json({ success: false, error: 'Payment verification failed' });
  }
});

module.exports = router;
`

	// Server.js setup code - imports grouped together to prevent ordering issues
	serverSetupCode := `// Add these lines at the TOP of server.js (before other code):
require('dotenv').config();
const razorpayRoutes = require('./routes/razorpay');

// Add this line with your other app.use() middleware (AFTER the above imports):
// app.use('/api/razorpay', razorpayRoutes);
`

	files := []FileAction{
		{
			Action:      "create",
			Path:        "routes/razorpay." + ext,
			Code:        razorpayRoutesCode,
			Description: "Razorpay API routes for order creation and payment verification",
		},
		{
			Action:      "create",
			Path:        frontend.FileName,
			Code:        frontend.Code,
			Description: frontend.Description,
		},
		{
			Action:      "insert_code",
			Path:        "server.js",
			Description: "Add Razorpay setup to server.js - MUST be done in this exact order",
			Code:        serverSetupCode,
			Edits: []EditItem{
				{
					Line: "STEP 1 - At the VERY TOP of server.js (line 1, before any other code)",
					Add:  "require('dotenv').config();",
					Why:  "Must be first line to load env vars before anything else",
				},
				{
					Line: "STEP 2 - Immediately after dotenv, with other require/import statements at the top",
					Add:  "const razorpayRoutes = require('./routes/razorpay');",
					Why:  "Import MUST come before usage - add this near top with other imports",
				},
				{
					Line: "STEP 3 - Later in the file, with other app.use() middleware registrations",
					Add:  "app.use('/api/razorpay', razorpayRoutes);",
					Why:  "Uses razorpayRoutes - MUST come AFTER the require statement above",
				},
			},
		},
		{
			Action:      "wire_payment",
			Path:        "DISCOVER",
			Description: "CRITICAL: Discover and modify the actual checkout flow - DO NOT assume file names",
			Code: `STEP-BY-STEP DISCOVERY PROCESS:

1. FIND THE CHECKOUT HTML PAGE:
   - Look for: checkout.html, cart.html, payment.html, or checkout page in index.html
   - Check which HTML file contains the checkout form/button
   - Note: It may NOT be index.html

2. FIND WHICH JS FILE IS LOADED BY THAT HTML:
   - Look at <script> tags in the checkout HTML
   - Common names: checkout.js, cart.js, payment.js, app.js, main.js, bundle.js
   - The correct file is whatever the checkout HTML actually loads
   - Note: Do NOT assume app.js - check the actual HTML

3. ADD RAZORPAY SCRIPT TO THE CORRECT HTML:
   - Add <script src="/js/razorpay.js"></script> (or correct path)
   - Add it BEFORE the checkout JS file so it's available
   - Add to the HTML file that has the checkout, NOT just index.html

4. FIND THE PAYMENT/CHECKOUT FUNCTION:
   - Search for functions like: initiatePayment, handleCheckout, checkout,
     placeOrder, processPayment, submitOrder, handlePayment
   - Look for comments like "payment integration", "add payment here", "TODO"
   - Look for paymentMethod: 'cod' or placeholder payment code

5. MODIFY THAT FUNCTION to call initiateRazorpayPayment():

   CRITICAL: Before calling initiateRazorpayPayment(), you MUST:
   a) Collect all order data (cart items, customer info, shipping address, etc.)
   b) Save it to localStorage so it's available in the success callback

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
         phone: document.getElementById('phone-field').value,
       },
       shippingAddress: {
         line1: document.getElementById('address-line1').value,
         city: document.getElementById('city').value,
         // ... other address fields
       },
       // Include whatever data the original order creation needed
     };
     localStorage.setItem('pendingOrder', JSON.stringify(pendingOrder));

     // 3. THEN call Razorpay payment
     initiateRazorpayPayment(
       total,
       async (paymentResponse) => {
         // 4. On SUCCESS: retrieve saved data and create order
         const orderData = JSON.parse(localStorage.getItem('pendingOrder'));
         orderData.paymentMethod = 'razorpay';
         orderData.paymentId = paymentResponse.paymentId;
         // Call the existing order creation API
         localStorage.removeItem('pendingOrder');
       },
       (error) => {
         alert('Payment failed: ' + error.message);
         localStorage.removeItem('pendingOrder');
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
- Adding to app.js when checkout.html uses checkout.js
- Adding script to index.html when checkout is in checkout.html
- Creating new functions instead of modifying existing ones
- Leaving the original COD/placeholder code active
- NOT SAVING order data before payment (causes "order data not found" errors)
- Trying to access form fields in success callback (form may be gone/reset)
- MISSING CLOSING BRACE: The function must end with } after the initiateRazorpayPayment() call:
     );
   }  // <-- THIS CLOSING BRACE IS REQUIRED - don't forget it!`,
		},
	}

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Express + " + frontend.Framework,
		Files:   files,
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
			{Name: "dotenv", InstallCommand: "npm install dotenv"},
		},
		EnvVars: []EnvVar{
			{Name: "RAZORPAY_KEY_ID", Value: keyID},
			{Name: "RAZORPAY_KEY_SECRET", Value: keySecret},
		},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: YOU MUST COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!
DO NOT tell the user to "wire up payment" or give any "Next Steps" - YOU must do it all!
If you finish without modifying the checkout function to call initiateRazorpayPayment(), YOU HAVE FAILED.

⚠️ DO NOT JUST ANALYZE - YOU MUST WRITE CODE!
If you find placeholder comments or COD-only code in checkout.js, DO NOT just report "Found the issue".
YOU MUST USE Edit/StrReplace TO ACTUALLY FIX IT. Analyzing without fixing = FAILURE.

CRITICAL INSTRUCTIONS - THE INTEGRATION IS NOT COMPLETE UNTIL ALL STEPS ARE DONE:

**FIRST: DISCOVER THE PROJECT STRUCTURE**
Before making changes, find:
- Where is the checkout page? (checkout.html? cart.html? part of index.html?)
- Which JS file handles checkout? (checkout.js? cart.js? app.js? - CHECK the HTML)
- Where is the payment/checkout function? (search for initiatePayment, handleCheckout, placeOrder, etc.)

DO NOT ASSUME app.js or index.html - find the ACTUAL files used for checkout!

**THEN FOLLOW THESE STEPS:**

1) npm install razorpay dotenv

2) Create routes/razorpay.js (backend routes file)

3) Create ` + frontend.FileName + ` (frontend file)

4) Create .env file with the provided RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET

5) Edit server.js (or main server file) - **CRITICAL ORDER (WILL CRASH IF WRONG)**:

   ⚠️ WARNING: JavaScript const/let variables CANNOT be used before declaration!
   The require() MUST appear BEFORE the app.use() in the file, or you'll get:
   "ReferenceError: Cannot access 'razorpayRoutes' before initialization"

   a) Add at TOP (line 1): require('dotenv').config();
   b) Add IMMEDIATELY AFTER with other requires at TOP of file: const razorpayRoutes = require('./routes/razorpay');
   c) Add app.use('/api/razorpay', razorpayRoutes); at the TOP LEVEL of the file.
      It MUST be at the same indentation level as app.listen().
      NEVER place it inside app.get(), app.post(), or any route handler/callback.

   CORRECT ORDER IN FILE:
   Line 1:  require('dotenv').config();
   Line 2+: const razorpayRoutes = require('./routes/razorpay');  // <-- DECLARE FIRST
   ...
   Line 50+: app.use('/api/razorpay', razorpayRoutes);  // <-- TOP LEVEL, with other middleware

   WRONG (WILL CRASH):
   app.use('/api/razorpay', razorpayRoutes);  // ❌ Used before declared
   const razorpayRoutes = require(...);        // ❌ Declared after use

   WRONG (ROUTES WON'T WORK):
   app.post('/some-route', (req, res) => {
     app.use('/api/razorpay', razorpayRoutes);  // ❌ INSIDE a route handler!
   });

6) Add razorpay.js script to the CORRECT HTML file:
   - Find which HTML has the checkout (may be checkout.html, NOT index.html)
   - Add <script src="/js/razorpay.js"></script> to THAT file
   - Add it BEFORE the checkout JS file in the script order

7) **CRITICAL - WIRE UP PAYMENT (DO NOT SKIP):**
   - Find the ACTUAL checkout JS file (the one loaded by the checkout HTML)
   - Find the payment/checkout function in THAT file
   - Modify it to call initiateRazorpayPayment() BEFORE creating the order

   The flow must be:
   User clicks checkout -> initiateRazorpayPayment(total, onSuccess, onError)
   -> On SUCCESS: create order with paymentMethod: 'razorpay', paymentId
   -> On FAILURE: show error, don't create order

8) **SECURITY - SERVER-SIDE PRICE (MANDATORY - DO NOT SKIP):**
   - routes/razorpay.js contains getOrderAmount(req) as a stub that throws on purpose.
   - You MUST implement getOrderAmount(req) against this project's real cart/order model.
     It must return { amountInSubunits, currency, reference } read from the server-side record.
   - NEVER take the amount from the request body - the order route ignores it by design.
   - When marking an order as paid, do it ONLY when the verify response's 'reference' matches
     the order being marked AND its 'amount' equals that order's own server-side total.

THE INTEGRATION IS **NOT COMPLETE** IF:
- You modified a JS file that isn't loaded by the checkout HTML
- The checkout still creates orders without calling initiateRazorpayPayment()
- The razorpay.js script isn't in the checkout HTML
- You tell the user to "wire up the payment" as a next step
- getOrderAmount(req) is still the throwing stub (you did not implement it against the real cart/order model)

DO NOT give "Next Steps" - complete EVERYTHING including discovering the correct files.

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER - READ THIS BEFORE FINISHING:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you FIND the actual checkout HTML/JS files (not assume app.js)?
✅ Did you ADD razorpay.js script to the checkout HTML?
✅ Did you MODIFY the checkout function to call initiateRazorpayPayment()?
✅ Did you put require() BEFORE app.use() in server.js?
✅ Is app.use('/api/razorpay', ...) at the TOP LEVEL (not inside any route handler)?
✅ Did you include the closing } brace for ALL functions?
✅ Did you use NORMAL quotes (not escaped \')?
✅ Did you implement getOrderAmount(req) against the real cart/order model?
✅ Did you mark orders paid only when the verify response reference and amount match that order?

❌ DO NOT give "Next Steps" - the integration must be COMPLETE
❌ DO NOT escape quotes/backticks in JavaScript (\' is WRONG, ' is CORRECT)
❌ DO NOT tell user to "wire up payment" - YOU must do it
❌ DO NOT finish without modifying the checkout function
❌ DO NOT just ANALYZE and report "Found issue" - USE Edit TO FIX IT
❌ DO NOT take the payment amount from the request body - getOrderAmount(req) reads it from the server-side cart or order record

If ANY checkbox above is NO, GO BACK AND FIX IT NOW.
If you said "Found the issue" but didn't edit the file, GO BACK AND EDIT IT NOW.`,
	}
}

// =============================================================================
// NEXT.JS + REACT INTEGRATION
// =============================================================================

func getNextjsReactIntegration(language string) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	orderRouteCode := `import { NextRequest, NextResponse } from 'next/server';
import Razorpay from 'razorpay';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID!,
  key_secret: process.env.RAZORPAY_KEY_SECRET!,
});

async function getOrderAmount(request: NextRequest): Promise<{ amountInSubunits: number; currency: string; reference: string }> {
  // Return { amountInSubunits, currency, reference } for the cart this
  // request belongs to. The price MUST come from the server: read it from
  // the cart or order record this project already has. Never read it from
  // the request body. The Razorpay signature covers the order id and the
  // payment id, not the price, so a shopper who chooses the amount can pay
  // 1 paise for any cart and still produce a signature that verifies.
  // Replace this body with a lookup, for example:
  //   const cart = await Cart.findOne({ user: request.user.id });
  //   return { amountInSubunits: Math.round(cart.total * 100), currency: 'INR', reference: String(cart.id) };
  throw new Error('getOrderAmount must read the amount from the server-side cart or order record. See the integration instructions.');
}

export async function POST(request: NextRequest) {
  try {
    const { amountInSubunits, currency, reference } = await getOrderAmount(request);

    if (!amountInSubunits || amountInSubunits <= 0) {
      return NextResponse.json({ success: false, error: 'Invalid amount' }, { status: 400 });
    }

    const order = await razorpay.orders.create({
      amount: amountInSubunits,
      currency,
      receipt: reference,
    });

    return NextResponse.json({
      success: true,
      orderId: order.id,
      amount: order.amount,
      currency: order.currency,
      keyId: process.env.RAZORPAY_KEY_ID,
    });
  } catch (error) {
    console.error('Razorpay order creation failed:', error);
    return NextResponse.json({ success: false, error: 'Failed to create order' }, { status: 500 });
  }
}
`

	verifyRouteCode := `import { NextRequest, NextResponse } from 'next/server';
import Razorpay from 'razorpay';
import crypto from 'crypto';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID!,
  key_secret: process.env.RAZORPAY_KEY_SECRET!,
});

export async function POST(request: NextRequest) {
  try {
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = await request.json();

    if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
      return NextResponse.json({ success: false, error: 'Missing payment details' }, { status: 400 });
    }

    const expectedSignature = crypto
      .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET!)
      .update(razorpay_order_id + '|' + razorpay_payment_id)
      .digest('hex');

    const expectedBuf = Buffer.from(expectedSignature, 'utf8');
    const receivedBuf = Buffer.from(razorpay_signature || '', 'utf8');
    const isValid =
      expectedBuf.length === receivedBuf.length &&
      crypto.timingSafeEqual(
      expectedBuf,
      receivedBuf
    );

    if (isValid) {
      const payment = await razorpay.payments.fetch(razorpay_payment_id);
      if (payment.order_id !== razorpay_order_id) {
        return NextResponse.json({ success: false, error: 'Payment does not match the order' }, { status: 400 });
      }
      if (!['authorized', 'captured'].includes(payment.status)) {
        return NextResponse.json({ success: false, error: 'Payment does not match the order' }, { status: 400 });
      }
      const order = await razorpay.orders.fetch(razorpay_order_id);
      if (Number(payment.amount) !== Number(order.amount)) {
        return NextResponse.json({ success: false, error: 'Payment does not match the order' }, { status: 400 });
      }
      if (payment.currency !== order.currency) {
        return NextResponse.json({ success: false, error: 'Payment does not match the order' }, { status: 400 });
      }

      return NextResponse.json({
        success: true,
        message: 'Payment verified',
        paymentId: razorpay_payment_id,
        orderId: razorpay_order_id,
        reference: order.receipt,
        amount: order.amount,
        currency: order.currency,
      });
    } else {
      return NextResponse.json({ success: false, error: 'Invalid signature' }, { status: 400 });
    }
  } catch (error) {
    console.error('Verification failed:', error);
    return NextResponse.json({ success: false, error: 'Verification failed' }, { status: 500 });
  }
}
`

	checkoutComponentCode := `'use client';

import { useState } from 'react';
import Script from 'next/script';

interface RazorpayCheckoutProps {
  amount: number;
  onSuccess?: (data: { paymentId: string; orderId: string }) => void;
  onError?: (error: Error) => void;
  buttonText?: string;
  className?: string;
}

export function RazorpayCheckout({
  amount,
  onSuccess,
  onError,
  buttonText = 'Pay Now',
  className = ''
}: RazorpayCheckoutProps) {
  const [loading, setLoading] = useState(false);
  const [scriptLoaded, setScriptLoaded] = useState(false);

  const handlePayment = async () => {
    if (!scriptLoaded || loading) return;
    setLoading(true);

    try {
      const orderRes = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });

      const orderData = await orderRes.json();
      if (!orderData.success) throw new Error(orderData.error);

      const options = {
        key: orderData.keyId,
        amount: orderData.amount,
        currency: orderData.currency,
        name: 'Payment',
        order_id: orderData.orderId,
        handler: async (response: any) => {
          const verifyRes = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const verifyData = await verifyRes.json();

          if (verifyData.success) {
            onSuccess?.({ paymentId: verifyData.paymentId, orderId: verifyData.orderId });
          } else {
            onError?.(new Error(verifyData.error));
          }
          setLoading(false);
        },
        modal: { ondismiss: () => setLoading(false) },
        theme: { color: '#528FF0' },
      };

      const razorpay = new (window as any).Razorpay(options);
      razorpay.on('payment.failed', (res: any) => {
        onError?.(new Error(res.error.description));
        setLoading(false);
      });
      razorpay.open();
    } catch (error) {
      onError?.(error as Error);
      setLoading(false);
    }
  };

  return (
    <>
      <Script
        src="https://checkout.razorpay.com/v1/checkout.js"
        onLoad={() => setScriptLoaded(true)}
      />
      <button
        onClick={handlePayment}
        disabled={loading || !scriptLoaded}
        className={className || 'bg-blue-600 text-white px-6 py-2 rounded disabled:opacity-50'}
      >
        {loading ? 'Processing...' : buttonText}
      </button>
    </>
  );
}
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Next.js + React",
		Files: []FileAction{
			{
				Action:      "create",
				Path:        "app/api/razorpay/order/route.ts",
				Code:        orderRouteCode,
				Description: "API route for creating Razorpay orders",
			},
			{
				Action:      "create",
				Path:        "app/api/razorpay/verify/route.ts",
				Code:        verifyRouteCode,
				Description: "API route for verifying payment signatures",
			},
			{
				Action:      "create",
				Path:        "components/RazorpayCheckout.tsx",
				Code:        checkoutComponentCode,
				Description: "React component for Razorpay checkout button",
			},
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
		},
		EnvVars: []EnvVar{
			{Name: "RAZORPAY_KEY_ID", Value: keyID},
			{Name: "RAZORPAY_KEY_SECRET", Value: keySecret},
		},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

STEPS:
1) Install razorpay package
2) Create ALL the API route files
3) Create the RazorpayCheckout component
4) Add env vars to .env.local
5) WIRE UP the component in your checkout page - import and use <RazorpayCheckout />
6) SECURITY - SERVER-SIDE PRICE (MANDATORY): app/api/razorpay/order/route.ts contains
   getOrderAmount(request) as a stub that throws. You MUST implement it against this
   project's real cart/order model so it returns { amountInSubunits, currency, reference }
   from the server-side record. NEVER take the amount from the request body - the API
   route ignores the posted amount by design. Mark an order as paid ONLY when the verify
   response's 'reference' matches the order being marked AND its 'amount' equals that
   order's own server-side total.

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you CREATE all API routes?
✅ Did you CREATE the RazorpayCheckout component?
✅ Did you ADD the component to the checkout page?
✅ Did you ADD env vars to .env.local?
✅ Did you implement getOrderAmount(request) against the real cart/order model?
✅ Did you mark orders paid only when the verify response reference and amount match that order?

❌ DO NOT give "Next Steps" - integration must be COMPLETE
❌ DO NOT tell user to "add the component" - YOU must do it
❌ DO NOT take the payment amount from the request body - getOrderAmount(request) reads it server-side`,
	}
}

// =============================================================================
// NODE.JS ADDITIONAL FRAMEWORKS (Fastify, Koa, Nuxt)
// =============================================================================

func getFastifyIntegration(language string, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	routesCode := `const Razorpay = require('razorpay');
const crypto = require('crypto');

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

async function getOrderAmount(request) {
  // Return { amountInSubunits, currency, reference } for the cart this
  // request belongs to. The price MUST come from the server: read it from
  // the cart or order record this project already has. Never read it from
  // the request body. The Razorpay signature covers the order id and the
  // payment id, not the price, so a shopper who chooses the amount can pay
  // 1 paise for any cart and still produce a signature that verifies.
  // Replace this body with a lookup, for example:
  //   const cart = await Cart.findOne({ user: request.user.id });
  //   return { amountInSubunits: Math.round(cart.total * 100), currency: 'INR', reference: String(cart.id) };
  throw new Error('getOrderAmount must read the amount from the server-side cart or order record. See the integration instructions.');
}

async function razorpayRoutes(fastify, options) {
  fastify.post('/api/razorpay/order', async (request, reply) => {
    try {
      const { amountInSubunits, currency, reference } = await getOrderAmount(request);

      if (!amountInSubunits || amountInSubunits <= 0) {
        return reply.status(400).send({ success: false, error: 'Invalid amount' });
      }

      const order = await razorpay.orders.create({
        amount: amountInSubunits,
        currency,
        receipt: reference,
      });

      return { success: true, orderId: order.id, amount: order.amount, currency: order.currency, keyId: process.env.RAZORPAY_KEY_ID };
    } catch (error) {
      return reply.status(500).send({ success: false, error: 'Failed to create order' });
    }
  });

  fastify.post('/api/razorpay/verify', async (request, reply) => {
    try {
      const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = request.body;

      if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
        return reply.status(400).send({ success: false, error: 'Missing payment details' });
      }

      const expectedSignature = crypto
        .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
        .update(razorpay_order_id + '|' + razorpay_payment_id)
        .digest('hex');

      const expectedBuf = Buffer.from(expectedSignature, 'utf8');
      const receivedBuf = Buffer.from(razorpay_signature || '', 'utf8');
      const isValidSignature =
        expectedBuf.length === receivedBuf.length &&
        crypto.timingSafeEqual(expectedBuf, receivedBuf);

      if (isValidSignature) {
        const payment = await razorpay.payments.fetch(razorpay_payment_id);
        if (payment.order_id !== razorpay_order_id) {
          return reply.status(400).send({ success: false, error: 'Payment does not match the order' });
        }
        if (!['authorized', 'captured'].includes(payment.status)) {
          return reply.status(400).send({ success: false, error: 'Payment does not match the order' });
        }
        const order = await razorpay.orders.fetch(razorpay_order_id);
        if (Number(payment.amount) !== Number(order.amount)) {
          return reply.status(400).send({ success: false, error: 'Payment does not match the order' });
        }
        if (payment.currency !== order.currency) {
          return reply.status(400).send({ success: false, error: 'Payment does not match the order' });
        }
        return { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id, reference: order.receipt, amount: order.amount, currency: order.currency };
      }
      return reply.status(400).send({ success: false, error: 'Invalid signature' });
    } catch (error) {
      return reply.status(500).send({ success: false, error: 'Verification failed' });
    }
  });
}

module.exports = razorpayRoutes;
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Fastify + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "routes/razorpay.js", Code: routesCode, Description: "Fastify routes for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "app.js", Description: "Register routes", Edits: []EditItem{
				{Line: "After require statements", Add: "require('dotenv').config();", Why: "Load env vars"},
				{Line: "At TOP LEVEL after fastify instance creation (same level as fastify.listen) - NEVER inside a route handler", Add: "fastify.register(require('./routes/razorpay'));", Why: "MUST be top-level. If placed inside a route handler, Razorpay endpoints won't exist"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
			{Name: "dotenv", InstallCommand: "npm install dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCIES**
- Run: npm install razorpay dotenv

**STEP 2: CREATE ROUTES**
- Create routes/razorpay.js with the provided Fastify route handlers

**STEP 3: REGISTER ROUTES**
- Find the main server file (commonly app.js, server.js, or index.js)
- Add require('dotenv').config() at the top
- Add fastify.register(require('./routes/razorpay')) at the TOP LEVEL of the file
  It MUST be at the same indentation level as fastify.listen().
  NEVER place it inside a route handler, callback, or function.

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET

**STEP 5: SECURITY - SERVER-SIDE PRICE (MANDATORY)**
- routes/razorpay.js contains getOrderAmount(request) as a stub that throws on purpose.
- You MUST implement getOrderAmount(request) against this project's real cart/order model.
  It must return { amountInSubunits, currency, reference } read from the server-side record.
- NEVER take the amount from the request body - the order route ignores it by design.
- Mark an order as paid ONLY when the verify response's 'reference' matches the order being
  marked AND its 'amount' equals that order's own server-side total.

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in route files
❌ DO NOT assume the main file is app.js - find the actual entry point
❌ DO NOT skip payment verification
❌ DO NOT place fastify.register() inside a route handler - it must be top-level
❌ DO NOT take the payment amount from the request body - getOrderAmount(request) reads it server-side

FINAL CHECKLIST:
✅ Did you run npm install razorpay dotenv?
✅ Did you create routes/razorpay.js?
✅ Did you register routes at TOP LEVEL in the main server file (not inside any handler)?
✅ Did you create .env with credentials?
✅ Did you implement getOrderAmount(request) against the real cart/order model?
✅ Did you mark orders paid only when the verify response reference and amount match that order?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

func getKoaIntegration(language string, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	routerCode := `const Router = require('@koa/router');
const Razorpay = require('razorpay');
const crypto = require('crypto');

const router = new Router({ prefix: '/api/razorpay' });

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

async function getOrderAmount(ctx) {
  // Return { amountInSubunits, currency, reference } for the cart this
  // request belongs to. The price MUST come from the server: read it from
  // the cart or order record this project already has. Never read it from
  // the request body. The Razorpay signature covers the order id and the
  // payment id, not the price, so a shopper who chooses the amount can pay
  // 1 paise for any cart and still produce a signature that verifies.
  // Replace this body with a lookup, for example:
  //   const cart = await Cart.findOne({ user: ctx.user.id });
  //   return { amountInSubunits: Math.round(cart.total * 100), currency: 'INR', reference: String(cart.id) };
  throw new Error('getOrderAmount must read the amount from the server-side cart or order record. See the integration instructions.');
}

router.post('/order', async (ctx) => {
  try {
    const { amountInSubunits, currency, reference } = await getOrderAmount(ctx);

    if (!amountInSubunits || amountInSubunits <= 0) {
      ctx.status = 400;
      ctx.body = { success: false, error: 'Invalid amount' };
      return;
    }

    const order = await razorpay.orders.create({
      amount: amountInSubunits,
      currency,
      receipt: reference,
    });

    ctx.body = { success: true, orderId: order.id, amount: order.amount, currency: order.currency, keyId: process.env.RAZORPAY_KEY_ID };
  } catch (error) {
    ctx.status = 500;
    ctx.body = { success: false, error: 'Failed to create order' };
  }
});

router.post('/verify', async (ctx) => {
  try {
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = ctx.request.body;

    if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
      ctx.status = 400;
      ctx.body = { success: false, error: 'Missing payment details' };
      return;
    }

    const expectedSignature = crypto
      .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
      .update(razorpay_order_id + '|' + razorpay_payment_id)
      .digest('hex');

    const expectedBuf = Buffer.from(expectedSignature, 'utf8');
    const receivedBuf = Buffer.from(razorpay_signature || '', 'utf8');
    const isValidSignature =
      expectedBuf.length === receivedBuf.length &&
      crypto.timingSafeEqual(expectedBuf, receivedBuf);

    if (isValidSignature) {
      const payment = await razorpay.payments.fetch(razorpay_payment_id);
      if (payment.order_id !== razorpay_order_id) {
        ctx.status = 400;
        ctx.body = { success: false, error: 'Payment does not match the order' };
        return;
      }
      if (!['authorized', 'captured'].includes(payment.status)) {
        ctx.status = 400;
        ctx.body = { success: false, error: 'Payment does not match the order' };
        return;
      }
      const order = await razorpay.orders.fetch(razorpay_order_id);
      if (Number(payment.amount) !== Number(order.amount)) {
        ctx.status = 400;
        ctx.body = { success: false, error: 'Payment does not match the order' };
        return;
      }
      if (payment.currency !== order.currency) {
        ctx.status = 400;
        ctx.body = { success: false, error: 'Payment does not match the order' };
        return;
      }
      ctx.body = { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id, reference: order.receipt, amount: order.amount, currency: order.currency };
    } else {
      ctx.status = 400;
      ctx.body = { success: false, error: 'Invalid signature' };
    }
  } catch (error) {
    ctx.status = 500;
    ctx.body = { success: false, error: 'Verification failed' };
  }
});

module.exports = router;
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Koa + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "routes/razorpay.js", Code: routerCode, Description: "Koa router for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "app.js", Description: "Use router", Edits: []EditItem{
				{Line: "At top", Add: "require('dotenv').config();", Why: "Load env vars"},

				{Line: "After imports at the top of the file", Add: "const razorpayRouter = require('./routes/razorpay');", Why: "Import router"},
				{Line: "At TOP LEVEL after app creation (same level as app.listen) - NEVER inside a route handler", Add: "app.use(razorpayRouter.routes()).use(razorpayRouter.allowedMethods());", Why: "MUST be top-level. If placed inside a route handler, Razorpay endpoints won't exist"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
			{Name: "@koa/router", InstallCommand: "npm install @koa/router"},
			{Name: "koa-bodyparser", InstallCommand: "npm install koa-bodyparser"},
			{Name: "dotenv", InstallCommand: "npm install dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCIES**
- Run: npm install razorpay @koa/router koa-bodyparser dotenv

**STEP 2: CREATE ROUTER**
- Create routes/razorpay.js with the provided Koa router

**STEP 3: WIRE UP ROUTER**
- Find the main server file (commonly app.js, server.js, or index.js)
- Add require('dotenv').config() at the top
- Add require('koa-bodyparser') and use it: app.use(bodyParser())
- Import and use the Razorpay router at the TOP LEVEL: app.use(razorpayRouter.routes())
  It MUST be at the same indentation level as app.listen().
  NEVER place it inside a route handler, callback, or function.

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET

**STEP 5: SECURITY - SERVER-SIDE PRICE (MANDATORY)**
- routes/razorpay.js contains getOrderAmount(ctx) as a stub that throws on purpose.
- You MUST implement getOrderAmount(ctx) against this project's real cart/order model.
  It must return { amountInSubunits, currency, reference } read from the server-side record.
- NEVER take the amount from the request body - the order route ignores it by design.
- Mark an order as paid ONLY when the verify response's 'reference' matches the order being
  marked AND its 'amount' equals that order's own server-side total.

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in route files
❌ DO NOT assume the main file is app.js - find the actual entry point
❌ DO NOT forget koa-bodyparser - Koa needs it to parse request bodies
❌ DO NOT skip payment verification
❌ DO NOT place app.use(router) inside a route handler - it must be top-level
❌ DO NOT take the payment amount from the request body - getOrderAmount(ctx) reads it server-side

FINAL CHECKLIST:
✅ Did you run npm install razorpay @koa/router koa-bodyparser dotenv?
✅ Did you create routes/razorpay.js?
✅ Did you register the router at TOP LEVEL in the main server file (not inside any handler)?
✅ Did you add koa-bodyparser middleware?
✅ Did you create .env with credentials?
✅ Did you implement getOrderAmount(ctx) against the real cart/order model?
✅ Did you mark orders paid only when the verify response reference and amount match that order?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

func getNuxtIntegration(language string) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	orderAPICode := `import Razorpay from 'razorpay';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

async function getOrderAmount(event): Promise<{ amountInSubunits: number; currency: string; reference: string }> {
  // Return { amountInSubunits, currency, reference } for the cart this
  // request belongs to. The price MUST come from the server: read it from
  // the cart or order record this project already has. Never read it from
  // the request body. The Razorpay signature covers the order id and the
  // payment id, not the price, so a shopper who chooses the amount can pay
  // 1 paise for any cart and still produce a signature that verifies.
  // Replace this body with a lookup, for example:
  //   const cart = await Cart.findOne({ user: event.context.user.id });
  //   return { amountInSubunits: Math.round(cart.total * 100), currency: 'INR', reference: String(cart.id) };
  throw new Error('getOrderAmount must read the amount from the server-side cart or order record. See the integration instructions.');
}

export default defineEventHandler(async (event) => {
  try {
    const { amountInSubunits, currency, reference } = await getOrderAmount(event);

    if (!amountInSubunits || amountInSubunits <= 0) {
      throw createError({ statusCode: 400, message: 'Invalid amount' });
    }

    const order = await razorpay.orders.create({
      amount: amountInSubunits,
      currency,
      receipt: reference,
    });

    return { success: true, orderId: order.id, amount: order.amount, currency: order.currency, keyId: process.env.RAZORPAY_KEY_ID };
  } catch (error) {
    throw createError({ statusCode: 500, message: 'Failed to create order' });
  }
});
`

	verifyAPICode := `import Razorpay from 'razorpay';
import crypto from 'crypto';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

export default defineEventHandler(async (event) => {
  try {
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = await readBody(event);

    if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
      throw createError({ statusCode: 400, message: 'Missing payment details' });
    }

    const expectedSignature = crypto
      .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
      .update(razorpay_order_id + '|' + razorpay_payment_id)
      .digest('hex');

    const expectedBuf = Buffer.from(expectedSignature, 'utf8');
    const receivedBuf = Buffer.from(razorpay_signature || '', 'utf8');
    const isValidSignature =
      expectedBuf.length === receivedBuf.length &&
      crypto.timingSafeEqual(expectedBuf, receivedBuf);

    if (isValidSignature) {
      const payment = await razorpay.payments.fetch(razorpay_payment_id);
      if (payment.order_id !== razorpay_order_id) {
        throw createError({ statusCode: 400, message: 'Payment does not match the order' });
      }
      if (!['authorized', 'captured'].includes(payment.status)) {
        throw createError({ statusCode: 400, message: 'Payment does not match the order' });
      }
      const order = await razorpay.orders.fetch(razorpay_order_id);
      if (Number(payment.amount) !== Number(order.amount)) {
        throw createError({ statusCode: 400, message: 'Payment does not match the order' });
      }
      if (payment.currency !== order.currency) {
        throw createError({ statusCode: 400, message: 'Payment does not match the order' });
      }
      return { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id, reference: order.receipt, amount: order.amount, currency: order.currency };
    }
    throw createError({ statusCode: 400, message: 'Invalid signature' });
  } catch (error: any) {
    if (error && error.statusCode) {
      throw error;
    }
    throw createError({ statusCode: 500, message: 'Verification failed' });
  }
});
`

	composableCode := `export const useRazorpay = () => {
  const loading = ref(false);
  const scriptLoaded = ref(false);

  onMounted(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => scriptLoaded.value = true;
    document.head.appendChild(script);
  });

  const pay = async (amount, onSuccess, onError) => {
    if (!scriptLoaded.value || loading.value) return;
    loading.value = true;

    try {
      const { data } = await useFetch('/api/razorpay/order', {
        method: 'POST',

        body: { amount },
      });

      if (!data.value?.success) throw new Error(data.value?.error || 'Failed to create order');

      const options = {
        key: data.value.keyId,
        amount: data.value.amount,
        currency: data.value.currency,
        order_id: data.value.orderId,
        handler: async (response) => {
          const { data: verifyData } = await useFetch('/api/razorpay/verify', {
            method: 'POST',
            body: response,
          });
          verifyData.value?.success ? onSuccess?.(verifyData.value) : onError?.(new Error(verifyData.value?.error));
          loading.value = false;
        },
        modal: { ondismiss: () => loading.value = false },
      };

      new window.Razorpay(options).open();
    } catch (e) {
      onError?.(e);
      loading.value = false;
    }
  };

  return { pay, loading, ready: scriptLoaded };
};
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Nuxt 3 + Vue",
		Files: []FileAction{
			{Action: "create", Path: "server/api/razorpay/order.post.ts", Code: orderAPICode, Description: "Nuxt API route for orders"},
			{Action: "create", Path: "server/api/razorpay/verify.post.ts", Code: verifyAPICode, Description: "Nuxt API route for verification"},
			{Action: "create", Path: "composables/useRazorpay.ts", Code: composableCode, Description: "Vue composable for Razorpay"},
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `IMPORTANT:
1) npm install razorpay
2) Create the API routes in server/api/razorpay/
3) Create the composable in composables/
4) Add env vars to .env
5) Use the composable in your checkout page: const { pay, loading, ready } = useRazorpay()
6) SECURITY (MANDATORY): implement getOrderAmount(event) in server/api/razorpay/order.post.ts against the real cart/order model. It must return { amountInSubunits, currency, reference } from the server-side record. NEVER take the amount from the request body.
7) SECURITY (MANDATORY): mark an order paid ONLY when the verify response's 'reference' matches the order being marked AND 'amount' equals that order's own server-side total.
8) DO NOT give "Next Steps" - complete the integration.`,
	}
}
