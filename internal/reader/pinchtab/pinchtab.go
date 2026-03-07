// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package pinchtab // import "miniflux.app/v2/internal/reader/pinchtab"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"miniflux.app/v2/internal/config"
)

const (
	httpRequestTimeout  = 30 * time.Second
	healthCheckInterval = 500 * time.Millisecond
	healthCheckMaxWait  = 30 * time.Second
)

type createInstanceResponse struct {
	ID string `json:"id"`
}

type createTabResponse struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

// RenderPage starts an ephemeral pinchtab subprocess, renders the given URL
// through a headless Chrome instance, extracts the text content, and tears
// down both the Chrome instance and the subprocess.
//
// proxyURL is optional: when non-empty it is forwarded to Chrome via
// --proxy-server so that the page fetch goes through the specified proxy.
// This supports per-feed proxy_url, the global HTTP_CLIENT_PROXY, and
// the proxy rotator — whichever the caller resolves.
//
// feedID isolates browser state (cookies, localStorage) per feed so that
// login sessions are preserved across fetches but different feeds cannot
// leak state to each other.
func RenderPage(pageURL, proxyURL string, feedID int64) (string, error) {
	port, err := findFreePort()
	if err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}

	cmd, err := startSubprocess(port, proxyURL, feedID)
	if err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}
	defer killSubprocess(cmd)

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := waitForReady(baseURL); err != nil {
		return "", fmt.Errorf("pinchtab: %w", err)
	}

	client := &http.Client{Timeout: httpRequestTimeout}

	instanceID, err := createInstance(client, baseURL)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to create instance: %w", err)
	}
	defer func() {
		if deleteErr := deleteInstance(client, baseURL, instanceID); deleteErr != nil {
			slog.Warn("pinchtab: failed to delete instance",
				slog.String("instance_id", instanceID),
				slog.Any("error", deleteErr),
			)
		}
	}()

	text, err := createTab(client, baseURL, instanceID, pageURL)
	if err != nil {
		return "", fmt.Errorf("pinchtab: failed to create tab for %q: %w", pageURL, err)
	}

	return text, nil
}

func startSubprocess(port int, proxyURL string, feedID int64) (*exec.Cmd, error) {
	binaryPath := config.Opts.PinchTabBinaryPath()
	cmd := exec.Command(binaryPath)

	profileName := fmt.Sprintf("miniflux-feed-%d", feedID)
	env := append(os.Environ(),
		"BRIDGE_STEALTH=full",
		"BRIDGE_PORT="+strconv.Itoa(port),
		"BRIDGE_NO_DASHBOARD=true",
		"BRIDGE_PROFILE="+profileName,
	)
	if proxyURL != "" {
		env = append(env, "CHROME_FLAGS=--proxy-server="+proxyURL)
	}
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	slog.Debug("pinchtab: starting ephemeral subprocess",
		slog.String("binary", binaryPath),
		slog.Int("port", port),
		slog.String("proxy_url", proxyURL),
	)

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start %q: %w", binaryPath, err)
	}

	return cmd, nil
}

func killSubprocess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	if err := cmd.Process.Kill(); err != nil && !isProcessDone(err) {
		slog.Warn("pinchtab: failed to kill subprocess", slog.Any("error", err))
	}
	// Reap zombie process.
	_ = cmd.Wait()
}

func isProcessDone(err error) bool {
	return err != nil && err.Error() == os.ErrProcessDone.Error()
}

func waitForReady(baseURL string) error {
	healthURL := baseURL + "/api/health"
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

	return fmt.Errorf("health check timed out after %s", healthCheckMaxWait)
}

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("unable to find free port: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()
	return port, nil
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
