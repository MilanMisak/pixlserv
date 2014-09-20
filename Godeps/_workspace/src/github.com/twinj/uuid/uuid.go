// This package provides RFC4122 UUID capabilities.
// Also included in the package is a distinctly go way of creating a
// unique id.
//
// NewV1, NewV3, NewV4, NewV5, for generating versions 1, 3, 4
// and 5 UUIDs as specified in RFC 4122.
//
// New([]byte), unsafe NewHex(string) and safe ParseUUID(string) for
// creating UUIDs from existing data
//
// The original version was from Krzysztof Kowalik <chris@nu7hat.ch>
// Although his version was non compliant with RFC4122 and bugged.
// I forked it but have since heavily redesigned it to suit my purposes
// The example code in the specification was also used to help design this
// as such I have started a new repository.
//  Copyright (C) 2014 twinj@github.com  2014 MIT style licence
package uuid

/****************
 * Date: 31/01/14
 * Time: 3:35 PM
 ***************/

import (
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"regexp"
	seed "math/rand"
	"strings"
)

const (
	ReservedNCS byte        = 0x00
	ReservedRFC4122   byte  = 0x80 // or and A0 if masked with 1F
	ReservedMicrosoft  byte = 0xC0
	ReservedFuture byte     = 0xE0
	TakeBack byte           = 0xF0
)

const (

	// Pattern used to parse string representation of the UUID.
	// Current one allows to parse string where only one opening
	// or closing bracket or any of the hyphens are optional.
	// It is only used to extract the main bytes to create a UUID,
	// so these imperfections are of no consequence.
	hexPattern = `^(urn\:uuid\:)?[\{(\[]?([A-Fa-f0-9]{8})-?([A-Fa-f0-9]{4})-?([1-5][A-Fa-f0-9]{3})-?([A-Fa-f0-9]{4})-?([A-Fa-f0-9]{12})[\]\})]?$`
)

var (
	parseUUIDRegex = regexp.MustCompile(hexPattern)
	format Format
)

func init() {
	SwitchFormatUpper(CurlyHyphen)
}

// ******************************************************  UUID

// The main interface for UUIDs
// Each implementation must also implement the UniqueName interface
type UUID interface {

	// Marshals the UUID bytes or data
	Bytes() (data []byte)

	// Organises data into a new UUID
	Unmarshal(pData []byte)

	// Size is used where different implementations require
	// different sizes. Should return the number of bytes in
	// the implementation.
	// Enables unmarshal and Bytes to screen for size
	Size() int

	// Version returns a version number of the algorithm used
	// to generate the UUID.
	// This may may behave independently across non RFC4122 UUIDs
	Version() int

	// Variant returns the UUID Variant
	// This will be one of the constants:
	// ReservedRFC4122,
	// ReservedMicrosoft,
	// ReservedFuture,
	// ReservedNCS.
	// This may behave differently across non RFC4122 UUIDs
	Variant() byte

	// UUID can be used as a Name within a namespace
	// Is simply just a String() string method
	// Returns a formatted version of the UUID.
	String() string
}

// New creates a UUID object from a data byte slice.
// It will truncate any bytes past the default length of 16
// It will panic if data slice is too small
func New(pData []byte) UUID {
	o := new(UUIDArray)
	o.Unmarshal(pData[:length])
	return o
}

// GoId creates a UUID object based on timestamps and a hash.
// It will truncate any bytes past the length of the initial hash.
// This creates a UUID based on a Namespace, UniqueName and an existing
// hash.
func GoId(pNs UUID, pName UniqueName, pHash hash.Hash) UUID {
	o := new(UUIDStruct)
	o.size = pHash.Size()
	Digest(o, pNs, pName, pHash)
	now := currentUUIDTimestamp()
	sequence := uint16(seed.Int()) & 0x3FFF
	return formatGoId(o, now, 15, ReservedFuture, sequence)
}

// Unmarshals data into struct for GoId UUIDs
func formatGoId(o *UUIDStruct, pNow Timestamp, pVersion uint16, pVariant byte, pSequence uint16) UUID {
	o.timeLow = uint32(pNow & 0xFFFFFFFF)
	o.timeMid = uint16((pNow >> 32) & 0xFFFF)
	o.timeHiAndVersion = uint16((pNow >> 48) & 0x0FFF)
	o.timeHiAndVersion |= uint16(pVersion << 12)
	o.sequenceLow = byte(pSequence & 0xFF)
	o.sequenceHiAndVariant = byte(( pSequence & 0x3F00) >> 8)
	o.sequenceHiAndVariant |= pVariant
	return o
}

// Creates a UUID from a valid hex string
// Will panic if hex string is invalid - will panic even with hyphens and brackets
func NewHex(pUuid string) UUID {
	bytes, err := hex.DecodeString(pUuid)
	if err != nil {
		panic(err)
	}
	return New(bytes)
}

