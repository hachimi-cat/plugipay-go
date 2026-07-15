# plugipay (Go)

Official Go SDK for [Plugipay](https://plugipay.com) — HMAC-signed REST
client with full resource coverage (customers, plans, checkout sessions,
invoices, subscriptions, refunds, adapters, payouts, ledger, reports,
templates, webhooks, account, workspaces, admin).

Mirrors the public surface of [`@forjio/plugipay-node`](../node) and the
Python [`plugipay`](../python) package. Stdlib-only — no external deps.

## Install

```bash
go get github.com/hachimi-cat/plugipay-go
```

## Quickstart

Env vars (or pass them to `NewClient`):

```bash
PLUGIPAY_KEY_ID=ak_live_...
PLUGIPAY_SECRET=sk_live_...
PLUGIPAY_BASE_URL=https://plugipay.com         # optional; default
PLUGIPAY_ON_BEHALF_OF=acc_...                  # optional; platform keys only
```

### Create a customer + checkout session

```go
package main

import (
    "context"
    "fmt"
    "log"

    plugipay "github.com/hachimi-cat/plugipay-go"
)

func ptr[T any](v T) *T { return &v }

func main() {
    c, err := plugipay.NewClient(plugipay.ClientOptions{})
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    cust, err := c.Customers.Create(ctx, plugipay.CustomerCreateInput{
        Email: ptr("buyer@example.com"),
        Name:  ptr("Buyer Adi"),
    })
    if err != nil {
        log.Fatal(err)
    }

    sess, err := c.CheckoutSessions.Create(ctx, plugipay.CheckoutSessionCreateInput{
        Amount:     50000,
        Currency:   plugipay.CurrencyIDR,
        Methods:    []plugipay.CheckoutMethod{plugipay.CheckoutMethodQRIS, plugipay.CheckoutMethodVA},
        SuccessURL: "https://yourapp.com/done",
        CancelURL:  "https://yourapp.com/cancel",
        CustomerID: &cust.ID,
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("hosted url:", sess.HostedURL)
}
```

### Verify webhooks

Mount your handler so you can read the raw body — if a framework parses
JSON first and re-stringifies it, whitespace drift will break the
signature.

```go
http.HandleFunc("/webhooks/plugipay", func(w http.ResponseWriter, r *http.Request) {
    body, err := io.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "bad body", 400)
        return
    }
    ev, err := plugipay.VerifyWebhook(
        body,
        r.Header.Get("X-Plugipay-Signature"),
        os.Getenv("PLUGIPAY_WEBHOOK_SECRET"),
        nil,
    )
    if err != nil {
        http.Error(w, "bad signature", 400)
        return
    }
    switch ev.Type {
    case "plugipay.invoice.paid.v1":
        var inv plugipay.Invoice
        _ = json.Unmarshal(ev.Data.Object, &inv)
        // ...
    }
    w.WriteHeader(204)
})
```

`VerifyWebhookSignature(body, header, secret, nil)` returns a bool if
you just want a boolean check.

### Platform-admin keys across many merchants

```go
master, _ := plugipay.NewClient(plugipay.ClientOptions{
    KeyID: "ak_platform_...", Secret: "sk_platform_...",
})

for _, accountID := range merchants {
    scoped := master.ForMerchant(accountID)
    invs, err := scoped.Invoices.List(ctx, plugipay.InvoiceListParams{})
    // ...
}
```

`ForMerchant` returns a shallow clone with `X-Plugipay-On-Behalf-Of` set
on every request. The underlying `*http.Client` is shared.

## What's in the box

| Symbol | Purpose |
|---|---|
| `NewClient(ClientOptions{})` | Construct a client. KeyID + Secret default from env. |
| `c.Customers` / `Plans` / `CheckoutSessions` / `Invoices` / `Subscriptions` | Core commerce primitives. |
| `c.Refunds` / `Payouts` / `Ledger` / `Reports` | Money movement + accounting. |
| `c.Adapters` / `ApiKeys` / `WebhookEndpoints` / `Templates` / `Uploads` | Configuration. |
| `c.Account` / `Workspaces` / `Onboarding` / `Billing` / `CheckoutSettings` | Merchant self-service. |
| `c.AdminPortal` / `Admin` | Plugipay-internal + platform-admin. |
| `VerifyWebhook(body, hdr, secret, opts)` | Verify + parse webhook payload. |
| `VerifyWebhookSignature(...)` | Bool-only variant. |
| `Sign(secret, SignInput{...})` | Low-level HMAC signing helper. |
| `*Error` | Single error type — branch on `.Code` / `.Status`. |

## Errors

Every operation returns `*plugipay.Error` on failure (wrapped via the
standard `error` interface). Extract with `errors.As`:

```go
var pe *plugipay.Error
if errors.As(err, &pe) {
    switch pe.Code {
    case "rate_limited":
        time.Sleep(time.Second)
        // retry
    case "timeout", "network_error":
        // transport — also retry
    case "signature_invalid", "signature_stale":
        // webhook only
    }
}
```

## License

MIT
