//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

// =============================================================================
// MOBILE INTEGRATIONS (React Native, Flutter)
// =============================================================================

func getReactNativeIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

MOBILE SETUP:
1) npm install react-native-razorpay
2) iOS: cd ios && pod install
3) Android: No additional setup needed
4) Create the payment service file
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in mobile app - only use on backend
7) Import and use initiatePayment() in your checkout screen - FIND the checkout screen and ADD the import

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you CREATE the payment service file?
✅ Did you IMPORT it in the checkout screen?
✅ Did you WIRE UP the payment button to call initiatePayment()?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

func getFlutterIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

MOBILE SETUP:
1) flutter pub add razorpay_flutter http
2) Android: Add proguard rules if using minification
3) iOS: No additional setup needed
4) Create lib/services/razorpay_service.dart
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in mobile app - only use on backend
7) FIND the checkout screen and ADD RazorpayService import and usage

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you CREATE lib/services/razorpay_service.dart?
✅ Did you IMPORT it in the checkout screen?
✅ Did you WIRE UP the payment button?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

// =============================================================================
// SPRING BOOT INTEGRATION
// =============================================================================

// =============================================================================
// ANDROID (KOTLIN) INTEGRATION
// =============================================================================

func getAndroidIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	paymentServiceCode := `package com.example.app.payment

import android.app.Activity
import com.razorpay.Checkout
import com.razorpay.PaymentResultListener
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONObject
import java.net.HttpURLConnection
import java.net.URL

class RazorpayService(private val activity: Activity) : PaymentResultListener {

    companion object {
        private const val BACKEND_URL = "YOUR_BACKEND_URL" // Replace with your backend
    }

    private var onPaymentSuccess: ((String, String) -> Unit)? = null
    private var onPaymentError: ((String) -> Unit)? = null
    private var lastOrderId: String = ""

    init {
        Checkout.preload(activity.applicationContext)
    }

    suspend fun initiatePayment(
        amount: Double,
        onSuccess: (paymentId: String, orderId: String) -> Unit,
        onError: (error: String) -> Unit
    ) {
        this.onPaymentSuccess = onSuccess
        this.onPaymentError = onError

        try {
            val orderData = createOrder(amount)
            if (orderData.optBoolean("success")) {
                openCheckout(orderData)
            } else {
                onError(orderData.optString("error", "Failed to create order"))
            }
        } catch (e: Exception) {
            onError(e.message ?: "Unknown error")
        }
    }

    private suspend fun createOrder(amount: Double): JSONObject = withContext(Dispatchers.IO) {
        val url = URL("$BACKEND_URL/api/razorpay/order")
        val connection = url.openConnection() as HttpURLConnection
        connection.requestMethod = "POST"
        connection.setRequestProperty("Content-Type", "application/json")
        connection.doOutput = true

        val body = JSONObject().put("amount", amount)
        connection.outputStream.write(body.toString().toByteArray())

        val response = connection.inputStream.bufferedReader().readText()
        JSONObject(response)
    }

    private fun openCheckout(orderData: JSONObject) {
        lastOrderId = orderData.getString("orderId")
        val checkout = Checkout()
        checkout.setKeyID(orderData.getString("keyId"))

        val options = JSONObject().apply {
            put("name", "Your App Name")
            put("description", "Payment")
            put("order_id", orderData.getString("orderId"))
            put("currency", orderData.getString("currency"))
            put("amount", orderData.getInt("amount"))
            put("theme", JSONObject().put("color", "#528FF0"))
        }

        checkout.open(activity, options)
    }

    override fun onPaymentSuccess(razorpayPaymentID: String?) {
        razorpayPaymentID?.let { paymentId ->
            kotlinx.coroutines.CoroutineScope(Dispatchers.IO).launch {
                try {
                    val verifyResult = verifyPayment(paymentId)
                    withContext(Dispatchers.Main) {
                        if (verifyResult.optBoolean("success")) {
                            onPaymentSuccess?.invoke(paymentId, verifyResult.optString("orderId", ""))
                        } else {
                            onPaymentError?.invoke(verifyResult.optString("error", "Verification failed"))
                        }
                    }
                } catch (e: Exception) {
                    withContext(Dispatchers.Main) {
                        onPaymentError?.invoke("Verification failed: ${e.message}")
                    }
                }
            }
        }
    }

    private suspend fun verifyPayment(paymentId: String): JSONObject = withContext(Dispatchers.IO) {
        val url = URL("$BACKEND_URL/api/razorpay/verify")
        val connection = url.openConnection() as HttpURLConnection
        connection.requestMethod = "POST"
        connection.setRequestProperty("Content-Type", "application/json")
        connection.doOutput = true

        val body = JSONObject().apply {
            put("razorpay_payment_id", paymentId)
            put("razorpay_order_id", lastOrderId)
            put("razorpay_signature", "")
        }
        connection.outputStream.write(body.toString().toByteArray())

        val response = connection.inputStream.bufferedReader().readText()
        JSONObject(response)
    }

    override fun onPaymentError(code: Int, response: String?) {
        onPaymentError?.invoke(response ?: "Payment failed with code: $code")
    }
}
`

	activityCode := `// Add to your Activity that handles payment:

class CheckoutActivity : AppCompatActivity(), PaymentResultListener {

    private lateinit var razorpayService: RazorpayService

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        razorpayService = RazorpayService(this)
    }

    private fun handlePayment(amount: Double) {
        lifecycleScope.launch {
            razorpayService.initiatePayment(
                amount = amount,
                onSuccess = { paymentId, orderId ->
                    // Payment successful
                    Toast.makeText(this@CheckoutActivity, "Payment successful!", Toast.LENGTH_SHORT).show()
                    // Navigate to success screen or update UI
                },
                onError = { error ->
                    // Payment failed
                    Toast.makeText(this@CheckoutActivity, "Payment failed: $error", Toast.LENGTH_SHORT).show()
                }
            )
        }
    }

    // Required: Implement PaymentResultListener
    override fun onPaymentSuccess(razorpayPaymentID: String?) {
        razorpayService.onPaymentSuccess(razorpayPaymentID)
    }

    override fun onPaymentError(code: Int, response: String?) {
        razorpayService.onPaymentError(code, response)
    }
}
`

	gradleDependency := `// Add to app/build.gradle dependencies
implementation 'com.razorpay:checkout:1.6.33'
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for Android (Kotlin)",
		Files: []FileAction{
			{Action: "create", Path: "app/src/main/java/com/example/app/payment/RazorpayService.kt", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.kt", Code: activityCode, Description: "Activity usage example"},
			{Action: "manual_edit", Path: "app/build.gradle", Description: "Add Razorpay dependency", Edits: []EditItem{
				{Line: "In dependencies block", Add: gradleDependency, Why: "Razorpay Android SDK"},
			}},
		},
		Dependencies:     []Dependency{{Name: "razorpay-checkout", InstallCommand: "Add to app/build.gradle: implementation 'com.razorpay:checkout:1.6.33'"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

ANDROID SETUP:
1) Add Razorpay dependency to app/build.gradle
2) Sync Gradle
3) Create RazorpayService.kt in your payment package
4) Your Activity must implement PaymentResultListener
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in Android app - only use on backend
7) Add internet permission in AndroidManifest.xml: <uses-permission android:name="android.permission.INTERNET"/>
8) FIND the checkout Activity and WIRE UP the payment service

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you ADD the dependency to build.gradle?
✅ Did you CREATE RazorpayService.kt?
✅ Did you IMPLEMENT PaymentResultListener in the Activity?
✅ Did you WIRE UP the payment button to call initiatePayment()?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

// =============================================================================
// IOS (SWIFT) INTEGRATION
// =============================================================================

func getIOSIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	paymentServiceCode := `import Foundation
import Razorpay

