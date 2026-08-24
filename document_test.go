package promptrepo

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestTemplateDocumentDescriptorPathAndBinding(t *testing.T) {
	template := TemplateRole{Role: "main", Locale: "zh-CN", Path: "solutions/drama/character/prompts/main.zh-CN.yaml", Digest: digestBytes([]byte("task: {}\n"))}
	got, err := TemplateDocumentDescriptorPath(template)
	if err != nil {
		t.Fatal(err)
	}
	if got != "solutions/drama/character/contracts/documents/main.zh-CN.document.json" {
		t.Fatalf("descriptor path = %q", got)
	}
	descriptor := testDocumentDescriptor(t, DocumentFormatYAML, template.Path, []byte("task: {}\n"))
	if err := ValidateTemplateDocumentDescriptorBinding(descriptor, Solution{PackageID: "drama", ID: "character", Version: "1.0.0"}, template); err != nil {
		t.Fatal(err)
	}
	descriptor.SourceDigest = digestBytes([]byte("changed"))
	if code := ErrorCode(ValidateTemplateDocumentDescriptorBinding(descriptor, Solution{PackageID: "drama", ID: "character", Version: "1.0.0"}, template)); code != CodeDigestMismatch {
		t.Fatalf("stale binding code = %s", code)
	}
}

func TestDecodeTemplateDocumentDescriptorRejectsUnknownAndDuplicateFields(t *testing.T) {
	payload := []byte(`{"schema_version":"promptrepo.template-document.v0.1","schema_version":"promptrepo.template-document.v0.1"}`)
	if _, err := DecodeTemplateDocumentDescriptor(payload); ErrorCode(err) != CodeDocumentDuplicateKey {
		t.Fatalf("duplicate descriptor field error = %v", err)
	}
	descriptor := testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", []byte(`{"task":{}}`))
	encoded, err := json.Marshal(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	encoded = bytes.Replace(encoded, []byte(`"digest":`), []byte(`"unknown":true,"digest":`), 1)
	if _, err := DecodeTemplateDocumentDescriptor(encoded); ErrorCode(err) != CodeDocumentDescriptorInvalid {
		t.Fatalf("unknown descriptor field error = %v", err)
	}
}

func TestStructuredDocumentJSONAndYAMLCanonicalDigestMatch(t *testing.T) {
	jsonBody := []byte(`{"count":1,"enabled":true,"task":{"id":"A","views":[{"name":"front"}]}}`)
	yamlBody := []byte("task:\n  id: A\n  views:\n    - name: front\nenabled: true\ncount: 1\n")
	jsonDocument, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", jsonBody), jsonBody)
	if err != nil {
		t.Fatal(err)
	}
	yamlDocument, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatYAML, "solutions/drama/character/prompts/main.zh-CN.yaml", yamlBody), yamlBody)
	if err != nil {
		t.Fatal(err)
	}
	if jsonDocument.CanonicalDigest != yamlDocument.CanonicalDigest {
		t.Fatalf("canonical digests differ: json=%s yaml=%s", jsonDocument.CanonicalDigest, yamlDocument.CanonicalDigest)
	}
	if jsonDocument.SourceDigest == yamlDocument.SourceDigest {
		t.Fatal("source digests unexpectedly match")
	}
}

func TestStructuredDocumentRejectsDuplicateKeysAndUnsafeYAML(t *testing.T) {
	jsonBody := []byte(`{"task":1,"task":2}`)
	if _, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", jsonBody), jsonBody); ErrorCode(err) != CodeDocumentDuplicateKey {
		t.Fatalf("duplicate JSON error = %v", err)
	}

	for name, body := range map[string][]byte{
		"duplicate": []byte("task: 1\ntask: 2\n"),
		"anchor":    []byte("task: &shared\n  id: A\ncopy: *shared\n"),
		"tag":       []byte("task: !custom A\n"),
		"map-key":   []byte("? [a, b]\n: value\n"),
		"nonfinite": []byte("count: .nan\n"),
	} {
		t.Run(name, func(t *testing.T) {
			_, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatYAML, "solutions/drama/character/prompts/main.zh-CN.yaml", body), body)
			if err == nil {
				t.Fatal("unsafe YAML was accepted")
			}
		})
	}
}

