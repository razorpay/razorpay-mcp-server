"""Owned loopback regression for actual Go-generated Django output."""
from __future__ import annotations
import argparse, atexit, base64, hashlib, hmac, importlib.metadata, json, os
import socket, sys, tempfile, textwrap, threading
from pathlib import Path
from http.server import BaseHTTPRequestHandler, HTTPServer
from wsgiref.simple_server import make_server, WSGIRequestHandler

parser=argparse.ArgumentParser();parser.add_argument('--generated',required=True);parser.add_argument('--cases',default='');args=parser.parse_args()
output=json.loads(Path(args.generated).read_text(encoding='utf-8'))
work=tempfile.TemporaryDirectory(prefix='rzp-binding-');atexit.register(work.cleanup)
root=Path(work.name);app=root/'app';app.mkdir()
def put(path:str,content:str)->None:
 p=app/path;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(textwrap.dedent(content),encoding='utf-8')
for item in output['files']:
 if item['action']=='create' and item['path'].startswith('razorpay_payments/'):
  put(item['path'],item['code'])
fixed=(app/'razorpay_payments/checkout.py').exists()
put('fixture/__init__.py','');put('orders/__init__.py','')
put('fixture/settings.py',f"""
SECRET_KEY='owned-synthetic-django-key'
DEBUG=False
ALLOWED_HOSTS=['127.0.0.1','testserver']
INSTALLED_APPS=['django.contrib.contenttypes','django.contrib.sessions','orders','razorpay_payments']
DATABASES={{'default':{{'ENGINE':'django.db.backends.sqlite3','NAME':{str(root/'db.sqlite3')!r}}}}}
MIDDLEWARE=['django.contrib.sessions.middleware.SessionMiddleware']
ROOT_URLCONF='fixture.urls'
RAZORPAY_KEY_ID='rzp_test_owned'
RAZORPAY_KEY_SECRET='owned-synthetic-razorpay-key'
RAZORPAY_CHECKOUT_ADAPTER='orders.adapter.quote'
RAZORPAY_CHECKOUT_TTL=900
USE_TZ=True
DEFAULT_AUTO_FIELD='django.db.models.AutoField'
""")
put('orders/models.py',"""
from django.db import models
class Product(models.Model):
 sku=models.CharField(max_length=20,unique=True)
 price_paise=models.PositiveBigIntegerField()
class Order(models.Model):
 total=models.PositiveBigIntegerField()
 payment_status=models.CharField(max_length=16,default='pending')
 payment_method=models.CharField(max_length=16,default='cod')
 razorpay_order_id=models.CharField(max_length=64,null=True,blank=True)
 razorpay_payment_id=models.CharField(max_length=64,null=True,blank=True)
 razorpay_amount_paise=models.PositiveBigIntegerField(null=True)
 razorpay_currency=models.CharField(max_length=3,default='')
 razorpay_checkout_fingerprint=models.CharField(max_length=64,default='')
""")
put('orders/adapter.py',"""
from razorpay_payments.checkout import CheckoutQuote
from .models import Product,Order
def quote(request):
 cart=request.session['cart'];p=Product.objects.get(sku=cart['sku'])
 order=Order.objects.select_for_update().filter(pk=request.session.get('fixture_order')).first()
 if order is None:
  order=Order.objects.create(total=p.price_paise*cart['qty'])
  request.session['fixture_order']=order.pk
 return CheckoutQuote(order=order,amount_paise=p.price_paise*cart['qty'],currency='INR',snapshot={'sku':p.sku,'quantity':cart['qty'],'price':p.price_paise})
""")
common="""
import json
from django.http import JsonResponse
from django.middleware.csrf import get_token
from django.views.decorators.csrf import ensure_csrf_cookie
from django.db import transaction
from razorpay_payments import views as razorpay_views
from .models import Product,Order
def get_order_snapshot(request):
 if hasattr(razorpay_views,'get_order_snapshot'):
  return razorpay_views.get_order_snapshot(request)
 return None
@ensure_csrf_cookie
def home(request):
 request.session.setdefault('cart',{'sku':'B','qty':1});get_token(request)
 return JsonResponse({'ready':True})
def cart(request):
 request.session['cart']=json.loads(request.body)
 return JsonResponse({'ok':True})
def state(order):
 order.refresh_from_db()
 return {'id':order.pk,'total':order.total,'status':order.payment_status,'provider_order_id':order.razorpay_order_id,'payment_id':order.razorpay_payment_id}
def cod(request):
 return JsonResponse(state(Order.objects.create(total=100000,payment_method='cod')))
"""
if fixed:
 checkout="""
from razorpay_payments.checkout import complete_checkout,CheckoutError
def checkout(request):
 try:
  with transaction.atomic():
   order=complete_checkout(request,request.POST.get('razorpay_payment_id',''))
   if request.GET.get('rollback')=='1':raise RuntimeError('owned rollback control')
   result=state(order)
  return JsonResponse(result)
 except CheckoutError as exc:return JsonResponse({'error':str(exc)},status=409)
 except RuntimeError:return JsonResponse({'error':'owned rollback'},status=409)
"""
else:
 instructions=output['aiInstructions'].replace('\r\n','\n')
 markers=('Add this before order.save():\n','Find the checkout view, add before order.save():\n')
 marker=next((candidate for candidate in markers if candidate in instructions),None)
 if marker is None:raise RuntimeError('checkout instruction marker missing')
 snippet=textwrap.dedent(instructions.split(marker,1)[1].split('Replace order.total_amount',1)[0].split('The checkout form',1)[0])
 snippet=snippet.replace('order.total_amount * 100','order.total')
 checkout="def _checkout(request):\n    cart=request.session['cart']\n    product=Product.objects.get(sku=cart['sku'])\n    order=Order.objects.create(total=product.price_paise*cart['qty'])\n"+textwrap.indent(snippet,'    ')+"\n    order.save()\n    return JsonResponse(state(order))\n"
