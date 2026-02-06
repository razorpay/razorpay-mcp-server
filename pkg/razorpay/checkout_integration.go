//nolint:lll // File contains embedded code templates requiring longer lines
package razorpay

import (
	"context"
	"encoding/json"

	rzpsdk "github.com/razorpay/razorpay-go"
	"github.com/spf13/viper"

	"github.com/razorpay/razorpay-mcp-server/pkg/mcpgo"
	"github.com/razorpay/razorpay-mcp-server/pkg/observability"
)

// Credentials holds Razorpay API credentials
type Credentials struct {
	KeyID     string
	KeySecret string
}

// FileAction represents an action to perform on a file
type FileAction struct {
	Action          string     `json:"action"`
	Path            string     `json:"path"`
	Code            string     `json:"code,omitempty"`
	Description     string     `json:"description"`
	Edits           []EditItem `json:"edits,omitempty"`
	FunctionName    string     `json:"functionName,omitempty"`
	FindCode        string     `json:"findCode,omitempty"`
	ReplaceWithCode string     `json:"replaceWithCode,omitempty"`
}

// EditItem represents a manual edit instruction
type EditItem struct {
	Line string `json:"line"`
	Add  string `json:"add"`
	Why  string `json:"why"`
}

// Dependency represents a dependency to install
type Dependency struct {
	Name           string `json:"name"`
	InstallCommand string `json:"installCommand"`
}

// EnvVar represents an environment variable
type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// IntegrateCheckoutOutput is the response from integrate_razorpay_checkout
type IntegrateCheckoutOutput struct {
	Summary          string       `json:"summary"`
	Files            []FileAction `json:"files"`
	Dependencies     []Dependency `json:"dependencies"`
	EnvVars          []EnvVar     `json:"envVars"`
	TestInstructions string       `json:"testInstructions"`
	AIInstructions   string       `json:"aiInstructions"`
}

// DetectStackOutput is the response from detect_stack
type DetectStackOutput struct {
	Language       string   `json:"language"`
	Framework      string   `json:"framework"`
	Frontend       string   `json:"frontend,omitempty"`
	PackageManager string   `json:"packageManager"`
	IsFullStack    bool     `json:"isFullStack"`
	Confidence     float64  `json:"confidence"`
	Notes          []string `json:"notes"`
}

// IntegrateRazorpayCheckout returns a tool for complete Razorpay checkout integration
//nolint:gocyclo // Complex detection logic requires multiple conditions
func IntegrateRazorpayCheckout(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithString(
			"language",
			mcpgo.Description("Programming language: javascript, typescript, python, go, java, php, ruby, rust, csharp, dart, swift, or kotlin"),
			mcpgo.Required(),
			mcpgo.Enum("javascript", "typescript", "python", "go", "java", "php", "ruby", "rust", "csharp", "dart", "swift", "kotlin"),
		),
		mcpgo.WithString(
			"backendFramework",
			mcpgo.Description("Backend framework: express, fastify, koa, nextjs, nuxt, django, flask, fastapi, gin, echo, fiber, spring, laravel, rails, actix, aspnet, react-native, or flutter"),
			mcpgo.Required(),
			mcpgo.Enum("express", "fastify", "koa", "nextjs", "nuxt", "django", "flask", "fastapi", "gin", "echo", "fiber", "spring", "laravel", "rails", "actix", "aspnet", "react-native", "flutter"),
		),
		mcpgo.WithString(
			"frontendFramework",
			mcpgo.Description("Frontend framework: vanilla, react, nextjs, vue, angular, svelte, or native (for mobile apps)"),
			mcpgo.Required(),
			mcpgo.Enum("vanilla", "react", "nextjs", "vue", "angular", "svelte", "native"),
		),
		mcpgo.WithString(
			"existingOrderEndpoint",
			mcpgo.Description("Existing order creation endpoint path if any (e.g., /api/orders/create)"),
		),
		mcpgo.WithString(
			"existingPaymentFunction",
			mcpgo.Description("Existing payment/checkout function name in frontend if any"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		args, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("Invalid arguments"), nil
		}

		language, _ := args["language"].(string)
		backendFramework, _ := args["backendFramework"].(string)
		frontendFramework, _ := args["frontendFramework"].(string)

		// Get credentials from config (set via MCP config env vars)
		creds := Credentials{
			KeyID:     viper.GetString("key"),
			KeySecret: viper.GetString("secret"),
		}

		var output IntegrateCheckoutOutput

		// Get frontend code based on frontend framework
		frontendCode := getFrontendIntegration(frontendFramework, creds)

		// Route to appropriate backend integration
		switch backendFramework {
		// Python frameworks
		case "django":
			output = getDjangoIntegration(creds, frontendCode)
		case "flask":
			output = getFlaskIntegration(creds, frontendCode)
		case "fastapi":
			output = getFastAPIIntegration(creds, frontendCode)
		// Go frameworks
		case "gin":
			output = getGinIntegration(creds, frontendCode)
		case "echo":
			output = getEchoIntegration(creds, frontendCode)
		case "fiber":
			output = getFiberIntegration(creds, frontendCode)
		// Node.js frameworks
		case "nextjs":
			output = getNextjsReactIntegration(language, creds)
		case "nuxt":
			output = getNuxtIntegration(language, creds)
		case "fastify":
			output = getFastifyIntegration(language, creds, frontendCode)
		case "koa":
			output = getKoaIntegration(language, creds, frontendCode)
		// Java
		case "spring":
			output = getSpringIntegration(creds, frontendCode)
		// PHP
		case "laravel":
			output = getLaravelIntegration(creds, frontendCode)
		// Ruby
		case "rails":
			output = getRailsIntegration(creds, frontendCode)
		// Rust
		case "actix":
			output = getActixIntegration(creds, frontendCode)
		// .NET
		case "aspnet":
			output = getAspNetIntegration(creds, frontendCode)
		// Mobile
		case "react-native":
			output = getReactNativeIntegration(creds)
		case "flutter":
			output = getFlutterIntegration(creds)
		default: // express
			output = getExpressVanillaIntegration(language, creds, frontendCode)
		}

		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"integrate_razorpay_checkout",
		"Complete Razorpay Standard Checkout integration. Returns ALL code needed - "+
			"backend routes, frontend integration, and payment verification. "+
			"Use this single tool to get everything needed for Razorpay payment integration. "+
			"The AI should apply ALL returned files and modifications without asking the user for additional steps.",
		parameters,
		handler,
	)
}

// DetectStack returns a tool for detecting project technology stack
func DetectStack(
	obs *observability.Observability,
	client *rzpsdk.Client,
) mcpgo.Tool {
	parameters := []mcpgo.ToolParameter{
		mcpgo.WithArray(
			"files",
			mcpgo.Description("List of file paths in the project"),
			mcpgo.Required(),
			mcpgo.Items(map[string]interface{}{"type": "string"}),
		),
		mcpgo.WithObject(
			"packageJson",
			mcpgo.Description("Contents of package.json if it exists (Node.js)"),
		),
		mcpgo.WithString(
			"requirementsTxt",
			mcpgo.Description("Contents of requirements.txt if it exists (Python)"),
		),
		mcpgo.WithString(
			"goMod",
			mcpgo.Description("Contents of go.mod if it exists (Go)"),
		),
		mcpgo.WithString(
			"pubspecYaml",
			mcpgo.Description("Contents of pubspec.yaml if it exists (Flutter)"),
		),
		mcpgo.WithString(
			"composerJson",
			mcpgo.Description("Contents of composer.json if it exists (PHP)"),
		),
		mcpgo.WithString(
			"gemfile",
			mcpgo.Description("Contents of Gemfile if it exists (Ruby)"),
		),
		mcpgo.WithString(
			"cargoToml",
			mcpgo.Description("Contents of Cargo.toml if it exists (Rust)"),
		),
		mcpgo.WithString(
			"pomXml",
			mcpgo.Description("Contents of pom.xml if it exists (Java/Maven)"),
		),
		mcpgo.WithString(
			"csproj",
			mcpgo.Description("Contents of .csproj if it exists (.NET)"),
		),
	}

	handler := func(
		ctx context.Context,
		r mcpgo.CallToolRequest,
	) (*mcpgo.ToolResult, error) {
		args, ok := r.Arguments.(map[string]interface{})
		if !ok {
			return mcpgo.NewToolResultError("Invalid arguments"), nil
		}

		output := detectProjectStack(args)
		return mcpgo.NewToolResultJSON(output)
	}

	return mcpgo.NewTool(
		"detect_stack",
		"Detect the technology stack of a project based on file information. "+
			"Returns language, framework, frontend framework, and package manager. "+
			"Use this to determine which integration approach to use.",
		parameters,
		handler,
	)
}

// =============================================================================
// EXPRESS + VANILLA JS INTEGRATION
// =============================================================================

// FrontendIntegration holds frontend code for different frameworks
type FrontendIntegration struct {
	Framework   string
	Code        string
	FileName    string
	ScriptTag   string
	Description string
}

