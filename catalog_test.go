package promptrepo_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/conformance"
	"github.com/yeisme/promptrepo/engine"
)

func TestCatalogDigestIsFrozenWithoutOptionalFields(t *testing.T) {
	catalog := conformance.MinimalCatalog()
	digest, err := promptrepo.CanonicalCatalogDigest(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if digest != privateV010CatalogDigest {
		t.Fatalf("catalog digest drifted: %s", digest)
	}
	payload, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"inputs", "license", "permissions"} {
		if strings.Contains(string(payload), `"`+field+`"`) {
			t.Fatalf("absent optional field %q was serialized: %s", field, payload)
		}
	}
}

func TestTemplateInputDefinitionValidation(t *testing.T) {
	minimum, maximum, minLength := 1.0, 5.0, 2
	valid := []promptrepo.InputDefinition{
		{Name: "subject", Type: promptrepo.InputTypeString, Required: true, Regex: `^[A-Za-z]+$`, MinLength: &minLength, Labels: map[string]string{"en": "Subject"}, Descriptions: map[string]string{"zh-CN": "主题"}},
		{Name: "count", Type: promptrepo.InputTypeInteger, Min: &minimum, Max: &maximum, Default: 2},
		{Name: "tone", Type: promptrepo.InputTypeEnum, Enum: []any{"calm", "bright"}, Default: "calm"},
		{Name: "private_key", Type: promptrepo.InputTypeString, Sensitivity: "sensitive"},
	}
	if err := promptrepo.ValidateTemplateInputs(valid); err != nil {
		t.Fatal(err)
	}
	for _, inputs := range [][]promptrepo.InputDefinition{
		{{Name: "not-valid", Type: promptrepo.InputTypeString}},
		{{Name: "tone", Type: promptrepo.InputTypeEnum}},
		{{Name: "secret", Type: promptrepo.InputTypeString, Sensitivity: "sensitive", Default: "nope"}},
		{{Name: "count", Type: promptrepo.InputTypeInteger, MinLength: &minLength}},
	} {
		if err := promptrepo.ValidateTemplateInputs(inputs); err == nil {
			t.Fatalf("expected invalid input schema: %+v", inputs)
		}
	}
}

func TestTemplateContractDigestValidation(t *testing.T) {
	for _, contract := range []promptrepo.TemplateContract{
		{},
		{Digest: "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", License: "internal", Permissions: []string{"preview"}},
		{License: "MIT OR Apache-2.0"},
	} {
		if err := promptrepo.ValidateTemplateContract(contract); err != nil {
			t.Fatalf("valid contract rejected: %v", err)
		}
	}
	for _, digest := range []string{"sha256:ABC", "sha256:0123", "sha512:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef", "not-a-digest"} {
		if err := promptrepo.ValidateTemplateContract(promptrepo.TemplateContract{Digest: digest}); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
			t.Fatalf("malformed contract digest %q: %v", digest, err)
		}
	}
	for _, contract := range []promptrepo.TemplateContract{
		{License: "https://license.example/private"},
		{License: " internal"},
		{License: "internal\nprivate"},
		{Permissions: []string{"preview", "inspect"}},
		{Permissions: []string{"preview", "preview"}},
		{Permissions: []string{"preview\nsecret"}},
	} {
		if err := promptrepo.ValidateTemplateContract(contract); promptrepo.ErrorCode(err) != promptrepo.CodeInvalidRequest {
			t.Fatalf("unsafe contract projection accepted: %+v, err=%v", contract, err)
		}
	}
}

func TestManagerStillImplementsExistingClient(t *testing.T) {
	var _ promptrepo.Client = (*engine.Manager)(nil)
}

func TestReleasedCatalogStructsKeepUnkeyedCompositeCompatibility(t *testing.T) {
	_ = promptrepo.TemplateRole{"main", "zh-CN", "prompts/main.md", "sha256:template"}
	_ = promptrepo.Solution{"official", "audio", "podcast", "1.0.0", "sha256:solution", "audio", nil, nil, "internal", "first-support", map[string]promptrepo.LocalizedText{}, nil}
}

