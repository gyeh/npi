package npi

import "fmt"

// NppesError represents errors that can occur during NPPES operations.
type NppesError struct {
	Kind    string
	Message string
	Detail  string
}

func (e *NppesError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Kind, e.Message, e.Detail)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

func ErrFileNotFound(path string) error {
	return &NppesError{
		Kind:    "FileNotFound",
		Message: fmt.Sprintf("file not found: %s", path),
		Detail:  "check the file path and try again",
	}
}

func ErrInvalidNPI(npi string) error {
	return &NppesError{
		Kind:    "InvalidNPI",
		Message: fmt.Sprintf("invalid NPI: %s", npi),
		Detail:  "NPI must be exactly 10 digits",
	}
}

func ErrInvalidEntityType(code string) error {
	return &NppesError{
		Kind:    "InvalidEntityType",
		Message: fmt.Sprintf("invalid entity type code: %s", code),
		Detail:  "valid codes are '1' (Individual) or '2' (Organization)",
	}
}

func ErrSchemaMismatch(expected, found int, mismatchCol *SchemaMismatchDetail) error {
	msg := fmt.Sprintf("schema mismatch: expected %d columns, found %d", expected, found)
	detail := ""
	if mismatchCol != nil {
		detail = fmt.Sprintf("column %d: expected '%s', found '%s'", mismatchCol.Index, mismatchCol.Expected, mismatchCol.Found)
	}
	return &NppesError{
		Kind:    "SchemaMismatch",
		Message: msg,
		Detail:  detail,
	}
}

type SchemaMismatchDetail struct {
	Index    int
	Expected string
	Found    string
}

func ErrDateParse(value, expectedFormat string) error {
	return &NppesError{
		Kind:    "DateParse",
		Message: fmt.Sprintf("failed to parse date: %s", value),
		Detail:  fmt.Sprintf("expected format: %s", expectedFormat),
	}
}

func ErrCsvParse(message string, line int) error {
	return &NppesError{
		Kind:    "CsvParse",
		Message: message,
		Detail:  fmt.Sprintf("line %d", line),
	}
}

func ErrDataValidation(message, field string) error {
	return &NppesError{
		Kind:    "DataValidation",
		Message: message,
		Detail:  fmt.Sprintf("field: %s", field),
	}
}

func ErrExport(message string) error {
	return &NppesError{
		Kind:    "Export",
		Message: message,
	}
}

func ErrCustom(message string) error {
	return &NppesError{
		Kind:    "Error",
		Message: message,
	}
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(bytes uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unitIndex := 0
	for size >= 1024.0 && unitIndex < len(units)-1 {
		size /= 1024.0
		unitIndex++
	}
	return fmt.Sprintf("%.2f %s", size, units[unitIndex])
}
