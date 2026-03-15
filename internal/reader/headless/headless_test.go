// SPDX-FileCopyrightText: Copyright The Miniflux Authors. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package headless

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

func findTestPort() int {
	l, _ := net.Listen("tcp", "127.0.0.1:0")
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return port
}

func startLightpanda(t *testing.T, port int) *exec.Cmd {
	t.Helper()

	binaryPath := os.Getenv("LIGHTPANDA_BINARY_PATH")
	if binaryPath == "" {
		binaryPath = "/usr/local/bin/lightpanda"
	}
	if _, err := os.Stat(binaryPath); err != nil {
		t.Skipf("Lightpanda binary not found at %s, skipping e2e test", binaryPath)
	}

	cmd := exec.Command(binaryPath, "serve", "--host", "127.0.0.1", "--port", fmt.Sprintf("%d", port))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start Lightpanda: %v", err)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		wsURL, err := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil && wsURL != "" {
			return cmd
		}
		time.Sleep(300 * time.Millisecond)
	}

	cmd.Process.Kill()
	cmd.Wait()
	t.Fatal("Lightpanda CDP server did not become ready within 15s")
	return nil
}

func TestLightpandaCDPConnection(t *testing.T) {
	port := findTestPort()
	cmd := startLightpanda(t, port)
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	wsURL, err := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to resolve CDP URL: %v", err)
	}
	t.Logf("CDP WebSocket URL: %s", wsURL)

	browser := rod.New().ControlURL(wsURL)
	err = browser.Connect()
	if err != nil {
		t.Fatalf("Failed to connect to Lightpanda CDP: %v", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	defer page.Close()

	err = page.WaitLoad()
	if err != nil {
		t.Fatalf("WaitLoad failed: %v", err)
	}

	result, err := page.Eval(`() => document.title`)
	if err != nil {
		t.Fatalf("Eval document.title failed: %v", err)
	}

	title := result.Value.Str()
	t.Logf("Page title: %q", title)
	if title == "" {
		t.Error("Expected non-empty page title from example.com")
	}
}

func TestReadableContentExtraction(t *testing.T) {
	port := findTestPort()
	cmd := startLightpanda(t, port)
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	wsURL, err := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to resolve CDP URL: %v", err)
	}

	browser := rod.New().ControlURL(wsURL)
	err = browser.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	defer page.Close()

	err = page.WaitLoad()
	if err != nil {
		t.Fatalf("WaitLoad failed: %v", err)
	}

	content, err := extractReadableContent(page, "https://example.com")
	if err != nil {
		t.Fatalf("extractReadableContent failed: %v", err)
	}

	t.Logf("Extracted content (%d chars): %.200s", len(content), content)

	if content == "" {
		t.Error("Expected non-empty content from example.com")
	}

	if !strings.Contains(content, "documentation") {
		t.Error("Expected content to contain 'documentation'")
	}
}

func TestFullHTMLExtraction(t *testing.T) {
	port := findTestPort()
	cmd := startLightpanda(t, port)
	defer func() {
		cmd.Process.Kill()
		cmd.Wait()
	}()

	wsURL, err := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		t.Fatalf("Failed to resolve CDP URL: %v", err)
	}

	browser := rod.New().ControlURL(wsURL)
	err = browser.Connect()
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer browser.Close()

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://example.com"})
	if err != nil {
		t.Fatalf("Failed to create page: %v", err)
	}
	defer page.Close()

	err = page.WaitLoad()
	if err != nil {
		t.Fatalf("WaitLoad failed: %v", err)
	}

	html, err := extractFullHTML(page)
	if err != nil {
		t.Fatalf("extractFullHTML failed: %v", err)
	}

	t.Logf("Full HTML length: %d", len(html))

	if !strings.Contains(html, "<html") {
		t.Error("Expected full HTML to contain <html tag")
	}
	if !strings.Contains(html, "Example Domain") {
		t.Error("Expected full HTML to contain 'Example Domain'")
	}
}

