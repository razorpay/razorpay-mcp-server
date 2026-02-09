// Package integrations provides code templates for Payment Gateway integration.
package integrations

// Order creation templates by language
// Note: These templates show how to call Razorpay Orders API (https://api.razorpay.com/v1/orders)
// from your backend server. The SDK handles the API call internally.
var OrderCreateTemplates = map[string]string{
	LangNodeJS: `// Server-side: Create Razorpay Order (Node.js)
// This code runs on YOUR backend server and calls Razorpay's Orders API
//
// Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
// SDK Reference: https://razorpay.com/docs/payments/server-integration/nodejs/
// API Reference: https://razorpay.com/docs/api/orders/create/

const Razorpay = require('razorpay');

// Initialize Razorpay instance with your credentials
// The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
const razorpay = new Razorpay({
  key_id: process.env.RAZORPAY_KEY_ID,
  key_secret: process.env.RAZORPAY_KEY_SECRET,
});

// YOUR backend endpoint that clients will call
// This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
app.post('/api/create-order', async (req, res) => {
  try {
    const { amount, currency = 'INR', receipt, notes } = req.body;

    // Options for Razorpay Orders API
    const options = {
      amount: amount, // Amount in smallest currency unit (paise for INR)
      currency: currency,
      receipt: receipt || ` + "`receipt_${Date.now()}`" + `,
      notes: notes || {},
    };

    // SDK calls: POST https://api.razorpay.com/v1/orders
    const order = await razorpay.orders.create(options);
    
    // Return order details to client
    // IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
    res.json({
      success: true,
      order_id: order.id,
      amount: order.amount,
      currency: order.currency,
      key_id: process.env.RAZORPAY_KEY_ID,
    });
  } catch (error) {
    console.error('Error creating order:', error);
    res.status(500).json({ success: false, error: error.message });
  }
});`,

	LangPython: `# Server-side: Create Razorpay Order (Python)
# This code runs on YOUR backend server and calls Razorpay's Orders API
#
# Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
# SDK Reference: https://razorpay.com/docs/payments/server-integration/python/
# API Reference: https://razorpay.com/docs/api/orders/create/

import razorpay
import os
import time
from flask import Flask, request, jsonify

# Initialize Razorpay client with your credentials
# The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
client = razorpay.Client(
    auth=(os.environ.get('RAZORPAY_KEY_ID'), os.environ.get('RAZORPAY_KEY_SECRET'))
)

# YOUR backend endpoint that clients will call
# This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
@app.route('/api/create-order', methods=['POST'])
def create_order():
    try:
        data = request.get_json()
        amount = data.get('amount')  # Amount in smallest currency unit (paise for INR)
        currency = data.get('currency', 'INR')
        receipt = data.get('receipt', f'receipt_{int(time.time())}')
        notes = data.get('notes', {})

        # Options for Razorpay Orders API
        order_data = {
            'amount': amount,
            'currency': currency,
            'receipt': receipt,
            'notes': notes,
        }

        # SDK calls: POST https://api.razorpay.com/v1/orders
        order = client.order.create(data=order_data)
        
        # Return order details to client
        # IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
        return jsonify({
            'success': True,
            'order_id': order['id'],
            'amount': order['amount'],
            'currency': order['currency'],
            'key_id': os.environ.get('RAZORPAY_KEY_ID'),
        })
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)}), 500`,

	LangGo: `// Server-side: Create Razorpay Order (Go)
// This code runs on YOUR backend server and calls Razorpay's Orders API
//
// Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
// SDK Reference: https://razorpay.com/docs/payments/server-integration/go/
// API Reference: https://razorpay.com/docs/api/orders/create/

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	razorpay "github.com/razorpay/razorpay-go"
)

// Initialize Razorpay client with your credentials
// The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
var client = razorpay.NewClient(
	os.Getenv("RAZORPAY_KEY_ID"),
	os.Getenv("RAZORPAY_KEY_SECRET"),
)

type CreateOrderRequest struct {
	Amount   int64             ` + "`json:\"amount\"`" + `   // Amount in smallest currency unit (paise for INR)
	Currency string            ` + "`json:\"currency\"`" + ` // e.g., "INR"
	Receipt  string            ` + "`json:\"receipt\"`" + `
	Notes    map[string]string ` + "`json:\"notes\"`" + `
}

// YOUR backend endpoint that clients will call
// This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
func CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Currency == "" {
		req.Currency = "INR"
	}
	if req.Receipt == "" {
		req.Receipt = fmt.Sprintf("receipt_%d", time.Now().Unix())
	}

	// Options for Razorpay Orders API
	data := map[string]interface{}{
		"amount":   req.Amount,
		"currency": req.Currency,
		"receipt":  req.Receipt,
		"notes":    req.Notes,
	}

	// SDK calls: POST https://api.razorpay.com/v1/orders
	order, err := client.Order.Create(data, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Return order details to client
	// IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"order_id": order["id"],
		"amount":   order["amount"],
		"currency": order["currency"],
		"key_id":   os.Getenv("RAZORPAY_KEY_ID"),
	})
}`,

	LangPHP: `<?php
// Server-side: Create Razorpay Order (PHP)
// This code runs on YOUR backend server and calls Razorpay's Orders API
//
// Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
// SDK Reference: https://razorpay.com/docs/payments/server-integration/php/
// API Reference: https://razorpay.com/docs/api/orders/create/

require 'vendor/autoload.php';

use Razorpay\Api\Api;

// Initialize Razorpay API with your credentials
// The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
$api = new Api(getenv('RAZORPAY_KEY_ID'), getenv('RAZORPAY_KEY_SECRET'));

// YOUR backend endpoint that clients will call
// This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
$input = json_decode(file_get_contents('php://input'), true);

try {
    $amount = $input['amount'];  // Amount in smallest currency unit (paise for INR)
    $currency = $input['currency'] ?? 'INR';
    $receipt = $input['receipt'] ?? 'receipt_' . time();
    $notes = $input['notes'] ?? [];

    // Options for Razorpay Orders API
    $orderData = [
        'amount' => $amount,
        'currency' => $currency,
        'receipt' => $receipt,
        'notes' => $notes,
    ];

    // SDK calls: POST https://api.razorpay.com/v1/orders
    $order = $api->order->create($orderData);
    
    // Return order details to client
    // IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
    echo json_encode([
        'success' => true,
        'order_id' => $order->id,
        'amount' => $order->amount,
        'currency' => $order->currency,
        'key_id' => getenv('RAZORPAY_KEY_ID'),
    ]);
} catch (Exception $e) {
    http_response_code(500);
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage(),
    ]);
}`,

	LangRuby: `# Server-side: Create Razorpay Order (Ruby)
# This code runs on YOUR backend server and calls Razorpay's Orders API
#
# Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
# SDK Reference: https://razorpay.com/docs/payments/server-integration/ruby/
# API Reference: https://razorpay.com/docs/api/orders/create/

require 'razorpay'

# Initialize Razorpay with your credentials
# The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
Razorpay.setup(ENV['RAZORPAY_KEY_ID'], ENV['RAZORPAY_KEY_SECRET'])

class PaymentsController < ApplicationController
  # YOUR backend endpoint that clients will call
  # This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
  def create_order
    amount = params[:amount]  # Amount in smallest currency unit (paise for INR)
    currency = params[:currency] || 'INR'
    receipt = params[:receipt] || "receipt_#{Time.now.to_i}"
    notes = params[:notes] || {}

    # SDK calls: POST https://api.razorpay.com/v1/orders
    order = Razorpay::Order.create(
      amount: amount,
      currency: currency,
      receipt: receipt,
      notes: notes
    )

    # Return order details to client
    # IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
    render json: {
      success: true,
      order_id: order.id,
      amount: order.amount,
      currency: order.currency,
      key_id: ENV['RAZORPAY_KEY_ID']
    }
  rescue StandardError => e
    render json: { success: false, error: e.message }, status: :internal_server_error
  end
end`,

	LangJava: `// Server-side: Create Razorpay Order (Java)
// This code runs on YOUR backend server and calls Razorpay's Orders API
//
// Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
// SDK Reference: https://razorpay.com/docs/payments/server-integration/java/
// API Reference: https://razorpay.com/docs/api/orders/create/

import com.razorpay.Order;
import com.razorpay.RazorpayClient;
import com.razorpay.RazorpayException;
import org.json.JSONObject;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api")
public class PaymentController {

    private RazorpayClient razorpayClient;

    public PaymentController() throws RazorpayException {
        // Initialize Razorpay client with your credentials
        // The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
        this.razorpayClient = new RazorpayClient(
            System.getenv("RAZORPAY_KEY_ID"),
            System.getenv("RAZORPAY_KEY_SECRET")
        );
    }

    // YOUR backend endpoint that clients will call
    // This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
    @PostMapping("/create-order")
    public String createOrder(@RequestBody OrderRequest request) {
        try {
            // Options for Razorpay Orders API
            JSONObject orderRequest = new JSONObject();
            orderRequest.put("amount", request.getAmount());  // Amount in paise
            orderRequest.put("currency", request.getCurrency() != null ? request.getCurrency() : "INR");
            orderRequest.put("receipt", request.getReceipt() != null ? 
                request.getReceipt() : "receipt_" + System.currentTimeMillis());
            
            if (request.getNotes() != null) {
                orderRequest.put("notes", new JSONObject(request.getNotes()));
            }

            // SDK calls: POST https://api.razorpay.com/v1/orders
            Order order = razorpayClient.orders.create(orderRequest);
            
            // Return order details to client
            // IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
            JSONObject response = new JSONObject();
            response.put("success", true);
            response.put("order_id", order.get("id"));
            response.put("amount", order.get("amount"));
            response.put("currency", order.get("currency"));
            response.put("key_id", System.getenv("RAZORPAY_KEY_ID"));
            
            return response.toString();
        } catch (RazorpayException e) {
            JSONObject error = new JSONObject();
            error.put("success", false);
            error.put("error", e.getMessage());
            return error.toString();
        }
    }
}`,

	LangDotNet: `// Server-side: Create Razorpay Order (C# / .NET)
// This code runs on YOUR backend server and calls Razorpay's Orders API
//
// Razorpay API Endpoint: POST https://api.razorpay.com/v1/orders
// SDK Reference: https://razorpay.com/docs/payments/server-integration/dotnet/
// API Reference: https://razorpay.com/docs/api/orders/create/

using Microsoft.AspNetCore.Mvc;
using Razorpay.Api;
using System.Collections.Generic;

[ApiController]
[Route("api")]
public class PaymentController : ControllerBase
{
    private readonly RazorpayClient _client;

    public PaymentController()
    {
        // Initialize Razorpay client with your credentials
        // The SDK will make authenticated requests to https://api.razorpay.com/v1/orders
        _client = new RazorpayClient(
            Environment.GetEnvironmentVariable("RAZORPAY_KEY_ID"),
            Environment.GetEnvironmentVariable("RAZORPAY_KEY_SECRET")
        );
    }

    // YOUR backend endpoint that clients will call
    // This internally calls Razorpay's API: POST https://api.razorpay.com/v1/orders
    [HttpPost("create-order")]
    public IActionResult CreateOrder([FromBody] CreateOrderRequest request)
    {
        try
        {
            // Options for Razorpay Orders API
            var options = new Dictionary<string, object>
            {
                { "amount", request.Amount },  // Amount in smallest currency unit (paise for INR)
                { "currency", request.Currency ?? "INR" },
                { "receipt", request.Receipt ?? $"receipt_{DateTimeOffset.UtcNow.ToUnixTimeSeconds()}" }
            };

            if (request.Notes != null)
            {
                options.Add("notes", request.Notes);
            }

            // SDK calls: POST https://api.razorpay.com/v1/orders
            var order = _client.Order.Create(options);

            // Return order details to client
            // IMPORTANT: Send key_id (NOT key_secret) to client for checkout initialization
            return Ok(new
            {
                success = true,
                order_id = order["id"].ToString(),
                amount = order["amount"],
                currency = order["currency"].ToString(),
                key_id = Environment.GetEnvironmentVariable("RAZORPAY_KEY_ID")
            });
        }
        catch (Exception ex)
        {
            return StatusCode(500, new { success = false, error = ex.Message });
        }
    }
}

public class CreateOrderRequest
{
    public long Amount { get; set; }
    public string Currency { get; set; }
    public string Receipt { get; set; }
    public Dictionary<string, string> Notes { get; set; }
}`,
}