func getExpressVanillaIntegration(language string, creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	ext := "js"
	if language == "typescript" {
		ext = "ts"
	}

	// Use actual keys if provided, otherwise use placeholders
	keyID := creds.KeyID
	keySecret := creds.KeySecret
	if keyID == "" {
		keyID = "rzp_test_YOUR_KEY_ID"
	}
	if keySecret == "" {
		keySecret = "YOUR_KEY_SECRET"
	}

	razorpayRoutesCode := `const express = require('express');
const Razorpay = require('razorpay');
const crypto = require('crypto');

const router = express.Router();

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

// Create Razorpay Order
router.post('/order', async (req, res) => {
  try {
    const { amount, currency = 'INR', receipt } = req.body;

    if (!amount || amount <= 0) {
      return res.status(400).json({ success: false, error: 'Invalid amount' });
    }

    const order = await razorpay.orders.create({
      amount: Math.round(amount * 100), // Convert to paise
      currency,
      receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
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
router.post('/verify', (req, res) => {
  try {
    const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = req.body;

    if (!razorpay_order_id || !razorpay_payment_id || !razorpay_signature) {
      return res.status(400).json({ success: false, error: 'Missing payment details' });
    }

    const expectedSignature = crypto
      .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
      .update(razorpay_order_id + '|' + razorpay_payment_id)
      .digest('hex');

    if (crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpay_signature))) {
      res.json({
        success: true,
        message: 'Payment verified successfully',
        paymentId: razorpay_payment_id,
        orderId: razorpay_order_id,
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
     );
   }

COMMON MISTAKES TO AVOID:
- Adding to app.js when checkout.html uses checkout.js
- Adding script to index.html when checkout is in checkout.html
- Creating new functions instead of modifying existing ones
- Leaving the original COD/placeholder code active
- NOT SAVING order data before payment (causes "order data not found" errors)
- Trying to access form fields in success callback (form may be gone/reset)`,
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
		AIInstructions: `CRITICAL INSTRUCTIONS - THE INTEGRATION IS NOT COMPLETE UNTIL ALL STEPS ARE DONE:

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

5) Edit server.js (or main server file) - IMPORTANT ORDER:
   a) Add at TOP (line 1): require('dotenv').config();
   b) Add after dotenv, with other requires: const razorpayRoutes = require('./routes/razorpay');
   c) Add LATER with middleware: app.use('/api/razorpay', razorpayRoutes);

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

THE INTEGRATION IS **NOT COMPLETE** IF:
- You modified a JS file that isn't loaded by the checkout HTML
- The checkout still creates orders without calling initiateRazorpayPayment()
- The razorpay.js script isn't in the checkout HTML
- You tell the user to "wire up the payment" as a next step

DO NOT give "Next Steps" - complete EVERYTHING including discovering the correct files.`,
	}
}

// =============================================================================
// NEXT.JS + REACT INTEGRATION
// =============================================================================

func getNextjsReactIntegration(language string, creds Credentials) IntegrateCheckoutOutput {
	// Use actual keys if provided, otherwise use placeholders
	keyID := creds.KeyID
	keySecret := creds.KeySecret
	if keyID == "" {
		keyID = "rzp_test_YOUR_KEY_ID"
	}
	if keySecret == "" {
		keySecret = "YOUR_KEY_SECRET"
	}

	orderRouteCode := `import { NextRequest, NextResponse } from 'next/server';
import Razorpay from 'razorpay';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID!,
  key_secret: process.env.RAZORPAY_KEY_SECRET!,
});

export async function POST(request: NextRequest) {
  try {
    const { amount, currency = 'INR', receipt } = await request.json();

    if (!amount || amount <= 0) {
      return NextResponse.json({ success: false, error: 'Invalid amount' }, { status: 400 });
    }

    const order = await razorpay.orders.create({
      amount: Math.round(amount * 100),
      currency,
      receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
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
import crypto from 'crypto';

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

    const isValid = crypto.timingSafeEqual(
      Buffer.from(expectedSignature),
      Buffer.from(razorpay_signature)
    );

    if (isValid) {
      return NextResponse.json({
        success: true,
        message: 'Payment verified',
        paymentId: razorpay_payment_id,
        orderId: razorpay_order_id,
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
		AIInstructions: `IMPORTANT:
1) Install razorpay package
2) Create ALL the API route files
3) Create the RazorpayCheckout component
4) Add env vars to .env.local
5) Do NOT ask the user to do anything or give "Next Steps" - the integration must be complete and working.`,
	}
}

// =============================================================================
// FRONTEND INTEGRATIONS
// =============================================================================

func getFrontendIntegration(framework string, creds Credentials) FrontendIntegration {
	switch framework {
	case "react":
		return getReactFrontend()
	case "vue":
		return getVueFrontend()
	case "angular":
		return getAngularFrontend()
	case "svelte":
		return getSvelteFrontend()
	case "native":
		return FrontendIntegration{
			Framework:   "Native Mobile",
			Code:        "// See mobile-specific integration",
			FileName:    "N/A",
			ScriptTag:   "Use native SDK",
			Description: "Native mobile integration",
		}
	default: // vanilla
		return getVanillaFrontend()
	}
}

func getVanillaFrontend() FrontendIntegration {
	code := `// Razorpay Payment Integration
async function initiateRazorpayPayment(amount, onSuccess, onError) {
  try {
    if (!window.Razorpay) {
      await new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = 'https://checkout.razorpay.com/v1/checkout.js';
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
      });
    }

    const orderResponse = await fetch('/api/razorpay/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount }),
    });

    const orderData = await orderResponse.json();
    if (!orderData.success) throw new Error(orderData.error || 'Failed to create order');

    const options = {
      key: orderData.keyId,
      amount: orderData.amount,
      currency: orderData.currency,
      name: document.title || 'Payment',
      order_id: orderData.orderId,
      handler: async function(response) {
        const verifyResponse = await fetch('/api/razorpay/verify', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(response),
        });
        const verifyData = await verifyResponse.json();
        if (verifyData.success) { if (onSuccess) onSuccess(verifyData); }
        else { if (onError) onError(new Error(verifyData.error)); }
      },
      modal: { ondismiss: () => { if (onError) onError(new Error('Payment cancelled')); } },
      theme: { color: '#528FF0' },
    };

    const razorpay = new window.Razorpay(options);
    razorpay.on('payment.failed', (r) => { if (onError) onError(new Error(r.error.description)); });
    razorpay.open();
  } catch (error) {
    console.error('Payment failed:', error);
    if (onError) onError(error);
  }
}
`
	return FrontendIntegration{
		Framework:   "Vanilla JS",
		Code:        code,
		FileName:    "public/js/razorpay.js",
		ScriptTag:   "Add <script src=\"/js/razorpay.js\"></script> to the CHECKOUT HTML file (find which HTML has the checkout - may be checkout.html, cart.html, NOT just index.html)",
		Description: "Vanilla JS Razorpay payment helper",
	}
}

func getReactFrontend() FrontendIntegration {
	code := `import { useState, useEffect } from 'react';

export function useRazorpay() {
  const [loading, setLoading] = useState(false);
  const [scriptLoaded, setScriptLoaded] = useState(false);

  useEffect(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => setScriptLoaded(true);
    document.body.appendChild(script);
    return () => document.body.removeChild(script);
  }, []);

  const pay = async (amount, onSuccess, onError) => {
    if (!scriptLoaded || loading) return;
    setLoading(true);
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? onSuccess?.(result) : onError?.(new Error(result.error));
          setLoading(false);
        },
        modal: { ondismiss: () => setLoading(false) },
      };
      new window.Razorpay(options).open();
    } catch (e) {
      onError?.(e);
      setLoading(false);
    }
  };

  return { pay, loading, ready: scriptLoaded };
}

export function RazorpayButton({ amount, onSuccess, onError, children }) {
  const { pay, loading, ready } = useRazorpay();
  return (
    <button onClick={() => pay(amount, onSuccess, onError)} disabled={!ready || loading}>
      {loading ? 'Processing...' : children || 'Pay Now'}
    </button>
  );
}
`
	return FrontendIntegration{
		Framework:   "React",
		Code:        code,
		FileName:    "src/components/RazorpayButton.jsx",
		ScriptTag:   "Import and use <RazorpayButton amount={100} onSuccess={...} />",
		Description: "React hook and component for Razorpay payments",
	}
}

func getVueFrontend() FrontendIntegration {
	code := `<template>
  <button @click="pay" :disabled="!ready || loading">
    {{ loading ? 'Processing...' : 'Pay Now' }}
  </button>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const props = defineProps({ amount: Number });
const emit = defineEmits(['success', 'error']);

const loading = ref(false);
const ready = ref(false);

onMounted(() => {
  const script = document.createElement('script');
  script.src = 'https://checkout.razorpay.com/v1/checkout.js';
  script.onload = () => ready.value = true;
  document.head.appendChild(script);
});

const pay = async () => {
  if (!ready.value || loading.value) return;
  loading.value = true;
  try {
    const res = await fetch('/api/razorpay/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount: props.amount }),
    });
    const data = await res.json();
    if (!data.success) throw new Error(data.error);

    const options = {
      key: data.keyId,
      amount: data.amount,
      currency: data.currency,
      order_id: data.orderId,
      handler: async (response) => {
        const verify = await fetch('/api/razorpay/verify', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(response),
        });
        const result = await verify.json();
        result.success ? emit('success', result) : emit('error', new Error(result.error));
        loading.value = false;
      },
      modal: { ondismiss: () => loading.value = false },
    };
    new window.Razorpay(options).open();
  } catch (e) {
    emit('error', e);
    loading.value = false;
  }
};
</script>
`
	return FrontendIntegration{
		Framework:   "Vue",
		Code:        code,
		FileName:    "src/components/RazorpayButton.vue",
		ScriptTag:   "Import and use <RazorpayButton :amount=\"100\" @success=\"...\" />",
		Description: "Vue 3 component for Razorpay payments",
	}
}

func getAngularFrontend() FrontendIntegration {
	code := `import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

