package privacyfilter

import (
	"bytes"
	"encoding/json"
	"math"
	"regexp"
	"strings"
	"unicode"
)

const maxRedactDepth = 32

const (
	TypeIPAddress   = "ip_address"
	TypeEmail       = "email"
	TypePhone       = "phone"
	TypeIDCard      = "id_card"
	TypeBankCard    = "bank_card"
	TypeAPIKey      = "api_key"
	TypeToken       = "token"
	TypePrivateKey  = "private_key"
	TypeRandomValue = "random_string"
)

const (
	replacementIP          = "[FILTERED:IP]"
	replacementEmail       = "[FILTERED:EMAIL]"
	replacementPhone       = "[FILTERED:PHONE]"
	replacementIDCard      = "[FILTERED:ID_CARD]"
	replacementBankCard    = "[FILTERED:BANK_CARD]"
	replacementAPIKey      = "[FILTERED:API_KEY]"
	replacementToken       = "[FILTERED:TOKEN]"
	replacementPrivateKey  = "[FILTERED:PRIVATE_KEY]"
	replacementRandomValue = "[FILTERED:RANDOM]"
)

var defaultTypes = []string{
	TypeIPAddress,
	TypeEmail,
	TypePhone,
	TypeIDCard,
	TypeBankCard,
	TypeAPIKey,
	TypeToken,
	TypePrivateKey,
	TypeRandomValue,
}

var validTypes = map[string]struct{}{
	TypeIPAddress:   {},
	TypeEmail:       {},
	TypePhone:       {},
	TypeIDCard:      {},
	TypeBankCard:    {},
	TypeAPIKey:      {},
	TypeToken:       {},
	TypePrivateKey:  {},
	TypeRandomValue: {},
}

var (
	privateKeyPattern = regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?-----END [A-Z0-9 ]*PRIVATE KEY-----`)
	apiKeyPattern     = regexp.MustCompile(`\b(?:sk-[A-Za-z0-9_-]{20,}|xai-[A-Za-z0-9_-]{20,}|AIza[0-9A-Za-z_-]{20,}|GOCSPX-[0-9A-Za-z_-]{20,})\b`)
	jwtPattern        = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\b`)
	tokenKVPattern    = regexp.MustCompile(`(?i)\b(access_token|refresh_token|id_token|token|authorization)\b(\s*[:=]\s*)(["']?)[^"',;\s}]+`)
	emailPattern      = regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`)
	cnPhonePattern    = regexp.MustCompile(`\b1[3-9]\d{9}\b`)
	usPhonePattern    = regexp.MustCompile(`(?i)(?:\+?1[-.\s]?)?\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4}`)
	idCardPattern     = regexp.MustCompile(`(?i)\b[1-9]\d{5}(?:18|19|20)\d{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12]\d|3[01])\d{3}[\dX]\b`)
	bankCardPattern   = regexp.MustCompile(`\b(?:\d[ -]?){13,19}\b`)
	ipv4Pattern       = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4]\d|1?\d?\d)\.){3}(?:25[0-5]|2[0-4]\d|1?\d?\d)\b`)
	ipv6Pattern       = regexp.MustCompile(`\b(?:[0-9a-fA-F]{1,4}:){2,7}[0-9a-fA-F]{1,4}\b`)
	randomPattern     = regexp.MustCompile(`\b[A-Za-z0-9_-]{32,}\b`)
)

type Config struct {
	Enabled bool     `json:"enabled"`
	Types   []string `json:"types"`
}

func DefaultConfig() Config {
	return Config{
		Enabled: false,
		Types:   copyDefaultTypes(),
	}
}

func NormalizeConfig(config Config) Config {
	config.Types = NormalizeTypes(config.Types)
	return config
}

func NormalizeTypes(types []string) []string {
	if types == nil {
		return copyDefaultTypes()
	}
	seen := make(map[string]struct{}, len(types))
	normalized := make([]string, 0, len(types))
	for _, item := range types {
		item = strings.ToLower(strings.TrimSpace(item))
		if _, ok := validTypes[item]; !ok {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized
}

func ParseConfig(raw string) Config {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultConfig()
	}

	var config Config
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return DefaultConfig()
	}
	return NormalizeConfig(config)
}

func RedactJSONBody(body []byte, config Config) ([]byte, bool, error) {
	config = NormalizeConfig(config)
	if !config.Enabled || len(config.Types) == 0 || len(bytes.TrimSpace(body)) == 0 {
		return body, false, nil
	}
	if !json.Valid(body) {
		return body, false, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return body, false, nil
	}

	selected := selectedTypes(config.Types)
	redacted, changed := redactValue(value, "", selected, 0)
	if !changed {
		return body, false, nil
	}

	out, err := json.Marshal(redacted)
	if err != nil {
		return body, false, err
	}
	return out, true, nil
}

func RedactText(input string, config Config) (string, bool) {
	config = NormalizeConfig(config)
	if !config.Enabled || len(config.Types) == 0 || input == "" {
		return input, false
	}
	out := redactTextWithTypes(input, selectedTypes(config.Types))
	return out, out != input
}