class RazorpayService: NSObject {

    static let shared = RazorpayService()

    private let backendURL = "YOUR_BACKEND_URL" // Replace with your backend
    private var razorpay: RazorpayCheckout?
    private var lastOrderId: String = ""

    private var onSuccess: ((String, String) -> Void)?
    private var onError: ((String) -> Void)?

    private override init() {
        super.init()
    }

    func initiatePayment(
        amount: Double,
        viewController: UIViewController,
        onSuccess: @escaping (String, String) -> Void,
        onError: @escaping (String) -> Void
    ) {
        self.onSuccess = onSuccess
        self.onError = onError

        createOrder(amount: amount) { [weak self] result in
            switch result {
            case .success(let orderData):
                self?.openCheckout(orderData: orderData, viewController: viewController)
            case .failure(let error):
                onError(error.localizedDescription)
            }
        }
    }

    private func createOrder(amount: Double, completion: @escaping (Result<[String: Any], Error>) -> Void) {
        guard let url = URL(string: "\(backendURL)/api/razorpay/order") else {
            completion(.failure(NSError(domain: "", code: -1, userInfo: [NSLocalizedDescriptionKey: "Invalid URL"])))
            return
        }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONSerialization.data(withJSONObject: ["amount": amount])

        URLSession.shared.dataTask(with: request) { data, _, error in
            if let error = error {
                DispatchQueue.main.async { completion(.failure(error)) }
                return
            }

            guard let data = data,
                  let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else {
                DispatchQueue.main.async {
                    completion(.failure(NSError(domain: "", code: -1, userInfo: [NSLocalizedDescriptionKey: "Invalid response"])))
                }
                return
            }

            DispatchQueue.main.async { completion(.success(json)) }
        }.resume()
    }

