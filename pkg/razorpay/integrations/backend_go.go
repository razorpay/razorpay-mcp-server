//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

// =============================================================================
// GO BACKEND INTEGRATIONS
// =============================================================================

func getGinIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"

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

// OrderAmount is what the server decided this cart costs.
type OrderAmount struct {
	AmountInSubunits int64
	Currency         string
	Reference        string
}

// getOrderAmount returns what the server says this cart costs. The price MUST
// come from the server: read it from the cart or order record this project
// already has. Never read it from the request body. The Razorpay signature
// covers the order id and the payment id, not the price, so a shopper who
// chooses the amount can pay 1 paise for any cart and still produce a
// signature that verifies. Replace the body with a real lookup.
func getOrderAmount(c *gin.Context) (OrderAmount, error) {
	return OrderAmount{}, errors.New("getOrderAmount must read the amount from the server-side cart or order record; see the integration instructions")
}

func toInt64(value interface{}) (int64, bool) {
	number, ok := value.(float64)
	if !ok {
		return 0, false
	}
	return int64(number), true
}

func CreateOrder(c *gin.Context) {
	var req OrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// The request body fields are deliberately ignored. The server, not the
	// shopper, decides what this cart costs.
	amount, err := getOrderAmount(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	data := map[string]interface{}{
		"amount":   amount.AmountInSubunits,
		"currency": amount.Currency,
		"receipt":  amount.Reference,
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
		// The signature proves only that Razorpay issued this payment for
		// some order. Bind the payment to this order and check the amounts
		// before trusting it.
		payment, err := client.Payment.Fetch(req.PaymentID, nil, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		paymentOrderID, _ := payment["order_id"].(string)
		if paymentOrderID != req.OrderID {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Payment does not match the order"})
			return
		}
		paymentStatus, _ := payment["status"].(string)
		if paymentStatus != "authorized" && paymentStatus != "captured" {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Payment does not match the order"})
			return
		}
		order, err := client.Order.Fetch(req.OrderID, nil, nil)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		paymentAmount, okPaymentAmount := toInt64(payment["amount"])
		orderAmount, okOrderAmount := toInt64(order["amount"])
		paymentCurrency, _ := payment["currency"].(string)
		orderCurrency, _ := order["currency"].(string)
		if !okPaymentAmount || !okOrderAmount || paymentAmount != orderAmount || paymentCurrency != orderCurrency {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Payment does not match the order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"paymentId": req.PaymentID,
			"orderId":   req.OrderID,
			"reference": order["receipt"],
			"amount":    paymentAmount,
			"currency":  paymentCurrency,
		})
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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints
- You MUST implement getOrderAmount against the real cart or order model of this project, reading the amount from the server-side record. NEVER take the amount from the request body - a shopper who controls the amount can pay 1 paise for any cart.
- You MUST mark an order as paid only when the verify response "reference" matches the order being marked AND "amount" equals that order's own server-side total. Never mark an order paid on signature success alone.

**STEP 3: WIRE UP ROUTES**
- Find main.go and add routes at the TOP LEVEL in main() where other routes are defined
- Add r.POST("/api/razorpay/order", handlers.CreateOrder)
- Add r.POST("/api/razorpay/verify", handlers.VerifyPayment)
- These MUST be at the same level as other r.GET/r.POST calls, NOT inside another handler

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security
❌ DO NOT take the payment amount from the request body - the amount MUST come from the server-side cart or order record
❌ DO NOT mark an order paid unless the verify response "reference" and "amount" match that order's own server-side record

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you implement getOrderAmount to read the amount from your server-side cart or order record?
✅ Did you add routes at TOP LEVEL in main.go (not inside another handler)?
✅ Did you create .env with credentials?
✅ Do you mark an order paid only when the verify response "reference" and "amount" match that order?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// ECHO INTEGRATION
// =============================================================================

func getEchoIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"os"

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

// OrderAmount is what the server decided this cart costs.
type OrderAmount struct {
	AmountInSubunits int64
	Currency         string
	Reference        string
}

// getOrderAmount returns what the server says this cart costs. The price MUST
// come from the server: read it from the cart or order record this project
// already has. Never read it from the request body. The Razorpay signature
// covers the order id and the payment id, not the price, so a shopper who
// chooses the amount can pay 1 paise for any cart and still produce a
// signature that verifies. Replace the body with a real lookup.
func getOrderAmount(c echo.Context) (OrderAmount, error) {
	return OrderAmount{}, errors.New("getOrderAmount must read the amount from the server-side cart or order record; see the integration instructions")
}

func toInt64(value interface{}) (int64, bool) {
	number, ok := value.(float64)
	if !ok {
		return 0, false
	}
	return int64(number), true
}

func CreateOrder(c echo.Context) error {
	var req OrderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": err.Error()})
	}

	// The request body fields are deliberately ignored. The server, not the
	// shopper, decides what this cart costs.
	amount, err := getOrderAmount(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"success": false, "error": err.Error()})
	}

	data := map[string]interface{}{"amount": amount.AmountInSubunits, "currency": amount.Currency, "receipt": amount.Reference}
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
		// The signature proves only that Razorpay issued this payment for
		// some order. Bind the payment to this order and check the amounts
		// before trusting it.
		payment, err := client.Payment.Fetch(req.PaymentID, nil, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": err.Error()})
		}
		paymentOrderID, _ := payment["order_id"].(string)
		if paymentOrderID != req.OrderID {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Payment does not match the order"})
		}
		paymentStatus, _ := payment["status"].(string)
		if paymentStatus != "authorized" && paymentStatus != "captured" {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Payment does not match the order"})
		}
		order, err := client.Order.Fetch(req.OrderID, nil, nil)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": err.Error()})
		}
		paymentAmount, okPaymentAmount := toInt64(payment["amount"])
		orderAmount, okOrderAmount := toInt64(order["amount"])
		paymentCurrency, _ := payment["currency"].(string)
		orderCurrency, _ := order["currency"].(string)
		if !okPaymentAmount || !okOrderAmount || paymentAmount != orderAmount || paymentCurrency != orderCurrency {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{"success": false, "error": "Payment does not match the order"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true, "paymentId": req.PaymentID, "orderId": req.OrderID,
			"reference": order["receipt"], "amount": paymentAmount, "currency": paymentCurrency,
		})
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
				{Line: "In the TOP-LEVEL router setup in main() (where other e.GET/e.POST routes are defined) - NEVER inside another handler function", Add: "e.POST(\"/api/razorpay/order\", handlers.CreateOrder)", Why: "Order endpoint - must be top-level in main(), not nested inside another handler"},
				{Line: "Immediately after the order route", Add: "e.POST(\"/api/razorpay/verify\", handlers.VerifyPayment)", Why: "Verify endpoint"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-go", InstallCommand: "go get github.com/razorpay/razorpay-go"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints
- You MUST implement getOrderAmount against the real cart or order model of this project, reading the amount from the server-side record. NEVER take the amount from the request body - a shopper who controls the amount can pay 1 paise for any cart.
- You MUST mark an order as paid only when the verify response "reference" matches the order being marked AND "amount" equals that order's own server-side total. Never mark an order paid on signature success alone.

**STEP 3: WIRE UP ROUTES**
- Find main.go and add routes at the TOP LEVEL in main() where other routes are defined
- Add e.POST("/api/razorpay/order", handlers.CreateOrder)
- Add e.POST("/api/razorpay/verify", handlers.VerifyPayment)
- These MUST be at the same level as other e.GET/e.POST calls, NOT inside another handler

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security
❌ DO NOT take the payment amount from the request body - the amount MUST come from the server-side cart or order record
❌ DO NOT mark an order paid unless the verify response "reference" and "amount" match that order's own server-side record

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you implement getOrderAmount to read the amount from your server-side cart or order record?
✅ Did you add routes at TOP LEVEL in main.go (not inside another handler)?
✅ Did you create .env with credentials?
✅ Do you mark an order paid only when the verify response "reference" and "amount" match that order?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

func getFiberIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	handlerCode := `package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"

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

// OrderAmount is what the server decided this cart costs.
type OrderAmount struct {
	AmountInSubunits int64
	Currency         string
	Reference        string
}

// getOrderAmount returns what the server says this cart costs. The price MUST
// come from the server: read it from the cart or order record this project
// already has. Never read it from the request body. The Razorpay signature
// covers the order id and the payment id, not the price, so a shopper who
// chooses the amount can pay 1 paise for any cart and still produce a
// signature that verifies. Replace the body with a real lookup.
func getOrderAmount(c *fiber.Ctx) (OrderAmount, error) {
	return OrderAmount{}, errors.New("getOrderAmount must read the amount from the server-side cart or order record; see the integration instructions")
}

func toInt64(value interface{}) (int64, bool) {
	number, ok := value.(float64)
	if !ok {
		return 0, false
	}
	return int64(number), true
}

func CreateOrder(c *fiber.Ctx) error {
	var req OrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	// The request body fields are deliberately ignored. The server, not the
	// shopper, decides what this cart costs.
	amount, err := getOrderAmount(c)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"success": false, "error": err.Error()})
	}

	data := map[string]interface{}{"amount": amount.AmountInSubunits, "currency": amount.Currency, "receipt": amount.Reference}
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
		// The signature proves only that Razorpay issued this payment for
		// some order. Bind the payment to this order and check the amounts
		// before trusting it.
		payment, err := client.Payment.Fetch(req.PaymentID, nil, nil)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
		}
		paymentOrderID, _ := payment["order_id"].(string)
		if paymentOrderID != req.OrderID {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Payment does not match the order"})
		}
		paymentStatus, _ := payment["status"].(string)
		if paymentStatus != "authorized" && paymentStatus != "captured" {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Payment does not match the order"})
		}
		order, err := client.Order.Fetch(req.OrderID, nil, nil)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": err.Error()})
		}
		paymentAmount, okPaymentAmount := toInt64(payment["amount"])
		orderAmount, okOrderAmount := toInt64(order["amount"])
		paymentCurrency, _ := payment["currency"].(string)
		orderCurrency, _ := order["currency"].(string)
		if !okPaymentAmount || !okOrderAmount || paymentAmount != orderAmount || paymentCurrency != orderCurrency {
			return c.Status(400).JSON(fiber.Map{"success": false, "error": "Payment does not match the order"})
		}

		return c.JSON(fiber.Map{
			"success": true, "paymentId": req.PaymentID, "orderId": req.OrderID,
			"reference": order["receipt"], "amount": paymentAmount, "currency": paymentCurrency,
		})
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
				{Line: "In the TOP-LEVEL router setup in main() (where other app.Get/app.Post routes are defined) - NEVER inside another handler function", Add: "app.Post(\"/api/razorpay/order\", handlers.CreateOrder)", Why: "Order endpoint - must be top-level in main(), not nested inside another handler"},
				{Line: "Immediately after the order route", Add: "app.Post(\"/api/razorpay/verify\", handlers.VerifyPayment)", Why: "Verify endpoint"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-go", InstallCommand: "go get github.com/razorpay/razorpay-go"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints
- You MUST implement getOrderAmount against the real cart or order model of this project, reading the amount from the server-side record. NEVER take the amount from the request body - a shopper who controls the amount can pay 1 paise for any cart.
- You MUST mark an order as paid only when the verify response "reference" matches the order being marked AND "amount" equals that order's own server-side total. Never mark an order paid on signature success alone.

**STEP 3: WIRE UP ROUTES**
- Find main.go and add routes at the TOP LEVEL in main() where other routes are defined
- Add app.Post("/api/razorpay/order", handlers.CreateOrder)
- Add app.Post("/api/razorpay/verify", handlers.VerifyPayment)
- These MUST be at the same level as other app.Get/app.Post calls, NOT inside another handler

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security
❌ DO NOT take the payment amount from the request body - the amount MUST come from the server-side cart or order record
❌ DO NOT mark an order paid unless the verify response "reference" and "amount" match that order's own server-side record

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you implement getOrderAmount to read the amount from your server-side cart or order record?
✅ Did you add routes at TOP LEVEL in main.go (not inside another handler)?
✅ Did you create .env with credentials?
✅ Do you mark an order paid only when the verify response "reference" and "amount" match that order?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}
