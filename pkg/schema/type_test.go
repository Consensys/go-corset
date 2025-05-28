package schema

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/stretchr/testify/assert"
)

func TestFieldTypeValidation(t *testing.T) {
	field := &FieldType{}

	// Test BitWidth returns 254
	assert.Equal(t, uint(254), field.BitWidth(), "FieldType.BitWidth() should return 254")

	// Test ByteWidth returns 32 (ceiling of 254/8)
	assert.Equal(t, uint(32), field.ByteWidth(), "FieldType.ByteWidth() should return 32")

	// Test Accept with various values
	// Note: fr.Element automatically reduces values modulo r (the BLS12-377 scalar field order)
	// so all fr.Element values are valid field elements
	tests := []struct {
		name  string
		value *big.Int
	}{
		{
			name:  "253 bit number",
			value: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 253), big.NewInt(1)),
		},
		{
			name:  "254 bit number",
			value: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 254), big.NewInt(1)),
		},
		{
			name:  "255 bit number (will be reduced mod r)",
			value: new(big.Int).Lsh(big.NewInt(1), 254),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var elem fr.Element
			elem.SetBigInt(tt.value)
			// All fr.Element values should be accepted since they are automatically reduced mod r
			assert.True(t, field.Accept(elem), "FieldType.Accept() should accept all fr.Element values")
		})
	}
}
