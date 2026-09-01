package promptrepo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestUITemplateAddressRoundTripUsesIndependentCanonicalOrder(t *testing.T) {
	raw := "promptrepo://official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=zh-CN&role=review&path=ui%2Freview.zh-CN.html&digest=sha256%3A0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef&snapshot=sha256%3Afedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	address, err := ParseUITemplateAddress(raw)
	if err != nil {
		t.Fatal(err)
	}
	if address.RepositoryID != "official" || address.PackageID != "scaena" || address.SolutionID != "storyboard-review" || address.Role != "review" || address.Path != "ui/review.zh-CN.html" {
		t.Fatalf("unexpected address: %+v", address)
	}
	if got := FormatUITemplateAddress(address); got != raw {
		t.Fatalf("canonical round trip changed:\nwant %s\n got %s", raw, got)
	}
	if _, err := ParseTemplateAddress(raw); err == nil {
		t.Fatal("released template parser must continue rejecting kind=ui-template")
	}
}

func TestUITemplateAddressRejectsUnsafeOrAmbiguousInputs(t *testing.T) {
	base := "promptrepo://official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=zh-CN"
	for _, raw := range []string{
		base + "&path=../private.html",
		base + "&path=%2Fprivate.html",
		base + "&selector=json-pointer:%2Fbody",
		base + "&unknown=value",
		base + "&locale=en",
		base + "&digest=not-a-digest",
		"promptrepo://user@official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=zh-CN",
		base + "#private",
		"promptrepo://official/scaena/storyboard-review@1.0.0?kind=template&locale=zh-CN",
		"promptrepo://official/scaena/storyboard-review@1.0.0?kind=ui-template&locale=not_a_locale",
	} {
		if _, err := ParseUITemplateAddress(raw); ErrorCode(err) != CodeUITemplateAddressInvalid {
			t.Fatalf("expected %s for %q, got %v", CodeUITemplateAddressInvalid, raw, err)
		}
	}
}

func TestUITemplateBundleLimitsAndSlots(t *testing.T) {
	bundle := validUITemplateBundle(t)
	if err := ValidateUITemplateBundle(bundle); err != nil {
		t.Fatal(err)
	}

	zeroLimits := bundle
	zeroLimits.Limits = UITemplateLimitsV1{}
	if err := ValidateUITemplateBundle(zeroLimits); ErrorCode(err) != CodeUITemplateLimitExceeded {
		t.Fatalf("zero limits: %v", err)
	}

	tooLarge := bundle
	tooLarge.HTMLFragment = make([]byte, MaxUITemplateHTMLBytesV1+1)
	if err := ValidateUITemplateBundle(tooLarge); ErrorCode(err) != CodeUITemplateLimitExceeded {
		t.Fatalf("large HTML: %v", err)
	}

	invalidUTF8 := bundle
	invalidUTF8.CSS = []byte{0xff}
	if err := ValidateUITemplateBundle(invalidUTF8); ErrorCode(err) != CodeUITemplateLimitExceeded {
		t.Fatalf("invalid UTF-8: %v", err)
	}

	duplicate := bundle
	duplicate.Slots = append(append([]UITemplateSlotV1(nil), bundle.Slots...), bundle.Slots[0])
	if err := ValidateUITemplateBundle(duplicate); ErrorCode(err) != CodeUITemplateSlotInvalid {
		t.Fatalf("duplicate slot: %v", err)
	}
}

func TestUITemplateErrorCodesAreStable(t *testing.T) {
	want := map[string]string{
		"address":  CodeUITemplateAddressInvalid,
		"limit":    CodeUITemplateLimitExceeded,
		"html":     CodeUITemplateHTMLForbidden,
		"css":      CodeUITemplateCSSForbidden,
		"slot":     CodeUITemplateSlotInvalid,
		"digest":   CodeUITemplateDigestMismatch,
		"snapshot": CodeUITemplateSnapshotMismatch,
	}
	if want["address"] != "UI_TEMPLATE_ADDRESS_INVALID" ||
		want["limit"] != "UI_TEMPLATE_LIMIT_EXCEEDED" ||
		want["html"] != "UI_TEMPLATE_HTML_FORBIDDEN" ||
		want["css"] != "UI_TEMPLATE_CSS_FORBIDDEN" ||
		want["slot"] != "UI_TEMPLATE_SLOT_INVALID" ||
		want["digest"] != "UI_TEMPLATE_DIGEST_MISMATCH" ||
		want["snapshot"] != "UI_TEMPLATE_SNAPSHOT_MISMATCH" {
		t.Fatalf("UI template error code drift: %+v", want)
	}
}

