package plugipay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// VerifyWebhookOptions tunes the webhook verifier.
type VerifyWebhookOptions struct {
	// ToleranceSeconds rejects signatures with a timestamp older than this
	// many seconds. Zero or negative means "use the default" (300).
	ToleranceSeconds int64
	// Now is the clock used for the freshness check. Nil means time.Now.
	// Useful in tests.
	Now func() time.Time
}

// VerifyWebhookSignature returns true iff the X-Plugipay-Signature
// header over the given raw request body validates against the secret
// within the tolerance window.
//
// Plugipay signs webhooks with HMAC-SHA256 over `${timestamp}.${rawBody}`
// and sends the signature as
//
//	X-Plugipay-Signature: t=<unix>,v1=<hex>
//
// Mount your webhook handler so that you can read the raw body — if a
// framework parses JSON first and re-stringifies, whitespace drift will
// break the signature. Pass the exact bytes that arrived on the wire.
//
// net/http example:
//
//	func plugipayWebhook(w http.ResponseWriter, r *http.Request) {
//	    body, err := io.ReadAll(r.Body)
//	    if err != nil { http.Error(w, "bad body", 400); return }
//	    sig := r.Header.Get("X-Plugipay-Signature")
//	    if !plugipay.VerifyWebhookSignature(body, sig, os.Getenv("PLUGIPAY_WEBHOOK_SECRET"), nil) {
//	        http.Error(w, "bad signature", 400); return
//	    }
//	    // unmarshal body into plugipay.WebhookEvent, dispatch.
//	}
func VerifyWebhookSignature(
	rawBody []byte,
	signatureHeader string,
	secret string,
	options *VerifyWebhookOptions,
) bool {
	if signatureHeader == "" || secret == "" {
		return false
	}

	tolerance := int64(300)
	nowFn := time.Now
	if options != nil {
		if options.ToleranceSeconds > 0 {
			tolerance = options.ToleranceSeconds
		}
		if options.Now != nil {
			nowFn = options.Now
		}
	}

	timestamp, expectedV1, ok := parsePlugipaySignature(signatureHeader)
	if !ok {
		return false
	}
	nowTs := nowFn().Unix()
	if diff := nowTs - timestamp; diff > tolerance || diff < -tolerance {
		return false
	}

	payload := strconv.FormatInt(timestamp, 10) + "." + string(rawBody)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	computed := hex.EncodeToString(mac.Sum(nil))

	expectedBytes, err := hex.DecodeString(expectedV1)
	if err != nil {
		return false
	}
	computedBytes, err := hex.DecodeString(computed)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedBytes, computedBytes)
}

// VerifyWebhook is the typed convenience wrapper: verifies the signature
// AND unmarshals the body into a WebhookEvent. Returns *Error on any
// failure (signature_missing, signature_malformed, signature_stale,
// signature_invalid, invalid_payload).
func VerifyWebhook(
	rawBody []byte,
	signatureHeader string,
	secret string,
	options *VerifyWebhookOptions,
) (*WebhookEvent, error) {
	if signatureHeader == "" {
		return nil, newErr(401, "signature_missing", "X-Plugipay-Signature header is missing")
	}
	if secret == "" {
		return nil, newErr(401, "signature_missing", "webhook secret is empty")
	}

	tolerance := int64(300)
	nowFn := time.Now
	if options != nil {
		if options.ToleranceSeconds > 0 {
			tolerance = options.ToleranceSeconds
		}
		if options.Now != nil {
			nowFn = options.Now
		}
	}

	timestamp, expectedV1, ok := parsePlugipaySignature(signatureHeader)
	if !ok {
		return nil, newErr(401, "signature_malformed", "missing t= or v1= in signature header")
	}
	nowTs := nowFn().Unix()
	age := nowTs - timestamp
	if age < 0 {
		age = -age
	}
	if age > tolerance {
		return nil, newErr(401, "signature_stale",
			"timestamp outside "+strconv.FormatInt(tolerance, 10)+"s tolerance")
	}

	payload := strconv.FormatInt(timestamp, 10) + "." + string(rawBody)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	computed := mac.Sum(nil)

	expectedBytes, err := hex.DecodeString(expectedV1)
	if err != nil {
		return nil, newErr(401, "signature_malformed", "v1 is not hex")
	}
	if len(expectedBytes) != len(computed) || !hmac.Equal(expectedBytes, computed) {
		return nil, newErr(401, "signature_invalid", "signature did not match")
	}

	var ev WebhookEvent
	if err := json.Unmarshal(rawBody, &ev); err != nil {
		return nil, newErr(400, "invalid_payload", "webhook body is not valid JSON: "+err.Error())
	}
	return &ev, nil
}

func parsePlugipaySignature(header string) (timestamp int64, v1 string, ok bool) {
	parts := strings.Split(header, ",")
	var tSeen, v1Seen bool
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		k, v, found := strings.Cut(p, "=")
		if !found {
			continue
		}
		switch k {
		case "t":
			if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
				timestamp = ts
				tSeen = true
			}
		case "v1":
			v1 = v
			v1Seen = true
		}
	}
	return timestamp, v1, tSeen && v1Seen
}
