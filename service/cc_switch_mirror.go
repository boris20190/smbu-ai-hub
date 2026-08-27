package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/bytedance/gopkg/util/gopool"
)

const (
	ccSwitchMirrorDefaultDir     = "downloads/cc-switch"
	ccSwitchMirrorDefaultAPIURL  = "https://api.github.com/repos/farion1231/cc-switch/releases/latest"
	ccSwitchMirrorRepositoryURL  = "https://github.com/farion1231/cc-switch"
	ccSwitchMirrorManifestName   = "manifest.json"
	ccSwitchMirrorUserAgent      = "new-api-cc-switch-mirror"
	ccSwitchMirrorDefaultTimeout = 20 * time.Minute
	ccSwitchMirrorDownloadTries  = 3
	ccSwitchMirrorRetryDelay     = 3 * time.Second
	ccSwitchMirrorTicketTTL      = 5 * time.Minute
)

type ccSwitchMirrorPlatformSpec struct {
	Key                   string
	Label                 string
	PackageType           string
	Extension             string
	PreferredNameContains []string
}

var ccSwitchMirrorPlatformSpecs = []ccSwitchMirrorPlatformSpec{
	{Key: "windows", Label: "Windows", PackageType: ".msi", Extension: ".msi", PreferredNameContains: []string{"windows.msi"}},
	{Key: "linux", Label: "Linux", PackageType: ".AppImage", Extension: ".appimage", PreferredNameContains: []string{"linux-x86_64.appimage", "x86_64.appimage", "amd64.appimage"}},
	{Key: "mac", Label: "Mac", PackageType: ".dmg", Extension: ".dmg", PreferredNameContains: []string{"macos.dmg", "mac.dmg"}},
}

type ccSwitchGitHubRelease struct {
	TagName     string                `json:"tag_name"`
	Name        string                `json:"name"`
	HTMLURL     string                `json:"html_url"`
	PublishedAt string                `json:"published_at"`
	Assets      []ccSwitchGitHubAsset `json:"assets"`
}

type ccSwitchGitHubAsset struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
	UpdatedAt          string `json:"updated_at"`
}

type CCSwitchMirrorAsset struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	FileName       string `json:"file_name"`
	StoredFileName string `json:"stored_file_name"`
	RelativePath   string `json:"relative_path"`
	SourceURL      string `json:"source_url,omitempty"`
	Size           int64  `json:"size"`
	Version        string `json:"version"`
	MirroredAt     int64  `json:"mirrored_at"`
}

type CCSwitchMirrorPlatform struct {
	Key         string                `json:"key"`
	Label       string                `json:"label"`
	PackageType string                `json:"package_type"`
	Assets      []CCSwitchMirrorAsset `json:"assets"`
}

type CCSwitchMirrorManifest struct {
	RepositoryURL string                            `json:"repository_url"`
	ReleaseURL    string                            `json:"release_url,omitempty"`
	Platforms     map[string]CCSwitchMirrorPlatform `json:"platforms"`
}

type CCSwitchMirrorAssetView struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	FileName    string `json:"file_name"`
	Size        int64  `json:"size"`
	Version     string `json:"version"`
	MirroredAt  int64  `json:"mirrored_at"`
	DownloadURL string `json:"download_url"`
}

type CCSwitchMirrorPlatformView struct {
	Key         string                    `json:"key"`
	Label       string                    `json:"label"`
	PackageType string                    `json:"package_type"`
	Available   bool                      `json:"available"`
	Assets      []CCSwitchMirrorAssetView `json:"assets"`
}

type CCSwitchMirrorManifestView struct {
	RepositoryURL string                       `json:"repository_url"`
	Platforms     []CCSwitchMirrorPlatformView `json:"platforms"`
}

type CCSwitchMirrorDownload struct {
	Path     string
	FileName string
}

type CCSwitchMirrorDownloadTicket struct {
	DownloadURL string `json:"download_url"`
	ExpiresAt   int64  `json:"expires_at"`
}

