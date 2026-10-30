package region

import (
	"encoding/base64"
	"encoding/json"
	"strings"
)

var (
	jwtURLAlphabet = strings.NewReplacer("-", "+", "_", "/")
	jwtAlgs        = map[string]bool{
		"HS256": true, "HS384": true, "HS512": true,
		"RS256": true, "RS384": true, "RS512": true,
		"ES256": true, "ES384": true,
	}
)

// FromToken reads the country_code claim of an OAuth JWT like the edge does.
func FromToken(token string) string {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return ""
	}
	headerJSON, ok := decodeJWTPart(parts[0])
	if !ok {
		return ""
	}
	claimsJSON, ok := decodeJWTPart(parts[1])
	if !ok {
		return ""
	}
	if _, ok := decodeJWTPart(parts[2]); !ok {
		return ""
	}

	var header struct {
		Alg string `json:"alg"`
	}
	if json.Unmarshal(headerJSON, &header) != nil || !jwtAlgs[header.Alg] {
		return ""
	}
	var claims struct {
		CountryCode string `json:"country_code"`
	}
	if json.Unmarshal(claimsJSON, &claims) != nil {
		return ""
	}
	return strings.ToUpper(claims.CountryCode)
}

// decodeJWTPart mirrors the edge: URL alphabet, padding dropped, no newlines.
func decodeJWTPart(part string) ([]byte, bool) {
	part = jwtURLAlphabet.Replace(part)
	if i := strings.IndexByte(part, '='); i >= 0 {
		part = part[:i]
	}
	if strings.ContainsAny(part, "\r\n") {
		return nil, false
	}
	decoded, err := base64.RawStdEncoding.DecodeString(part)
	return decoded, err == nil
}