if not fixed:
 checkout += "\ndef checkout(request):\n try:\n  with transaction.atomic():\n   response=_checkout(request)\n   if request.GET.get('rollback')=='1':raise RuntimeError('owned rollback control')\n  return response\n except RuntimeError:return JsonResponse({'error':'owned rollback'},status=409)\n"
put('orders/views.py',common+'\n'+checkout)
put('fixture/urls.py',"""
from django.urls import include,path
from orders import views
urlpatterns=[path('',views.home),path('cart',views.cart),path('checkout',views.checkout),path('cod',views.cod),path('api/razorpay/',include('razorpay_payments.urls'))]
""")
for key in list(os.environ):
 if key.upper().startswith('RAZORPAY') or key.upper() in ('HTTP_PROXY','HTTPS_PROXY','ALL_PROXY','NO_PROXY'):os.environ.pop(key,None)
os.environ['DJANGO_SETTINGS_MODULE']='fixture.settings';sys.path.insert(0,str(app))
import django;django.setup()
from django.conf import settings
from django.core.management import call_command
from django.core.wsgi import get_wsgi_application
from django.db import connections
from orders.models import Product,Order
from razorpay_payments import views
import requests
original_quote=views.get_order_amount_paise

def trusted_quote(request):
 cart=request.session['cart'];p=Product.objects.get(sku=cart['sku'])
 return p.price_paise*cart['qty'],'INR'
views.get_order_amount_paise=trusted_quote
if hasattr(views,'get_order_snapshot'):
 def snapshot(request):
  cart=request.session['cart'];p=Product.objects.get(sku=cart['sku'])
  return {'sku':p.sku,'qty':cart['qty'],'price':p.price_paise}
 views.get_order_snapshot=snapshot
call_command('makemigrations','orders',verbosity=0);call_command('migrate',verbosity=0)
Product.objects.bulk_create([Product(sku='B',price_paise=100000),Product(sku='C',price_paise=100000)])
orders={};payments={};api_log=[];mode={};denied=[];allowed=set();sessions=[]
class API(BaseHTTPRequestHandler):
 def log_message(self,*args):pass
 def respond(self,code,data):
  body=json.dumps(data).encode();self.send_response(code);self.send_header('Content-Type','application/json');self.send_header('Content-Length',str(len(body)));self.end_headers();self.wfile.write(body)
 def checkauth(self):
  want='Basic '+base64.b64encode((settings.RAZORPAY_KEY_ID+':'+settings.RAZORPAY_KEY_SECRET).encode()).decode()
  if self.headers.get('Authorization')!=want:raise AssertionError('fixture auth mismatch')
 def do_POST(self):
  self.checkauth();assert self.path=='/v1/orders'
  body=json.loads(self.rfile.read(int(self.headers['Content-Length'])));oid='order_'+str(len(orders)+1)
  item={'id':oid,'amount':body['amount'],'currency':body['currency'],'status':'created'}
  item.update(mode.get('create_override',{}));orders[oid]=item;api_log.append({'method':'POST','path':self.path,'amount':body['amount'],'currency':body['currency']});self.respond(200,item)
 def do_GET(self):
  self.checkauth();assert self.path.startswith('/v1/payments/')
  pid=self.path.rsplit('/',1)[-1];api_log.append({'method':'GET','path':self.path})
  if mode.get('fetch_error'):return self.respond(503,{'error':{'code':'SERVER_ERROR','description':'fixture private canary'}})
  return self.respond(200,payments[pid])
