package integrations

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// backendCode returns every generated source file joined together.
func backendCode(out IntegrateCheckoutOutput) string {
	var b strings.Builder
	for _, f := range out.Files {
		if f.Action == "create" {
			b.WriteString(f.Code)
			b.WriteString("\n")
		}
	}
	return b.String()
}

// TestOrderAmountComesFromTheServer guards ISS-3655598 for every framework.
//
// The generated order endpoint is unauthenticated. If it takes the price from
// the request body, a shopper pays 1 paise for any cart and still receives a
// genuine signature, because the signature covers the order id and the payment
// id and never the price. So every template must read the amount from a
// server-side record, and every verify handler must check what was actually
// paid before the caller marks anything paid.
func TestOrderAmountComesFromTheServer(t *testing.T) {
	fe := getReactFrontend()

	integrations := map[string]IntegrateCheckoutOutput{
		"django":     getDjangoIntegration(fe),
		"flask":      getFlaskIntegration(fe),
		"fastapi":    getFastAPIIntegration(fe),
		"express":    getExpressVanillaIntegration("javascript", fe),
		"nextjs":     getNextjsReactIntegration("typescript"),
		"fastify":    getFastifyIntegration("javascript", fe),
		"koa":        getKoaIntegration("javascript", fe),
		"nuxt":       getNuxtIntegration("typescript"),
		"gin":        getGinIntegration(fe),
		"echo":       getEchoIntegration(fe),
		"fiber":      getFiberIntegration(fe),
		"spring":     getSpringIntegration(fe),
		"springboot": getSpringBootIntegration(fe),
		"laravel":    getLaravelIntegration(fe),
		"rails":      getRailsIntegration(fe),
		"actix":      getActixIntegration(fe),
		"aspnet":     getAspNetIntegration(fe),
	}

	// Any read of an amount off a request-ish object. These are the shapes
	// the templates used before the fix. Reading payment["amount"] or
	// order["amount"] back from Razorpay is fine and is not matched here.
	bodyAmountPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(data|request|req|params|body|input)\b` +
			`\s*(\.get_json\(\))?\.get\(\s*['"]amount['"]`),
		regexp.MustCompile(`(?i)\b(req|request|ctx|event|body)\w*` +
			`\.(body\.|request\.)?amount\b`),
		regexp.MustCompile(`params\[:amount\]`),
		regexp.MustCompile(`(?i)input\(\s*['"]amount['"]`),
		regexp.MustCompile(`(?i)getProperty\(\s*"amount"\s*\)` +
			`\s*\.\s*Get\w+\(\)\s*\*`),
		regexp.MustCompile(`amount\s*:\s*float`),
		regexp.MustCompile(`(?i)amount\s*\*\s*100`),
	}

	// Each template names its price lookup in its own casing.
	lookupNames := []string{
		"get_order_amount", "getOrderAmount", "GetOrderAmount",
	}

	// Each template fails closed until the agent implements the lookup.
	failClosedMarkers := []string{
		"NotImplementedError", "UnsupportedOperationException",
		"InvalidOperationException", "throw new Error(",
		"errors.New(", "raise ", "Err(", "throw new \\Exception(",
	}

	for name, out := range integrations {
		t.Run(name, func(t *testing.T) {
			code := backendCode(out)

			for _, bad := range bodyAmountPatterns {
				assert.NotRegexp(t, bad, code,
					"the price must not come from the request body")
			}

			assert.True(t, containsAny(code, lookupNames),
				"needs a server-side price lookup")
			assert.True(t, containsAny(code, failClosedMarkers),
				"the price lookup must fail closed until implemented")

			// Django binds through the session
			// instead of the receipt, so it
			// compares against its own pending record rather than the order.
			if name == "django" {
				assert.Contains(t, code, "razorpay_pending",
					"django records the expected amount in the session")
			} else {
				assert.True(t,
					containsAny(code, []string{"receipt: reference",
						"'receipt': reference", "receipt = reference",
						"receipt\", reference", "\"receipt\": reference",
						"receipt: reference,", "Reference", "reference"}),
					"the receipt must carry the server-side reference")
			}

			assert.Contains(t, code, "Payment does not match the order",
				"verify must reject a payment that does not match")

			instructions := strings.ToLower(out.AIInstructions)
			assert.Contains(t, instructions, "request body",
				"instructions must warn against the request body")
		})
	}
}

func containsAny(haystack string, needles []string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}
	return false
}
