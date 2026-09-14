package engine

import (
	"context"
	"github.com/yeisme/promptrepo"
	"testing"
	"time"
)

func TestBrowseFiltersStatsAndPagination(t *testing.T) {
	a := &discoveryAdapter{}
	m := discoveryManager(t, a, time.Now)
	addDiscoverySource(t, m, "a")
	addDiscoverySource(t, m, "b")
	b, err := m.Browse(context.Background(), promptrepo.BrowseRequest{DiscoveryRequest: promptrepo.DiscoveryRequest{Limit: 1}, MediaTypes: []string{"image", "audio"}, TemplateLocales: []string{"en"}, GroupBy: []string{"repository", "media"}})
	if err != nil {
		t.Fatal(err)
	}
	if b.Matched.Documents != 2 || b.Returned.Documents != 1 || b.Remaining != 1 || len(b.Browse.Groups) != 2 || b.Browse.Totals.Entries != 2 {
		t.Fatalf("bad filtered pagination: %+v", b)
	}
	if a.calls != 2 {
		t.Fatal("source refreshed more than once")
	}
	only, err := m.Browse(context.Background(), promptrepo.BrowseRequest{Repositories: []string{"b"}, DiscoveryRequest: promptrepo.DiscoveryRequest{Offline: true}})
	if err != nil || len(only.Sources) != 1 || only.Sources[0].ID != "b" {
		t.Fatal("repository selection escaped")
	}
	_, err = m.Browse(context.Background(), promptrepo.BrowseRequest{Repositories: []string{"missing"}})
	if err == nil {
		t.Fatal("missing repository was hidden")
	}
}