declare var Razorpay: any;

@Component({
  selector: 'app-razorpay-button',
  template: ` + "`" + `
    <button (click)="pay()" [disabled]="!ready || loading">
      {{ loading ? 'Processing...' : 'Pay Now' }}
    </button>
  ` + "`" + `,
})
export class RazorpayButtonComponent implements OnInit {
  @Input() amount: number = 0;
  @Output() success = new EventEmitter<any>();
  @Output() error = new EventEmitter<Error>();

  loading = false;
  ready = false;

  ngOnInit() {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => this.ready = true;
    document.head.appendChild(script);
  }

  async pay() {
    if (!this.ready || this.loading) return;
    this.loading = true;
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount: this.amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response: any) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? this.success.emit(result) : this.error.emit(new Error(result.error));
          this.loading = false;
        },
        modal: { ondismiss: () => this.loading = false },
      };
      new Razorpay(options).open();
    } catch (e) {
      this.error.emit(e as Error);
      this.loading = false;
    }
  }
}
`
	return FrontendIntegration{
		Framework:   "Angular",
		Code:        code,
		FileName:    "src/app/components/razorpay-button.component.ts",
		ScriptTag:   "Add to module declarations and use <app-razorpay-button [amount]=\"100\" (success)=\"...\">",
		Description: "Angular component for Razorpay payments",
	}
}

func getSvelteFrontend() FrontendIntegration {
	code := `<script>
  import { onMount } from 'svelte';
  export let amount = 0;

  let loading = false;
  let ready = false;

  onMount(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => ready = true;
    document.head.appendChild(script);
  });

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  async function pay() {
    if (!ready || loading) return;
    loading = true;
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? dispatch('success', result) : dispatch('error', new Error(result.error));
          loading = false;
        },
        modal: { ondismiss: () => loading = false },
      };
      new window.Razorpay(options).open();
    } catch (e) {
      dispatch('error', e);
      loading = false;
    }
  }
</script>

<button on:click={pay} disabled={!ready || loading}>
  {loading ? 'Processing...' : 'Pay Now'}
</button>
`
	return FrontendIntegration{
		Framework:   "Svelte",
		Code:        code,
		FileName:    "src/components/RazorpayButton.svelte",
		ScriptTag:   "Import and use <RazorpayButton amount={100} on:success={...} />",
		Description: "Svelte component for Razorpay payments",
	}
}

// =============================================================================
// PYTHON BACKEND INTEGRATIONS
// =============================================================================

func getDjangoIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	viewsCode := `import json
import time
import razorpay
import hmac
import hashlib
from django.http import JsonResponse
from django.views.decorators.csrf import csrf_exempt
from django.views.decorators.http import require_POST
from django.conf import settings

client = razorpay.Client(auth=(settings.RAZORPAY_KEY_ID, settings.RAZORPAY_KEY_SECRET))

@csrf_exempt
@require_POST
def create_order(request):
    try:
        data = json.loads(request.body)
        amount = data.get('amount', 0)

        if amount <= 0:
            return JsonResponse({'success': False, 'error': 'Invalid amount'}, status=400)

        order = client.order.create({
            'amount': int(amount * 100),  # Convert to paise
            'currency': data.get('currency', 'INR'),
            'receipt': data.get('receipt', f'receipt_{int(time.time())}'),
        })

        return JsonResponse({
            'success': True,
            'orderId': order['id'],
            'amount': order['amount'],
            'currency': order['currency'],
            'keyId': settings.RAZORPAY_KEY_ID,
        })
    except Exception as e:
        return JsonResponse({'success': False, 'error': str(e)}, status=500)

@csrf_exempt
@require_POST
def verify_payment(request):
    try:
        data = json.loads(request.body)
        razorpay_order_id = data.get('razorpay_order_id')
        razorpay_payment_id = data.get('razorpay_payment_id')
        razorpay_signature = data.get('razorpay_signature')

        if not all([razorpay_order_id, razorpay_payment_id, razorpay_signature]):
            return JsonResponse({'success': False, 'error': 'Missing payment details'}, status=400)

        msg = f'{razorpay_order_id}|{razorpay_payment_id}'
        expected_signature = hmac.new(
            settings.RAZORPAY_KEY_SECRET.encode(),
            msg.encode(),
            hashlib.sha256
        ).hexdigest()

        if hmac.compare_digest(expected_signature, razorpay_signature):
            return JsonResponse({
                'success': True,
                'message': 'Payment verified',
                'paymentId': razorpay_payment_id,
                'orderId': razorpay_order_id,
            })
        else:
            return JsonResponse({'success': False, 'error': 'Invalid signature'}, status=400)
    except Exception as e:
        return JsonResponse({'success': False, 'error': str(e)}, status=500)
`

	urlsCode := `from django.urls import path
from . import views

urlpatterns = [
    path('order/', views.create_order, name='razorpay_order'),
    path('verify/', views.verify_payment, name='razorpay_verify'),
]
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Django + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "razorpay_payments/views.py", Code: viewsCode, Description: "Django views for Razorpay"},
			{Action: "create", Path: "razorpay_payments/urls.py", Code: urlsCode, Description: "Django URL patterns"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{
				Action:      "manual_edit",
				Path:        "settings.py",
				Description: "Add to settings.py",
				Edits: []EditItem{
					{Line: "After other settings", Add: "RAZORPAY_KEY_ID = os.environ.get('RAZORPAY_KEY_ID')", Why: "Razorpay key ID"},
					{Line: "After RAZORPAY_KEY_ID", Add: "RAZORPAY_KEY_SECRET = os.environ.get('RAZORPAY_KEY_SECRET')", Why: "Razorpay key secret"},
				},
			},
			{
				Action:      "manual_edit",
				Path:        "urls.py",
				Description: "Add to main urls.py",
				Edits: []EditItem{
					{Line: "In urlpatterns", Add: "path('api/razorpay/', include('razorpay_payments.urls')),", Why: "Mount Razorpay URLs"},
				},
			},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay", InstallCommand: "pip install razorpay"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV",
		AIInstructions: `BACKEND SETUP:
1) pip install razorpay
2) Create razorpay_payments app with views.py and urls.py
3) Add RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET to settings.py
4) Add import os at top of settings.py if not present
5) Include razorpay_payments.urls in main urls.py
6) Create .env with Razorpay keys` + getFrontendWiringInstructions(frontend),
	}
}

func getFlaskIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	appCode := `import os
import time
import hmac
import hashlib
import razorpay
from flask import Flask, request, jsonify
from dotenv import load_dotenv

load_dotenv()

app = Flask(__name__)
client = razorpay.Client(auth=(os.environ['RAZORPAY_KEY_ID'], os.environ['RAZORPAY_KEY_SECRET']))

@app.route('/api/razorpay/order', methods=['POST'])
def create_order():
    try:
        data = request.get_json()
        amount = data.get('amount', 0)

        if amount <= 0:
            return jsonify({'success': False, 'error': 'Invalid amount'}), 400

        order = client.order.create({
            'amount': int(amount * 100),
            'currency': data.get('currency', 'INR'),
            'receipt': data.get('receipt', f'receipt_{int(time.time())}'),
        })

        return jsonify({
            'success': True,
            'orderId': order['id'],
            'amount': order['amount'],
            'currency': order['currency'],
            'keyId': os.environ['RAZORPAY_KEY_ID'],
        })
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)}), 500

@app.route('/api/razorpay/verify', methods=['POST'])
def verify_payment():
    try:
        data = request.get_json()
        razorpay_order_id = data.get('razorpay_order_id')
        razorpay_payment_id = data.get('razorpay_payment_id')
        razorpay_signature = data.get('razorpay_signature')

        if not all([razorpay_order_id, razorpay_payment_id, razorpay_signature]):
            return jsonify({'success': False, 'error': 'Missing payment details'}), 400

        msg = f'{razorpay_order_id}|{razorpay_payment_id}'
        expected = hmac.new(os.environ['RAZORPAY_KEY_SECRET'].encode(), msg.encode(), hashlib.sha256).hexdigest()

        if hmac.compare_digest(expected, razorpay_signature):
            return jsonify({'success': True, 'paymentId': razorpay_payment_id, 'orderId': razorpay_order_id})
        return jsonify({'success': False, 'error': 'Invalid signature'}), 400
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True)
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Flask + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "razorpay_routes.py", Code: appCode, Description: "Flask routes for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "pip install razorpay"},
			{Name: "python-dotenv", InstallCommand: "pip install python-dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) pip install razorpay python-dotenv
2) Create razorpay_routes.py with the Razorpay endpoints
3) Import and register the blueprint in your main app.py
4) Create .env with Razorpay keys` + getFrontendWiringInstructions(frontend),
	}
}

func getFastAPIIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	routerCode := `import os
import time
import hmac
import hashlib
import razorpay
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from dotenv import load_dotenv

load_dotenv()

router = APIRouter(prefix="/api/razorpay")
client = razorpay.Client(auth=(os.environ['RAZORPAY_KEY_ID'], os.environ['RAZORPAY_KEY_SECRET']))

class OrderRequest(BaseModel):
    amount: float
    currency: str = "INR"
    receipt: str = None

class VerifyRequest(BaseModel):
    razorpay_order_id: str
    razorpay_payment_id: str
    razorpay_signature: str

@router.post("/order")
async def create_order(req: OrderRequest):
    if req.amount <= 0:
        raise HTTPException(status_code=400, detail="Invalid amount")
    try:
        order = client.order.create({
            'amount': int(req.amount * 100),
            'currency': req.currency,
            'receipt': req.receipt or f'receipt_{int(time.time())}',
        })
        return {
            'success': True,
            'orderId': order['id'],
            'amount': order['amount'],
            'currency': order['currency'],
            'keyId': os.environ['RAZORPAY_KEY_ID'],
        }
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@router.post("/verify")
async def verify_payment(req: VerifyRequest):
    msg = f'{req.razorpay_order_id}|{req.razorpay_payment_id}'
    expected = hmac.new(os.environ['RAZORPAY_KEY_SECRET'].encode(), msg.encode(), hashlib.sha256).hexdigest()

    if hmac.compare_digest(expected, req.razorpay_signature):
        return {'success': True, 'paymentId': req.razorpay_payment_id, 'orderId': req.razorpay_order_id}
    raise HTTPException(status_code=400, detail="Invalid signature")
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for FastAPI + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "routers/razorpay.py", Code: routerCode, Description: "FastAPI router for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "main.py", Description: "Add router", Edits: []EditItem{
				{Line: "After imports", Add: "from routers.razorpay import router as razorpay_router", Why: "Import router"},
				{Line: "After app creation", Add: "app.include_router(razorpay_router)", Why: "Include router"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "pip install razorpay"},
			{Name: "python-dotenv", InstallCommand: "pip install python-dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) pip install razorpay python-dotenv
2) Create routers/razorpay.py with the Razorpay endpoints
3) Import and include router in main.py
4) Create .env with Razorpay keys` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// GO BACKEND INTEGRATIONS
// =============================================================================

func getGinIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
)

var client = razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

type OrderRequest struct {
	Amount   float64 ` + "`json:\"amount\"`" + `
	Currency string  ` + "`json:\"currency\"`" + `
	Receipt  string  ` + "`json:\"receipt\"`" + `
}

type VerifyRequest struct {
	OrderID   string ` + "`json:\"razorpay_order_id\"`" + `
	PaymentID string ` + "`json:\"razorpay_payment_id\"`" + `
	Signature string ` + "`json:\"razorpay_signature\"`" + `
}

func CreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid amount"})
		return
	}
	if req.Currency == "" {
		req.Currency = "INR"
	}
	if req.Receipt == "" {
		req.Receipt = fmt.Sprintf("receipt_%d", time.Now().Unix())
	}

	data := map[string]interface{}{
		"amount":   int(req.Amount * 100),
		"currency": req.Currency,
		"receipt":  req.Receipt,
	}
	order, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"orderId":  order["id"],
		"amount":   order["amount"],
		"currency": order["currency"],
		"keyId":    os.Getenv("RAZORPAY_KEY_ID"),
	})
}

func VerifyPayment(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	msg := req.OrderID + "|" + req.PaymentID
	h := hmac.New(sha256.New, []byte(os.Getenv("RAZORPAY_KEY_SECRET")))
	h.Write([]byte(msg))
	expected := hex.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(expected), []byte(req.Signature)) {
		c.JSON(http.StatusOK, gin.H{"success": true, "paymentId": req.PaymentID, "orderId": req.OrderID})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid signature"})
	}
}
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Gin + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "handlers/razorpay.go", Code: handlerCode, Description: "Gin handlers for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "main.go", Description: "Add routes", Edits: []EditItem{
				{Line: "In router setup", Add: "r.POST(\"/api/razorpay/order\", handlers.CreateOrder)", Why: "Order endpoint"},
				{Line: "After order route", Add: "r.POST(\"/api/razorpay/verify\", handlers.VerifyPayment)", Why: "Verify endpoint"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-go", InstallCommand: "go get github.com/razorpay/razorpay-go"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) go get github.com/razorpay/razorpay-go
2) Create handlers/razorpay.go with the Razorpay handlers
3) Add routes in main.go to wire up the handlers
4) Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET env vars` + getFrontendWiringInstructions(frontend),
	}
}

func getEchoIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	razorpay "github.com/razorpay/razorpay-go"
)

var client = razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

type OrderRequest struct {
	Amount   float64 ` + "`json:\"amount\"`" + `
	Currency string  ` + "`json:\"currency\"`" + `
	Receipt  string  ` + "`json:\"receipt\"`" + `
}

type VerifyRequest struct {
	OrderID   string ` + "`json:\"razorpay_order_id\"`" + `
	PaymentID string ` + "`json:\"razorpay_payment_id\"`" + `
	Signature string ` + "`json:\"razorpay_signature\"`" + `
}

func CreateOrder(c echo.Context) error {
	var req OrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": err.Error()})
	}
	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Invalid amount"})
	}
	if req.Currency == "" { req.Currency = "INR" }
	if req.Receipt == "" { req.Receipt = fmt.Sprintf("receipt_%d", time.Now().Unix()) }

	data := map[string]interface{}{"amount": int(req.Amount * 100), "currency": req.Currency, "receipt": req.Receipt}
	order, err := client.Order.Create(data, nil)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true, "orderId": order["id"], "amount": order["amount"],
		"currency": order["currency"], "keyId": os.Getenv("RAZORPAY_KEY_ID"),
	})
}

func VerifyPayment(c echo.Context) error {
	var req VerifyRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": err.Error()})
	}

	msg := req.OrderID + "|" + req.PaymentID
	h := hmac.New(sha256.New, []byte(os.Getenv("RAZORPAY_KEY_SECRET")))
	h.Write([]byte(msg))
	expected := hex.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(expected), []byte(req.Signature)) {
		return c.JSON(http.StatusOK, map[string]interface{}{"success": true, "paymentId": req.PaymentID, "orderId": req.OrderID})
	}
	return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Invalid signature"})
}
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Echo + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "handlers/razorpay.go", Code: handlerCode, Description: "Echo handlers for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "main.go", Description: "Add routes", Edits: []EditItem{
				{Line: "In router setup", Add: "e.POST(\"/api/razorpay/order\", handlers.CreateOrder)", Why: "Order endpoint"},
				{Line: "After order route", Add: "e.POST(\"/api/razorpay/verify\", handlers.VerifyPayment)", Why: "Verify endpoint"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-go", InstallCommand: "go get github.com/razorpay/razorpay-go"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) go get github.com/razorpay/razorpay-go
2) Create handlers/razorpay.go with the Razorpay handlers
3) Add routes in main.go to wire up the handlers
4) Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET env vars` + getFrontendWiringInstructions(frontend),
	}
}

func getFiberIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	razorpay "github.com/razorpay/razorpay-go"
)

var client = razorpay.NewClient(os.Getenv("RAZORPAY_KEY_ID"), os.Getenv("RAZORPAY_KEY_SECRET"))

type OrderRequest struct {
	Amount   float64 ` + "`json:\"amount\"`" + `
	Currency string  ` + "`json:\"currency\"`" + `
	Receipt  string  ` + "`json:\"receipt\"`" + `
}

type VerifyRequest struct {
	OrderID   string ` + "`json:\"razorpay_order_id\"`" + `
	PaymentID string ` + "`json:\"razorpay_payment_id\"`" + `
	Signature string ` + "`json:\"razorpay_signature\"`" + `
}

func CreateOrder(c *fiber.Ctx) error {
	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}
	if req.Amount <= 0 {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid amount"})
	}
	if req.Currency == "" { req.Currency = "INR" }
	if req.Receipt == "" { req.Receipt = fmt.Sprintf("receipt_%d", time.Now().Unix()) }

	data := map[string]interface{}{"amount": int(req.Amount * 100), "currency": req.Currency, "receipt": req.Receipt}
	order, err := client.Order.Create(data, nil)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"success": true, "orderId": order["id"], "amount": order["amount"],
		"currency": order["currency"], "keyId": os.Getenv("RAZORPAY_KEY_ID"),
	})
}

func VerifyPayment(c *fiber.Ctx) error {
	var req VerifyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	msg := req.OrderID + "|" + req.PaymentID
	h := hmac.New(sha256.New, []byte(os.Getenv("RAZORPAY_KEY_SECRET")))
	h.Write([]byte(msg))
	expected := hex.EncodeToString(h.Sum(nil))

	if hmac.Equal([]byte(expected), []byte(req.Signature)) {
		return c.JSON(fiber.Map{"success": true, "paymentId": req.PaymentID, "orderId": req.OrderID})
	}
	return c.Status(400).JSON(fiber.Map{"success": false, "error": "Invalid signature"})
}
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Fiber + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "handlers/razorpay.go", Code: handlerCode, Description: "Fiber handlers for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "main.go", Description: "Add routes", Edits: []EditItem{
				{Line: "In router setup", Add: "app.Post(\"/api/razorpay/order\", handlers.CreateOrder)", Why: "Order endpoint"},
				{Line: "After order route", Add: "app.Post(\"/api/razorpay/verify\", handlers.VerifyPayment)", Why: "Verify endpoint"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-go", InstallCommand: "go get github.com/razorpay/razorpay-go"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) go get github.com/razorpay/razorpay-go
2) Create handlers/razorpay.go with the Razorpay handlers
3) Add routes in main.go to wire up the handlers
4) Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET env vars` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// NODE.JS ADDITIONAL FRAMEWORKS (Fastify, Koa, Nuxt)
// =============================================================================

func getFastifyIntegration(language string, creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	routesCode := `const Razorpay = require('razorpay');
const crypto = require('crypto');

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

async function razorpayRoutes(fastify, options) {
  fastify.post('/api/razorpay/order', async (request, reply) => {
    try {
      const { amount, currency = 'INR', receipt } = request.body;

      if (!amount || amount <= 0) {
        return reply.status(400).send({ success: false, error: 'Invalid amount' });
      }

      const order = await razorpay.orders.create({
        amount: Math.round(amount * 100),
        currency,
        receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
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

      if (crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpay_signature))) {
        return { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id };
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
				{Line: "After fastify instance creation", Add: "fastify.register(require('./routes/razorpay'));", Why: "Register Razorpay routes"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
			{Name: "dotenv", InstallCommand: "npm install dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) npm install razorpay dotenv
2) Create routes/razorpay.js with Fastify routes
3) Register the routes in your main app file
4) Create .env with Razorpay keys` + getFrontendWiringInstructions(frontend),
	}
}

