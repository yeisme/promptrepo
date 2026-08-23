package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func TestInspectAndPreviewAreSafeAndNonExecuting(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello {{subject}} {{tone}} {{count}} {{published}}")
	configRoot := filepath.Join(t.TempDir(), "config")
	manager, err := New(Options{ConfigRoot: configRoot, CacheRoot: filepath.Join(t.TempDir(), "cache")})
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
	ref := "promptrepo://official/audio/preview@1.0.0?locale=zh-CN"
	contract := previewContract()
	inspect, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: ref, Contract: contract, Values: map[string]any{"count": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if inspect.Ready || len(inspect.Issues) != 1 || inspect.Issues[0].Code != promptrepo.CodeInputRequired || inspect.Inputs[0].Status != "missing" {
		t.Fatalf("inspect readiness: ready=%t issue_count=%d first_code=%q first_status=%q", inspect.Ready, len(inspect.Issues), inspect.Issues[0].Code, inspect.Inputs[0].Status)
	}
	before, err := os.ReadFile(filepath.Join(configRoot, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	preview, err := manager.Preview(ctx, promptrepo.PreviewRequest{Ref: ref, Contract: contract, Values: map[string]any{"subject": "Ada", "count": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if !preview.Ready || preview.RenderedBody != "Hello Ada calm 2 true" || preview.RenderedDigest == "" || preview.RenderedBytes != len(preview.RenderedBody) || preview.RenderedRunes != len([]rune(preview.RenderedBody)) {
		t.Fatalf("preview metadata is invalid: ready=%t digest=%q bytes=%d runes=%d", preview.Ready, preview.RenderedDigest, preview.RenderedBytes, preview.RenderedRunes)
	}
	if preview.ProviderCalls || preview.StateWrites || preview.UsageRecorded {
		t.Fatalf("preview side effect flags are not all false")
	}
	payload, err := json.Marshal(preview)
	if err != nil {
		t.Fatal(err)
	}
	if string(payload) == "" || containsString(string(payload), preview.RenderedBody) || containsString(string(payload), "Ada") {
		t.Fatalf("preview body or supplied values leaked: %s", payload)
	}
	after, err := os.ReadFile(filepath.Join(configRoot, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatal("preview changed persisted state")
	}
	for _, values := range []map[string]any{
		{"subject": "Ada", "count": "two"},
		{"subject": "Ada", "tone": "loud", "count": 2},
		{"subject": "Ada", "count": 2, "unknown": "x"},
	} {
		result, err := manager.Preview(ctx, promptrepo.PreviewRequest{Ref: ref, Contract: contract, Values: values})
		if err != nil || result.Ready || len(result.Issues) == 0 {
			t.Fatalf("expected validation issue: ready=%t issues=%d err=%v", result.Ready, len(result.Issues), err)
		}
	}
}

func TestResolveTemplateContractCompanion(t *testing.T) {
	repositoryRoot := t.TempDir()
	body := "Hello {{subject}} {{tone}} {{count}} {{published}}"
	writePreviewCatalog(t, repositoryRoot, body)
	writePreviewContractCompanion(t, repositoryRoot, body)
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
	ref := "promptrepo://official/audio/preview@1.0.0?locale=zh-CN"
	resolved, err := manager.ResolveTemplateContract(ctx, promptrepo.ResolveTemplateContractRequest{Ref: ref, Role: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Path != "contracts/main.zh-CN.json" || resolved.Consistency != promptrepo.ContractConsistencyContentBound || resolved.Contract.Digest == "" || len(resolved.Contract.Inputs) != 4 {
		t.Fatalf("resolved companion: %+v", resolved)
	}
	preview, err := manager.Preview(ctx, promptrepo.PreviewRequest{Ref: ref, Role: "main", Contract: resolved.Contract, Values: map[string]any{"subject": "Ada", "count": 2}})
	if err != nil || !preview.Ready || preview.RenderedBody != "Hello Ada calm 2 true" {
		t.Fatalf("preview from companion: ready=%t body=%q err=%v", preview.Ready, preview.RenderedBody, err)
	}
}

func TestContractConsistencyMatchesSourceGuarantee(t *testing.T) {
	for sourceKind, want := range map[string]string{
		"git":  promptrepo.ContractConsistencySnapshotPinned,
		"file": promptrepo.ContractConsistencyContentBound,
		"s3":   promptrepo.ContractConsistencyContentBound,
	} {
		if got := contractConsistency(sourceKind); got != want {
			t.Fatalf("source %q consistency = %q, want %q", sourceKind, got, want)
		}
	}
}

func TestPreviewRejectsUndeclaredPlaceholders(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello {{subject}} {{undeclared}}")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Preview(ctx, promptrepo.PreviewRequest{Ref: "promptrepo://official/audio/preview@1.0.0?locale=zh-CN", Contract: previewContract(), Values: map[string]any{"subject": "Ada", "count": 2}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Ready || len(result.Issues) != 1 || result.Issues[0].Code != promptrepo.CodeTemplatePlaceholder {
		t.Fatalf("placeholder validation: ready=%t issue_count=%d", result.Ready, len(result.Issues))
	}
}

func TestInspectDoesNotReadTemplateBody(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello {{subject}}")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(repositoryRoot, "prompts", "main.zh-CN.md")); err != nil {
		t.Fatal(err)
	}
	result, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: "promptrepo://official/audio/preview@1.0.0?locale=zh-CN", Contract: previewContract(), Values: map[string]any{"subject": "Ada", "count": 2}})
	if err != nil || !result.Ready {
		t.Fatalf("inspect should use only catalog metadata: ready=%t err=%v", result.Ready, err)
	}
}

func TestInspectRightsBlockedDoesNotSuggestSupplyingInputs(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalogWithRights(t, repositoryRoot, "Hello {{subject}}", "blocked")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	contract := promptrepo.TemplateContract{Inputs: []promptrepo.InputDefinition{{Name: "subject", Type: promptrepo.InputTypeString, Required: true}}}
	result, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: "promptrepo://official/audio/preview@1.0.0?locale=zh-CN", Contract: contract, Values: map[string]any{"subject": "Ada"}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Ready || result.NextAction.Kind != "blocked" || len(result.NextAction.RequiredInputs) != 0 || len(result.Issues) != 1 || result.Issues[0].Code != promptrepo.CodeRightsBlocked {
		t.Fatalf("blocked rights projection: ready=%t next=%+v issues=%+v", result.Ready, result.NextAction, result.Issues)
	}
}

func TestAddressQualifiersAndSelectorsFailClosed(t *testing.T) {
	for _, raw := range []string{
		"promptrepo://official/audio/preview@1.0.0?locale=zh-CN&digest=sha256%3A0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		"promptrepo://official/audio/preview@1.0.0?locale=zh-CN&path=prompts%2Fmain.zh-CN.md",
		"promptrepo://official/audio/preview@1.0.0?locale=zh-CN&selector=heading%3Atitle",
	} {
		if _, _, err := parseAddressIfPresent(raw); err == nil {
			t.Fatalf("address qualifier bypassed validation: %s", raw)
		}
	}

	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello {{subject}} {{tone}} {{count}} {{published}}")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	address := "promptrepo://official/audio/preview@1.0.0?kind=template&locale=zh-CN&path=prompts%2Fmain.zh-CN.md&selector=heading%3Atitle"
	inspection, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: address, Contract: previewContract(), Values: map[string]any{"subject": "Ada", "count": 2}})
	if err != nil || !strings.Contains(inspection.Address, "selector=heading%3Atitle") {
		t.Fatalf("inspect should preserve selector metadata: err=%v address=%q", err, inspection.Address)
	}
	_, err = manager.Preview(ctx, promptrepo.PreviewRequest{Ref: address, Contract: previewContract(), Values: map[string]any{"subject": "Ada", "count": 2}})
	if promptrepo.ErrorCode(err) != promptrepo.CodeSelectorUnsupported {
		t.Fatalf("selector preview error: %v", err)
	}
}

func TestAddressPathOnlyAndRolePathMismatch(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	pathOnly := "promptrepo://official/audio/preview@1.0.0?kind=template&locale=zh-CN&path=prompts%2Fmain.zh-CN.md"
	result, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: pathOnly})
	if err != nil || result.Role != "main" {
		t.Fatalf("path-only address did not choose the unique template: role=%q err=%v", result.Role, err)
	}
	mismatch := "promptrepo://official/audio/preview@1.0.0?kind=template&locale=zh-CN&role=other&path=prompts%2Fmain.zh-CN.md"
	if _, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: mismatch}); promptrepo.ErrorCode(err) != promptrepo.CodeAddressMismatch {
		t.Fatalf("role/path mismatch: %v", err)
	}
}

