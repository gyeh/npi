package npi

import (
	"fmt"
	"sort"
	"strings"
)

// Analyzer is the interface for analytics and aggregation over provider data.
type Analyzer interface {
	FindByNPI(npi NPI) *NppesRecord
	FindByName(name string) []*NppesRecord
	FindByState(state string) []*NppesRecord
	FindByTaxonomyCode(code string) []*NppesRecord
	ProviderCountByState() []CountEntry
	ProviderCountByTaxonomy() []CountEntry
	TopStatesByProviderCount(n int) []CountEntry
	TopTaxonomyCodesByProviderCount(n int) []CountEntry
	ComputeDatasetStats() DatasetStats
}

var _ Analyzer = (*NppesAnalytics)(nil)

// NppesAnalytics provides analytics and aggregation over provider data.
type NppesAnalytics struct {
	providers   []NppesRecord
	taxonomyMap map[string]TaxonomyReference
}

// NewNppesAnalytics creates a new analytics engine.
func NewNppesAnalytics(providers []NppesRecord) *NppesAnalytics {
	return &NppesAnalytics{providers: providers}
}

// WithTaxonomyMap sets the taxonomy reference map for enriched analytics.
func (a *NppesAnalytics) WithTaxonomyMap(m map[string]TaxonomyReference) *NppesAnalytics {
	a.taxonomyMap = m
	return a
}

// FindByNPI finds a provider by NPI.
func (a *NppesAnalytics) FindByNPI(npi NPI) *NppesRecord {
	for i := range a.providers {
		if a.providers[i].NPI == npi {
			return &a.providers[i]
		}
	}
	return nil
}

// FindByName finds providers whose name contains the given string (case-insensitive).
func (a *NppesAnalytics) FindByName(name string) []*NppesRecord {
	lower := strings.ToLower(name)
	var results []*NppesRecord
	for i := range a.providers {
		if strings.Contains(strings.ToLower(a.providers[i].DisplayName()), lower) {
			results = append(results, &a.providers[i])
		}
	}
	return results
}

// FindByState finds all providers in a given state.
func (a *NppesAnalytics) FindByState(state string) []*NppesRecord {
	sc, ok := StateCodeFromCode(state)
	if !ok {
		return nil
	}
	var results []*NppesRecord
	for i := range a.providers {
		if a.providers[i].MailingAddress.State != nil && *a.providers[i].MailingAddress.State == sc {
			results = append(results, &a.providers[i])
		}
	}
	return results
}

// FindByTaxonomyCode finds all providers with a given taxonomy code.
func (a *NppesAnalytics) FindByTaxonomyCode(code string) []*NppesRecord {
	var results []*NppesRecord
	for i := range a.providers {
		for _, tc := range a.providers[i].TaxonomyCodes {
			if tc.Code == code {
				results = append(results, &a.providers[i])
				break
			}
		}
	}
	return results
}

// CountEntry is a key-count pair used for aggregation results.
type CountEntry struct {
	Key   string
	Count int
}

// ProviderCountByState returns provider counts grouped by state.
func (a *NppesAnalytics) ProviderCountByState() []CountEntry {
	counts := make(map[string]int)
	for _, p := range a.providers {
		if p.MailingAddress.State != nil {
			counts[p.MailingAddress.State.AsCode()]++
		}
	}
	return mapToSortedEntries(counts)
}

// ProviderCountByTaxonomy returns provider counts grouped by taxonomy code.
func (a *NppesAnalytics) ProviderCountByTaxonomy() []CountEntry {
	counts := make(map[string]int)
	for _, p := range a.providers {
		for _, tc := range p.TaxonomyCodes {
			counts[tc.Code]++
		}
	}
	return mapToSortedEntries(counts)
}

// TopStatesByProviderCount returns the top N states by provider count.
func (a *NppesAnalytics) TopStatesByProviderCount(n int) []CountEntry {
	entries := a.ProviderCountByState()
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// TopTaxonomyCodesByProviderCount returns the top N taxonomy codes by provider count.
func (a *NppesAnalytics) TopTaxonomyCodesByProviderCount(n int) []CountEntry {
	entries := a.ProviderCountByTaxonomy()
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n]
}

// DatasetStats holds summary stats for the analytics engine.
type DatasetStats struct {
	TotalProviders        int
	IndividualProviders   int
	OrganizationProviders int
	ActiveProviders       int
	InactiveProviders     int
	UniqueStates          int
	UniqueTaxonomyCodes   int
}

// ComputeDatasetStats calculates summary statistics.
func (a *NppesAnalytics) ComputeDatasetStats() DatasetStats {
	stats := DatasetStats{TotalProviders: len(a.providers)}
	states := make(map[string]bool)
	taxonomies := make(map[string]bool)

	for _, p := range a.providers {
		if p.EntityType != nil {
			switch *p.EntityType {
			case EntityTypeIndividual:
				stats.IndividualProviders++
			case EntityTypeOrganization:
				stats.OrganizationProviders++
			}
		}
		if p.IsActive() {
			stats.ActiveProviders++
		} else {
			stats.InactiveProviders++
		}
		if p.MailingAddress.State != nil {
			states[p.MailingAddress.State.AsCode()] = true
		}
		for _, tc := range p.TaxonomyCodes {
			taxonomies[tc.Code] = true
		}
	}

	stats.UniqueStates = len(states)
	stats.UniqueTaxonomyCodes = len(taxonomies)
	return stats
}

// PrintSummary prints a formatted summary of dataset stats.
func (s *DatasetStats) PrintSummary() {
	fmt.Println("=== Dataset Stats ===")
	fmt.Printf("Total Providers: %d\n", s.TotalProviders)
	fmt.Printf("  Individual: %d\n", s.IndividualProviders)
	fmt.Printf("  Organization: %d\n", s.OrganizationProviders)
	fmt.Printf("Active: %d\n", s.ActiveProviders)
	fmt.Printf("Inactive: %d\n", s.InactiveProviders)
	fmt.Printf("Unique States: %d\n", s.UniqueStates)
	fmt.Printf("Unique Taxonomy Codes: %d\n", s.UniqueTaxonomyCodes)
}

func mapToSortedEntries(m map[string]int) []CountEntry {
	entries := make([]CountEntry, 0, len(m))
	for k, v := range m {
		entries = append(entries, CountEntry{Key: k, Count: v})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Count > entries[j].Count
	})
	return entries
}