type stagedCCSwitchMirrorAsset struct {
	asset      CCSwitchMirrorAsset
	stagedPath string
	finalPath  string
}

var (
	ccSwitchMirrorOnce    sync.Once
	ccSwitchMirrorRunning atomic.Bool
)

func StartCCSwitchMirrorTask() {
	ccSwitchMirrorOnce.Do(func() {
		if !common.IsMasterNode || !common.GetEnvOrDefaultBool("CC_SWITCH_MIRROR_ENABLED", true) {
			return
		}

		gopool.Go(func() {
			startupDelay := time.Duration(common.GetEnvOrDefault("CC_SWITCH_MIRROR_STARTUP_DELAY_SECONDS", 15)) * time.Second
			logger.LogInfo(context.Background(), fmt.Sprintf("cc-switch mirror task started: startup_delay=%s", startupDelay))
			if startupDelay > 0 {
				time.Sleep(startupDelay)
			}
			runCCSwitchMirrorOnce()

			for {
				wait := durationUntilNextShanghaiMidnight(time.Now())
				timer := time.NewTimer(wait)
				<-timer.C
				runCCSwitchMirrorOnce()
			}
		})
	})
}

func runCCSwitchMirrorOnce() {
	if !ccSwitchMirrorRunning.CompareAndSwap(false, true) {
		return
	}
	defer ccSwitchMirrorRunning.Store(false)

	timeout := time.Duration(common.GetEnvOrDefault("CC_SWITCH_MIRROR_TIMEOUT_SECONDS", int(ccSwitchMirrorDefaultTimeout.Seconds()))) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := SyncCCSwitchMirror(ctx); err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("cc-switch mirror sync failed: %v", err))
	}
}

func SyncCCSwitchMirror(ctx context.Context) error {
	mirrorDir := getCCSwitchMirrorDir()
	if err := os.MkdirAll(mirrorDir, 0o755); err != nil {
		return fmt.Errorf("create mirror dir failed: %w", err)
	}

	release, err := fetchLatestCCSwitchRelease(ctx)
	if err != nil {
		return err
	}
	if len(release.Assets) == 0 {
		return errors.New("release has no assets")
	}

	oldManifest := loadCCSwitchMirrorManifest()
	newManifest := cloneCCSwitchMirrorManifest(oldManifest)
	newManifest.RepositoryURL = ccSwitchMirrorRepositoryURL
	newManifest.ReleaseURL = release.HTMLURL
	if newManifest.Platforms == nil {
		newManifest.Platforms = make(map[string]CCSwitchMirrorPlatform)
	}

	mirroredCount := 0
	downloadedCount := 0
	cleanupJobs := make([]struct {
		oldAssets []CCSwitchMirrorAsset
		newAssets []CCSwitchMirrorAsset
	}, 0, len(ccSwitchMirrorPlatformSpecs))
	for _, spec := range ccSwitchMirrorPlatformSpecs {
		candidates := selectCCSwitchReleaseAssets(release.Assets, spec)
		if len(candidates) == 0 {
			continue
		}

		oldPlatform := newManifest.Platforms[spec.Key]
		if reused, ok := reuseCurrentCCSwitchAssets(oldPlatform.Assets, release.TagName, candidates); ok {
			newManifest.Platforms[spec.Key] = CCSwitchMirrorPlatform{
				Key:         spec.Key,
				Label:       spec.Label,
				PackageType: spec.PackageType,
				Assets:      reused,
			}
			mirroredCount += len(reused)
			continue
		}

		assets, err := mirrorCCSwitchPlatformAssets(ctx, mirrorDir, release, spec, candidates)
		if err != nil {
			logger.LogWarn(ctx, fmt.Sprintf("cc-switch mirror sync skipped platform=%s: %v", spec.Key, err))
			continue
		}

		newManifest.Platforms[spec.Key] = CCSwitchMirrorPlatform{
			Key:         spec.Key,
			Label:       spec.Label,
			PackageType: spec.PackageType,
			Assets:      assets,
		}
		cleanupJobs = append(cleanupJobs, struct {
			oldAssets []CCSwitchMirrorAsset
			newAssets []CCSwitchMirrorAsset
		}{
			oldAssets: oldPlatform.Assets,
			newAssets: assets,
		})
		mirroredCount += len(assets)
		downloadedCount += len(assets)
	}

	if mirroredCount == 0 {
		return errors.New("no platform assets were mirrored")
	}
	if downloadedCount == 0 {
		logger.LogInfo(ctx, fmt.Sprintf("cc-switch mirror sync already up to date: assets=%d", mirroredCount))
		return nil
	}
	if err := saveCCSwitchMirrorManifest(mirrorDir, newManifest); err != nil {
		return err
	}
	for _, job := range cleanupJobs {
		cleanupReplacedCCSwitchFiles(mirrorDir, job.oldAssets, job.newAssets)
	}

	logger.LogInfo(ctx, fmt.Sprintf("cc-switch mirror sync completed: assets=%d", downloadedCount))
	return nil
}

