package codon

var headerLogics = `
type RandSrc interface {
	GetBool() bool
	GetInt() int
	GetInt8() int8
	GetInt16() int16
	GetInt32() int32
	GetInt64() int64
	GetUint() uint
	GetUint8() uint8
	GetUint16() uint16
	GetUint32() uint32
	GetUint64() uint64
	GetFloat32() float32
	GetFloat64() float64
	GetString(n int) string
	GetBytes(n int) []byte
}

func codonEncodeBool(w *[]byte, v bool) {
	if v {
		*w = append(*w, byte(1))
	} else {
		*w = append(*w, byte(0))
	}
}
func codonEncodeVarint(w *[]byte, v int64) {
	var buf [10]byte
	n := binary.PutVarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonEncodeInt8(w *[]byte, v int8) {
	*w = append(*w, byte(v))
}
func codonEncodeInt16(w *[]byte, v int16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], uint16(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeUvarint(w *[]byte, v uint64) {
	var buf [10]byte
	n := binary.PutUvarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonEncodeUint8(w *[]byte, v uint8) {
	*w = append(*w, byte(v))
}
func codonEncodeUint16(w *[]byte, v uint16) {
	var buf [2]byte
	binary.LittleEndian.PutUint16(buf[:], v)
	*w = append(*w, buf[:]...)
}
func codonEncodeFloat32(w *[]byte, v float32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], math.Float32bits(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeFloat64(w *[]byte, v float64) {
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], math.Float64bits(v))
	*w = append(*w, buf[:]...)
}
func codonEncodeByteSlice(w *[]byte, v []byte) {
	codonEncodeVarint(w, int64(len(v)))
	*w = append(*w, v...)
}
func codonEncodeString(w *[]byte, v string) {
	codonEncodeByteSlice(w, []byte(v))
}
func codonDecodeBool(bz []byte, n *int, err *error) bool {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return false
	}
	*n = 1
	*err = nil
	return bz[0]!=0
}
func codonDecodeInt(bz []byte, m *int, err *error) int {
	i, n := binary.Varint(bz)
	if n == 0 {
		// buf too small
		*err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		*err = errors.New("EOF decoding varint")
	}
	*m = n
	return int(i)
}
func codonDecodeInt8(bz []byte, n *int, err *error) int8 {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*err = nil
	*n = 1
	return int8(bz[0])
}
func codonDecodeInt16(bz []byte, n *int, err *error) int16 {
	if len(bz) < 2 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 2
	*err = nil
	return int16(binary.LittleEndian.Uint16(bz[:2]))
}
func codonDecodeInt32(bz []byte, n *int, err *error) int32 {
	i := codonDecodeInt64(bz, n, err)
	return int32(i)
}
func codonDecodeInt64(bz []byte, m *int, err *error) int64 {
	i, n := binary.Varint(bz)
	if n == 0 {
		// buf too small
		*err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		*err = errors.New("EOF decoding varint")
	}
	*m = n
	*err = nil
	return int64(i)
}
func codonDecodeUint(bz []byte, n *int, err *error) uint {
	i := codonDecodeUint64(bz, n, err)
	return uint(i)
}
func codonDecodeUint8(bz []byte, n *int, err *error) uint8 {
	if len(bz) < 1 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 1
	*err = nil
	return uint8(bz[0])
}
func codonDecodeUint16(bz []byte, n *int, err *error) uint16 {
	if len(bz) < 2 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 2
	*err = nil
	return uint16(binary.LittleEndian.Uint16(bz[:2]))
}
func codonDecodeUint32(bz []byte, n *int, err *error) uint32 {
	i := codonDecodeUint64(bz, n, err)
	return uint32(i)
}
func codonDecodeUint64(bz []byte, m *int, err *error) uint64 {
	i, n := binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		*err = errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		*err = errors.New("EOF decoding varint")
	}
	*m = n
	*err = nil
	return uint64(i)
}
func codonDecodeFloat64(bz []byte, n *int, err *error) float64 {
	if len(bz) < 8 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 8
	*err = nil
	i := binary.LittleEndian.Uint64(bz[:8])
	return math.Float64frombits(i)
}
func codonDecodeFloat32(bz []byte, n *int, err *error) float32 {
	if len(bz) < 4 {
		*err = errors.New("Not enough bytes to read")
		return 0
	}
	*n = 4
	*err = nil
	i := binary.LittleEndian.Uint32(bz[:4])
	return math.Float32frombits(i)
}
func codonGetByteSlice(bz []byte, length int) ([]byte, int, error) {
	if len(bz) < length {
		return nil, 0, errors.New("Not enough bytes to read")
	}
	return bz[:length], length, nil
}
func codonDecodeString(bz []byte, n *int, err *error) string {
	var m int
	length := codonDecodeInt64(bz, &m, err)
	if *err != nil {
		return ""
	}
	var bs []byte
	var l int
	bs, l, *err = codonGetByteSlice(bz[m:], int(length))
	*n = m + l
	return string(bs)
}
`

