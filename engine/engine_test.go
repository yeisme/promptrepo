package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
)

func TestRepositorySearchResolveAndStage(t *testing.T) {
	repositoryRoot := t.TempDir()
	writeFixtureCatalog(t, repositoryRoot)
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache"), Now: func() time.Time { return time.Unix(100, 0).UTC() }})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String(), Trust: "official"}})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Results) != 1 || receipt.Results[0].State != "ready" {
		t.Fatalf("sync: %+v", receipt)
	}
	search, err := manager.Search(ctx, promptrepo.SearchRequest{Query: "中文播客旁白", Locale: "zh-CN", RequiredCapabilities: []string{"audio", "voice", "tts"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(search.Results) != 1 || !search.Results[0].Compatible {
		t.Fatalf("search: %+v", search)
	}
	resolved, err := manager.Resolve(ctx, promptrepo.ResolveRequest{Ref: search.Results[0].Ref, RequiredCapabilities: []string{"audio", "voice", "tts"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Display.Title != "中文播客旁白" {
		t.Fatalf("resolve: %+v", resolved)
	}
	content, err := manager.ReadTemplate(ctx, promptrepo.ReadTemplateRequest{Ref: resolved.Ref, Role: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if content.Body != fixtureTemplateBody || content.Digest == "" {
		t.Fatalf("template: %+v", content)
	}
	stage, err := manager.Stage(ctx, promptrepo.StageRequest{ResolveRequest: promptrepo.ResolveRequest{Ref: resolved.Ref, RequiredCapabilities: []string{"audio", "voice", "tts"}}, Consumer: "sonora"})
	if err != nil {
		t.Fatal(err)
	}
	if stage.State != "staged_pending_review" || stage.ProviderCalls != "disabled" {
		t.Fatalf("stage: %+v", stage)
	}
}

func TestFutureStateSchemaFailsClosed(t *testing.T) {
	configRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(configRoot, "state.json"), []byte(`{"schema_version":"promptrepo.state.v9.0"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	manager, err := New(Options{ConfigRoot: configRoot, CacheRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.ListRepositories(context.Background(), promptrepo.ListRepositoriesRequest{})
	if promptrepo.ErrorCode(err) != promptrepo.CodeStateSchemaTooNew {
		t.Fatalf("error: %v", err)
	}
}

func TestStaleWriterLockIsRecovered(t *testing.T) {
	configRoot := t.TempDir()
	lockPath := filepath.Join(configRoot, "state.lock")
	if err := os.WriteFile(lockPath, []byte("abandoned"), 0o600); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(lockPath, old, old); err != nil {
		t.Fatal(err)
	}
	manager, err := New(Options{ConfigRoot: configRoot, CacheRoot: t.TempDir(), StaleLockAge: time.Minute})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.AddRepository(context.Background(), promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "local", Source: "file:///tmp/prompts"}})
	if err != nil {
		t.Fatal(err)
	}
}

func writeFixtureCatalog(t *testing.T, root string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "main.zh-CN.md"), []byte(fixtureTemplateBody), 0o600); err != nil {
		t.Fatal(err)
	}
	templateDigest := sha256.Sum256([]byte(fixtureTemplateBody))
	catalog := promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "official", Name: "Yeisme Prompt Templates", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(1, 0).UTC(),
		Solutions:     []promptrepo.Solution{{PackageID: "audio", ID: "podcast-narration", Version: "1.0.0", Category: "audio", Tags: []string{"job:generate", "artifact:voiceover"}, Capabilities: []string{"audio", "voice", "tts"}, Rights: "internal", Maturity: "first-support", Digest: "sha256:solution", Locales: map[string]promptrepo.LocalizedText{"zh-CN": {Title: "中文播客旁白", Summary: "生成自然、清晰的中文播客旁白", Aliases: []string{"播客配音"}}, "en": {Title: "Chinese podcast narration", Summary: "Generate natural Chinese podcast narration"}}, Templates: []promptrepo.TemplateRole{{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.md", Digest: "sha256:" + hex.EncodeToString(templateDigest[:])}}}},
	}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	payload, _ := json.MarshalIndent(catalog, "", "  ")
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
}

const fixtureTemplateBody = "# 中文播客旁白\n\n请生成自然、清晰的旁白。\n"
