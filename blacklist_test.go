package router

import (
	"github.com/stretchr/testify/assert"
	"router/internal/rules"
	"testing"
)

func TestBlacklist(t *testing.T) {
	data, err := LoadBlacklistData("examples/rules/blacklist.csv")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(data["Mail"]), 10)

	client := NewClient()

	t.Run("CheckEmailBlacklist", func(t *testing.T) {
		program, err := client.Compile("not(Mail in blacklisted.Mail)")
		assert.NoError(t, err)

		paymentIn := rules.Payment{Mail: "test1@example.com"}
		result, err := client.Run(program, paymentIn, data)
		assert.NoError(t, err)
		assert.False(t, result.(bool))
	})

	t.Run("CheckNameBlacklistRegex", func(t *testing.T) {
		program, err := client.Compile("not(Name matches '^User1$')")
		assert.NoError(t, err)

		paymentIn := rules.Payment{Name: "User1"}
		result, err := client.Run(program, paymentIn, data)
		assert.NoError(t, err)
		assert.False(t, result.(bool))
	})

	t.Run("CheckBalanceBlacklist", func(t *testing.T) {
		program, err := client.Compile("not(Amount in blacklisted.Balance)")
		assert.NoError(t, err)

		paymentIn := rules.Payment{Amount: 1000.50}
		result, err := client.Run(program, paymentIn, data)
		assert.NoError(t, err)
		assert.False(t, result.(bool))

		paymentOut := rules.Payment{Amount: 999.0}
		result, err = client.Run(program, paymentOut, data)
		assert.NoError(t, err)
		assert.True(t, result.(bool))
	})

	t.Run("CheckActiveFlag", func(t *testing.T) {
		program, err := client.Compile("not(IsActive in blacklisted.IsActive)")
		assert.NoError(t, err)

		paymentIn := rules.Payment{IsActive: false}
		result, err := client.Run(program, paymentIn, data)
		assert.NoError(t, err)
		assert.False(t, result.(bool))

		paymentOut := rules.Payment{IsActive: true}
		result, err = client.Run(program, paymentOut, data)
		assert.NoError(t, err)
		assert.False(t, result.(bool))
	})
}
