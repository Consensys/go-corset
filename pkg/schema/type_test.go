package schema

import (
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	"github.com/stretchr/testify/assert"
)

func TestFieldTypeValidation(t *testing.T) {
	field := &FieldType{}

	// Test BitWidth returns fieldElementBitWidth
	assert.Equal(t, uint(fieldElementBitWidth), field.BitWidth(), "FieldType.BitWidth() should return fieldElementBitWidth")

	// Test ByteWidth returns ceiling of fieldElementBitWidth/8
	assert.Equal(t, uint((fieldElementBitWidth+7)/8), field.ByteWidth(), "FieldType.ByteWidth() should return ceiling of fieldElementBitWidth/8")

	// Test Accept with various values
	tests := []struct {
		name     string
		value    *big.Int
		expected bool
	}{
		{
			name:     "Zero",
			value:    big.NewInt(0),
			expected: true,
		},
		{
			name:     "Small positive number",
			value:    big.NewInt(42),
			expected: true,
		},
		{
			name:     "253 bit number (valid)",
			value:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth-1), big.NewInt(1)),
			expected: true,
		},
		{
			name:     "254 bit number (max valid)",
			value:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth), big.NewInt(1)),
			expected: true,
		},
		{
			name:     "255 bit number (invalid)",
			value:    new(big.Int).Lsh(big.NewInt(1), fieldElementBitWidth),
			expected: false,
		},
		{
			name:     "256 bit number (invalid)",
			value:    new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1)),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var elem fr.Element
			elem.SetBigInt(tt.value)
			result := field.Accept(elem)
			assert.Equal(t, tt.expected, result,
				"FieldType.Accept() returned %v for %s, expected %v",
				result, tt.name, tt.expected)
		})
	}
}
