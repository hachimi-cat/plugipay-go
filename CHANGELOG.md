# Changelog

## 0.5.0
- `WebhookEndpoints.Update(ctx, id, WebhookEndpointUpdateInput)` (`Active: &true` re-enables an endpoint Plugipay switched off for failing and clears its failure streak), and the delivery log: `WebhookEndpoints.ListDeliveries(ctx, WebhookDeliveryListParams)` (a `Page[WebhookDelivery]`), `GetDelivery`, `RetryDelivery`. New types `WebhookDelivery`, `WebhookDeliveryAttempt`, `WebhookDeliveryStatus`.
- `WebhookEndpoint` has `Mode`, `ConsecutiveFailures`, `FailingSince`, `DisabledAt` and `DisabledReason`.
- `ApiKey` is the key the API returns: `ID`, `Name`, `KeyPrefix`, `Environment`, `Scopes`, `LastUsedAt`, `CreatedAt`, `Key` (it had `KeyID`, `Description`, `Scope`, `Secret`, `RevokedAt`, which the API never sent).
- `BillingTier` is the tier the API returns: `PriceMonthlyIDR`, `PriceMonthlyUSDCents`, `ChannelFeeRate`, `MonthlyTxnCap` / `MaxWebhookEndpoints` / `MaxAPIKeys` (nil is unlimited), `CustomBranding`, `DailyPayouts`, `Support` (`BillingTierSupport`), `Tagline`, `AgentCredits`, `Features` (it had `Monthly`, which the API never sent).
- `client.API`: `WebhooksDeliveries`, `WebhooksGetDeliveries`, `WebhooksDeliveriesRetry` (regenerated).

## 0.4.0
Hand-written methods that could never succeed against the API now send what their routes require. Some signatures and structs changed (a minor bump in 0.x); every changed call failed before.
- `Customers.Update` sends an `Idempotency-Key` (every call was `400`); `CustomerUpdateInput` gains `ExternalID`, `TaxID`, `DefaultPaymentTokenID`, `Metadata`.
- `PlanCreateInput` is the API's shape: `Name`, `Interval`, `Prices []PriceInput`, `PortalFeatures *PortalFeatures` (nil sends `DefaultPortalFeatures`), `Description`, `IntervalCount`, `TrialDays`, …. `Currency` + `Amount` are the shorthand for one flat price. `Plan` is what the API returns (`Prices []Price`, `PortalFeatures`, `IntervalCount`, …; no plan-level `Currency` / `Amount`).
- Templates: `Template.Config`, `TemplateCreateInput{Kind, Name, IsDefault, Config}`, `TemplateUpdateInput{Name, Config}`, `TemplatePreviewInput{Kind, Config}` (the API never took `Document`; `SampleData` is gone). `Preview` returns the HTML page the route answers with (it failed with `invalid_response`). `Duplicate` with a name renames the copy with a second request.
- Any 2xx with an empty body (every DELETE answers 204) is success; it failed with `invalid_response`, so `Templates.Delete`, `WebhookEndpoints.Delete`, `ApiKeys.Revoke`, … never succeeded.
- `Adapters.UpdateXendit` / `UpdatePaypal` / `UpdateMidtrans` / `UpdateManual` take typed inputs (`XenditAdapterInput`, …) and send an `Idempotency-Key` (every call was `400`). `Adapters.List` returns the adapters sorted by kind (it failed decoding the object the API answers). `AdapterConfig` and `ManagedOnboardingState` are the API's shapes.
- `ManagedOnboardingStartInput{Email}`, `ManagedOnboardingSimulateInput{KybStatus, CapabilitiesStatus, PayoutsReady}`; `Onboarding.ProvisionManaged(ctx)` takes no input and returns `*ProvisionManagedResult`.
- `CheckoutSettingsUpdateInput` has the fields the API takes (`BrandName`, `BrandAccentColor`, `Business*`, `EnabledMethods`, …); `BrandColor`, `DefaultTemplateID`, `TermsURL`, `PrivacyURL` were refused. `CheckoutSettings` is the API's shape. `Billing.ListPlans` returns `[]BillingPlan`.
- `ApiKeyCreateInput{Name, Environment, Scopes}` (person-only: a key still gets `403`). `Invoices.Create` and `Invoices.Finalize` send an `Idempotency-Key`.
- The routes-exist test now also fails a hand-written method that sends no `Idempotency-Key` where the route takes one, or whose body (every input field filled in, by reflection) the route's schema refuses.

## 0.3.0
- `Uploads.Image` sends the image as `multipart/form-data` in the field `file`, which is what `POST /api/v1/uploads/image` takes; it sent JSON before and always got 400. `UploadImageInput` gains `File io.Reader` and `ContentType` (`Base64` + `Filename` + `Mime` still work). `UploadedFile` is what the API returns: `URL`, `FileName`, `FileSize` (it listed fields the API never sent).
- `API.UploadsImage(ctx, &UploadsImageArgs{File: FormFile{…}})` and `API.PublicCheckoutSessionsProofImage` upload a file (`FormFile{Name, Content io.Reader}`); they took no file before.
- Every request is signed with a fresh `X-Plugipay-Timestamp` when it is sent; the API now refuses a signed time more than 300 s off its clock either way (`401 timestamp_skew`, `401 invalid_timestamp`).

## 0.2.0
- `Client.API`: every Plugipay feature route, one method each (generated from the API spec: `api_generated.go`), signed like every other call, with an idempotency key on writes.
- `ClientOptions.APIKey` (or `PLUGIPAY_API_KEY`): a key minted in the dashboard (`pk_live_…` / `pk_test_…`) is sent as `Authorization: Bearer <key>`; `KeyID` + `Secret` are then not needed.
- `WebhookEndpoints.List/Create/Delete` called `/api/v1/webhook-endpoints`, which the API does not serve (every call 404'd); they now call `/api/v1/webhooks`.
- The API now takes a key on every customer feature. Person-only (403 person_only with a key): `ApiKeys`, `Account`, `Workspaces.Create/Update/Delete`; `AdminPortal` is for Plugipay's operators.
- A test checks every hand-written method's route against the API spec (`backend/openapi.json`).

## 0.1.0
- Initial release: full resource coverage. Module path is github.com/hachimi-cat/plugipay-go.
