package util

import (
	"fmt"
	"slices"
)

// Path is a construct for describing paths through trees.  A path can be either
// *absolute* or *relative*.  An absolute path always starts from the root of
// the tree, whilst a relative path can begin from any point within the tree.
type Path struct {
	// Indicates whether or not this is an absolute path.
	absolute bool
	// Segments in the path.
	segments []string
}

// NewAbsolutePath constructs a new absolute path from the given segments.
func NewAbsolutePath(segments ...string) Path {
	return Path{true, segments}
}

// NewRelativePath constructs a new absolute path from the given segments.
func NewRelativePath(segments ...string) Path {
	return Path{false, segments}
}

// Depth returns the number of segments in this path (a.k.a its depth).
func (p *Path) Depth() uint {
	return uint(len(p.segments))
}

// IsAbsolute determines whether or not this is an absolute path.
func (p *Path) IsAbsolute() bool {
	return p.absolute
}

// Head returns the first (i.e. outermost) segment in this path.
func (p *Path) Head() string {
	return p.segments[0]
}

// Dehead removes the head from this path, returning an otherwise identical
// path.  Observe that, if this were absolute, it is no longer!
func (p *Path) Dehead() *Path {
	return &Path{false, p.segments[1:]}
}

// Tail returns the last (i.e. innermost) segment in this path.
func (p *Path) Tail() string {
	n := len(p.segments) - 1
	return p.segments[n]
}

// Get returns the nth segment of this path.
func (p *Path) Get(nth uint) string {
	return p.segments[nth]
}

// Equals determines whether two paths are the same.
func (p *Path) Equals(other Path) bool {
	return p.absolute == other.absolute && slices.Equal(p.segments, other.segments)
}

// PrefixOf checks whether this path is a prefix of the other.
func (p *Path) PrefixOf(other Path) bool {
	if len(p.segments) > len(other.segments) {
		return false
	}
	//
	for i := range p.segments {
		if p.segments[i] != other.segments[i] {
			return false
		}
	}
	// Looks good
	return true
}

// Slice returns the subpath starting from the given segment.
func (p *Path) Slice(start uint) *Path {
	return &Path{false, p.segments[start:]}
}

// PushRoot converts a relative path into an absolute path by pushing the "root"
// of the tree onto the head (i.e. outermost) position.
func (p *Path) PushRoot(tail string) *Path {
	if p.absolute {
		panic("cannot push root onto absolute path")
	}
	// Prepend root to segments
	nsegments := Prepend(tail, p.segments)
	// Convert to absolute path
	return &Path{true, nsegments}
}

// Parent returns the parent of this path.
func (p *Path) Parent() *Path {
	n := p.Depth() - 1
	return &Path{p.absolute, p.segments[0:n]}
}

// Extend returns this path extended with a new innermost segment.
func (p *Path) Extend(tail string) *Path {
	return &Path{p.absolute, Append(p.segments, tail)}
}

// Return a string representation of this path.
func (p *Path) String() string {
	if p.IsAbsolute() {
		switch len(p.segments) {
		case 0:
			return ""
		case 1:
			return p.segments[0]
		case 2:
			return fmt.Sprintf("%s.%s", p.segments[0], p.segments[1])
		default:
			return fmt.Sprintf("%s/%s", p.Parent().String(), p.Tail())
		}
	}
	//
	switch len(p.segments) {
	case 0:
		// Non-sensical case really
		return "/"
	case 1:
		return fmt.Sprintf("/%s", p.segments[0])
	default:
		return fmt.Sprintf("%s/%s", p.Parent().String(), p.Tail())
	}
}