func getKoaIntegration(language string, creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	routerCode := `const Router = require('@koa/router');
const Razorpay = require('razorpay');
const crypto = require('crypto');

const router = new Router({ prefix: '/api/razorpay' });

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

router.post('/order', async (ctx) => {
  try {
    const { amount, currency = 'INR', receipt } = ctx.request.body;

    if (!amount || amount <= 0) {
      ctx.status = 400;
      ctx.body = { success: false, error: 'Invalid amount' };
      return;
    }

    const order = await razorpay.orders.create({
      amount: Math.round(amount * 100),
      currency,
      receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
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

    if (crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpay_signature))) {
      ctx.body = { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id };
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
				{Line: "After imports", Add: "const razorpayRouter = require('./routes/razorpay');", Why: "Import router"},
				{Line: "After app creation", Add: "app.use(razorpayRouter.routes()).use(razorpayRouter.allowedMethods());", Why: "Use router"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "npm install razorpay"},
			{Name: "@koa/router", InstallCommand: "npm install @koa/router"},
			{Name: "dotenv", InstallCommand: "npm install dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) npm install razorpay @koa/router dotenv
2) Create routes/razorpay.js with Koa router
3) Import and use the router in your main app
4) Create .env with Razorpay keys` + getFrontendWiringInstructions(frontend),
	}
}

func getNuxtIntegration(language string, creds Credentials) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	orderAPICode := `import Razorpay from 'razorpay';

const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

export default defineEventHandler(async (event) => {
  try {
    const { amount, currency = 'INR', receipt } = await readBody(event);

    if (!amount || amount <= 0) {
      throw createError({ statusCode: 400, message: 'Invalid amount' });
    }

    const order = await razorpay.orders.create({
      amount: Math.round(amount * 100),
      currency,
      receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
    });

    return { success: true, orderId: order.id, amount: order.amount, currency: order.currency, keyId: process.env.RAZORPAY_KEY_ID };
  } catch (error) {
    throw createError({ statusCode: 500, message: 'Failed to create order' });
  }
});
`

	verifyAPICode := `import crypto from 'crypto';

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

    if (crypto.timingSafeEqual(Buffer.from(expectedSignature), Buffer.from(razorpay_signature))) {
      return { success: true, paymentId: razorpay_payment_id, orderId: razorpay_order_id };
    }
    throw createError({ statusCode: 400, message: 'Invalid signature' });
  } catch (error) {
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
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `IMPORTANT:
1) npm install razorpay
2) Create the API routes in server/api/razorpay/
3) Create the composable in composables/
4) Add env vars to .env
5) Use the composable in your checkout page: const { pay, loading, ready } = useRazorpay()
6) DO NOT give "Next Steps" - complete the integration.`,
	}
}

// =============================================================================
// JAVA (Spring Boot) INTEGRATION
// =============================================================================

func getSpringIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	controllerCode := `package com.example.razorpay;

import com.razorpay.*;
import org.json.JSONObject;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/razorpay")
public class RazorpayController {

    @Value("${razorpay.key.id}")
    private String keyId;

    @Value("${razorpay.key.secret}")
    private String keySecret;

    @PostMapping("/order")
    public ResponseEntity<Map<String, Object>> createOrder(@RequestBody Map<String, Object> request) {
        Map<String, Object> response = new HashMap<>();
        try {
            RazorpayClient razorpay = new RazorpayClient(keyId, keySecret);

            double amount = ((Number) request.get("amount")).doubleValue();
            if (amount <= 0) {
                response.put("success", false);
                response.put("error", "Invalid amount");
                return ResponseEntity.badRequest().body(response);
            }

            JSONObject orderRequest = new JSONObject();
            orderRequest.put("amount", (int) (amount * 100));
            orderRequest.put("currency", request.getOrDefault("currency", "INR"));
            orderRequest.put("receipt", request.getOrDefault("receipt", "receipt_" + System.currentTimeMillis()));

            Order order = razorpay.orders.create(orderRequest);

            response.put("success", true);
            response.put("orderId", order.get("id"));
            response.put("amount", order.get("amount"));
            response.put("currency", order.get("currency"));
            response.put("keyId", keyId);
            return ResponseEntity.ok(response);
        } catch (Exception e) {
            response.put("success", false);
            response.put("error", "Failed to create order");
            return ResponseEntity.internalServerError().body(response);
        }
    }

    @PostMapping("/verify")
    public ResponseEntity<Map<String, Object>> verifyPayment(@RequestBody Map<String, String> request) {
        Map<String, Object> response = new HashMap<>();
        try {
            String orderId = request.get("razorpay_order_id");
            String paymentId = request.get("razorpay_payment_id");
            String signature = request.get("razorpay_signature");

            if (orderId == null || paymentId == null || signature == null) {
                response.put("success", false);
                response.put("error", "Missing payment details");
                return ResponseEntity.badRequest().body(response);
            }

            String data = orderId + "|" + paymentId;
            Mac sha256Hmac = Mac.getInstance("HmacSHA256");
            sha256Hmac.init(new SecretKeySpec(keySecret.getBytes(), "HmacSHA256"));
            String expectedSignature = bytesToHex(sha256Hmac.doFinal(data.getBytes()));

            if (expectedSignature.equals(signature)) {
                response.put("success", true);
                response.put("paymentId", paymentId);
                response.put("orderId", orderId);
                return ResponseEntity.ok(response);
            }
            response.put("success", false);
            response.put("error", "Invalid signature");
            return ResponseEntity.badRequest().body(response);
        } catch (Exception e) {
            response.put("success", false);
            response.put("error", "Verification failed");
            return ResponseEntity.internalServerError().body(response);
        }
    }

    private String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) sb.append(String.format("%02x", b));
        return sb.toString();
    }
}
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Spring Boot + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "src/main/java/com/example/razorpay/RazorpayController.java", Code: controllerCode, Description: "Spring controller for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "pom.xml", Description: "Add Razorpay dependency", Edits: []EditItem{
				{Line: "In dependencies section", Add: "<dependency><groupId>com.razorpay</groupId><artifactId>razorpay-java</artifactId><version>1.4.3</version></dependency>", Why: "Razorpay SDK"},
			}},
			{Action: "manual_edit", Path: "src/main/resources/application.properties", Description: "Add Razorpay config", Edits: []EditItem{
				{Line: "Add properties", Add: "razorpay.key.id=${RAZORPAY_KEY_ID}\nrazorpay.key.secret=${RAZORPAY_KEY_SECRET}", Why: "Razorpay credentials"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-java", InstallCommand: "Add to pom.xml: com.razorpay:razorpay-java:1.4.3"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) Add razorpay-java dependency to pom.xml
2) Create RazorpayController.java
3) Add config to application.properties
4) Set environment variables` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// PHP (Laravel) INTEGRATION
// =============================================================================

func getLaravelIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	controllerCode := `<?php

namespace App\Http\Controllers;

use Illuminate\Http\Request;
use Razorpay\Api\Api;

class RazorpayController extends Controller
{
    private $razorpay;

    public function __construct()
    {
        $this->razorpay = new Api(env('RAZORPAY_KEY_ID'), env('RAZORPAY_KEY_SECRET'));
    }