func TestDocumentFormatMismatchDepthAndUnicodeViolations(t *testing.T) {
	body := []byte(`{"task":{"nested":{"too":"deep"}}}`)
	descriptor := testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", body)
	descriptor.Limits.MaxDepth = 2
	resignDocumentDescriptor(t, &descriptor)
	if _, err := DecodeDocument(descriptor, body); ErrorCode(err) != CodeDocumentParseFailed {
		t.Fatalf("depth limit error = %v", err)
	}

	formatMismatch := testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", body)
	formatMismatch.TemplatePath = "solutions/drama/character/prompts/main.zh-CN.yaml"
	resignDocumentDescriptor(t, &formatMismatch)
	if err := ValidateTemplateDocumentDescriptor(formatMismatch); ErrorCode(err) != CodeDocumentFormatMismatch {
		t.Fatalf("format mismatch error = %v", err)
	}

	invalidUnicode := []byte(`{"value":"\ud800"}`)
	if _, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", invalidUnicode), invalidUnicode); ErrorCode(err) != CodeDocumentCanonicalizeFailed {
		t.Fatalf("invalid Unicode scalar error = %v", err)
	}
}

func TestJSONLLoadAndSelectors(t *testing.T) {
	body := []byte("{\"id\":\"front\",\"yaw\":0}\n{\"id\":\"back\",\"yaw\":180}\n")
	descriptor := testDocumentDescriptor(t, DocumentFormatJSONL, "solutions/drama/character/prompts/views.zh-CN.jsonl", body)
	loaded, err := DecodeDocument(descriptor, body)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.RecordCount != 2 || loaded.CanonicalDigest == "" {
		t.Fatalf("loaded JSONL metadata = %+v", loaded)
	}
	selected, err := SelectLoadedDocument(loaded, "jsonl-id:back")
	if err != nil {
		t.Fatal(err)
	}
	if selected.RecordID != "back" || selected.CanonicalDigest == "" {
		t.Fatalf("selected JSONL record = %+v", selected)
	}

	duplicate := []byte("{\"id\":\"front\"}\n{\"id\":\"front\"}\n")
	if _, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSONL, "solutions/drama/character/prompts/views.zh-CN.jsonl", duplicate), duplicate); ErrorCode(err) != CodeJSONLRecordDuplicate {
		t.Fatalf("duplicate JSONL record error = %v", err)
	}
	missingFinalNewline := []byte(`{"id":"front"}`)
	if _, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSONL, "solutions/drama/character/prompts/views.zh-CN.jsonl", missingFinalNewline), missingFinalNewline); ErrorCode(err) != CodeDocumentParseFailed {
		t.Fatalf("missing final newline error = %v", err)
	}

	tooLarge := []byte("{\"id\":\"front\",\"description\":\"too-large\"}\n")
	tooLargeDescriptor := testDocumentDescriptor(t, DocumentFormatJSONL, "solutions/drama/character/prompts/views.zh-CN.jsonl", tooLarge)
	tooLargeDescriptor.Limits.MaxRecordBytes = 16
	resignDocumentDescriptor(t, &tooLargeDescriptor)
	if _, err := DecodeDocument(tooLargeDescriptor, tooLarge); ErrorCode(err) != CodeJSONLRecordTooLarge {
		t.Fatalf("oversized JSONL record error = %v", err)
	}
}