func TestUITemplateHTMLValidatorFailsClosed(t *testing.T) {
	bundle := validUITemplateBundle(t)
	for name, fragment := range map[string]string{
		"script":                 `<section><script>alert(1)</script><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"event attribute":        `<section ONCLICK="alert(1)"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"URL attribute":          `<section><a href="https://example.test">x</a><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"inline style":           `<section style="color:red"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"document node":          `<html><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></html>`,
		"SVG namespace":          `<section><svg></svg><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"MathML namespace":       `<section><math></math><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"custom element":         `<x-review><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></x-review>`,
		"content editable":       `<section contenteditable="true"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"duplicate attribute":    `<section class="first" CLASS="second"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"namespace attribute":    `<section xml:lang="en"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"framework directive":    `<section hx-get="/private"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"arbitrary data hook":    `<section data-controller="review"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"attribution URL":        `<section attributionsrc="https://example.test/register"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"template element":       `<section><template><div data-promptrepo-slot="title"></div></template><div data-promptrepo-slot="scenes"></div></section>`,
		"media element":          `<section><video></video><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"mustache template":      `<section>{{private_action}}<div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"template attribute":     `<section class="{{private_class}}"><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"EJS template":           `<section><%= privateAction %><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"EL template":            `<section>${privateAction}<div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"unclosed element":       `<section><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div>`,
		"mismatched element":     `<section><div data-promptrepo-slot="title"></section><div data-promptrepo-slot="scenes"></div>`,
		"non-void self closing":  `<section><div data-promptrepo-slot="title"/><div data-promptrepo-slot="scenes"></div></section>`,
		"unterminated attribute": `<section class="review><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
	} {
		t.Run(name, func(t *testing.T) {
			candidate := bundle
			candidate.HTMLFragment = []byte(fragment)
			if err := ValidateUITemplateBundle(candidate); ErrorCode(err) != CodeUITemplateHTMLForbidden {
				t.Fatalf("expected HTML rejection, got %v", err)
			}
		})
	}

	for name, fragment := range map[string]string{
		"undeclared": `<section><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="unknown"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"duplicate":  `<section><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`,
		"missing":    `<section><div data-promptrepo-slot="title"></div></section>`,
	} {
		t.Run(name, func(t *testing.T) {
			candidate := bundle
			candidate.HTMLFragment = []byte(fragment)
			if err := ValidateUITemplateBundle(candidate); ErrorCode(err) != CodeUITemplateSlotInvalid {
				t.Fatalf("expected slot rejection, got %v", err)
			}
		})
	}
}

func TestUITemplateHTMLStaticReviewAllowlistIsStable(t *testing.T) {
	for tag := range allowedUITemplateHTMLTags {
		t.Run("element/"+tag, func(t *testing.T) {
			fragment := "<" + tag + "></" + tag + ">"
			if _, void := uiTemplateHTMLVoidTags[tag]; void {
				fragment = "<" + tag + ">"
			}
			if err := validateUITemplateHTML([]byte(fragment), nil, DefaultUITemplateLimitsV1()); err != nil {
				t.Fatalf("allowlisted element %q rejected: %v", tag, err)
			}
		})
	}

	for attribute := range allowedUITemplateHTMLAttributes {
		t.Run("attribute/"+attribute, func(t *testing.T) {
			fragment := `<div ` + attribute + `="value"></div>`
			if err := validateUITemplateHTML([]byte(fragment), nil, DefaultUITemplateLimitsV1()); err != nil {
				t.Fatalf("allowlisted attribute %q rejected: %v", attribute, err)
			}
		})
	}

	if err := validateUITemplateHTML([]byte(`<div aria-label="Review"></div>`), nil, DefaultUITemplateLimitsV1()); err != nil {
		t.Fatalf("allowlisted ARIA attribute rejected: %v", err)
	}
	if err := validateUITemplateHTML([]byte(`<div aria-="Review"></div>`), nil, DefaultUITemplateLimitsV1()); ErrorCode(err) != CodeUITemplateHTMLForbidden {
		t.Fatalf("malformed ARIA attribute accepted: %v", err)
	}
	if err := validateUITemplateHTML([]byte(`<div unknown="value"></div>`), nil, DefaultUITemplateLimitsV1()); ErrorCode(err) != CodeUITemplateHTMLForbidden {
		t.Fatalf("unknown attribute accepted: %v", err)
	}
}

func TestUITemplateCSSValidatorRejectsBypassesAndParserErrors(t *testing.T) {
	bundle := validUITemplateBundle(t)
	for name, stylesheet := range map[string]string{
		"import":            `@import "https://example.test/review.css";`,
		"escaped URL":       `.review { background: u\72l(https://example.test/a.png); }`,
		"comment URL":       `.review { background: u/**/rl(https://example.test/a.png); }`,
		"expression":        `.review { width: e\78pression(alert(1)); }`,
		"behavior":          `.review { behavior: url(private.htc); }`,
		"binding":           `.review { -moz-binding: url(private.xml); }`,
		"font source":       `@font-face { font-family: private; src: local(private); }`,
		"image set":         `.review { background: image-set("https://example.test/a.png" 1x); }`,
		"HTML break":        `.review::before { content: "</style><script>"; }`,
		"unterminated rule": `.review { color: red;`,
	} {
		t.Run(name, func(t *testing.T) {
			candidate := bundle
			candidate.CSS = []byte(stylesheet)
			if err := ValidateUITemplateBundle(candidate); ErrorCode(err) != CodeUITemplateCSSForbidden {
				t.Fatalf("expected CSS rejection, got %v", err)
			}
		})
	}
}

func TestUITemplateDigestCoversMetadataAndBodiesButNotSnapshot(t *testing.T) {
	bundle := validUITemplateBundle(t)
	baseDigest, err := CanonicalUITemplateDigest(bundle)
	if err != nil {
		t.Fatal(err)
	}

	changedHTML := bundle
	changedHTML.HTMLFragment = append(append([]byte(nil), bundle.HTMLFragment...), '\n')
	assertUITemplateDigestChanges(t, baseDigest, changedHTML)

	changedCSS := bundle
	changedCSS.CSS = append(append([]byte(nil), bundle.CSS...), '\n')
	assertUITemplateDigestChanges(t, baseDigest, changedCSS)

	changedMetadata := bundle
	changedMetadata.Slots = append([]UITemplateSlotV1(nil), bundle.Slots...)
	changedMetadata.Slots[0].Required = false
	assertUITemplateDigestChanges(t, baseDigest, changedMetadata)

	reordered := bundle
	reordered.Slots = []UITemplateSlotV1{bundle.Slots[1], bundle.Slots[0]}
	reorderedDigest, err := CanonicalUITemplateDigest(reordered)
	if err != nil {
		t.Fatal(err)
	}
	if reorderedDigest != baseDigest {
		t.Fatalf("slot order must normalize: %s != %s", reorderedDigest, baseDigest)
	}

	changedSnapshot := bundle
	changedSnapshot.Snapshot = testUITemplateDigest("another-snapshot")
	changedSnapshot.Address.Snapshot = changedSnapshot.Snapshot
	snapshotDigest, err := CanonicalUITemplateDigest(changedSnapshot)
	if err != nil {
		t.Fatal(err)
	}
	if snapshotDigest != baseDigest {
		t.Fatalf("snapshot must remain a separate binding: %s != %s", snapshotDigest, baseDigest)
	}
}

func TestUITemplateExactDigestAndSnapshotBindings(t *testing.T) {
	bundle := validUITemplateBundle(t)
	tampered := bundle
	tampered.HTMLFragment = append(append([]byte(nil), bundle.HTMLFragment...), []byte("<!-- changed -->")...)
	if err := ValidateUITemplateBundle(tampered); ErrorCode(err) != CodeUITemplateDigestMismatch {
		t.Fatalf("tampered body: %v", err)
	}

	wrongSnapshot := bundle
	wrongSnapshot.Address.Snapshot = testUITemplateDigest("wrong-snapshot")
	if err := ValidateUITemplateBundle(wrongSnapshot); ErrorCode(err) != CodeUITemplateSnapshotMismatch {
		t.Fatalf("wrong snapshot: %v", err)
	}
}

func TestUITemplateMachineProjectionOmitsBodiesAndPrivatePaths(t *testing.T) {
	const sentinel = "PRIVATE_UI_TEMPLATE_BODY_SENTINEL"
	bundle := validUITemplateBundle(t)
	bundle.HTMLFragment = []byte(`<section>` + sentinel + `<div data-promptrepo-slot="title"></div><div data-promptrepo-slot="scenes"></div></section>`)
	bundle.CSS = []byte(`.review { color: #222; } /* ` + sentinel + ` */`)
	digest, err := CanonicalUITemplateDigest(bundle)
	if err != nil {
		t.Fatal(err)
	}
	bundle.ContentDigest = digest
	bundle.Address.Digest = digest

	for name, value := range map[string]any{"bundle": bundle, "inspection": NewUITemplateInspection(bundle)} {
		payload, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(payload), sentinel) || strings.Contains(string(payload), "/private/workspace") {
			t.Fatalf("%s leaked body or private path: %s", name, payload)
		}
	}
	for _, fieldName := range []string{"HTMLFragment", "CSS"} {
		field, ok := reflect.TypeOf(UITemplateBundleV1{}).FieldByName(fieldName)
		if !ok || field.Tag.Get("json") != "-" || field.Tag.Get("yaml") != "-" {
			t.Fatalf("%s must be body-redacted", fieldName)
		}
	}

	unsafe := bundle
	unsafe.HTMLFragment = []byte(`<script>` + sentinel + `</script>`)
	err = ValidateUITemplateBundle(unsafe)
	if err == nil || strings.Contains(err.Error(), sentinel) {
		t.Fatalf("validation error leaked body: %v", err)
	}
}

func TestUITemplateFileDescriptorIsStrictAndBodyFree(t *testing.T) {
	bundle := validUITemplateBundle(t)
	descriptor := UITemplateFileDescriptorV1{
		SchemaVersion: bundle.SchemaVersion,
		Address:       bundle.Address,
		HTMLPath:      bundle.Address.Path,
		CSSPath:       "ui/review.zh-CN.css",
		Slots:         bundle.Slots,
		Security:      bundle.Security,
		Limits:        bundle.Limits,
		HTMLBytes:     len(bundle.HTMLFragment),
		CSSBytes:      len(bundle.CSS),
		BodyBytes:     len(bundle.HTMLFragment) + len(bundle.CSS),
		ContentDigest: bundle.ContentDigest,
		Snapshot:      bundle.Snapshot,
	}
	payload, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeUITemplateFileDescriptorV1(payload)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decoded, descriptor) {
		t.Fatalf("descriptor round trip changed:\nwant %+v\n got %+v", descriptor, decoded)
	}
	if strings.Contains(string(payload), string(bundle.HTMLFragment)) || strings.Contains(string(payload), string(bundle.CSS)) {
		t.Fatalf("descriptor leaked body: %s", payload)
	}
	if path, err := UITemplateFileDescriptorPath(bundle.Address); err != nil || path != "ui/review.zh-CN.ui-template.json" {
		t.Fatalf("descriptor path: %q, %v", path, err)
	}
	withUnknown := append(append([]byte(nil), payload[:len(payload)-1]...), []byte(`,"unknown":true}`)...)
	if _, err := DecodeUITemplateFileDescriptorV1(withUnknown); ErrorCode(err) != CodeUITemplateAddressInvalid {
		t.Fatalf("unknown descriptor field: %v", err)
	}
}

