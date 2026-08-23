package promptrepo

import "testing"

func TestRefRoundTrip(t *testing.T) {
	raw := "promptrepo://official/audio/podcast-narration@1.0.0?locale=zh-CN"
	ref, err := ParseRef(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatRef(ref); got != raw {
		t.Fatalf("round trip: %s", got)
	}
}

func TestRefRequiresExactVersion(t *testing.T) {
	if _, err := ParseRef("promptrepo://official/audio/podcast-narration"); err == nil {
		t.Fatal("expected exact version error")
	}
}