func TestDocumentSelectorPointerHeadingAndDocumentProjectionBodyLeak(t *testing.T) {
	jsonBody := []byte(`{"task":{"views":[{"name":"front"},{"name":"back"}]},"secret":"BODY_SENTINEL_8f4c"}`)
	loaded, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatJSON, "solutions/drama/character/prompts/main.zh-CN.json", jsonBody), jsonBody)
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectLoadedDocument(loaded, "json-pointer:/task/views/1")
	if err != nil {
		t.Fatal(err)
	}
	selectedMap, ok := selected.Value.(map[string]any)
	if !ok || selectedMap["name"] != "back" {
		t.Fatalf("selected pointer value = %#v", selected.Value)
	}
	if len(selected.Document.Body) != 0 || selected.Document.Value != nil {
		t.Fatal("selected document retained the unselected source body or parsed node")
	}
	mutated := loaded
	mutated.Body = append([]byte(nil), loaded.Body...)
	mutated.Body[0] = '['
	if _, err := SelectLoadedDocument(mutated, "json-pointer:/task"); ErrorCode(err) != CodeDigestMismatch {
		t.Fatalf("mutated loaded body error = %v", err)
	}
	for _, value := range []any{loaded, selected} {
		jsonProjection, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		yamlProjection, err := yaml.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(jsonProjection), "BODY_SENTINEL_8f4c") || strings.Contains(string(yamlProjection), "BODY_SENTINEL_8f4c") {
			t.Fatalf("body leaked from safe projection: json=%s yaml=%s", jsonProjection, yamlProjection)
		}
	}

	markdown := []byte("# Intro\nhello\n## Detail\nkept\n# Outro\nbye\n")
	markdownDocument, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatMarkdown, "solutions/drama/character/prompts/main.zh-CN.md", markdown), markdown)
	if err != nil {
		t.Fatal(err)
	}
	heading, err := SelectLoadedDocument(markdownDocument, "heading:Intro")
	if err != nil {
		t.Fatal(err)
	}
	if string(heading.Body) != "# Intro\nhello\n## Detail\nkept\n" {
		t.Fatalf("heading selection = %q", heading.Body)
	}
	if _, err := SelectLoadedDocument(markdownDocument, "heading:Missing"); ErrorCode(err) != CodeSelectorNotFound {
		t.Fatalf("missing heading error = %v", err)
	}
	ambiguousMarkdown := []byte("# Intro\none\n# Intro\ntwo\n")
	ambiguousDocument, err := DecodeDocument(testDocumentDescriptor(t, DocumentFormatMarkdown, "solutions/drama/character/prompts/main.zh-CN.md", ambiguousMarkdown), ambiguousMarkdown)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := SelectLoadedDocument(ambiguousDocument, "heading:Intro"); ErrorCode(err) != CodeSelectorAmbiguous {
		t.Fatalf("ambiguous heading error = %v", err)
	}
}

func TestTemplateAddressAcceptsJSONLIDSelector(t *testing.T) {
	address := "promptrepo://official/drama/character@1.0.0?kind=template&locale=zh-CN&role=views&selector=jsonl-id%3Afront"
	parsed, err := ParseTemplateAddress(address)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Selector != "jsonl-id:front" {
		t.Fatalf("selector = %q", parsed.Selector)
	}
}

func testDocumentDescriptor(t *testing.T, format DocumentFormat, templatePath string, body []byte) TemplateDocumentDescriptor {
	t.Helper()
	mediaType, canonicalization, selectors, _, ok := documentFormatContract(format)
	if !ok {
		t.Fatalf("unsupported test format %q", format)
	}
	selectorKinds := make([]string, 0, len(selectors))
	for selector := range selectors {
		selectorKinds = append(selectorKinds, selector)
	}
	selectorKinds = canonicalStrings(selectorKinds)
	document := TemplateDocumentDescriptor{
		SchemaVersion: TemplateDocumentSchemaVersion,
		PackageID:     "drama", SolutionID: "character", Version: "1.0.0", Role: "main", Locale: "zh-CN",
		TemplatePath: templatePath, SourceDigest: digestBytes(body), Format: format, MediaType: mediaType,
		Canonicalization: canonicalization, SelectorKinds: selectorKinds,
		Limits: DocumentLimits{MaxBytes: 64 << 10, MaxDepth: 32},
	}
	if format == DocumentFormatJSONL {
		document.RecordIDField = "id"
		document.Limits.MaxRecordBytes = 16 << 10
		document.Limits.MaxRecords = 100
	}
	resignDocumentDescriptor(t, &document)
	return document
}

func resignDocumentDescriptor(t *testing.T, document *TemplateDocumentDescriptor) {
	t.Helper()
	digest, err := CanonicalTemplateDocumentDescriptorDigest(*document)
	if err != nil {
		t.Fatal(err)
	}
	document.Digest = digest
}
