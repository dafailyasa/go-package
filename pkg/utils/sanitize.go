package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

// SensitiveKey defines a sensitive field key and its masking type.
type SensitiveKey struct {
	Pattern     *regexp.Regexp
	MaskingType MaskingType
}

// defaultSensitiveKeys is the built-in list of sensitive field patterns.
// Rules:
//   - Do NOT add container keys (e.g. "credentials") here — they are arrays/objects,
//     not strings, so Mask() won't fire. Let the recursion walk into them naturally.
//   - Only add leaf-level string keys whose values should be masked.
var defaultSensitiveKeys = []SensitiveKey{
	// --- auth / secrets ---
	{Pattern: regexp.MustCompile(`(?i)^password$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^pass$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^secret$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^access_token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^refresh_token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^id_token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^api_key$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^apikey$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^authorization$`), MaskingType: Any},
	// credential object leaf — "value" inside credentials[].value
	{Pattern: regexp.MustCompile(`(?i)^value$`), MaskingType: Any},
	// --- PII ---
	{Pattern: regexp.MustCompile(`(?i)^email$`), MaskingType: Middle},
	{Pattern: regexp.MustCompile(`(?i)^phone$`), MaskingType: Right},
	{Pattern: regexp.MustCompile(`(?i)^phone_number$`), MaskingType: Right},
	{Pattern: regexp.MustCompile(`(?i)^card_number$`), MaskingType: Left},
	{Pattern: regexp.MustCompile(`(?i)^credit_card$`), MaskingType: Left},
	{Pattern: regexp.MustCompile(`(?i)^cvv$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^ssn$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^national_id$`), MaskingType: Any},
}

// defaultSensitiveHeaders is the list of HTTP header names that should be masked.
var defaultSensitiveHeaders = []SensitiveKey{
	{Pattern: regexp.MustCompile(`(?i)^authorization$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^proxy-authorization$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^x-api-key$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^x-auth-token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^x-access-token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^x-refresh-token$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^x-secret$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^cookie$`), MaskingType: Any},
	{Pattern: regexp.MustCompile(`(?i)^set-cookie$`), MaskingType: Any},
}

// SanitizeOption allows customizing the sanitizer behavior.
type SanitizeOption func(*sanitizeConfig)

type sanitizeConfig struct {
	extraKeys     []SensitiveKey
	replaceKeys   []SensitiveKey
	useCustomOnly bool
}

// WithExtraKeys adds additional sensitive keys on top of defaults.
func WithExtraKeys(keys []SensitiveKey) SanitizeOption {
	return func(c *sanitizeConfig) {
		c.extraKeys = append(c.extraKeys, keys...)
	}
}

// WithCustomKeysOnly replaces the default sensitive key list entirely.
func WithCustomKeysOnly(keys []SensitiveKey) SanitizeOption {
	return func(c *sanitizeConfig) {
		c.replaceKeys = keys
		c.useCustomOnly = true
	}
}

func resolveKeys(defaults []SensitiveKey, cfg *sanitizeConfig) []SensitiveKey {
	active := defaults
	if cfg.useCustomOnly {
		active = cfg.replaceKeys
	}
	return append(active, cfg.extraKeys...)
}

// SanitizeBody masks sensitive fields in a JSON body ([]byte).
// Recursively traverses nested objects and arrays.
// Returns sanitized JSON bytes. Returns original bytes if input is not valid JSON.
func SanitizeBody(bodyBytes []byte, opts ...SanitizeOption) []byte {
	if len(bodyBytes) == 0 {
		return bodyBytes
	}
	cfg := &sanitizeConfig{}
	for _, o := range opts {
		o(cfg)
	}
	var parsed any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return bodyBytes
	}
	out, err := json.Marshal(maskValue(parsed, resolveKeys(defaultSensitiveKeys, cfg)))
	if err != nil {
		return bodyBytes
	}
	return out
}

// SanitizeBodyParsed sanitizes sensitive fields and returns the result as a parsed
// any (map[string]any, []any, etc.) instead of JSON bytes.
// Preferred for structured loggers like logrus — avoids escaped \n in log output.
func SanitizeBodyParsed(bodyBytes []byte, opts ...SanitizeOption) any {
	if len(bodyBytes) == 0 {
		return nil
	}
	cfg := &sanitizeConfig{}
	for _, o := range opts {
		o(cfg)
	}
	var parsed any
	if err := json.Unmarshal(bodyBytes, &parsed); err != nil {
		return string(bodyBytes)
	}
	return maskValue(parsed, resolveKeys(defaultSensitiveKeys, cfg))
}

// SanitizeBodyIndent is the same as SanitizeBody but returns pretty-printed JSON.
func SanitizeBodyIndent(bodyBytes []byte, opts ...SanitizeOption) []byte {
	if len(bodyBytes) == 0 {
		return bodyBytes
	}
	sanitized := SanitizeBody(bodyBytes, opts...)
	var parsed any
	if err := json.Unmarshal(sanitized, &parsed); err != nil {
		return sanitized
	}
	out, err := json.MarshalIndent(parsed, "", "  ")
	if err != nil {
		return sanitized
	}
	return out
}

// SanitizeHeaders masks sensitive values in an HTTP header map (map[string]string).
// Returns a new map — the original is never mutated.
func SanitizeHeaders(headers map[string]string, opts ...SanitizeOption) map[string]string {
	if len(headers) == 0 {
		return headers
	}
	cfg := &sanitizeConfig{}
	for _, o := range opts {
		o(cfg)
	}
	activeKeys := resolveKeys(defaultSensitiveHeaders, cfg)
	result := make(map[string]string, len(headers))
	for k, v := range headers {
		if mt, ok := matchSensitiveKey(k, activeKeys); ok {
			result[k] = Mask(mt, v)
		} else {
			result[k] = v
		}
	}
	return result
}

// maskValue recursively walks any JSON-decoded value and masks sensitive keys.
func maskValue(v any, keys []SensitiveKey) any {
	switch val := v.(type) {
	case map[string]any:
		return maskObject(val, keys)
	case []any:
		return maskArray(val, keys)
	default:
		return v
	}
}

func maskObject(obj map[string]any, keys []SensitiveKey) map[string]any {
	result := make(map[string]any, len(obj))
	for k, v := range obj {
		if mt, ok := matchSensitiveKey(k, keys); ok {
			switch typed := v.(type) {
			case string:
				result[k] = Mask(mt, typed)
			case map[string]any:
				result[k] = maskObject(typed, keys)
			case []any:
				result[k] = maskArray(typed, keys)
			default:
				result[k] = v
			}
		} else {
			result[k] = maskValue(v, keys)
		}
	}
	return result
}

func maskArray(arr []any, keys []SensitiveKey) []any {
	result := make([]any, len(arr))
	for i, item := range arr {
		result[i] = maskValue(item, keys)
	}
	return result
}

// matchSensitiveKey checks whether a key matches any sensitive pattern.
func matchSensitiveKey(key string, keys []SensitiveKey) (MaskingType, bool) {
	normalized := strings.TrimSpace(key)
	for _, sk := range keys {
		if sk.Pattern.MatchString(normalized) {
			return sk.MaskingType, true
		}
	}
	return "", false
}
