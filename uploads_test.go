package plugipay

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// POST /api/v1/uploads/image takes multipart/form-data with the image in the field
// "file" (backend/src/routes/uploads.ts). Until 0.3 Uploads.Image sent JSON and got 400.
var testPNG, _ = base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY42YAAAAASUVORK5CYII=")

type seenUpload struct {
	method, path, auth, want, ts, idem string
	filename, partType                 string
	content                            []byte
	fields                             []string
}

// uploadServer parses each request's multipart body the way the server does.
func uploadServer(t *testing.T, seen *[]seenUpload) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := seenUpload{method: r.Method, path: r.URL.Path, auth: r.Header.Get("Authorization"), ts: r.Header.Get("X-Plugipay-Timestamp"), idem: r.Header.Get("Idempotency-Key")}
		// a multipart body is signed as an empty one (middleware/hmac-auth.ts)
		s.want = "Plugipay-HMAC-SHA256 keyId=ak_test, scope=*, signature=" + serverSignature(r, nil, "sk_test")
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("not a multipart body: %v", err)
		} else {
			for k := range r.MultipartForm.File {
				s.fields = append(s.fields, k)
			}
			if f, hdr, err := r.FormFile("file"); err == nil {
				s.filename, s.partType = hdr.Filename, hdr.Header.Get("Content-Type")
				s.content, _ = io.ReadAll(f)
			}
		}
		*seen = append(*seen, s)
		_, _ = w.Write([]byte(`{"data":{"url":"/api/v1/files/1-ab.png","fileName":"logo.png","fileSize":70},"error":null,"meta":{"requestId":"r"}}`))
	}))
}

func TestUploadsImageSendsTheFileAsMultipart(t *testing.T) {
	var seen []seenUpload
	srv := uploadServer(t, &seen)
	defer srv.Close()
	c, _ := NewClient(ClientOptions{KeyID: "ak_test", Secret: "sk_test", BaseURL: srv.URL})

	out, err := c.Uploads.Image(context.Background(), UploadImageInput{File: bytes.NewReader(testPNG), Filename: "logo.png", ContentType: "image/png"})
	if err != nil {
		t.Fatal(err)
	}
	if out.URL != "/api/v1/files/1-ab.png" || out.FileName != "logo.png" || out.FileSize != 70 {
		t.Fatalf("out = %+v", out)
	}
	s := seen[0]
	if s.method != "POST" || s.path != "/api/v1/uploads/image" || s.idem == "" || s.auth != s.want {
		t.Fatalf("sent %+v", s)
	}
	if s.filename != "logo.png" || s.partType != "image/png" || !bytes.Equal(s.content, testPNG) || strings.Join(s.fields, ",") != "file" {
		t.Fatalf("part %+v", s)
	}

	// the 0.2 shape: base64 + filename + mime
	if _, err := c.Uploads.Image(context.Background(), UploadImageInput{Base64: base64.StdEncoding.EncodeToString(testPNG), Filename: "old.png", Mime: "image/png"}); err != nil {
		t.Fatal(err)
	}
	if s := seen[1]; s.filename != "old.png" || !bytes.Equal(s.content, testPNG) {
		t.Fatalf("part %+v", s)
	}

	// no bytes: refused before any request
	if _, err := c.Uploads.Image(context.Background(), UploadImageInput{}); err == nil || len(seen) != 2 {
		t.Fatalf("err = %v, requests = %d", err, len(seen))
	}
}

func TestAPIUploadsImageSendsTheSameForm(t *testing.T) {
	var seen []seenUpload
	srv := uploadServer(t, &seen)
	defer srv.Close()
	c, _ := NewClient(ClientOptions{KeyID: "ak_test", Secret: "sk_test", BaseURL: srv.URL})

	if _, err := c.API.UploadsImage(context.Background(), &UploadsImageArgs{File: FormFile{Name: "gen.png", Content: bytes.NewReader(testPNG)}}); err != nil {
		t.Fatal(err)
	}
	s := seen[0]
	if s.auth != s.want || s.filename != "gen.png" || !bytes.Equal(s.content, testPNG) {
		t.Fatalf("sent %+v", s)
	}
	// a fresh timestamp on every request
	if ts, _ := strconv.ParseInt(s.ts, 10, 64); time.Now().Unix()-ts > 2 || ts-time.Now().Unix() > 2 {
		t.Fatalf("timestamp %s", s.ts)
	}
	if _, err := c.API.UploadsImage(context.Background(), nil); err == nil {
		t.Fatal("no file: want an error")
	}
}
