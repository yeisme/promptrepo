package promptrepo

import "time"

const (
	StateSchemaVersion   = "promptrepo.state.v0.1"
	CatalogSchemaVersion = "promptrepo.catalog.v0.1"
	ReceiptSchemaVersion = "promptrepo.stage_receipt.v0.1"
	RankingProfileV1     = "promptrepo.ranking.zh.v0.1"
)

type RepositoryProfile struct {
	ID               string    `json:"id"`
	Source           string    `json:"source"`
	SourceKind       string    `json:"source_kind"`
	Scope            string    `json:"scope"`
	Trust            string    `json:"trust"`
	Enabled          bool      `json:"enabled"`
	Revision         string    `json:"revision,omitempty"`
	Channel          string    `json:"channel,omitempty"`
	PreferredLocales []string  `json:"preferred_locales,omitempty"`
	CredentialRef    string    `json:"credential_ref,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type RepositoryHealth struct {
	State      string    `json:"state"`
	Code       string    `json:"code,omitempty"`
	Message    string    `json:"message,omitempty"`
	CheckedAt  time.Time `json:"checked_at,omitempty"`
	SnapshotAt time.Time `json:"snapshot_at,omitempty"`
}

type RepositoryView struct {
	Profile  RepositoryProfile `json:"profile"`
	Health   RepositoryHealth  `json:"health"`
	Snapshot *SnapshotMetadata `json:"snapshot,omitempty"`
}

type RepositoryMetadata struct {
	ID              string `json:"id"`
	Name            string `json:"name,omitempty"`
	DefaultLocale   string `json:"default_locale"`
	TaxonomyVersion string `json:"taxonomy_version"`
}

type LocalizedText struct {
	Title   string   `json:"title"`
	Summary string   `json:"summary"`
	Usage   string   `json:"usage,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type TemplateRole struct {
	Role   string `json:"role"`
	Locale string `json:"locale"`
	Path   string `json:"path"`
	Digest string `json:"digest"`
}

type Solution struct {
	RepositoryID string                   `json:"repository_id,omitempty"`
	PackageID    string                   `json:"package_id"`
	ID           string                   `json:"id"`
	Version      string                   `json:"version"`
	Digest       string                   `json:"digest"`
	Category     string                   `json:"category"`
	Tags         []string                 `json:"tags,omitempty"`
	Capabilities []string                 `json:"capabilities,omitempty"`
	Rights       string                   `json:"rights"`
	Maturity     string                   `json:"maturity"`
	Locales      map[string]LocalizedText `json:"locales"`
	Templates    []TemplateRole           `json:"templates,omitempty"`
}

type Catalog struct {
	SchemaVersion string             `json:"schema_version"`
	Repository    RepositoryMetadata `json:"repository"`
	GeneratedAt   time.Time          `json:"generated_at"`
	Digest        string             `json:"digest"`
	Solutions     []Solution         `json:"solutions"`
}

type Snapshot struct {
	RepositoryID string    `json:"repository_id"`
	Revision     string    `json:"revision"`
	Digest       string    `json:"digest"`
	FetchedAt    time.Time `json:"fetched_at"`
	Catalog      Catalog   `json:"catalog"`
}

type SnapshotMetadata struct {
	RepositoryID string    `json:"repository_id"`
	Revision     string    `json:"revision"`
	Digest       string    `json:"digest"`
	FetchedAt    time.Time `json:"fetched_at"`
}

type AddRepositoryRequest struct {
	Profile RepositoryProfile `json:"profile"`
}

type ListRepositoriesRequest struct{}

type RepositoryPage struct {
	Repositories []RepositoryView `json:"repositories"`
}

type SyncRequest struct {
	RepositoryIDs []string `json:"repository_ids,omitempty"`
	All           bool     `json:"all,omitempty"`
}

type RepositorySyncResult struct {
	RepositoryID  string `json:"repository_id"`
	State         string `json:"state"`
	Revision      string `json:"revision,omitempty"`
	Digest        string `json:"digest,omitempty"`
	SolutionCount int    `json:"solution_count,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	Message       string `json:"message,omitempty"`
}

type SyncReceipt struct {
	SchemaVersion string                 `json:"schema_version"`
	StartedAt     time.Time              `json:"started_at"`
	FinishedAt    time.Time              `json:"finished_at"`
	Results       []RepositorySyncResult `json:"results"`
}

type SearchRequest struct {
	Query                string   `json:"query"`
	Locale               string   `json:"locale,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
	IncludeIncompatible  bool     `json:"include_incompatible,omitempty"`
}

type SolutionCard struct {
	Ref            string   `json:"ref"`
	Title          string   `json:"title"`
	Summary        string   `json:"summary"`
	Locale         string   `json:"locale"`
	Category       string   `json:"category"`
	Tags           []string `json:"tags,omitempty"`
	Capabilities   []string `json:"capabilities,omitempty"`
	Rights         string   `json:"rights"`
	Maturity       string   `json:"maturity"`
	Trust          string   `json:"trust"`
	Health         string   `json:"health"`
	Compatible     bool     `json:"compatible"`
	Missing        []string `json:"missing_capabilities,omitempty"`
	MatchReasons   []string `json:"match_reasons,omitempty"`
	Score          int      `json:"score"`
	SnapshotDigest string   `json:"snapshot_digest"`
	SolutionDigest string   `json:"solution_digest"`
}

type SearchResult struct {
	RankingProfile  string         `json:"ranking_profile"`
	RequestedLocale string         `json:"requested_locale"`
	Results         []SolutionCard `json:"results"`
}

type ResolveRequest struct {
	Ref                  string   `json:"ref"`
	Locale               string   `json:"locale,omitempty"`
	RequiredCapabilities []string `json:"required_capabilities,omitempty"`
}

type ResolvedSolution struct {
	Ref        string           `json:"ref"`
	Solution   Solution         `json:"solution"`
	Locale     string           `json:"locale"`
	Display    LocalizedText    `json:"display"`
	Snapshot   SnapshotMetadata `json:"snapshot"`
	Trust      string           `json:"trust"`
	Health     string           `json:"health"`
	Compatible bool             `json:"compatible"`
	Missing    []string         `json:"missing_capabilities,omitempty"`
}

type StageRequest struct {
	ResolveRequest ResolveRequest `json:"resolve"`
	Consumer       string         `json:"consumer"`
}

type ReadTemplateRequest struct {
	Ref    string `json:"ref"`
	Locale string `json:"locale,omitempty"`
	Role   string `json:"role,omitempty"`
}

type TemplateContent struct {
	Ref      string           `json:"ref"`
	Role     string           `json:"role"`
	Locale   string           `json:"locale"`
	Path     string           `json:"path"`
	Digest   string           `json:"digest"`
	Body     string           `json:"body"`
	Snapshot SnapshotMetadata `json:"snapshot"`
}

type StageReceipt struct {
	SchemaVersion    string    `json:"schema_version"`
	ID               string    `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Consumer         string    `json:"consumer"`
	Ref              string    `json:"ref"`
	RepositoryID     string    `json:"repository_id"`
	SnapshotRevision string    `json:"snapshot_revision"`
	SnapshotDigest   string    `json:"snapshot_digest"`
	SolutionDigest   string    `json:"solution_digest"`
	Locale           string    `json:"locale"`
	Rights           string    `json:"rights"`
	Trust            string    `json:"trust"`
	Compatibility    string    `json:"compatibility"`
	State            string    `json:"state"`
	ProviderCalls    string    `json:"provider_calls"`
}
