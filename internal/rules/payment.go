package rules

// Payment represents the entity to be evaluated by rules.
type Payment struct {
	Amount   float64
	Currency string
	UserType string
	Country  string
	User     bool
}