    public function createOrder(Request $request)
    {
        try {
            $amount = $request->input('amount', 0);

            if ($amount <= 0) {
                return response()->json(['success' => false, 'error' => 'Invalid amount'], 400);
            }

            $order = $this->razorpay->order->create([
                'amount' => (int)($amount * 100),
                'currency' => $request->input('currency', 'INR'),
                'receipt' => $request->input('receipt', 'receipt_' . time()),
            ]);

            return response()->json([
                'success' => true,
                'orderId' => $order->id,
                'amount' => $order->amount,
                'currency' => $order->currency,
                'keyId' => env('RAZORPAY_KEY_ID'),
            ]);
        } catch (\Exception $e) {
            return response()->json(['success' => false, 'error' => 'Failed to create order'], 500);
        }
    }

    public function verifyPayment(Request $request)
    {
        try {
            $orderId = $request->input('razorpay_order_id');
            $paymentId = $request->input('razorpay_payment_id');
            $signature = $request->input('razorpay_signature');

            if (!$orderId || !$paymentId || !$signature) {
                return response()->json(['success' => false, 'error' => 'Missing payment details'], 400);
            }

            $expectedSignature = hash_hmac('sha256', $orderId . '|' . $paymentId, env('RAZORPAY_KEY_SECRET'));

            if (hash_equals($expectedSignature, $signature)) {
                return response()->json([
                    'success' => true,
                    'paymentId' => $paymentId,
                    'orderId' => $orderId,
                ]);
            }

            return response()->json(['success' => false, 'error' => 'Invalid signature'], 400);
        } catch (\Exception $e) {
            return response()->json(['success' => false, 'error' => 'Verification failed'], 500);
        }
    }
}
`

	routesCode := `// Add to routes/api.php
Route::post('/razorpay/order', [App\Http\Controllers\RazorpayController::class, 'createOrder']);
Route::post('/razorpay/verify', [App\Http\Controllers\RazorpayController::class, 'verifyPayment']);
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Laravel + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "app/Http/Controllers/RazorpayController.php", Code: controllerCode, Description: "Laravel controller for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "routes/api.php", Description: "Add routes", Edits: []EditItem{
				{Line: "Add routes", Add: routesCode, Why: "Razorpay API routes"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay/razorpay", InstallCommand: "composer require razorpay/razorpay"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) composer require razorpay/razorpay
2) Create RazorpayController.php
3) Add routes to routes/api.php
4) Add keys to .env` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// RUBY (Rails) INTEGRATION
// =============================================================================

func getRailsIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	controllerCode := `class RazorpayController < ApplicationController
  skip_before_action :verify_authenticity_token

  def create_order
    begin
      amount = params[:amount].to_f

      if amount <= 0
        render json: { success: false, error: 'Invalid amount' }, status: :bad_request
        return
      end

      require 'razorpay'
      Razorpay.setup(ENV['RAZORPAY_KEY_ID'], ENV['RAZORPAY_KEY_SECRET'])

      order = Razorpay::Order.create(
        amount: (amount * 100).to_i,
        currency: params[:currency] || 'INR',
        receipt: params[:receipt] || "receipt_#{Time.now.to_i}"
      )

      render json: {
        success: true,
        orderId: order.id,
        amount: order.amount,
        currency: order.currency,
        keyId: ENV['RAZORPAY_KEY_ID']
      }
    rescue => e
      render json: { success: false, error: 'Failed to create order' }, status: :internal_server_error
    end
  end

  def verify_payment
    begin
      order_id = params[:razorpay_order_id]
      payment_id = params[:razorpay_payment_id]
      signature = params[:razorpay_signature]

      if order_id.blank? || payment_id.blank? || signature.blank?
        render json: { success: false, error: 'Missing payment details' }, status: :bad_request
        return
      end

      data = "#{order_id}|#{payment_id}"
      expected_signature = OpenSSL::HMAC.hexdigest('sha256', ENV['RAZORPAY_KEY_SECRET'], data)

      if Rack::Utils.secure_compare(expected_signature, signature)
        render json: { success: true, paymentId: payment_id, orderId: order_id }
      else
        render json: { success: false, error: 'Invalid signature' }, status: :bad_request
      end
    rescue => e
      render json: { success: false, error: 'Verification failed' }, status: :internal_server_error
    end
  end
end
`

	routesCode := `# Add to config/routes.rb
post '/api/razorpay/order', to: 'razorpay#create_order'
post '/api/razorpay/verify', to: 'razorpay#verify_payment'
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Rails + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "app/controllers/razorpay_controller.rb", Code: controllerCode, Description: "Rails controller for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "config/routes.rb", Description: "Add routes", Edits: []EditItem{
				{Line: "Inside Rails.application.routes.draw", Add: routesCode, Why: "Razorpay API routes"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay", InstallCommand: "bundle add razorpay"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) bundle add razorpay
2) Create app/controllers/razorpay_controller.rb
3) Add routes to config/routes.rb
4) Set environment variables` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// RUST (Actix-web) INTEGRATION
// =============================================================================

func getActixIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	handlerCode := `use actix_web::{web, HttpResponse, Result};
use hmac::{Hmac, Mac};
use sha2::Sha256;
use serde::{Deserialize, Serialize};
use reqwest::Client;
use std::env;

type HmacSha256 = Hmac<Sha256>;

#[derive(Deserialize)]
pub struct OrderRequest {
    amount: f64,
    currency: Option<String>,
    receipt: Option<String>,
}

#[derive(Deserialize)]
pub struct VerifyRequest {
    razorpay_order_id: String,
    razorpay_payment_id: String,
    razorpay_signature: String,
}

#[derive(Serialize)]
pub struct OrderResponse {
    success: bool,
    #[serde(rename = "orderId", skip_serializing_if = "Option::is_none")]
    order_id: Option<String>,
    amount: Option<i64>,
    currency: Option<String>,
    #[serde(rename = "keyId", skip_serializing_if = "Option::is_none")]
    key_id: Option<String>,
    error: Option<String>,
}

pub async fn create_order(req: web::Json<OrderRequest>) -> Result<HttpResponse> {
    if req.amount <= 0.0 {
        return Ok(HttpResponse::BadRequest().json(OrderResponse {
            success: false, order_id: None, amount: None, currency: None, key_id: None,
            error: Some("Invalid amount".to_string()),
        }));
    }

    let key_id = env::var("RAZORPAY_KEY_ID").unwrap();
    let key_secret = env::var("RAZORPAY_KEY_SECRET").unwrap();

    let client = Client::new();
    let amount = (req.amount * 100.0) as i64;
    let currency = req.currency.clone().unwrap_or_else(|| "INR".to_string());

    let response = client
        .post("https://api.razorpay.com/v1/orders")
        .basic_auth(&key_id, Some(&key_secret))
        .json(&serde_json::json!({
            "amount": amount,
            "currency": currency,
            "receipt": req.receipt.clone().unwrap_or_else(|| format!("receipt_{}", chrono::Utc::now().timestamp()))
        }))
        .send()
        .await;

    match response {
        Ok(res) => {
            let order: serde_json::Value = res.json().await.unwrap();
            Ok(HttpResponse::Ok().json(OrderResponse {
                success: true,
                order_id: Some(order["id"].as_str().unwrap().to_string()),
                amount: Some(order["amount"].as_i64().unwrap()),
                currency: Some(order["currency"].as_str().unwrap().to_string()),
                key_id: Some(key_id),
                error: None,
            }))
        }
        Err(_) => Ok(HttpResponse::InternalServerError().json(OrderResponse {
            success: false, order_id: None, amount: None, currency: None, key_id: None,
            error: Some("Failed to create order".to_string()),
        })),
    }
}

pub async fn verify_payment(req: web::Json<VerifyRequest>) -> Result<HttpResponse> {
    let key_secret = env::var("RAZORPAY_KEY_SECRET").unwrap();

    let data = format!("{}|{}", req.razorpay_order_id, req.razorpay_payment_id);
    let mut mac = HmacSha256::new_from_slice(key_secret.as_bytes()).unwrap();
    mac.update(data.as_bytes());
    let expected = hex::encode(mac.finalize().into_bytes());

    if expected == req.razorpay_signature {
        Ok(HttpResponse::Ok().json(serde_json::json!({
            "success": true,
            "paymentId": req.razorpay_payment_id,
            "orderId": req.razorpay_order_id
        })))
    } else {
        Ok(HttpResponse::BadRequest().json(serde_json::json!({
            "success": false,
            "error": "Invalid signature"
        })))
    }
}

// In main.rs, add routes:
// .route("/api/razorpay/order", web::post().to(handlers::razorpay::create_order))
// .route("/api/razorpay/verify", web::post().to(handlers::razorpay::verify_payment))
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Actix-web + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "src/handlers/razorpay.rs", Code: handlerCode, Description: "Actix handlers for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "Cargo.toml", Description: "Add dependencies", Edits: []EditItem{
				{Line: "In dependencies", Add: "reqwest = { version = \"0.11\", features = [\"json\"] }\nhmac = \"0.12\"\nsha2 = \"0.10\"\nhex = \"0.4\"\nchrono = \"0.4\"", Why: "Required crates"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "reqwest + hmac + sha2", InstallCommand: "Add to Cargo.toml"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) Add dependencies to Cargo.toml
2) Create src/handlers/razorpay.rs
3) Register routes in main.rs
4) Set environment variables` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// .NET (ASP.NET Core) INTEGRATION
// =============================================================================

func getAspNetIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	controllerCode := `using Microsoft.AspNetCore.Mvc;
using System.Security.Cryptography;
using System.Text;
using System.Text.Json;