func GetCCSwitchMirrorManifestView() CCSwitchMirrorManifestView {
	manifest := loadCCSwitchMirrorManifest()
	result := CCSwitchMirrorManifestView{
		RepositoryURL: ccSwitchMirrorRepositoryURL,
		Platforms:     make([]CCSwitchMirrorPlatformView, 0, len(ccSwitchMirrorPlatformSpecs)),
	}
	if manifest.RepositoryURL != "" {
		result.RepositoryURL = manifest.RepositoryURL
	}

	for _, spec := range ccSwitchMirrorPlatformSpecs {
		platform := manifest.Platforms[spec.Key]
		view := CCSwitchMirrorPlatformView{
			Key:         spec.Key,
			Label:       spec.Label,
			PackageType: spec.PackageType,
			Assets:      make([]CCSwitchMirrorAssetView, 0, len(platform.Assets)),
		}
		for _, asset := range platform.Assets {
			if !ccSwitchMirrorFileExists(asset.RelativePath) {
				continue
			}
			view.Assets = append(view.Assets, CCSwitchMirrorAssetView{
				ID:          asset.ID,
				Name:        asset.Name,
				FileName:    asset.FileName,
				Size:        asset.Size,
				Version:     asset.Version,
				MirroredAt:  asset.MirroredAt,
				DownloadURL: fmt.Sprintf("/api/downloads/cc-switch/%s/%s", spec.Key, asset.ID),
			})
		}
		view.Available = len(view.Assets) > 0
		result.Platforms = append(result.Platforms, view)
	}

	return result
}

func ResolveCCSwitchMirrorDownload(platformKey, assetID string) (CCSwitchMirrorDownload, error) {
	manifest := loadCCSwitchMirrorManifest()
	platform, ok := manifest.Platforms[platformKey]
	if !ok {
		return CCSwitchMirrorDownload{}, errors.New("download is not available")
	}
	for _, asset := range platform.Assets {
		if asset.ID != assetID {
			continue
		}
		path := ccSwitchMirrorAbsolutePath(asset.RelativePath)
		if path == "" {
			return CCSwitchMirrorDownload{}, errors.New("download is not available")
		}
		if _, err := os.Stat(path); err != nil {
			return CCSwitchMirrorDownload{}, errors.New("download is not available")
		}
		return CCSwitchMirrorDownload{Path: path, FileName: asset.FileName}, nil
	}
	return CCSwitchMirrorDownload{}, errors.New("download is not available")
}

