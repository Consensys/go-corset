package util

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strings"
)

// A reasonably simple hashset implementation which permits collisions.  Observe
// that, for example, hashicorp's go-set is *not* a suitable replacement here,
// since that does not handle collisions.  Specifically, it assumes the hash
// function always uniquely identifies the data in question.  I don't want to
// make that assumption here.

// Hasher provides a generic definition of a hashing function suitable for use
// within the hashset.  This is similar to the Hasher interface provided in
// go-set, except that it additionally includes equality.
type Hasher[T any] interface {
	// Check whether two items are equal (or not).
	Equals(T) bool
	// Return a suitable hashcode.
	Hash() uint64
}

// HashSet defines a generic set implementation backed by a map.  This is a true
// hashtable in that collisions are handle gracefully using buckets, rather than
// simply discarding them.
type HashSet[T Hasher[T]] struct {
	// items maps hashcodes to *buckets* of items.
	items map[uint64]bucket[T]
}

// NewHashSet creates a new HashSet with a given underlying capacity.
func NewHashSet[T Hasher[T]](size uint) *HashSet[T] {
	items := make(map[uint64]bucket[T], size)
	return &HashSet[T]{items}
}

// Size returns the number of unique items stored in this HashSet.
func (p *HashSet[T]) Size() uint {
	count := uint(0)
	for _, b := range p.items {
		count += b.size()
	}

	return count
}

// MaxBucket returns the size of the largest bucket.
func (p *HashSet[T]) MaxBucket() uint {
	m := uint(0)
	for _, b := range p.items {
		m = max(m, b.size())
	}

	return m
}

// Insert a new item into this map, returning true if it was already contained
// and false otherwise.
func (p *HashSet[T]) Insert(item T) bool {
	var b1 bucket[T]
	// Compute item's hashcode
	hash := item.Hash()
	// Lookup existing bucket
	b1 = p.items[hash]
	// Insert new item
	r := b1.insert(item)
	// Update map
	p.items[hash] = b1
	// Done
	return r
}

// Contains checks whether the given item is contained within this map, or not.
func (p *HashSet[T]) Contains(item T) bool {
	hash := item.Hash()

	if bucket, ok := p.items[hash]; ok {
		return bucket.contains(item)
	}

	return false
}

func (p *HashSet[T]) String() string {
	var r strings.Builder
	//
	first := true
	// Write opening brace
	r.WriteString("{")
	// Iterate all buckets
	for _, b := range p.items {
		// Iterate all items in bucket
		for _, i := range b.items {
			if !first {
				r.WriteString(",")
			}

			first = false

			r.WriteString(fmt.Sprintf("%s", any(i)))
		}
	}
	// Write closing brace
	r.WriteString("}")
	// Done
	return r.String()
}

// ============================================================================
// Bucket
// ============================================================================

type bucket[T Hasher[T]] struct {
	items []T
}

// Get the number of items in this bucket.
//
//nolint:revive
func (b *bucket[T]) size() uint {
	return uint(len(b.items))
}

// Insert a new item into this bucket
//
//nolint:revive
func (b *bucket[T]) insert(item T) bool {
	if b.contains(item) {
		// Item already present, so nothing to do.
		return true
	}
	// Append item
	b.items = append(b.items, item)
	// Item not present
	return false
}

// Check whether this bucket contains a given item, or not.
//
//nolint:revive
func (b *bucket[T]) contains(item T) bool {
	for _, i := range b.items {
		if item.Equals(i) {
			return true
		}
	}

	return false
}

// ============================================================================
// Key Implementations
// ============================================================================

// BytesKey wraps a bytes array as something which can be safely placed into a
// HashSet.
type BytesKey struct {
	bytes []byte
}

// NewBytesKey constructs a new bytes key.
func NewBytesKey(bytes []byte) BytesKey {
	return BytesKey{bytes}
}

// Equals compares two BytesKeys to check whether they represent the same
// underlying byte array (or not).
func (p BytesKey) Equals(other BytesKey) bool {
	return bytes.Equal(p.bytes, other.bytes)
}

// Hash generat6es a 64-bit hashcode from the underlying bytes array.
func (p BytesKey) Hash() uint64 {
	hash := fnv.New64a()
	hash.Write(p.bytes)
	// Done
	return hash.Sum64()
}
