//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

func getDjangoIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	viewsCode := `import json
import time
import razorpay
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

        # Use Razorpay SDK to verify signature
        client.utility.verify_payment_signature({
            'razorpay_order_id': razorpay_order_id,
            'razorpay_payment_id': razorpay_payment_id,
            'razorpay_signature': razorpay_signature,
        })

        return JsonResponse({
            'success': True,
            'message': 'Payment verified',
            'paymentId': razorpay_payment_id,
            'orderId': razorpay_order_id,
        })
    except razorpay.errors.SignatureVerificationError:
        return JsonResponse({'success': False, 'error': 'Invalid signature'}, status=400)
    except Exception as e:
        return JsonResponse({'success': False, 'error': str(e)}, status=500)
`

	urlsCode := `from django.urls import path
from . import views

urlpatterns = [
    path('order', views.create_order, name='razorpay_order'),
    path('verify', views.verify_payment, name='razorpay_verify'),
]
`

	initCode := `# Razorpay payments app
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for Django + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "razorpay_payments/__init__.py", Code: initCode, Description: "Django app init file"},
			{Action: "create", Path: "razorpay_payments/views.py", Code: viewsCode, Description: "Django views for Razorpay"},
			{Action: "create", Path: "razorpay_payments/urls.py", Code: urlsCode, Description: "Django URL patterns"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{
				Action:      "manual_edit",
				Path:        "requirements.txt",
				Description: "Add razorpay to requirements.txt",
				Edits: []EditItem{
					{Line: "At end of file", Add: "razorpay>=1.2.0", Why: "Razorpay Python SDK"},
				},
			},
			{
				Action:      "manual_edit",
				Path:        "settings.py",
				Description: "Add to settings.py",
				Edits: []EditItem{
					{Line: "In INSTALLED_APPS list", Add: "'razorpay_payments',", Why: "Register the Razorpay app"},
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
		AIInstructions: `
███████████████████████████████████████████████████████████████████████████████
█ IMPORTANT: USE THE CODE FROM "Files" SECTION - DON'T REPLACE IT             █
███████████████████████████████████████████████████████████████████████████████

The "Files" section above contains TESTED, WORKING code. Use it as your primary source.
You MAY use your knowledge to integrate it properly with the existing project.
You MUST NOT replace the provided code with generic boilerplate.
You MUST NOT create new models or run migrations - the project already has what it needs.
You MUST complete everything - no "Next Steps" for the user.

═══════════════════════════════════════════════════════════════════════════════
🚫 DO NOT DO THESE:
═══════════════════════════════════════════════════════════════════════════════
- DO NOT delete or revert existing files/code - you are ADDING to the project
- DO NOT create Order model or PaymentOrder model (project already has Order model)
- DO NOT create razorpay_checkout.html (modify EXISTING checkout template instead)
- DO NOT run migrations unless model has no razorpay choice yet
- DO NOT hardcode keys like "YOUR_KEY_ID" (use os.environ.get())

═══════════════════════════════════════════════════════════════════════════════
🎯 GOAL: Add Razorpay as a payment option to the EXISTING checkout page
═══════════════════════════════════════════════════════════════════════════════
- The site already has a working checkout at /orders/checkout/
- You are ADDING Razorpay as a second payment option alongside COD
- Users should see both "Cash on Delivery" and "Razorpay" radio buttons
- DO NOT create a separate payment page - MODIFY the existing checkout

═══════════════════════════════════════════════════════════════════════════════
FILES TO CREATE (new files):
═══════════════════════════════════════════════════════════════════════════════
1. razorpay_payments/__init__.py (empty file)
2. razorpay_payments/views.py (Razorpay order creation + verification endpoints)
3. razorpay_payments/urls.py (URL routing for the above)
4. static/js/razorpay.js (frontend payment handling)
5. .env (with ACTUAL keys provided - not placeholders!)

═══════════════════════════════════════════════════════════════════════════════
FILES TO MODIFY (existing files - make MINIMAL changes):
═══════════════════════════════════════════════════════════════════════════════
1. config/settings.py:
   - Add 'razorpay_payments' to INSTALLED_APPS list (check if not already there!)
   - Add RAZORPAY_KEY_ID/SECRET using os.environ.get() at bottom
   - DO NOT use INSTALLED_APPS += [...] syntax - edit the list directly
2. config/urls.py (or main urls.py) - add path('api/razorpay/', include('razorpay_payments.urls'))
3. orders/views.py - add: order.payment_method = request.POST.get('payment_method', 'cod')
4. templates/orders/checkout.html - add payment method radios + razorpay.js script tag
5. requirements.txt - add 'razorpay' line (if not already there)

