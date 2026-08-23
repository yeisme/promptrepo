package promptrepo_test

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/conformance"
	"github.com/yeisme/promptrepo/engine"
)

const privateV010CatalogDigest = "sha256:cbc117e93a2e904b01ba1494ef7e4e2e82a100d9e1e5e739cf035c4aa53858fc"

// TestPrivateV010FixtureCompatibility freezes observable behavior from the
// private sdk/go/promptrepo/v0.1.0 tag. It is entirely local so public CI never
// needs access to the private repository.
func TestPrivateV010FixtureCompatibility(t *testing.T) {
	t.Parallel()

	repositoryRoot := t.TempDir()
	conformance.WriteCatalog(t, repositoryRoot, conformance.MinimalCatalog())
	manager, err := engine.New(engine.Options{
		ConfigRoot: filepath.Join(t.TempDir(), "config"),
		CacheRoot:  filepath.Join(t.TempDir(), "cache"),
		Now:        func() time.Time { return time.Unix(100, 0).UTC() },
	})
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	view, err := manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{
		ID: "fixture", Source: (&url.URL{Scheme: "file", Path: repositoryRoot}).String(), Trust: "official",
	}})
	if err != nil {
		t.Fatal(err)
	}
	if view.Profile.SourceKind != "file" || view.Profile.Scope != "user" || !view.Profile.Enabled {
		t.Fatalf("private v0.1.0 profile defaults changed: %+v", view.Profile)
	}

	syncReceipt, err := manager.SyncRepositories(ctx, promptrepo.SyncRequest{All: true})
	if err != nil {
		t.Fatal(err)
	}
	if syncReceipt.SchemaVersion != "promptrepo.sync_receipt.v0.1" || len(syncReceipt.Results) != 1 {
		t.Fatalf("private v0.1.0 sync receipt changed: %+v", syncReceipt)
	}
	if got := syncReceipt.Results[0].Digest; got != privateV010CatalogDigest {
		t.Fatalf("private v0.1.0 catalog digest changed: got %s", got)
	}

	search, err := manager.Search(ctx, promptrepo.SearchRequest{Query: "播客配音", Locale: "zh-CN", RequiredCapabilities: []string{"audio", "voice", "tts"}})
	if err != nil {
		t.Fatal(err)
	}
	const privateV010Ref = "promptrepo://fixture/audio/podcast-narration@1.0.0?locale=zh-CN"
	if len(search.Results) != 1 || search.Results[0].Ref != privateV010Ref || search.Results[0].Score != 30 || !search.Results[0].Compatible {
		t.Fatalf("private v0.1.0 search changed: %+v", search)
	}

	resolved, err := manager.Resolve(ctx, promptrepo.ResolveRequest{Ref: privateV010Ref, RequiredCapabilities: []string{"audio", "voice", "tts"}})
	if err != nil {
		t.Fatal(err)
	}
	if resolved.Snapshot.Digest != privateV010CatalogDigest || resolved.Locale != "zh-CN" {
		t.Fatalf("private v0.1.0 resolve changed: %+v", resolved)
	}
	content, err := manager.ReadTemplate(ctx, promptrepo.ReadTemplateRequest{Ref: resolved.Ref, Role: "main"})
	if err != nil {
		t.Fatal(err)
	}
	if content.Body != conformance.MinimalTemplateBody {
		t.Fatalf("private v0.1.0 template read changed: %+v", content)
	}

	stage, err := manager.Stage(ctx, promptrepo.StageRequest{ResolveRequest: promptrepo.ResolveRequest{Ref: resolved.Ref, RequiredCapabilities: []string{"audio", "voice", "tts"}}, Consumer: "compatibility-test"})
	if err != nil {
		t.Fatal(err)
	}
	if stage.State != "staged_pending_review" || stage.ProviderCalls != "disabled" || stage.SnapshotDigest != privateV010CatalogDigest {
		t.Fatalf("private v0.1.0 stage changed: %+v", stage)
	}

	if _, err := manager.Resolve(ctx, promptrepo.ResolveRequest{Ref: "promptrepo://fixture/audio/podcast-narration"}); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
		t.Fatalf("private v0.1.0 invalid exact ref error changed: %v", err)
	}
}