func CreateCCSwitchMirrorDownloadTicket(platformKey, assetID string, userID int) (CCSwitchMirrorDownloadTicket, error) {
	if _, err := ResolveCCSwitchMirrorDownload(platformKey, assetID); err != nil {
		return CCSwitchMirrorDownloadTicket{}, err
	}

	expiresAt := time.Now().Add(ccSwitchMirrorTicketTTL).Unix()
	userIDStr := strconv.Itoa(userID)
	expiresAtStr := strconv.FormatInt(expiresAt, 10)
	signature := signCCSwitchMirrorDownloadTicket(platformKey, assetID, userIDStr, expiresAtStr)
	params := url.Values{}
	params.Set("user_id", userIDStr)
	params.Set("expires_at", expiresAtStr)
	params.Set("signature", signature)

	return CCSwitchMirrorDownloadTicket{
		DownloadURL: fmt.Sprintf("/api/downloads/cc-switch/%s/%s/file?%s", platformKey, assetID, params.Encode()),
		ExpiresAt:   expiresAt,
	}, nil
}

func ValidateCCSwitchMirrorDownloadTicket(platformKey, assetID, userID, expiresAt, signature string) error {
	if userID == "" || expiresAt == "" || signature == "" {
		return errors.New("download ticket is required")
	}
	expiresAtUnix, err := strconv.ParseInt(expiresAt, 10, 64)
	if err != nil {
		return errors.New("download ticket is invalid")
	}
	if time.Now().Unix() > expiresAtUnix {
		return errors.New("download ticket has expired")
	}
	expected := signCCSwitchMirrorDownloadTicket(platformKey, assetID, userID, expiresAt)
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("download ticket is invalid")
	}
	return nil
}

func signCCSwitchMirrorDownloadTicket(platformKey, assetID, userID, expiresAt string) string {
	payload := strings.Join([]string{platformKey, assetID, userID, expiresAt}, "\n")
	return common.GenerateHMAC(payload)
}

func fetchLatestCCSwitchRelease(ctx context.Context) (ccSwitchGitHubRelease, error) {
	apiURL := common.GetEnvOrDefaultString("CC_SWITCH_RELEASE_API_URL", ccSwitchMirrorDefaultAPIURL)
	if err := validateCCSwitchOutboundURL(apiURL); err != nil {
		return ccSwitchGitHubRelease{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return ccSwitchGitHubRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", ccSwitchMirrorUserAgent)

	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return ccSwitchGitHubRelease{}, fmt.Errorf("fetch release failed: %w", err)
	}
	defer CloseResponseBodyGracefully(resp)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ccSwitchGitHubRelease{}, fmt.Errorf("fetch release failed: status=%d", resp.StatusCode)
	}

	var release ccSwitchGitHubRelease
	if err := common.DecodeJson(resp.Body, &release); err != nil {
		return ccSwitchGitHubRelease{}, fmt.Errorf("parse release failed: %w", err)
	}
	return release, nil
}

func mirrorCCSwitchPlatformAssets(ctx context.Context, mirrorDir string, release ccSwitchGitHubRelease, spec ccSwitchMirrorPlatformSpec, assets []ccSwitchGitHubAsset) ([]CCSwitchMirrorAsset, error) {
	stageDir := filepath.Join(mirrorDir, ".tmp", fmt.Sprintf("%s-%d", spec.Key, time.Now().UnixNano()))
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return nil, fmt.Errorf("create staging dir failed: %w", err)
	}
	defer os.RemoveAll(stageDir)

	stagedAssets := make([]stagedCCSwitchMirrorAsset, 0, len(assets))
	mirrored := make([]CCSwitchMirrorAsset, 0, len(assets))
	for _, asset := range assets {
		stagedAsset, err := stageCCSwitchAsset(ctx, mirrorDir, stageDir, release, spec, asset)
		if err != nil {
			return nil, err
		}
		stagedAssets = append(stagedAssets, stagedAsset)
		mirrored = append(mirrored, stagedAsset.asset)
	}

	for _, stagedAsset := range stagedAssets {
		if err := os.MkdirAll(filepath.Dir(stagedAsset.finalPath), 0o755); err != nil {
			return nil, fmt.Errorf("create platform dir failed: %w", err)
		}
		if err := os.Rename(stagedAsset.stagedPath, stagedAsset.finalPath); err != nil {
			return nil, fmt.Errorf("replace asset %s failed: %w", stagedAsset.asset.Name, err)
		}
	}
	return mirrored, nil
}

