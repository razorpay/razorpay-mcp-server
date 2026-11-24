package integrations

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDjangoGeneratedRuntime(t *testing.T) {
	python := os.Getenv("RZP_DJANGO_PYTHON")
	if python == "" {
		t.Skip("set RZP_DJANGO_PYTHON to a Python with Django and razorpay")
	}
	output := getDjangoIntegration(getVanillaFrontend())
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(t.TempDir(), "generated.json")
	if err := os.WriteFile(generated, data, 0600); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join("testdata", "django_checkout_regression.py")
	cmd := exec.Command(python, script, "--generated", generated)
	result, err := cmd.CombinedOutput()
	t.Log(string(result))
	if err != nil {
		t.Fatalf("generated Django regression: %v", err)
	}
}

func TestDjangoCheckoutCSRFHeaders(t *testing.T) {
	for _, framework := range []string{
		"vanilla", "react", "vue", "angular", "svelte",
	} {
		frontend := getFrontendIntegration(framework)
		output := getDjangoIntegration(frontend)
		found := false
		for _, file := range output.Files {
			if file.Path == frontend.FileName {
				found = true
				if strings.Count(file.Code, "X-CSRFToken") != 2 {
					t.Errorf(
						"%s needs CSRF headers on both requests",
						framework,
					)
				}
			}
		}
		if !found {
			t.Errorf("%s frontend output missing", framework)
		}
	}
}

func TestDjangoConfigurationEdits(t *testing.T) {
	output := getDjangoIntegration(getVanillaFrontend())
	edits := map[string]string{}
	views := ""
	for _, file := range output.Files {
		for _, edit := range file.Edits {
			edits[file.Path] += edit.Add + "\n"
		}
		if file.Path == "razorpay_payments/views.py" {
			views = file.Code
		}
	}
	for _, setting := range []string{
		"import os",
		"RAZORPAY_KEY_ID = os.environ.get('RAZORPAY_KEY_ID')",
		"RAZORPAY_KEY_SECRET = os.environ.get('RAZORPAY_KEY_SECRET')",
	} {
		if !strings.Contains(edits["settings.py"], setting) {
			t.Errorf("generated settings edits omit %s", setting)
		}
	}
	for _, hook := range []string{
		"def get_order_amount_paise(request):",
		"def get_order_snapshot(request):",
	} {
		if !strings.Contains(views, hook) {
			t.Errorf(
				"generated Django views omit required server-side hook %s",
				hook,
			)
		}
	}
	if strings.Count(views, "@csrf_protect") != 2 {
		t.Error(
			"generated Django views must protect both POST endpoints directly",
		)
	}
	for _, requirement := range []string{
		"JsonResponse from django.http",
		"transaction.atomic()",
		"get_token(request)",
		"CsrfViewMiddleware",
		"ensure_csrf_cookie",
	} {
		if !strings.Contains(output.AIInstructions, requirement) {
			t.Errorf("generated checkout instructions omit %s", requirement)
		}
	}
	if !strings.Contains(edits["requirements.txt"], "razorpay>=1.2.0") {
		t.Error("generated requirements omit the Python SDK")
	}
}
