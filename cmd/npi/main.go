package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"nppes_golang/npi"

	"github.com/spf13/cobra"
)

// loadPostalCodesFromFile reads a JSON file and extracts postal codes.
// If key is provided, it navigates to that nested key first (dot-separated path,
// e.g. "new_york_city.manhattan"). Otherwise walks the entire JSON tree.
func loadPostalCodesFromFile(path, key string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read postal codes file: %w", err)
	}

	var raw interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("failed to parse postal codes JSON: %w", err)
	}

	// Navigate to the specified key path
	if key != "" {
		parts := strings.Split(key, ".")
		current := raw
		for _, part := range parts {
			m, ok := current.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("key %q: expected object at %q", key, part)
			}
			current, ok = m[part]
			if !ok {
				return nil, fmt.Errorf("key %q: %q not found", key, part)
			}
		}
		raw = current
	}

	var codes []string
	collectPostalCodes(raw, &codes)
	if len(codes) == 0 {
		return nil, fmt.Errorf("no postal codes found in %s (key=%q)", path, key)
	}
	return codes, nil
}

func collectPostalCodes(v interface{}, codes *[]string) {
	switch val := v.(type) {
	case []interface{}:
		for _, item := range val {
			if s, ok := item.(string); ok {
				*codes = append(*codes, s)
			} else {
				collectPostalCodes(item, codes)
			}
		}
	case map[string]interface{}:
		for _, item := range val {
			collectPostalCodes(item, codes)
		}
	}
}

// parsePostalCodes returns postal codes from either a file or comma-separated string.
func parsePostalCodes(file, key, csv string) ([]string, error) {
	if file != "" {
		return loadPostalCodesFromFile(file, key)
	}
	if csv != "" {
		parts := strings.Split(csv, ",")
		var codes []string
		for _, p := range parts {
			if s := strings.TrimSpace(p); s != "" {
				codes = append(codes, s)
			}
		}
		return codes, nil
	}
	return nil, nil
}

func logf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "npcli",
		Short: "NPPES Data CLI - Query, analyze, and export NPPES healthcare provider data",
	}

	rootCmd.AddCommand(statsCmd())
	rootCmd.AddCommand(queryCmd())
	rootCmd.AddCommand(exportCmd())
	rootCmd.AddCommand(downloadCmd())

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func statsCmd() *cobra.Command {
	var dataDir string
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show summary statistics for a dataset",
		Run: func(cmd *cobra.Command, args []string) {
			dataset, err := npi.LoadStandard(dataDir, npi.WithLogger(logf))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
				os.Exit(1)
			}
			stats := dataset.Statistics()
			stats.PrintSummary()
		},
	}
	cmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Path to directory containing NPPES data files")
	cmd.MarkFlagRequired("data-dir")
	return cmd
}

func queryCmd() *cobra.Command {
	var (
		dataDir         string
		state           string
		specialty       string
		npiFilter       string
		postalCode      string
		postalCodesFile string
		postalCodesKey  string
		active          bool
		limit           int
	)
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query providers by state, specialty, NPI, or postal code",
		Run: func(cmd *cobra.Command, args []string) {
			dataset, err := npi.LoadStandard(dataDir, npi.WithLogger(logf))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
				os.Exit(1)
			}

			query := dataset.Query()
			if state != "" {
				query = query.State(state)
			}
			if specialty != "" {
				query = query.Specialty(specialty)
			}
			codes, err := parsePostalCodes(postalCodesFile, postalCodesKey, postalCode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if len(codes) > 0 {
				fmt.Printf("Filtering by %d postal codes\n", len(codes))
				query = query.PostalCodes(codes)
			}
			if active {
				query = query.ActiveOnly()
			}

			results := query.Execute()

			// Apply NPI filter post-query (matching Rust CLI behavior)
			if npiFilter != "" {
				var filtered []*npi.NppesRecord
				for _, p := range results {
					if string(p.NPI) == npiFilter {
						filtered = append(filtered, p)
					}
				}
				results = filtered
			}

			for i, provider := range results {
				if i >= limit {
					break
				}
				stateStr := ""
				if provider.MailingAddress.State != nil {
					stateStr = provider.MailingAddress.State.AsCode()
				}
				fmt.Printf("%s | %s | %s | %s\n",
					provider.NPI,
					provider.DisplayName(),
					npi.OptionDisplay(provider.EntityType),
					stateStr,
				)
			}
			fmt.Printf("Total matches: %d\n", len(results))
		},
	}
	cmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Path to directory containing NPPES data files")
	cmd.Flags().StringVar(&state, "state", "", "State code (e.g. CA, NY)")
	cmd.Flags().StringVar(&specialty, "specialty", "", "Specialty (taxonomy display name)")
	cmd.Flags().StringVar(&npiFilter, "npi", "", "NPI number")
	cmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code prefix, comma-separated (e.g. 90210,90211)")
	cmd.Flags().StringVar(&postalCodesFile, "postal-codes-file", "", "JSON file containing postal codes")
	cmd.Flags().StringVar(&postalCodesKey, "postal-codes-key", "", "Dot-separated key path in postal codes JSON (e.g. new_york_city.manhattan)")
	cmd.Flags().BoolVar(&active, "active", false, "Only show active providers")
	cmd.Flags().IntVar(&limit, "limit", 20, "Limit number of results")
	cmd.MarkFlagRequired("data-dir")
	return cmd
}

