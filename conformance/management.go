package conformance

import "github.com/yeisme/promptrepo"

// MinimalRepositorySetRequest is the common embedded/service parity fixture.
func MinimalRepositorySetRequest() promptrepo.EffectiveRepositorySetRequest {
	return promptrepo.EffectiveRepositorySetRequest{
		Consumer: "conformance",
		Mode:     promptrepo.RepositoryModeEmbedded,
		Repositories: []promptrepo.RepositorySetCandidate{{
			RepositoryID: "fixture", Scope: promptrepo.RepositoryScopeOrganization,
			Trust: "official", Health: "ready", Enabled: true,
		}},
		Bindings: []promptrepo.RepositoryScopeBinding{{
			Scope: promptrepo.RepositoryScopeOrganization, ScopeRef: "fixture-organization",
			RepositoryIDs: []string{"fixture"}, PolicyRef: "fixture-policy", PolicyDigest: "sha256:fixture-policy", Enabled: true,
		}},
	}
}

// MinimalPolicyRequest is an allowed decision fixture for implementers of the
// optional PolicyEvaluator interface.
func MinimalPolicyRequest(repository promptrepo.RepositorySetEntry) promptrepo.EvaluatePolicyRequest {
	return promptrepo.EvaluatePolicyRequest{
		Repository: repository, OperationID: promptrepo.OperationTemplatePreview,
		Rights: "internal", Capabilities: []string{"text"}, RequiredCapabilities: []string{"text"},
		Policies: []promptrepo.PolicyConstraint{{
			Scope: promptrepo.RepositoryScopeOrganization, Ref: "fixture-policy", Digest: "sha256:fixture-policy",
			MinimumTrust: "verified", AllowedRights: []string{"internal"}, AllowedOperations: []string{promptrepo.OperationTemplatePreview},
		}},
	}
}
