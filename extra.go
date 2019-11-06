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
	var buf [binary.MaxVarintLen64]byte
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
	var buf [binary.MaxVarintLen64]byte
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

var ImportsForBridgeLogic = []string{`"io"`, `"fmt"`, `"reflect"`, `"encoding/json"`, `amino "github.com/tendermint/go-amino"`}

var disfixWrapperDefine = "type disfixWrapper struct {\n"+
			"	Name string          `json:\"type\"`\n"+
			"	Data json.RawMessage `json:\"value\"`\n"+
			"}\n"

var BridgeLogic = disfixWrapperDefine+`
// ========= BridgeBegin ============
type CodecImp struct {
	sealed          bool
	structPath2Name map[string]string
}

func (cdc *CodecImp) MarshalBinaryBare(o interface{}) ([]byte, error) {
	s := CodonStub{}
	return s.MarshalBinaryBare(o)
}
func (cdc *CodecImp) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	s := CodonStub{}
	return s.MarshalBinaryLengthPrefixed(o)
}
func (cdc *CodecImp) MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error) {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	m, err := w.Write(bz)
	return int64(m), err
}
func (cdc *CodecImp) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	s := CodonStub{}
	return s.UnmarshalBinaryBare(bz, ptr)
}
func (cdc *CodecImp) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	s := CodonStub{}
	return s.UnmarshalBinaryLengthPrefixed(bz, ptr)
}
func (cdc *CodecImp) UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{}, maxSize int64) (n int64, err error) {
	if maxSize < 0 {
		panic("maxSize cannot be negative.")
	}

	// Read byte-length prefix.
	var l int64
	var buf [binary.MaxVarintLen64]byte
	for i := 0; i < len(buf); i++ {
		_, err = r.Read(buf[i : i+1])
		if err != nil {
			return
		}
		n += 1
		if buf[i]&0x80 == 0 {
			break
		}
		if n >= maxSize {
			err = fmt.Errorf("Read overflow, maxSize is %v but uvarint(length-prefix) is itself greater than maxSize.", maxSize)
		}
	}
	u64, _ := binary.Uvarint(buf[:])
	if err != nil {
		return
	}
	if maxSize > 0 {
		if uint64(maxSize) < u64 {
			err = fmt.Errorf("Read overflow, maxSize is %v but this amino binary object is %v bytes.", maxSize, u64)
			return
		}
		if (maxSize - n) < int64(u64) {
			err = fmt.Errorf("Read overflow, maxSize is %v but this length-prefixed amino binary object is %v+%v bytes.", maxSize, n, u64)
			return
		}
	}
	l = int64(u64)
	if l < 0 {
		err = fmt.Errorf("Read overflow, this implementation can't read this because, why would anyone have this much data?")
	}

	// Read that many bytes.
	var bz = make([]byte, l, l)
	_, err = io.ReadFull(r, bz)
	if err != nil {
		return
	}
	n += l

	// Decode.
	err = cdc.UnmarshalBinaryBare(bz, ptr)
	return
}

//------

func (cdc *CodecImp) MustMarshalBinaryBare(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryBare(o)
	if err!=nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := cdc.MarshalBinaryLengthPrefixed(o)
	if err!=nil {
		panic(err)
	}
	return bz
}
func (cdc *CodecImp) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryBare(bz, ptr)
	if err!=nil {
		panic(err)
	}
}
func (cdc *CodecImp) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	err := cdc.UnmarshalBinaryLengthPrefixed(bz, ptr)
	if err!=nil {
		panic(err)
	}
}

// ====================
func derefPtr(v interface{}) reflect.Type {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func (cdc *CodecImp) PrintTypes(out io.Writer) error {
	for _, entry := range GetSupportList() {
		_, err := out.Write([]byte(entry))
		if err != nil {
			return err
		}
		_, err = out.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
func (cdc *CodecImp) RegisterConcrete(o interface{}, name string, copts *amino.ConcreteOptions) {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	t := derefPtr(o)
	path := t.PkgPath() + "." + t.Name()
	cdc.structPath2Name[path] = name
}
func (cdc *CodecImp) RegisterInterface(_ interface{}, _ *amino.InterfaceOptions) {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	//Nop
}
func (cdc *CodecImp) SealImp() *CodecImp {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	cdc.sealed = true
	return cdc
}

// ========================================

type CodonStub struct {
}

func (_ *CodonStub) NewCodecImp() *CodecImp {
	return &CodecImp{
		structPath2Name: make(map[string]string),
	}
}
func (_ *CodonStub) DeepCopy(o interface{}) (r interface{}) {
	r = DeepCopyAny(o)
	return
}

func (_ *CodonStub) MarshalBinaryBare(o interface{}) ([]byte, error) {
	if _, err := getMagicBytesOfVar(o); err!=nil {
		return nil, err
	}
	buf := make([]byte, 0, 1024)
	EncodeAny(&buf, o)
	return buf, nil
}
func (s *CodonStub) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	if _, err := getMagicBytesOfVar(o); err!=nil {
		return nil, err
	}
	bz, err := s.MarshalBinaryBare(o)
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	return append(buf[:n], bz...), err
}
func (_ *CodonStub) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	if _, err := getMagicBytesOfVar(ptr); err!=nil {
		return err
	}
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}
	if len(bz) <= 4 {
		return fmt.Errorf("Byte slice is too short: %d", len(bz))
	}
	magicBytes, err := getMagicBytesOfVar(ptr)
	if err!=nil {
		return err
	}
	if bz[0]!=magicBytes[0] || bz[1]!=magicBytes[1] || bz[2]!=magicBytes[2] || bz[3]!=magicBytes[3] {
		return fmt.Errorf("MagicBytes Missmatch %v vs %v", bz[0:4], magicBytes[:])
	}
	o, _, err := DecodeAny(bz)
	rv.Elem().Set(reflect.ValueOf(o))
	return err
}
func (s *CodonStub) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	if _, err := getMagicBytesOfVar(ptr); err!=nil {
		return err
	}
	if len(bz) == 0 {
		return errors.New("UnmarshalBinaryLengthPrefixed cannot decode empty bytes")
	}
	// Read byte-length prefix.
	u64, n := binary.Uvarint(bz)
	if n < 0 {
		return fmt.Errorf("Error reading msg byte-length prefix: got code %v", n)
	}
	if u64 > uint64(len(bz)-n) {
		return fmt.Errorf("Not enough bytes to read in UnmarshalBinaryLengthPrefixed, want %v more bytes but only have %v",
			u64, len(bz)-n)
	} else if u64 < uint64(len(bz)-n) {
		return fmt.Errorf("Bytes left over in UnmarshalBinaryLengthPrefixed, should read %v more bytes but have %v",
			u64, len(bz)-n)
	}
	bz = bz[n:]
	return s.UnmarshalBinaryBare(bz, ptr)
}
func (s *CodonStub) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	bz, err := s.MarshalBinaryLengthPrefixed(o)
	if err!=nil {
		panic(err)
	}
	return bz
}

// ========================================
func (_ *CodonStub) UvarintSize(u uint64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], u)
	return n
}
func (_ *CodonStub) EncodeByteSlice(w io.Writer, bz []byte) error {
	buf := make([]byte, 0, binary.MaxVarintLen64+len(bz))
	codonEncodeByteSlice(&buf, bz)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUvarint(w io.Writer, u uint64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (s *CodonStub) ByteSliceSize(bz []byte) int {
	return s.UvarintSize(uint64(len(bz))) + len(bz)
}
func (_ *CodonStub) EncodeInt8(w io.Writer, i int8) error {
	_, err := w.Write([]byte{byte(i)})
	return err
}
func (_ *CodonStub) EncodeInt16(w io.Writer, i int16) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeInt16(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeInt32(w io.Writer, i int32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, int64(i))
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeInt64(w io.Writer, i int64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeVarint(w io.Writer, i int64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeVarint(&buf, i)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeByte(w io.Writer, b byte) error {
	_, err := w.Write([]byte{b})
	return err
}
func (_ *CodonStub) EncodeUint8(w io.Writer, u uint8) error {
	_, err := w.Write([]byte{u})
	return err
}
func (_ *CodonStub) EncodeUint16(w io.Writer, u uint16) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUint16(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUint32(w io.Writer, u uint32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, uint64(u))
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeUint64(w io.Writer, u uint64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeUvarint(&buf, u)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeBool(w io.Writer, b bool) error {
	u := byte(0)
	if b {
		u = byte(1)
	}
	_, err := w.Write([]byte{u})
	return err
}
func (_ *CodonStub) EncodeFloat32(w io.Writer, f float32) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeFloat32(&buf, f)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeFloat64(w io.Writer, f float64) error {
	buf := make([]byte, 0, binary.MaxVarintLen64)
	codonEncodeFloat64(&buf, f)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) EncodeString(w io.Writer, s string) error {
	buf := make([]byte, 0, binary.MaxVarintLen64+len(s))
	codonEncodeString(&buf, s)
	_, err := w.Write(buf)
	return err
}
func (_ *CodonStub) DecodeInt8(bz []byte) (i int8, n int, err error) {
	i = codonDecodeInt8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt16(bz []byte) (i int16, n int, err error) {
	i = codonDecodeInt16(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt32(bz []byte) (i int32, n int, err error) {
	i = codonDecodeInt32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeInt64(bz []byte) (i int64, n int, err error) {
	i = codonDecodeInt64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeVarint(bz []byte) (i int64, n int, err error) {
	i = codonDecodeInt64(bz, &n, &err)
	return
}
func (s *CodonStub) DecodeByte(bz []byte) (b byte, n int, err error) {
	b = codonDecodeUint8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint8(bz []byte) (u uint8, n int, err error) {
	u = codonDecodeUint8(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint16(bz []byte) (u uint16, n int, err error) {
	u = codonDecodeUint16(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint32(bz []byte) (u uint32, n int, err error) {
	u = codonDecodeUint32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUint64(bz []byte) (u uint64, n int, err error) {
	u = codonDecodeUint64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeUvarint(bz []byte) (u uint64, n int, err error) {
	u = codonDecodeUint64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeBool(bz []byte) (b bool, n int, err error) {
	b = codonDecodeBool(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeFloat32(bz []byte) (f float32, n int, err error) {
	f = codonDecodeFloat32(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeFloat64(bz []byte) (f float64, n int, err error) {
	f = codonDecodeFloat64(bz, &n, &err)
	return
}
func (_ *CodonStub) DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	length := codonDecodeInt(bz, &n, &err)
	if err != nil {
		return
	}
	bz = bz[n:]
	n += length
	bz2, m, err := codonGetByteSlice(bz, length)
	n += m
	return
}
func (_ *CodonStub) DecodeString(bz []byte) (s string, n int, err error) {
	s = codonDecodeString(bz, &n, &err)
	return
}
func (_ *CodonStub) VarintSize(i int64) int {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)
	return n
}
// ========= BridgeEnd ============
`

