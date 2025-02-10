package bit

import (
	"fmt"
	"strings"
)

// Set provides a straightforward bitset implementation.  That is, a set of
// (unsigned) integer values implemented as an array of bits.
type Set struct {
	words []uint64
}

// Insert a given value into this set.
func (p *Set) Insert(val uint) {
	word := val / 64
	bit := val % 64
	//
	for uint(len(p.words)) <= word {
		p.words = append(p.words, 0)
	}
	// Set bit
	mask := uint64(1) << bit
	p.words[word] = p.words[word] | mask
}

// InsertAll inserts all elements from a given bitset into this bitset.
func (p *Set) InsertAll(bits Set) {
	for len(p.words) < len(bits.words) {
		p.words = append(p.words, 0)
	}
	// Insert all
	for w := 0; w < len(bits.words); w++ {
		p.words[w] = p.words[w] | bits.words[w]
	}
}

// Contains checks whether a given value is contained, or not.
func (p *Set) Contains(val uint) bool {
	word := val / 64
	bit := val % 64
	//
	if uint(len(p.words)) <= word {
		return false
	}
	// Set mask
	mask := uint64(1) << bit
	//
	return (p.words[word] & mask) != 0
}

// Count returns the number of bits in the bitset which are set to one.
func (p *Set) Count() uint {
	count := uint(0)
	//
	for word := uint(0); word < uint(len(p.words)); word++ {
		bits := p.words[word]
		//
		for bits != 0 {
			if bits&1 == 1 {
				count++
			}
			//
			bits = bits >> 1
		}
	}
	//
	return count
}

func (p *Set) String() string {
	var (
		builder strings.Builder
		first   = true
	)
	//
	for word := uint(0); word < uint(len(p.words)); word++ {
		for bit := uint(0); bit < 64; bit++ {
			value := (word * 64) + bit

			if p.Contains(value) {
				if !first {
					builder.WriteString(", ")
				}
				//
				first = false
				//
				builder.WriteString(fmt.Sprintf("%d", value))
			}
		}
	}
	//
	return builder.String()
}
