package plugipay

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"strings"
	"testing"
)

// TestLiveApp drives the hand-written methods fixed in 0.4.0 against a running
// Plugipay API. Skipped unless PLUGIPAY_LIVE_BASE_URL, PLUGIPAY_LIVE_KEY_ID and
// PLUGIPAY_LIVE_SECRET are set. The backend's own suite sets them: it starts the real
// app on a local socket with a test key and runs this test
// (backend/src/__tests__/integration/sdk-hand-written.test.ts). Never point it at
// production: it creates customers, plans and templates, and connects adapters.
func TestLiveApp(t *testing.T) {
	base, keyID, secret := os.Getenv("PLUGIPAY_LIVE_BASE_URL"), os.Getenv("PLUGIPAY_LIVE_KEY_ID"), os.Getenv("PLUGIPAY_LIVE_SECRET")
	if base == "" || keyID == "" || secret == "" {
		t.Skip("set PLUGIPAY_LIVE_BASE_URL / PLUGIPAY_LIVE_KEY_ID / PLUGIPAY_LIVE_SECRET to run against an API")
	}
	c, err := NewClient(ClientOptions{KeyID: keyID, Secret: secret, BaseURL: base})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	tag := func() string {
		var b [4]byte
		_, _ = rand.Read(b[:])
		return hex.EncodeToString(b[:])
	}
	str := func(s string) *string { return &s }
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("Customers.Update", func(t *testing.T) {
		cust, err := c.Customers.Create(ctx, CustomerCreateInput{Email: str("go-" + tag() + "@example.com"), Name: str("Before")})
		must(err)
		up, err := c.Customers.Update(ctx, cust.ID, CustomerUpdateInput{Name: str("After"), Metadata: map[string]string{"source": "go"}})
		must(err)
		if up.Name == nil || *up.Name != "After" {
			t.Fatalf("name not updated: %+v", up)
		}
	})

	t.Run("Plans.Create", func(t *testing.T) {
		flat, err := c.Plans.Create(ctx, PlanCreateInput{Name: "Go flat " + tag(), Interval: "month", Currency: "IDR", Amount: 150000})
		must(err)
		if len(flat.Prices) != 1 || flat.Prices[0].Model != "flat" || flat.Prices[0].UnitAmount == nil || *flat.Prices[0].UnitAmount != 150000 {
			t.Fatalf("flat shorthand: %+v", flat.Prices)
		}
		if flat.PortalFeatures == nil || !flat.PortalFeatures.SelfServeCancel {
			t.Fatalf("default portal features: %+v", flat.PortalFeatures)
		}
		trial := 7
		amount := int64(9900)
		full, err := c.Plans.Create(ctx, PlanCreateInput{
			Name: "Go full " + tag(), Interval: "year", TrialDays: &trial,
			PortalFeatures: &PortalFeatures{SelfServeCancel: true, SelfServePause: true, UpdatePaymentMethod: true},
			Prices:         []PriceInput{{Currency: "USD", Model: "flat", UnitAmount: &amount, TaxMode: str("inclusive")}},
		})
		must(err)
		if full.TrialDays != 7 || full.PortalFeatures == nil || !full.PortalFeatures.SelfServePause || full.Prices[0].Currency != "USD" {
			t.Fatalf("full plan: %+v", full)
		}
	})

	t.Run("Templates", func(t *testing.T) {
		name := "Go receipt " + tag()
		tpl, err := c.Templates.Create(ctx, TemplateCreateInput{
			Kind: "receipt", Name: name, Config: map[string]any{"footerText": "Thanks", "accentColor": "#0F766E"},
		})
		must(err)
		if tpl.Config["footerText"] != "Thanks" {
			t.Fatalf("config: %+v", tpl.Config)
		}
		kind := TemplateKind("receipt")
		list, err := c.Templates.List(ctx, TemplateListParams{Kind: &kind})
		must(err)
		found := false
		for _, x := range list {
			found = found || x.ID == tpl.ID
		}
		if !found {
			t.Fatal("created template not listed")
		}
		got, err := c.Templates.Get(ctx, tpl.ID)
		must(err)
		if got.Name != name {
			t.Fatalf("get: %+v", got)
		}
		up, err := c.Templates.Update(ctx, tpl.ID, TemplateUpdateInput{Name: str(name + " v2"), Config: map[string]any{"footerText": "Terima kasih"}})
		must(err)
		if up.Name != name+" v2" || up.Config["footerText"] != "Terima kasih" {
			t.Fatalf("update: %+v", up)
		}
		def, err := c.Templates.MakeDefault(ctx, tpl.ID)
		must(err)
		if !def.IsDefault {
			t.Fatal("not default")
		}
		cp, err := c.Templates.Duplicate(ctx, tpl.ID, str(name+" copy"))
		must(err)
		if cp.Name != name+" copy" || cp.ID == tpl.ID {
			t.Fatalf("duplicate: %+v", cp)
		}
		prev, err := c.Templates.Preview(ctx, TemplatePreviewInput{Kind: "invoice", Config: map[string]any{"termsText": "Net 14"}})
		must(err)
		if !strings.HasPrefix(strings.ToLower(strings.TrimSpace(prev.HTML)), "<!doctype html") || !strings.Contains(prev.HTML, "Net 14") {
			t.Fatalf("preview: %.120q", prev.HTML)
		}
		must(c.Templates.Delete(ctx, cp.ID))
		_, err = c.Templates.Get(ctx, cp.ID)
		var pe *Error
		if !errors.As(err, &pe) || pe.Status != 404 {
			t.Fatalf("deleted template still there: %v", err)
		}
	})

	t.Run("Adapters", func(t *testing.T) {
		m, err := c.Adapters.UpdateManual(ctx, ManualAdapterInput{
			BankAccounts: []ManualBankAccount{{BankName: "BCA", AccountNumber: "1234567890", AccountHolder: "PT Contoh"}},
			Instructions: str("Transfer, then upload the receipt"),
		})
		must(err)
		if m.Kind != "manual" || m.Status != "active" {
			t.Fatalf("manual: %+v", m)
		}
		x, err := c.Adapters.UpdateXendit(ctx, XenditAdapterInput{SecretKey: "xnd_sk_test_go_live_0001"})
		must(err)
		if x.SecretKeyLast4 == nil || *x.SecretKeyLast4 != "0001" {
			t.Fatalf("xendit: %+v", x)
		}
		_, err = c.Adapters.UpdatePaypal(ctx, PaypalAdapterInput{ClientID: "go-client", Secret: "go-secret", Mode: str("sandbox")})
		must(err)
		_, err = c.Adapters.UpdateMidtrans(ctx, MidtransAdapterInput{ServerKey: "SB-Mid-server-go", ClientKey: "SB-Mid-client-go", MerchantID: "G000"})
		must(err)
		list, err := c.Adapters.List(ctx)
		must(err)
		kinds := map[string]bool{}
		for _, a := range list {
			kinds[a.Kind] = true
		}
		for _, k := range []string{"manual", "xendit", "paypal", "midtrans"} {
			if !kinds[k] {
				t.Fatalf("adapters.List missing %s: %+v", k, list)
			}
		}
	})

	t.Run("CheckoutSettings.Update", func(t *testing.T) {
		s, err := c.CheckoutSettings.Update(ctx, CheckoutSettingsUpdateInput{BrandName: str("Toko Go"), BrandAccentColor: str("#654321")})
		must(err)
		if s.BrandName == nil || *s.BrandName != "Toko Go" || s.BrandAccentColor == nil || *s.BrandAccentColor != "#654321" {
			t.Fatalf("settings: %+v", s)
		}
	})

	t.Run("Invoices.Create+Finalize", func(t *testing.T) {
		cust, err := c.Customers.Create(ctx, CustomerCreateInput{Email: str("inv-" + tag() + "@example.com")})
		must(err)
		inv, err := c.Invoices.Create(ctx, InvoiceCreateInput{
			CustomerID: cust.ID, Currency: "IDR",
			Lines: []InvoiceCreateLine{{Description: "Consulting", Quantity: 1, UnitAmount: 250000}},
		})
		must(err)
		fin, err := c.Invoices.Finalize(ctx, inv.ID)
		must(err)
		if fin.Status != "open" {
			t.Fatalf("finalize: %+v", fin)
		}
	})

	t.Run("ManagedOnboarding", func(t *testing.T) {
		if os.Getenv("PLUGIPAY_LIVE_MANAGED") != "1" {
			t.Skip("managed payments are switched off on this API")
		}
		st, err := c.Adapters.StartManagedOnboarding(ctx, ManagedOnboardingStartInput{Email: "payouts-" + tag() + "@example.com"})
		must(err)
		v, err := c.Adapters.SimulateManagedOnboarding(ctx, ManagedOnboardingSimulateInput{})
		must(err)
		if v.KybStatus != "live" || !v.PayoutsReady {
			t.Fatalf("simulate: %+v", v)
		}
		p, err := c.Onboarding.ProvisionManaged(ctx)
		must(err)
		if p.SubAccountID != st.SubAccountID {
			t.Fatalf("provision: %+v vs %+v", p, st)
		}
	})
}