func exportCmd() *cobra.Command {
	var (
		dataDir         string
		output          string
		format          string
		state           string
		specialty       string
		postalCode      string
		postalCodesFile string
		postalCodesKey  string
		active          bool
	)
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data to JSON, CSV, or SQL",
		Run: func(cmd *cobra.Command, args []string) {
			dataset, err := npi.LoadStandard(dataDir, npi.WithLogger(logf))
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading dataset: %v\n", err)
				os.Exit(1)
			}

			postalCodes, err := parsePostalCodes(postalCodesFile, postalCodesKey, postalCode)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if len(postalCodes) > 0 {
				fmt.Printf("Filtering by %d postal codes\n", len(postalCodes))
			}
			// Build a postal code matcher using the same logic as QueryBuilder
			var postalSet map[string]bool
			var postalPrefixes []string
			if len(postalCodes) > 0 {
				postalSet = make(map[string]bool, len(postalCodes))
				for _, p := range postalCodes {
					p = strings.TrimSpace(p)
					if len(p) == 5 {
						postalSet[p] = true
					} else if p != "" {
						postalPrefixes = append(postalPrefixes, p)
					}
				}
			}

			filter := func(p *npi.NppesRecord) bool {
				ok := true
				if state != "" {
					if p.MailingAddress.State == nil || p.MailingAddress.State.AsCode() != strings.ToUpper(state) {
						ok = false
					}
				}
				if len(postalCodes) > 0 && ok {
					matched := false
					for _, pc := range []string{p.MailingAddress.PostalCode, p.PracticeAddress.PostalCode} {
						if pc == "" {
							continue
						}
						zip5 := pc
						if len(zip5) > 5 {
							zip5 = zip5[:5]
						}
						if postalSet[zip5] {
							matched = true
							break
						}
						for _, pfx := range postalPrefixes {
							if strings.HasPrefix(pc, pfx) {
								matched = true
								break
							}
						}
						if matched {
							break
						}
					}
					if !matched {
						ok = false
					}
				}
				if specialty != "" && ok {
					found := false
					lower := strings.ToLower(specialty)
					for _, tc := range p.TaxonomyCodes {
						ref := dataset.GetTaxonomyDescription(tc.Code)
						if ref != nil && ref.DisplayName != "" {
							if strings.Contains(strings.ToLower(ref.DisplayName), lower) {
								found = true
								break
							}
						}
					}
					if !found {
						ok = false
					}
				}
				if active && ok {
					ok = p.IsActive()
				}
				return ok
			}

			var exportFormat npi.ExportFormat
			switch strings.ToLower(format) {
			case "json":
				exportFormat = npi.ExportFormatJSON
			case "csv":
				exportFormat = npi.ExportFormatCSV
			case "sql":
				exportFormat = npi.ExportFormatSQL
			default:
				exportFormat = npi.ExportFormatJSON
			}

			if err := npi.ExportSubset(dataset, output, filter, exportFormat); err != nil {
				fmt.Fprintf(os.Stderr, "Export error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Exported to %s\n", output)
		},
	}
	cmd.Flags().StringVarP(&dataDir, "data-dir", "d", "", "Path to directory containing NPPES data files")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output file path")
	cmd.Flags().StringVar(&format, "format", "json", "Export format (json, csv, sql)")
	cmd.Flags().StringVar(&state, "state", "", "State filter")
	cmd.Flags().StringVar(&specialty, "specialty", "", "Specialty filter")
	cmd.Flags().StringVar(&postalCode, "postal-code", "", "Postal code prefix, comma-separated (e.g. 90210,90211)")
	cmd.Flags().StringVar(&postalCodesFile, "postal-codes-file", "", "JSON file containing postal codes")
	cmd.Flags().StringVar(&postalCodesKey, "postal-codes-key", "", "Dot-separated key path in postal codes JSON (e.g. new_york_city.manhattan)")
	cmd.Flags().BoolVar(&active, "active", false, "Only export active providers")
	cmd.MarkFlagRequired("data-dir")
	cmd.MarkFlagRequired("output")
	return cmd
}

func downloadCmd() *cobra.Command {
	var outDir string
	cmd := &cobra.Command{
		Use:   "download",
		Short: "Download the latest NPPES data",
		Run: func(cmd *cobra.Command, args []string) {
			config := npi.DefaultDownloadConfig()
			config.OutputDir = outDir
			config.KeepFiles = true

			downloader := npi.NewNppesDownloaderWithConfig(config)
			downloader.Logger = logf
			files, err := downloader.DownloadLatestNPPES()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Download and extraction complete: %s\n", files.Summary())
			fmt.Printf("Files saved to %s\n", files.Directory)
		},
	}
	cmd.Flags().StringVarP(&outDir, "out-dir", "o", "", "Output directory for downloaded files")
	cmd.MarkFlagRequired("out-dir")
	return cmd
}
