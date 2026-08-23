package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
)

func TestLocalAdapterReadsCatalog(t *testing.T) {
	root := t.TempDir()
	catalog := promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "fixture", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(1, 0).UTC(),
		Solutions:     []promptrepo.Solution{{PackageID: "audio", ID: "podcast", Version: "1.0.0", Category: "audio", Rights: "internal", Maturity: "first-support", Locales: map[string]promptrepo.LocalizedText{"zh-CN": {Title: "中文播客旁白", Summary: "生成自然旁白"}}}},
	}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	payload, _ := json.Marshal(catalog)
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	result, err := (LocalAdapter{}).Sync(context.Background(), promptrepo.RepositoryProfile{Source: (&url.URL{Scheme: "file", Path: root}).String()}, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Catalog.Digest != digest || len(result.Catalog.Solutions) != 1 {
		t.Fatalf("result: %+v", result)
	}
}

func TestGitHubURIValidation(t *testing.T) {
	remote, err := normalizeGitRemote("github://yeisme/prompt-templates")
	if err != nil {
		t.Fatal(err)
	}
	if remote != "https://github.com/yeisme/prompt-templates.git" {
		t.Fatalf("remote: %s", remote)
	}
}

func TestGitAdapterPinsExactCommit(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	repositoryRoot := t.TempDir()
	writeTestCatalog(t, repositoryRoot)
	runTestGit(t, repositoryRoot, "init", "-b", "main")
	runTestGit(t, repositoryRoot, "config", "user.name", "Prompt Repo Test")
	runTestGit(t, repositoryRoot, "config", "user.email", "promptrepo@example.invalid")
	runTestGit(t, repositoryRoot, "add", ".")
	runTestGit(t, repositoryRoot, "commit", "-m", "catalog fixture")
	cacheRoot := t.TempDir()
	profile := promptrepo.RepositoryProfile{Source: "git+" + (&url.URL{Scheme: "file", Path: repositoryRoot}).String(), Revision: "main"}
	result, err := (GitAdapter{}).Sync(context.Background(), profile, cacheRoot)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Revision) != 40 || result.Catalog.Repository.ID != "fixture" {
		t.Fatalf("result: %+v", result)
	}
	payload, err := (GitAdapter{}).ReadTemplate(context.Background(), profile, promptrepo.SnapshotMetadata{Revision: result.Revision}, result.Catalog.Solutions[0].Templates[0], cacheRoot)
	if err != nil || string(payload) != testTemplateBody {
		t.Fatalf("template: %q, %v", payload, err)
	}
}

func TestS3AdapterReadsPathStyleCatalog(t *testing.T) {
	catalog := testCatalog(t)
	payload, _ := json.Marshal(catalog)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		switch request.URL.Path {
		case "/prompt-bucket/catalog/catalog.json":
			writer.Header().Set("ETag", `"fixture-etag"`)
			_, _ = writer.Write(payload)
		case "/prompt-bucket/catalog/prompts/main.zh-CN.md":
			_, _ = writer.Write([]byte(testTemplateBody))
		default:
			t.Fatalf("path: %s", request.URL.Path)
		}
	}))
	defer server.Close()
	sourceURI := "s3://prompt-bucket/catalog?endpoint=" + url.QueryEscape(server.URL) + "&path_style=true"
	result, err := (S3Adapter{HTTPClient: server.Client()}).Sync(context.Background(), promptrepo.RepositoryProfile{Source: sourceURI}, "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Revision != "s3:fixture-etag" || result.Catalog.Repository.ID != "fixture" {
		t.Fatalf("result: %+v", result)
	}
	templatePayload, err := (S3Adapter{HTTPClient: server.Client()}).ReadTemplate(context.Background(), promptrepo.RepositoryProfile{Source: sourceURI}, promptrepo.SnapshotMetadata{Revision: result.Revision}, result.Catalog.Solutions[0].Templates[0], "")
	if err != nil || string(templatePayload) != testTemplateBody {
		t.Fatalf("template: %q, %v", templatePayload, err)
	}
}

func writeTestCatalog(t *testing.T, root string) {
	t.Helper()
	catalog := testCatalog(t)
	if err := os.MkdirAll(filepath.Join(root, "prompts"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "prompts", "main.zh-CN.md"), []byte(testTemplateBody), 0o600); err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(catalog)
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func testCatalog(t *testing.T) promptrepo.Catalog {
	t.Helper()
	templateDigest := sha256.Sum256([]byte(testTemplateBody))
	catalog := promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "fixture", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(1, 0).UTC(),
		Solutions:     []promptrepo.Solution{{PackageID: "audio", ID: "podcast", Version: "1.0.0", Category: "audio", Rights: "internal", Maturity: "first-support", Locales: map[string]promptrepo.LocalizedText{"zh-CN": {Title: "中文播客旁白", Summary: "生成自然旁白"}}, Templates: []promptrepo.TemplateRole{{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.md", Digest: "sha256:" + hex.EncodeToString(templateDigest[:])}}}},
	}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	return catalog
}

const testTemplateBody = "# 中文播客旁白\n\n请生成自然旁白。\n"

func runTestGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
}
