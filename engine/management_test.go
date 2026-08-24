package engine_test

import (
	"context"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/engine"
)

func TestEffectiveRepositorySetUsesCanonicalUserStateWithoutWrites(t *testing.T) {
	t.Parallel()

	configRoot := filepath.Join(t.TempDir(), "config")
	manager, err := engine.New(engine.Options{ConfigRoot: configRoot, CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	for _, id := range []string{"alpha", "beta"} {
		_, err := manager.AddRepository(ctx, promptrepo.AddRepositoryRequest{Profile: promptrepo.RepositoryProfile{
			ID: id, Source: (&url.URL{Scheme: "file", Path: filepath.Join(t.TempDir(), id)}).String(), Trust: "user_trusted",
		}})
		if err != nil {
			t.Fatal(err)
		}
	}
	statePath := filepath.Join(configRoot, "state.json")
	before, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}

	set, err := manager.EffectiveRepositorySet(ctx, promptrepo.EffectiveRepositorySetRequest{
		Consumer: "eikona",
		Repositories: []promptrepo.RepositorySetCandidate{
			{RepositoryID: "alpha", Scope: promptrepo.RepositoryScopeOrganization, Trust: "official", Health: "ready", Enabled: true},
			{RepositoryID: "external", Scope: promptrepo.RepositoryScopeOrganization, Trust: "verified", Health: "ready", Enabled: true},
		},
		Bindings: []promptrepo.RepositoryScopeBinding{
			{Scope: promptrepo.RepositoryScopeUser, ScopeRef: "preference", RepositoryIDs: []string{"beta", "alpha"}, Priority: 100, Enabled: true},
			{Scope: promptrepo.RepositoryScopeOrganization, ScopeRef: "organization", RepositoryIDs: []string{"external"}, Enabled: true},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(set.EffectiveRepositories))
	for _, entry := range set.EffectiveRepositories {
		got = append(got, entry.RepositoryID)
		if entry.RepositoryID == "alpha" && (entry.Health != "configured" || entry.Trust != "user_trusted" || entry.RepositoryScope != promptrepo.RepositoryScopeUser) {
			t.Fatalf("caller overrode canonical user metadata: %+v", entry)
		}
	}
	if !reflect.DeepEqual(got, []string{"beta", "alpha", "external"}) {
		t.Fatalf("embedded effective set order changed: %v", got)
	}
	after, err := os.ReadFile(statePath)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("effective repository set modified durable state")
	}
}

func TestEvaluatePolicyAdapterIsSideEffectFree(t *testing.T) {
	t.Parallel()

	configRoot := filepath.Join(t.TempDir(), "config")
	manager, err := engine.New(engine.Options{ConfigRoot: configRoot, CacheRoot: filepath.Join(t.TempDir(), "cache")})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := manager.EvaluatePolicy(context.Background(), promptrepo.EvaluatePolicyRequest{
		Repository:  promptrepo.RepositorySetEntry{RepositoryID: "fixture", Scope: promptrepo.RepositoryScopeUser, Trust: "official", Health: "ready", Enabled: true},
		OperationID: promptrepo.OperationRepositoryList,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed || decision.Readiness != promptrepo.ReadinessReady {
		t.Fatalf("unexpected policy decision: %+v", decision)
	}
	if _, err := os.Stat(filepath.Join(configRoot, "state.json")); !os.IsNotExist(err) {
		t.Fatalf("policy evaluation created durable state: %v", err)
	}
}
