//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

// =============================================================================
// PHP (Laravel) INTEGRATION
// =============================================================================

func getLaravelIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: composer require razorpay/razorpay

**STEP 2: CREATE CONTROLLER**
- Create app/Http/Controllers/RazorpayController.php with the provided code
- Includes order creation and payment verification endpoints

**STEP 3: ADD ROUTES**
- Add Razorpay API routes to routes/api.php

**STEP 4: SET UP ENVIRONMENT**
- Add RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET to .env file

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in controllers
❌ DO NOT skip payment verification

FINAL CHECKLIST:
✅ Did you run composer require razorpay/razorpay?
✅ Did you create RazorpayController.php?
✅ Did you add routes to routes/api.php?
✅ Did you add keys to .env?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// RUBY (Rails) INTEGRATION
// =============================================================================

func getRailsIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: INSTALL DEPENDENCY**
- Run: bundle add razorpay

**STEP 2: CREATE CONTROLLER**
- Create app/controllers/razorpay_controller.rb with the provided code
- Includes order creation and payment verification endpoints

**STEP 3: ADD ROUTES**
- Add Razorpay routes to config/routes.rb

**STEP 4: SET UP ENVIRONMENT**
- Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET environment variables
- Use dotenv-rails or credentials.yml for local development

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in controllers
❌ DO NOT skip payment verification

FINAL CHECKLIST:
✅ Did you run bundle add razorpay?
✅ Did you create razorpay_controller.rb?
✅ Did you add routes to config/routes.rb?
✅ Did you set up environment variables?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// RUST (Actix-web) INTEGRATION
// =============================================================================

func getActixIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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

    let key_id = env::var("RAZORPAY_KEY_ID").expect("RAZORPAY_KEY_ID must be set");
    let key_secret = env::var("RAZORPAY_KEY_SECRET").expect("RAZORPAY_KEY_SECRET must be set");

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
            let order: serde_json::Value = res.json().await.map_err(|_| actix_web::error::ErrorInternalServerError("Failed to parse response"))?;
            Ok(HttpResponse::Ok().json(OrderResponse {
                success: true,
                order_id: order["id"].as_str().map(|s| s.to_string()),
                amount: order["amount"].as_i64(),
                currency: order["currency"].as_str().map(|s| s.to_string()),
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
    let key_secret = env::var("RAZORPAY_KEY_SECRET").expect("RAZORPAY_KEY_SECRET must be set");

    let data = format!("{}|{}", req.razorpay_order_id, req.razorpay_payment_id);
    let mut mac = HmacSha256::new_from_slice(key_secret.as_bytes()).expect("HMAC key error");
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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: ADD DEPENDENCIES**
- Add reqwest, hmac, sha2, hex, and chrono to Cargo.toml

**STEP 2: CREATE HANDLERS**
- Create src/handlers/razorpay.rs with the provided code
- Includes order creation and payment verification endpoints

**STEP 3: REGISTER ROUTES**
- In main.rs, add the Razorpay routes:
  .route("/api/razorpay/order", web::post().to(handlers::razorpay::create_order))
  .route("/api/razorpay/verify", web::post().to(handlers::razorpay::verify_payment))

**STEP 4: SET UP ENVIRONMENT**
- Set RAZORPAY_KEY_ID and RAZORPAY_KEY_SECRET environment variables
- The app will panic with a clear message if these are not set

⛔ FORBIDDEN:
❌ DO NOT hardcode API keys in source code
❌ DO NOT skip payment verification

FINAL CHECKLIST:
✅ Did you add dependencies to Cargo.toml?
✅ Did you create src/handlers/razorpay.rs?
✅ Did you register routes in main.rs?
✅ Did you set up environment variables?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// .NET (ASP.NET Core) INTEGRATION
// =============================================================================

func getAspNetIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: CREATE CONTROLLER**
- Create Controllers/RazorpayController.cs with the provided code
- The controller uses IConfiguration to read "Razorpay:KeyId" and "Razorpay:KeySecret"

**STEP 2: CONFIGURE SERVICES**
- In Program.cs, add: builder.Services.AddHttpClient(); before builder.Build()
- This MUST be at the top level of Program.cs, NOT inside any method or endpoint handler

**STEP 3: ADD CONFIGURATION**
- In appsettings.json, add the Razorpay section with your credentials
- NOTE: ASP.NET uses "Razorpay:KeyId" in config (colon-separated)
- For environment variables, use double underscore: Razorpay__KeyId, Razorpay__KeySecret

**STEP 4: SET UP FRONTEND**
- Create the frontend payment helper file
- Add the script tag to the checkout page

⛔ FORBIDDEN:
❌ DO NOT mix up config naming - use "Razorpay:KeyId" in appsettings.json and Razorpay__KeyId for env vars
❌ DO NOT hardcode credentials in the controller

FINAL CHECKLIST:
✅ Did you create RazorpayController.cs?
✅ Did you add HttpClient to Program.cs?
✅ Did you add Razorpay config to appsettings.json?
✅ Did you wire up the frontend payment flow?` + getFrontendWiringInstructions(frontend),
	}
}

// =============================================================================
// MOBILE INTEGRATIONS (React Native, Flutter)
// =============================================================================