    private func openCheckout(orderData: [String: Any], viewController: UIViewController) {
        lastOrderId = orderData["orderId"] as? String ?? ""
        guard let keyId = orderData["keyId"] as? String else {
            onError?("Missing key ID")
            return
        }

        razorpay = RazorpayCheckout.initWithKey(keyId, andDelegate: self)

        let options: [String: Any] = [
            "name": "Your App Name",
            "description": "Payment",
            "order_id": orderData["orderId"] as? String ?? "",
            "currency": orderData["currency"] as? String ?? "INR",
            "amount": orderData["amount"] as? Int ?? 0,
            "theme": ["color": "#528FF0"]
        ]

        razorpay?.open(options, displayController: viewController)
    }
}

extension RazorpayService: RazorpayPaymentCompletionProtocol {
    func onPaymentSuccess(_ payment_id: String) {
        verifyPayment(paymentId: payment_id) { [weak self] success, orderId, error in
            DispatchQueue.main.async {
                if success {
                    self?.onSuccess?(payment_id, orderId)
                } else {
                    self?.onError?(error ?? "Verification failed")
                }
            }
        }
    }

    func onPaymentError(_ code: Int32, description str: String) {
        onError?(str)
    }

    private func verifyPayment(paymentId: String, completion: @escaping (Bool, String, String?) -> Void) {
        guard let url = URL(string: "\(backendURL)/api/razorpay/verify") else {
            completion(false, "", "Invalid URL")
            return
        }

        var request = URLRequest(url: url)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONSerialization.data(withJSONObject: [
            "razorpay_payment_id": paymentId,
            "razorpay_order_id": lastOrderId,
            "razorpay_signature": ""
        ])

        URLSession.shared.dataTask(with: request) { data, _, error in
            if let error = error {
                completion(false, "", error.localizedDescription)
                return
            }
            guard let data = data,
                  let json = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else {
                completion(false, "", "Invalid response")
                return
            }
            let success = json["success"] as? Bool ?? false
            let orderId = json["orderId"] as? String ?? ""
            let errorMsg = json["error"] as? String
            completion(success, orderId, errorMsg)
        }.resume()
    }
}
`

	usageCode := `// Usage in your ViewController:

class CheckoutViewController: UIViewController {

    @IBAction func payButtonTapped(_ sender: UIButton) {
        let amount = 100.0 // Amount in your base currency

        RazorpayService.shared.initiatePayment(
            amount: amount,
            viewController: self,
            onSuccess: { paymentId, orderId in
                print("Payment successful: \(paymentId)")
                // Navigate to success screen
                DispatchQueue.main.async {
                    self.showSuccessAlert()
                }
            },
            onError: { error in
                print("Payment failed: \(error)")
                DispatchQueue.main.async {
                    self.showErrorAlert(message: error)
                }
            }
        )
    }

