package promptrepo_test

import (
	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/conformance"
	"strings"
	"testing"
)

func TestBrowseConformanceCounts(t *testing.T) {
	cases := conformance.BrowseCases()
	if cases[0].ExpectedTotals.Entries != 3 || cases[0].ExpectedTotals.UniqueContents != 2 {
		t.Fatal("duplicate bytes did not collapse")
	}
	if len(cases[1].ExpectedRefs) != 2 || len(cases[2].ExpectedRefs) != 1 || len(cases[3].ExpectedRefs) != 1 {
		t.Fatal("OR/AND/unknown/language filters wrong")
	}
	for _, c := range cases {
		if c.Name == "localized-alias" && len(c.ExpectedRefs) != 2 {
			t.Fatal("localized metadata was not searchable")
		}
	}
	groups, overlap := promptrepo.GroupBrowse(cases[0].Records, []string{"media"})
	if !overlap || len(groups) != 3 {
		t.Fatal("multi-valued membership lost")
	}
}
func TestBrowseValidationAndMissingHashes(t *testing.T) {
	for _, q := range []promptrepo.BrowseRequest{{GroupBy: []string{"media", "media"}}, {GroupBy: []string{"repository", "media", "category"}}, {Sort: "body"}, {Order: "random"}, {Dedupe: "title"}, {Metric: "cost"}} {
		if _, err := promptrepo.NormalizeBrowseRequest(q); err == nil {
			t.Fatalf("invalid query accepted: %+v", q)
		}
	}
	rows := []promptrepo.BrowseRecord{{DiscoveryTemplate: promptrepo.DiscoveryTemplate{Ref: "a", Digest: "invalid"}}, {DiscoveryTemplate: promptrepo.DiscoveryTemplate{Ref: "b", Digest: "invalid"}}, {DiscoveryTemplate: promptrepo.DiscoveryTemplate{Digest: ""}}}
	if got := promptrepo.CountBrowse(rows); got.UniqueContents != 3 {
		t.Fatalf("invalid digests merged: %+v", got)
	}
	row := conformance.BrowseCases()[0].Records[0]
	row.Digest = "SHA256:" + strings.Repeat("A", 64)
	if promptrepo.BrowseContentKey(row) != "sha256:"+strings.Repeat("a", 64) {
		t.Fatal("digest normalization failed")
	}
	q, err := promptrepo.NormalizeBrowseRequest(promptrepo.BrowseRequest{DiscoveryRequest: promptrepo.DiscoveryRequest{Media: "image"}, MediaTypes: []string{"video"}})
	if err != nil || len(q.MediaTypes) != 2 {
		t.Fatal("legacy selector not preserved")
	}
}
