package conformance

import (
	"github.com/yeisme/promptrepo"
	"strings"
)

type BrowseCase struct {
	Name           string                    `json:"name"`
	Records        []promptrepo.BrowseRecord `json:"records"`
	Filters        map[string][]string       `json:"filters"`
	Query          string                    `json:"query"`
	Dimensions     []string                  `json:"dimensions"`
	ExpectedRefs   []string                  `json:"expected_refs"`
	ExpectedTotals promptrepo.BrowseTotals   `json:"expected_totals"`
	ExpectedGroups []promptrepo.BrowseGroup  `json:"expected_groups"`
}

func BrowseCases() []BrowseCase {
	a := promptrepo.BrowseRecord{DiscoveryTemplate: promptrepo.DiscoveryTemplate{Ref: "promptrepo://official/image/hero@1?locale=en&role=main", Title: "人物参考 👩‍🎨", Summary: "Studio portrait / reference", Role: "main", Locale: "en", Digest: "sha256:" + strings.Repeat("a", 64), Media: []string{"image", "video"}, Consumers: []string{"eikona", "scaena"}, CompilerStatus: "not_checked"}, Repository: "official", Solution: "image/hero", SolutionTitle: "Hero", Category: "image", Tags: []string{"job:generate", "subject:hero"}, Capabilities: []string{"character_asset"}, Maturity: "exploratory", Rights: "internal", SourceState: "ready"}
	a.SearchAliases = []string{"肖像档案"}
	b := a
	b.Repository = "personal"
	b.Ref = "promptrepo://personal/image/hero@1?locale=zh-CN&role=main"
	b.Locale = "zh-CN"
	b.Media = []string{"image"}
	b.SourceState = "degraded"
	c := promptrepo.BrowseRecord{DiscoveryTemplate: promptrepo.DiscoveryTemplate{Ref: "promptrepo://personal/writing/meeting@1?locale=en", Title: "Meeting", Summary: "Collect actions", Role: "main", Locale: "en", Digest: "invalid"}, Repository: "personal", Solution: "writing/meeting", Category: "office"}
	records := []promptrepo.BrowseRecord{a, b, c}
	cases := []BrowseCase{
		{Name: "all-and-overlap", Records: records, Filters: map[string][]string{}, Dimensions: []string{"repository", "media"}},
		{Name: "or-and-query", Records: records, Filters: map[string][]string{"media": {"image", "audio"}, "consumer": {"eikona"}}, Query: "人物 reference", Dimensions: []string{"media"}},
		{Name: "template-language", Records: records, Filters: map[string][]string{"template_locale": {"zh-CN"}}, Dimensions: []string{"repository"}},
		{Name: "unknown-metadata", Records: records, Filters: map[string][]string{"rights": {"unknown"}}, Dimensions: []string{"rights"}},
		{Name: "tag-and-maturity", Records: records, Filters: map[string][]string{"tag": {"job:generate"}, "maturity": {"exploratory"}}, Dimensions: []string{"consumer", "media"}},
		{Name: "localized-alias", Records: records, Filters: map[string][]string{}, Query: "肖像", Dimensions: []string{"repository"}},
		{Name: "no-matches", Records: records, Filters: map[string][]string{"compiler_status": {"owner_only"}}, Dimensions: []string{"category"}},
	}
	for i := range cases {
		test := &cases[i]
		q := promptrepo.BrowseRequest{MediaTypes: test.Filters["media"], Consumers: test.Filters["consumer"], TemplateLocales: test.Filters["template_locale"], Rights: test.Filters["rights"], Tags: test.Filters["tag"], Maturities: test.Filters["maturity"], CompilerStates: test.Filters["compiler_status"], Query: test.Query}
		rows := []promptrepo.BrowseRecord{}
		test.ExpectedRefs = []string{}
		for _, r := range test.Records {
			if promptrepo.MatchBrowseRecord(r, q) {
				rows = append(rows, r)
				test.ExpectedRefs = append(test.ExpectedRefs, r.Ref)
			}
		}
		test.ExpectedTotals = promptrepo.CountBrowse(rows)
		test.ExpectedGroups, _ = promptrepo.GroupBrowse(rows, test.Dimensions)
	}
	return cases
}