func TestUITemplateAdditiveContractPreservesReleasedAddressBehavior(t *testing.T) {
	oldRaw := "promptrepo://official/audio/podcast-narration@1.0.0?kind=template&locale=zh-CN&role=main&path=prompts%2Fmain.zh-CN.md"
	oldAddress, err := ParseTemplateAddress(oldRaw)
	if err != nil {
		t.Fatal(err)
	}
	if FormatTemplateAddress(oldAddress) != oldRaw {
		t.Fatalf("released TemplateAddress formatting changed: %s", FormatTemplateAddress(oldAddress))
	}
	if _, err := ParseUITemplateAddress(oldRaw); ErrorCode(err) != CodeUITemplateAddressInvalid {
		t.Fatalf("new parser must not absorb old kind=template input: %v", err)
	}
}

func validUITemplateBundle(t *testing.T) UITemplateBundleV1 {
	t.Helper()
	snapshot := testUITemplateDigest("fixture-snapshot")
	bundle := UITemplateBundleV1{
		SchemaVersion: UITemplateSchemaVersionV1,
		Address: UITemplateAddress{
			RepositoryID: "official",
			PackageID:    "scaena",
			SolutionID:   "storyboard-review",
			Version:      "1.0.0",
			Locale:       "zh-CN",
			Role:         "review",
			Path:         "ui/review.zh-CN.html",
			Snapshot:     snapshot,
		},
		HTMLFragment: []byte(`<section class="review"><h1 data-promptrepo-slot="title"></h1><div data-promptrepo-slot="scenes"></div></section>`),
		CSS:          []byte(`.review { display: grid; gap: 1rem; color: #222; } @media (min-width: 40rem) { .review { gap: 2rem; } }`),
		Slots: []UITemplateSlotV1{
			{Name: "title", Kind: UITemplateSlotKindText, Required: true, Cardinality: UITemplateSlotCardinalityOne},
			{Name: "scenes", Kind: UITemplateSlotKindCollection, Required: true, Cardinality: UITemplateSlotCardinalityMany},
		},
		Security: UITemplateSecurityStaticReviewFragmentV1,
		Limits:   DefaultUITemplateLimitsV1(),
		Snapshot: snapshot,
	}
	digest, err := CanonicalUITemplateDigest(bundle)
	if err != nil {
		t.Fatalf("build valid UI template fixture: %v", err)
	}
	bundle.ContentDigest = digest
	bundle.Address.Digest = digest
	return bundle
}

func testUITemplateDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}

func assertUITemplateDigestChanges(t *testing.T, base string, bundle UITemplateBundleV1) {
	t.Helper()
	digest, err := CanonicalUITemplateDigest(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if digest == base {
		t.Fatalf("canonical digest did not change for modified bundle: %s", digest)
	}
}
