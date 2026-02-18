//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints

**STEP 3: WIRE UP ROUTES**
- Find main.go and add the Razorpay routes to the Gin router
- Add r.POST("/api/razorpay/order", handlers.CreateOrder)
- Add r.POST("/api/razorpay/verify", handlers.VerifyPayment)

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you add routes in main.go?
✅ Did you create .env with credentials?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// ECHO INTEGRATION
// =============================================================================

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints

**STEP 3: WIRE UP ROUTES**
- Find main.go and add the Razorpay routes to the Echo instance
- Add e.POST("/api/razorpay/order", handlers.CreateOrder)
- Add e.POST("/api/razorpay/verify", handlers.VerifyPayment)

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you add routes in main.go?
✅ Did you create .env with credentials?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: go get github.com/razorpay/razorpay-go

**STEP 2: CREATE HANDLERS**
- Create handlers/razorpay.go with the provided code
- This includes order creation and payment verification endpoints

**STEP 3: WIRE UP ROUTES**
- Find main.go and add the Razorpay routes to the Fiber app
- Add app.Post("/api/razorpay/order", handlers.CreateOrder)
- Add app.Post("/api/razorpay/verify", handlers.VerifyPayment)

**STEP 4: SET UP ENVIRONMENT**
- Create .env file with RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET
- Ensure your app loads env vars (use godotenv or os.Getenv)

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification - it is critical for security

FINAL CHECKLIST:
✅ Did you run go get github.com/razorpay/razorpay-go?
✅ Did you create handlers/razorpay.go?
✅ Did you add routes in main.go?
✅ Did you create .env with credentials?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}
