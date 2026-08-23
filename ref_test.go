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

func TestTemplateAddressRoundTripUsesCanonicalOrder(t *testing.T) {
	raw := "promptrepo://official/audio/podcast-narration@1.0.0?kind=template&locale=zh-CN&role=main&path=prompts%2Fmain.zh-CN.md&selector=json-pointer%3A%2Ftitle&digest=sha256%3A0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef&snapshot=sha256%3Afedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	address, err := ParseTemplateAddress(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got := FormatTemplateAddress(address); got != raw {
		t.Fatalf("round trip: %s", got)
	}
}

func TestTemplateAddressRejectsInvalidFields(t *testing.T) {
	valid := "promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN"
	for _, raw := range []string{
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN&path=../secret.md",
		"promptrepo://user@official/audio/podcast@1.0.0?kind=template&locale=zh-CN",
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN#private",
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN&selector=json-pointer:field",
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN&digest=not-a-digest",
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN&locale=en",
		"promptrepo://official/audio/podcast@1.0.0?kind=solution&locale=zh-CN",
		"promptrepo://official/audio/podcast@1.0.0?kind=template&locale=zh-CN&unknown=x",
	} {
		if _, err := ParseTemplateAddress(raw); err == nil {
			t.Fatalf("expected rejection for %s", raw)
		}
	}
	if _, err := ParseTemplateAddress(valid); err != nil {
		t.Fatal(err)
	}
	for _, selector := range []string{"heading:Introduction", "json-pointer:/title", "yaml-pointer:/metadata/title"} {
		if _, err := ParseTemplateAddress(valid + "&selector=" + selector); err != nil {
			t.Fatalf("selector %q: %v", selector, err)
		}
	}
}
