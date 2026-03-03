// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinchtab // import "miniflux.app/v2/internal/reader/pinchtab"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"miniflux.app/v2/internal/config"
)

var (
	instance *subprocess
	once     sync.Once
)

const (
	httpRequestTimeout  = 30 * time.Second
	healthCheckInterval = 1 * time.Second
	healthCheckMaxWait  = 30 * time.Second
)

type subprocess struct {
	cmd *exec.Cmd
	mu  sync.Mutex
	// running tracks whether the subprocess is alive to prevent
	// use-after-exit calls to the pinchtab API.
	running bool
}

type createInstanceResponse struct {
	ID string `json:"id"`
}

type createTabResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// RenderPage creates a browser instance, opens a tab with the given URL,
// extracts the text content, and cleans up the instance.
func RenderPage(pageURL string) (string, error) {
	if instance == nil || !instance.isRunning() {
		return "", fmt.Errorf("pinchtab: subprocess is not running")
	}

	baseURL := config.Opts.PinchTabURL()
	client := &http.Client{Timeout: httpRequestTimeout}

	// Step 1: Create a browser instance.
	instanceID, err := createInstance(client, baseURL)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to create instance: %w", err)
	}

	// Always clean up the instance after use.
	defer func() {
		if deleteErr := deleteInstance(client, baseURL, instanceID); deleteErr != nil {
			slog.Warn("pinchtab: failed to delete instance",
				slog.String("instance_id", instanceID),
				slog.Any("error", deleteErr),
			)
		}
	}()

	// Step 2: Create a tab and get the rendered text content.
	text, err := createTab(client, baseURL, instanceID, pageURL)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to create tab for %q: %w", pageURL, err)
	}

	return text, nil
}

// StartIfEnabled starts the pinchtab subprocess if PINCHTAB_ENABLED is true.
// This function is safe to call multiple times; only the first call takes effect.
func StartIfEnabled() {
	if !config.Opts.PinchTabEnabled() {
		return
	}

	once.Do(func() {
		instance = &subprocess{}
		if err := instance.start(); err != nil {
			slog.Error("pinchtab: failed to start subprocess", slog.Any("error", err))
			instance = nil
		}
	})
}

// Stop gracefully stops the pinchtab subprocess.
func Stop() {
	if instance == nil {
		return
	}
	instance.stop()
}

func (s *subprocess) start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	binaryPath := config.Opts.PinchTabBinaryPath()
	s.cmd = exec.Command(binaryPath)
	s.cmd.Stdout = os.Stdout
	s.cmd.Stderr = os.Stderr

	slog.Info("pinchtab: starting subprocess", slog.String("binary", binaryPath))

	if err := s.cmd.Start(); err != nil {
		return fmt.Errorf("pinchtab: failed to start %q: %w", binaryPath, err)
	}

	s.running = true

	// Monitor the subprocess in a goroutine to detect unexpected exits.
	go func() {
		if err := s.cmd.Wait(); err != nil {
			slog.Error("pinchtab: subprocess exited with error", slog.Any("error", err))
		} else {
			slog.Info("pinchtab: subprocess exited normally")
		}
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	// Wait for pinchtab to be ready by polling the health endpoint.
	if err := s.waitForReady(); err != nil {
		s.stop()
		return err
	}

	slog.Info("pinchtab: subprocess is ready")
	return nil
}

func (s *subprocess) stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil || s.cmd.Process == nil {
		return
	}

	slog.Info("pinchtab: stopping subprocess")

	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		slog.Warn("pinchtab: failed to send interrupt, killing process", slog.Any("error", err))
		_ = s.cmd.Process.Kill()
	}

	s.running = false
}

func (s *subprocess) isRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// waitForReady polls the pinchtab health endpoint until it responds or
// the maximum wait time is exceeded.
func (s *subprocess) waitForReady() error {
	healthURL := config.Opts.PinchTabURL() + "/api/health"
	client := &http.Client{Timeout: 2 * time.Second}
	deadline := time.Now().Add(healthCheckMaxWait)

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(healthCheckInterval)
	}

	return fmt.Errorf("pinchtab: health check timed out after %s", healthCheckMaxWait)
}

func createInstance(client *http.Client, baseURL string) (string, error) {
	resp, err := client.Post(baseURL+"/api/instances", "application/json", nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result createInstanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
}

func createTab(client *http.Client, baseURL, instanceID, pageURL string) (string, error) {
	payload, err := json.Marshal(map[string]string{"url": pageURL})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/api/instances/%s/tabs", baseURL, instanceID)
	resp, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	var result createTabResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Text, nil
}

func deleteInstance(client *http.Client, baseURL, instanceID string) error {
	url := fmt.Sprintf("%s/api/instances/%s", baseURL, instanceID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}

	return nil
}
