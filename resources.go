package plugipay

import "encoding/json"

// ─────────────────────────────────────────────────────────────────
// API envelope
// ─────────────────────────────────────────────────────────────────

// APIEnvelope is the standard Plugipay response shape: { data, error,
// meta }. Generic-free for simplicity — callers usually decode `Data`
// into a concrete type via json.Unmarshal a second time.
type APIEnvelope struct {
	Data  json.RawMessage `json:"data"`
	Error *APIError       `json:"error,omitempty"`
	Meta  *EnvelopeMeta   `json:"meta,omitempty"`
}

// APIError is the error block inside an envelope.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	DocURL  string `json:"docUrl,omitempty"`
}

// EnvelopeMeta covers all known meta fields. Cursor + HasMore appear on
// list endpoints; RequestID + Timestamp appear on every response.
type EnvelopeMeta struct {
	RequestID string  `json:"requestId,omitempty"`
	Timestamp string  `json:"timestamp,omitempty"`
	Cursor    *string `json:"cursor,omitempty"`
	HasMore   bool    `json:"hasMore,omitempty"`
}

// Page wraps a list-endpoint result. Mirrors the Node SDK's
// requestList return shape: { data, cursor, hasMore }.
type Page[T any] struct {
	Data    []T     `json:"data"`
	Cursor  *string `json:"cursor"`
	HasMore bool    `json:"hasMore"`
}

// ─────────────────────────────────────────────────────────────────
// Enums (string types — keep them simple)
// ─────────────────────────────────────────────────────────────────

type CurrencyCode string

const (
	CurrencyIDR CurrencyCode = "IDR"
	CurrencyUSD CurrencyCode = "USD"
)

type CheckoutMethod string

const (
	CheckoutMethodQRIS    CheckoutMethod = "qris"
	CheckoutMethodVA      CheckoutMethod = "va"
	CheckoutMethodEwallet CheckoutMethod = "ewallet"
	CheckoutMethodCard    CheckoutMethod = "card"
	CheckoutMethodRetail  CheckoutMethod = "retail"
	CheckoutMethodPaypal  CheckoutMethod = "paypal"
)

type PayoutStatus string

const (
	PayoutStatusPending   PayoutStatus = "pending"
	PayoutStatusInTransit PayoutStatus = "in_transit"
	PayoutStatusPaid      PayoutStatus = "paid"
	PayoutStatusFailed    PayoutStatus = "failed"
	PayoutStatusCancelled PayoutStatus = "cancelled"
)

type PayoutMethod string

const (
	PayoutMethodManual             PayoutMethod = "manual"
	PayoutMethodXenditDisbursement PayoutMethod = "xendit_disbursement"
)

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "pending"
	RefundStatusSucceeded RefundStatus = "succeeded"
	RefundStatusFailed    RefundStatus = "failed"
)

type AdapterKind string

const (
	AdapterKindXendit   AdapterKind = "xendit"
	AdapterKindPaypal   AdapterKind = "paypal"
	AdapterKindMidtrans AdapterKind = "midtrans"
	AdapterKindManual   AdapterKind = "manual"
)

type TemplateKind string

const (
	TemplateKindCheckout TemplateKind = "checkout"
	TemplateKindReceipt  TemplateKind = "receipt"
	TemplateKindInvoice  TemplateKind = "invoice"
)

type SourceType string

const (
	SourceTypeCheckoutSession SourceType = "checkout_session"
	SourceTypeInvoice         SourceType = "invoice"
)

// ─────────────────────────────────────────────────────────────────
// Resource shapes — mirror sdk/node/src/types.ts
// ─────────────────────────────────────────────────────────────────