    private func showSuccessAlert() {
        let alert = UIAlertController(title: "Success", message: "Payment completed!", preferredStyle: .alert)
        alert.addAction(UIAlertAction(title: "OK", style: .default))
        present(alert, animated: true)
    }

    private func showErrorAlert(message: String) {
        let alert = UIAlertController(title: "Error", message: message, preferredStyle: .alert)
        alert.addAction(UIAlertAction(title: "OK", style: .default))
        present(alert, animated: true)
    }
}
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for iOS (Swift)",
		Files: []FileAction{
			{Action: "create", Path: "Services/RazorpayService.swift", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.swift", Code: usageCode, Description: "ViewController usage example"},
		},
		Dependencies:     []Dependency{{Name: "razorpay-pod", InstallCommand: "Add to Podfile: pod 'razorpay-pod' && pod install"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

IOS SETUP:
1) Add to Podfile: pod 'razorpay-pod'
2) Run: cd ios && pod install
3) Open .xcworkspace (not .xcodeproj)
4) Create Services/RazorpayService.swift
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in iOS app - only use on backend
7) Import Razorpay in your Swift files
8) FIND the checkout ViewController and WIRE UP the payment service

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you ADD razorpay-pod to Podfile?
✅ Did you CREATE RazorpayService.swift?
✅ Did you WIRE UP the payment button in the checkout screen?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

// =============================================================================
// CORDOVA INTEGRATION
// =============================================================================

func getCordovaIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	paymentServiceCode := `// www/js/razorpay-service.js

var RazorpayService = {
    backendUrl: 'YOUR_BACKEND_URL', // Replace with your backend

    initiatePayment: function(amount, onSuccess, onError) {
        var self = this;

        // Create order on backend
        this.createOrder(amount)
            .then(function(orderData) {
                if (!orderData.success) {
                    throw new Error(orderData.error || 'Failed to create order');
                }
                return self.openCheckout(orderData);
            })
            .then(function(paymentData) {
                return self.verifyPayment(paymentData);
            })
            .then(function(verifyData) {
                if (verifyData.success) {
                    onSuccess(verifyData);
                } else {
                    throw new Error(verifyData.error || 'Verification failed');
                }
            })
            .catch(function(error) {
                onError(error);
            });
    },

    createOrder: function(amount) {
        return new Promise(function(resolve, reject) {
            var xhr = new XMLHttpRequest();
            xhr.open('POST', RazorpayService.backendUrl + '/api/razorpay/order');
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onload = function() {
                if (xhr.status === 200) {
                    resolve(JSON.parse(xhr.responseText));
                } else {
                    reject(new Error('Failed to create order'));
                }
            };
            xhr.onerror = function() { reject(new Error('Network error')); };
            xhr.send(JSON.stringify({ amount: amount }));
        });
    },

    openCheckout: function(orderData) {
        return new Promise(function(resolve, reject) {
            var options = {
                description: 'Payment',
                currency: orderData.currency,
                key: orderData.keyId,
                amount: orderData.amount,
                name: 'Your App Name',
                order_id: orderData.orderId,
                theme: { color: '#528FF0' }
            };

            var successCallback = function(payment_id) {
                resolve({
                    razorpay_payment_id: payment_id,
                    razorpay_order_id: orderData.orderId
                });
            };

            var errorCallback = function(code, description) {
                reject(new Error(description || 'Payment failed'));
            };

            RazorpayCheckout.open(options, successCallback, errorCallback);
        });
    },

    verifyPayment: function(paymentData) {
        return new Promise(function(resolve, reject) {
            var xhr = new XMLHttpRequest();
            xhr.open('POST', RazorpayService.backendUrl + '/api/razorpay/verify');
            xhr.setRequestHeader('Content-Type', 'application/json');
            xhr.onload = function() {
                if (xhr.status === 200) {
                    resolve(JSON.parse(xhr.responseText));
                } else {
                    reject(new Error('Verification failed'));
                }
            };
            xhr.onerror = function() { reject(new Error('Network error')); };
            xhr.send(JSON.stringify(paymentData));
        });
    }
};
`

	usageCode := `// Usage in your Cordova app:

document.getElementById('payButton').addEventListener('click', function() {
    var amount = 100; // Amount in your base currency

    RazorpayService.initiatePayment(
        amount,
        function(data) {
            console.log('Payment successful:', data);
            alert('Payment successful!');
            // Navigate to success page
        },
        function(error) {
            console.error('Payment failed:', error);
            alert('Payment failed: ' + error.message);
        }
    );
});
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for Cordova",
		Files: []FileAction{
			{Action: "create", Path: "www/js/razorpay-service.js", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.js", Code: usageCode, Description: "Usage example"},
		},
		Dependencies:     []Dependency{{Name: "razorpay-cordova", InstallCommand: "cordova plugin add com.nicholaswilliams.nicepay.razorpay"}},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CORDOVA SETUP:
1) Install plugin: cordova plugin add com.nicholaswilliams.nicepay.razorpay
2) Create www/js/razorpay-service.js
3) Include script in index.html: <script src="js/razorpay-service.js"></script>
4) Replace YOUR_BACKEND_URL with your actual backend
5) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in the app - only use on backend
6) FIND the checkout page and WIRE UP the payment button
7) Build: cordova build android/ios

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you INSTALL the Cordova plugin?
✅ Did you CREATE razorpay-service.js?
✅ Did you ADD the script tag to index.html?
✅ Did you WIRE UP the payment button to call initiatePayment()?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

// =============================================================================
// IONIC INTEGRATION
// =============================================================================

func getIonicIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	paymentServiceCode := `// src/app/services/razorpay.service.ts

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Platform } from '@ionic/angular';

