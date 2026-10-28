//nolint:lll // File contains embedded code templates requiring longer lines
package integrations

func getFrontendIntegration(framework string) FrontendIntegration {
	switch framework {
	case "react":
		return getReactFrontend()
	case "vue":
		return getVueFrontend()
	case "angular":
		return getAngularFrontend()
	case "svelte":
		return getSvelteFrontend()
	case "solid":
		return getSolidFrontend()
	case "native":
		return FrontendIntegration{
			Framework:   "Native Mobile",
			Code:        "// See mobile-specific integration",
			FileName:    "N/A",
			ScriptTag:   "Use native SDK",
			Description: "Native mobile integration",
		}
	default: // vanilla
		return getVanillaFrontend()
	}
}

func getVanillaFrontend() FrontendIntegration {
	code := `// Razorpay Payment Integration
async function initiateRazorpayPayment(amount, onSuccess, onError) {
  try {
    if (!window.Razorpay) {
      await new Promise((resolve, reject) => {
        const script = document.createElement('script');
        script.src = 'https://checkout.razorpay.com/v1/checkout.js';
        script.onload = resolve;
        script.onerror = reject;
        document.head.appendChild(script);
      });
    }

    const orderResponse = await fetch('/api/razorpay/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount }),
    });

    const orderData = await orderResponse.json();
    if (!orderData.success) throw new Error(orderData.error || 'Failed to create order');

    const options = {
      key: orderData.keyId,
      amount: orderData.amount,
      currency: orderData.currency,
      name: document.title || 'Payment',
      order_id: orderData.orderId,
      handler: async function(response) {
        const verifyResponse = await fetch('/api/razorpay/verify', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(response),
        });
        const verifyData = await verifyResponse.json();
        if (verifyData.success) { if (onSuccess) onSuccess(verifyData); }
        else { if (onError) onError(new Error(verifyData.error)); }
      },
      modal: { ondismiss: () => { if (onError) onError(new Error('Payment cancelled')); } },
      theme: { color: '#528FF0' },
    };

    const razorpay = new window.Razorpay(options);
    razorpay.on('payment.failed', (r) => { if (onError) onError(new Error(r.error.description)); });
    razorpay.open();
  } catch (error) {
    console.error('Payment failed:', error);
    if (onError) onError(error);
  }
}
`
	return FrontendIntegration{
		Framework:   "Vanilla JS",
		Code:        code,
		FileName:    "public/js/razorpay.js",
		ScriptTag:   "Add <script src=\"/js/razorpay.js\"></script> to the CHECKOUT HTML file (find which HTML has the checkout - may be checkout.html, cart.html, NOT just index.html)",
		Description: "Vanilla JS Razorpay payment helper",
	}
}

func getReactFrontend() FrontendIntegration {
	code := `import { useState, useEffect } from 'react';

export function useRazorpay() {
  const [loading, setLoading] = useState(false);
  const [scriptLoaded, setScriptLoaded] = useState(false);

  useEffect(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => setScriptLoaded(true);
    document.body.appendChild(script);
    return () => document.body.removeChild(script);
  }, []);

  const pay = async (amount, onSuccess, onError) => {
    if (!scriptLoaded || loading) return;
    setLoading(true);
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? onSuccess?.(result) : onError?.(new Error(result.error));
          setLoading(false);
        },
        modal: { ondismiss: () => setLoading(false) },
      };
      new window.Razorpay(options).open();
    } catch (e) {
      onError?.(e);
      setLoading(false);
    }
  };

  return { pay, loading, ready: scriptLoaded };
}

export function RazorpayButton({ amount, onSuccess, onError, children }) {
  const { pay, loading, ready } = useRazorpay();
  return (
    <button onClick={() => pay(amount, onSuccess, onError)} disabled={!ready || loading}>
      {loading ? 'Processing...' : children || 'Pay Now'}
    </button>
  );
}
`
	return FrontendIntegration{
		Framework:   "React",
		Code:        code,
		FileName:    "src/components/RazorpayButton.jsx",
		ScriptTag:   "Import and use <RazorpayButton amount={100} onSuccess={...} />",
		Description: "React hook and component for Razorpay payments",
	}
}

