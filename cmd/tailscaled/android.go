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

	baseDir := "/data/adb/tailscale"
	_ = os.MkdirAll(baseDir, 0700)
	_ = os.MkdirAll(filepath.Join(baseDir, "state"), 0700)
	_ = os.MkdirAll(filepath.Join(baseDir, ".cache"), 0700)
	_ = os.MkdirAll(filepath.Join(baseDir, "config"), 0700)
	_ = os.MkdirAll(filepath.Join(baseDir, "tmp"), 0700)

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

	ensureAndroidBootService()
	if !shouldRunCLI() {
		setupAndroidLogging()
	}
}

const (
	androidServiceDir    = "/data/adb/service.d"
	androidServiceScript = "/data/adb/service.d/tailscaled.sh"
	androidLegacyLink    = "/data/adb/service.d/tailscaled"
	androidTargetBin     = "/data/adb/tailscale/tailscaled"
)

const androidBootScriptContent = `#!/system/bin/sh

if [ -x /system/bin/wait_for_prop ]; then
	/system/bin/wait_for_prop sys.boot_completed 1
else
	until [ "$(getprop sys.boot_completed)" = "1" ]; do
		sleep 1
	done
fi

WAIT_COUNT=0
until ip route get 8.8.8.8 >/dev/null 2>&1 || ip -6 route get 2001:4860:4860::8888 >/dev/null 2>&1 || [ "$WAIT_COUNT" -ge 30 ]; do
	WAIT_COUNT=$((WAIT_COUNT + 1))
	sleep 1
done

exec /data/adb/tailscale/tailscaled
`

func ensureAndroidBootService() {
	if fi, err := os.Stat("/data/adb"); err == nil && fi.IsDir() {
		_ = os.MkdirAll(androidServiceDir, 0755)
		_ = os.Remove(androidLegacyLink)
		if _, err := os.Stat(androidServiceScript); os.IsNotExist(err) {
			if _, err := os.Stat(androidTargetBin); err == nil {
				_ = os.WriteFile(androidServiceScript, []byte(androidBootScriptContent), 0755)
				_ = os.Chmod(androidServiceScript, 0755)
			}
		}
	}
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
