package promptrepo_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/conformance"
	"gopkg.in/yaml.v3"
)

func TestManagementContractOperationIDs(t *testing.T) {
	t.Parallel()

	if promptrepo.RepositorySetSchemaVersion != "promptrepo.repository-set.v0.1" ||
		promptrepo.PolicyDecisionSchemaVersion != "promptrepo.policy-decision.v0.1" ||
		promptrepo.ManagementProjectionSchemaVersion != "promptrepo.management-projection.v0.1" {
		t.Fatal("management schema versions changed")
	}
	operations := []string{
		promptrepo.OperationRepositoryList,
		promptrepo.OperationRepositorySync,
		promptrepo.OperationCatalogSearch,
		promptrepo.OperationTemplateInspect,
		promptrepo.OperationTemplatePreview,
	}
	for _, operation := range operations {
		if !strings.HasPrefix(operation, "promptrepo.") {
			t.Fatalf("operation id is not stable promptrepo identity: %q", operation)
		}
	}
}

func TestManagementConformanceFixtureHasEmbeddedServiceParity(t *testing.T) {
	t.Parallel()

	embeddedRequest := conformance.MinimalRepositorySetRequest()
	embedded, err := promptrepo.ComposeRepositorySet(embeddedRequest)
	if err != nil {
		t.Fatal(err)
	}
	serviceRequest := embeddedRequest
	serviceRequest.Mode = promptrepo.RepositoryModeService
	service, err := promptrepo.ComposeRepositorySet(serviceRequest)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(embedded.Layers, service.Layers) ||
		!reflect.DeepEqual(embedded.EffectiveRepositories, service.EffectiveRepositories) ||
		embedded.PolicyDigest != service.PolicyDigest ||
		!reflect.DeepEqual(embedded.HealthSummary, service.HealthSummary) {
		t.Fatalf("embedded/service repository-set semantics diverged:\nembedded=%+v\nservice=%+v", embedded, service)
	}
	decision, err := promptrepo.EvaluateRepositoryPolicy(conformance.MinimalPolicyRequest(embedded.EffectiveRepositories[0]))
	if err != nil {
		t.Fatal(err)
	}
	if !decision.Allowed || decision.Readiness != promptrepo.ReadinessReady {
		t.Fatalf("management conformance policy fixture is not allowed: %+v", decision)
	}
}

func TestRepositorySetPreferenceOrderAndDeterminism(t *testing.T) {
	t.Parallel()

	request := promptrepo.EffectiveRepositorySetRequest{
		Consumer:          "scaena",
		Mode:              promptrepo.RepositoryModeEmbedded,
		ExactRepositoryID: "exact",
		Repositories: []promptrepo.RepositorySetCandidate{
			{RepositoryID: "organization", Scope: promptrepo.RepositoryScopeOrganization, Trust: "verified", Health: "ready", Enabled: true},
			{RepositoryID: "user", Scope: promptrepo.RepositoryScopeUser, Trust: "user_trusted", Health: "ready", Enabled: true},
			{RepositoryID: "project", Scope: promptrepo.RepositoryScopeOrganization, Trust: "verified", Health: "ready", Enabled: true},
			{RepositoryID: "exact", Scope: promptrepo.RepositoryScopeOrganization, Trust: "official", Health: "quarantined", Enabled: true},
			{RepositoryID: "official", Scope: promptrepo.RepositoryScopeOrganization, Trust: "official", Health: "ready", Enabled: true},
		},
		Bindings: []promptrepo.RepositoryScopeBinding{
			{Scope: promptrepo.RepositoryScopeOrganization, ScopeRef: "private-organization-id", RepositoryIDs: []string{"organization"}, PolicyDigest: "sha256:org", Enabled: true},
			{Scope: promptrepo.RepositoryScopeUser, ScopeRef: "private-user-id", RepositoryIDs: []string{"user"}, Priority: 5, Enabled: true},
			{Scope: promptrepo.RepositoryScopeProject, ScopeRef: "private-project-id", RepositoryIDs: []string{"project"}, PolicyDigest: "sha256:project", Enabled: true},
		},
		OfficialFallbackRepositoryIDs: []string{"official"},
	}

	first, err := promptrepo.ComposeRepositorySet(request)
	if err != nil {
		t.Fatal(err)
	}
	second, err := promptrepo.ComposeRepositorySet(request)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repository set is not deterministic:\nfirst=%+v\nsecond=%+v", first, second)
	}
	want := []string{"exact", "project", "user", "organization", "official"}
	got := make([]string, 0, len(first.EffectiveRepositories))
	for _, entry := range first.EffectiveRepositories {
		got = append(got, entry.RepositoryID)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("preference order changed: got %v want %v", got, want)
	}
	if first.EffectiveRepositories[0].SelectionSource != promptrepo.SelectionSessionExact || first.EffectiveRepositories[0].Health != "quarantined" {
		t.Fatalf("exact selection should rank without changing health: %+v", first.EffectiveRepositories[0])
	}
	if first.PolicyDigest == "" || first.HealthSummary.Quarantined != 1 || first.HealthSummary.Ready != 4 {
		t.Fatalf("repository set summary is incomplete: %+v", first)
	}
	payload, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	for _, privateValue := range []string{"private-organization-id", "private-user-id", "private-project-id"} {
		if strings.Contains(string(payload), privateValue) {
			t.Fatalf("raw scope ref leaked: %s", payload)
		}
	}
}

