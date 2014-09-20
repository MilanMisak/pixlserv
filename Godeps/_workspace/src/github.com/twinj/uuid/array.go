package uuid

/****************
 * Date: 1/02/14
 * Time: 10:08 AM
 ***************/

const (
	variantIndex = 8
	versionIndex = 6
)

// A clean UUID type for simpler UUID versions
type UUIDArray [length]byte

func (UUIDArray) Size() int {
	return length
}

func (o UUIDArray) Version() int {
	return int(o[versionIndex]) >> 4
}

func (o *UUIDArray) setVersion(pVersion int) {
	o[versionIndex] &= 0x0F
	o[versionIndex] |= byte(pVersion)<<4
}

func (o *UUIDArray) Variant() byte {
	return variant(o[variantIndex])
}

func (o *UUIDArray) setVariant(pVariant byte) {
	setVariant(&o[variantIndex], pVariant)
}

func (o *UUIDArray) Unmarshal(pData []byte) {
	copy(o[:], pData)
}

func (o *UUIDArray) Bytes() []byte {
	return o[:]
}

func (o UUIDArray) String() string {
	return formatter(&o, format)
}

func (o UUIDArray) Format(pFormat Format) string {
	return formatter(&o, pFormat)
}

// Set the three most significant bits (bits 0, 1 and 2) of the
// sequenceHiAndVariant equivalent in the array to ReservedRFC4122.
func (o *UUIDArray) setRFC4122Variant() {
	o[variantIndex] &= 0x3F
	o[variantIndex] |= ReservedRFC4122
}

// Marshals the UUID bytes into a slice
func (o *UUIDArray) MarshalBinary() ([]byte, error) {
	return o.Bytes(), nil
}

// Un-marshals the data bytes into the UUID.
func (o *UUIDArray) UnmarshalBinary(pData []byte) (error) {
	return UnmarshalBinary(o, pData)
}