// Signature verification templates by language
// Note: Signature verification must be done on your server after receiving
// payment response from the client. The signature is: HMAC-SHA256(order_id|payment_id, key_secret)
var SignatureVerifyTemplates = map[string]string{
	LangNodeJS: `// Server-side: Verify Razorpay Payment Signature (Node.js)
// This MUST be done on your server - NEVER trust client-side verification!
//
// Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
// Reference: https://razorpay.com/docs/payments/server-integration/nodejs/payment-verification/

const crypto = require('crypto');

/**
 * Verifies the Razorpay payment signature
 * IMPORTANT: Use the order_id you stored when creating the order,
 * NOT the razorpay_order_id from the client response (they should match, but verify from your DB)
 */
function verifyPaymentSignature(orderId, paymentId, signature) {
  const body = orderId + '|' + paymentId;
  
  const expectedSignature = crypto
    .createHmac('sha256', process.env.RAZORPAY_KEY_SECRET)
    .update(body)
    .digest('hex');
  
  // Use constant-time comparison to prevent timing attacks
  return crypto.timingSafeEqual(
    Buffer.from(expectedSignature),
    Buffer.from(signature)
  );
}

// YOUR backend endpoint for payment verification
app.post('/api/verify-payment', (req, res) => {
  const { razorpay_order_id, razorpay_payment_id, razorpay_signature } = req.body;
  
  // 1. Retrieve the original order_id from YOUR database
  // This ensures the payment is for an order you created
  const storedOrderId = getOrderIdFromDatabase(razorpay_order_id);
  
  if (!storedOrderId) {
    return res.status(400).json({ success: false, error: 'Order not found' });
  }
  
  // 2. Verify the signature
  const isValid = verifyPaymentSignature(
    storedOrderId,
    razorpay_payment_id,
    razorpay_signature
  );
  
  if (isValid) {
    // 3. Payment is verified - update your database
    // Mark order as paid, fulfill the order, etc.
    updateOrderStatus(storedOrderId, 'paid', razorpay_payment_id);
    res.json({ success: true, message: 'Payment verified' });
  } else {
    res.status(400).json({ success: false, error: 'Invalid signature' });
  }
});`,

	LangPython: `# Server-side: Verify Razorpay Payment Signature (Python)
# This MUST be done on your server - NEVER trust client-side verification!
#
# Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
# Reference: https://razorpay.com/docs/payments/server-integration/python/payment-verification/

import hmac
import hashlib
import os

def verify_payment_signature(order_id, payment_id, signature):
    """
    Verifies the Razorpay payment signature
    IMPORTANT: Use the order_id you stored when creating the order,
    NOT the razorpay_order_id from the client response (verify from your DB)
    """
    body = f"{order_id}|{payment_id}"
    
    expected_signature = hmac.new(
        os.environ.get('RAZORPAY_KEY_SECRET').encode(),
        body.encode(),
        hashlib.sha256
    ).hexdigest()
    
    # Use constant-time comparison to prevent timing attacks
    return hmac.compare_digest(expected_signature, signature)

# YOUR backend endpoint for payment verification
@app.route('/api/verify-payment', methods=['POST'])
def verify_payment():
    data = request.get_json()
    razorpay_order_id = data.get('razorpay_order_id')
    razorpay_payment_id = data.get('razorpay_payment_id')
    razorpay_signature = data.get('razorpay_signature')
    
    # 1. Retrieve the original order_id from YOUR database
    # This ensures the payment is for an order you created
    stored_order_id = get_order_id_from_database(razorpay_order_id)
    
    if not stored_order_id:
        return jsonify({'success': False, 'error': 'Order not found'}), 400
    
    # 2. Verify the signature
    is_valid = verify_payment_signature(
        stored_order_id,
        razorpay_payment_id,
        razorpay_signature
    )
    
    if is_valid:
        # 3. Payment is verified - update your database
        # Mark order as paid, fulfill the order, etc.
        update_order_status(stored_order_id, 'paid', razorpay_payment_id)
        return jsonify({'success': True, 'message': 'Payment verified'})
    else:
        return jsonify({'success': False, 'error': 'Invalid signature'}), 400`,

	LangGo: `// Server-side: Verify Razorpay Payment Signature (Go)
// This MUST be done on your server - NEVER trust client-side verification!
//
// Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
// Reference: https://razorpay.com/docs/payments/server-integration/go/payment-verification/

package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

// VerifyPaymentSignature verifies the Razorpay payment signature
// IMPORTANT: Use the order_id you stored when creating the order,
// NOT the razorpay_order_id from the client response (verify from your DB)
func VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	body := orderID + "|" + paymentID
	
	h := hmac.New(sha256.New, []byte(os.Getenv("RAZORPAY_KEY_SECRET")))
	h.Write([]byte(body))
	expectedSignature := hex.EncodeToString(h.Sum(nil))
	
	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare([]byte(expectedSignature), []byte(signature)) == 1
}

type VerifyPaymentRequest struct {
	RazorpayOrderID   string ` + "`json:\"razorpay_order_id\"`" + `
	RazorpayPaymentID string ` + "`json:\"razorpay_payment_id\"`" + `
	RazorpaySignature string ` + "`json:\"razorpay_signature\"`" + `
}

// YOUR backend endpoint for payment verification
func VerifyPayment(w http.ResponseWriter, r *http.Request) {
	var req VerifyPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	// 1. Retrieve the original order_id from YOUR database
	// This ensures the payment is for an order you created
	storedOrderID := getOrderIDFromDatabase(req.RazorpayOrderID)
	if storedOrderID == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Order not found",
		})
		return
	}
	
	// 2. Verify the signature
	isValid := VerifyPaymentSignature(
		storedOrderID,
		req.RazorpayPaymentID,
		req.RazorpaySignature,
	)
	
	w.Header().Set("Content-Type", "application/json")
	if isValid {
		// 3. Payment is verified - update your database
		// Mark order as paid, fulfill the order, etc.
		updateOrderStatus(storedOrderID, "paid", req.RazorpayPaymentID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Payment verified",
		})
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Invalid signature",
		})
	}
}`,

	LangPHP: `<?php
// Server-side: Verify Razorpay Payment Signature (PHP)
// This MUST be done on your server - NEVER trust client-side verification!
//
// Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
// Reference: https://razorpay.com/docs/payments/server-integration/php/payment-verification/

/**
 * Verifies the Razorpay payment signature
 * IMPORTANT: Use the order_id you stored when creating the order,
 * NOT the razorpay_order_id from the client response (verify from your DB)
 */
function verifyPaymentSignature($orderId, $paymentId, $signature) {
    $body = $orderId . '|' . $paymentId;
    
    $expectedSignature = hash_hmac(
        'sha256',
        $body,
        getenv('RAZORPAY_KEY_SECRET')
    );
    
    // Use constant-time comparison to prevent timing attacks
    return hash_equals($expectedSignature, $signature);
}

// YOUR backend endpoint for payment verification
$input = json_decode(file_get_contents('php://input'), true);

$razorpayOrderId = $input['razorpay_order_id'];
$razorpayPaymentId = $input['razorpay_payment_id'];
$razorpaySignature = $input['razorpay_signature'];

// 1. Retrieve the original order_id from YOUR database
// This ensures the payment is for an order you created
$storedOrderId = getOrderIdFromDatabase($razorpayOrderId);

if (!$storedOrderId) {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Order not found']);
    exit;
}

// 2. Verify the signature
$isValid = verifyPaymentSignature(
    $storedOrderId,
    $razorpayPaymentId,
    $razorpaySignature
);

if ($isValid) {
    // 3. Payment is verified - update your database
    // Mark order as paid, fulfill the order, etc.
    updateOrderStatus($storedOrderId, 'paid', $razorpayPaymentId);
    echo json_encode(['success' => true, 'message' => 'Payment verified']);
} else {
    http_response_code(400);
    echo json_encode(['success' => false, 'error' => 'Invalid signature']);
}`,

	LangRuby: `# Server-side: Verify Razorpay Payment Signature (Ruby)
# This MUST be done on your server - NEVER trust client-side verification!
#
# Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
# Reference: https://razorpay.com/docs/payments/server-integration/ruby/payment-verification/

require 'openssl'

# Verifies the Razorpay payment signature
# IMPORTANT: Use the order_id you stored when creating the order,
# NOT the razorpay_order_id from the client response (verify from your DB)
def verify_payment_signature(order_id, payment_id, signature)
  body = "#{order_id}|#{payment_id}"
  
  expected_signature = OpenSSL::HMAC.hexdigest(
    'sha256',
    ENV['RAZORPAY_KEY_SECRET'],
    body
  )
  
  # Use constant-time comparison to prevent timing attacks
  ActiveSupport::SecurityUtils.secure_compare(expected_signature, signature)
end

# YOUR backend endpoint for payment verification
def verify_payment
  razorpay_order_id = params[:razorpay_order_id]
  razorpay_payment_id = params[:razorpay_payment_id]
  razorpay_signature = params[:razorpay_signature]
  
  # 1. Retrieve the original order_id from YOUR database
  # This ensures the payment is for an order you created
  stored_order_id = get_order_id_from_database(razorpay_order_id)
  
  unless stored_order_id
    render json: { success: false, error: 'Order not found' }, status: :bad_request
    return
  end
  
  # 2. Verify the signature
  is_valid = verify_payment_signature(
    stored_order_id,
    razorpay_payment_id,
    razorpay_signature
  )
  
  if is_valid
    # 3. Payment is verified - update your database
    # Mark order as paid, fulfill the order, etc.
    update_order_status(stored_order_id, 'paid', razorpay_payment_id)
    render json: { success: true, message: 'Payment verified' }
  else
    render json: { success: false, error: 'Invalid signature' }, status: :bad_request
  end
end`,

	LangJava: `// Server-side: Verify Razorpay Payment Signature (Java)
// This MUST be done on your server - NEVER trust client-side verification!
//
// Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
// Reference: https://razorpay.com/docs/payments/server-integration/java/payment-verification/

import javax.crypto.Mac;
import javax.crypto.spec.SecretKeySpec;
import java.security.MessageDigest;

public class PaymentVerification {

    /**
     * Verifies the Razorpay payment signature
     * IMPORTANT: Use the order_id you stored when creating the order,
     * NOT the razorpay_order_id from the client response (verify from your DB)
     */
    public static boolean verifyPaymentSignature(
            String orderId, String paymentId, String signature) {
        try {
            String body = orderId + "|" + paymentId;
            
            Mac sha256Hmac = Mac.getInstance("HmacSHA256");
            SecretKeySpec secretKey = new SecretKeySpec(
                System.getenv("RAZORPAY_KEY_SECRET").getBytes(), "HmacSHA256"
            );
            sha256Hmac.init(secretKey);
            
            byte[] hash = sha256Hmac.doFinal(body.getBytes());
            String expectedSignature = bytesToHex(hash);
            
            // Use constant-time comparison to prevent timing attacks
            return MessageDigest.isEqual(
                expectedSignature.getBytes(), signature.getBytes()
            );
        } catch (Exception e) {
            return false;
        }
    }
    
    private static String bytesToHex(byte[] bytes) {
        StringBuilder sb = new StringBuilder();
        for (byte b : bytes) {
            sb.append(String.format("%02x", b));
        }
        return sb.toString();
    }
}

// In your controller:
@PostMapping("/api/verify-payment")
public ResponseEntity<?> verifyPayment(@RequestBody VerifyPaymentRequest request) {
    // 1. Retrieve the original order_id from YOUR database
    String storedOrderId = getOrderIdFromDatabase(request.getRazorpayOrderId());
    
    if (storedOrderId == null) {
        return ResponseEntity.badRequest()
            .body(Map.of("success", false, "error", "Order not found"));
    }
    
    // 2. Verify the signature
    boolean isValid = PaymentVerification.verifyPaymentSignature(
        storedOrderId,
        request.getRazorpayPaymentId(),
        request.getRazorpaySignature()
    );
    
    if (isValid) {
        // 3. Payment is verified - update your database
        updateOrderStatus(storedOrderId, "paid", request.getRazorpayPaymentId());
        return ResponseEntity.ok(Map.of("success", true, "message", "Payment verified"));
    }
    
    return ResponseEntity.badRequest()
        .body(Map.of("success", false, "error", "Invalid signature"));
}`,

	LangDotNet: `// Server-side: Verify Razorpay Payment Signature (C# / .NET)
// This MUST be done on your server - NEVER trust client-side verification!
//
// Signature Formula: HMAC-SHA256(order_id + "|" + payment_id, key_secret)
// Reference: https://razorpay.com/docs/payments/server-integration/dotnet/payment-verification/

using System.Security.Cryptography;
using System.Text;

public class PaymentVerification
{
    /// <summary>
    /// Verifies the Razorpay payment signature
    /// IMPORTANT: Use the order_id you stored when creating the order,
    /// NOT the razorpay_order_id from the client response (verify from your DB)
    /// </summary>
    public static bool VerifyPaymentSignature(
        string orderId, string paymentId, string signature)
    {
        string body = $"{orderId}|{paymentId}";
        
        using var hmac = new HMACSHA256(
            Encoding.UTF8.GetBytes(Environment.GetEnvironmentVariable("RAZORPAY_KEY_SECRET"))
        );
        
        byte[] hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(body));
        string expectedSignature = BitConverter.ToString(hash)
            .Replace("-", "").ToLower();
        
        // Use constant-time comparison to prevent timing attacks
        return CryptographicOperations.FixedTimeEquals(
            Encoding.UTF8.GetBytes(expectedSignature),
            Encoding.UTF8.GetBytes(signature)
        );
    }
}

// In your controller:
[HttpPost("api/verify-payment")]
public IActionResult VerifyPayment([FromBody] VerifyPaymentRequest request)
{
    // 1. Retrieve the original order_id from YOUR database
    var storedOrderId = GetOrderIdFromDatabase(request.RazorpayOrderId);
    
    if (storedOrderId == null)
    {
        return BadRequest(new { success = false, error = "Order not found" });
    }
    
    // 2. Verify the signature
    var isValid = PaymentVerification.VerifyPaymentSignature(
        storedOrderId,
        request.RazorpayPaymentId,
        request.RazorpaySignature
    );
    
    if (isValid)
    {
        // 3. Payment is verified - update your database
        UpdateOrderStatus(storedOrderId, "paid", request.RazorpayPaymentId);
        return Ok(new { success = true, message = "Payment verified" });
    }
    
    return BadRequest(new { success = false, error = "Invalid signature" });
}`,
}

