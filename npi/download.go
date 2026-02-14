package npi

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Downloader is the interface for downloading and extracting NPPES data.
type Downloader interface {
	DownloadFile(url string, filename string) (string, error)
	DownloadAndExtractZip(url string) (*ExtractedFiles, error)
	ExtractZip(zipPath string) (*ExtractedFiles, error)
	DownloadLatestNPPES() (*ExtractedFiles, error)
}

var _ Downloader = (*NppesDownloader)(nil)

// DownloadConfig configures download behavior.
type DownloadConfig struct {
	TimeoutSeconds int
	MaxFileSize    int64
	OutputDir      string
	KeepFiles      bool
	UserAgent      string
}

// DefaultDownloadConfig returns default download settings.
func DefaultDownloadConfig() DownloadConfig {
	return DownloadConfig{
		TimeoutSeconds: 300,
		MaxFileSize:    20 * 1024 * 1024 * 1024, // 20GB
		OutputDir:      "",
		KeepFiles:      false,
		UserAgent:      "npi-golang/1.0",
	}
}

// NppesDownloader handles downloading and extracting NPPES data.
type NppesDownloader struct {
	Config DownloadConfig
	Logger func(format string, args ...interface{})
}

func (d *NppesDownloader) logf(format string, args ...interface{}) {
	if d.Logger != nil {
		d.Logger(format, args...)
	}
}

// NewNppesDownloader creates a downloader with default configuration.
func NewNppesDownloader() *NppesDownloader {
	return &NppesDownloader{Config: DefaultDownloadConfig()}
}

// NewNppesDownloaderWithConfig creates a downloader with custom configuration.
func NewNppesDownloaderWithConfig(config DownloadConfig) *NppesDownloader {
	return &NppesDownloader{Config: config}
}

// ExtractedFiles contains information about extracted NPPES files.
type ExtractedFiles struct {
	Directory             string
	Files                 []string
	MainDataFile          string
	TaxonomyFile          string
	OtherNamesFile        string
	PracticeLocationsFile string
	EndpointsFile         string
}

// HasMainData returns true if the main data file was found.
func (ef *ExtractedFiles) HasMainData() bool {
	return ef.MainDataFile != ""
}

// Summary returns a human-readable summary of found files.
func (ef *ExtractedFiles) Summary() string {
	var parts []string
	if ef.MainDataFile != "" {
		parts = append(parts, "Main Data")
	}
	if ef.TaxonomyFile != "" {
		parts = append(parts, "Taxonomy")
	}
	if ef.OtherNamesFile != "" {
		parts = append(parts, "Other Names")
	}
	if ef.PracticeLocationsFile != "" {
		parts = append(parts, "Practice Locations")
	}
	if ef.EndpointsFile != "" {
		parts = append(parts, "Endpoints")
	}
	if len(parts) == 0 {
		return "No recognized NPPES files found"
	}
	return "Found: " + strings.Join(parts, ", ")
}

// DownloadFile downloads a file from a URL.
func (d *NppesDownloader) DownloadFile(url string, filename string) (string, error) {
	d.logf("Downloading from: %s\n", url)

	client := &http.Client{
		Timeout: time.Duration(d.Config.TimeoutSeconds) * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", ErrCustom(fmt.Sprintf("failed to download: %v", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrCustom(fmt.Sprintf("HTTP error %d: %s", resp.StatusCode, url))
	}

	outputDir := d.Config.OutputDir
	if outputDir == "" {
		outputDir = os.TempDir()
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", ErrCustom(fmt.Sprintf("failed to create directory: %v", err))
	}

	if filename == "" {
		parts := strings.Split(url, "/")
		filename = parts[len(parts)-1]
		if filename == "" {
			filename = "nppes_download"
		}
	}

	filePath := filepath.Join(outputDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", ErrCustom(fmt.Sprintf("failed to create file: %v", err))
	}
	defer f.Close()

	written, err := io.Copy(f, resp.Body)
	if err != nil {
		return "", ErrCustom(fmt.Sprintf("error during download: %v", err))
	}

	d.logf("Downloaded %s to %s\n", FormatBytes(uint64(written)), filePath)
	return filePath, nil
}

// DownloadAndExtractZip downloads a ZIP file and extracts it.
func (d *NppesDownloader) DownloadAndExtractZip(url string) (*ExtractedFiles, error) {
	zipPath, err := d.DownloadFile(url, "")
	if err != nil {
		return nil, err
	}

	extracted, err := d.ExtractZip(zipPath)
	if err != nil {
		return nil, err
	}

	if !d.Config.KeepFiles {
		os.Remove(zipPath)
	}

	return extracted, nil
}

// ExtractZip extracts a ZIP file and categorizes the extracted files.
func (d *NppesDownloader) ExtractZip(zipPath string) (*ExtractedFiles, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, ErrCustom(fmt.Sprintf("failed to open ZIP file: %v", err))
	}
	defer r.Close()

	extractDir := d.Config.OutputDir
	if extractDir == "" {
		extractDir = os.TempDir()
	}
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return nil, ErrCustom(fmt.Sprintf("failed to create directory: %v", err))
	}

	d.logf("Extracting ZIP file to: %s\n", extractDir)

	extracted := &ExtractedFiles{
		Directory: extractDir,
	}

	for _, f := range r.File {
		filePath := filepath.Join(extractDir, f.Name)

		// Ensure parent directory exists
		if dir := filepath.Dir(filePath); dir != "" {
			os.MkdirAll(dir, 0755)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, 0755)
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return nil, ErrCustom(fmt.Sprintf("failed to read file from ZIP: %v", err))
		}

		outFile, err := os.Create(filePath)
		if err != nil {
			rc.Close()
			return nil, ErrCustom(fmt.Sprintf("failed to create file: %v", err))
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return nil, ErrCustom(fmt.Sprintf("failed to extract file: %v", err))
		}

		d.logf("Extracted: %s\n", f.Name)

		// Categorize files
		lower := strings.ToLower(f.Name)
		if strings.Contains(lower, "npidata_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			extracted.MainDataFile = filePath
		} else if strings.Contains(lower, "nucc_taxonomy") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			extracted.TaxonomyFile = filePath
		} else if strings.Contains(lower, "othername_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			extracted.OtherNamesFile = filePath
		} else if strings.Contains(lower, "pl_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			extracted.PracticeLocationsFile = filePath
		} else if strings.Contains(lower, "endpoint_pfile") && strings.HasSuffix(lower, ".csv") && !strings.Contains(lower, "_fileheader") {
			extracted.EndpointsFile = filePath
		}

		extracted.Files = append(extracted.Files, filePath)
	}

	d.logf("Extracted %d files\n", len(extracted.Files))
	return extracted, nil
}

// DownloadLatestNPPES downloads the latest NPPES data from CMS.
func (d *NppesDownloader) DownloadLatestNPPES() (*ExtractedFiles, error) {
	now := time.Now()
	month := now.Format("January") // Full month name
	year := now.Year()
	url := fmt.Sprintf("https://download.cms.gov/nppes/NPPES_Data_Dissemination_%s_%d_V2.zip", month, year)
	return d.DownloadAndExtractZip(url)
}

// LatestNPPESUrl returns the URL for the latest NPPES data download.
func LatestNPPESUrl() string {
	now := time.Now()
	month := now.Format("January")
	year := now.Year()
	return fmt.Sprintf("https://download.cms.gov/nppes/NPPES_Data_Dissemination_%s_%d_V2.zip", month, year)
}
