package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hazyforge/hazyctl/pkg/version"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/minio/selfupdate"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "ALPHA: Update hazyctl to the latest version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔍 Checking for updates...")

		currentVersion := version.Version
		latestVersion := getLatestVersion()

		fmt.Printf("Current version: %s\n", currentVersion)
		fmt.Printf("Latest version: %s\n", latestVersion)

		if currentVersion == latestVersion {
			fmt.Println("✅ Already using the latest version")
			return
		}

		fmt.Printf("🔄 Updating from %s to %s...\n", currentVersion, latestVersion)

		// Create a temporary directory for our downloaded archive
		tempDir, err := os.MkdirTemp("", "hazyctl-update-")
		if err != nil {
			panic(fmt.Errorf("failed to create temp dir: %w", err))
		}
		defer os.RemoveAll(tempDir)

		// Locate the current executable
		exePath, err := os.Executable()
		fmt.Println("exePath", exePath)
		if err != nil {
			panic(fmt.Errorf("failed to determine current exe path: %w", err))
		}

		archivePath := downloadArchive(tempDir, latestVersion)
		verifyChecksum(archivePath, latestVersion)
		newBinaryPath := extractBinary(tempDir, archivePath, latestVersion)

		// Replace current binary
		replaceCurrentBinary(exePath, newBinaryPath)

		fmt.Println("🎉 Update successful! Restart hazyctl to use the new version.")
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func getLatestVersion() string {
	resp, err := http.Get("https://api.github.com/repos/hazyforge/hazyctl/releases/latest")
	if err != nil {
		panic(fmt.Sprintf("Error checking version: %v", err))
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	// Basic, naive parse to get the "tag_name" from JSON:
	var releaseData struct {
		TagName string `json:"tag_name"`
	}

	if err := json.Unmarshal(body, &releaseData); err != nil {
		panic(fmt.Sprintf("Error parsing JSON response: %v", err))
	}

	return strings.TrimPrefix(releaseData.TagName, "v")
}

func downloadArchive(tempDir, version string) string {
	goos := runtime.GOOS
	arch := runtime.GOARCH
	ext := "tar.gz"

	if goos == "windows" {
		ext = "zip"
	}

	assetName := fmt.Sprintf("hazyctl_%s_%s_%s.%s", version, goos, arch, ext)
	baseURL := fmt.Sprintf("https://github.com/hazyforge/hazyctl/releases/download/v%s", version)

	url := fmt.Sprintf("%s/%s", baseURL, assetName)

	// Download the archive
	resp, err := http.Get(url)
	if err != nil {
		exitError("Error downloading update: %v", err)
	}
	defer resp.Body.Close()

	archivePath := filepath.Join(tempDir, assetName)
	out, err := os.Create(archivePath)
	if err != nil {
		exitError("Error creating temp file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		exitError("Error writing update: %v", err)
	}

	return archivePath
}

func verifyChecksum(archivePath, version string) {
	checksumUrl := fmt.Sprintf("https://github.com/HazyForge/hazyctl/releases/download/v%s/%s.sha256", version, filepath.Base(archivePath))
	resp, err := http.Get(checksumUrl)
	if err != nil {
		exitError("Error downloading checksum: %v", err)
	}
	defer resp.Body.Close()

	checksum, _ := io.ReadAll(resp.Body)
	expectedHash := strings.TrimSpace(string(checksum))
	hash, err := fileSHA256(archivePath)
	if err != nil {
		exitError("Error verifying checksum: %v", err)
	}
	actualHash := fmt.Sprintf("%s  %s", hash, filepath.Base(archivePath))


	if actualHash != expectedHash {
		exitError("Checksum mismatch!\nExpected: %s\nActual:   %s", expectedHash, actualHash)
	}
}

func extractBinary(tempDir, archivePath, version string) string {
	switch {
	case strings.HasSuffix(archivePath, ".tar.gz"):
		return extractTarGz(archivePath, tempDir)
	case strings.HasSuffix(archivePath, ".zip"):
		return extractZip(archivePath, tempDir)
	default:
		exitError("Unsupported archive format: %s", filepath.Ext(archivePath))
	}
	return ""
}

func extractTarGz(archivePath, tempDir string) string {
	file, err := os.Open(archivePath)
	if err != nil {
		exitError("Error opening archive: %v", err)
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		exitError("Error reading gzip: %v", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			exitError("Error reading tar: %v", err)
		}

		if strings.Contains(hdr.Name, "/hazyctl") && !strings.HasSuffix(hdr.Name, "/") {
			return extractFileFromArchive(tr, tempDir, hdr.Name)
		}
	}
	exitError("No binary found in archive")
	return ""
}

func extractZip(archivePath, tempDir string) string {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		exitError("Error opening zip: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.Contains(f.Name, "hazyctl") && !f.FileInfo().IsDir() {
			return extractFileFromZip(f, tempDir)
		}
	}
	exitError("No binary found in zip archive")
	return ""
}

func extractFileFromArchive(tr *tar.Reader, tempDir, filename string) string {
	outPath := filepath.Join(tempDir, filepath.Base(filename))
	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		exitError("Error creating file: %v", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, tr); err != nil {
		exitError("Error extracting file: %v", err)
	}

	return outPath
}

func extractFileFromZip(f *zip.File, tempDir string) string {
	outPath := filepath.Join(tempDir, filepath.Base(f.Name))
	outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		exitError("Error creating file: %v", err)
	}
	defer outFile.Close()

	rc, err := f.Open()
	if err != nil {
		exitError("Error opening zip file: %v", err)
	}
	defer rc.Close()

	if _, err := io.Copy(outFile, rc); err != nil {
		exitError("Error extracting file: %v", err)
	}

	return outPath
}

func replaceCurrentBinary(currentBinPath, freshBinPath string) {
	newFile, err := os.Open(freshBinPath)
	if err != nil {
		panic(fmt.Errorf("failed to open new binary: %w", err))
	}
	defer newFile.Close()

	// Overwrite the current binary with the new one
	if err := selfupdate.Apply(newFile, selfupdate.Options{}); err != nil {
		panic(fmt.Errorf("failed to apply update: %w", err))
	}
	fmt.Printf("Replaced %s with %s\n", currentBinPath, freshBinPath)
}

func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func findMatchingChecksum(checksums, filename string) string {
	for _, line := range strings.Split(checksums, "\n") {
		if strings.Contains(line, filename) {
			return strings.Split(line, " ")[0]
		}
	}
	panic(fmt.Sprintf("No checksum found for %s", filename))
}

func exitError(format string, args ...interface{}) {
	fmt.Printf("❌ "+format+"\n", args...)
	os.Exit(1)
}
