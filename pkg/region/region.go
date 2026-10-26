// Package region holds the per-region profile the MCP server runs with: the
// Razorpay API host and the tools enabled for the region. Country facts such
// as the default currency come from i18nify.
package region

import (
	"context"
	"embed"
	"fmt"
	"net/url"
	"os"
	"path"
	"slices"
	"sort"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/razorpay/i18nify/packages/i18nify-go/modules/country_metadata"
	"github.com/razorpay/i18nify/packages/i18nify-go/modules/currency"
)

// DefaultRegion is used when no region is configured.
const DefaultRegion = "IN"

const allTools = "*"

//go:embed profiles/*.yaml
var builtin embed.FS

// Profile describes how the server behaves for one region.
type Profile struct {
	CountryCode string   `yaml:"country_code"`
	APIBaseURL  string   `yaml:"api_base_url"`
	Tools       []string `yaml:"tools"`
}

// Currency returns the region's default currency from i18nify.
func (p Profile) Currency() string {
	return country_metadata.GetMetadataInformation(p.CountryCode).DefaultCurrency
}

// Localize fills the {{currency}}, {{minor_unit}} and {{symbol}}
// placeholders in text with the region's currency facts from i18nify.
func (p Profile) Localize(text string) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	code := p.Currency()
	info, _ := currency.GetCurrencyInformation(code)
	return strings.NewReplacer(
		"{{currency}}", code,
		"{{minor_unit}}", info.MinorUnit,
		"{{symbol}}", info.Symbol,
	).Replace(text)
}

// AllowsAll reports whether every registered tool is enabled.
func (p Profile) AllowsAll() bool {
	return len(p.Tools) == 1 && p.Tools[0] == allTools
}

// AllowsTool reports whether the named tool is enabled for the region.
func (p Profile) AllowsTool(name string) bool {
	return p.AllowsAll() || slices.Contains(p.Tools, name)
}

func (p Profile) validate(code string) error {
	if p.CountryCode != code {
		return fmt.Errorf("region %s: country_code %q does not match region %q",
			code, p.CountryCode, code)
	}
	u, err := url.Parse(p.APIBaseURL)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("region %s: api_base_url must use https", code)
	}
	if len(p.Tools) == 0 {
		return fmt.Errorf("region %s: tools must not be empty", code)
	}
	cur := p.Currency()
	if cur == "" {
		return fmt.Errorf("region %s: no default currency for country %q",
			code, p.CountryCode)
	}
	if _, err := currency.GetCurrencyInformation(cur); err != nil {
		return fmt.Errorf("region %s: currency %q not found in i18nify",
			code, cur)
	}
	return nil
}

func readBuiltin(name string) (Profile, error) {
	var p Profile
	data, err := builtin.ReadFile(path.Join("profiles", name))
	if err != nil {
		return p, err
	}
	if err := yaml.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("built-in profile %s: %w", name, err)
	}
	return p, nil
}

// Load returns the built-in profiles keyed by region code. When overridePath
// is set, each region in that file replaces the built-in profile for that
// region.
func Load(overridePath string) (map[string]Profile, error) {
	profiles := map[string]Profile{}

	entries, err := builtin.ReadDir("profiles")
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		p, err := readBuiltin(e.Name())
		if err != nil {
			return nil, err
		}
		profiles[strings.TrimSuffix(e.Name(), ".yaml")] = p
	}

	if overridePath != "" {
		data, err := os.ReadFile(overridePath)
		if err != nil {
			return nil, fmt.Errorf("region config %s: %w", overridePath, err)
		}
		var overrides map[string]Profile
		if err := yaml.Unmarshal(data, &overrides); err != nil {
			return nil, fmt.Errorf("region config %s: %w", overridePath, err)
		}
		for code, p := range overrides {
			profiles[strings.ToUpper(code)] = p
		}
	}

	for code, p := range profiles {
		if err := p.validate(code); err != nil {
			return nil, err
		}
	}
	return profiles, nil
}

// Get returns the profile for code. An empty code means DefaultRegion.
func Get(profiles map[string]Profile, code string) (Profile, error) {
	if code == "" {
		code = DefaultRegion
	}
	code = strings.ToUpper(code)
	if p, ok := profiles[code]; ok {
		return p, nil
	}
	available := make([]string, 0, len(profiles))
	for c := range profiles {
		available = append(available, c)
	}
	sort.Strings(available)
	return Profile{}, fmt.Errorf("no region profile for %q; available: %s",
		code, strings.Join(available, ", "))
}

var india = sync.OnceValue(func() Profile {
	p, err := readBuiltin(DefaultRegion + ".yaml")
	if err != nil {
		panic(err) // embedded at build time; covered by tests
	}
	return p
})

// India returns the built-in IN profile.
func India() Profile {
	return india()
}

type contextKey struct{}

// WithProfile returns a copy of ctx carrying p.
func WithProfile(ctx context.Context, p Profile) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// FromContext returns the profile in ctx, or India() when none is set.
func FromContext(ctx context.Context) Profile {
	if p, ok := ctx.Value(contextKey{}).(Profile); ok {
		return p
	}
	return India()
}