// Checkout open templates by stack
// Note: These templates show client-side code that opens the Razorpay Checkout
// The checkout communicates directly with Razorpay's servers
var CheckoutOpenTemplates = map[string]string{
	StackWebStandard: `<!-- Web Standard Checkout Integration -->
<!-- Reference: https://razorpay.com/docs/payments/payment-gateway/web-integration/standard/ -->

<!-- Step 1: Include Razorpay checkout.js script (before closing </body> tag) -->
<script src="https://checkout.razorpay.com/v1/checkout.js"></script>

<!-- Step 2: JavaScript to open checkout -->
<script>
async function initiatePayment(amount) {
  // Step 1: Create order on YOUR backend
  // Your backend calls Razorpay's API: POST https://api.razorpay.com/v1/orders
  const response = await fetch('/api/create-order', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      amount: amount, // Amount in paise (e.g., 50000 for ₹500)
      currency: 'INR',
    }),
  });
  
  const orderData = await response.json();
  
  if (!orderData.success) {
    alert('Error creating order: ' + orderData.error);
    return;
  }
  
  // Step 2: Configure Razorpay checkout options
  const options = {
    key: orderData.key_id, // Key ID from your server (NOT the secret!)
    amount: orderData.amount,
    currency: orderData.currency,
    name: 'Your Company Name',
    description: 'Payment for Order',
    order_id: orderData.order_id, // Order ID from Razorpay (via your server)
    handler: function(response) {
      // Step 3: Payment successful - verify on YOUR backend
      verifyPayment(response);
    },
    prefill: {
      name: 'Customer Name',
      email: 'customer@example.com',
      contact: '9999999999',
    },
    theme: {
      color: '#3399cc',
    },
    modal: {
      ondismiss: function() {
        console.log('Payment modal closed');
      },
    },
  };
  
  // Step 3: Open Razorpay checkout
  const rzp = new Razorpay(options);
  rzp.on('payment.failed', function(response) {
    alert('Payment failed: ' + response.error.description);
  });
  rzp.open();
}

async function verifyPayment(response) {
  // Send payment details to YOUR backend for verification
  // Your backend verifies signature using: HMAC-SHA256(order_id|payment_id, key_secret)
  const verifyResponse = await fetch('/api/verify-payment', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      razorpay_order_id: response.razorpay_order_id,
      razorpay_payment_id: response.razorpay_payment_id,
      razorpay_signature: response.razorpay_signature,
    }),
  });
  
  const result = await verifyResponse.json();
  if (result.success) {
    alert('Payment successful!');
    // Redirect to success page or update UI
  } else {
    alert('Payment verification failed');
  }
}
</script>

<!-- Step 4: Add payment button -->
<button onclick="initiatePayment(50000)">Pay ₹500</button>`,

	StackAndroidStandard: `// Android Standard SDK Integration
// Reference: https://razorpay.com/docs/payments/payment-gateway/android-integration/standard/
//
// Step 1: Add dependency in app/build.gradle:
// implementation 'com.razorpay:checkout:1.6.33'
//
// Step 2: Add INTERNET permission in AndroidManifest.xml:
// <uses-permission android:name="android.permission.INTERNET" />

import com.razorpay.Checkout
import com.razorpay.PaymentResultListener
import org.json.JSONObject

class PaymentActivity : AppCompatActivity(), PaymentResultListener {

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Preload Razorpay checkout for faster loading
        Checkout.preload(applicationContext)
    }
    
    /**
     * Start payment flow
     * @param orderId - Order ID from YOUR backend (created via Razorpay API)
     * @param amount - Amount in paise
     * @param keyId - Razorpay Key ID (NOT the secret!)
     */
    fun startPayment(orderId: String, amount: Int, keyId: String) {
        val checkout = Checkout()
        checkout.setKeyID(keyId)
        
        try {
            val options = JSONObject().apply {
                put("name", "Your Company Name")
                put("description", "Payment for Order")
                put("order_id", orderId) // Order ID from Razorpay (via your backend)
                put("currency", "INR")
                put("amount", amount) // Amount in paise
                put("prefill", JSONObject().apply {
                    put("email", "customer@example.com")
                    put("contact", "9999999999")
                })
                put("theme", JSONObject().apply {
                    put("color", "#3399cc")
                })
            }
            
            // Open Razorpay checkout
            checkout.open(this, options)
        } catch (e: Exception) {
            Log.e("PaymentActivity", "Error: ${e.message}")
        }
    }
    
    override fun onPaymentSuccess(razorpayPaymentID: String?) {
        // Payment successful - verify on YOUR backend
        // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
        verifyPaymentOnServer(razorpayPaymentID)
    }
    
    override fun onPaymentError(code: Int, response: String?) {
        Log.e("PaymentActivity", "Payment failed: $code - $response")
        Toast.makeText(this, "Payment failed", Toast.LENGTH_SHORT).show()
    }
    
    private fun verifyPaymentOnServer(paymentId: String?) {
        // Make API call to YOUR backend to verify payment
        // Include: razorpay_order_id, razorpay_payment_id, razorpay_signature
    }
}`,

	StackIOSStandard: `// iOS Standard SDK Integration (Swift)
// Reference: https://razorpay.com/docs/payments/payment-gateway/ios-integration/standard/
//
// Step 1: Add via CocoaPods: pod 'razorpay-pod'
// Or Swift Package Manager: https://github.com/nickyhelali/razorpay-pod

import Razorpay

class PaymentViewController: UIViewController, RazorpayPaymentCompletionProtocol {
    
    var razorpay: RazorpayCheckout!
    
    override func viewDidLoad() {
        super.viewDidLoad()
        // Initialize Razorpay with your Key ID (NOT the secret!)
        razorpay = RazorpayCheckout.initWithKey(
            "YOUR_KEY_ID",
            andDelegate: self
        )
    }
    
    /**
     * Start payment flow
     * - orderId: Order ID from YOUR backend (created via Razorpay API)
     * - amount: Amount in paise
     * - keyId: Razorpay Key ID (NOT the secret!)
     */
    func startPayment(orderId: String, amount: Int, keyId: String) {
        let options: [String: Any] = [
            "name": "Your Company Name",
            "description": "Payment for Order",
            "order_id": orderId,  // Order ID from Razorpay (via your backend)
            "currency": "INR",
            "amount": amount,  // Amount in paise
            "prefill": [
                "email": "customer@example.com",
                "contact": "9999999999"
            ],
            "theme": [
                "color": "#3399cc"
            ]
        ]
        
        // Open Razorpay checkout
        razorpay.open(options)
    }
    
    // MARK: - RazorpayPaymentCompletionProtocol
    
    func onPaymentSuccess(_ payment_id: String) {
        // Payment successful - verify on YOUR backend
        // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
        print("Payment successful: \(payment_id)")
        verifyPaymentOnServer(paymentId: payment_id)
    }
    
    func onPaymentError(_ code: Int32, description str: String) {
        print("Payment failed: \(code) - \(str)")
        showAlert(message: "Payment failed: \(str)")
    }
    
    private func verifyPaymentOnServer(paymentId: String) {
        // Make API call to YOUR backend to verify payment
        // Include: razorpay_order_id, razorpay_payment_id, razorpay_signature
    }
    
    private func showAlert(message: String) {
        let alert = UIAlertController(
            title: "Payment Status",
            message: message,
            preferredStyle: .alert
        )
        alert.addAction(UIAlertAction(title: "OK", style: .default))
        present(alert, animated: true)
    }
}`,

	StackReactNativeStandard: `// React Native Standard SDK Integration
// Reference: https://razorpay.com/docs/payments/payment-gateway/react-native-integration/standard/
//
// Step 1: Install SDK
// npm install react-native-razorpay
// cd ios && pod install  // For iOS

import RazorpayCheckout from 'react-native-razorpay';

const PaymentScreen = () => {
  /**
   * Start payment flow
   * @param orderId - Order ID from YOUR backend (created via Razorpay API)
   * @param amount - Amount in paise
   * @param keyId - Razorpay Key ID (NOT the secret!)
   */
  const startPayment = async (orderId, amount, keyId) => {
    const options = {
      description: 'Payment for Order',
      image: 'https://your-logo-url.png',
      currency: 'INR',
      key: keyId,  // Key ID (NOT the secret!)
      amount: amount,  // Amount in paise
      name: 'Your Company Name',
      order_id: orderId,  // Order ID from Razorpay (via your backend)
      prefill: {
        email: 'customer@example.com',
        contact: '9999999999',
        name: 'Customer Name',
      },
      theme: { color: '#3399cc' },
    };

    try {
      // Open Razorpay checkout
      const data = await RazorpayCheckout.open(options);
      // Payment successful
      // data contains: razorpay_payment_id, razorpay_order_id, razorpay_signature
      console.log('Payment success:', data);
      verifyPaymentOnServer(data);
    } catch (error) {
      // Payment failed
      console.log('Payment failed:', error);
      Alert.alert('Payment Failed', error.description || 'Something went wrong');
    }
  };

  const verifyPaymentOnServer = async (paymentData) => {
    // Send payment details to YOUR backend for verification
    // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
    try {
      const response = await fetch('/api/verify-payment', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          razorpay_order_id: paymentData.razorpay_order_id,
          razorpay_payment_id: paymentData.razorpay_payment_id,
          razorpay_signature: paymentData.razorpay_signature,
        }),
      });
      const result = await response.json();
      if (result.success) {
        Alert.alert('Success', 'Payment verified!');
      }
    } catch (error) {
      console.error('Verification failed:', error);
    }
  };

  return (
    <Button title="Pay Now" onPress={() => startPayment(orderId, 50000, keyId)} />
  );
};`,

	StackFlutterStandard: `// Flutter Standard SDK Integration
// Reference: https://razorpay.com/docs/payments/payment-gateway/flutter-integration/standard/
//
// Step 1: Add dependency in pubspec.yaml:
// dependencies:
//   razorpay_flutter: ^1.3.5
//
// Step 2: For iOS, run: cd ios && pod install

import 'package:razorpay_flutter/razorpay_flutter.dart';

class PaymentScreen extends StatefulWidget {
  @override
  _PaymentScreenState createState() => _PaymentScreenState();
}

class _PaymentScreenState extends State<PaymentScreen> {
  late Razorpay _razorpay;

  @override
  void initState() {
    super.initState();
    _razorpay = Razorpay();
    _razorpay.on(Razorpay.EVENT_PAYMENT_SUCCESS, _handlePaymentSuccess);
    _razorpay.on(Razorpay.EVENT_PAYMENT_ERROR, _handlePaymentError);
    _razorpay.on(Razorpay.EVENT_EXTERNAL_WALLET, _handleExternalWallet);
  }

  @override
  void dispose() {
    _razorpay.clear();  // Important: Clear to prevent memory leaks
    super.dispose();
  }

  /// Start payment flow
  /// - orderId: Order ID from YOUR backend (created via Razorpay API)
  /// - amount: Amount in paise
  /// - keyId: Razorpay Key ID (NOT the secret!)
  void startPayment(String orderId, int amount, String keyId) {
    var options = {
      'key': keyId,  // Key ID (NOT the secret!)
      'amount': amount,  // Amount in paise
      'name': 'Your Company Name',
      'description': 'Payment for Order',
      'order_id': orderId,  // Order ID from Razorpay (via your backend)
      'prefill': {
        'contact': '9999999999',
        'email': 'customer@example.com',
      },
      'theme': {
        'color': '#3399cc',
      },
    };

    try {
      // Open Razorpay checkout
      _razorpay.open(options);
    } catch (e) {
      debugPrint('Error: $e');
    }
  }

  void _handlePaymentSuccess(PaymentSuccessResponse response) {
    // Payment successful - verify on YOUR backend
    // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
    print('Payment success: ${response.paymentId}');
    verifyPaymentOnServer(
      response.orderId!,
      response.paymentId!,
      response.signature!,
    );
  }

  void _handlePaymentError(PaymentFailureResponse response) {
    print('Payment failed: ${response.code} - ${response.message}');
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('Payment failed: ${response.message}')),
    );
  }

  void _handleExternalWallet(ExternalWalletResponse response) {
    print('External wallet: ${response.walletName}');
  }

  Future<void> verifyPaymentOnServer(
    String orderId,
    String paymentId,
    String signature,
  ) async {
    // Make API call to YOUR backend to verify payment
  }

  @override
  Widget build(BuildContext context) {
    return ElevatedButton(
      onPressed: () => startPayment(orderId, 50000, keyId),
      child: Text('Pay Now'),
    );
  }
}`,

	StackCordovaStandard: `// Cordova Standard Plugin Integration
// Reference: https://razorpay.com/docs/payments/payment-gateway/cordova-integration/
//
// Step 1: Install plugin
// cordova plugin add com.nicholaswilliams.nicepay.razorpay --save

/**
 * Start payment flow
 * @param orderId - Order ID from YOUR backend (created via Razorpay API: POST https://api.razorpay.com/v1/orders)
 * @param amount - Amount in paise
 * @param keyId - Razorpay Key ID (NOT the secret!)
 */
function startPayment(orderId, amount, keyId) {
  var options = {
    description: 'Payment for Order',
    image: 'https://your-logo-url.png',
    currency: 'INR',
    key: keyId,  // Key ID (NOT the secret!)
    amount: amount,  // Amount in paise
    name: 'Your Company Name',
    order_id: orderId,  // Order ID from Razorpay (via your backend)
    prefill: {
      email: 'customer@example.com',
      contact: '9999999999',
      name: 'Customer Name',
    },
    theme: {
      color: '#3399cc',
    },
  };

  var successCallback = function(payment_id) {
    // Payment successful - verify on YOUR backend
    // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
    console.log('Payment success:', payment_id);
    verifyPaymentOnServer(payment_id);
  };

  var errorCallback = function(error) {
    console.log('Payment failed:', error.code, error.description);
    alert('Payment failed: ' + error.description);
  };

  // Open Razorpay checkout
  RazorpayCheckout.open(options, successCallback, errorCallback);
}

function verifyPaymentOnServer(paymentId) {
  // Make API call to YOUR backend to verify payment
  // Include: razorpay_order_id, razorpay_payment_id, razorpay_signature
}`,

	StackIonicStandard: `// Ionic Standard Plugin Integration (uses Cordova plugin)
// Reference: https://razorpay.com/docs/payments/payment-gateway/cordova-integration/
//
// Step 1: Install plugin
// ionic cordova plugin add com.nicholaswilliams.nicepay.razorpay --save
// npm install @nickyhelali/nickyhelali-razorpay

import { Component } from '@angular/core';
import { RazorpayCheckout } from '@nickyhelali/nickyhelali-razorpay';

@Component({
  selector: 'app-payment',
  templateUrl: './payment.page.html',
})
export class PaymentPage {
  
  /**
   * Start payment flow
   * @param orderId - Order ID from YOUR backend (created via Razorpay API)
   * @param amount - Amount in paise
   * @param keyId - Razorpay Key ID (NOT the secret!)
   */
  async startPayment(orderId: string, amount: number, keyId: string) {
    const options = {
      description: 'Payment for Order',
      currency: 'INR',
      key: keyId,  // Key ID (NOT the secret!)
      amount: amount,  // Amount in paise
      name: 'Your Company Name',
      order_id: orderId,  // Order ID from Razorpay (via your backend)
      prefill: {
        email: 'customer@example.com',
        contact: '9999999999',
        name: 'Customer Name',
      },
      theme: {
        color: '#3399cc',
      },
    };

    try {
      // Open Razorpay checkout
      const response = await RazorpayCheckout.open(options);
      // Payment successful
      // response contains: razorpay_payment_id, razorpay_order_id, razorpay_signature
      console.log('Payment success:', response);
      await this.verifyPaymentOnServer(response);
    } catch (error) {
      console.log('Payment failed:', error);
      this.showAlert('Payment Failed', error.description || 'Something went wrong');
    }
  }

  async verifyPaymentOnServer(paymentData: any) {
    // Send payment details to YOUR backend for verification
    // Your backend verifies: HMAC-SHA256(order_id|payment_id, key_secret)
    const response = await fetch('/api/verify-payment', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        razorpay_order_id: paymentData.razorpay_order_id,
        razorpay_payment_id: paymentData.razorpay_payment_id,
        razorpay_signature: paymentData.razorpay_signature,
      }),
    });
    return response.json();
  }

  showAlert(title: string, message: string) {
    // Show alert using Ionic AlertController
  }
}`,
}

