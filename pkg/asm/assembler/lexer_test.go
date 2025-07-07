package assembler

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLexerNumberRule(t *testing.T) {
	require.Equal(t, uint(4), number([]int32{'0', 'b', '1', '0', 'a'}))
	require.Equal(t, uint(3), number([]int32{'0', 'x', '1', 'p'}))
	require.Equal(t, uint(2), number([]int32{'1', '2', 'a'}))
	require.Equal(t, uint(5), number([]int32{'0', 'x', 'A', '_', '0', '='}))
}