═══════════════════════════════════════════════════════════════════════════════
STEP-BY-STEP INSTRUCTIONS:
═══════════════════════════════════════════════════════════════════════════════

**STEP 1: INSTALL RAZORPAY SDK**
- Check for venv/, .venv/, or env/ folder
- If venv/ exists: ./venv/bin/pip install razorpay
- If .venv/ exists: ./.venv/bin/pip install razorpay
- If no venv: pip3 install razorpay
- Add 'razorpay' to requirements.txt

**STEP 2: CREATE razorpay_payments APP**
Create these 3 files:
- razorpay_payments/__init__.py (empty)
- razorpay_payments/views.py (use the provided code)
- razorpay_payments/urls.py (use the provided code)

**STEP 3: UPDATE settings.py**
Add to INSTALLED_APPS: 'razorpay_payments'
Add at bottom (MUST use os.environ.get, NOT hardcoded strings!):
  RAZORPAY_KEY_ID = os.environ.get('RAZORPAY_KEY_ID')
  RAZORPAY_KEY_SECRET = os.environ.get('RAZORPAY_KEY_SECRET')

**STEP 4: CREATE .env FILE**
Create .env in project root with the ACTUAL keys provided:
  RAZORPAY_KEY_ID=<actual key from tool output>
  RAZORPAY_KEY_SECRET=<actual secret from tool output>

**STEP 5: UPDATE main urls.py**
Add: path('api/razorpay/', include('razorpay_payments.urls'))

**STEP 6: UPDATE orders/views.py (MINIMAL CHANGE)**
Find the checkout view, add before order.save():
  order.payment_method = request.POST.get('payment_method', 'cod')
  if order.payment_method == 'razorpay':
      order.payment_status = 'paid'

**STEP 6b: CHECK orders/models.py FOR PAYMENT_METHOD_CHOICES**
If the Order model has PAYMENT_METHOD_CHOICES, add ('razorpay', 'Razorpay') to it:
  PAYMENT_METHOD_CHOICES = [
      ('cod', 'Cash on Delivery'),
      ('razorpay', 'Razorpay'),  # <-- ADD THIS
  ]
Then run: ./venv/bin/python manage.py makemigrations && ./venv/bin/python manage.py migrate

**STEP 7: CREATE static/js/razorpay.js**
Use the provided razorpay.js code (from Files section)

**STEP 8: UPDATE THE EXISTING CHECKOUT TEMPLATE (READ IT FIRST!)**
READ the checkout template to understand its structure before making changes.

PRINCIPLES TO FOLLOW:
1. {% load static %} must be at TOP with other {% load %} tags (after {% extends %})
2. Payment method radios MUST be INSIDE the <form> tag - find where payment options belong
3. Add a hidden input for the cart total so JS can read it
4. Scripts go in the JS block ({% block extra_js %} or similar) - NOT loose after </form>

CORRECT JS FLOW (CRITICAL!):
- User selects Razorpay and clicks submit
- JS intercepts, calls initiateRazorpayPayment() -> Razorpay modal opens
- User completes payment in modal -> payment verified server-side
- ONLY THEN submit the form to create Django order
- WRONG: Creating Django order first via fetch(), then paying (order exists before payment!)

WHAT TO ADD:
- {% load static %} at top if not present
- Payment method radio buttons (cod + razorpay) - find appropriate spot INSIDE form
- Hidden input with cart total value for JS to read
- Script tag for razorpay.js
- Form submit handler that calls initiateRazorpayPayment() for Razorpay option