declare var RazorpayCheckout: any;

@Injectable({
  providedIn: 'root'
})
export class RazorpayService {
  private backendUrl = 'YOUR_BACKEND_URL'; // Replace with your backend

  constructor(
    private http: HttpClient,
    private platform: Platform
  ) {}

  async initiatePayment(
    amount: number,
    onSuccess: (data: any) => void,
    onError: (error: Error) => void
  ): Promise<void> {
    try {
      const orderData = await this.createOrder(amount);
      if (!orderData.success) {
        throw new Error(orderData.error || 'Failed to create order');
      }

      const paymentData = await this.openCheckout(orderData);
      const verifyData = await this.verifyPayment(paymentData);

      if (verifyData.success) {
        onSuccess(verifyData);
      } else {
        throw new Error(verifyData.error || 'Verification failed');
      }
    } catch (error: any) {
      onError(error);
    }
  }

  private createOrder(amount: number): Promise<any> {
    return this.http.post(` + "`${this.backendUrl}/api/razorpay/order`" + `, { amount }).toPromise();
  }

  private openCheckout(orderData: any): Promise<any> {
    return new Promise((resolve, reject) => {
      const options = {
        description: 'Payment',
        currency: orderData.currency,
        key: orderData.keyId,
        amount: orderData.amount,
        name: 'Your App Name',
        order_id: orderData.orderId,
        theme: { color: '#528FF0' }
      };

      const successCallback = (data: any) => {
        resolve({
          razorpay_payment_id: data.razorpay_payment_id || data,
          razorpay_order_id: data.razorpay_order_id || orderData.orderId,
          razorpay_signature: data.razorpay_signature || ''
        });
      };

      const errorCallback = (code: number, description: string) => {
        reject(new Error(description || 'Payment failed'));
      };

      RazorpayCheckout.open(options, successCallback, errorCallback);
    });
  }

  private verifyPayment(paymentData: any): Promise<any> {
    return this.http.post(` + "`${this.backendUrl}/api/razorpay/verify`" + `, paymentData).toPromise();
  }
}
`

	componentCode := `// Usage in your Ionic component:

import { Component } from '@angular/core';
import { RazorpayService } from '../services/razorpay.service';
import { ToastController, LoadingController } from '@ionic/angular';

