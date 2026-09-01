// Package uitemplatefs provides a bounded local fixture loader for exact
// Promptrepo UI template bundles. It performs no network or execution work.
package uitemplatefs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/yeisme/promptrepo"
)

const maxDescriptorBytes = 256 << 10

type Loader struct {
	root string
}

var _ promptrepo.UITemplateInspector = (*Loader)(nil)
var _ promptrepo.UITemplateLoader = (*Loader)(nil)

func New(rootPath string) (*Loader, error) {
	if !filepath.IsAbs(rootPath) {
		return nil, promptrepo.NewError(promptrepo.CodeUITemplateAddressInvalid, "UI template fixture root must be absolute", false, nil)
	}
	root, err := filepath.EvalSymlinks(filepath.Clean(rootPath))
	if err != nil {
		return nil, promptrepo.NewError(promptrepo.CodeNotFound, "UI template fixture root is unavailable", false, err)
	}
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		return nil, promptrepo.NewError(promptrepo.CodeNotFound, "UI template fixture root is unavailable", false, err)
	}
	return &Loader{root: root}, nil
}

func (loader *Loader) InspectUITemplate(ctx context.Context, request promptrepo.InspectUITemplateRequest) (promptrepo.UITemplateInspectionV1, error) {
	bundle, err := loader.readBundle(ctx, request.Address)
	if err != nil {
		return promptrepo.UITemplateInspectionV1{}, err
	}
	inspection := promptrepo.NewUITemplateInspection(bundle)
	if !inspection.Validation.Valid {
		return inspection, promptrepo.ValidateUITemplateBundle(bundle)
	}
	return inspection, nil
}

func (loader *Loader) LoadUITemplate(ctx context.Context, request promptrepo.LoadUITemplateRequest) (promptrepo.UITemplateBundleV1, error) {
	bundle, err := loader.readBundle(ctx, request.Address)
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	if err := promptrepo.ValidateUITemplateBundle(bundle); err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	return bundle, nil
}

func (loader *Loader) readBundle(ctx context.Context, address promptrepo.UITemplateAddress) (promptrepo.UITemplateBundleV1, error) {
	if err := ctx.Err(); err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	if err := promptrepo.ValidateUITemplateAddress(address); err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	if address.Role == "" || address.Path == "" || address.Digest == "" || address.Snapshot == "" {
		return promptrepo.UITemplateBundleV1{}, promptrepo.NewError(promptrepo.CodeUITemplateAddressInvalid, "exact UI template load requires role, path, digest, and snapshot", false, nil)
	}
	descriptorPath, err := promptrepo.UITemplateFileDescriptorPath(address)
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	descriptorPayload, err := loader.readBounded(descriptorPath, maxDescriptorBytes, promptrepo.CodeUITemplateAddressInvalid, "UI template descriptor is unavailable or invalid")
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	descriptor, err := promptrepo.DecodeUITemplateFileDescriptorV1(descriptorPayload)
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	if descriptor.Address != address {
		return promptrepo.UITemplateBundleV1{}, promptrepo.NewError(promptrepo.CodeUITemplateAddressInvalid, "UI template descriptor identity does not match the requested address", false, nil)
	}
	htmlFragment, err := loader.readBounded(descriptor.HTMLPath, descriptor.Limits.MaxHTMLBytes, promptrepo.CodeUITemplateLimitExceeded, "UI template HTML is unavailable or exceeds its limit")
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	stylesheet, err := loader.readBounded(descriptor.CSSPath, descriptor.Limits.MaxCSSBytes, promptrepo.CodeUITemplateLimitExceeded, "UI template CSS is unavailable or exceeds its limit")
	if err != nil {
		return promptrepo.UITemplateBundleV1{}, err
	}
	if len(htmlFragment) != descriptor.HTMLBytes || len(stylesheet) != descriptor.CSSBytes || len(htmlFragment)+len(stylesheet) != descriptor.BodyBytes {
		return promptrepo.UITemplateBundleV1{}, promptrepo.NewError(promptrepo.CodeUITemplateDigestMismatch, "UI template body sizes do not match the descriptor", false, nil)
	}
	return descriptor.Bundle(htmlFragment, stylesheet), nil
}

func (loader *Loader) readBounded(relativePath string, maxBytes int, code, message string) ([]byte, error) {
	if maxBytes <= 0 || filepath.IsAbs(relativePath) || strings.Contains(relativePath, "\\") {
		return nil, promptrepo.NewError(code, message, false, nil)
	}
	candidate := filepath.Join(loader.root, filepath.FromSlash(relativePath))
	if !withinRoot(loader.root, candidate) {
		return nil, promptrepo.NewError(code, message, false, nil)
	}
	info, err := os.Lstat(candidate)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() || info.Size() > int64(maxBytes) {
		return nil, promptrepo.NewError(code, message, false, err)
	}
	resolved, err := filepath.EvalSymlinks(candidate)
	if err != nil || !withinRoot(loader.root, resolved) {
		return nil, promptrepo.NewError(code, message, false, err)
	}
	file, err := os.Open(resolved)
	if err != nil {
		return nil, promptrepo.NewError(code, message, false, err)
	}
	defer file.Close()
	payload, err := io.ReadAll(io.LimitReader(file, int64(maxBytes)+1))
	if err != nil || len(payload) > maxBytes {
		return nil, promptrepo.NewError(code, message, false, err)
	}
	return payload, nil
}

func withinRoot(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
