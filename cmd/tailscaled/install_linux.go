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

func installSystemDaemonAndroid(args []string) error {
	if len(args) > 0 {
		return errors.New("install-system-daemon takes no arguments")
	}

	if err := os.MkdirAll(androidServiceDir, 0755); err != nil {
		return fmt.Errorf("failed to create %s: %w", androidServiceDir, err)
	}

	_ = os.Remove(androidLegacyLink)

	if err := os.WriteFile(androidServiceScript, []byte(androidBootScriptContent), 0755); err != nil {
		return fmt.Errorf("failed to write %s: %w", androidServiceScript, err)
	}
	_ = os.Chmod(androidServiceScript, 0755)

	fmt.Printf("Tailscale Android boot service successfully installed at %s\n", androidServiceScript)
	return nil
}

func uninstallSystemDaemonAndroid(args []string) error {
	if len(args) > 0 {
		return errors.New("uninstall-system-daemon takes no arguments")
	}

	_ = os.Remove(androidLegacyLink)
	if err := os.Remove(androidServiceScript); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove %s: %w", androidServiceScript, err)
	}

	fmt.Printf("Tailscale Android boot service successfully uninstalled from %s\n", androidServiceScript)
	return nil
}
