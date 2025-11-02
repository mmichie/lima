package beancount

import (
	"time"

	"github.com/shopspring/decimal"
)

// Directive represents any Beancount directive
type Directive interface {
	GetDate() time.Time
	GetType() DirectiveType
}

// DirectiveType represents the type of Beancount directive
type DirectiveType string

const (
	DirectiveTypeTransaction DirectiveType = "transaction"
	DirectiveTypeOpen        DirectiveType = "open"
	DirectiveTypeClose       DirectiveType = "close"
	DirectiveTypeBalance     DirectiveType = "balance"
	DirectiveTypePrice       DirectiveType = "price"
	DirectiveTypeCommodity   DirectiveType = "commodity"
	DirectiveTypePad         DirectiveType = "pad"
	DirectiveTypeNote        DirectiveType = "note"
	DirectiveTypeDocument    DirectiveType = "document"
	DirectiveTypeEvent       DirectiveType = "event"
	DirectiveTypeQuery       DirectiveType = "query"
	DirectiveTypeCustom      DirectiveType = "custom"
)

// Transaction represents a Beancount transaction
type Transaction struct {
	Date      time.Time
	Flag      string // "*" (cleared) or "!" (pending)
	Payee     string
	Narration string
	Tags      []string
	Links     []string
	Postings  []Posting
	Metadata  map[string]string

	// For lazy loading - track position in file
	FilePosition int64
	LineNumber   int
}

func (t Transaction) GetDate() time.Time       { return t.Date }
func (t Transaction) GetType() DirectiveType   { return DirectiveTypeTransaction }

// Posting represents a single posting within a transaction
type Posting struct {
	Account  string
	Amount   *Amount // nil for auto-balanced postings
	Cost     *Amount // cost basis (optional)
	Price    *Amount // price (optional)
	Metadata map[string]string
}

// Amount represents a monetary amount with commodity
type Amount struct {
	Number    decimal.Decimal
	Commodity string
}

// OpenAccount represents an account opening directive
type OpenAccount struct {
	Date        time.Time
	Account     string
	Commodities []string
	Metadata    map[string]string
	LineNumber  int
}

func (o OpenAccount) GetDate() time.Time     { return o.Date }
func (o OpenAccount) GetType() DirectiveType { return DirectiveTypeOpen }

// CloseAccount represents an account closing directive
type CloseAccount struct {
	Date       time.Time
	Account    string
	Metadata   map[string]string
	LineNumber int
}

func (c CloseAccount) GetDate() time.Time     { return c.Date }
func (c CloseAccount) GetType() DirectiveType { return DirectiveTypeClose }

// Balance represents a balance assertion
type Balance struct {
	Date       time.Time
	Account    string
	Amount     Amount
	Metadata   map[string]string
	LineNumber int
}

func (b Balance) GetDate() time.Time     { return b.Date }
func (b Balance) GetType() DirectiveType { return DirectiveTypeBalance }

// Price represents a price directive
type Price struct {
	Date       time.Time
	Commodity  string
	Amount     Amount
	Metadata   map[string]string
	LineNumber int
}

func (p Price) GetDate() time.Time     { return p.Date }
func (p Price) GetType() DirectiveType { return DirectiveTypePrice }

// Commodity represents a commodity declaration
type Commodity struct {
	Date       time.Time
	Name       string
	Metadata   map[string]string
	LineNumber int
}

func (c Commodity) GetDate() time.Time     { return c.Date }
func (c Commodity) GetType() DirectiveType { return DirectiveTypeCommodity }

// Pad represents a pad directive
type Pad struct {
	Date       time.Time
	Account    string
	SourceAccount string
	Metadata   map[string]string
	LineNumber int
}

func (p Pad) GetDate() time.Time     { return p.Date }
func (p Pad) GetType() DirectiveType { return DirectiveTypePad }

// Note represents a note directive
type Note struct {
	Date       time.Time
	Account    string
	Comment    string
	Metadata   map[string]string
	LineNumber int
}

func (n Note) GetDate() time.Time     { return n.Date }
func (n Note) GetType() DirectiveType { return DirectiveTypeNote }
