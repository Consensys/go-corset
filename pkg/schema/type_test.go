package schema

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/stretchr/testify/assert"
)

func TestFieldTypeValidation(t *testing.T) {
	field := &FieldType{}

	// Test BitWidth returns fieldElementBitWidth
	assert.Equal(t, uint(fieldElementBitWidth), field.BitWidth(), "FieldType.BitWidth() should return fieldElementBitWidth")

	// Test ByteWidth returns ceiling of fieldElementBitWidth/8
	assert.Equal(t, uint((fieldElementBitWidth+7)/8), field.ByteWidth(), "FieldType.ByteWidth() should return ceiling of fieldElementBitWidth/8")

	// Test Accept with various values
	// Note: fr.Element automatically reduces values modulo r (the BN254 scalar field order)
	// so all fr.Element values are valid field elements
	tests := []struct {
		name  string
		value *big.Int
	}{
		{
			name:  "253 bit number",
			value: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth-1), big.NewInt(1)),
		},
		{
			name:  "254 bit number",
			value: new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth), big.NewInt(1)),
		},
		{
			name:  "255 bit number (will be reduced mod r)",
			value: new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth),
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