func TestPolicyDenyWinsOverExactAndProjectAllow(t *testing.T) {
	t.Parallel()

	entry := promptrepo.RepositorySetEntry{
		RepositoryID: "restricted", Scope: promptrepo.RepositoryScopeSession, RepositoryScope: promptrepo.RepositoryScopeOrganization,
		SelectionSource: promptrepo.SelectionSessionExact, Trust: "official", Health: "ready", Enabled: true,
	}
	decision, err := promptrepo.EvaluateRepositoryPolicy(promptrepo.EvaluatePolicyRequest{
		Repository: entry, OperationID: promptrepo.OperationTemplatePreview, Rights: "internal", Capabilities: []string{"image"},
		Policies: []promptrepo.PolicyConstraint{
			{Scope: promptrepo.RepositoryScopeOrganization, Ref: "org-policy", Digest: "sha256:org-policy", DeniedRepositoryIDs: []string{"restricted"}},
			{Scope: promptrepo.RepositoryScopeProject, Ref: "project-policy", Digest: "sha256:project-policy", AllowedRepositoryIDs: []string{"restricted"}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || decision.Readiness != promptrepo.ReadinessBlocked || !contains(decision.ReasonCodes, promptrepo.ReasonRepositoryDenied) {
		t.Fatalf("project allow or exact selection bypassed deny: %+v", decision)
	}
	if !reflect.DeepEqual(decision.MatchedPolicyRefs, []string{"org-policy", "project-policy"}) {
		t.Fatalf("policy lineage is unstable: %+v", decision.MatchedPolicyRefs)
	}
	if decision.PolicyDigest == "" {
		t.Fatal("combined policy digest is missing")
	}
}

func TestExactQuarantinedRepositoryRemainsBlocked(t *testing.T) {
	t.Parallel()

	set, err := promptrepo.ComposeRepositorySet(promptrepo.EffectiveRepositorySetRequest{
		Consumer: "scaena", ExactRepositoryID: "exact",
		Repositories: []promptrepo.RepositorySetCandidate{{
			RepositoryID: "exact", Scope: promptrepo.RepositoryScopeOrganization, Trust: "official", Health: "quarantined", Enabled: true,
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	decision, err := promptrepo.EvaluateRepositoryPolicy(promptrepo.EvaluatePolicyRequest{
		Repository: set.EffectiveRepositories[0], OperationID: promptrepo.OperationTemplatePreview,
	})
	if err != nil {
		t.Fatal(err)
	}
	if decision.Allowed || !contains(decision.ReasonCodes, promptrepo.ReasonSourceQuarantined) {
		t.Fatalf("exact selection bypassed quarantine: %+v", decision)
	}
}

func TestPolicyEvaluatesTrustRightsCapabilityAndOperation(t *testing.T) {
	t.Parallel()

	decision, err := promptrepo.EvaluateRepositoryPolicy(promptrepo.EvaluatePolicyRequest{
		Repository:           promptrepo.RepositorySetEntry{RepositoryID: "community", Scope: promptrepo.RepositoryScopeProject, Trust: "untrusted", Health: "ready", Enabled: true},
		OperationID:          promptrepo.OperationTemplatePreview,
		Rights:               "prohibited",
		Capabilities:         []string{"text"},
		RequiredCapabilities: []string{"image", "text"},
		Policies: []promptrepo.PolicyConstraint{{
			Scope: promptrepo.RepositoryScopeOrganization, Ref: "strict", MinimumTrust: "verified",
			AllowedOperations: []string{promptrepo.OperationTemplateInspect}, AllowedRights: []string{"internal"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	wantReasons := []string{
		promptrepo.ReasonCapabilityMissing,
		promptrepo.ReasonOperationNotAllowed,
		promptrepo.ReasonRightsBlocked,
		promptrepo.ReasonRightsNotAllowed,
		promptrepo.ReasonTrustBelowMinimum,
	}
	if !reflect.DeepEqual(decision.ReasonCodes, wantReasons) {
		t.Fatalf("policy reasons changed: got %v want %v", decision.ReasonCodes, wantReasons)
	}
	if !reflect.DeepEqual(decision.MissingCapabilities, []string{"image"}) || decision.MinimumTrust != "verified" {
		t.Fatalf("policy detail changed: %+v", decision)
	}
	if decision.RightsResult != promptrepo.PolicyResultBlocked || decision.CapabilityResult != promptrepo.PolicyResultBlocked {
		t.Fatalf("policy result projection changed: %+v", decision)
	}
}

func TestManagementProjectionIsBodyAndCredentialFree(t *testing.T) {
	t.Parallel()

	entry := promptrepo.RepositorySetEntry{
		RepositoryID: "official", Scope: promptrepo.RepositoryScopeOrganization, RepositoryScope: promptrepo.RepositoryScopeOrganization,
		ScopeRefDigest: "sha256:scope", Trust: "official", Health: "ready", Enabled: true,
	}
	decision := promptrepo.PolicyDecision{SchemaVersion: promptrepo.PolicyDecisionSchemaVersion, Allowed: true, Readiness: promptrepo.ReadinessReady, PolicyDigest: "sha256:decision-policy", RightsResult: promptrepo.PolicyResultAllowed, CapabilityResult: promptrepo.PolicyResultAllowed}
	projection, err := promptrepo.BuildManagementProjection(promptrepo.ManagementProjectionRequest{
		OperationID: promptrepo.OperationTemplatePreview, Command: "prompt-asset catalog preview", Mode: promptrepo.RepositoryModeEmbedded,
		Repository: entry, Decision: decision, SnapshotDigest: "sha256:snapshot",
		NextActions: []promptrepo.ManagementAction{{ActionID: "eikona.catalog.install", Owner: "eikona", Risk: "local_write", Writes: "owner_store"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	jsonPayload, err := json.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	yamlPayload, err := yaml.Marshal(projection)
	if err != nil {
		t.Fatal(err)
	}
	for _, payload := range [][]byte{jsonPayload, yamlPayload} {
		lower := strings.ToLower(string(payload))
		for _, forbidden := range []string{"template_body", "rendered_body", "input_values", "credential_ref", "credential_value", "authorization", "source_url", "local_path", "provider_payload", "chain_of_thought"} {
			if strings.Contains(lower, forbidden) {
				t.Fatalf("management projection leaked forbidden field %q: %s", forbidden, payload)
			}
		}
	}
	if projection.Status != "ok" || projection.OperationID != promptrepo.OperationTemplatePreview || projection.ProviderCalls || projection.DurableWrites {
		t.Fatalf("management projection changed: %+v", projection)
	}
	if projection.PolicyDigest != decision.PolicyDigest {
		t.Fatalf("management projection lost policy digest: %+v", projection)
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
