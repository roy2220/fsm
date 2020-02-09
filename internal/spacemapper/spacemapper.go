// Package spacemapper defines space mappers.
package spacemapper

// SpaceMapper represents a space mapper.
type SpaceMapper interface {
	// MapSpace maps space with the given size.
	MapSpace(spaceSize int) error

	// AccessSpace returns a byte slice as a space accessor
	// which may get *INVALIDATED* after calling MapSpace.
	AccessSpace() []byte
}