// Customer is a customer of the merchant.
type Customer struct {
	ID         string  `json:"id"`
	ARN        string  `json:"arn"`
	AccountID  string  `json:"accountId"`
	ExternalID *string `json:"externalId"` // your own id for it; unique per account and mode
	Email      *string `json:"email"`
	Name       *string `json:"name"`
	Phone      *string `json:"phone"`
	// TaxID is the customer's tax id (NPWP in Indonesia).
	TaxID                 *string           `json:"taxId"`
	DefaultPaymentTokenID *string           `json:"defaultPaymentTokenId"`
	Metadata              map[string]string `json:"metadata"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
}

type Plan struct {
	ID              string            `json:"id"`
	ARN             string            `json:"arn"`
	AccountID       string            `json:"accountId"`
	Name            string            `json:"name"`
	Description     *string           `json:"description"`
	Interval        string            `json:"interval"`
	IntervalCount   int               `json:"intervalCount"`
	TrialDays       int               `json:"trialDays"`
	Prices          []Price           `json:"prices"` // a plan's amounts live on its prices
	UsageAggregate  *string           `json:"usageAggregate"`
	MeteredUnit     *string           `json:"meteredUnit"`
	PortalFeatures  *PortalFeatures   `json:"portalFeatures"`
	DunningPolicyID *string           `json:"dunningPolicyId"`
	Active          bool              `json:"active"`
	ArchivedAt      *string           `json:"archivedAt"`
	Metadata        map[string]string `json:"metadata"`
	CreatedAt       string            `json:"createdAt"`
	UpdatedAt       string            `json:"updatedAt"`
}

type Price struct {
	ID         string       `json:"id"`
	PlanID     string       `json:"planId"`
	Currency   CurrencyCode `json:"currency"`
	Model      string       `json:"model"` // flat | tiered | volume | usage
	UnitAmount *int64       `json:"unitAmount"`
	Tiers      []PriceTier  `json:"tiers"`
	TaxMode    string       `json:"taxMode"`
	Active     bool         `json:"active"`
	CreatedAt  string       `json:"createdAt"`
}

// CheckoutSessionCustomer is the customer a checkout session is for, as the session shows it.
type CheckoutSessionCustomer struct {
	ID         string  `json:"id"`
	Name       *string `json:"name"`
	Email      *string `json:"email"`
	Phone      *string `json:"phone"`
	ExternalID *string `json:"externalId"`
}

// CheckoutSession is a hosted checkout. Status is open, pending, pending_review,
// completed, failed, expired or canceled.
type CheckoutSession struct {
	ID            string                   `json:"id"`
	ARN           string                   `json:"arn"`
	AccountID     string                   `json:"accountId"`
	WorkspaceName *string                  `json:"workspaceName"` // the workspace's display name, where the response carries it
	CustomerID    *string                  `json:"customerId"`
	Customer      *CheckoutSessionCustomer `json:"customer"`
	Mode          string                   `json:"mode"` // live or test
	Status        string                   `json:"status"`
	Amount        int64                    `json:"amount"`
	Currency      CurrencyCode             `json:"currency"`
	Methods       []CheckoutMethod         `json:"methods"`
	PaymentMethod *string                  `json:"paymentMethod"` // the method the buyer paid with, once known
	Adapter       *string                  `json:"adapter"`
	LineItems     json.RawMessage          `json:"lineItems"`
	SuccessURL    string                   `json:"successUrl"`
	CancelURL     string                   `json:"cancelUrl"`
	HostedURL     string                   `json:"hostedUrl"`
	ExpiresAt     string                   `json:"expiresAt"`
	CompletedAt   *string                  `json:"completedAt"`
	// PaymentID is the provider's charge id once paid — what Refunds.Create's ChargeID takes.
	PaymentID *string           `json:"paymentId"`
	Metadata  map[string]string `json:"metadata"`
	CreatedAt string            `json:"createdAt"`
	UpdatedAt string            `json:"updatedAt"`
}

// InvoiceLine is one line of an invoice.
type InvoiceLine struct {
	ID          string            `json:"id"`
	Description string            `json:"description"`
	Quantity    int64             `json:"quantity"`
	UnitAmount  int64             `json:"unitAmount"`
	Amount      int64             `json:"amount"`
	PriceID     *string           `json:"priceId"`
	Metadata    map[string]string `json:"metadata"`
}

// Invoice is an invoice. Status is draft, open, past_due, paid, void or uncollectible.
type Invoice struct {
	ID               string       `json:"id"`
	ARN              string       `json:"arn"`
	AccountID        string       `json:"accountId"`
	CustomerID       string       `json:"customerId"`
	SubscriptionID   *string      `json:"subscriptionId"`
	Status           string       `json:"status"`
	Number           string       `json:"number"`
	Currency         CurrencyCode `json:"currency"`
	Subtotal         int64        `json:"subtotal"`
	Discount         int64        `json:"discount"`
	Tax              int64        `json:"tax"`
	Total            int64        `json:"total"`
	AmountPaid       int64        `json:"amountPaid"`
	AmountDue        int64        `json:"amountDue"`
	DueAt            *string      `json:"dueAt"`
	IssuedAt         *string      `json:"issuedAt"`
	PaidAt           *string      `json:"paidAt"`
	VoidedAt         *string      `json:"voidedAt"`
	HostedInvoiceURL *string      `json:"hostedInvoiceUrl"`
	// ChargeID is the provider charge that settled the invoice — set by Invoices.Get when a
	// checkout session paid it (nil in lists and events, and when it was paid otherwise).
	// Pass it to Refunds.Create; nil means it can't be refunded through a provider.
	ChargeID           *string           `json:"chargeId"`
	CollectionAttempts int               `json:"collectionAttempts"`
	Lines              []InvoiceLine     `json:"lines"`
	Metadata           map[string]string `json:"metadata"`
	CreatedAt          string            `json:"createdAt"`
	UpdatedAt          string            `json:"updatedAt"`
}

// Subscription is a customer's subscription to a plan. Status is trialing, active,
// past_due, canceled, paused or incomplete.
type Subscription struct {
	ID                 string  `json:"id"`
	ARN                string  `json:"arn"`
	AccountID          string  `json:"accountId"`
	CustomerID         string  `json:"customerId"`
	PlanID             string  `json:"planId"`
	PriceID            *string `json:"priceId"`
	Status             string  `json:"status"`
	CurrentPeriodStart string  `json:"currentPeriodStart"`
	CurrentPeriodEnd   string  `json:"currentPeriodEnd"`
	TrialEnd           *string `json:"trialEnd"` // nil without a trial
	// CancelAt is set by a period-end cancel: the subscription cancels then.
	CancelAt   *string `json:"cancelAt"`
	CanceledAt *string `json:"canceledAt"`
	// CanceledReason is customer_portal, merchant, failed_payment or user_request.
	CanceledReason        *string           `json:"canceledReason"`
	PausedAt              *string           `json:"pausedAt"`
	DefaultPaymentTokenID *string           `json:"defaultPaymentTokenId"`
	DiscountCouponID      *string           `json:"discountCouponId"`
	CollectionMethod      string            `json:"collectionMethod"` // charge_automatically or send_invoice
	Metadata              map[string]string `json:"metadata"`
	CreatedAt             string            `json:"createdAt"`
	UpdatedAt             string            `json:"updatedAt"`
}

// PortalSession is a billing-portal link for a customer: open URL in their browser.
type PortalSession struct {
	ID         string `json:"id"`
	ARN        string `json:"arn"`
	AccountID  string `json:"accountId"`
	CustomerID string `json:"customerId"`
	URL        string `json:"url"`
	ReturnURL  string `json:"returnUrl"`
	ExpiresAt  string `json:"expiresAt"`
	CreatedAt  string `json:"createdAt"`
}

type PartnerWorkspace struct {
	AccountID     string  `json:"accountId"`
	Partner       string  `json:"partner"`
	DiscountRate  float64 `json:"discountRate"`
	BrandName     *string `json:"brandName"`
	BusinessEmail *string `json:"businessEmail"`
	CreatedAt     string  `json:"createdAt"`
}

type Payout struct {
	ID                  string       `json:"id"`
	AccountID           string       `json:"accountId"`
	Amount              int64        `json:"amount"`
	Currency            string       `json:"currency"`
	Status              PayoutStatus `json:"status"`
	Method              PayoutMethod `json:"method"`
	BankCode            *string      `json:"bankCode"`
	BankName            string       `json:"bankName"`
	BankAccountNumber   string       `json:"bankAccountNumber"`
	BankAccountHolder   string       `json:"bankAccountHolder"`
	Note                *string      `json:"note"`
	Reference           *string      `json:"reference"`
	FailureReason       *string      `json:"failureReason"`
	LedgerTransactionID *string      `json:"ledgerTransactionId"`
	ProcessedAt         *string      `json:"processedAt"`
	CompletedAt         *string      `json:"completedAt"`
	CreatedAt           string       `json:"createdAt"`
	UpdatedAt           string       `json:"updatedAt"`
}

type AvailableBalance struct {
	LedgerBalance int64   `json:"ledgerBalance"`
	Locked        int64   `json:"locked"`
	Available     int64   `json:"available"`
	Currency      *string `json:"currency"`
}

type BankAccount struct {
	BankCode          *string `json:"bankCode"`
	BankName          *string `json:"bankName"`
	BankAccountNumber *string `json:"bankAccountNumber"`
	BankAccountHolder *string `json:"bankAccountHolder"`
	Configured        bool    `json:"configured"`
}

type LedgerEntry struct {
	ID         string  `json:"id"`
	AccountID  string  `json:"accountId"`
	TxID       string  `json:"txId"`
	Code       string  `json:"code"`
	Direction  string  `json:"direction"`
	Amount     int64   `json:"amount"`
	Currency   string  `json:"currency"`
	SourceType string  `json:"sourceType"`
	SourceID   string  `json:"sourceId"`
	Memo       *string `json:"memo"`
	PostedAt   string  `json:"postedAt"`
}

type LedgerBalance struct {
	Code    string `json:"code"`
	Debits  int64  `json:"debits"`
	Credits int64  `json:"credits"`
	Balance int64  `json:"balance"`
}

type PnLReportLine struct {
	Code   string `json:"code"`
	Amount int64  `json:"amount"`
}

type PnLReport struct {
	From         string          `json:"from"`
	To           string          `json:"to"`
	Currency     string          `json:"currency"`
	Revenue      int64           `json:"revenue"`
	Refunds      int64           `json:"refunds"`
	PlatformFees int64           `json:"platformFees"`
	Tax          int64           `json:"tax"`
	Net          int64           `json:"net"`
	Lines        []PnLReportLine `json:"lines"`
}

type CashFlowBucket struct {
	Day     string `json:"day"`
	Inflow  int64  `json:"inflow"`
	Outflow int64  `json:"outflow"`
	Net     int64  `json:"net"`
}

type CashFlowReport struct {
	From         string           `json:"from"`
	To           string           `json:"to"`
	Currency     string           `json:"currency"`
	TotalInflow  int64            `json:"totalInflow"`
	TotalOutflow int64            `json:"totalOutflow"`
	Net          int64            `json:"net"`
	Buckets      []CashFlowBucket `json:"buckets"`
}

type WebhookEndpoint struct {
	ID          string   `json:"id"`
	AccountID   string   `json:"accountId,omitempty"` // only on create
	Mode        string   `json:"mode"`                // live or test: it receives only that mode's events
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description *string  `json:"description"`
	Active      bool     `json:"active"`
	// ConsecutiveFailures counts failed delivery attempts in a row since the last 2xx;
	// FailingSince is when that run started (nil while healthy).
	ConsecutiveFailures int     `json:"consecutiveFailures"`
	FailingSince        *string `json:"failingSince"`
	// DisabledAt / DisabledReason are set when Plugipay switched the endpoint off because
	// it kept failing (20 failed attempts in a row over at least 24 hours); nil for a
	// manual pause. Re-enable with Update(…, Active: true).
	DisabledAt     *string `json:"disabledAt"`
	DisabledReason *string `json:"disabledReason"`
	Secret         string  `json:"secret,omitempty"` // only on create
	CreatedAt      string  `json:"createdAt"`
	UpdatedAt      string  `json:"updatedAt"`
}

type WebhookDeliveryStatus string

const (
	WebhookDeliveryPending   WebhookDeliveryStatus = "pending"
	WebhookDeliverySucceeded WebhookDeliveryStatus = "succeeded"
	WebhookDeliveryFailed    WebhookDeliveryStatus = "failed"
)

// WebhookDelivery is one event sent to one endpoint (WebhookEndpoints.ListDeliveries).
type WebhookDelivery struct {
	ID         string `json:"id"`
	EndpointID string `json:"endpointId"`
	// EventID is the event's id (evt_…): the body's id, the same on every attempt.
	EventID string `json:"eventId"`
	Type    string `json:"type"`
	// Body is the exact JSON sent on every attempt.
	Body     string                `json:"body"`
	Status   WebhookDeliveryStatus `json:"status"`
	Attempts int                   `json:"attempts"`
	// NextRetryAt is when it is next due; nil once succeeded or failed.
	NextRetryAt   *string `json:"nextRetryAt"`
	LastAttemptAt *string `json:"lastAttemptAt"`
	DeliveredAt   *string `json:"deliveredAt"`
	ResponseCode  *int    `json:"responseCode"`
	LastError     *string `json:"lastError"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
	// AttemptLog is every attempt, oldest first.
	AttemptLog []WebhookDeliveryAttempt `json:"attemptLog"`
}

