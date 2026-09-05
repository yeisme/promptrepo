package promptrepo

import "testing"

func TestRecipeDependencyOrderAndCycle(t *testing.T) {
	r := RecipeV1{SchemaVersion: RecipeSchemaV1, ID: "story", Steps: []RecipeStepV1{
		{ID: "shots", Ref: "promptrepo://official/video/shots@1.0.0", Bindings: map[string]InputBindingV1{"script": {Kind: "step_output", Ref: "script"}}},
		{ID: "script", Ref: "promptrepo://official/writing/script@1.0.0"},
	}}
	order, err := RecipeOrder(r)
	if err != nil || len(order) != 2 || order[0] != "script" {
		t.Fatalf("dependency order: %v %v", order, err)
	}
	r.Steps[1].DependsOn = []string{"shots"}
	if _, err := RecipeOrder(r); err == nil {
		t.Fatal("cycle accepted")
	}
}

func TestPackagePortablePathsAndTampering(t *testing.T) {
	for _, bad := range []string{"../escape", "a/../escape", "C:/escape", "a\\b", "CON.txt", "AUX", "a/.", "a/space ", "/absolute"} {
		if SafePackagePath(bad) {
			t.Fatalf("unsafe path accepted: %q", bad)
		}
	}
	body := []byte("synthetic fixture")
	m := PromptPackageV1{SchemaVersion: PromptPackageSchemaV1, Compiler: "test-v1", Recipe: RecipeV1{SchemaVersion: RecipeSchemaV1, ID: "test", Steps: []RecipeStepV1{{ID: "main", Ref: "promptrepo://official/general/test@1.0.0"}}}, Steps: []PackageStepV1{{ID: "main", Status: "ready", Path: "prompts/main.txt"}}, Files: map[string]PackageFileV1{"prompts/main.txt": {Digest: PromptBytesDigest(body), Bytes: int64(len(body)), Kind: "prompt"}}}
	m.Digest, _ = PromptPackageDigest(m)
	if err := VerifyPromptPackageFiles(m, map[string][]byte{"prompts/main.txt": body}); err != nil {
		t.Fatal(err)
	}
	if err := VerifyPromptPackageFiles(m, map[string][]byte{"prompts/main.txt": []byte("tampered")}); err == nil {
		t.Fatal("tampering accepted")
	}
	m.Files["prompts/MAIN.txt"] = m.Files["prompts/main.txt"]
	m.Digest, _ = PromptPackageDigest(m)
	if err := ValidatePromptPackage(m); err == nil {
		t.Fatal("case-insensitive collision accepted")
	}
}