func killAndWait(cmd *exec.Cmd) {
	if cmd != nil && cmd.Process != nil {
		cmd.Process.Kill()
		cmd.Wait()
	}
}

func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

func TestProcessCleanupAfterSuccess(t *testing.T) {
	port := findTestPort()
	cmd := startLightpanda(t, port)
	pid := cmd.Process.Pid
	activeProcessCount.Add(1)

	stopSubprocess(cmd)

	time.Sleep(500 * time.Millisecond)
	if isProcessAlive(pid) {
		t.Errorf("Lightpanda process %d still alive after stopSubprocess", pid)
	}
}

func TestProcessCleanupAfterCDPDisconnect(t *testing.T) {
	port := findTestPort()
	cmd := startLightpanda(t, port)
	pid := cmd.Process.Pid
	activeProcessCount.Add(1)

	wsURL, err := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		killAndWait(cmd)
		t.Fatalf("Failed to resolve CDP URL: %v", err)
	}

	browser := rod.New().ControlURL(wsURL)
	if err := browser.Connect(); err != nil {
		killAndWait(cmd)
		t.Fatalf("Connect failed: %v", err)
	}

	page, err := browser.Page(proto.TargetCreateTarget{URL: "https://example.com"})
	if err != nil {
		browser.Close()
		killAndWait(cmd)
		t.Fatalf("Page creation failed: %v", err)
	}
	page.WaitLoad()

	_ = page.Close()
	browser.Close()
	stopSubprocess(cmd)

	time.Sleep(500 * time.Millisecond)
	if isProcessAlive(pid) {
		t.Errorf("Lightpanda process %d still alive after cleanup", pid)
	}
}

func TestActiveProcessCountAccuracy(t *testing.T) {
	before := ActiveProcessCount()

	port := findTestPort()
	cmd := startLightpanda(t, port)
	activeProcessCount.Add(1)

	if got := ActiveProcessCount(); got != before+1 {
		t.Errorf("ActiveProcessCount after start: got %d, want %d", got, before+1)
	}

	stopSubprocess(cmd)

	time.Sleep(500 * time.Millisecond)
	if got := ActiveProcessCount(); got != before {
		t.Errorf("ActiveProcessCount after stop: got %d, want %d", got, before)
	}
}

func TestStopSubprocessNilSafe(t *testing.T) {
	stopSubprocess(nil)

	cmd := &exec.Cmd{}
	stopSubprocess(cmd)
}

func TestMultipleSequentialRenders(t *testing.T) {
	before := ActiveProcessCount()

	for i := 0; i < 3; i++ {
		port := findTestPort()
		cmd := startLightpanda(t, port)
		activeProcessCount.Add(1)

		wsURL, _ := launcher.ResolveURL(fmt.Sprintf("127.0.0.1:%d", port))
		browser := rod.New().ControlURL(wsURL)
		browser.Connect()

		page, _ := browser.Page(proto.TargetCreateTarget{URL: "https://example.com"})
		page.WaitLoad()

		result, _ := page.Eval(`() => document.title`)
		t.Logf("Iteration %d: title=%s", i, result.Value.Str())

		_ = page.Close()
		browser.Close()
		stopSubprocess(cmd)
	}

	time.Sleep(500 * time.Millisecond)
	after := ActiveProcessCount()
	if after != before {
		t.Errorf("Process leak: ActiveProcessCount before=%d after=%d", before, after)
	}

	lightpandaCount := LightpandaProcessCount()
	t.Logf("LightpandaProcessCount via /proc: %d", lightpandaCount)
}

func TestNoZombieProcesses(t *testing.T) {
	var pids []int
	for i := 0; i < 3; i++ {
		port := findTestPort()
		cmd := startLightpanda(t, port)
		pids = append(pids, cmd.Process.Pid)
		activeProcessCount.Add(1)
		stopSubprocess(cmd)
	}

	time.Sleep(1 * time.Second)
	for _, pid := range pids {
		if isProcessAlive(pid) {
			t.Errorf("Zombie process detected: pid %d", pid)
		}
	}
}