func redactValue(value any, key string, types map[string]struct{}, depth int) (any, bool) {
	if depth > maxRedactDepth {
		return value, false
	}

	switch v := value.(type) {
	case map[string]any:
		changed := false
		out := make(map[string]any, len(v))
		for k, item := range v {
			redacted, itemChanged := redactValue(item, k, types, depth+1)
			if itemChanged {
				changed = true
			}
			out[k] = redacted
		}
		return out, changed
	case []any:
		changed := false
		out := make([]any, len(v))
		for i, item := range v {
			redacted, itemChanged := redactValue(item, key, types, depth+1)
			if itemChanged {
				changed = true
			}
			out[i] = redacted
		}
		return out, changed
	case string:
		if replacement, ok := replacementForSensitiveKey(key, types); ok && strings.TrimSpace(v) != "" {
			return replacement, true
		}
		out := redactTextWithTypes(v, types)
		return out, out != v
	default:
		return value, false
	}
}

func redactTextWithTypes(input string, types map[string]struct{}) string {
	out := input
	if isTypeSelected(types, TypePrivateKey) {
		out = privateKeyPattern.ReplaceAllString(out, replacementPrivateKey)
	}
	if isTypeSelected(types, TypeAPIKey) {
		out = apiKeyPattern.ReplaceAllString(out, replacementAPIKey)
	}
	if isTypeSelected(types, TypeToken) {
		out = tokenKVPattern.ReplaceAllString(out, `$1$2$3`+replacementToken)
		out = jwtPattern.ReplaceAllString(out, replacementToken)
	}
	if isTypeSelected(types, TypeEmail) {
		out = emailPattern.ReplaceAllString(out, replacementEmail)
	}
	if isTypeSelected(types, TypeIDCard) {
		out = idCardPattern.ReplaceAllString(out, replacementIDCard)
	}
	if isTypeSelected(types, TypeBankCard) {
		out = bankCardPattern.ReplaceAllStringFunc(out, func(match string) string {
			digits := onlyDigits(match)
			if len(digits) < 13 || len(digits) > 19 || !isLuhnValid(digits) {
				return match
			}
			return replacementBankCard
		})
	}
	if isTypeSelected(types, TypePhone) {
		out = cnPhonePattern.ReplaceAllString(out, replacementPhone)
		out = usPhonePattern.ReplaceAllString(out, replacementPhone)
	}
	if isTypeSelected(types, TypeIPAddress) {
		out = ipv4Pattern.ReplaceAllString(out, replacementIP)
		out = ipv6Pattern.ReplaceAllString(out, replacementIP)
	}
	if isTypeSelected(types, TypeRandomValue) {
		out = randomPattern.ReplaceAllStringFunc(out, func(match string) string {
			if isLikelyRandomValue(match) {
				return replacementRandomValue
			}
			return match
		})
	}
	return out
}

func replacementForSensitiveKey(key string, types map[string]struct{}) (string, bool) {
	key = strings.ToLower(strings.TrimSpace(key))
	compact := strings.NewReplacer("-", "", "_", "", " ", "").Replace(key)
	switch {
	case isTypeSelected(types, TypePrivateKey) && strings.Contains(compact, "privatekey"):
		return replacementPrivateKey, true
	case isTypeSelected(types, TypeAPIKey) && (strings.Contains(compact, "apikey") || strings.Contains(compact, "secretkey") || strings.Contains(compact, "clientsecret")):
		return replacementAPIKey, true
	case isTypeSelected(types, TypeToken) && (strings.Contains(compact, "token") || compact == "authorization"):
		return replacementToken, true
	default:
		return "", false
	}
}

func selectedTypes(types []string) map[string]struct{} {
	selected := make(map[string]struct{}, len(types))
	for _, item := range types {
		if _, ok := validTypes[item]; ok {
			selected[item] = struct{}{}
		}
	}
	return selected
}

func isTypeSelected(types map[string]struct{}, item string) bool {
	_, ok := types[item]
	return ok
}

func copyDefaultTypes() []string {
	out := make([]string, len(defaultTypes))
	copy(out, defaultTypes)
	return out
}

func onlyDigits(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))
	for _, r := range input {
		if r >= '0' && r <= '9' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func isLuhnValid(digits string) bool {
	sum := 0
	alternate := false
	for i := len(digits) - 1; i >= 0; i-- {
		n := int(digits[i] - '0')
		if alternate {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
		alternate = !alternate
	}
	return sum > 0 && sum%10 == 0
}

func isLikelyRandomValue(input string) bool {
	if len(input) < 32 || strings.HasPrefix(input, "FILTERED") {
		return false
	}

	hasLetter := false
	hasDigit := false
	hasSymbol := false
	for _, r := range input {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		case r == '_' || r == '-':
			hasSymbol = true
		default:
			return false
		}
	}
	if !hasLetter || !hasDigit {
		return false
	}
	if !hasSymbol && shannonEntropy(input) < 3.5 {
		return false
	}
	return shannonEntropy(input) >= 3.2
}

func shannonEntropy(input string) float64 {
	if input == "" {
		return 0
	}
	counts := make(map[rune]int)
	for _, r := range input {
		counts[r]++
	}
	length := float64(len([]rune(input)))
	var entropy float64
	for _, count := range counts {
		p := float64(count) / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}
