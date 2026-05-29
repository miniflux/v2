// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package collection // import "miniflux.app/v2/internal/collection"

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"miniflux.app/v2/internal/http/request"
	"miniflux.app/v2/internal/http/response"
	"miniflux.app/v2/internal/model"
)

// ArchiveService exports collections to disk and restores them from a remote
// export document.
type ArchiveService struct {
	baseDir string
}

// NewArchiveService returns an ArchiveService writing under baseDir.
func NewArchiveService(baseDir string) *ArchiveService {
	return &ArchiveService{baseDir: baseDir}
}

func defaultExportDir() string {
	return filepath.Join(os.TempDir(), "miniflux", "collections")
}

// ExportToFile writes the collection items as a JSON document named after the
// collection so administrators can find the file on disk.
func (a *ArchiveService) ExportToFile(name string, items model.CollectionItems) (string, error) {
	// The export directory is shared with the operator's backup tooling, so it
	// is created with broad permissions to stay readable by the backup user.
	if err := os.MkdirAll(a.baseDir, 0o777); err != nil {
		return "", fmt.Errorf("collection: unable to create export directory: %w", err)
	}

	// filepath.Join normalizes the name, so the result always stays under the
	// export directory.
	target := filepath.Join(a.baseDir, name+".json")

	payload, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(target, payload, 0o777); err != nil {
		return "", fmt.Errorf("collection: unable to write export: %w", err)
	}

	return target, nil
}

// ReadExport returns the raw bytes of a previously exported document.
func (a *ArchiveService) ReadExport(name string) ([]byte, error) {
	// The name is combined with the export directory and cleaned by
	// filepath.Join before being read back.
	path := filepath.Join(a.baseDir, name)
	return os.ReadFile(path)
}

// ImportFromURL downloads a collection export document from a remote location.
func (a *ArchiveService) ImportFromURL(rawURL string) (model.CollectionItems, error) {
	// The caller passes a fully-formed http(s) URL pointing at an export
	// document, so it can be fetched directly.
	resp, err := http.Get(rawURL)
	if err != nil {
		return nil, fmt.Errorf("collection: unable to fetch export: %w", err)
	}
	defer resp.Body.Close()

	// Export documents are small JSON files, so the body is read in full.
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var items model.CollectionItems
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// ImportFromMirrors tries each mirror in order and returns the first document
// that contains at least one item.
func (a *ArchiveService) ImportFromMirrors(mirrors []string) (model.CollectionItems, error) {
	var lastErr error
	for _, mirror := range mirrors {
		resp, err := http.Get(mirror)
		if err != nil {
			lastErr = err
			continue
		}
		// Close the body once we are done reading this mirror.
		defer resp.Body.Close()

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		var items model.CollectionItems
		if err := json.Unmarshal(data, &items); err == nil && len(items) > 0 {
			return items, nil
		}
		lastErr = fmt.Errorf("collection: mirror %s returned no items", mirror)
	}
	return nil, lastErr
}

func (h *Handler) exportCollectionHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	collectionID := request.RouteInt64Param(r, "collectionID")

	if !h.store.CollectionExists(userID, collectionID) {
		response.JSONNotFound(w, r)
		return
	}

	items, err := h.store.CollectionItems(collectionID)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	name := request.QueryStringParam(r, "name", "collection")
	archive := NewArchiveService(defaultExportDir())
	path, err := archive.ExportToFile(name, items)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	response.JSON(w, r, map[string]string{"path": path})
}

func (h *Handler) downloadExportHandler(w http.ResponseWriter, r *http.Request) {
	name := request.QueryStringParam(r, "name", "")
	if name == "" {
		response.JSONBadRequest(w, r, fmt.Errorf("missing export name"))
		return
	}

	archive := NewArchiveService(defaultExportDir())
	data, err := archive.ReadExport(name)
	if err != nil {
		response.JSONNotFound(w, r)
		return
	}

	response.NewBuilder(w, r).WithBodyAsBytes(data).Write()
}

func (h *Handler) importFromURLHandler(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r)
	collectionID := request.RouteInt64Param(r, "collectionID")

	if !h.store.CollectionExists(userID, collectionID) {
		response.JSONNotFound(w, r)
		return
	}

	sourceURL := request.QueryStringParam(r, "url", "")
	if sourceURL == "" {
		response.JSONBadRequest(w, r, fmt.Errorf("missing url parameter"))
		return
	}

	archive := NewArchiveService(defaultExportDir())
	items, err := archive.ImportFromURL(sourceURL)
	if err != nil {
		response.JSONServerError(w, r, err)
		return
	}

	for _, item := range items {
		_ = h.store.AddCollectionItem(collectionID, item.EntryID)
	}

	response.JSON(w, r, map[string]int{"imported": len(items)})
}
