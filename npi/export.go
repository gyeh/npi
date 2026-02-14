package npi

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

// ExportFormat represents the format for data export.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
	ExportFormatSQL  ExportFormat = "sql"
)

// Exporter is the interface for exporting provider data.
type Exporter interface {
	Export(providers []*NppesRecord, path string) error
}

// JSONExporter exports providers as JSON.
type JSONExporter struct{}

// Export writes providers as a pretty-printed JSON array.
func (e *JSONExporter) Export(providers []*NppesRecord, path string) error {
	return ExportJSON(providers, path)
}

// CSVExporter exports providers as CSV.
type CSVExporter struct{}

// Export writes providers as normalized CSV files.
func (e *CSVExporter) Export(providers []*NppesRecord, path string) error {
	return ExportCSV(providers, path)
}

// SQLExporter exports providers as SQL.
type SQLExporter struct{}

// Export writes providers as PostgreSQL INSERT statements.
func (e *SQLExporter) Export(providers []*NppesRecord, path string) error {
	return ExportSQL(providers, path)
}

// NewExporter returns an Exporter for the given format.
func NewExporter(format ExportFormat) (Exporter, error) {
	switch format {
	case ExportFormatJSON:
		return &JSONExporter{}, nil
	case ExportFormatCSV:
		return &CSVExporter{}, nil
	case ExportFormatSQL:
		return &SQLExporter{}, nil
	default:
		return nil, ErrExport(fmt.Sprintf("unsupported export format: %s", format))
	}
}

// ExportJSON exports providers as a pretty-printed JSON array.
func ExportJSON(providers []*NppesRecord, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return ErrExport(fmt.Sprintf("failed to create file: %s", path))
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(providers); err != nil {
		return ErrExport(fmt.Sprintf("JSON encoding error: %v", err))
	}
	return nil
}

// ExportCSV exports providers as normalized CSV files (providers + taxonomies).
func ExportCSV(providers []*NppesRecord, basePath string) error {
	dir := "."
	baseName := basePath
	if idx := strings.LastIndex(basePath, "/"); idx >= 0 {
		dir = basePath[:idx]
		baseName = basePath[idx+1:]
	}
	// Strip extension
	if idx := strings.LastIndex(baseName, "."); idx >= 0 {
		baseName = baseName[:idx]
	}

	// Export providers
	providersPath := fmt.Sprintf("%s/%s_providers.csv", dir, baseName)
	pf, err := os.Create(providersPath)
	if err != nil {
		return ErrExport(fmt.Sprintf("failed to create file: %s", providersPath))
	}
	defer pf.Close()

	pw := csv.NewWriter(pf)
	for _, p := range providers {
		state := ""
		if p.MailingAddress.State != nil {
			state = p.MailingAddress.State.AsCode()
		}
		etCode := ""
		if p.EntityType != nil {
			etCode = p.EntityType.ToCode()
		}
		pw.Write([]string{
			string(p.NPI),
			etCode,
			p.DisplayName(),
			state,
			p.MailingAddress.PostalCode,
		})
	}
	pw.Flush()
	if err := pw.Error(); err != nil {
		return ErrExport(fmt.Sprintf("CSV write error: %v", err))
	}

	// Export taxonomies
	taxonomyPath := fmt.Sprintf("%s/%s_taxonomies.csv", dir, baseName)
	tf, err := os.Create(taxonomyPath)
	if err != nil {
		return ErrExport(fmt.Sprintf("failed to create file: %s", taxonomyPath))
	}
	defer tf.Close()

	tw := csv.NewWriter(tf)
	tw.Write([]string{"npi", "taxonomy_code", "is_primary", "license_number", "license_state"})
	for _, p := range providers {
		for _, tc := range p.TaxonomyCodes {
			isPrimary := "N"
			if tc.IsPrimary {
				isPrimary = "Y"
			}
			tw.Write([]string{
				string(p.NPI),
				tc.Code,
				isPrimary,
				tc.LicenseNumber,
				tc.LicenseState,
			})
		}
	}
	tw.Flush()
	if err := tw.Error(); err != nil {
		return ErrExport(fmt.Sprintf("CSV write error: %v", err))
	}

	return nil
}

// ExportSQL exports providers as PostgreSQL CREATE TABLE + INSERT statements.
func ExportSQL(providers []*NppesRecord, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return ErrExport(fmt.Sprintf("failed to create file: %s", path))
	}
	defer f.Close()

	w := io.Writer(f)
	writeSchema(w)
	writeProviderInserts(w, providers)
	return nil
}

