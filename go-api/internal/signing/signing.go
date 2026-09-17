package signing

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Sign uses only c2patool's public development credentials. It does not assert
// camera capture, authorship, or the identity of the uploader.
func Sign(ctx context.Context, tool, source, output string) error {
	toolPath, err := exec.LookPath(tool)
	if err != nil {
		return fmt.Errorf("find c2patool: %w", err)
	}
	toolPath, err = filepath.Abs(toolPath)
	if err != nil {
		return err
	}
	dir, err := os.MkdirTemp("", "c2dp-sign-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	manifest := `{"claim_generator_info":[{"name":"C2DP upload demo","version":"1.0.0"}],"title":"C2DP uploaded image","assertions":[{"label":"org.c2dp.upload","data":{"operation":"Received and signed an uploaded image","credentials":"Public C2PA development test certificate","note":"Does not attest to the uploader's identity, original capture, or truth of the scene."}}]}`
	manifestPath := filepath.Join(dir, "manifest.json")
	settingsPath := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(manifestPath, []byte(manifest), 0600); err != nil {
		return err
	}
	if err := os.WriteFile(settingsPath, []byte(`{}`), 0600); err != nil {
		return err
	}
	// Explicit settings and filtered environment prevent accidentally using a
	// developer's real signing credentials from their local c2patool setup.
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, toolPath, source, "--settings", settingsPath, "--manifest", manifestPath, "--output", output)
	cmd.Stdout = nil
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "C2PA") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("signing interrupted: %w", ctx.Err())
		}
		return fmt.Errorf("c2patool signing failed: %w", err)
	}
	info, err := os.Stat(output)
	if err != nil {
		return fmt.Errorf("missing signed output: %w", err)
	}
	if info.Size() == 0 {
		return fmt.Errorf("empty signed output")
	}
	return nil
}

// Inspect runs c2patool against an already-signed asset and returns its
// manifest report (the same JSON `c2patool <path>` prints to stdout).
func Inspect(ctx context.Context, tool, path string) (json.RawMessage, error) {
	toolPath, err := exec.LookPath(tool)
	if err != nil {
		return nil, fmt.Errorf("find c2patool: %w", err)
	}
	toolPath, err = filepath.Abs(toolPath)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, toolPath, path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	for _, env := range os.Environ() {
		if !strings.HasPrefix(env, "C2PA") {
			cmd.Env = append(cmd.Env, env)
		}
	}
	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("inspecting interrupted: %w", ctx.Err())
		}
		return nil, fmt.Errorf("c2patool inspect failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	raw := bytes.TrimSpace(stdout.Bytes())
	if !json.Valid(raw) {
		return nil, fmt.Errorf("c2patool returned invalid JSON report")
	}
	return json.RawMessage(raw), nil
}