CRITICAL RULES:
- Payment radios MUST be INSIDE <form>, not after </form> (or they won't submit)
- Find the actual element IDs/classes by reading the template - don't guess
- Use the template's existing patterns (CSS classes, structure) for consistency

═══════════════════════════════════════════════════════════════════════════════
⛔ FORBIDDEN - NEVER DO THESE:
═══════════════════════════════════════════════════════════════════════════════
❌ NEVER create Order/PaymentOrder models - project already has Order model
❌ NEVER run makemigrations or migrate - NO DATABASE CHANGES NEEDED
❌ NEVER create razorpay_checkout.html - modify EXISTING templates/orders/checkout.html
❌ NEVER create a separate payment page - ADD Razorpay option to existing checkout
❌ NEVER hardcode "YOUR_KEY_ID" or "your_key_secret" - use os.environ.get()
❌ NEVER say "Next Steps" or ask user to do things manually - YOU complete everything
❌ NEVER replace the provided Files code with generic boilerplate
❌ NEVER forget __init__.py in razorpay_payments/
❌ NEVER add razorpay_payments to INSTALLED_APPS twice - check first!
❌ NEVER use INSTALLED_APPS += [...] - edit the list directly to avoid duplicates
❌ NEVER add scripts loose after </form> - put them in the JS block ({% block extra_js %} etc)
❌ NEVER add form inputs OUTSIDE <form> - radios/hidden fields must be INSIDE to be submitted
❌ NEVER reference elements without checking they exist - READ the template first!
❌ NEVER use {% static %} without {% load static %} at the top of the template
❌ NEVER set order.razorpay_payment_id - this field doesn't exist on the model
❌ NEVER break the existing COD payment flow - both options must work
❌ NEVER delete or revert existing files - you are ADDING integration, not replacing
❌ NEVER create Django order BEFORE Razorpay payment - payment must happen FIRST
❌ NEVER forget to add ('razorpay', 'Razorpay') to PAYMENT_METHOD_CHOICES if model has it

═══════════════════════════════════════════════════════════════════════════════
FINAL CHECKLIST - ALL MUST BE YES:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you install razorpay using the correct venv pip path?
✅ Did you create razorpay_payments/__init__.py?
✅ Did you add 'razorpay_payments' to INSTALLED_APPS?
✅ Did you use os.environ.get() for credentials (NOT hardcoded)?
✅ Did you create .env with the ACTUAL keys (not placeholders)?
✅ Did you create static/js/razorpay.js?
✅ Did you MODIFY templates/orders/checkout.html (not create a new template)?
✅ Did you add payment method radio buttons to the EXISTING checkout form?
✅ Does the existing COD flow still work?

If ANY answer is NO, GO BACK AND FIX IT NOW.` + getFrontendWiringInstructions(frontend),
	}
}

func getFlaskIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	appCode := `import os
import time
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

        # Use Razorpay SDK to verify signature
        client.utility.verify_payment_signature({
            'razorpay_order_id': razorpay_order_id,
            'razorpay_payment_id': razorpay_payment_id,
            'razorpay_signature': razorpay_signature,
        })

        return jsonify({'success': True, 'paymentId': razorpay_payment_id, 'orderId': razorpay_order_id})
    except razorpay.errors.SignatureVerificationError:
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
			{
				Action:      "manual_edit",
				Path:        "requirements.txt",
				Description: "Add dependencies to requirements.txt",
				Edits: []EditItem{
					{Line: "At end of file", Add: "razorpay>=1.2.0", Why: "Razorpay Python SDK"},
					{Line: "At end of file", Add: "python-dotenv>=1.0.0", Why: "Environment variable loading"},
				},
			},
			{Action: "manual_edit", Path: "app.py", Description: "Register Razorpay routes in main Flask app", Edits: []EditItem{
				{Line: "After imports", Add: "from razorpay_routes import app as razorpay_app", Why: "Import Razorpay routes"},
				{Line: "After app creation", Add: "app.register_blueprint(razorpay_app) # or import routes directly", Why: "Register routes"},
			}},
			getWirePaymentAction(),
		},
		Dependencies: []Dependency{
			{Name: "razorpay", InstallCommand: "pip install razorpay"},
			{Name: "python-dotenv", InstallCommand: "pip install python-dotenv"},
		},
		EnvVars:          []EnvVar{{Name: "RAZORPAY_KEY_ID", Value: keyID}, {Name: "RAZORPAY_KEY_SECRET", Value: keySecret}},
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: CHECK FOR VIRTUAL ENVIRONMENT AND INSTALL DEPENDENCIES**
- Look for venv/, .venv/, env/ folders in project root
- If venv/ exists: ./venv/bin/pip install razorpay python-dotenv
- If .venv/ exists: ./.venv/bin/pip install razorpay python-dotenv
- If no venv: pip3 install razorpay python-dotenv
- Add 'razorpay' and 'python-dotenv' to requirements.txt

**STEP 2: CREATE razorpay_routes.py**
- Create razorpay_routes.py with the provided Razorpay endpoints
- Register the routes in the main Flask app (find app.py or main file and import/register the routes)

**STEP 3: CREATE .env FILE**
- CREATE a new .env file in project root with placeholder keys
- The user will fill in actual credentials from the Razorpay Dashboard

**STEP 4: TEMPLATE INTEGRATION**
- Put razorpay.js in static/js/ folder
- Add script tag in template
- Modify checkout to call initiateRazorpayPayment() before submitting

═══════════════════════════════════════════════════════════════════════════════
⛔ FORBIDDEN - DO NOT DO ANY OF THESE:
═══════════════════════════════════════════════════════════════════════════════
❌ NEVER modify database models - the integration works without adding fields
❌ NEVER run migrations - they cause conflicts and are not needed
❌ DO NOT delete the .env file
❌ DO NOT use bare except: - catch specific exceptions

