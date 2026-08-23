package promptrepo

import (
	"bytes"
	"testing"
)

func TestTemplateContractCompanionPath(t *testing.T) {
	path, err := TemplateContractCompanionPath(TemplateRole{Role: "main", Locale: "zh-CN", Path: "solutions/audio/podcast/prompts/main.zh-CN.md"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "solutions/audio/podcast/contracts/main.zh-CN.json" {
		t.Fatalf("companion path = %q", path)
	}
	nested, err := TemplateContractCompanionPath(TemplateRole{Role: "system", Locale: "en", Path: "solutions/audio/podcast/prompts/system/main.en.md"})
	if err != nil || nested != "solutions/audio/podcast/contracts/system.en.json" {
		t.Fatalf("nested companion path = %q, err=%v", nested, err)
	}
	for _, invalid := range []string{
		"../prompts/main.zh-CN.md",
		"solutions/audio/podcast/../prompts/main.zh-CN.md",
		"solutions/audio/podcast/main.zh-CN.md",
		"/solutions/audio/podcast/prompts/main.zh-CN.md",
	} {
		if _, err := TemplateContractCompanionPath(TemplateRole{Role: "main", Locale: "zh-CN", Path: invalid}); err == nil {
			t.Fatalf("invalid template path accepted: %q", invalid)
		}
	}
}

func TestTemplateContractDocumentDigestAndUnknownField(t *testing.T) {
	document := TemplateContractDocument{
		SchemaVersion: TemplateContractSchemaVersion,
		PackageID:     "audio", SolutionID: "podcast", Version: "1.0.0", Role: "main", Locale: "zh-CN",
		TemplatePath:   "solutions/audio/podcast/prompts/main.zh-CN.md",
		TemplateDigest: "sha256:0000000000000000000000000000000000000000000000000000000000000000",
		License:        "internal", Permissions: []string{"inspect", "preview"},
		Inputs: []InputDefinition{{Name: "script", Type: InputTypeString, Required: true, Sensitivity: "sensitive"}},
	}
	digest, err := CanonicalTemplateContractDigest(document)
	if err != nil {
		t.Fatal(err)
	}
	document.Digest = digest
	if err := ValidateTemplateContractDocument(document); err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeTemplateContractDocument([]byte(`{"schema_version":"promptrepo.template-contract.v0.1","unknown":true}`)); ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unknown field error = %v", err)
	}
}

func TestTemplateContractDocumentRejectsOversizeAndInvalidUTF8(t *testing.T) {
	for _, payload := range [][]byte{
		bytes.Repeat([]byte{' '}, maxTemplateContractDocumentBytes+1),
		[]byte("{\"schema_version\":\"promptrepo.template-contract.v0.1\",\"license\":\"\xff\"}"),
	} {
		if _, err := DecodeTemplateContractDocument(payload); ErrorCode(err) != CodeInvalidRequest {
			t.Fatalf("unsafe companion payload error = %v", err)
		}
	}
}