func TestRenderTemplateOmitsBodyFromJSON(t *testing.T) {
	result, err := promptrepo.RenderTemplate(promptrepo.RenderRequest{Template: "Hello {{name}}", Inputs: []promptrepo.InputDefinition{{Name: "name", Type: promptrepo.InputTypeString, Required: true}}, Values: map[string]any{"name": "Ada"}})
	if err != nil || !result.Ready || result.RenderedBody != "Hello Ada" {
		t.Fatalf("render result metadata invalid: ready=%t err=%v", result.Ready, err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), "Hello Ada") || strings.Contains(string(payload), "Ada") {
		t.Fatalf("render body leaked into JSON: %s", payload)
	}
	field, ok := reflect.TypeOf(promptrepo.RenderResult{}).FieldByName("RenderedBody")
	if !ok || field.Tag.Get("json") != "-" || field.Tag.Get("yaml") != "-" {
		t.Fatal("rendered body must be omitted by both JSON and YAML contracts")
	}
	previewField, ok := reflect.TypeOf(promptrepo.PreviewResult{}).FieldByName("RenderedBody")
	if !ok || previewField.Tag.Get("json") != "-" || previewField.Tag.Get("yaml") != "-" {
		t.Fatal("preview body must be omitted by both JSON and YAML contracts")
	}
}

func TestRenderTemplateRendersMissingOptionalPlaceholderAsEmpty(t *testing.T) {
	result, err := promptrepo.RenderTemplate(promptrepo.RenderRequest{
		Template: "Subject {{subject}} ref={{character_ref}} ending={{ending}}",
		Inputs: []promptrepo.InputDefinition{
			{Name: "subject", Type: promptrepo.InputTypeString, Required: true},
			{Name: "character_ref", Type: promptrepo.InputTypeString, Regex: `^@[A-Za-z0-9]+$`},
			{Name: "ending", Type: promptrepo.InputTypeString},
		},
		Values: map[string]any{"subject": "Ada"},
	})
	if err != nil || !result.Ready {
		t.Fatalf("optional placeholders should render: ready=%t err=%v issues=%+v", result.Ready, err, result.Issues)
	}
	if result.RenderedBody != "Subject Ada ref= ending=" {
		t.Fatalf("optional placeholders were not empty: %q", result.RenderedBody)
	}
	if len(result.Inputs) != 3 || result.Inputs[1].Status != "missing" || result.Inputs[2].Status != "missing" {
		t.Fatalf("optional input status changed: %+v", result.Inputs)
	}
}

func TestRenderTemplateStillRejectsUndeclaredPlaceholder(t *testing.T) {
	result, err := promptrepo.RenderTemplate(promptrepo.RenderRequest{
		Template: "Hello {{undeclared}}",
		Inputs:   []promptrepo.InputDefinition{},
		Values:   map[string]any{},
	})
	if err != nil || result.Ready || len(result.Issues) != 1 || result.Issues[0].Code != promptrepo.CodeTemplatePlaceholder {
		t.Fatalf("undeclared placeholder must fail closed: ready=%t err=%v issues=%+v", result.Ready, err, result.Issues)
	}
}

func TestRenderTemplateRejectsMalformedBraceRuns(t *testing.T) {
	inputs := []promptrepo.InputDefinition{{Name: "name", Type: promptrepo.InputTypeString, Required: true}}
	for _, body := range []string{"{{{name}}}", "{{name}}}", "prefix }} suffix", "{{name}"} {
		result, err := promptrepo.RenderTemplate(promptrepo.RenderRequest{Template: body, Inputs: inputs, Values: map[string]any{"name": "Ada"}})
		if err != nil || result.Ready || len(result.Issues) != 1 || result.Issues[0].Code != promptrepo.CodeTemplateSyntax {
			t.Fatalf("expected strict syntax error for %q: ready=%t err=%v issues=%d", body, result.Ready, err, len(result.Issues))
		}
	}
	result, err := promptrepo.RenderTemplate(promptrepo.RenderRequest{Template: "literal { brace } and {{name}}", Inputs: inputs, Values: map[string]any{"name": "Ada"}})
	if err != nil || !result.Ready || result.RenderedBody != "literal { brace } and Ada" {
		t.Fatalf("single braces should remain literal: ready=%t err=%v", result.Ready, err)
	}
}