// WebhookDeliveryAttempt is one try at a delivery.
type WebhookDeliveryAttempt struct {
	AttemptNumber int    `json:"attemptNumber"`
	Status        string `json:"status"` // succeeded or failed
	// ResponseCode is nil when no response came back (a timeout, a refused connection).
	ResponseCode *int `json:"responseCode"`
	DurationMs   int  `json:"durationMs"`
	// Error is why it failed: "HTTP 503", "timed out after 10000ms", …
	Error *string `json:"error"`
	// NextRetryAt is the retry this failure scheduled; nil on success or when it gave up.
	NextRetryAt *string `json:"nextRetryAt"`
	AttemptedAt string  `json:"attemptedAt"`
}

// EventRecord is an event as Events.List / Events.Get return it.
type EventRecord struct {
	ID        string  `json:"id"`
	Type      string  `json:"type"`
	AccountID *string `json:"accountId"`
	// Mode is live or test; nil for events from before events carried one.
	Mode       *string         `json:"mode"`
	OccurredAt string          `json:"occurredAt"`
	Data       json.RawMessage `json:"data"`
	// Metadata is delivery bookkeeping: idempotencyKey, mode, requestId, source, replayOf.
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt string          `json:"createdAt"`
	// PublishedAt is when the outbox worker sent it to your endpoints; nil while queued.
	PublishedAt *string `json:"publishedAt"`
}

