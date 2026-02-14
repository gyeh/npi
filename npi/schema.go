package npi

import "fmt"

// MainSchemaColumnCount is the total number of columns in the main NPPES file.
const MainSchemaColumnCount = 330

// MainSchemaColumnNames returns all column names in exact order for the main CSV.
func MainSchemaColumnNames() []string {
	columns := []string{
		// Core identifiers (0-3)
		"NPI",
		"Entity Type Code",
		"Replacement NPI",
		"Employer Identification Number (EIN)",

		// Organization name (4)
		"Provider Organization Name (Legal Business Name)",

		// Individual provider name (5-10)
		"Provider Last Name (Legal Name)",
		"Provider First Name",
		"Provider Middle Name",
		"Provider Name Prefix Text",
		"Provider Name Suffix Text",
		"Provider Credential Text",

		// Other organization name (11-12)
		"Provider Other Organization Name",
		"Provider Other Organization Name Type Code",

		// Other individual name (13-19)
		"Provider Other Last Name",
		"Provider Other First Name",
		"Provider Other Middle Name",
		"Provider Other Name Prefix Text",
		"Provider Other Name Suffix Text",
		"Provider Other Credential Text",
		"Provider Other Last Name Type Code",

		// Mailing address (20-27)
		"Provider First Line Business Mailing Address",
		"Provider Second Line Business Mailing Address",
		"Provider Business Mailing Address City Name",
		"Provider Business Mailing Address State Name",
		"Provider Business Mailing Address Postal Code",
		"Provider Business Mailing Address Country Code (If outside U.S.)",
		"Provider Business Mailing Address Telephone Number",
		"Provider Business Mailing Address Fax Number",

		// Practice location address (28-35)
		"Provider First Line Business Practice Location Address",
		"Provider Second Line Business Practice Location Address",
		"Provider Business Practice Location Address City Name",
		"Provider Business Practice Location Address State Name",
		"Provider Business Practice Location Address Postal Code",
		"Provider Business Practice Location Address Country Code (If outside U.S.)",
		"Provider Business Practice Location Address Telephone Number",
		"Provider Business Practice Location Address Fax Number",

		// Dates (36-40)
		"Provider Enumeration Date",
		"Last Update Date",
		"NPI Deactivation Reason Code",
		"NPI Deactivation Date",
		"NPI Reactivation Date",

		// Provider gender (41)
		"Provider Sex Code",

		// Authorized official (42-46)
		"Authorized Official Last Name",
		"Authorized Official First Name",
		"Authorized Official Middle Name",
		"Authorized Official Title or Position",
		"Authorized Official Telephone Number",
	}

	// Taxonomy columns (47-106): 15 sets of 4 columns each
	for i := 1; i <= MaxTaxonomyCodes; i++ {
		columns = append(columns,
			fmt.Sprintf("Healthcare Provider Taxonomy Code_%d", i),
			fmt.Sprintf("Provider License Number_%d", i),
			fmt.Sprintf("Provider License Number State Code_%d", i),
			fmt.Sprintf("Healthcare Provider Primary Taxonomy Switch_%d", i),
		)
	}

	// Other identifier columns (107-306): 50 sets of 4 columns each
	for i := 1; i <= MaxOtherIdentifiers; i++ {
		columns = append(columns,
			fmt.Sprintf("Other Provider Identifier_%d", i),
			fmt.Sprintf("Other Provider Identifier Type Code_%d", i),
			fmt.Sprintf("Other Provider Identifier State_%d", i),
			fmt.Sprintf("Other Provider Identifier Issuer_%d", i),
		)
	}

	// Organization flags (307-310)
	columns = append(columns,
		"Is Sole Proprietor",
		"Is Organization Subpart",
		"Parent Organization LBN",
		"Parent Organization TIN",
	)

	// Authorized official additional fields (311-313)
	columns = append(columns,
		"Authorized Official Name Prefix Text",
		"Authorized Official Name Suffix Text",
		"Authorized Official Credential Text",
	)

	// Taxonomy group columns (314-328)
	for i := 1; i <= MaxTaxonomyCodes; i++ {
		columns = append(columns, fmt.Sprintf("Healthcare Provider Taxonomy Group_%d", i))
	}

	// Certification date (329)
	columns = append(columns, "Certification Date")

	return columns
}

// OtherNameSchemaColumnNames returns column names for the other name CSV.
func OtherNameSchemaColumnNames() []string {
	return []string{
		"NPI",
		"Provider Other Organization Name",
		"Provider Other Organization Name Type Code",
	}
}

// OtherNameSchemaColumnCount is the number of columns in the other name file.
const OtherNameSchemaColumnCount = 3

// PracticeLocationSchemaColumnNames returns column names for the practice location CSV.
func PracticeLocationSchemaColumnNames() []string {
	return []string{
		"NPI",
		"Provider Secondary Practice Location Address- Address Line 1",
		"Provider Secondary Practice Location Address-  Address Line 2",
		"Provider Secondary Practice Location Address - City Name",
		"Provider Secondary Practice Location Address - State Name",
		"Provider Secondary Practice Location Address - Postal Code",
		"Provider Secondary Practice Location Address - Country Code (If outside U.S.)",
		"Provider Secondary Practice Location Address - Telephone Number",
		"Provider Secondary Practice Location Address - Telephone Extension",
		"Provider Practice Location Address - Fax Number",
	}
}

// PracticeLocationSchemaColumnCount is the number of columns in the practice location file.
const PracticeLocationSchemaColumnCount = 10

// EndpointSchemaColumnNames returns column names for the endpoint CSV.
func EndpointSchemaColumnNames() []string {
	return []string{
		"NPI",
		"Endpoint Type",
		"Endpoint Type Description",
		"Endpoint",
		"Affiliation",
		"Endpoint Description",
		"Affiliation Legal Business Name",
		"Use Code",
		"Use Description",
		"Other Use Description",
		"Content Type",
		"Content Description",
		"Other Content Description",
		"Affiliation Address Line One",
		"Affiliation Address Line Two",
		"Affiliation Address City",
		"Affiliation Address State",
		"Affiliation Address Country",
		"Affiliation Address Postal Code",
	}
}

// EndpointSchemaColumnCount is the number of columns in the endpoint file.
const EndpointSchemaColumnCount = 19

// TaxonomySchemaColumnNames returns column names for the taxonomy reference CSV.
func TaxonomySchemaColumnNames() []string {
	return []string{
		"Code",
		"Grouping",
		"Classification",
		"Specialization",
		"Definition",
		"Notes",
		"Display Name",
		"Section",
	}
}

// TaxonomySchemaColumnCount is the number of columns in the taxonomy reference file.
const TaxonomySchemaColumnCount = 8

// ValidateHeaders validates that CSV headers match the expected schema.
func ValidateHeaders(expected, actual []string) error {
	if len(expected) != len(actual) {
		return ErrSchemaMismatch(len(expected), len(actual), nil)
	}
	for i := range expected {
		if expected[i] != actual[i] {
			return ErrSchemaMismatch(len(expected), len(actual), &SchemaMismatchDetail{
				Index:    i,
				Expected: expected[i],
				Found:    actual[i],
			})
		}
	}
	return nil
}