class Quiet(WSGIRequestHandler):
 def log_message(self,*args):pass
api=HTTPServer(('127.0.0.1',0),API);server=make_server('127.0.0.1',0,get_wsgi_application(),handler_class=Quiet)
allowed.update([('127.0.0.1',api.server_port),('127.0.0.1',server.server_port)])
def audit(event,data):
 if event=='socket.connect' and (not isinstance(data[1],tuple) or data[1][:2] not in allowed):
  denied.append(str(data[1]));raise RuntimeError('nonfixture destination blocked')
sys.addaudithook(audit)
views.client.base_url=f'http://127.0.0.1:{api.server_port}';views.client.session.trust_env=False
threads=[threading.Thread(target=s.serve_forever,daemon=True) for s in (api,server)]
for t in threads:t.start()
base=f'http://127.0.0.1:{server.server_port}'
def fresh():
 s=requests.Session();s.trust_env=False;s.get(base+'/',timeout=5).raise_for_status();s.headers['X-CSRFToken']=s.cookies['csrftoken'];sessions.append(s);return s
def post(s,path,payload=None,form=False):
 q=s.post(base+path,**({'data':payload} if form else {'json':payload}),timeout=8)
 try:data=q.json()
 except ValueError:data={'error':'non-json','http':q.status_code}
 return q.status_code,data
def create(s):
 code,item=post(s,'/api/razorpay/order',{'amount':1,'currency':'USD'});assert code==200,(code,item);return item
def receipt(order,**override):
 oid=order['orderId'];pid='pay_'+str(len(payments)+1)
 payments[pid]={'id':pid,'order_id':oid,'amount':order['amount'],'currency':order['currency'],'status':'captured','captured':True};payments[pid].update(override)
 sig=hmac.new(settings.RAZORPAY_KEY_SECRET.encode(),(oid+'|'+pid).encode(),hashlib.sha256).hexdigest()
 return {'razorpay_order_id':oid,'razorpay_payment_id':pid,'razorpay_signature':sig}
def verify(s,r):return post(s,'/api/razorpay/verify',r)
def checkout(s,r,extra=''):return post(s,'/checkout'+extra,{'payment_method':'razorpay','razorpay_payment_id':r['razorpay_payment_id']},True)
results=[];http_trace=[]
def rows():return list(Order.objects.order_by('id').values())
original_post=post
def post(s,path,payload=None,form=False):
 code,body=original_post(s,path,payload,form)
 http_trace.append({'path':path,'code':code,'response':body})
 return code,body

def check(name,fn):
 if args.cases and name not in args.cases.split(','):return
 before=rows();start=len(http_trace);api_start=len(api_log)
 try:details=fn();item={'id':name,'pass':True,'details':details}
 except Exception as exc:item={'id':name,'pass':False,'error':str(exc),'exception':type(exc).__name__}
 item.update({'before_rows':before,'after_rows':rows(),'http':http_trace[start:],'provider_requests':api_log[api_start:]})
 results.append(item)
def authority():
 s=fresh();o=create(s);assert (o['amount'],o['currency'])==(100000,'INR'),o;return o
def positive():
 s=fresh();o=create(s);r=receipt(o);assert verify(s,r)[0]==200
 code,body=checkout(s,r);assert code==200 and body['status']=='paid',body
 repeat=checkout(s,r);assert repeat[0] in (400,409),repeat;return {'paid':body,'repeat':repeat}
def reject(label,overrides=None,receipt_change=None,cart_change=False):
 s=fresh();o=create(s);r=receipt(o,**(overrides or {}));r.update(receipt_change or {})
 if cart_change:post(s,'/cart',{'sku':'C','qty':1})
 code,body=verify(s,r)
 after=checkout(s,r)
 assert code in (400,409,502,503),(label,code,body,after)
 assert after[0] in (400,409),after
 assert not Order.objects.filter(razorpay_payment_id=r['razorpay_payment_id'],payment_status='paid').exists()
 return {'verify':code,'checkout':after[0]}
def substitution():
 s=fresh();create(s);r=receipt({'orderId':'order_oldlow','amount':100,'currency':'INR'})
 code,body=verify(s,r)
 after=checkout(s,r);assert code in (400,409),(code,body,after)
 assert after[0] in (400,409),after;return {'verify':code,'checkout':after[0]}
def changed_after_verify():
 s=fresh();o=create(s);r=receipt(o);assert verify(s,r)[0]==200
 post(s,'/cart',{'sku':'C','qty':1});code,body=checkout(s,r);assert code in (400,409),body;return {'checkout':code}
