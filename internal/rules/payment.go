package rules

type Pepe string

const (
	GBP Pepe = "GBP"
	CAD Pepe = "CAD"
	EUR Pepe = "EUR"
)

// Payment represents the entity to be evaluated by rules.
type Payment struct {
	Amount   float64
	Currency Pepe
	UserType string
	Country  string
	User     bool
	Mail     string
	Name     string
	IsActive bool
}
