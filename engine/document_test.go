package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
)

func TestManagerLoadsAndSelectsStructuredDocument(t *testing.T) {
	repositoryRoot := t.TempDir()
	body := []byte(`{"task":{"id":"TURNAROUND","views":[{"name":"front"},{"name":"back"}]},"secret":"ENGINE_BODY_SENTINEL_41ac"}`)
	writeStructuredDocumentRepository(t, repositoryRoot, body, true, "")
	manager := newStructuredDocumentManager(t, repositoryRoot)
	ctx := context.Background()
	ref := "promptrepo://official/drama/character@1.0.0?locale=zh-CN"

	resolved, err := manager.ResolveDocumentDescriptor(ctx, promptrepo.ResolveDocumentDescriptorRequest{Ref: ref, Role: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Path != "contracts/documents/main.zh-CN.document.json" || resolved.Document.Format != promptrepo.DocumentFormatJSON {
		t.Fatalf("resolved descriptor = %+v", resolved)
	}
	loaded, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: ref, Role: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Ready || loaded.CanonicalDigest == "" || loaded.SourceDigest != resolved.Document.SourceDigest {
		t.Fatalf("loaded document = %+v", loaded)
	}
	projection, err := json.Marshal(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(projection), "ENGINE_BODY_SENTINEL_41ac") {
		t.Fatalf("loaded projection leaked body: %s", projection)
	}

	selected, err := manager.SelectDocument(ctx, promptrepo.SelectDocumentRequest{Ref: ref, Role: "main", Selector: "json-pointer:/task/views/1"})
	if err != nil {
		t.Fatal(err)
	}
	value, ok := selected.Value.(map[string]any)
	if !ok || value["name"] != "back" {
		t.Fatalf("selected value = %#v", selected.Value)
	}

	address := "promptrepo://official/drama/character@1.0.0?kind=template&locale=zh-CN&role=main&selector=json-pointer%3A%2Ftask%2Fid"
	if _, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: address}); promptrepo.ErrorCode(err) != promptrepo.CodeSelectorUnsupported {
		t.Fatalf("LoadDocument selector error = %v", err)
	}
	selected, err = manager.SelectDocument(ctx, promptrepo.SelectDocumentRequest{Ref: address})
	if err != nil {
		t.Fatal(err)
	}
	if selected.Value != "TURNAROUND" {
		t.Fatalf("address selector value = %#v", selected.Value)
	}
}

func TestManagerRejectsMissingAndStaleDocumentDescriptor(t *testing.T) {
	ctx := context.Background()
	body := []byte(`{"task":{"id":"A"}}`)

	t.Run("missing", func(t *testing.T) {
		repositoryRoot := t.TempDir()
		writeStructuredDocumentRepository(t, repositoryRoot, body, false, "")
		manager := newStructuredDocumentManager(t, repositoryRoot)
		_, err := manager.ResolveDocumentDescriptor(ctx, promptrepo.ResolveDocumentDescriptorRequest{Ref: "promptrepo://official/drama/character@1.0.0?locale=zh-CN", Role: "main"})
		if promptrepo.ErrorCode(err) != promptrepo.CodeDocumentDescriptorMissing {
			t.Fatalf("missing descriptor error = %v", err)
		}
	})

	t.Run("stale", func(t *testing.T) {
		repositoryRoot := t.TempDir()
		writeStructuredDocumentRepository(t, repositoryRoot, body, true, "sha256:0000000000000000000000000000000000000000000000000000000000000000")
		manager := newStructuredDocumentManager(t, repositoryRoot)
		_, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: "promptrepo://official/drama/character@1.0.0?locale=zh-CN", Role: "main"})
		if promptrepo.ErrorCode(err) != promptrepo.CodeDigestMismatch {
			t.Fatalf("stale descriptor error = %v", err)
		}
	})
}

func newStructuredDocumentManager(t *testing.T, repositoryRoot string) *Manager {
	t.Helper()
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String(), Trust: "official"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	return manager
}

func writeStructuredDocumentRepository(t *testing.T, root string, body []byte, withDescriptor bool, descriptorSourceDigest string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "main.zh-CN.json"), body, 0o600); err != nil {
		t.Fatal(err)
	}
	templateDigest := digestForTest(body)
	catalog := promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "official", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(1, 0).UTC(),
		Solutions: []promptrepo.Solution{{
			PackageID: "drama", ID: "character", Version: "1.0.0", Digest: "sha256:solution", Category: "drama", Rights: "internal", Maturity: "first-support",
			Locales:   map[string]promptrepo.LocalizedText{"zh-CN": {Title: "角色资产", Summary: "结构化角色资产测试"}},
			Templates: []promptrepo.TemplateRole{{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.json", Digest: templateDigest}},
		}},
	}
	catalogDigest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = catalogDigest
	catalogPayload, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), catalogPayload, 0o600); err != nil {
		t.Fatal(err)
	}
	if !withDescriptor {
		return
	}
	if descriptorSourceDigest == "" {
		descriptorSourceDigest = templateDigest
	}
	descriptor := promptrepo.TemplateDocumentDescriptor{
		SchemaVersion: promptrepo.TemplateDocumentSchemaVersion,
		PackageID:     "drama", SolutionID: "character", Version: "1.0.0", Role: "main", Locale: "zh-CN",
		TemplatePath: "prompts/main.zh-CN.json", SourceDigest: descriptorSourceDigest,
		Format: promptrepo.DocumentFormatJSON, MediaType: "application/json", Canonicalization: promptrepo.DocumentCanonicalizationJCS,
		SelectorKinds: []string{"json-pointer"}, Limits: promptrepo.DocumentLimits{MaxBytes: 64 << 10, MaxDepth: 32},
	}
	descriptorDigest, err := promptrepo.CanonicalTemplateDocumentDescriptorDigest(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	descriptor.Digest = descriptorDigest
	descriptorPayload, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts", "documents"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "contracts", "documents", "main.zh-CN.document.json"), descriptorPayload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func digestForTest(payload []byte) string {
	digest := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(digest[:])
}
