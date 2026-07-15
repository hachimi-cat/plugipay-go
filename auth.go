package plugipay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// SignInput captures the bytes that go into a Plugipay HMAC signature.
// The string-to-sign is:
//
//	METHOD\nPATH\nTIMESTAMP\nHEX(SHA256(BODY))[\nIDEMPOTENCY_KEY]
//
// where BODY is the raw request body as serialized over the wire (the
// empty string if no body), and IDEMPOTENCY_KEY is appended only when
// non-empty. This must match sdk/node/src/client.ts#sign and
// sdk/python/plugipay/_client.py#_sign byte-for-byte — drift will break
// every authenticated request.
type SignInput struct {
	Method         string
	Path           string
	Body           string
	IdempotencyKey string
	Timestamp      int64 // unix seconds. If 0, time.Now().Unix() is used.
}

// SignResult is the output of Sign — signature plus the timestamp that
// went into it (callers need both to set the request headers).
type SignResult struct {
	Signature string
	Timestamp string
}

// Sign computes the HMAC-SHA256 signature for a Plugipay API request.
// Exposed so callers building lower-level integrations (proxies,
// per-merchant fan-out, custom retry loops) can re-derive the exact
// signature without going through Client.
func Sign(secret string, in SignInput) SignResult {
	ts := in.Timestamp
	if ts == 0 {
		ts = time.Now().Unix()
	}
	tsStr := strconv.FormatInt(ts, 10)

	bodyHashBytes := sha256.Sum256([]byte(in.Body))
	bodyHash := hex.EncodeToString(bodyHashBytes[:])

	stringToSign := in.Method + "\n" + in.Path + "\n" + tsStr + "\n" + bodyHash
	if in.IdempotencyKey != "" {
		stringToSign += "\n" + in.IdempotencyKey
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	sig := hex.EncodeToString(mac.Sum(nil))

	return SignResult{Signature: sig, Timestamp: tsStr}
}

// AuthorizationHeader returns the value for the Authorization header
// given a keyId and signature. Format is:
//
//	Plugipay-HMAC-SHA256 keyId=<id>, scope=*, signature=<hex>
func AuthorizationHeader(keyID, signature string) string {
	return fmt.Sprintf("Plugipay-HMAC-SHA256 keyId=%s, scope=*, signature=%s",
		keyID, signature)
}
