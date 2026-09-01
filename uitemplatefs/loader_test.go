package uitemplatefs_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yeisme/promptrepo"
	"github.com/yeisme/promptrepo/uitemplatefs"
)

func TestLoadUITemplateAndInspectUITemplateBodySafeBundle(t *testing.T) {
	const sentinel = "PRIVATE_FILESYSTEM_TEMPLATE_SENTINEL"
	bundle := filesystemBundle(t, sentinel)
	root := writeFilesystemBundle(t, bundle)
	loader, err := uitemplatefs.New(root)
	if err != nil {
		t.Fatal(err)
	}

	loaded, err := loader.LoadUITemplate(context.Background(), promptrepo.LoadUITemplateRequest{Address: bundle.Address})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(loaded.HTMLFragment), sentinel) || loaded.ContentDigest != bundle.ContentDigest {
		t.Fatalf("unexpected loaded bundle: digest=%s", loaded.ContentDigest)
	}

	inspection, err := loader.InspectUITemplate(context.Background(), promptrepo.InspectUITemplateRequest{Address: bundle.Address})
	if err != nil {
		t.Fatal(err)
	}
	if !inspection.Validation.Valid || inspection.HTMLBytes != len(bundle.HTMLFragment) || inspection.CSSBytes != len(bundle.CSS) {
		t.Fatalf("unexpected inspection: %+v", inspection)
	}
	payload, err := json.Marshal(inspection)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), sentinel) || strings.Contains(string(payload), root) {
		t.Fatalf("inspection leaked body or absolute root: %s", payload)
	}
}

func TestLoadUITemplateRejectsTamperingAndSymlinkEscape(t *testing.T) {
	bundle := filesystemBundle(t, "ORIGINAL")
	root := writeFilesystemBundle(t, bundle)
	loader, err := uitemplatefs.New(root)
	if err != nil {
		t.Fatal(err)
	}
	htmlPath := filepath.Join(root, filepath.FromSlash(bundle.Address.Path))
	if err := os.WriteFile(htmlPath, []byte(`<section><div data-promptrepo-slot="title">changed</div></section>`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loader.LoadUITemplate(context.Background(), promptrepo.LoadUITemplateRequest{Address: bundle.Address}); promptrepo.ErrorCode(err) != promptrepo.CodeUITemplateDigestMismatch {
		t.Fatalf("tampered body: %v", err)
	}

	outside := filepath.Join(t.TempDir(), "outside.html")
	if err := os.WriteFile(outside, bundle.HTMLFragment, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(htmlPath); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, htmlPath); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := loader.LoadUITemplate(context.Background(), promptrepo.LoadUITemplateRequest{Address: bundle.Address}); promptrepo.ErrorCode(err) != promptrepo.CodeUITemplateLimitExceeded {
		t.Fatalf("symlink body: %v", err)
	}
}

func TestUITemplateRedactionLoaderErrorsDoNotExposePrivateRootOrBody(t *testing.T) {
	const sentinel = "DO_NOT_LOG_TEMPLATE_BODY"
	bundle := filesystemBundle(t, sentinel)
	root := writeFilesystemBundle(t, bundle)
	if err := os.Remove(filepath.Join(root, filepath.FromSlash(bundle.Address.Path))); err != nil {
		t.Fatal(err)
	}
	loader, err := uitemplatefs.New(root)
	if err != nil {
		t.Fatal(err)
	}
	_, err = loader.LoadUITemplate(context.Background(), promptrepo.LoadUITemplateRequest{Address: bundle.Address})
	if err == nil || strings.Contains(err.Error(), root) || strings.Contains(err.Error(), sentinel) {
		t.Fatalf("loader error leaked private data: %v", err)
	}
}

func filesystemBundle(t *testing.T, sentinel string) promptrepo.UITemplateBundleV1 {
	t.Helper()
	snapshot := filesystemDigest("fixture-snapshot")
	bundle := promptrepo.UITemplateBundleV1{
		SchemaVersion: promptrepo.UITemplateSchemaVersionV1,
		Address: promptrepo.UITemplateAddress{
			RepositoryID: "official",
			PackageID:    "scaena",
			SolutionID:   "storyboard-review",
			Version:      "1.0.0",
			Locale:       "zh-CN",
			Role:         "review",
			Path:         "ui/review.zh-CN.html",
			Snapshot:     snapshot,
		},
		HTMLFragment: []byte(`<section class="review">` + sentinel + `<div data-promptrepo-slot="title"></div></section>`),
		CSS:          []byte(`.review { display: grid; color: #222; }`),
		Slots: []promptrepo.UITemplateSlotV1{{
			Name: "title", Kind: promptrepo.UITemplateSlotKindText, Required: true, Cardinality: promptrepo.UITemplateSlotCardinalityOne,
		}},
		Security: promptrepo.UITemplateSecurityStaticReviewFragmentV1,
		Limits:   promptrepo.DefaultUITemplateLimitsV1(),
		Snapshot: snapshot,
	}
	digest, err := promptrepo.CanonicalUITemplateDigest(bundle)
	if err != nil {
		t.Fatal(err)
	}
	bundle.ContentDigest = digest
	bundle.Address.Digest = digest
	return bundle
}

func writeFilesystemBundle(t *testing.T, bundle promptrepo.UITemplateBundleV1) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "ui"), 0o700); err != nil {
		t.Fatal(err)
	}
	cssPath := "ui/review.zh-CN.css"
	for relative, payload := range map[string][]byte{bundle.Address.Path: bundle.HTMLFragment, cssPath: bundle.CSS} {
		if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(relative)), payload, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	descriptor := promptrepo.UITemplateFileDescriptorV1{
		SchemaVersion: bundle.SchemaVersion,
		Address:       bundle.Address,
		HTMLPath:      bundle.Address.Path,
		CSSPath:       cssPath,
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
	descriptorPath, err := promptrepo.UITemplateFileDescriptorPath(bundle.Address)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, filepath.FromSlash(descriptorPath)), payload, 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func filesystemDigest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return "sha256:" + hex.EncodeToString(sum[:])
}
