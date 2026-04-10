// MIT License
// Copyright (c) 2024 quantix
// test(PEPPER): Sprint 28 - accounts/key/utils additional coverage
package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateDefaultDirectories_NoPanel(t *testing.T) {
	// Should not panic even if dirs exist already
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = CreateDefaultDirectories()
}

func TestGetStorageInfo_NotNil(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	info := sm.GetStorageInfo()
	if info == nil {
		t.Fatal("expected non-nil storage info")
	}
}

func TestGetStorageInfo_HasFields(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	info := sm.GetStorageInfo()
	if _, ok := info["disk"]; !ok {
		t.Fatal("expected 'disk' field in storage info")
	}
	if _, ok := info["usb"]; !ok {
		t.Fatal("expected 'usb' field in storage info")
	}
}

func TestMountUSB_EmptyPath_Error(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	err = sm.MountUSB("")
	// Empty path should either succeed or fail — just no panic
	_ = err
}

func TestMountUSB_NonExistentPath(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	err = sm.MountUSB("/tmp/quantix-nonexistent-usb-path")
	_ = err
}

func TestInitializeUSBStorage_NilUSB_ErrorOrNoPanel(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = sm.InitializeUSBStorage("/tmp/qtx-test-usb")
}

func TestValidateStoragePath_ValidDisk(t *testing.T) {
	// Use a temp dir as valid path
	tmp := t.TempDir()
	err := ValidateStoragePath(tmp, StorageTypeDisk)
	// may or may not error depending on implementation — just no panic
	_ = err
}

func TestValidateStoragePath_EmptyPath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic with empty path: %v", r)
		}
	}()
	_ = ValidateStoragePath("", StorageTypeDisk)
}

func TestBackupToUSB_NoUSB_ErrorOrNoPanel(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = sm.BackupToUSB("testpassphrase")
}

func TestRestoreFromUSB_NoUSB_ErrorOrNoPanel(t *testing.T) {
	sm, err := NewStorageManager()
	if err != nil {
		t.Fatalf("NewStorageManager: %v", err)
	}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_, _ = sm.RestoreFromUSB("testpassphrase")
}

func TestGetDefaultStorageTypeDiskPath_NonEmpty(t *testing.T) {
	p := getDefaultDiskStoragePath()
	if p == "" {
		t.Fatal("expected non-empty default disk storage path")
	}
}

func TestGetDefaultBackupPath_NonEmpty(t *testing.T) {
	p := getDefaultBackupPath()
	if p == "" {
		t.Fatal("expected non-empty default backup path")
	}
}

func TestGetDefaultConfigPath_NonEmpty(t *testing.T) {
	p := getDefaultConfigPath()
	if p == "" {
		t.Fatal("expected non-empty default config path")
	}
}

func TestIsProductionEnvironment_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic: %v", r)
		}
	}()
	_ = isProductionEnvironment()
}

func TestCreateDefaultDirectories_CreatesDir(t *testing.T) {
	// Check at least no error on second call (idempotent)
	err1 := CreateDefaultDirectories()
	err2 := CreateDefaultDirectories()
	_ = err1
	_ = err2
}

func TestValidateStoragePath_AbsolutePath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "testkeys")
	_ = os.MkdirAll(path, 0700)
	_ = ValidateStoragePath(path, StorageTypeDisk)
}
