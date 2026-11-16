// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package main

import (
	"errors"
	"fmt"
	"os"

	"tailscale.com/version/distro"
)

func init() {
	if distro.Get() == distro.Android {
		installSystemDaemon = installSystemDaemonAndroid
		uninstallSystemDaemon = uninstallSystemDaemonAndroid
	}
}

const (
	androidServiceDir     = "/data/adb/service.d"
	androidServiceSymlink = "/data/adb/service.d/tailscaled"
	androidTargetBin      = "/data/adb/tailscale/tailscaled"
)

func installSystemDaemonAndroid(args []string) error {
	if len(args) > 0 {
		return errors.New("install-system-daemon takes no arguments")
	}

	exe, err := os.Executable()
	if err != nil {
		exe = androidTargetBin
	}

	if err := os.MkdirAll(androidServiceDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", androidServiceDir, err)
	}

	_ = os.Remove(androidServiceSymlink)

	if err := os.Symlink(exe, androidServiceSymlink); err != nil {
		return fmt.Errorf("failed to create boot service link %s -> %s: %w", androidServiceSymlink, exe, err)
	}

	_ = os.Chmod(androidServiceSymlink, 0755)
	fmt.Printf("Tailscale Android boot service successfully installed at %s\n", androidServiceSymlink)
	return nil
}

func uninstallSystemDaemonAndroid(args []string) error {
	if len(args) > 0 {
		return errors.New("uninstall-system-daemon takes no arguments")
	}

	if err := os.Remove(androidServiceSymlink); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", androidServiceSymlink, err)
	}

	fmt.Printf("Tailscale Android boot service successfully uninstalled from %s\n", androidServiceSymlink)
	return nil
}