type ReceiptSummary struct {
	ID         string  `json:"id"`
	Number     string  `json:"number"`
	SourceType string  `json:"sourceType"`
	SourceID   string  `json:"sourceId"`
	CustomerID *string `json:"customerId"`
	Amount     int64   `json:"amount"`
	Currency   string  `json:"currency"`
	Method     *string `json:"method"`
	Adapter    *string `json:"adapter"`
	IssuedAt   string  `json:"issuedAt"`
	EmailedAt  *string `json:"emailedAt"`
	EmailedTo  *string `json:"emailedTo"`
}

type PartnerUsageLine struct {
	AccountID        string  `json:"accountId"`
	BrandName        *string `json:"brandName"`
	TransactionCount int64   `json:"transactionCount"`
	GrossVolume      int64   `json:"grossVolume"`
	DiscountRate     float64 `json:"discountRate"`
	Fee              int64   `json:"fee"`
}

type PartnerUsageSummary struct {
	Partner  string             `json:"partner"`
	From     string             `json:"from"`
	To       string             `json:"to"`
	Currency string             `json:"currency"`
	Lines    []PartnerUsageLine `json:"lines"`
	Total    int64              `json:"total"`
}

// Refund is money sent back for a charge. Status is pending, succeeded or failed.
type Refund struct {
	ID        string `json:"id"`
	ARN       string `json:"arn"`
	AccountID string `json:"accountId"`
	// ChargeID is the provider charge refunded (a checkout session's PaymentID).
	ChargeID  string       `json:"chargeId"`
	InvoiceID *string      `json:"invoiceId"` // the invoice it refunds, when the charge paid one
	Amount    int64        `json:"amount"`
	Currency  CurrencyCode `json:"currency"`
	// Reason is requested_by_customer, duplicate, fraudulent or other.
	Reason string       `json:"reason"`
	Status RefundStatus `json:"status"`
	// FailureCode / FailureMessage are the provider's reason (failed only).
	FailureCode    *string           `json:"failureCode"`
	FailureMessage *string           `json:"failureMessage"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

// GiftCard is a gift card (CustomerID nil) or store credit (CustomerID set). Status is
// active, redeemed, void or expired.
type GiftCard struct {
	ID             string            `json:"id"`
	ARN            string            `json:"arn"`
	AccountID      string            `json:"accountId"`
	Mode           string            `json:"mode"`
	Code           string            `json:"code"`
	Currency       CurrencyCode      `json:"currency"`
	InitialBalance int64             `json:"initialBalance"`
	Balance        int64             `json:"balance"`
	Status         string            `json:"status"`
	CustomerID     *string           `json:"customerId"`
	Kind           string            `json:"kind"` // store_credit or gift_card
	ExpiresAt      *string           `json:"expiresAt"`
	Note           *string           `json:"note"`
	IssuedSource   *string           `json:"issuedSource"`
	IssuedRef      *string           `json:"issuedRef"`
	Metadata       map[string]string `json:"metadata"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

// GiftCardEntry is one movement of a card's balance (Kind issue, redeem, topup or void).
type GiftCardEntry struct {
	ID                string       `json:"id"`
	AccountID         string       `json:"accountId"`
	GiftCardID        string       `json:"giftCardId"`
	Kind              string       `json:"kind"`
	Delta             int64        `json:"delta"`
	BalanceAfter      int64        `json:"balanceAfter"`
	Currency          CurrencyCode `json:"currency"`
	ExternalSource    string       `json:"externalSource"`
	ExternalRef       string       `json:"externalRef"`
	CheckoutSessionID *string      `json:"checkoutSessionId"`
	LedgerTxID        *string      `json:"ledgerTxId"`
	Note              *string      `json:"note"`
	CreatedAt         string       `json:"createdAt"`
}

// AdapterConfig is a connected provider, per mode. Secrets are never returned:
// SecretKeyLast4 and the masked PublicConfig stand for them.
type AdapterConfig struct {
	Kind           string         `json:"kind"`   // xendit | paypal | midtrans | manual | managed
	Status         string         `json:"status"` // unconfigured | active | error
	SecretKeyLast4 *string        `json:"secretKeyLast4"`
	PublicConfig   map[string]any `json:"publicConfig"`
	ConfiguredAt   *string        `json:"configuredAt"`
	LastErrorAt    *string        `json:"lastErrorAt"`
	LastErrorCode  *string        `json:"lastErrorCode"`
}

// ManagedOnboardingState is the managed (xenPlatform) sub-account behind managed payments.
type ManagedOnboardingState struct {
	SubAccountID       string  `json:"subAccountId"`
	Email              *string `json:"email"`
	OnboardingURL      *string `json:"onboardingUrl"`
	KybStatus          string  `json:"kybStatus"`
	CapabilitiesStatus string  `json:"capabilitiesStatus"`
	PayoutsReady       bool    `json:"payoutsReady"`
	LastWebhookAt      *string `json:"lastWebhookAt"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}

// ApiKey is a dashboard API key (pk_test_… / pk_live_…). Revoking deletes it.
type ApiKey struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// KeyPrefix is the key's first 12 characters, to tell keys apart.
	KeyPrefix   string   `json:"keyPrefix"`
	Environment string   `json:"environment"` // test or live
	Scopes      []string `json:"scopes"`
	LastUsedAt  *string  `json:"lastUsedAt"`
	CreatedAt   string   `json:"createdAt"`
	// Key is the whole key — only on create.
	Key string `json:"key,omitempty"`
}

type Template struct {
	ID        string         `json:"id"`
	AccountID string         `json:"accountId"`
	Kind      TemplateKind   `json:"kind"`
	Name      string         `json:"name"`
	IsDefault bool           `json:"isDefault"`
	Config    map[string]any `json:"config"` // the kind's settings
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// UploadedFile is an uploaded image (POST /api/v1/uploads/image).
type UploadedFile struct {
	// URL is where it is served: /api/v1/files/<name>, relative to plugipay.com.
	URL string `json:"url"`
	// FileName is the name it was sent under.
	FileName string `json:"fileName"`
	// FileSize is its size in bytes.
	FileSize int64 `json:"fileSize"`
}

// Workspace is a workspace (a Huudis account). A key's Workspaces.List is its own
// workspace, with ID, Name, Slug and Role only; a person's comes from Huudis with the rest.
type Workspace struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Slug              string  `json:"slug"`
	Role              string  `json:"role,omitempty"` // owner, admin or member
	CreatedAt         string  `json:"createdAt,omitempty"`
	JoinedAt          string  `json:"joinedAt,omitempty"`
	IsActive          bool    `json:"isActive,omitempty"`
	IsForjioInternal  bool    `json:"isForjioInternal,omitempty"`
	PendingDeletionAt *string `json:"pendingDeletionAt,omitempty"`
}

// WorkspaceDeletion is Workspaces.Delete's answer: the deletion happens at PendingDeletionAt.
type WorkspaceDeletion struct {
	Scheduled         bool   `json:"scheduled"`
	PendingDeletionAt string `json:"pendingDeletionAt"`
}

// AccountProfile is the signed-in person's Huudis profile. When Huudis can't be reached
// the API answers from the session with ID, Email, Name and EmailVerified only.
type AccountProfile struct {
	ID                string                     `json:"id"`
	Email             string                     `json:"email"`
	Name              *string                    `json:"name"`
	EmailVerified     bool                       `json:"emailVerified"`
	Locale            *string                    `json:"locale,omitempty"`
	HasPassword       *bool                      `json:"hasPassword,omitempty"` // false: signs in with Google / Apple only
	MfaEnabled        *bool                      `json:"mfaEnabled,omitempty"`
	PendingDeletionAt *string                    `json:"pendingDeletionAt,omitempty"`
	CreatedAt         *string                    `json:"createdAt,omitempty"`
	LastLoginAt       *string                    `json:"lastLoginAt,omitempty"`
	Memberships       []AccountProfileMembership `json:"memberships,omitempty"`
}

// AccountProfileMembership is one workspace the person belongs to.
type AccountProfileMembership struct {
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
	Account  struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"account"`
}

// AccountProfileUpdate is Account.Update's answer: the fields it changes.
type AccountProfileUpdate struct {
	ID     string  `json:"id"`
	Name   *string `json:"name"`
	Locale *string `json:"locale"`
}

// BrowserSession is one of the person's signed-in sessions.
type BrowserSession struct {
	ID         string  `json:"id"`
	UserAgent  *string `json:"userAgent"`
	IP         *string `json:"ip"`
	CreatedAt  string  `json:"createdAt"`
	LastUsedAt *string `json:"lastUsedAt"`
	ExpiresAt  string  `json:"expiresAt"`
	Current    bool    `json:"current"` // the session making this request
}

// LinkedAccount is a Google or Apple sign-in linked to the person.
type LinkedAccount struct {
	ID       string  `json:"id"`
	Provider string  `json:"provider"` // google or apple
	Email    *string `json:"email"`
	LinkedAt string  `json:"linkedAt"`
}

// LinkedAccounts is Account.ListLinked's answer: the linked sign-ins, and whether the
// person also has a password.
type LinkedAccounts struct {
	HasPassword bool            `json:"hasPassword"`
	Providers   []LinkedAccount `json:"providers"`
}

// WorkspaceMember is a member of the active workspace (its Huudis IAM users).
type WorkspaceMember struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          *string `json:"name"`
	EmailVerified bool    `json:"emailVerified"`
	Role          string  `json:"role"` // owner, admin or member
	JoinedAt      string  `json:"joinedAt"`
	LastLoginAt   *string `json:"lastLoginAt"`
	CreatedAt     string  `json:"createdAt"`
	IsYou         bool    `json:"isYou"` // the person making this request
	Groups        []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"groups"`
}

// BillingTier is one of Plugipay's own plans (Billing.ListTiers). A nil limit is
// unlimited.
type BillingTier struct {
	ID   string `json:"id"` // starter, growth, scale or enterprise
	Name string `json:"name"`
	// PriceMonthlyIDR is IDR per month; nil when negotiated.
	PriceMonthlyIDR *int64 `json:"priceMonthlyIdr"`
	// PriceMonthlyUSDCents is USD cents per month for merchants billed in USD; nil when
	// the tier has no USD price.
	PriceMonthlyUSDCents *int64 `json:"priceMonthlyUsdCents"`
	// ChannelFeeRate is Plugipay's fee per transaction, 0–1.
	ChannelFeeRate      float64            `json:"channelFeeRate"`
	MonthlyTxnCap       *int64             `json:"monthlyTxnCap"`
	MaxWebhookEndpoints *int64             `json:"maxWebhookEndpoints"`
	MaxAPIKeys          *int64             `json:"maxApiKeys"`
	CustomBranding      bool               `json:"customBranding"`
	DailyPayouts        bool               `json:"dailyPayouts"`
	Support             BillingTierSupport `json:"support"`
	Tagline             string             `json:"tagline"`
	// AgentCredits is the monthly assistant credits the tier grants.
	AgentCredits int64    `json:"agentCredits"`
	Features     []string `json:"features"`
}

// BillingTierSupport is a tier's support level.
type BillingTierSupport struct {
	Tier          string `json:"tier"` // community, email, priority or dedicated
	ResponseHours *int   `json:"responseHours"`
	SLA           bool   `json:"sla"`
}

// BillingPlan is one of Plugipay's own plans (Billing.ListPlans): Price is IDR per
// month, -1 when the plan has no fixed price.
type BillingPlan struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

// CheckoutSettings: the hosted checkout's payment methods, branding and business details.
type CheckoutSettings struct {
	EnabledMethods   []string            `json:"enabledMethods"`
	MethodOrder      []string            `json:"methodOrder"`
	MethodAdapter    map[string]string   `json:"methodAdapter"`
	BrandName        *string             `json:"brandName"`
	BrandLogoURL     *string             `json:"brandLogoUrl"`
	BrandAccentColor *string             `json:"brandAccentColor"`
	BrandTagline     *string             `json:"brandTagline"`
	BusinessPhone    *string             `json:"businessPhone"`
	BusinessEmail    *string             `json:"businessEmail"`
	BusinessAddress  *string             `json:"businessAddress"`
	BusinessTaxID    *string             `json:"businessTaxId"`
	ReceiptTemplate  map[string]any      `json:"receiptTemplate"`
	AvailableMethods []string            `json:"availableMethods"` // computed
	MethodSupport    map[string][]string `json:"methodSupport"`    // computed
}

// AdminPortalIdentity is AdminPortal.Me's answer: the account the probe was scoped to,
// and whether it is a Forjio-internal workspace.
type AdminPortalIdentity struct {
	AccountID        string `json:"accountId"`
	IsForjioInternal bool   `json:"isForjioInternal"`
}

// WebhookEvent is the parsed webhook payload. The shape is type-tagged:
// inspect Type, then unmarshal Data.Object into the matching resource
// struct (CheckoutSession, Invoice, Subscription). Mirrors the
// discriminated-union in sdk/node/src/types.ts#WebhookEvent.
type WebhookEvent struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	AccountID  string           `json:"accountId"`
	OccurredAt string           `json:"occurredAt"`
	Data       WebhookEventData `json:"data"`
}

// WebhookEventData is an event's data: Object is the resource after the change (decode it
// into the struct its Type names); AggregateType / AggregateID name that resource. A few
// types carry more — To (invoice.sent), Reason (invoice.failed), Entry (gift_card.redeemed
// and .topped_up), DaysAdvanced (subscription.renewed from a test clock).
type WebhookEventData struct {
	Object        json.RawMessage `json:"object"`
	AggregateType string          `json:"aggregateType"`
	AggregateID   string          `json:"aggregateId"`
	To            *string         `json:"to,omitempty"`
	Reason        *string         `json:"reason,omitempty"`
	Entry         json.RawMessage `json:"entry,omitempty"`
	DaysAdvanced  *int            `json:"daysAdvanced,omitempty"`
}
