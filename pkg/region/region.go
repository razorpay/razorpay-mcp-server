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

// country uses GetCountriesByCodes: GetMetadataInformation prints to stdout.
func (p Profile) country() (country_metadata.CountryInfo, bool) {
	if p.CountryCode == "" {
		return country_metadata.CountryInfo{}, false
	}
	countries, err := country_metadata.GetCountriesByCodes(
		[]string{p.CountryCode})
	if err != nil || len(countries) == 0 {
		return country_metadata.CountryInfo{}, false
	}
	return countries[0], true
}

// Currency returns the region's default currency from i18nify, or "".
func (p Profile) Currency() string {
	c, _ := p.country()
	return c.DefaultCurrency
}

// Localize fills region placeholders with the country's facts from i18nify.
func (p Profile) Localize(text string) string {
	if !strings.Contains(text, "{{") {
		return text
	}
	c, _ := p.country()
	info, _ := currency.GetCurrencyInformation(c.DefaultCurrency)
	return strings.NewReplacer(
		"{{currency}}", c.DefaultCurrency,
		"{{minor_unit}}", info.MinorUnit,
		"{{symbol}}", info.Symbol,
		"{{dial_code}}", c.DialCode,
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

// Load returns built-in profiles, with regions in overridePath replacing them.
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

// keyCountry is the edge geo-router key regex, so countries match the edge.
var keyCountry = regexp.MustCompile(
	`^rzp_(?:test|live)_([a-zA-Z]{2})_(?:(?:partner|oauth)_)?\w{14}`)

// Resolve picks the explicit region, else the key's country, else IN.
func Resolve(explicit, keyID string) (string, error) {
	explicit = strings.ToUpper(explicit)
	fromKey := ""
	if m := keyCountry.FindStringSubmatch(keyID); m != nil {
		fromKey = strings.ToUpper(m[1])
	}

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
		return p, nil
	}
	available := slices.Sorted(maps.Keys(profiles))
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