@Component({
  selector: 'app-checkout',
  template: ` + "`" + `
    <ion-button (click)="handlePayment()" expand="block">
      Pay Now
    </ion-button>
  ` + "`" + `
})
export class CheckoutPage {
  constructor(
    private razorpayService: RazorpayService,
    private toastController: ToastController,
    private loadingController: LoadingController
  ) {}

  async handlePayment() {
    const loading = await this.loadingController.create({
      message: 'Processing payment...'
    });
    await loading.present();

    const amount = 100; // Amount in your base currency

    this.razorpayService.initiatePayment(
      amount,
      async (data) => {
        await loading.dismiss();
        const toast = await this.toastController.create({
          message: 'Payment successful!',
          duration: 3000,
          color: 'success'
        });
        await toast.present();
        // Navigate to success page
      },
      async (error) => {
        await loading.dismiss();
        const toast = await this.toastController.create({
          message: 'Payment failed: ' + error.message,
          duration: 3000,
          color: 'danger'
        });
        await toast.present();
      }
    );
  }
}
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for Ionic",
		Files: []FileAction{
			{Action: "create", Path: "src/app/services/razorpay.service.ts", Code: paymentServiceCode, Description: "Razorpay payment service"},
			{Action: "create", Path: "USAGE_EXAMPLE.ts", Code: componentCode, Description: "Component usage example"},
		},
		Dependencies: []Dependency{
			{Name: "razorpay-cordova", InstallCommand: "ionic cordova plugin add com.nicholaswilliams.nicepay.razorpay"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

IONIC SETUP:
1) Install plugin: ionic cordova plugin add com.nicholaswilliams.nicepay.razorpay
2) Create src/app/services/razorpay.service.ts
3) Import HttpClientModule in app.module.ts
4) Replace YOUR_BACKEND_URL with your actual backend
5) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in the app - only use on backend
6) Declare RazorpayCheckout in your service or global typings
7) FIND the checkout component and INJECT and USE the RazorpayService
8) Build: ionic cordova build android/ios

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you INSTALL the Cordova plugin?
✅ Did you CREATE razorpay.service.ts?
✅ Did you IMPORT HttpClientModule?
✅ Did you INJECT and USE the service in the checkout component?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}

// =============================================================================
// CAPACITOR INTEGRATION
// =============================================================================