// Environment template
var EnvTemplate = `# Razorpay API Credentials
# IMPORTANT: Never commit this file to version control!
# Add .env to your .gitignore file

# =====================================================
# Razorpay API Reference: https://razorpay.com/docs/api/
# Orders API: POST https://api.razorpay.com/v1/orders
# =====================================================

# Test Mode Keys (use for development)
# Get these from: https://dashboard.razorpay.com/app/keys
RAZORPAY_KEY_ID=rzp_test_xxxxxxxxxxxx
RAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx

# Live Mode Keys (use for production)
# RAZORPAY_KEY_ID=rzp_live_xxxxxxxxxxxx
# RAZORPAY_KEY_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx

# Webhook Secret (optional, for webhook signature verification)
# Get this from: Dashboard > Settings > Webhooks
# RAZORPAY_WEBHOOK_SECRET=xxxxxxxxxxxxxxxxxxxxxxxx
`

// Webhook guidance
var WebhookGuidance = `# Razorpay Webhooks Setup Guide

## Overview
Webhooks allow Razorpay to notify YOUR server about events like successful payments,
failed payments, refunds, etc. This is essential for reliable payment tracking.

Reference: https://razorpay.com/docs/webhooks/

## Integration Flow Recap

1. YOUR Backend → Razorpay API: Create Order (POST https://api.razorpay.com/v1/orders)
2. Client → Razorpay Checkout: Customer makes payment
3. Razorpay → Client: Payment response with signature
4. Client → YOUR Backend: Send payment details for verification
5. YOUR Backend: Verify signature (HMAC-SHA256)
6. Razorpay → YOUR Backend (Webhook): Asynchronous payment notifications

## Steps to Configure Webhooks

1. **Login to Razorpay Dashboard**
   - Go to https://dashboard.razorpay.com/
   - Navigate to Settings → Webhooks

2. **Add Webhook Endpoint**
   - Click "Add New Webhook"
   - Enter YOUR webhook URL (e.g., https://yoursite.com/api/webhooks/razorpay)
   - Select events to subscribe to:
     - payment.authorized
     - payment.captured
     - payment.failed
     - order.paid
     - refund.created

3. **Set Webhook Secret**
   - Generate and copy the webhook secret
   - Store it in your environment variables as RAZORPAY_WEBHOOK_SECRET

4. **Verify Webhook Signature**
   - Always verify the webhook signature before processing
   - Use the X-Razorpay-Signature header
   - Compare HMAC SHA256 of request body with the signature

## Webhook Payload Example

{
  "entity": "event",
  "account_id": "acc_xxxxxxxxxxxx",
  "event": "payment.captured",
  "contains": ["payment"],
  "payload": {
    "payment": {
      "entity": {
        "id": "pay_xxxxxxxxxxxx",
        "order_id": "order_xxxxxxxxxxxx",
        "amount": 50000,
        "currency": "INR",
        "status": "captured"
      }
    }
  },
  "created_at": 1234567890
}

## Security Best Practices

1. Always verify webhook signatures
2. Use HTTPS for webhook endpoints
3. Implement idempotency (handle duplicate webhooks)
4. Respond with 200 status quickly, process asynchronously
5. Log all webhook events for debugging
`

// Security warnings for all integrations
var SecurityWarnings = []string{
	"Never expose RAZORPAY_KEY_SECRET on the client-side",
	"Always verify payment signatures on your server before fulfilling orders",
	"Use HTTPS for all API communications",
	"Store API credentials in environment variables, never in code",
	"Add .env files to .gitignore to prevent accidental commits",
	"Use the order_id from your server/database, not from client response, for signature verification",
	"Implement webhook signature verification for reliable payment status updates",
	"Use test mode keys during development, switch to live keys only in production",
	"The SDK calls Razorpay API (https://api.razorpay.com/v1/orders) - you don't call it directly",
}