def replay():
 s=fresh();o=create(s);r=receipt(o);assert verify(s,r)[0]==200;assert checkout(s,r)[0]==200
 s2=fresh();create(s2);code,body=verify(s2,r);assert code in (400,409),(code,body);return {'verify_new_session':code}
def csrf():
 s=fresh();s.headers.pop('X-CSRFToken');a=post(s,'/api/razorpay/order',{'amount':1})[0];b=post(s,'/api/razorpay/verify',{})[0];assert a==403 and b==403,(a,b);return {'order':a,'verify':b}
def rollback():
 s=fresh();o=create(s);r=receipt(o);assert verify(s,r)[0]==200
 before=Order.objects.filter(payment_status='paid').count();code,body=checkout(s,r,'?rollback=1');assert code==409,(code,body)
 assert Order.objects.filter(payment_status='paid').count()==before
 code,body=checkout(s,r);assert code==200 and body['status']=='paid',(code,body);return body
def fetch_error():
 s=fresh();o=create(s);r=receipt(o);mode['fetch_error']=True
 try:code,body=verify(s,r)
 finally:mode.pop('fetch_error',None)
 assert code in (502,503),(code,body);assert 'canary' not in json.dumps(body)
 assert checkout(s,r)[0] in (400,409);return {'verify':code}
def unconfigured():
 views.get_order_amount_paise=original_quote
 try:code,body=post(fresh(),'/api/razorpay/order',{'amount':1})
 finally:views.get_order_amount_paise=trusted_quote
 assert code==503,(code,body);return {'code':code,'property':'original mandatory server-price stub fails closed'}
def bad_order_response():
 mode['create_override']={'amount':1}
 try:code,body=post(fresh(),'/api/razorpay/order',{'amount':1})
 finally:mode.pop('create_override',None)
 assert code in (502,503),(code,body);return {'code':code}
def cod():
 code,body=post(fresh(),'/cod',{},True);assert code==200 and body['status']=='pending';return body
try:
 check('client_amount_currency_ignored',authority)
 check('valid_captured_payment_and_consumed_repeat',positive)
 check('valid_foreign_low_receipt_rejected',substitution)
 for name,extra in [('amount',{'amount':100}),('currency',{'currency':'USD'}),('provider_order',{'order_id':'order_other'}),('authorized_not_captured',{'status':'authorized','captured':False})]:
  check('payment_'+name+'_mismatch',lambda extra=extra,name=name:reject(name,extra))
 check('tampered_signature',lambda:reject('signature',receipt_change={'razorpay_signature':'0'*64}))
 check('missing_signature',lambda:reject('missing',receipt_change={'razorpay_signature':''}))
 check('same_price_cart_change_before_verify',lambda:reject('cart',cart_change=True))
 check('same_price_cart_change_after_verify',changed_after_verify)
 check('same_receipt_new_session',replay)
 check('csrf_required',csrf)
 check('business_failure_rollback_then_retry',rollback)
 check('provider_error_not_paid_or_leaked',fetch_error)
 check('adapter_unconfigured_fails_closed',unconfigured)
 check('unexpected_provider_order_response',bad_order_response)
 check('cod_remains_unpaid',cod)
finally:
 for s in (server,api):s.shutdown();s.server_close()
 for t in threads:t.join(timeout=4)
 for s in sessions:s.close()
 views.client.session.close();connections.close_all()
closed=[]
for port in (api.server_port,server.server_port):
 sock=socket.socket();sock.settimeout(.2);closed.append(sock.connect_ex(('127.0.0.1',port))!=0);sock.close()
healthy=all(closed) and not denied and all(not t.is_alive() for t in threads)
summary={'schema':'django-binding-regression/1','fixed_generator':fixed,'versions':{'python':sys.version.split()[0],'django':django.get_version(),'razorpay':importlib.metadata.version('razorpay')},'cases':results,'passed':sum(x['pass'] for x in results),'total':len(results),'cleanup':{'listeners_closed':closed,'threads_joined':all(not t.is_alive() for t in threads),'denied_connections':denied},'provider_requests':api_log,'boundary':'Actual Go-generated Django files, HTTP WSGI, SQLite sessions/ORM and SDK HMAC verifier; owned catalog adapter and loopback provider; no live Razorpay/fulfilment.'}
if os.environ.get('RZP_DJANGO_RESULT'):
 Path(os.environ['RZP_DJANGO_RESULT']).write_text(json.dumps(summary,ensure_ascii=False,indent=2),encoding='utf-8')
print(json.dumps(summary,ensure_ascii=False))
sys.exit(0 if healthy and all(x['pass'] for x in results) else 1)
