// Package region holds per-region API host and tools; facts from i18nify.
package region

import (
	"context"
	"embed"
	"fmt"
	"maps"
	"net/url"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"

	countrydata "github.com/razorpay/i18nify/i18nify-data/go/country/metadata"
	currencydata "github.com/razorpay/i18nify/i18nify-data/go/currency"
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

// country returns the region's i18nify metadata, or nil if unknown.
func (p Profile) country() *countrydata.CountryMetadata {
	d, err := countrydata.GetCountryMetadataData()
	if err != nil {
		return nil
	}
	return d.GetMetadataInformation()[p.CountryCode]
}

// currencyInfo returns i18nify's facts for an ISO currency code, or nil.
func currencyInfo(code string) *currencydata.CurrencyInfo {
	d, err := currencydata.GetCurrencyData()
	if err != nil {
		return nil
	}
	return d.GetCurrencyInformation()[code]
}

// Currency returns the region's default currency from i18nify, or "".
func (p Profile) Currency() string {
	return p.country().GetDefaultCurrency()
}

// replacers caches one placeholder replacer per country code.
var replacers sync.Map

// Localize fills region placeholders with the country's facts from i18nify.
func (p Profile) Localize(text string) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	if r, ok := replacers.Load(p.CountryCode); ok {
		return r.(*strings.Replacer).Replace(text)
	}
	c := p.country()
	info := currencyInfo(c.GetDefaultCurrency())
	r := strings.NewReplacer(
		"{{currency}}", c.GetDefaultCurrency(),
		"{{minor_unit}}", info.GetMinorUnit(),
		"{{symbol}}", info.GetSymbol(),
		"{{dial_code}}", c.GetDialCode(),
	)
	replacers.Store(p.CountryCode, r)
	return r.Replace(text)
}

// AllowsAll reports whether every registered tool is enabled.
func (p Profile) AllowsAll() bool {
	return len(p.Tools) == 1 && p.Tools[0] == allTools
}

// AllowsTool reports whether the named tool is enabled for the region.
func (p Profile) AllowsTool(name string) bool {
	return p.AllowsAll() || slices.Contains(p.Tools, name)
}

// Validate reports whether the profile can serve a region.
func (p Profile) Validate() error {
	if p.CountryCode == "" {
		return fmt.Errorf("country_code must be set")
	}
	if err := validateBaseURL(p.APIBaseURL); err != nil {
		return err
	}
	if len(p.Tools) == 0 {
		return fmt.Errorf("tools must not be empty")
	}
	cur := p.Currency()
	if cur == "" {
		return fmt.Errorf("no default currency for country %q", p.CountryCode)
	}
	info := currencyInfo(cur)
	if info == nil {
		return fmt.Errorf("currency %q not found in i18nify", cur)
	}
	if info.GetMinorUnit() != "2" {
		return fmt.Errorf("currency %q has %s decimal places; amount "+
			"descriptions in pkg/razorpay/descriptions.go assume 2",
			cur, info.GetMinorUnit())
	}
	return nil
}

func (p Profile) validate(code string) error {
	if p.CountryCode != "" && p.CountryCode != code {
		return fmt.Errorf("region %s: country_code %q does not match region %q",
			code, p.CountryCode, code)
	}
	if err := p.Validate(); err != nil {
		return fmt.Errorf("region %s: %w", code, err)
	}
	return nil
}

// apiDomains are the only hosts the key and secret may be sent to.
var apiDomains = []string{"razorpay.com", "razorpay.in", "curlec.com"}

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid api_base_url: %w", err)
	}
	if u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("api_base_url must use https")
	}
	if u.User != nil || u.Port() != "" || u.Path != "" || u.RawQuery != "" ||
		u.Fragment != "" || u.Opaque != "" {
		return fmt.Errorf("api_base_url must be scheme and host only, "+
			"e.g. https://api.razorpay.com; got %q", raw)
	}
	host := strings.ToLower(u.Hostname())
	for _, d := range apiDomains {
		if host == d || strings.HasSuffix(host, "."+d) {
			return nil
		}
	}
	return fmt.Errorf("api_base_url host must be under %s; got %q",
		strings.Join(apiDomains, ", "), host)
}