func stageCCSwitchAsset(ctx context.Context, mirrorDir string, stageDir string, release ccSwitchGitHubRelease, spec ccSwitchMirrorPlatformSpec, asset ccSwitchGitHubAsset) (stagedCCSwitchMirrorAsset, error) {
	var lastErr error
	for attempt := 1; attempt <= ccSwitchMirrorDownloadTries; attempt++ {
		stagedAsset, err := stageCCSwitchAssetOnce(ctx, mirrorDir, stageDir, release, spec, asset)
		if err == nil {
			return stagedAsset, nil
		}
		lastErr = err
		if attempt == ccSwitchMirrorDownloadTries {
			break
		}
		logger.LogWarn(ctx, fmt.Sprintf("cc-switch mirror retry asset=%s attempt=%d/%d: %v", asset.Name, attempt+1, ccSwitchMirrorDownloadTries, err))
		if err := sleepCCSwitchMirrorRetry(ctx, time.Duration(attempt)*ccSwitchMirrorRetryDelay); err != nil {
			return stagedCCSwitchMirrorAsset{}, err
		}
	}
	return stagedCCSwitchMirrorAsset{}, lastErr
}

func stageCCSwitchAssetOnce(ctx context.Context, mirrorDir string, stageDir string, release ccSwitchGitHubRelease, spec ccSwitchMirrorPlatformSpec, asset ccSwitchGitHubAsset) (stagedCCSwitchMirrorAsset, error) {
	if asset.BrowserDownloadURL == "" {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("asset %s has no download url", asset.Name)
	}
	if err := validateCCSwitchOutboundURL(asset.BrowserDownloadURL); err != nil {
		return stagedCCSwitchMirrorAsset{}, err
	}

	fileName := sanitizeCCSwitchFileName(asset.Name, spec.Extension)
	assetID := ccSwitchAssetID(asset.Name)
	storedFileName := fmt.Sprintf("%s-%s", assetID, fileName)
	relativePath := filepath.ToSlash(filepath.Join("files", spec.Key, storedFileName))
	finalPath := filepath.Join(mirrorDir, relativePath)
	stagedPath := filepath.Join(stageDir, storedFileName)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return stagedCCSwitchMirrorAsset{}, err
	}
	req.Header.Set("User-Agent", ccSwitchMirrorUserAgent)

	client := GetHttpClient()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("download asset %s failed: %w", asset.Name, err)
	}
	defer CloseResponseBodyGracefully(resp)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("download asset %s failed: status=%d", asset.Name, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp(stageDir, "cc-switch-*")
	if err != nil {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("create temp file failed: %w", err)
	}
	tmpName := tmpFile.Name()
	defer os.Remove(tmpName)

	written, copyErr := io.Copy(tmpFile, resp.Body)
	closeErr := tmpFile.Close()
	if copyErr != nil {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("write asset %s failed: %w", asset.Name, copyErr)
	}
	if closeErr != nil {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("close temp asset %s failed: %w", asset.Name, closeErr)
	}
	if written <= 0 {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("download asset %s produced an empty file", asset.Name)
	}
	if asset.Size > 0 && written != asset.Size {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("download asset %s size mismatch: expected=%d got=%d", asset.Name, asset.Size, written)
	}
	if resp.ContentLength > 0 && written != resp.ContentLength {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("download asset %s content length mismatch: expected=%d got=%d", asset.Name, resp.ContentLength, written)
	}
	if err := os.Rename(tmpName, stagedPath); err != nil {
		return stagedCCSwitchMirrorAsset{}, fmt.Errorf("stage asset %s failed: %w", asset.Name, err)
	}

	return stagedCCSwitchMirrorAsset{
		asset: CCSwitchMirrorAsset{
			ID:             assetID,
			Name:           asset.Name,
			FileName:       fileName,
			StoredFileName: storedFileName,
			RelativePath:   relativePath,
			SourceURL:      asset.BrowserDownloadURL,
			Size:           written,
			Version:        release.TagName,
			MirroredAt:     time.Now().Unix(),
		},
		stagedPath: stagedPath,
		finalPath:  finalPath,
	}, nil
}

func sleepCCSwitchMirrorRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func selectCCSwitchReleaseAssets(assets []ccSwitchGitHubAsset, spec ccSwitchMirrorPlatformSpec) []ccSwitchGitHubAsset {
	extension := strings.ToLower(spec.Extension)
	selected := make([]ccSwitchGitHubAsset, 0)
	for _, asset := range assets {
		if strings.HasSuffix(strings.ToLower(asset.Name), extension) {
			selected = append(selected, asset)
		}
	}
	sort.SliceStable(selected, func(i, j int) bool {
		return strings.ToLower(selected[i].Name) < strings.ToLower(selected[j].Name)
	})
	if len(selected) <= 1 {
		return selected
	}
	for _, preferred := range spec.PreferredNameContains {
		preferred = strings.ToLower(preferred)
		for _, asset := range selected {
			if strings.Contains(strings.ToLower(asset.Name), preferred) {
				return []ccSwitchGitHubAsset{asset}
			}
		}
	}
	return selected[:1]
}

func reuseCurrentCCSwitchAssets(oldAssets []CCSwitchMirrorAsset, version string, candidates []ccSwitchGitHubAsset) ([]CCSwitchMirrorAsset, bool) {
	if len(candidates) == 0 {
		return nil, false
	}
	currentByID := make(map[string]CCSwitchMirrorAsset, len(oldAssets))
	for _, asset := range oldAssets {
		currentByID[asset.ID] = asset
	}
	reused := make([]CCSwitchMirrorAsset, 0, len(candidates))
	for _, candidate := range candidates {
		current, ok := currentByID[ccSwitchAssetID(candidate.Name)]
		if !ok {
			return nil, false
		}
		if current.Version != version || current.Name != candidate.Name || current.Size != candidate.Size {
			return nil, false
		}
		if !ccSwitchMirrorFileExists(current.RelativePath) {
			return nil, false
		}
		reused = append(reused, current)
	}
	return reused, true
}

func loadCCSwitchMirrorManifest() CCSwitchMirrorManifest {
	manifest := defaultCCSwitchMirrorManifest()
	path := filepath.Join(getCCSwitchMirrorDir(), ccSwitchMirrorManifestName)
	f, err := os.Open(path)
	if err != nil {
		return manifest
	}
	defer f.Close()
	if err := common.DecodeJson(f, &manifest); err != nil {
		logger.LogWarn(context.Background(), fmt.Sprintf("cc-switch mirror manifest parse failed: %v", err))
		return defaultCCSwitchMirrorManifest()
	}
	if manifest.Platforms == nil {
		manifest.Platforms = make(map[string]CCSwitchMirrorPlatform)
	}
	return manifest
}

func saveCCSwitchMirrorManifest(mirrorDir string, manifest CCSwitchMirrorManifest) error {
	data, err := common.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("marshal manifest failed: %w", err)
	}
	tmpPath := filepath.Join(mirrorDir, ccSwitchMirrorManifestName+".tmp")
	finalPath := filepath.Join(mirrorDir, ccSwitchMirrorManifestName)
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write manifest failed: %w", err)
	}
	if err := os.Rename(tmpPath, finalPath); err != nil {
		return fmt.Errorf("replace manifest failed: %w", err)
	}
	return nil
}