namespace YourApp.Controllers;

[ApiController]
[Route("api/razorpay")]
public class RazorpayController : ControllerBase
{
    private readonly IConfiguration _config;
    private readonly HttpClient _httpClient;

    public RazorpayController(IConfiguration config, IHttpClientFactory httpClientFactory)
    {
        _config = config;
        _httpClient = httpClientFactory.CreateClient();
    }

    [HttpPost("order")]
    public async Task<IActionResult> CreateOrder([FromBody] OrderRequest request)
    {
        if (request.Amount <= 0)
            return BadRequest(new { success = false, error = "Invalid amount" });

        var keyId = _config["Razorpay:KeyId"];
        var keySecret = _config["Razorpay:KeySecret"];

        var authValue = Convert.ToBase64String(Encoding.ASCII.GetBytes($"{keyId}:{keySecret}"));
        _httpClient.DefaultRequestHeaders.Authorization = new System.Net.Http.Headers.AuthenticationHeaderValue("Basic", authValue);

        var orderData = new
        {
            amount = (int)(request.Amount * 100),
            currency = request.Currency ?? "INR",
            receipt = request.Receipt ?? $"receipt_{DateTimeOffset.UtcNow.ToUnixTimeSeconds()}"
        };

        var response = await _httpClient.PostAsJsonAsync("https://api.razorpay.com/v1/orders", orderData);
        var order = await response.Content.ReadFromJsonAsync<JsonElement>();

        return Ok(new
        {
            success = true,
            orderId = order.GetProperty("id").GetString(),
            amount = order.GetProperty("amount").GetInt64(),
            currency = order.GetProperty("currency").GetString(),
            keyId = keyId
        });
    }

    [HttpPost("verify")]
    public IActionResult VerifyPayment([FromBody] VerifyRequest request)
    {
        if (string.IsNullOrEmpty(request.RazorpayOrderId) || string.IsNullOrEmpty(request.RazorpayPaymentId) || string.IsNullOrEmpty(request.RazorpaySignature))
            return BadRequest(new { success = false, error = "Missing payment details" });

        var keySecret = _config["Razorpay:KeySecret"];
        var data = $"{request.RazorpayOrderId}|{request.RazorpayPaymentId}";

        using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(keySecret));
        var hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(data));
        var expectedSignature = BitConverter.ToString(hash).Replace("-", "").ToLower();

        if (expectedSignature == request.RazorpaySignature)
            return Ok(new { success = true, paymentId = request.RazorpayPaymentId, orderId = request.RazorpayOrderId });

        return BadRequest(new { success = false, error = "Invalid signature" });
    }
}

public record OrderRequest(double Amount, string? Currency, string? Receipt);
public record VerifyRequest(string RazorpayOrderId, string RazorpayPaymentId, string RazorpaySignature);
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for ASP.NET Core + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "Controllers/RazorpayController.cs", Code: controllerCode, Description: "ASP.NET controller for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "appsettings.json", Description: "Add Razorpay config", Edits: []EditItem{
				{Line: "In configuration", Add: "\"Razorpay\": { \"KeyId\": \"YOUR_KEY_ID\", \"KeySecret\": \"YOUR_KEY_SECRET\" }", Why: "Razorpay credentials"},
			}},
			{Action: "manual_edit", Path: "Program.cs", Description: "Add HttpClient", Edits: []EditItem{
				{Line: "Before builder.Build()", Add: "builder.Services.AddHttpClient();", Why: "Register HttpClient factory"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "HttpClient", InstallCommand: "Built-in - just add services.AddHttpClient()"}},
		EnvVars:          []EnvVar{{Name: "Razorpay__KeyId", Value: keyID}, {Name: "Razorpay__KeySecret", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111",
		AIInstructions: `BACKEND SETUP:
1) Create Controllers/RazorpayController.cs
2) Add HttpClient in Program.cs
3) Add config to appsettings.json or use environment variables
4) Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// MOBILE INTEGRATIONS (React Native, Flutter)
// =============================================================================

func getReactNativeIntegration(creds Credentials) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	paymentServiceCode := `import RazorpayCheckout from 'react-native-razorpay';

const RAZORPAY_KEY_ID = 'YOUR_RAZORPAY_KEY_ID'; // Replace or use env

export const createOrder = async (amount: number) => {
  // Call your backend API to create order
  const response = await fetch('YOUR_BACKEND_URL/api/razorpay/order', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount }),
  });
  return response.json();
};

export const verifyPayment = async (paymentData: any) => {
  const response = await fetch('YOUR_BACKEND_URL/api/razorpay/verify', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(paymentData),
  });
  return response.json();
};

export const initiatePayment = async (
  amount: number,
  onSuccess: (data: any) => void,
  onError: (error: Error) => void
) => {
  try {
    const orderData = await createOrder(amount);
    if (!orderData.success) throw new Error(orderData.error);

    const options = {
      description: 'Payment',
      currency: orderData.currency,
      key: orderData.keyId,
      amount: orderData.amount,
      name: 'Your App Name',
      order_id: orderData.orderId,
      theme: { color: '#528FF0' },
    };

    const paymentResponse = await RazorpayCheckout.open(options);
    const verifyData = await verifyPayment(paymentResponse);

    if (verifyData.success) {
      onSuccess(verifyData);
    } else {
      onError(new Error(verifyData.error));
    }
  } catch (error: any) {
    onError(error);
  }
};
`

	usageCode := `// Usage in your component:
import { initiatePayment } from './services/razorpay';

const handlePayment = () => {
  initiatePayment(
    100, // amount
    (data) => {
      console.log('Payment successful:', data);
      // Navigate to success screen
    },
    (error) => {
      console.error('Payment failed:', error);
      Alert.alert('Payment Failed', error.message);
    }
  );
};

// In your JSX:
<TouchableOpacity onPress={handlePayment}>
  <Text>Pay Now</Text>
</TouchableOpacity>
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for React Native",
		Files: []FileAction{
			{Action: "create", Path: "src/services/razorpay.ts", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.md", Code: usageCode, Description: "Usage example"},
		},
		Dependencies: []Dependency{
			{Name: "react-native-razorpay", InstallCommand: "npm install react-native-razorpay && cd ios && pod install"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret + " (BACKEND ONLY - never expose in mobile app)"}},
		TestInstructions: "Use test card: 4111 1111 1111 1111. Test UPI: success@razorpay",
		AIInstructions: `MOBILE SETUP:
1) npm install react-native-razorpay
2) iOS: cd ios && pod install
3) Android: No additional setup needed
4) Create the payment service file
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in mobile app - only use on backend
7) Import and use initiatePayment() in your checkout screen`,
	}
}

func getFlutterIntegration(creds Credentials) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	paymentServiceCode := `import 'package:razorpay_flutter/razorpay_flutter.dart';
import 'package:http/http.dart' as http;
import 'dart:convert';

class RazorpayService {
  late Razorpay _razorpay;
  Function(Map<String, dynamic>)? onSuccess;
  Function(String)? onError;

  static const String backendUrl = 'YOUR_BACKEND_URL'; // Replace with your backend
  static const String keyId = 'YOUR_RAZORPAY_KEY_ID'; // Replace or use env

  RazorpayService() {
    _razorpay = Razorpay();
    _razorpay.on(Razorpay.EVENT_PAYMENT_SUCCESS, _handlePaymentSuccess);
    _razorpay.on(Razorpay.EVENT_PAYMENT_ERROR, _handlePaymentError);
    _razorpay.on(Razorpay.EVENT_EXTERNAL_WALLET, _handleExternalWallet);
  }

  Future<Map<String, dynamic>> _createOrder(double amount) async {
    final response = await http.post(
      Uri.parse('$backendUrl/api/razorpay/order'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({'amount': amount}),
    );
    return jsonDecode(response.body);
  }

  Future<Map<String, dynamic>> _verifyPayment(Map<String, dynamic> paymentData) async {
    final response = await http.post(
      Uri.parse('$backendUrl/api/razorpay/verify'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode(paymentData),
    );
    return jsonDecode(response.body);
  }

  Future<void> initiatePayment({
    required double amount,
    required Function(Map<String, dynamic>) onPaymentSuccess,
    required Function(String) onPaymentError,
  }) async {
    onSuccess = onPaymentSuccess;
    onError = onPaymentError;

    try {
      final orderData = await _createOrder(amount);
      if (orderData['success'] != true) {
        onError?.call(orderData['error'] ?? 'Failed to create order');
        return;
      }

      var options = {
        'key': orderData['keyId'],
        'amount': orderData['amount'],
        'currency': orderData['currency'],
        'name': 'Your App Name',
        'order_id': orderData['orderId'],
        'theme': {'color': '#528FF0'},
      };

      _razorpay.open(options);
    } catch (e) {
      onError?.call(e.toString());
    }
  }

  void _handlePaymentSuccess(PaymentSuccessResponse response) async {
    try {
      final verifyData = await _verifyPayment({
        'razorpay_order_id': response.orderId,
        'razorpay_payment_id': response.paymentId,
        'razorpay_signature': response.signature,
      });

      if (verifyData['success'] == true) {
        onSuccess?.call(verifyData);
      } else {
        onError?.call(verifyData['error'] ?? 'Verification failed');
      }
    } catch (e) {
      onError?.call(e.toString());
    }
  }

  void _handlePaymentError(PaymentFailureResponse response) {
    onError?.call(response.message ?? 'Payment failed');
  }

  void _handleExternalWallet(ExternalWalletResponse response) {
    // Handle external wallet selection if needed
  }

  void dispose() {
    _razorpay.clear();
  }
}
`

	usageCode := `// Usage in your widget:
final razorpayService = RazorpayService();

void handlePayment() {
  razorpayService.initiatePayment(
    amount: 100,
    onPaymentSuccess: (data) {
      print('Payment successful: $data');
      // Navigate to success screen
    },
    onPaymentError: (error) {
      print('Payment failed: $error');
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Payment failed: $error')),
      );
    },
  );
}

// In your build method:
ElevatedButton(
  onPressed: handlePayment,
  child: Text('Pay Now'),
)

// Don't forget to dispose:
@override
void dispose() {
  razorpayService.dispose();
  super.dispose();
}
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for Flutter",
		Files: []FileAction{
			{Action: "create", Path: "lib/services/razorpay_service.dart", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.md", Code: usageCode, Description: "Usage example"},
		},
		Dependencies: []Dependency{
			{Name: "razorpay_flutter", InstallCommand: "flutter pub add razorpay_flutter http"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret + " (BACKEND ONLY - never expose in mobile app)"}},
		TestInstructions: "Use test card: 4111 1111 1111 1111. Test UPI: success@razorpay",
		AIInstructions: `MOBILE SETUP:
1) flutter pub add razorpay_flutter http
2) Android: Add proguard rules if using minification
3) iOS: No additional setup needed
4) Create lib/services/razorpay_service.dart
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in mobile app - only use on backend
7) Use RazorpayService in your checkout screen`,
	}
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
		Code: `STEP-BY-STEP DISCOVERY PROCESS:

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
     );
   }

COMMON MISTAKES TO AVOID:
- Modifying wrong file (e.g., app.js when checkout.html uses checkout.js)
- Adding script to index.html when checkout is in checkout.html
- Creating new functions instead of modifying existing checkout flow
- Leaving COD/placeholder payment code active
- NOT SAVING order data before payment (causes "order data not found" errors)
- Trying to access form fields in success callback (form may be gone/reset)`,
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

DO NOT give "Next Steps" - complete EVERYTHING including frontend wiring.`
}

// =============================================================================
// DETECT STACK HELPER
// =============================================================================

//nolint:gocyclo // Complex detection logic requires multiple conditions
func detectProjectStack(args map[string]interface{}) DetectStackOutput {
	files := []string{}
	if f, ok := args["files"].([]interface{}); ok {
		for _, v := range f {
			if s, ok := v.(string); ok {
				files = append(files, s)
			}
		}
	}

	packageJsonRaw, hasPackageJson := args["packageJson"].(map[string]interface{})
	requirementsTxt, _ := args["requirementsTxt"].(string)
	goMod, _ := args["goMod"].(string)
	pubspecYaml, _ := args["pubspecYaml"].(string)

	notes := []string{}

	// Flutter detection
	if pubspecYaml != "" || containsSuffix(files, "pubspec.yaml") {
		return DetectStackOutput{
			Language:       "dart",
			Framework:      "flutter",
			PackageManager: "pub",
			IsFullStack:    false,
			Confidence:     0.95,
			Notes:          []string{"Flutter mobile app detected"},
		}
	}

	// Go detection
	if goMod != "" || containsSuffix(files, "go.mod") {
		framework := "gin" // default to gin for Go web
		//nolint:gocritic // if-else chain is clearer here
		if contains(goMod, "github.com/labstack/echo") {
			framework = "echo"
		} else if contains(goMod, "github.com/gofiber/fiber") {
			framework = "fiber"
		}

		return DetectStackOutput{
			Language:       "go",
			Framework:      framework,
			PackageManager: "go-mod",
			IsFullStack:    true,
			Confidence:     0.9,
			Notes:          []string{"Go project with " + framework},
		}
	}

	// Rust detection
	if containsSuffix(files, "Cargo.toml") {
		framework := "actix" // default for Rust web
		return DetectStackOutput{
			Language:       "rust",
			Framework:      framework,
			PackageManager: "cargo",
			IsFullStack:    true,
			Confidence:     0.9,
			Notes:          []string{"Rust project detected"},
		}
	}

	// Java/Spring detection
	if containsSuffix(files, "pom.xml") || containsSuffix(files, "build.gradle") {
		return DetectStackOutput{
			Language:       "java",
			Framework:      "spring",
			PackageManager: "maven",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Java project detected - assuming Spring Boot"},
		}
	}

	// PHP/Laravel detection
	if containsSuffix(files, "composer.json") {
		framework := "laravel" // default for PHP
		if containsPath(files, "artisan") {
			framework = "laravel"
		}
		return DetectStackOutput{
			Language:       "php",
			Framework:      framework,
			PackageManager: "composer",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"PHP project detected"},
		}
	}

	// Ruby/Rails detection
	if containsSuffix(files, "Gemfile") {
		framework := "rails" // default for Ruby web
		if containsPath(files, "config/routes.rb") {
			framework = "rails"
		}
		return DetectStackOutput{
			Language:       "ruby",
			Framework:      framework,
			PackageManager: "bundler",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Ruby project detected"},
		}
	}

	// .NET detection
	if containsSuffix(files, ".csproj") || containsSuffix(files, ".sln") {
		return DetectStackOutput{
			Language:       "csharp",
			Framework:      "aspnet",
			PackageManager: "nuget",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{".NET project detected - assuming ASP.NET Core"},
		}
	}

	// Python detection
	if requirementsTxt != "" || containsSuffix(files, "requirements.txt") || containsSuffix(files, "pyproject.toml") {
		framework := "flask" // default
		pythonFrameworks := []string{"django", "flask", "fastapi", "starlette"}
		for _, fw := range pythonFrameworks {
			if contains(requirementsTxt, fw) {
				framework = fw
				break
			}
		}

		if containsPath(files, "manage.py") {
			framework = "django"
		}
		if containsSuffix(files, "app.py") && framework == "flask" {
			framework = "flask"
		}

		return DetectStackOutput{
			Language:       "python",
			Framework:      framework,
			PackageManager: "pip",
			IsFullStack:    true,
			Confidence:     0.85,
			Notes:          []string{"Python project with " + framework},
		}
	}

	// Node.js detection
	if hasPackageJson || containsSuffix(files, "package.json") {
		deps := map[string]bool{}
		if packageJsonRaw != nil {
			if d, ok := packageJsonRaw["dependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}
			if d, ok := packageJsonRaw["devDependencies"].(map[string]interface{}); ok {
				for k := range d {
					deps[k] = true
				}
			}
		}

		isTypeScript := containsSuffix(files, ".ts") || containsSuffix(files, ".tsx") || deps["typescript"]
		language := "javascript"
		if isTypeScript {
			language = "typescript"
		}

		// Detect package manager
		packageManager := "npm"
		//nolint:gocritic // if-else chain is clearer here
		if containsPath(files, "yarn.lock") {
			packageManager = "yarn"
		} else if containsPath(files, "pnpm-lock.yaml") {
			packageManager = "pnpm"
		} else if containsPath(files, "bun.lockb") {
			packageManager = "bun"
		}

		// Detect backend framework
		framework := "express" // default for Node
		nodeFrameworks := map[string]string{
			"next":         "nextjs",
			"express":      "express",
			"fastify":      "fastify",
			"koa":          "koa",
			"hono":         "hono",
			"nuxt":         "nuxt",
			"@nestjs/core": "nestjs",
		}
		for pkg, fw := range nodeFrameworks {
			if deps[pkg] {
				framework = fw
				notes = append(notes, "Found "+pkg+" in dependencies")
				break
			}
		}

		// Detect frontend framework
		frontend := ""
		frontendFrameworks := map[string]string{
			"react":         "react",
			"vue":           "vue",
			"@angular/core": "angular",
			"svelte":        "svelte",
			"solid-js":      "solid",
			"react-native":  "react-native",
			"expo":          "react-native",
		}
		for pkg, fw := range frontendFrameworks {
			if deps[pkg] {
				frontend = fw
				notes = append(notes, "Found "+pkg+" for frontend")
				break
			}
		}

		// React Native special case
		if frontend == "react-native" {
			return DetectStackOutput{
				Language:       language,
				Framework:      "react-native",
				PackageManager: packageManager,
				IsFullStack:    false,
				Confidence:     0.95,
				Notes:          []string{"React Native mobile app detected"},
			}
		}

		// Determine if fullstack
		isFullStack := framework == "nextjs" || framework == "nuxt" || framework == "nestjs" ||
			(framework != "node" && frontend == "")

		return DetectStackOutput{
			Language:       language,
			Framework:      framework,
			Frontend:       frontend,
			PackageManager: packageManager,
			IsFullStack:    isFullStack,
			Confidence:     0.9,
			Notes:          notes,
		}
	}

	// Default fallback
	return DetectStackOutput{
		Language:       "unknown",
		Framework:      "unknown",
		PackageManager: "unknown",
		IsFullStack:    false,
		Confidence:     0.1,
		Notes:          []string{"Could not detect project stack"},
	}
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