func readBuiltin(name string) (Profile, error) {
	var p Profile
	data, err := builtin.ReadFile(path.Join("profiles", name))
	if err != nil {
		return p, fmt.Errorf("built-in profile %s: %w", name, err)
	}
	if err := yaml.Unmarshal(data, &p); err != nil {
		return p, fmt.Errorf("built-in profile %s: %w", name, err)
	}
	return p, nil
}

// Load returns built-in profiles merged with overrides; Get validates them.
func Load(overridePath string) (map[string]Profile, error) {
	profiles := map[string]Profile{}

	entries, err := builtin.ReadDir("profiles")
	if err != nil {
		return nil, fmt.Errorf("built-in profiles: %w", err)
	}
	for _, e := range entries {
		p, err := readBuiltin(e.Name())
		if err != nil {
			return nil, err
		}
		profiles[strings.TrimSuffix(e.Name(), ".yaml")] = p
	}

	if overridePath == "" {
		return profiles, nil
	}
	data, err := os.ReadFile(overridePath)
	if err != nil {
		return nil, fmt.Errorf("region config %s: %w", overridePath, err)
	}
	var overrides map[string]yaml.Node
	if err := yaml.Unmarshal(data, &overrides); err != nil {
		return nil, fmt.Errorf("region config %s: %w", overridePath, err)
	}
	for code, node := range overrides {
		code = strings.ToUpper(code)
		// Fields the override leaves out keep their built-in values.
		p := profiles[code]
		p.Tools = slices.Clone(p.Tools)
		if err := node.Decode(&p); err != nil {
			return nil, fmt.Errorf("region config %s: region %s: %w",
				overridePath, code, err)
		}
		profiles[code] = p
	}
	return profiles, nil
}

// keyCountry is the edge geo-router key regex, so countries match the edge.
var keyCountry = regexp.MustCompile(
	`^rzp_(?:test|live)_([a-zA-Z]{2})_(?:(?:partner|oauth)_)?\w{14}`)

// keyCountryOf returns the country segment of an API key, or "" if none.
func keyCountryOf(keyID string) string {
	if m := keyCountry.FindStringSubmatch(keyID); m != nil {
		return strings.ToUpper(m[1])
	}
	return ""
}

// Resolve picks the explicit region, else the key's country, else IN.
func Resolve(explicit, keyID string) (string, error) {
	explicit = strings.ToUpper(explicit)
	fromKey := keyCountryOf(keyID)

	switch {
	case explicit != "" && fromKey != "" && explicit != fromKey:
		return "", fmt.Errorf("region %s does not match API key region %s",
			explicit, fromKey)
	case explicit != "":
		return explicit, nil
	case fromKey != "":
		return fromKey, nil
	default:
		return DefaultRegion, nil
	}
}

// Get returns the profile for code. An empty code means DefaultRegion.
func Get(profiles map[string]Profile, code string) (Profile, error) {
	if code == "" {
		code = DefaultRegion
	}
	code = strings.ToUpper(code)
	if p, ok := profiles[code]; ok {
		return p, p.validate(code)
	}
	available := slices.Sorted(maps.Keys(profiles))
	return Profile{}, fmt.Errorf("no region profile for %q; available: %s",
		code, strings.Join(available, ", "))
}

// India returns the built-in IN profile; profiles/IN.yaml documents it.
func India() Profile {
	return Profile{
		CountryCode: DefaultRegion,
		APIBaseURL:  "https://api.razorpay.com",
		Tools:       []string{allTools},
	}
}

type contextKey struct{}

// WithProfile returns a copy of ctx carrying p.
func WithProfile(ctx context.Context, p Profile) context.Context {
	return context.WithValue(ctx, contextKey{}, p)
}

// Lookup returns the profile in ctx and whether one was set.
func Lookup(ctx context.Context) (Profile, bool) {
	p, ok := ctx.Value(contextKey{}).(Profile)
	return p, ok
}

// FromContext returns the profile in ctx, or India() when none is set.
func FromContext(ctx context.Context) Profile {
	if p, ok := Lookup(ctx); ok {
		return p
	}
	return India()
}
