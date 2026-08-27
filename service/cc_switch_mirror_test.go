package service

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestGetCCSwitchMirrorDir(t *testing.T) {
	t.Setenv("CC_SWITCH_MIRROR_DIR", "")
	if got := getCCSwitchMirrorDir(); got != "downloads/cc-switch" {
		t.Fatalf("expected portable default mirror dir, got %q", got)
	}

	override := t.TempDir()
	t.Setenv("CC_SWITCH_MIRROR_DIR", override)
	if got := getCCSwitchMirrorDir(); got != override {
		t.Fatalf("expected mirror dir override %q, got %q", override, got)
	}
}

func TestSelectCCSwitchReleaseAssets(t *testing.T) {
	assets := []ccSwitchGitHubAsset{
		{Name: "cc-switch.AppImage"},
		{Name: "cc-switch.dmg"},
		{Name: "cc-switch.msi"},
		{Name: "source.zip"},
	}

	selected := selectCCSwitchReleaseAssets(assets, ccSwitchMirrorPlatformSpec{Extension: ".AppImage"})
	if len(selected) != 1 {
		t.Fatalf("expected 1 AppImage asset, got %d", len(selected))
	}
	if selected[0].Name != "cc-switch.AppImage" {
		t.Fatalf("unexpected selected asset: %s", selected[0].Name)
	}
}

func TestSelectCCSwitchReleaseAssetsPrefersRecommendedAsset(t *testing.T) {
	assets := []ccSwitchGitHubAsset{
		{Name: "CC-Switch-v3.14.1-Linux-arm64.AppImage"},
		{Name: "CC-Switch-v3.14.1-Linux-x86_64.AppImage"},
		{Name: "CC-Switch-v3.14.1-Linux-x86_64.deb"},
	}
	spec := ccSwitchMirrorPlatformSpec{
		Extension:             ".AppImage",
		PreferredNameContains: []string{"linux-x86_64.appimage", "x86_64.appimage"},
	}

	selected := selectCCSwitchReleaseAssets(assets, spec)
	if len(selected) != 1 {
		t.Fatalf("expected 1 preferred AppImage asset, got %d", len(selected))
	}
	if selected[0].Name != "CC-Switch-v3.14.1-Linux-x86_64.AppImage" {
		t.Fatalf("unexpected selected asset: %s", selected[0].Name)
	}
}

func TestReuseCurrentCCSwitchAssets(t *testing.T) {
	mirrorDir := t.TempDir()
	t.Setenv("CC_SWITCH_MIRROR_DIR", mirrorDir)

	assetName := "CC-Switch-v3.14.1-Windows.msi"
	relativePath := filepath.ToSlash(filepath.Join("files", "windows", assetName))
	if err := os.MkdirAll(filepath.Dir(filepath.Join(mirrorDir, relativePath)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(mirrorDir, relativePath), []byte("installer"), 0o644); err != nil {
		t.Fatal(err)
	}

	oldAssets := []CCSwitchMirrorAsset{{
		ID:           ccSwitchAssetID(assetName),
		Name:         assetName,
		RelativePath: relativePath,
		Size:         9,
		Version:      "v3.14.1",
	}}
	candidates := []ccSwitchGitHubAsset{{
		Name: assetName,
		Size: 9,
	}}

	reused, ok := reuseCurrentCCSwitchAssets(oldAssets, "v3.14.1", candidates)
	if !ok {
		t.Fatal("expected current asset to be reused")
	}
	if len(reused) != 1 || reused[0].Name != assetName {
		t.Fatalf("unexpected reused assets: %+v", reused)
	}
}

func TestDurationUntilNextShanghaiMidnight(t *testing.T) {
	now := time.Date(2026, 5, 9, 15, 30, 0, 0, time.FixedZone("UTC", 0))
	got := durationUntilNextShanghaiMidnight(now)
	want := 30 * time.Minute
	if got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}
}

func TestValidateCCSwitchMirrorDownloadTicket(t *testing.T) {
	platform := "linux"
	assetID := "asset-1"
	userID := "123"
	expiresAt := strconv.FormatInt(time.Now().Add(time.Minute).Unix(), 10)
	signature := signCCSwitchMirrorDownloadTicket(platform, assetID, userID, expiresAt)

	if err := ValidateCCSwitchMirrorDownloadTicket(platform, assetID, userID, expiresAt, signature); err != nil {
		t.Fatalf("expected valid ticket, got %v", err)
	}
	if err := ValidateCCSwitchMirrorDownloadTicket(platform, assetID, userID, expiresAt, signature+"x"); err == nil {
		t.Fatal("expected tampered signature to fail")
	}
}

func TestValidateCCSwitchMirrorDownloadTicketRejectsExpiredTicket(t *testing.T) {
	platform := "linux"
	assetID := "asset-1"
	userID := "123"
	expiresAt := strconv.FormatInt(time.Now().Add(-time.Minute).Unix(), 10)
	signature := signCCSwitchMirrorDownloadTicket(platform, assetID, userID, expiresAt)

	if err := ValidateCCSwitchMirrorDownloadTicket(platform, assetID, userID, expiresAt, signature); err == nil {
		t.Fatal("expected expired ticket to fail")
	}
}
