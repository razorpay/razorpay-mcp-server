package region

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeOverride(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "regions.yaml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestLoad(t *testing.T) {
	t.Run("loads built-in IN and SG profiles", func(t *testing.T) {
		profiles, err := Load("")
		require.NoError(t, err)

		assert.Equal(t, "https://api.razorpay.com", profiles["IN"].APIBaseURL)
		assert.Equal(t, []string{"*"}, profiles["IN"].Tools)
		assert.Equal(t, "https://api-sg.razorpay.com", profiles["SG"].APIBaseURL)
		assert.NotEmpty(t, profiles["SG"].Tools)
	})

	t.Run("loads built-in MY and US profiles", func(t *testing.T) {
		profiles, err := Load("")
		require.NoError(t, err)

		assert.Equal(t, "https://api.razorpay.com", profiles["MY"].APIBaseURL)
		assert.NotEmpty(t, profiles["MY"].Tools)
		assert.Equal(t, "https://api-us.razorpay.com", profiles["US"].APIBaseURL)
		assert.NotEmpty(t, profiles["US"].Tools)
	})

	t.Run("India matches the built-in IN profile", func(t *testing.T) {
		profiles, err := Load("")
		require.NoError(t, err)
		assert.Equal(t, profiles["IN"], India())
	})

	t.Run("override replaces only the regions it names", func(t *testing.T) {
		path := writeOverride(t, `
SG:
  country_code: SG
  api_base_url: https://api-sg.stage.example.com
  tools: ["fetch_payment"]
`)
		profiles, err := Load(path)
		require.NoError(t, err)

		assert.Equal(t, "https://api-sg.stage.example.com",
			profiles["SG"].APIBaseURL)
		assert.Equal(t, []string{"fetch_payment"}, profiles["SG"].Tools)
		assert.Equal(t, "https://api.razorpay.com", profiles["IN"].APIBaseURL)
	})

	t.Run("override can add a new region", func(t *testing.T) {
		path := writeOverride(t, `
US:
  country_code: US
  api_base_url: https://api-us.example.com
  tools: ["*"]
`)
		profiles, err := Load(path)
		require.NoError(t, err)

		assert.Equal(t, "https://api-us.example.com", profiles["US"].APIBaseURL)
	})

	t.Run("missing override file errors", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "missing.yaml"))
		assert.ErrorContains(t, err, "region config")
	})

	invalid := []struct {
		name    string
		content string
		errMsg  string
	}{
		{
			name:    "bad yaml",
			content: "SG: [",
			errMsg:  "region config",
		},
		{
			name: "non-https base url",
			content: `
SG:
  country_code: SG
  api_base_url: http://api-sg.razorpay.com
  tools: ["*"]
`,
			errMsg: "api_base_url must use https",
		},
		{
			name: "empty tools",
			content: `
SG:
  country_code: SG
  api_base_url: https://api-sg.razorpay.com
  tools: []
`,
			errMsg: "tools must not be empty",
		},
		{
			name: "country code does not match key",
			content: `
SG:
  country_code: MY
  api_base_url: https://api-sg.razorpay.com
  tools: ["*"]
`,
			errMsg: `country_code "MY" does not match region "SG"`,
		},
		{
			name: "country unknown to i18nify",
			content: `
XX:
  country_code: XX
  api_base_url: https://api-xx.razorpay.com
  tools: ["*"]
`,
			errMsg: `no default currency for country "XX"`,
		},
	}
	for _, tc := range invalid {
		t.Run("rejects override with "+tc.name, func(t *testing.T) {
			_, err := Load(writeOverride(t, tc.content))
			assert.ErrorContains(t, err, tc.errMsg)
		})
	}
}

func TestGet(t *testing.T) {
	profiles, err := Load("")
	require.NoError(t, err)

	t.Run("empty code returns IN", func(t *testing.T) {
		p, err := Get(profiles, "")
		require.NoError(t, err)
		assert.Equal(t, "IN", p.CountryCode)
	})

	t.Run("code is case insensitive", func(t *testing.T) {
		p, err := Get(profiles, "sg")
		require.NoError(t, err)
		assert.Equal(t, "SG", p.CountryCode)
	})

	t.Run("unknown code lists available regions", func(t *testing.T) {
		_, err := Get(profiles, "XX")
		assert.EqualError(t, err,
			`no region profile for "XX"; available: IN, MY, SG, US`)
	})
}

func TestProfileCurrency(t *testing.T) {
	assert.Equal(t, "INR", India().Currency())

	profiles, err := Load("")
	require.NoError(t, err)
	assert.Equal(t, "SGD", profiles["SG"].Currency())
	assert.Equal(t, "MYR", profiles["MY"].Currency())
	assert.Equal(t, "USD", profiles["US"].Currency())
}

func TestProfileLocalize(t *testing.T) {
	profiles, err := Load("")
	require.NoError(t, err)
	sg := profiles["SG"]

	t.Run("fills placeholders from i18nify", func(t *testing.T) {
		got := sg.Localize(
			"smallest unit of {{currency}} ({{minor_unit}} decimal places): " +
				"100 = {{symbol}}1")
		assert.Equal(t,
			"smallest unit of SGD (2 decimal places): 100 = S$1", got)
	})

	t.Run("India uses INR and rupee symbol", func(t *testing.T) {
		assert.Equal(t, "INR ₹", India().Localize("{{currency}} {{symbol}}"))
	})

	t.Run("text without placeholders is unchanged", func(t *testing.T) {
		assert.Equal(t, `{"amount": 100}`, sg.Localize(`{"amount": 100}`))
	})
}

func TestProfileAllowsTool(t *testing.T) {
	t.Run("star allows every tool", func(t *testing.T) {
		p := Profile{Tools: []string{"*"}}
		assert.True(t, p.AllowsAll())
		assert.True(t, p.AllowsTool("anything"))
	})

	t.Run("list allows only listed tools", func(t *testing.T) {
		p := Profile{Tools: []string{"fetch_payment"}}
		assert.False(t, p.AllowsAll())
		assert.True(t, p.AllowsTool("fetch_payment"))
		assert.False(t, p.AllowsTool("fetch_all_payouts"))
	})
}

func TestContext(t *testing.T) {
	t.Run("no profile in context returns India", func(t *testing.T) {
		assert.Equal(t, India(), FromContext(context.Background()))
	})

	t.Run("returns the profile stored in context", func(t *testing.T) {
		sg := Profile{CountryCode: "SG"}
		ctx := WithProfile(context.Background(), sg)
		assert.Equal(t, sg, FromContext(ctx))
	})
}

func TestUnknownCountryPrintsNothingToStdout(t *testing.T) {
	// stdout is the MCP protocol channel in stdio mode.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w

	_, loadErr := Load(writeOverride(t, `
XX:
  country_code: XX
  api_base_url: https://api-xx.razorpay.com
  tools: ["*"]
`))
	_ = Profile{}.Currency()

	os.Stdout = orig
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	assert.Error(t, loadErr)
	assert.Empty(t, string(out))
}
