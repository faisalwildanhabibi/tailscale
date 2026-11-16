// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package main

import (
	"os"
	"path/filepath"

	"tailscale.com/version/distro"
)

func init() {
	if distro.Get() != distro.Android {
		return
	}

	baseDir := "/data/adb/tailscale"
	if os.Getenv("HOME") == "" || os.Getenv("HOME") == "/" {
		os.Setenv("HOME", baseDir)
	}
	if os.Getenv("XDG_CACHE_HOME") == "" {
		os.Setenv("XDG_CACHE_HOME", filepath.Join(baseDir, ".cache"))
	}
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		os.Setenv("XDG_CONFIG_HOME", filepath.Join(baseDir, "config"))
	}
	if os.Getenv("XDG_DATA_HOME") == "" {
		os.Setenv("XDG_DATA_HOME", filepath.Join(baseDir, "state"))
	}
	if os.Getenv("TMPDIR") == "" {
		os.Setenv("TMPDIR", filepath.Join(baseDir, "tmp"))
	}

	if os.Getenv("SSL_CERT_DIR") == "" {
		if _, err := os.Stat("/system/etc/security/cacerts"); err == nil {
			os.Setenv("SSL_CERT_DIR", "/system/etc/security/cacerts")
		}
	}
	if os.Getenv("SSL_CERT_FILE") == "" {
		if _, err := os.Stat("/data/adb/tailscale/cacert.pem"); err == nil {
			os.Setenv("SSL_CERT_FILE", "/data/adb/tailscale/cacert.pem")
		}
	}
}
