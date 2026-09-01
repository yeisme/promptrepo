package engine

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
)

const graphKitBodySentinel = "GRAPH_KIT_BODY_SENTINEL_73d1"

type graphKitFixtureChild struct {
	Role     string `json:"role"`
	Path     string `json:"path"`
	Digest   string `json:"digest"`
	Selector string `json:"selector"`
	Expected string `json:"expected"`
}

type graphKitFixtureManifest struct {
	SchemaVersion string                 `json:"schema_version"`
	Children      []graphKitFixtureChild `json:"children"`
}

func TestGraphKitConformancePinsClosureToOneGitSnapshot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	repositoryRoot := t.TempDir()
	writeGraphKitRepository(t, repositoryRoot, "", false)
	commit := commitGraphKitRepository(t, repositoryRoot)
	manager := newGraphKitGitManager(t, repositoryRoot)
	ctx := context.Background()
	ref := "promptrepo://official/knowledge/graph-kit@1.0.0?locale=zh-CN"

	loadedManifest, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: ref, Role: "manifest"})
	if err != nil {
		t.Fatal(err)
	}
	if loadedManifest.Snapshot.Revision != commit || !regexp.MustCompile(`^[a-f0-9]{40,64}$`).MatchString(loadedManifest.Snapshot.Revision) {
		t.Fatalf("manifest snapshot is not the exact committed revision: %+v", loadedManifest.Snapshot)
	}
	if loadedManifest.Snapshot.Digest == "" || loadedManifest.CanonicalDigest == "" || !loadedManifest.Ready {
		t.Fatalf("manifest is not digest-ready: %+v", loadedManifest)
	}
	projection, err := json.Marshal(loadedManifest)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(projection), graphKitBodySentinel) {
		t.Fatalf("manifest projection leaked the structured body: %s", projection)
	}

	var manifest graphKitFixtureManifest
	if err := json.Unmarshal(loadedManifest.Body, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.SchemaVersion != "promptrepo.graph-kit-conformance.v1" || len(manifest.Children) != 3 {
		t.Fatalf("manifest closure = %+v", manifest)
	}
	for _, child := range manifest.Children {
		address := promptrepo.FormatTemplateAddress(promptrepo.TemplateAddress{
			RepositoryID: "official",
			PackageID:    "knowledge",
			SolutionID:   "graph-kit",
			Version:      "1.0.0",
			Kind:         "template",
			Locale:       "zh-CN",
			Role:         child.Role,
			Path:         child.Path,
			Digest:       child.Digest,
			Snapshot:     loadedManifest.Snapshot.Digest,
		})
		resolved, err := manager.ResolveDocumentDescriptor(ctx, promptrepo.ResolveDocumentDescriptorRequest{Ref: address})
		if err != nil {
			t.Fatalf("resolve %s: %v", child.Role, err)
		}
		loaded, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: address})
		if err != nil {
			t.Fatalf("load %s: %v", child.Role, err)
		}
		if resolved.Snapshot != loadedManifest.Snapshot || loaded.Snapshot != loadedManifest.Snapshot {
			t.Fatalf("%s escaped the manifest snapshot: resolved=%+v loaded=%+v manifest=%+v", child.Role, resolved.Snapshot, loaded.Snapshot, loadedManifest.Snapshot)
		}
		if loaded.SourceDigest != child.Digest || loaded.CanonicalDigest == "" || !loaded.Ready {
			t.Fatalf("%s digest lineage = %+v", child.Role, loaded)
		}
		selected, err := manager.SelectDocument(ctx, promptrepo.SelectDocumentRequest{Ref: address, Selector: child.Selector})
		if err != nil {
			t.Fatalf("select %s: %v", child.Role, err)
		}
		if selected.Value != child.Expected || selected.Document.Snapshot != loadedManifest.Snapshot || selected.CanonicalDigest == "" || !selected.Ready {
			t.Fatalf("%s selected projection = %+v", child.Role, selected)
		}
	}
}

