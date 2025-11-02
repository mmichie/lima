package categorizer

import (
	"regexp"
	"testing"

	"github.com/mmichie/lima/internal/beancount"
	"github.com/shopspring/decimal"
)

func TestPattern_Matches(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		tx      *beancount.Transaction
		want    bool
	}{
		{
			name: "matches payee",
			pattern: Pattern{
				Pattern: "STARBUCKS",
				Regex:   regexp.MustCompile("STARBUCKS"),
				Fields:  []string{"payee"},
			},
			tx: &beancount.Transaction{
				Payee:     "STARBUCKS #12345",
				Narration: "Coffee",
			},
			want: true,
		},
		{
			name: "matches narration",
			pattern: Pattern{
				Pattern: "coffee",
				Regex:   regexp.MustCompile("(?i)coffee"),
				Fields:  []string{"narration"},
			},
			tx: &beancount.Transaction{
				Payee:     "Local Shop",
				Narration: "Morning Coffee",
			},
			want: true,
		},
		{
			name: "matches any field",
			pattern: Pattern{
				Pattern: "groceries",
				Regex:   regexp.MustCompile("(?i)groceries"),
				Fields:  []string{"any"},
			},
			tx: &beancount.Transaction{
				Payee:     "Safeway",
				Narration: "groceries",
			},
			want: true,
		},
		{
			name: "does not match wrong field",
			pattern: Pattern{
				Pattern: "STARBUCKS",
				Regex:   regexp.MustCompile("STARBUCKS"),
				Fields:  []string{"narration"},
			},
			tx: &beancount.Transaction{
				Payee:     "STARBUCKS #12345",
				Narration: "Coffee",
			},
			want: false,
		},
		{
			name: "matches with amount constraint",
			pattern: Pattern{
				Pattern:   "UBER",
				Regex:     regexp.MustCompile("UBER"),
				Fields:    []string{"payee"},
				MinAmount: floatPtr(10.0),
				MaxAmount: floatPtr(50.0),
			},
			tx: &beancount.Transaction{
				Payee:     "UBER TRIP",
				Narration: "Ride",
				Postings: []beancount.Posting{
					{
						Account: "Assets:Checking",
						Amount:  &beancount.Amount{Number: decimal.NewFromFloat(-25.0), Commodity: "USD"},
					},
					{
						Account: "Expenses:Transport",
						Amount:  &beancount.Amount{Number: decimal.NewFromFloat(25.0), Commodity: "USD"},
					},
				},
			},
			want: true,
		},
		{
			name: "fails amount constraint - too low",
			pattern: Pattern{
				Pattern:   "UBER",
				Regex:     regexp.MustCompile("UBER"),
				Fields:    []string{"payee"},
				MinAmount: floatPtr(50.0),
			},
			tx: &beancount.Transaction{
				Payee:     "UBER TRIP",
				Narration: "Ride",
				Postings: []beancount.Posting{
					{
						Account: "Assets:Checking",
						Amount:  &beancount.Amount{Number: decimal.NewFromFloat(-25.0), Commodity: "USD"},
					},
				},
			},
			want: false,
		},
		{
			name: "matches with required tags",
			pattern: Pattern{
				Pattern: "business",
				Regex:   regexp.MustCompile("(?i)business"),
				Fields:  []string{"narration"},
				Tags:    []string{"work", "reimbursable"},
			},
			tx: &beancount.Transaction{
				Payee:     "Office Supply Store",
				Narration: "Business supplies",
				Tags:      []string{"work", "reimbursable", "expense"},
			},
			want: true,
		},
		{
			name: "fails required tags",
			pattern: Pattern{
				Pattern: "business",
				Regex:   regexp.MustCompile("(?i)business"),
				Fields:  []string{"narration"},
				Tags:    []string{"work", "reimbursable"},
			},
			tx: &beancount.Transaction{
				Payee:     "Office Supply Store",
				Narration: "Business supplies",
				Tags:      []string{"work"}, // missing "reimbursable"
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.Matches(tt.tx); got != tt.want {
				t.Errorf("Pattern.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPattern_UpdateStatistics(t *testing.T) {
	pattern := Pattern{
		Statistics: PatternStatistics{},
	}

	// First acceptance
	pattern.UpdateStatistics(true)
	if pattern.Statistics.MatchCount != 1 {
		t.Errorf("MatchCount = %d, want 1", pattern.Statistics.MatchCount)
	}
	if pattern.Statistics.AcceptCount != 1 {
		t.Errorf("AcceptCount = %d, want 1", pattern.Statistics.AcceptCount)
	}
	if pattern.Statistics.Accuracy != 1.0 {
		t.Errorf("Accuracy = %f, want 1.0", pattern.Statistics.Accuracy)
	}

	// First rejection
	pattern.UpdateStatistics(false)
	if pattern.Statistics.MatchCount != 2 {
		t.Errorf("MatchCount = %d, want 2", pattern.Statistics.MatchCount)
	}
	if pattern.Statistics.RejectCount != 1 {
		t.Errorf("RejectCount = %d, want 1", pattern.Statistics.RejectCount)
	}
	if pattern.Statistics.Accuracy != 0.5 {
		t.Errorf("Accuracy = %f, want 0.5", pattern.Statistics.Accuracy)
	}

	// Another acceptance
	pattern.UpdateStatistics(true)
	expectedAccuracy := 2.0 / 3.0
	if pattern.Statistics.Accuracy != expectedAccuracy {
		t.Errorf("Accuracy = %f, want %f", pattern.Statistics.Accuracy, expectedAccuracy)
	}

	// Check that LastMatched was updated
	if pattern.Statistics.LastMatched.IsZero() {
		t.Error("LastMatched was not set")
	}
}

func TestSuggestionSource_Constants(t *testing.T) {
	// Test that constants are defined
	sources := []SuggestionSource{
		SourcePattern,
		SourceML,
		SourceHistory,
		SourceManual,
	}

	expected := []string{"pattern", "ml", "history", "manual"}

	for i, source := range sources {
		if string(source) != expected[i] {
			t.Errorf("Source constant %d = %s, want %s", i, source, expected[i])
		}
	}
}

func TestPattern_matchesAmount(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		tx      *beancount.Transaction
		want    bool
	}{
		{
			name: "no constraints - always matches",
			pattern: Pattern{
				MinAmount: nil,
				MaxAmount: nil,
			},
			tx: &beancount.Transaction{
				Postings: []beancount.Posting{
					{Amount: &beancount.Amount{Number: decimal.NewFromFloat(100.0)}},
				},
			},
			want: true,
		},
		{
			name: "within min and max",
			pattern: Pattern{
				MinAmount: floatPtr(10.0),
				MaxAmount: floatPtr(100.0),
			},
			tx: &beancount.Transaction{
				Postings: []beancount.Posting{
					{Amount: &beancount.Amount{Number: decimal.NewFromFloat(50.0)}},
				},
			},
			want: true,
		},
		{
			name: "below min",
			pattern: Pattern{
				MinAmount: floatPtr(100.0),
			},
			tx: &beancount.Transaction{
				Postings: []beancount.Posting{
					{Amount: &beancount.Amount{Number: decimal.NewFromFloat(50.0)}},
				},
			},
			want: false,
		},
		{
			name: "above max",
			pattern: Pattern{
				MaxAmount: floatPtr(50.0),
			},
			tx: &beancount.Transaction{
				Postings: []beancount.Posting{
					{Amount: &beancount.Amount{Number: decimal.NewFromFloat(100.0)}},
				},
			},
			want: false,
		},
		{
			name: "handles negative amounts",
			pattern: Pattern{
				MinAmount: floatPtr(10.0),
				MaxAmount: floatPtr(100.0),
			},
			tx: &beancount.Transaction{
				Postings: []beancount.Posting{
					{Amount: &beancount.Amount{Number: decimal.NewFromFloat(-50.0)}},
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.matchesAmount(tt.tx); got != tt.want {
				t.Errorf("Pattern.matchesAmount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPattern_matchesTags(t *testing.T) {
	tests := []struct {
		name    string
		pattern Pattern
		tx      *beancount.Transaction
		want    bool
	}{
		{
			name: "no required tags - always matches",
			pattern: Pattern{
				Tags: []string{},
			},
			tx: &beancount.Transaction{
				Tags: []string{"any", "tags"},
			},
			want: true,
		},
		{
			name: "has all required tags",
			pattern: Pattern{
				Tags: []string{"work", "expense"},
			},
			tx: &beancount.Transaction{
				Tags: []string{"work", "expense", "reimbursable"},
			},
			want: true,
		},
		{
			name: "missing one required tag",
			pattern: Pattern{
				Tags: []string{"work", "expense"},
			},
			tx: &beancount.Transaction{
				Tags: []string{"work"},
			},
			want: false,
		},
		{
			name: "no tags on transaction",
			pattern: Pattern{
				Tags: []string{"work"},
			},
			tx: &beancount.Transaction{
				Tags: []string{},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.matchesTags(tt.tx); got != tt.want {
				t.Errorf("Pattern.matchesTags() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		item  string
		want  bool
	}{
		{
			name:  "found",
			slice: []string{"a", "b", "c"},
			item:  "b",
			want:  true,
		},
		{
			name:  "not found",
			slice: []string{"a", "b", "c"},
			item:  "d",
			want:  false,
		},
		{
			name:  "empty slice",
			slice: []string{},
			item:  "a",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := contains(tt.slice, tt.item); got != tt.want {
				t.Errorf("contains() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to create float pointers
func floatPtr(f float64) *float64 {
	return &f
}