func defaultCCSwitchMirrorManifest() CCSwitchMirrorManifest {
	return CCSwitchMirrorManifest{
		RepositoryURL: ccSwitchMirrorRepositoryURL,
		Platforms:     make(map[string]CCSwitchMirrorPlatform),
	}
}

func cloneCCSwitchMirrorManifest(manifest CCSwitchMirrorManifest) CCSwitchMirrorManifest {
	cloned := CCSwitchMirrorManifest{
		RepositoryURL: manifest.RepositoryURL,
		ReleaseURL:    manifest.ReleaseURL,
		Platforms:     make(map[string]CCSwitchMirrorPlatform, len(manifest.Platforms)),
	}
	for key, platform := range manifest.Platforms {
		assets := make([]CCSwitchMirrorAsset, len(platform.Assets))
		copy(assets, platform.Assets)
		platform.Assets = assets
		cloned.Platforms[key] = platform
	}
	return cloned
}

func cleanupReplacedCCSwitchFiles(mirrorDir string, oldAssets []CCSwitchMirrorAsset, newAssets []CCSwitchMirrorAsset) {
	keep := make(map[string]struct{}, len(newAssets))
	for _, asset := range newAssets {
		keep[asset.RelativePath] = struct{}{}
	}
	for _, asset := range oldAssets {
		if _, ok := keep[asset.RelativePath]; ok {
			continue
		}
		path := filepath.Join(mirrorDir, filepath.FromSlash(asset.RelativePath))
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			logger.LogWarn(context.Background(), fmt.Sprintf("cc-switch mirror cleanup failed: %v", err))
		}
	}
}

func ccSwitchMirrorFileExists(relativePath string) bool {
	path := ccSwitchMirrorAbsolutePath(relativePath)
	if path == "" {
		return false
	}
	_, err := os.Stat(path)
	return err == nil
}

func ccSwitchMirrorAbsolutePath(relativePath string) string {
	if relativePath == "" || filepath.IsAbs(relativePath) {
		return ""
	}
	clean := filepath.Clean(filepath.FromSlash(relativePath))
	if clean == "." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || clean == ".." {
		return ""
	}
	return filepath.Join(getCCSwitchMirrorDir(), clean)
}

func getCCSwitchMirrorDir() string {
	return common.GetEnvOrDefaultString("CC_SWITCH_MIRROR_DIR", ccSwitchMirrorDefaultDir)
}

func validateCCSwitchOutboundURL(url string) error {
	fetchSetting := system_setting.GetFetchSetting()
	if err := common.ValidateURLWithFetchSetting(url, fetchSetting.EnableSSRFProtection, fetchSetting.AllowPrivateIp, fetchSetting.DomainFilterMode, fetchSetting.IpFilterMode, fetchSetting.DomainList, fetchSetting.IpList, fetchSetting.AllowedPorts, fetchSetting.ApplyIPFilterForDomain); err != nil {
		return fmt.Errorf("request reject: %v", err)
	}
	return nil
}

func sanitizeCCSwitchFileName(name, extension string) string {
	base := filepath.Base(strings.TrimSpace(name))
	if base == "" || base == "." || base == string(os.PathSeparator) {
		base = "cc-switch" + extension
	}
	var builder strings.Builder
	for _, r := range base {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-' || r == '+' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune('_')
	}
	sanitized := builder.String()
	if sanitized == "" || sanitized == "." || sanitized == ".." {
		return "cc-switch" + extension
	}
	return sanitized
}

func ccSwitchAssetID(name string) string {
	sum := sha256.Sum256([]byte(name))
	return hex.EncodeToString(sum[:])[:16]
}

func durationUntilNextShanghaiMidnight(now time.Time) time.Duration {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		loc = time.FixedZone("Asia/Shanghai", 8*60*60)
	}
	localNow := now.In(loc)
	next := time.Date(localNow.Year(), localNow.Month(), localNow.Day()+1, 0, 0, 0, 0, loc)
	return next.Sub(localNow)
}
