//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: ADD DEPENDENCY**
- Add razorpay-java dependency to pom.xml (com.razorpay:razorpay-java:1.4.3)
- Run: mvn clean install

**STEP 2: CREATE CONTROLLER**
- Create RazorpayController.java with the provided code
- Includes order creation and payment verification endpoints

**STEP 3: CONFIGURE APPLICATION**
- Add razorpay.key.id and razorpay.key.secret to application.properties
- These read from environment variables: ${RAZORPAY_KEY_ID} and ${RAZORPAY_KEY_SECRET}

**STEP 4: SET UP ENVIRONMENT**
- Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET environment variables

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification

FINAL CHECKLIST:
✅ Did you add razorpay-java to pom.xml?
✅ Did you create RazorpayController.java?
✅ Did you add config to application.properties?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// SPRING BOOT INTEGRATION
// =============================================================================

func getSpringBootIntegration(creds Credentials, frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := getKeysOrPlaceholders(creds)

	controllerCode := `package com.example.payment;

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
@CrossOrigin(origins = "*")
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
            response.put("error", "Failed to create order: " + e.getMessage());
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
            response.put("error", "Verification failed: " + e.getMessage());
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

	configCode := `# Razorpay Configuration
razorpay.key.id=${RAZORPAY_KEY_ID}
razorpay.key.secret=${RAZORPAY_KEY_SECRET}
`

	pomDependency := `<!-- Add to pom.xml dependencies section -->
<dependency>
    <groupId>com.razorpay</groupId>
    <artifactId>razorpay-java</artifactId>
    <version>1.4.3</version>
</dependency>
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Spring Boot + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "src/main/java/com/example/payment/RazorpayController.java", Code: controllerCode, Description: "Spring Boot REST controller for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{Action: "manual_edit", Path: "pom.xml", Description: "Add Razorpay dependency", Edits: []EditItem{
				{Line: "In <dependencies> section", Add: pomDependency, Why: "Razorpay Java SDK"},
			}},
			{Action: "manual_edit", Path: "src/main/resources/application.properties", Description: "Add Razorpay config", Edits: []EditItem{
				{Line: "Add at end of file", Add: configCode, Why: "Razorpay credentials from environment"},
			}},
			getWirePaymentAction(),
		},
		Dependencies:     []Dependency{{Name: "razorpay-java", InstallCommand: "Add to pom.xml: com.razorpay:razorpay-java:1.4.3"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: ADD DEPENDENCY**
- Add razorpay-java dependency to pom.xml (com.razorpay:razorpay-java:1.4.3)
- Run: mvn clean install

**STEP 2: CREATE CONTROLLER**
- Create RazorpayController.java with the provided code in your controller package
- Includes order creation and payment verification endpoints
- Uses @RestController, @RequestMapping, @PostMapping

**STEP 3: CONFIGURE APPLICATION**
- Add razorpay.key.id and razorpay.key.secret to application.properties
- These read from environment variables: ${RAZORPAY_KEY_ID} and ${RAZORPAY_KEY_SECRET}

**STEP 4: SET UP ENVIRONMENT**
- Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET environment variables

**STEP 5: RUN APPLICATION**
- Run: mvn spring-boot:run
- Ensure your main class has @SpringBootApplication

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification

FINAL CHECKLIST:
✅ Did you add razorpay-java to pom.xml?
✅ Did you create RazorpayController.java?
✅ Did you add config to application.properties?
✅ Did you set up environment variables?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