func TestAddressDuplicateMatchesFailClosed(t *testing.T) {
	repositoryRoot := t.TempDir()
	writeDuplicateTemplateCatalog(t, repositoryRoot)
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	ref := "promptrepo://official/audio/preview@1.0.0?kind=template&locale=zh-CN&path=prompts%2Fmain.zh-CN.md"
	if _, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: ref}); promptrepo.ErrorCode(err) != promptrepo.CodeAddressMismatch {
		t.Fatalf("duplicate address should fail closed: %v", err)
	}
}

func TestSyncStateRemainsV020Shaped(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello")
	configRoot := filepath.Join(t.TempDir(), "config")
	manager, err := New(Options{ConfigRoot: configRoot, CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.AddRepository(context.Background(), promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(context.Background(), promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	payload, err := os.ReadFile(filepath.Join(configRoot, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"inputs"`, `"license"`, `"permissions"`} {
		if strings.Contains(string(payload), field) {
			t.Fatalf("v0.2 state gained field %s", field)
		}
	}
}

func TestManagerRejectsMalformedTemplateContractDigest(t *testing.T) {
	repositoryRoot := t.TempDir()
	writePreviewCatalog(t, repositoryRoot, "Hello")
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	_, err = manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: "official", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String()}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true}); err != nil {
		t.Fatal(err)
	}
	ref := "promptrepo://official/audio/preview@1.0.0?locale=zh-CN"
	bad := promptrepo.TemplateContract{Digest: "sha256:bad"}
	if _, err := manager.Inspect(ctx, promptrepo.InspectRequest{Ref: ref, Contract: bad}); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
		t.Fatalf("inspect contract digest: %v", err)
	}
	if _, err := manager.Validate(ctx, promptrepo.ValidateRequest{Ref: ref, Contract: bad}); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
		t.Fatalf("validate contract digest: %v", err)
	}
	if _, err := manager.Preview(ctx, promptrepo.PreviewRequest{Ref: ref, Contract: bad}); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
		t.Fatalf("preview contract digest: %v", err)
	}
}

func TestNormalizeSearchHandlesInvalidUTF8(t *testing.T) {
	value := normalizeSearch("  A\xffB  ")
	if value != "a�b" {
		t.Fatalf("normalized invalid UTF-8 = %q", value)
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

func writePreviewCatalog(t *testing.T, root, body string) {
	writePreviewCatalogWithRights(t, root, body, "internal")
}

func writePreviewCatalogWithRights(t *testing.T, root, body, rights string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "main.zh-CN.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	templateDigest := sha256.Sum256([]byte(body))
	catalog := promptrepo.Catalog{SchemaVersion: promptrepo.CatalogSchemaVersion, Repository: promptrepo.RepositoryMetadata{ID: "official", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"}, GeneratedAt: time.Unix(1, 0).UTC(), Solutions: []promptrepo.Solution{{PackageID: "audio", ID: "preview", Version: "1.0.0", Digest: "sha256:solution", Category: "audio", Rights: rights, Maturity: "first-support", Locales: map[string]promptrepo.LocalizedText{"zh-CN": {Title: "预览", Summary: "预览测试"}}, Templates: []promptrepo.TemplateRole{{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.md", Digest: "sha256:" + hex.EncodeToString(templateDigest[:])}}}}}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	payload, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func previewContract() promptrepo.TemplateContract {
	minimum, maximum := 1.0, 5.0
	return promptrepo.TemplateContract{Digest: "sha256:0000000000000000000000000000000000000000000000000000000000000000", License: "internal", Permissions: []string{"preview"}, Inputs: []promptrepo.InputDefinition{{Name: "subject", Type: promptrepo.InputTypeString, Required: true}, {Name: "tone", Type: promptrepo.InputTypeEnum, Enum: []any{"calm", "bright"}, Default: "calm"}, {Name: "count", Type: promptrepo.InputTypeInteger, Min: &minimum, Max: &maximum, Required: true}, {Name: "published", Type: promptrepo.InputTypeBoolean, Default: true}}}
}

func writePreviewContractCompanion(t *testing.T, root, body string) {
	t.Helper()
	contract := previewContract()
	inputs := append([]promptrepo.InputDefinition(nil), contract.Inputs...)
	sort.Slice(inputs, func(i, j int) bool { return inputs[i].Name < inputs[j].Name })
	templateDigest := sha256.Sum256([]byte(body))
	document := promptrepo.TemplateContractDocument{
		SchemaVersion: promptrepo.TemplateContractSchemaVersion,
		PackageID:     "audio", SolutionID: "preview", Version: "1.0.0", Role: "main", Locale: "zh-CN",
		TemplatePath: "prompts/main.zh-CN.md", TemplateDigest: "sha256:" + hex.EncodeToString(templateDigest[:]),
		License: "internal", Permissions: []string{"preview"}, Inputs: inputs,
	}
	digest, err := promptrepo.CanonicalTemplateContractDigest(document)
	if err != nil {
		t.Fatal(err)
	}
	document.Digest = digest
	payload, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "contracts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "contracts", "main.zh-CN.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeDuplicateTemplateCatalog(t *testing.T, root string) {
	t.Helper()
	body := "Hello"
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "main.zh-CN.md"), []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	templateDigest := sha256.Sum256([]byte(body))
	template := promptrepo.TemplateRole{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.md", Digest: "sha256:" + hex.EncodeToString(templateDigest[:])}
	catalog := promptrepo.Catalog{SchemaVersion: promptrepo.CatalogSchemaVersion, Repository: promptrepo.RepositoryMetadata{ID: "official", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"}, GeneratedAt: time.Unix(1, 0).UTC(), Solutions: []promptrepo.Solution{{PackageID: "audio", ID: "preview", Version: "1.0.0", Digest: "sha256:solution", Category: "audio", Rights: "internal", Maturity: "first-support", Locales: map[string]promptrepo.LocalizedText{"zh-CN": {Title: "预览", Summary: "预览测试"}}, Templates: []promptrepo.TemplateRole{template, template}}}}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	payload, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func containsString(value, part string) bool {
	return len(part) != 0 && len(value) >= len(part) && stringContains(value, part)
}

func stringContains(value, part string) bool {
	for index := 0; index+len(part) <= len(value); index++ {
		if value[index:index+len(part)] == part {
			return true
		}
	}
	return false
}