func TestGraphKitConformanceFailsClosed(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not available")
	}
	ctx := context.Background()

	t.Run("missing child", func(t *testing.T) {
		repositoryRoot := t.TempDir()
		writeGraphKitRepository(t, repositoryRoot, "missing-child", false)
		commitGraphKitRepository(t, repositoryRoot)
		manager := newGraphKitGitManager(t, repositoryRoot)
		loadedManifest, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: "promptrepo://official/knowledge/graph-kit@1.0.0?locale=zh-CN", Role: "manifest"})
		if err != nil {
			t.Fatal(err)
		}
		var manifest graphKitFixtureManifest
		if err := json.Unmarshal(loadedManifest.Body, &manifest); err != nil {
			t.Fatal(err)
		}
		missing := manifest.Children[len(manifest.Children)-1]
		_, err = manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: "promptrepo://official/knowledge/graph-kit@1.0.0?locale=zh-CN", Role: missing.Role})
		if promptrepo.ErrorCode(err) != promptrepo.CodeNotFound {
			t.Fatalf("missing child error = %v", err)
		}
	})

	t.Run("child digest drift", func(t *testing.T) {
		repositoryRoot := t.TempDir()
		writeGraphKitRepository(t, repositoryRoot, "", true)
		commitGraphKitRepository(t, repositoryRoot)
		manager := newGraphKitGitManager(t, repositoryRoot)
		_, err := manager.LoadDocument(ctx, promptrepo.LoadDocumentRequest{Ref: "promptrepo://official/knowledge/graph-kit@1.0.0?locale=zh-CN", Role: "lens"})
		if promptrepo.ErrorCode(err) != promptrepo.CodeDigestMismatch {
			t.Fatalf("digest drift error = %v", err)
		}
	})
}

func newGraphKitGitManager(t *testing.T, repositoryRoot string) *Manager {
	t.Helper()
	manager, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	_, err = manager.AddRepository(context.Background(), promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{
		ID: "official", Source: "git+" + (&url.URL{Scheme: "file", Path: repositoryRoot}).String(), Revision: "main", Trust: "official",
	}})
	if err != nil {
		t.Fatal(err)
	}
	receipt, err := manager.SyncRepositories(context.Background(), promptrepo.SyncRequest{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Results) != 1 || receipt.Results[0].State != "ready" || !regexp.MustCompile(`^[a-f0-9]{40,64}$`).MatchString(receipt.Results[0].Revision) {
		t.Fatalf("Git sync did not return an exact ready snapshot: %+v", receipt)
	}
	return manager
}

