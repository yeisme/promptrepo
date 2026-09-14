package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/source"
)

type discoveryAdapter struct {
	mu                  sync.Mutex
	calls, active, peak int
	fail                bool
	entered             chan struct{}
	release             chan struct{}
}

func (*discoveryAdapter) Kind() string { return "file" }
func (a *discoveryAdapter) Sync(ctx context.Context, p promptrepo.RepositoryProfile, _ string) (source.SyncResult, error) {
	a.mu.Lock()
	a.calls++
	a.active++
	if a.active > a.peak {
		a.peak = a.active
	}
	fail := a.fail
	a.mu.Unlock()
	defer func() { a.mu.Lock(); a.active--; a.mu.Unlock() }()
	if a.entered != nil {
		a.entered <- struct{}{}
	}
	if a.release != nil {
		select {
		case <-a.release:
		case <-ctx.Done():
			return source.SyncResult{}, ctx.Err()
		}
	}
	if fail {
		return source.SyncResult{}, promptrepo.NewError(promptrepo.CodeAuthorizationFailed, "denied", false, nil)
	}
	return source.SyncResult{Revision: p.Revision, Catalog: promptrepo.Catalog{Discovery: []promptrepo.TemplateDiscoveryAnnotation{{PackageID: "video", SolutionID: "character", Role: "front", Locale: "en", Media: []string{"image"}, Consumers: []string{"eikona"}}, {PackageID: "video", SolutionID: "character", Role: "front", Locale: "zh-CN", Media: []string{"image"}, Consumers: []string{"eikona"}}, {PackageID: "video", SolutionID: "character", Role: "motion", Locale: "en", Media: []string{"video"}}}, SchemaVersion: promptrepo.CatalogSchemaVersion, Repository: promptrepo.RepositoryMetadata{ID: p.ID, DefaultLocale: "en"}, Solutions: []promptrepo.Solution{{PackageID: "video", ID: "character", Version: "1.0.0", Category: "video", Capabilities: []string{"character_asset"}, Locales: map[string]promptrepo.LocalizedText{"en": {Title: "Character assets", Summary: "Create character references"}}, Templates: []promptrepo.TemplateRole{{Role: "front", Locale: "en", Digest: "same"}, {Role: "front", Locale: "zh-CN", Digest: "different"}, {Role: "motion", Locale: "en", Digest: "motion"}}}}}}, nil
}
func (*discoveryAdapter) ReadTemplate(context.Context, promptrepo.RepositoryProfile, promptrepo.SnapshotMetadata, promptrepo.TemplateRole, string) ([]byte, error) {
	panic("discovery must not read bodies")
}
func discoveryManager(t *testing.T, a *discoveryAdapter, now func() time.Time) *Manager {
	t.Helper()
	m, err := New(Options{ConfigRoot: filepath.Join(t.TempDir(), "config"), CacheRoot: t.TempDir(), Sources: source.NewRegistry(a), Now: now})
	if err != nil {
		t.Fatal(err)
	}
	return m
}
func addDiscoverySource(t *testing.T, m *Manager, id string) {
	t.Helper()
	_, err := m.AddRepository(context.Background(), promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{ID: id, Source: "file:///" + id, Revision: "pinned"}})
	if err != nil {
		t.Fatal(err)
	}
}
func TestDiscoverCountsFiltersPaginationAndCache(t *testing.T) {
	ctx := context.Background()
	now := time.Unix(1000, 0)
	a := &discoveryAdapter{}
	m := discoveryManager(t, a, func() time.Time { return now })
	addDiscoverySource(t, m, "official")
	addDiscoverySource(t, m, "personal")
	d, err := m.Discover(ctx, promptrepo.DiscoveryRequest{Media: "image", Consumer: "eikona", Limit: 1})
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != "success" || d.Counts.Solutions != 2 || d.Counts.Roles != 4 || d.Counts.Documents != 6 || d.Matched.Roles != 2 || d.Matched.Documents != 4 || d.Returned.Documents != 2 || d.Remaining != 1 {
		t.Fatalf("bad counts: %+v", d)
	}
	if len(d.Solutions[0].Templates[0].DuplicateOf) != 1 {
		t.Fatal("cross-source duplicate missing")
	}
	if _, err := promptrepo.ParseTemplateAddress(d.Solutions[0].Templates[0].Ref); err != nil {
		t.Fatal(err)
	}
	_, _ = m.Discover(ctx, promptrepo.DiscoveryRequest{})
	if a.calls != 2 {
		t.Fatal("fresh cache fetched")
	}
	now = now.Add(6 * time.Minute)
	a.fail = true
	stale, err := m.Discover(ctx, promptrepo.DiscoveryRequest{})
	if err != nil || stale.Status != "partial" || stale.Counts.Documents != 6 || !stale.Sources[0].Stale || stale.Sources[0].ErrorCode != promptrepo.CodeAuthorizationFailed {
		t.Fatalf("stale: %+v %v", stale, err)
	}
	_, _ = m.Discover(ctx, promptrepo.DiscoveryRequest{Offline: true})
	if a.calls != 4 {
		t.Fatal("offline made network call")
	}
	for _, req := range []promptrepo.DiscoveryRequest{{Limit: -1}, {Offset: -1}, {Offline: true, Refresh: true}, {Repository: "absent"}} {
		if _, err := m.Discover(ctx, req); err == nil {
			t.Fatalf("accepted invalid request %+v", req)
		}
	}
}
func TestDiscoverMissingDisabledAndSet(t *testing.T) {
	a := &discoveryAdapter{fail: true}
	m := discoveryManager(t, a, time.Now)
	ctx := context.Background()
	addDiscoverySource(t, m, "one")
	d, err := m.Discover(ctx, promptrepo.DiscoveryRequest{})
	if err != nil || d.Status != "failed" || d.Sources[0].Counts.Documents != 0 {
		t.Fatalf("%+v %v", d, err)
	}
	_, _ = m.SetRepositoryEnabled(ctx, "one", false)
	d, _ = m.Discover(ctx, promptrepo.DiscoveryRequest{})
	if d.Counts.Repositories != 0 || d.Sources[0].Enabled || a.calls != 1 {
		t.Fatal("disabled source fetched")
	}
	_, _ = m.SetRepositoryEnabled(ctx, "one", true)
	rev := "other"
	_, err = m.SetRepository(ctx, "one", promptrepo.RepositoryPatch{Revision: &rev})
	if err != nil {
		t.Fatal(err)
	}
	p, _ := m.ShowRepository(ctx, "one")
	if p.Profile.Revision != rev {
		t.Fatal("patch missing")
	}
}
func TestDiscoverConcurrencyAndRemovedSource(t *testing.T) {
	a := &discoveryAdapter{entered: make(chan struct{}, 8), release: make(chan struct{})}
	m := discoveryManager(t, a, time.Now)
	for i := 0; i < 7; i++ {
		addDiscoverySource(t, m, fmt.Sprint("r", i))
	}
	done := make(chan error, 1)
	go func() { _, err := m.Discover(context.Background(), promptrepo.DiscoveryRequest{}); done <- err }()
	for i := 0; i < 4; i++ {
		select {
		case <-a.entered:
		case <-time.After(3 * time.Second):
			t.Fatal("refresh did not parallelize")
		}
	}
	if err := m.RemoveRepository(context.Background(), "r0"); err != nil {
		t.Fatal(err)
	}
	close(a.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if a.peak > 4 {
		t.Fatalf("peak=%d", a.peak)
	}
	if _, err := m.ShowRepository(context.Background(), "r0"); err == nil {
		t.Fatal("deleted source resurrected")
	}
}
