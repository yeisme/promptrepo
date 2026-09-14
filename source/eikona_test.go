package source

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/yeisme/promptrepo"
)

func TestEikonaRemoteDiscoveryAndFailureModes(t *testing.T) {
	for _, mode := range []string{"success", "unsupported", "unauthorized", "truncated", "schema", "redirect"} {
		t.Run(mode, func(t *testing.T) {
			called := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if mode == "unauthorized" {
					http.Error(w, "private-token", 401)
					return
				}
				if mode == "redirect" {
					http.Redirect(w, r, "https://invalid.example/secret", 302)
					return
				}
				if strings.HasSuffix(r.URL.Path, "/actions") {
					action := "prompts.catalog.list"
					if mode == "unsupported" {
						action = "prompts.list"
					}
					_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"actions": []any{map[string]any{"action": action, "kind": "readonly", "input_schema": map[string]any{"type": "object"}}}}})
					return
				}
				called++
				var req struct {
					Action string `json:"action"`
				}
				_ = json.NewDecoder(r.Body).Decode(&req)
				if req.Action != "prompts.catalog.list" {
					t.Error("wrong owner action")
				}
				schema := "eikona.template_catalog.v1"
				if mode == "schema" {
					schema = "v999"
				}
				catalog := promptrepo.EikonaCatalog{SchemaVersion: schema, Entries: []promptrepo.EikonaTemplate{{ID: "p1", Title: "Portrait", Version: "1", Kind: "template", Category: "portraits", Media: "image", Digest: "sha256:" + strings.Repeat("a", 64), OwnerRef: "eikona://prompts/p1"}}}
				_ = json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{"status": "success", "truncated": mode == "truncated", "data": catalog}})
			}))
			defer server.Close()
			result, err := (EikonaAdapter{}).Sync(context.Background(), promptrepo.RepositoryProfile{ID: "remote", Source: "eikona+" + server.URL}, "")
			if mode == "success" {
				if err != nil || len(result.Catalog.Solutions) != 1 || result.Catalog.Discovery[0].CompilerStatus != "owner_only" {
					t.Fatalf("result: %+v %v", result, err)
				}
			} else {
				if err == nil {
					t.Fatal("failure hidden")
				}
				if strings.Contains(err.Error(), "private-token") {
					t.Fatal("error leaked remote payload")
				}
			}
			if mode == "unsupported" && called != 0 {
				t.Fatal("executed action absent from discovery")
			}
		})
	}
}
func TestEikonaRejectsUnsafeSourceAndBodies(t *testing.T) {
	for _, s := range []string{"eikona+https://user:secret@host", "eikona+https://host?token=secret", "eikona+file://host/root", "eikona+file:relative"} {
		if _, err := DetectKind(s); err == nil {
			t.Fatalf("accepted unsafe source %s", s)
		}
	}
	if _, err := (EikonaAdapter{}).ReadTemplate(context.Background(), promptrepo.RepositoryProfile{}, promptrepo.SnapshotMetadata{}, promptrepo.TemplateRole{}, ""); err == nil {
		t.Fatal("owner-only template read allowed")
	}
}
