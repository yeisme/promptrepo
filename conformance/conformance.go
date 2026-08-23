package conformance

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yeisme/promptrepo"
)

func MinimalCatalog() promptrepo.Catalog {
	templateDigest := sha256.Sum256([]byte(MinimalTemplateBody))
	return promptrepo.Catalog{
		SchemaVersion: promptrepo.CatalogSchemaVersion,
		Repository:    promptrepo.RepositoryMetadata{ID: "fixture", Name: "Promptrepo fixture", DefaultLocale: "zh-CN", TaxonomyVersion: "v1"},
		GeneratedAt:   time.Unix(0, 0).UTC(),
		Solutions: []promptrepo.Solution{{
			PackageID: "audio", ID: "podcast-narration", Version: "1.0.0", Category: "audio",
			Tags: []string{"job:generate", "artifact:voiceover"}, Capabilities: []string{"audio", "voice", "tts"},
			Rights: "internal", Maturity: "first-support", Digest: "sha256:fixture-solution",
			Locales: map[string]promptrepo.LocalizedText{
				"zh-CN": {Title: "中文播客旁白", Summary: "生成自然、清晰的中文播客旁白", Aliases: []string{"播客配音"}},
				"en":    {Title: "Chinese podcast narration", Summary: "Generate natural Chinese podcast narration"},
			},
			Templates: []promptrepo.TemplateRole{{Role: "main", Locale: "zh-CN", Path: "prompts/main.zh-CN.md", Digest: "sha256:" + hex.EncodeToString(templateDigest[:])}},
		}},
	}
}

func WriteCatalog(t testing.TB, root string, catalog promptrepo.Catalog) string {
	t.Helper()
	for _, solution := range catalog.Solutions {
		for _, template := range solution.Templates {
			if template.Path != "prompts/main.zh-CN.md" || template.Digest == "" {
				continue
			}
			path := filepath.Join(root, filepath.FromSlash(template.Path))
			if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(MinimalTemplateBody), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	catalog.Digest = digest
	payload, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "catalog.json")
	if err := os.WriteFile(path, append(payload, '\n'), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

const MinimalTemplateBody = "# 中文播客旁白\n\n请生成自然、清晰的中文播客旁白。\n"
