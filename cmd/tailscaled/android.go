// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

//go:build linux

package main

import (
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
	"tailscale.com/version/distro"
)

func initAndroidRootEnv() {
	if distro.Get() != distro.Android {
		return
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

	setupAndroidLogging()
}

func setupAndroidLogging() {
	logDir := "/data/adb/tailscale"
	if fi, err := os.Stat(logDir); err != nil || !fi.IsDir() {
		return
	}
	logPath := filepath.Join(logDir, "tailscaled.log")

	// Rotate log if larger than 8MB
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > 8*1024*1024 {
		_ = os.Rename(logPath, logPath+".1")
	}

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}

	log.SetOutput(f)
	os.Stdout = f
	os.Stderr = f

	fd := int(f.Fd())
	_ = unix.Dup2(fd, 1)
	_ = unix.Dup2(fd, 2)
}
