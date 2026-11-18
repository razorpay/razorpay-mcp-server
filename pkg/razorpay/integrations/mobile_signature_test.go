package integrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMobileTemplatesForwardSDKSignature checks that every generated mobile
// client sends the razorpay_signature returned by the SDK success callback to
// the verify endpoint, because every generated backend requires it.
func TestMobileTemplatesForwardSDKSignature(t *testing.T) {
	tests := []struct {
		name string
		out  IntegrateCheckoutOutput
		// Each fragment must appear in the generated code.
		mustContain []string
		// None of these fragments may appear in the generated code.
		mustNotContain []string
	}{
		{
			name: "react-native",
			out:  getReactNativeIntegration(),
			mustContain: []string{
				"verifyPayment(paymentResponse)",
			},
		},
		{
			name: "flutter",
			out:  getFlutterIntegration(),
			mustContain: []string{
				"'razorpay_signature': response.signature",
			},
		},
		{
			name: "android",
			out:  getAndroidIntegration(),
			mustContain: []string{
				"import com.razorpay.PaymentData",
				"import com.razorpay.PaymentResultWithDataListener",
				"RazorpayService(private val activity: Activity) : " +
					"PaymentResultWithDataListener",
				"paymentData: PaymentData?",
				"paymentData?.signature",
				`put("razorpay_signature", signature)`,
				"AppCompatActivity(), PaymentResultWithDataListener",
				"razorpayService.onPaymentSuccess(razorpayPaymentID, paymentData)",
			},
			mustNotContain: []string{
				`put("razorpay_signature", "")`,
				"import com.razorpay.PaymentResultListener\n",
				"), PaymentResultListener {",
				": PaymentResultListener {",
			},
		},
		{
			name: "ios",
			out:  getIOSIntegration(),
			mustContain: []string{
				"RazorpayPaymentCompletionProtocolWithData",
				"andDelegateWithData: self",
				"andData response: [AnyHashable: Any]?",
				`response?["razorpay_signature"] as? String`,
				`"razorpay_signature": signature`,
			},
			mustNotContain: []string{
				`"razorpay_signature": ""`,
				"RazorpayPaymentCompletionProtocol {",
			},
		},
		{
			name: "cordova",
			out:  getCordovaIntegration(),
			mustContain: []string{
				"razorpay_signature: response.razorpay_signature",
			},
			mustNotContain: []string{
				"function(payment_id)",
			},
		},
		{
			name: "ionic",
			out:  getIonicIntegration(),
			mustContain: []string{
				"razorpay_signature: response.razorpay_signature",
			},
			mustNotContain: []string{
				"razorpay_signature: data.razorpay_signature || ''",
				"data.razorpay_payment_id || data,",
			},
		},
		{
			name: "capacitor",
			out:  getCapacitorIntegration(),
			mustContain: []string{
				"razorpay_signature: response.razorpay_signature",
			},
			mustNotContain: []string{
				"async (payment_id: string)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := backendCode(tt.out)
			for _, fragment := range tt.mustContain {
				assert.Contains(t, code, fragment)
			}
			for _, fragment := range tt.mustNotContain {
				assert.NotContains(t, code, fragment)
			}
		})
	}
}

// TestMobileTemplatesFailClosedWithoutSignature checks that the generated
// native clients stop before calling verify when the SDK gives no signature.
func TestMobileTemplatesFailClosedWithoutSignature(t *testing.T) {
	tests := map[string]IntegrateCheckoutOutput{
		"android":   getAndroidIntegration(),
		"ios":       getIOSIntegration(),
		"cordova":   getCordovaIntegration(),
		"ionic":     getIonicIntegration(),
		"capacitor": getCapacitorIntegration(),
	}
	for name, out := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Contains(t, backendCode(out), "Payment signature missing")
		})
	}
}

// TestMobileInstructionsRequireServerSideVerification checks that the agent
// instructions for every mobile platform require the backend to verify the
// signature and the paid amount before marking an order paid.
func TestMobileInstructionsRequireServerSideVerification(t *testing.T) {
	tests := map[string]IntegrateCheckoutOutput{
		"react-native": getReactNativeIntegration(),
		"flutter":      getFlutterIntegration(),
		"android":      getAndroidIntegration(),
		"ios":          getIOSIntegration(),
		"cordova":      getCordovaIntegration(),
		"ionic":        getIonicIntegration(),
		"capacitor":    getCapacitorIntegration(),
	}
	for name, out := range tests {
		t.Run(name, func(t *testing.T) {
			instructions := strings.ToLower(out.AIInstructions)
			assert.Contains(t, instructions, "razorpay_signature")
			assert.Contains(t, instructions, "verify the signature")
			assert.Contains(t, instructions, "amount")
		})
	}
}
