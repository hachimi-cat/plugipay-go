# Changelog

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