func getVueFrontend() FrontendIntegration {
	code := `<template>
  <button @click="pay" :disabled="!ready || loading">
    {{ loading ? 'Processing...' : 'Pay Now' }}
  </button>
</template>

<script setup>
import { ref, onMounted } from 'vue';

const props = defineProps({ amount: Number });
const emit = defineEmits(['success', 'error']);

const loading = ref(false);
const ready = ref(false);

onMounted(() => {
  const script = document.createElement('script');
  script.src = 'https://checkout.razorpay.com/v1/checkout.js';
  script.onload = () => ready.value = true;
  document.head.appendChild(script);
});

const pay = async () => {
  if (!ready.value || loading.value) return;
  loading.value = true;
  try {
    const res = await fetch('/api/razorpay/order', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ amount: props.amount }),
    });
    const data = await res.json();
    if (!data.success) throw new Error(data.error);

    const options = {
      key: data.keyId,
      amount: data.amount,
      currency: data.currency,
      order_id: data.orderId,
      handler: async (response) => {
        const verify = await fetch('/api/razorpay/verify', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(response),
        });
        const result = await verify.json();
        result.success ? emit('success', result) : emit('error', new Error(result.error));
        loading.value = false;
      },
      modal: { ondismiss: () => loading.value = false },
    };
    new window.Razorpay(options).open();
  } catch (e) {
    emit('error', e);
    loading.value = false;
  }
};
</script>
`
	return FrontendIntegration{
		Framework:   "Vue",
		Code:        code,
		FileName:    "src/components/RazorpayButton.vue",
		ScriptTag:   "Import and use <RazorpayButton :amount=\"100\" @success=\"...\" />",
		Description: "Vue 3 component for Razorpay payments",
	}
}

func getAngularFrontend() FrontendIntegration {
	code := `import { Component, Input, Output, EventEmitter, OnInit } from '@angular/core';

declare var Razorpay: any;

@Component({
  selector: 'app-razorpay-button',
  template: ` + "`" + `
    <button (click)="pay()" [disabled]="!ready || loading">
      {{ loading ? 'Processing...' : 'Pay Now' }}
    </button>
  ` + "`" + `,
})
export class RazorpayButtonComponent implements OnInit {
  @Input() amount: number = 0;
  @Output() success = new EventEmitter<any>();
  @Output() error = new EventEmitter<Error>();

  loading = false;
  ready = false;

  ngOnInit() {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => this.ready = true;
    document.head.appendChild(script);
  }

  async pay() {
    if (!this.ready || this.loading) return;
    this.loading = true;
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount: this.amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response: any) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? this.success.emit(result) : this.error.emit(new Error(result.error));
          this.loading = false;
        },
        modal: { ondismiss: () => this.loading = false },
      };
      new Razorpay(options).open();
    } catch (e) {
      this.error.emit(e as Error);
      this.loading = false;
    }
  }
}
`
	return FrontendIntegration{
		Framework:   "Angular",
		Code:        code,
		FileName:    "src/app/components/razorpay-button.component.ts",
		ScriptTag:   "Add to module declarations and use <app-razorpay-button [amount]=\"100\" (success)=\"...\">",
		Description: "Angular component for Razorpay payments",
	}
}

func getSvelteFrontend() FrontendIntegration {
	code := `<script>
  import { onMount } from 'svelte';
  export let amount = 0;

  let loading = false;
  let ready = false;

  onMount(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => ready = true;
    document.head.appendChild(script);
  });

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  async function pay() {
    if (!ready || loading) return;
    loading = true;
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? dispatch('success', result) : dispatch('error', new Error(result.error));
          loading = false;
        },
        modal: { ondismiss: () => loading = false },
      };
      new window.Razorpay(options).open();
    } catch (e) {
      dispatch('error', e);
      loading = false;
    }
  }
</script>

<button on:click={pay} disabled={!ready || loading}>
  {loading ? 'Processing...' : 'Pay Now'}
</button>
`
	return FrontendIntegration{
		Framework:   "Svelte",
		Code:        code,
		FileName:    "src/components/RazorpayButton.svelte",
		ScriptTag:   "Import and use <RazorpayButton amount={100} on:success={...} />",
		Description: "Svelte component for Razorpay payments",
	}
}

func getSolidFrontend() FrontendIntegration {
	code := `import { createSignal, onMount } from 'solid-js';

export function createRazorpayPayment() {
  const [loading, setLoading] = createSignal(false);
  const [ready, setReady] = createSignal(false);

  onMount(() => {
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => setReady(true);
    document.body.appendChild(script);
  });

  const pay = async (amount, onSuccess, onError) => {
    if (!ready() || loading()) return;
    setLoading(true);
    try {
      const res = await fetch('/api/razorpay/order', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount }),
      });
      const data = await res.json();
      if (!data.success) throw new Error(data.error);

      const options = {
        key: data.keyId,
        amount: data.amount,
        currency: data.currency,
        order_id: data.orderId,
        handler: async (response) => {
          const verify = await fetch('/api/razorpay/verify', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(response),
          });
          const result = await verify.json();
          result.success ? onSuccess?.(result) : onError?.(new Error(result.error));
          setLoading(false);
        },
        modal: { ondismiss: () => setLoading(false) },
      };
      new window.Razorpay(options).open();
    } catch (e) {
      onError?.(e);
      setLoading(false);
    }
  };

  return { pay, loading, ready };
}
`
	return FrontendIntegration{
		Framework:   "Solid.js",
		Code:        code,
		FileName:    "src/components/RazorpayPayment.jsx",
		ScriptTag:   "Import createRazorpayPayment and call pay(amount, onSuccess, onError)",
		Description: "Solid.js composable for Razorpay payments",
	}
}