func writeGraphKitRepository(t *testing.T, root, extraChildRole string, driftLens bool) {
	t.Helper()
	roles := []string{"lens", "validator", "view"}
	bodies := map[string][]byte{
		"lens":      []byte(`{"identity":{"id":"synthetic-lens-v1"},"rules":["stable-id"]}`),
		"validator": []byte(`{"checks":["schema","digest"],"identity":{"id":"synthetic-validator-v1"}}`),
		"view":      []byte(`{"identity":{"id":"synthetic-view-v1"},"panels":["nodes","relations"]}`),
	}
	children := make([]graphKitFixtureChild, 0, len(roles)+1)
	for _, role := range roles {
		children = append(children, graphKitFixtureChild{
			Role: role, Path: "graph-kit/prompts/" + role + ".zh-CN.json", Digest: graphKitDigest(bodies[role]),
			Selector: "json-pointer:/identity/id", Expected: "synthetic-" + role + "-v1",
		})
	}
	if extraChildRole != "" {
		children = append(children, graphKitFixtureChild{
			Role: extraChildRole, Path: "graph-kit/prompts/" + extraChildRole + ".zh-CN.json",
			Digest: graphKitDigest([]byte(`{"identity":{"id":"missing"}}`)), Selector: "json-pointer:/identity/id", Expected: "missing",
		})
	}
	// Keep the sentinel in a valid JSON field so serialization redaction is
	// exercised without making the fixture body invalid.
	manifestBody, err := json.Marshal(map[string]any{
		"body_sentinel":  graphKitBodySentinel,
		"children":       children,
		"schema_version": "promptrepo.graph-kit-conformance.v1",
	})
	if err != nil {
		t.Fatal(err)
	}

	allBodies := map[string][]byte{"manifest": manifestBody}
	for role, body := range bodies {
		allBodies[role] = body
	}
	roles = append([]string{"manifest"}, roles...)
	templates := make([]promptrepo.TemplateRole, 0, len(roles))
	for _, role := range roles {
		body := allBodies[role]
		path := "graph-kit/prompts/" + role + ".zh-CN.json"
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, path), body, 0o600); err != nil {
			t.Fatal(err)
		}
		template := promptrepo.TemplateRole{Role: role, Locale: "zh-CN", Path: path, Digest: graphKitDigest(body)}
		templates = append(templates, template)
		writeGraphKitDescriptor(t, root, template)
	}
	if driftLens {
		if err := os.WriteFile(filepath.Join(root, "graph-kit", "prompts", "lens.zh-CN.json"), []byte(`{"identity":{"id":"tampered-lens"}}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	catalog := promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "official", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(1, 0).UTC(),
		Solutions: []promptrepo.Solution{{
			PackageID: "knowledge", ID: "graph-kit", Version: "1.0.0", Digest: "sha256:synthetic-solution",
			Category: "knowledge", Rights: "internal", Maturity: "first-support",
			Locales:   map[string]promptrepo.LocalizedText{"zh-CN": {Title: "Synthetic Graph Kit", Summary: "Cross-domain structured document closure fixture"}},
			Templates: templates,
		}},
	}
	catalogDigest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = catalogDigest
	payload, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "catalog.json"), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func writeGraphKitDescriptor(t *testing.T, root string, template promptrepo.TemplateRole) {
	t.Helper()
	descriptor := promptrepo.TemplateDocumentDescriptor{
		SchemaVersion: promptrepo.TemplateDocumentSchemaVersion,
		PackageID:     "knowledge", SolutionID: "graph-kit", Version: "1.0.0", Role: template.Role, Locale: template.Locale,
		TemplatePath: template.Path, SourceDigest: template.Digest,
		Format: promptrepo.DocumentFormatJSON, MediaType: "application/json", Canonicalization: promptrepo.DocumentCanonicalizationJCS,
		SelectorKinds: []string{"json-pointer"}, Limits: promptrepo.DocumentLimits{MaxBytes: 64 << 10, MaxDepth: 32},
	}
	digest, err := promptrepo.CanonicalTemplateDocumentDescriptorDigest(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	descriptor.Digest = digest
	payload, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	path, err := promptrepo.TemplateDocumentDescriptorPath(template)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(filepath.Join(root, path)), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, path), payload, 0o600); err != nil {
		t.Fatal(err)
	}
}

func commitGraphKitRepository(t *testing.T, root string) string {
	t.Helper()
	runGraphKitGit(t, root, "init", "-b", "main")
	runGraphKitGit(t, root, "config", "user.name", "Promptrepo Graph Kit Test")
	runGraphKitGit(t, root, "config", "user.email", "promptrepo-graph-kit@example.invalid")
	runGraphKitGit(t, root, "add", ".")
	runGraphKitGit(t, root, "commit", "-m", "graph kit conformance fixture")
	return strings.TrimSpace(runGraphKitGit(t, root, "rev-parse", "HEAD"))
}

func runGraphKitGit(t *testing.T, root string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = root
	payload, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, payload)
	}
	return string(payload)
}

func graphKitDigest(payload []byte) string {
	sum := sha256.Sum256(payload)
	return "sha256:" + hex.EncodeToString(sum[:])
}