func writeSchema(w io.Writer) {
	fmt.Fprintln(w, "-- NPPES Database Schema for PostgreSQL")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "CREATE TABLE IF NOT EXISTS nppes_providers (")
	fmt.Fprintln(w, "  npi VARCHAR(10) PRIMARY KEY,")
	fmt.Fprintln(w, "  entity_type SMALLINT NOT NULL,")
	fmt.Fprintln(w, "  organization_name VARCHAR(255),")
	fmt.Fprintln(w, "  last_name VARCHAR(100),")
	fmt.Fprintln(w, "  first_name VARCHAR(100),")
	fmt.Fprintln(w, "  middle_name VARCHAR(100),")
	fmt.Fprintln(w, "  mailing_address_line1 VARCHAR(255),")
	fmt.Fprintln(w, "  mailing_address_city VARCHAR(100),")
	fmt.Fprintln(w, "  mailing_address_state VARCHAR(2),")
	fmt.Fprintln(w, "  mailing_address_postal_code VARCHAR(10),")
	fmt.Fprintln(w, "  enumeration_date DATE,")
	fmt.Fprintln(w, "  last_update_date DATE,")
	fmt.Fprintln(w, "  is_active BOOLEAN DEFAULT TRUE")
	fmt.Fprintln(w, ");")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "CREATE TABLE IF NOT EXISTS nppes_taxonomies (")
	fmt.Fprintln(w, "  id SERIAL PRIMARY KEY,")
	fmt.Fprintln(w, "  npi VARCHAR(10) REFERENCES nppes_providers(npi),")
	fmt.Fprintln(w, "  taxonomy_code VARCHAR(10) NOT NULL,")
	fmt.Fprintln(w, "  is_primary BOOLEAN DEFAULT FALSE,")
	fmt.Fprintln(w, "  license_number VARCHAR(50),")
	fmt.Fprintln(w, "  license_state VARCHAR(2)")
	fmt.Fprintln(w, ");")
	fmt.Fprintln(w)

	fmt.Fprintln(w, "CREATE INDEX idx_nppes_state ON nppes_providers(mailing_address_state);")
	fmt.Fprintln(w, "CREATE INDEX idx_nppes_taxonomy ON nppes_taxonomies(taxonomy_code);")
}

func writeProviderInserts(w io.Writer, providers []*NppesRecord) {
	batchSize := 1000
	fmt.Fprintln(w, "\n-- Provider data")

	for start := 0; start < len(providers); start += batchSize {
		end := start + batchSize
		if end > len(providers) {
			end = len(providers)
		}
		chunk := providers[start:end]

		fmt.Fprintln(w, "INSERT INTO nppes_providers (npi, entity_type, organization_name, last_name, first_name, middle_name, mailing_address_line1, mailing_address_city, mailing_address_state, mailing_address_postal_code, enumeration_date, last_update_date, is_active) VALUES")

		for i, p := range chunk {
			var values string
			if p.EntityType != nil && *p.EntityType == EntityTypeOrganization {
				values = fmt.Sprintf("('%s', %s, %s, NULL, NULL, NULL, %s, %s, %s, %s, %s, %s, %t)",
					p.NPI,
					sqlVal(p.EntityType.ToCode()),
					sqlStr(p.OrganizationName.LegalBusinessName),
					sqlStr(p.MailingAddress.Line1),
					sqlStr(p.MailingAddress.City),
					sqlStateStr(p.MailingAddress.State),
					sqlStr(p.MailingAddress.PostalCode),
					sqlDate(p.EnumerationDate),
					sqlDate(p.LastUpdateDate),
					p.IsActive(),
				)
			} else if p.EntityType != nil && *p.EntityType == EntityTypeIndividual {
				values = fmt.Sprintf("('%s', %s, NULL, %s, %s, %s, %s, %s, %s, %s, %s, %s, %t)",
					p.NPI,
					sqlVal(p.EntityType.ToCode()),
					sqlStr(p.ProviderName.Last),
					sqlStr(p.ProviderName.First),
					sqlStr(p.ProviderName.Middle),
					sqlStr(p.MailingAddress.Line1),
					sqlStr(p.MailingAddress.City),
					sqlStateStr(p.MailingAddress.State),
					sqlStr(p.MailingAddress.PostalCode),
					sqlDate(p.EnumerationDate),
					sqlDate(p.LastUpdateDate),
					p.IsActive(),
				)
			} else {
				values = fmt.Sprintf("('%s', NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL, NULL)", p.NPI)
			}

			if i < len(chunk)-1 {
				fmt.Fprintf(w, "  %s,\n", values)
			} else {
				fmt.Fprintf(w, "  %s;\n", values)
			}
		}
	}
}

func sqlStr(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func sqlVal(s string) string {
	if s == "" {
		return "NULL"
	}
	return s
}

func sqlStateStr(sc *StateCode) string {
	if sc == nil {
		return "NULL"
	}
	return "'" + sc.AsCode() + "'"
}

func sqlDate(t *time.Time) string {
	if t == nil {
		return "NULL"
	}
	return "'" + t.Format("2006-01-02") + "'"
}

// ExportSubset exports a filtered subset of providers in the given format.
func ExportSubset(ds *NppesDataset, path string, filter func(*NppesRecord) bool, format ExportFormat) error {
	var filtered []*NppesRecord
	for i := range ds.Providers {
		if filter(&ds.Providers[i]) {
			filtered = append(filtered, &ds.Providers[i])
		}
	}

	switch format {
	case ExportFormatJSON:
		return ExportJSON(filtered, path)
	case ExportFormatCSV:
		return ExportCSV(filtered, path)
	case ExportFormatSQL:
		return ExportSQL(filtered, path)
	default:
		return ErrExport(fmt.Sprintf("unsupported export format: %s", format))
	}
}