═══════════════════════════════════════════════════════════════════════════════
FINAL CHECKLIST:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you run pip install razorpay python-dotenv?
✅ Did you add dependencies to requirements.txt?
✅ Did you create razorpay_routes.py?
✅ Did you create .env file with actual keys?

If ANY checkbox is NO, GO BACK AND FIX IT NOW.` + getFrontendWiringInstructions(frontend),
	}
}

func getFastAPIIntegration(frontend FrontendIntegration) IntegrateCheckoutOutput {
	keyID, keySecret := placeholderKeyID, placeholderKeySecret

	routerCode := `import os
import time
import razorpay
from typing import Optional
from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from dotenv import load_dotenv

load_dotenv()

router = APIRouter(prefix="/api/razorpay")
client = razorpay.Client(auth=(os.environ['RAZORPAY_KEY_ID'], os.environ['RAZORPAY_KEY_SECRET']))

class OrderRequest(BaseModel):
    amount: float
    currency: str = "INR"
    receipt: Optional[str] = None

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
    try:
        # Use Razorpay SDK to verify signature
        client.utility.verify_payment_signature({
            'razorpay_order_id': req.razorpay_order_id,
            'razorpay_payment_id': req.razorpay_payment_id,
            'razorpay_signature': req.razorpay_signature,
        })
        return {'success': True, 'paymentId': req.razorpay_payment_id, 'orderId': req.razorpay_order_id}
    except razorpay.errors.SignatureVerificationError:
        raise HTTPException(status_code=400, detail="Invalid signature")
`

	return IntegrateCheckoutOutput{
		Summary: "Complete Razorpay Standard Checkout integration for FastAPI + " + frontend.Framework,
		Files: []FileAction{
			{Action: "create", Path: "routers/razorpay.py", Code: routerCode, Description: "FastAPI router for Razorpay"},
			{Action: "create", Path: frontend.FileName, Code: frontend.Code, Description: frontend.Description},
			{
				Action:      "manual_edit",
				Path:        "requirements.txt",
				Description: "Add dependencies to requirements.txt",
				Edits: []EditItem{
					{Line: "At end of file", Add: "razorpay>=1.2.0", Why: "Razorpay Python SDK"},
					{Line: "At end of file", Add: "python-dotenv>=1.0.0", Why: "Environment variable loading"},
				},
			},
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
		TestInstructions: "Use test card: 4111 1111 1111 1111, any future expiry, any CVV. Test UPI: success@razorpay",
		AIInstructions: `⚠️ MANDATORY: COMPLETE THE ENTIRE INTEGRATION - NO "NEXT STEPS" ALLOWED!

CRITICAL INSTRUCTIONS - FOLLOW THESE EXACT STEPS:

**STEP 1: CHECK FOR VIRTUAL ENVIRONMENT AND INSTALL DEPENDENCIES**
- Look for venv/, .venv/, env/ folders in project root
- If venv/ exists: ./venv/bin/pip install razorpay python-dotenv
- If .venv/ exists: ./.venv/bin/pip install razorpay python-dotenv
- If no venv: pip3 install razorpay python-dotenv
- Add 'razorpay' and 'python-dotenv' to requirements.txt

**STEP 2: CREATE RAZORPAY ROUTER**
- Create routers/razorpay.py with the provided endpoints
- Include router in main app: app.include_router(razorpay_router)

**STEP 3: CREATE .env FILE WITH ACTUAL KEYS (MUST DO THIS)**
- CREATE a new .env file in project root with:
  RAZORPAY_KEY_ID=<the actual key provided>
  RAZORPAY_KEY_SECRET=<the actual secret provided>

**STEP 4: TEMPLATE/FRONTEND INTEGRATION**
- Put razorpay.js in static/js/ folder
- Ensure StaticFiles mount exists
- Modify checkout to call initiateRazorpayPayment() before submitting

═══════════════════════════════════════════════════════════════════════════════
⛔ FORBIDDEN - DO NOT DO ANY OF THESE:
═══════════════════════════════════════════════════════════════════════════════
❌ NEVER modify database models - the integration works without adding fields
❌ NEVER run migrations - they cause conflicts and are not needed
❌ DO NOT delete the .env file
❌ DO NOT use bare except: - catch specific exceptions

═══════════════════════════════════════════════════════════════════════════════
FINAL CHECKLIST:
═══════════════════════════════════════════════════════════════════════════════
✅ Did you run pip install razorpay python-dotenv?
✅ Did you add dependencies to requirements.txt?
✅ Did you create routers/razorpay.py?
✅ Did you create .env file with actual keys?

If ANY checkbox is NO, GO BACK AND FIX IT NOW.` + getFrontendWiringInstructions(frontend),
	}
}