func getCapacitorIntegration() IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	paymentServiceCode := `// src/services/razorpay.service.ts

import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Capacitor } from '@capacitor/core';

// For Capacitor, we use a custom plugin or web fallback
declare var RazorpayCheckout: any;

@Injectable({
  providedIn: 'root'
})
export class RazorpayService {
  private backendUrl = 'YOUR_BACKEND_URL'; // Replace with your backend

  constructor(private http: HttpClient) {}

  async initiatePayment(
    amount: number,
    onSuccess: (data: any) => void,
    onError: (error: Error) => void
  ): Promise<void> {
    try {
      const orderData: any = await this.createOrder(amount);
      if (!orderData.success) {
        throw new Error(orderData.error || 'Failed to create order');
      }

      if (Capacitor.isNativePlatform()) {
        // Native platform - use Cordova plugin
        await this.openNativeCheckout(orderData, onSuccess, onError);
      } else {
        // Web platform - use Razorpay web checkout
        await this.openWebCheckout(orderData, onSuccess, onError);
      }
    } catch (error: any) {
      onError(error);
    }
  }

  private createOrder(amount: number): Promise<any> {
    return this.http.post(` + "`${this.backendUrl}/api/razorpay/order`" + `, { amount }).toPromise();
  }

  private async openNativeCheckout(
    orderData: any,
    onSuccess: (data: any) => void,
    onError: (error: Error) => void
  ): Promise<void> {
    const options = {
      description: 'Payment',
      currency: orderData.currency,
      key: orderData.keyId,
      amount: orderData.amount,
      name: 'Your App Name',
      order_id: orderData.orderId,
      theme: { color: '#528FF0' }
    };

    RazorpayCheckout.open(
      options,
      async (payment_id: string) => {
        const verifyData = await this.verifyPayment({
          razorpay_payment_id: payment_id,
          razorpay_order_id: orderData.orderId
        });
        if (verifyData.success) {
          onSuccess(verifyData);
        } else {
          onError(new Error(verifyData.error || 'Verification failed'));
        }
      },
      (code: number, description: string) => {
        onError(new Error(description || 'Payment failed'));
      }
    );
  }

  private async openWebCheckout(
    orderData: any,
    onSuccess: (data: any) => void,
    onError: (error: Error) => void
  ): Promise<void> {
    // Load Razorpay script if not loaded
    if (!(window as any).Razorpay) {
      await this.loadRazorpayScript();
    }

    const options = {
      key: orderData.keyId,
      amount: orderData.amount,
      currency: orderData.currency,
      name: 'Your App Name',
      order_id: orderData.orderId,
      handler: async (response: any) => {
        const verifyData: any = await this.verifyPayment(response);
        if (verifyData.success) {
          onSuccess(verifyData);
        } else {
          onError(new Error(verifyData.error || 'Verification failed'));
        }
      },
      modal: {
        ondismiss: () => {
          onError(new Error('Payment cancelled'));
        }
      },
      theme: { color: '#528FF0' }
    };

    const razorpay = new (window as any).Razorpay(options);
    razorpay.open();
  }

  private loadRazorpayScript(): Promise<void> {
    return new Promise((resolve, reject) => {
      const script = document.createElement('script');
      script.src = 'https://checkout.razorpay.com/v1/checkout.js';
      script.onload = () => resolve();
      script.onerror = () => reject(new Error('Failed to load Razorpay'));
      document.head.appendChild(script);
    });
  }

  private verifyPayment(paymentData: any): Promise<any> {
    return this.http.post(` + "`${this.backendUrl}/api/razorpay/verify`" + `, paymentData).toPromise();
  }
}
`

	usageCode := `// Usage in your Capacitor/Angular component:

import { Component } from '@angular/core';
import { RazorpayService } from '../services/razorpay.service';

@Component({
  selector: 'app-checkout',
  template: ` + "`" + `
    <button (click)="handlePayment()">Pay Now</button>
  ` + "`" + `
})
export class CheckoutComponent {
  constructor(private razorpayService: RazorpayService) {}

  handlePayment() {
    const amount = 100; // Amount in your base currency

    this.razorpayService.initiatePayment(
      amount,
      (data) => {
        console.log('Payment successful:', data);
        alert('Payment successful!');
        // Navigate to success page
      },
      (error) => {
        console.error('Payment failed:', error);
        alert('Payment failed: ' + error.message);
      }
    );
  }
}
`

	return IntegrateCheckoutOutput{
		Summary: "Razorpay integration for Capacitor",
		Files: []FileAction{
			{Action: "create", Path: "src/services/razorpay.service.ts", Code: paymentServiceCode, Description: "Razorpay payment service with web/native support"},
			{Action: "create", Path: "USAGE_EXAMPLE.ts", Code: usageCode, Description: "Component usage example"},
		},
		Dependencies: []Dependency{
			{Name: "razorpay-cordova", InstallCommand: "npm install cordova-plugin-razorpay && npx cap sync"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CAPACITOR SETUP:
1) Install Cordova plugin: npm install cordova-plugin-razorpay
2) Sync: npx cap sync
3) Create src/services/razorpay.service.ts
4) Import HttpClientModule in app.module.ts
5) Replace YOUR_BACKEND_URL with your actual backend
6) IMPORTANT: Never expose RAZORPAY_KEY_SECRET in the app - only use on backend
7) Works on both web and native (Android/iOS)
8) For web: Razorpay script is loaded dynamically
9) For native: Uses Cordova plugin
10) FIND the checkout component and INJECT and USE the RazorpayService

═══════════════════════════════════════════════════════════════════════════════
FINAL REMINDER:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you INSTALL cordova-plugin-razorpay and run npx cap sync?
✅ Did you CREATE razorpay.service.ts?
✅ Did you IMPORT HttpClientModule?
✅ Did you INJECT and USE the service in the checkout component?
❌ DO NOT give "Next Steps" - YOU must complete everything`,
	}
}
