package source

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/yeisme/promptrepo"
)

// sentinelCredential is a sentinel secret value: it may only ever appear in
// the outbound Authorization header. Any other occurrence (errors, results,
// receipts) is a leak.
const sentinelCredential = "sk-sentinel-do-not-leak"

type fakeResolver struct {
	calls atomic.Int64
	deny  atomic.Bool
	got   atomic.Value // last grant
}

func (r *fakeResolver) ResolveCredential(_ context.Context, ref string, grant CredentialGrant) (CredentialResolution, error) {
	r.calls.Add(1)
	r.got.Store(grant)
	if r.deny.Load() {
		return CredentialResolution{}, errors.New("consumer_not_allowed")
	}
	return CredentialResolution{Secret: []byte(sentinelCredential)}, nil
}

func newCredentialTestServer(t *testing.T, requireAuth bool, hits *atomic.Int64) *httptest.Server {
	t.Helper()
	catalog := testCatalog(t)
	payload, _ := json.Marshal(catalog)
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		hits.Add(1)
		if requireAuth && request.Header.Get("Authorization") != "Bearer "+sentinelCredential {
			writer.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch request.URL.Path {
		case "/prompt-bucket/catalog/catalog.json":
			writer.Header().Set("ETag", `"fixture-etag"`)
			_, _ = writer.Write(payload)
		case "/prompt-bucket/catalog/prompts/main.zh-CN.md":
			_, _ = writer.Write([]byte(testTemplateBody))
		case "/prompt-bucket/catalog/contracts/main.zh-CN.json":
			_, _ = writer.Write([]byte(`{"schema_version":"promptrepo.template-contract.v0.1"}`))
		default:
			t.Fatalf("path: %s", request.URL.Path)
		}
	}))
}

func TestS3CredentialAdapterResolvesAtUsePoint(t *testing.T) {
	var hits atomic.Int64
	server := newCredentialTestServer(t, true, &hits)
	defer server.Close()
	resolver := &fakeResolver{}
	adapter := S3CredentialAdapter{
		S3Adapter:  S3Adapter{HTTPClient: server.Client()},
		Resolver:   resolver,
		Consumer:   "template-registry",
		Capability: "fetch",
	}
	sourceURI := "s3://prompt-bucket/catalog?endpoint=" + url.QueryEscape(server.URL) + "&path_style=true"
	profile := promptrepo.RepositoryProfile{Source: sourceURI, CredentialRef: "yeisme-credential://prompt/s3"}

	result, err := adapter.Sync(context.Background(), profile, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != "s3:fixture-etag" || result.Catalog.Repository.ID != "fixture" {
		t.Fatalf("result: %+v", result)
	}
	grant := resolver.got.Load().(CredentialGrant)
	if grant.Consumer != "template-registry" || grant.Capability != "fetch" || grant.Operation != "sync" {
		t.Fatalf("grant: %+v", grant)
	}
	if resolver.calls.Load() != 1 {
		t.Fatalf("resolver must be called exactly once per request, got %d", resolver.calls.Load())
	}

	templatePayload, err := adapter.ReadTemplate(context.Background(), profile, promptrepo.SnapshotMetadata{Revision: result.Revision}, result.Catalog.Solutions[0].Templates[0], "")
	if err != nil || string(templatePayload) != testTemplateBody {
		t.Fatalf("template: %q, %v", templatePayload, err)
	}
	companionPayload, err := adapter.ReadCompanion(context.Background(), profile, promptrepo.SnapshotMetadata{Revision: result.Revision}, "contracts/main.zh-CN.json", "")
	if err != nil || len(companionPayload) == 0 {
		t.Fatalf("companion: %q, %v", companionPayload, err)
	}
	// one resolution per request: three requests, three resolutions
	if resolver.calls.Load() != 3 {
		t.Fatalf("expected 3 use-point resolutions, got %d", resolver.calls.Load())
	}
	// sentinel hygiene: the sentinel may not survive in any result surface
	rendered, _ := json.Marshal(result)
	if strings.Contains(string(rendered), sentinelCredential) {
		t.Fatalf("sync result leaks sentinel: %s", rendered)
	}
}

func TestS3CredentialAdapterDenialFailsClosedWithoutRequest(t *testing.T) {
	var hits atomic.Int64
	server := newCredentialTestServer(t, true, &hits)
	defer server.Close()
	resolver := &fakeResolver{}
	resolver.deny.Store(true)
	adapter := S3CredentialAdapter{
		S3Adapter:  S3Adapter{HTTPClient: server.Client()},
		Resolver:   resolver,
		Consumer:   "template-registry",
		Capability: "fetch",
	}
	profile := promptrepo.RepositoryProfile{
		Source:        "s3://prompt-bucket/catalog?endpoint=" + url.QueryEscape(server.URL) + "&path_style=true",
		CredentialRef: "yeisme-credential://prompt/s3",
	}
	_, err := adapter.Sync(context.Background(), profile, "")
	if promptrepo.ErrorCode(err) != promptrepo.CodeAuthRequired {
		t.Fatalf("denial must surface as auth-required, got %v", err)
	}
	if strings.Contains(err.Error(), sentinelCredential) || strings.Contains(err.Error(), "consumer_not_allowed") {
		t.Fatalf("denial must not echo resolver detail: %v", err)
	}
	if hits.Load() != 0 {
		t.Fatalf("denied resolution must not reach the source, got %d hits", hits.Load())
	}
}

func TestS3CredentialAdapterRevocationIsImmediate(t *testing.T) {
	var hits atomic.Int64
	server := newCredentialTestServer(t, true, &hits)
	defer server.Close()
	resolver := &fakeResolver{}
	adapter := S3CredentialAdapter{
		S3Adapter:  S3Adapter{HTTPClient: server.Client()},
		Resolver:   resolver,
		Consumer:   "template-registry",
		Capability: "fetch",
	}
	profile := promptrepo.RepositoryProfile{
		Source:        "s3://prompt-bucket/catalog?endpoint=" + url.QueryEscape(server.URL) + "&path_style=true",
		CredentialRef: "yeisme-credential://prompt/s3",
	}
	if _, err := adapter.Sync(context.Background(), profile, ""); err != nil {
		t.Fatal(err)
	}
	resolver.deny.Store(true)
	_, err := adapter.Sync(context.Background(), profile, "")
	if promptrepo.ErrorCode(err) != promptrepo.CodeAuthRequired {
		t.Fatalf("revoked grant must deny the next resolution, got %v", err)
	}
	if strings.Contains(err.Error(), sentinelCredential) {
		t.Fatalf("revocation error leaks sentinel: %v", err)
	}
}

func TestS3CredentialAdapterWithoutRefStaysUnsigned(t *testing.T) {
	var hits atomic.Int64
	server := newCredentialTestServer(t, false, &hits)
	defer server.Close()
	resolver := &fakeResolver{}
	adapter := S3CredentialAdapter{
		S3Adapter:  S3Adapter{HTTPClient: server.Client()},
		Resolver:   resolver,
		Consumer:   "template-registry",
		Capability: "fetch",
	}
	profile := promptrepo.RepositoryProfile{Source: "s3://prompt-bucket/catalog?endpoint=" + url.QueryEscape(server.URL) + "&path_style=true"}
	result, err := adapter.Sync(context.Background(), profile, "")
	if err != nil || result.Catalog.Repository.ID != "fixture" {
		t.Fatalf("unsigned sync: %+v, %v", result, err)
	}
	if resolver.calls.Load() != 0 {
		t.Fatalf("profile without CredentialRef must not resolve, got %d calls", resolver.calls.Load())
	}
}
