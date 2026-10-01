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
	RefundStatusCanceled  RefundStatus = "canceled"
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

type Customer struct {
	ID         string  `json:"id"`
	AccountID  string  `json:"accountId"`
	Email      *string `json:"email"`
	Name       *string `json:"name"`
	Phone      *string `json:"phone"`
	ExternalID *string `json:"externalId"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
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

type CheckoutSession struct {
	ID          string            `json:"id"`
	AccountID   string            `json:"accountId"`
	CustomerID  *string           `json:"customerId"`
	Amount      int64             `json:"amount"`
	Currency    CurrencyCode      `json:"currency"`
	Status      string            `json:"status"`
	Methods     []CheckoutMethod  `json:"methods"`
	Adapter     *string           `json:"adapter"`
	LineItems   json.RawMessage   `json:"lineItems"`
	SuccessURL  string            `json:"successUrl"`
	CancelURL   string            `json:"cancelUrl"`
	HostedURL   string            `json:"hostedUrl"`
	ExpiresAt   string            `json:"expiresAt"`
	CompletedAt *string           `json:"completedAt"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"createdAt"`
	UpdatedAt   string            `json:"updatedAt"`
}

type InvoiceLine struct {
	ID          string `json:"id"`
	Description string `json:"description"`
	Quantity    int64  `json:"quantity"`
	UnitAmount  int64  `json:"unitAmount"`
	Amount      int64  `json:"amount"`
}

type Invoice struct {
	ID               string        `json:"id"`
	AccountID        string        `json:"accountId"`
	CustomerID       string        `json:"customerId"`
	Status           string        `json:"status"`
	Number           string        `json:"number"`
	Currency         CurrencyCode  `json:"currency"`
	Subtotal         int64         `json:"subtotal"`
	Discount         int64         `json:"discount"`
	Tax              int64         `json:"tax"`
	Total            int64         `json:"total"`
	AmountPaid       int64         `json:"amountPaid"`
	AmountDue        int64         `json:"amountDue"`
	DueAt            *string       `json:"dueAt"`
	IssuedAt         *string       `json:"issuedAt"`
	PaidAt           *string       `json:"paidAt"`
	HostedInvoiceURL *string       `json:"hostedInvoiceUrl"`
	Lines            []InvoiceLine `json:"lines"`
	CreatedAt        string        `json:"createdAt"`
	UpdatedAt        string        `json:"updatedAt"`
}

type Subscription struct {
	ID                 string  `json:"id"`
	AccountID          string  `json:"accountId"`
	CustomerID         string  `json:"customerId"`
	PlanID             string  `json:"planId"`
	Status             string  `json:"status"`
	CurrentPeriodStart string  `json:"currentPeriodStart"`
	CurrentPeriodEnd   string  `json:"currentPeriodEnd"`
	CancelAtPeriodEnd  bool    `json:"cancelAtPeriodEnd"`
	TrialEndsAt        *string `json:"trialEndsAt"`
	CreatedAt          string  `json:"createdAt"`
	UpdatedAt          string  `json:"updatedAt"`
}

type PortalSession struct {
	ID         string `json:"id"`
	CustomerID string `json:"customerId"`
	URL        string `json:"url"`
	ReturnURL  string `json:"returnUrl"`
	ExpiresAt  string `json:"expiresAt"`
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
	AccountID   string   `json:"accountId"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Description *string  `json:"description"`
	Active      bool     `json:"active"`
	Secret      string   `json:"secret,omitempty"` // only on create
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

type EventRecord struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	AccountID  string          `json:"accountId"`
	OccurredAt string          `json:"occurredAt"`
	Data       json.RawMessage `json:"data"`
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

type Refund struct {
	ID            string       `json:"id"`
	AccountID     string       `json:"accountId"`
	Amount        int64        `json:"amount"`
	Currency      CurrencyCode `json:"currency"`
	Status        RefundStatus `json:"status"`
	Reason        *string      `json:"reason"`
	SourceType    SourceType   `json:"sourceType"`
	SourceID      string       `json:"sourceId"`
	FailureReason *string      `json:"failureReason"`
	CreatedAt     string       `json:"createdAt"`
	UpdatedAt     string       `json:"updatedAt"`
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

type ApiKey struct {
	ID          string  `json:"id"`
	AccountID   string  `json:"accountId"`
	KeyID       string  `json:"keyId"`
	Description *string `json:"description"`
	Scope       string  `json:"scope"`
	Secret      string  `json:"secret,omitempty"` // only on create
	CreatedAt   string  `json:"createdAt"`
	RevokedAt   *string `json:"revokedAt"`
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

type Workspace struct {
	ID            string  `json:"id"`
	AccountID     string  `json:"accountId"`
	BrandName     *string `json:"brandName"`
	BusinessEmail *string `json:"businessEmail"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

type AccountProfile struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"emailVerified"`
	Name          *string `json:"name"`
	MfaEnrolled   bool    `json:"mfaEnrolled"`
	CreatedAt     string  `json:"createdAt"`
}

type BrowserSession struct {
	ID         string  `json:"id"`
	UserAgent  *string `json:"userAgent"`
	IPAddress  *string `json:"ipAddress"`
	Current    bool    `json:"current"`
	CreatedAt  string  `json:"createdAt"`
	LastSeenAt string  `json:"lastSeenAt"`
}

type LinkedAccount struct {
	Provider string  `json:"provider"`
	Subject  string  `json:"subject"`
	Email    *string `json:"email"`
	LinkedAt string  `json:"linkedAt"`
}

type WorkspaceMember struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

type BillingTier struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Monthly  int64    `json:"monthly"`
	Features []string `json:"features"`
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

type AdminPortalIdentity struct {
	IdentityID string `json:"identityId"`
	Email      string `json:"email"`
	Role       string `json:"role"`
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

// WebhookEventData wraps the resource snapshot at event time. Object is
// kept as a RawMessage so callers can decode it into the concrete type
// they care about.
type WebhookEventData struct {
	Object json.RawMessage `json:"object"`
}
