package plugipay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"testing"
	"time"
)

// signFixture mirrors what the backend sends.
func signFixture(secret string, body []byte, ts int64) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(fmt.Sprintf("%d.", ts) + string(body)))
	return fmt.Sprintf("t=%d,v1=%s", ts, hex.EncodeToString(mac.Sum(nil)))
}

func TestVerifyWebhookSignature_RoundTrip(t *testing.T) {
	secret := "whsec_test_1234567890"
	body := []byte(`{"id":"evt_x","type":"plugipay.invoice.paid.v1","accountId":"acc_1","occurredAt":"2026-01-01T00:00:00Z","data":{"object":{}}}`)
	ts := int64(1_750_000_000)
	sig := signFixture(secret, body, ts)

	fixedClock := func() time.Time { return time.Unix(ts+10, 0) }
	if !VerifyWebhookSignature(body, sig, secret, &VerifyWebhookOptions{Now: fixedClock}) {
		t.Fatal("valid signature rejected")
	}
}

func TestVerifyWebhookSignature_Tampered(t *testing.T) {
	secret := "whsec_x"
	ts := int64(1_750_000_000)
	body := []byte(`{"hello":"world"}`)
	good := signFixture(secret, body, ts)

	tampered := []byte(`{"hello":"mars"}`) // same length, different content
	fixedClock := func() time.Time { return time.Unix(ts, 0) }

	if VerifyWebhookSignature(tampered, good, secret, &VerifyWebhookOptions{Now: fixedClock}) {
		t.Fatal("tampered body verified true")
	}
}

func TestVerifyWebhookSignature_Expired(t *testing.T) {
	secret := "whsec_x"
	ts := int64(1_750_000_000)
	body := []byte(`{}`)
	sig := signFixture(secret, body, ts)

	// 10 min after signing — outside default 5 min tolerance
	fixedClock := func() time.Time { return time.Unix(ts+600, 0) }
	if VerifyWebhookSignature(body, sig, secret, &VerifyWebhookOptions{Now: fixedClock}) {
		t.Fatal("stale signature accepted")
	}
}

func TestVerifyWebhookSignature_MalformedHeader(t *testing.T) {
	if VerifyWebhookSignature([]byte("x"), "not-a-valid-header", "s", nil) {
		t.Fatal("bogus header accepted")
	}
	if VerifyWebhookSignature([]byte("x"), "", "s", nil) {
		t.Fatal("empty header accepted")
	}
	if VerifyWebhookSignature([]byte("x"), "t=abc,v1=deadbeef", "s", nil) {
		t.Fatal("non-numeric timestamp accepted")
	}
	if VerifyWebhookSignature([]byte("x"), "t=123,v1=ZZZZ", "s", nil) {
		t.Fatal("non-hex v1 accepted")
	}
}

func TestVerifyWebhookSignature_EmptySecret(t *testing.T) {
	body := []byte(`{}`)
	ts := int64(1_750_000_000)
	sig := signFixture("whsec_x", body, ts)
	if VerifyWebhookSignature(body, sig, "", nil) {
		t.Fatal("empty secret accepted")
	}
}

// ─── typed VerifyWebhook ─────────────────────────────────────────

func TestVerifyWebhook_TypedParse(t *testing.T) {
	secret := "whsec_typed"
	body := []byte(`{"id":"evt_abc","type":"plugipay.checkout_session.completed.v1","accountId":"acc_1","occurredAt":"2026-01-01T00:00:00Z","data":{"object":{"id":"cs_1"}}}`)
	ts := int64(1_750_000_000)
	sig := signFixture(secret, body, ts)
	fixedClock := func() time.Time { return time.Unix(ts, 0) }

	ev, err := VerifyWebhook(body, sig, secret, &VerifyWebhookOptions{Now: fixedClock})
	if err != nil {
		t.Fatalf("VerifyWebhook: %v", err)
	}
	if ev.ID != "evt_abc" || ev.Type != "plugipay.checkout_session.completed.v1" {
		t.Fatalf("parsed event wrong: %+v", ev)
	}
}

func TestVerifyWebhook_ErrorCodes(t *testing.T) {
	secret := "whsec_x"
	ts := int64(1_750_000_000)
	body := []byte(`{"hello":"world"}`)
	good := signFixture(secret, body, ts)
	fixed := func() time.Time { return time.Unix(ts, 0) }

	// missing header
	if _, err := VerifyWebhook(body, "", secret, nil); err == nil {
		t.Error("expected signature_missing")
	} else {
		var pe *Error
		if !errors.As(err, &pe) || pe.Code != "signature_missing" {
			t.Errorf("missing: got %+v", err)
		}
	}

	// malformed header
	if _, err := VerifyWebhook(body, "not-a-header", secret, &VerifyWebhookOptions{Now: fixed}); err == nil {
		t.Error("expected signature_malformed")
	} else {
		var pe *Error
		if !errors.As(err, &pe) || pe.Code != "signature_malformed" {
			t.Errorf("malformed: got %+v", err)
		}
	}

	// stale
	stale := func() time.Time { return time.Unix(ts+10_000, 0) }
	if _, err := VerifyWebhook(body, good, secret, &VerifyWebhookOptions{Now: stale}); err == nil {
		t.Error("expected signature_stale")
	} else {
		var pe *Error
		if !errors.As(err, &pe) || pe.Code != "signature_stale" {
			t.Errorf("stale: got %+v", err)
		}
	}

	// invalid (wrong secret)
	if _, err := VerifyWebhook(body, good, "wrong_secret", &VerifyWebhookOptions{Now: fixed}); err == nil {
		t.Error("expected signature_invalid")
	} else {
		var pe *Error
		if !errors.As(err, &pe) || pe.Code != "signature_invalid" {
			t.Errorf("invalid: got %+v", err)
		}
	}

	// signed but non-JSON body
	badBody := []byte(`not json`)
	badSig := signFixture(secret, badBody, ts)
	if _, err := VerifyWebhook(badBody, badSig, secret, &VerifyWebhookOptions{Now: fixed}); err == nil {
		t.Error("expected invalid_payload")
	} else {
		var pe *Error
		if !errors.As(err, &pe) || pe.Code != "invalid_payload" {
			t.Errorf("invalid_payload: got %+v", err)
		}
	}
}

func TestVerifyWebhookSignature_CustomTolerance(t *testing.T) {
	secret := "whsec_x"
	body := []byte(`{}`)
	ts := int64(1_750_000_000)
	sig := signFixture(secret, body, ts)

	// 400s after signing — outside default 300s tolerance
	clock := func() time.Time { return time.Unix(ts+400, 0) }
	if VerifyWebhookSignature(body, sig, secret, &VerifyWebhookOptions{Now: clock}) {
		t.Fatal("default tolerance should reject 400s drift")
	}
	// But with a custom 600s tolerance it should accept.
	if !VerifyWebhookSignature(body, sig, secret,
		&VerifyWebhookOptions{Now: clock, ToleranceSeconds: 600}) {
		t.Fatal("custom 600s tolerance should accept 400s drift")
	}
}