// ParseUUID creates a UUID object from a valid string representation.
// Accepts UUID string in following formats:
//		6ba7b8149dad11d180b400c04fd430c8
//		6ba7b814-9dad-11d1-80b4-00c04fd430c8
//		{6ba7b814-9dad-11d1-80b4-00c04fd430c8}
//		urn:uuid:6ba7b814-9dad-11d1-80b4-00c04fd430c8
//		[6ba7b814-9dad-11d1-80b4-00c04fd430c8]
//
func ParseUUID(pUUID string) (UUID, error) {
	md := parseUUIDRegex.FindStringSubmatch(pUUID)
	if md == nil {
		return nil, errors.New("UUID.ParseUUID: invalid string")
	}
	return NewHex(md[2]+md[3]+md[4]+md[5]+md[6]), nil
}

// Digest a namespace UUID and a UniqueName, which then  marshals to
// a new UUID
func Digest(o, pNs UUID, pName UniqueName, pHash hash.Hash) {
	// Hash writer never returns an error
	pHash.Write(pNs.Bytes())
	pHash.Write([]byte(pName.String()))
	o.Unmarshal(pHash.Sum(nil)[:o.Size()])
}

// Checks for length
func UnmarshalBinary(o UUID , pData []byte) error {
	if (len(pData) != o.Size()) {
		return errors.New("UUID.UnmarshalBinary: invalid length")
	}
	o.Unmarshal(pData)
	return nil
}

// **********************************************  UUID Names

// Name is a simple string which implements UniqueName
type Name string

func (o Name) String() string {
	return string(o)
}

// NewName will create a name from several sources
func NewName(salt string, pNames... UniqueName) UniqueName {
	var s string
	for _, s2 := range pNames {
		s += s2.String()
	}
	return Name(s + salt)
}

// UniqueName is a Stinger interface
// Made for easy passing of IPs, URLs, the several Address types,
// Buffers and any other type which implements Stringer
// string, []byte types and Hash sums will need to be cast to
// the Name type or some other type which implements
// Stringer or UniqueName
type UniqueName interface {

	// Many go types implement this method for use with printing
	// Will convert the current type to its native string format
	String() string
}

// **********************************************  UUID Printing

type Format string

const (
	Clean Format         = "%x%x%x%x%x%x"
	Curly Format         = "{%x%x%x%x%x%x}"
	Bracket Format       = "(%x%x%x%x%x%x)"
	CleanHyphen Format   = "%x-%x-%x-%x%x-%x"
	CurlyHyphen Format   = "{%x-%x-%x-%x%x-%x}"
	BracketHyphen Format = "(%x-%x-%x-%x%x-%x)"
	GoIdFormat Format    = "[%X-%X-%x-%X%X-%x]"
)

// Switches the printing format for UUID strings
// When String() is called it will get the current format
// Default is CurlyHyphen
// A valid format will have 6 groups
func SwitchFormat(pFormat Format) {
	if (strings.Count(string(pFormat), "%") != 6) {
		panic(errors.New("UUID.SwitchFormat: invalid formatting"))
	}
	format = pFormat
}

// Same as SwitchFormat but will make it uppercase: will ruin GoId
// formatting
func SwitchFormatUpper(pFormat Format) {
	SwitchFormat(Format(strings.ToUpper(string(pFormat))))
}

// Gets the current default format string
func GetFormat() Format {
	return format
}

// Compares whether each UUID is the same
func Equal(p1 UUID, p2 UUID) bool {
	return p1.String() == p2.String()
}

// **********************************************  UUID Versions

type UUIDVersion int

const (
	NONE    UUIDVersion = iota
	RFC4122v1
	DunnoYetv2
	RFC4122v3
	RFC4122v4
	RFC4122v5
)

// ***************************************************  Helpers

// Retrieves the variant from the given byte
func variant(pVariant byte) byte {
	switch pVariant & variantGet {
	case ReservedRFC4122, 0xA0:
		return ReservedRFC4122
	case ReservedMicrosoft:
		return ReservedMicrosoft
	case ReservedFuture:
		return ReservedFuture
	}
	return ReservedNCS
}

// not strictly required
func setVariant(pByte *byte, pVariant byte) {
	switch pVariant {
	case ReservedRFC4122:
		*pByte  &= variantSet
	case ReservedFuture, ReservedMicrosoft:
		*pByte  &= 0x1F
	case ReservedNCS:
		*pByte  &= 0x7F
	default:
		panic(errors.New("UUID.setVariant: invalid variant mask"))
	}
	*pByte |= pVariant;
}

// format a UUID into a human readable string
func formatter(pUUID UUID, pFormat Format) string {
	b := pUUID.Bytes()
	return fmt.Sprintf(string(pFormat), b[0:4], b[4:6], b[6:8], b[8:9], b[9:10], b[10:pUUID.Size()])
}

// Format a UUID into a human readable string
func Formatter(pUUID UUID, pFormat Format) string {
	if (strings.Count(string(pFormat), "%") != 6) {
		panic(errors.New("UUID.SwitchFormat: invalid formatting"))
	}
	return formatter(pUUID, pFormat)
}










