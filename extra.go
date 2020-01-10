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

func codonWriteVarint(w *[]byte, v int64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}
func codonWriteUvarint(w *[]byte, v uint64) {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	*w = append(*w, buf[0:n]...)
}

func codonEncodeBool(n int, w *[]byte, v bool) {
	codonWriteUvarint(w, uint64(n)<<3)
	if v {
		codonWriteUvarint(w, uint64(1))
	} else {
		codonWriteUvarint(w, uint64(0))
	}
}
func codonEncodeVarint(n int, w *[]byte, v int64) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeInt8(n int, w *[]byte, v int8) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeInt16(n int, w *[]byte, v int16) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteVarint(w, int64(v))
}
func codonEncodeUvarint(n int, w *[]byte, v uint64) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, v)
}
func codonEncodeUint8(n int, w *[]byte, v uint8) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, uint64(v))
}
func codonEncodeUint16(n int, w *[]byte, v uint16) {
	codonWriteUvarint(w, uint64(n)<<3)
	codonWriteUvarint(w, uint64(v))
}

func codonEncodeByteSlice(n int, w *[]byte, v []byte) {
	codonWriteUvarint(w, (uint64(n)<<3)|2)
	codonWriteUvarint(w, uint64(len(v)))
	*w = append(*w, v...)
}
func codonEncodeString(n int, w *[]byte, v string) {
	codonEncodeByteSlice(n, w, []byte(v))
}
func codonDecodeBool(bz []byte, n *int, err *error) bool {
	return codonDecodeInt64(bz, n, err) != 0
}
func codonDecodeInt(bz []byte, n *int, err *error) int {
	return int(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt8(bz []byte, n *int, err *error) int8 {
	return int8(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt16(bz []byte, n *int, err *error) int16 {
	return int16(codonDecodeInt64(bz, n, err))
}
func codonDecodeInt32(bz []byte, n *int, err *error) int32 {
	return int32(codonDecodeInt64(bz, n, err))
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
	return uint(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint8(bz []byte, n *int, err *error) uint8 {
	return uint8(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint16(bz []byte, n *int, err *error) uint16 {
	return uint16(codonDecodeUint64(bz, n, err))
}
func codonDecodeUint32(bz []byte, n *int, err *error) uint32 {
	return uint32(codonDecodeUint64(bz, n, err))
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
func codonGetByteSlice(res *[]byte, bz []byte) (int, error) {
	length, n := binary.Uvarint(bz)
	if n == 0 {
		// buf too small
		return n, errors.New("buffer too small")
	} else if n < 0 {
		// value larger than 64 bits (overflow)
		// and -n is the number of bytes read
		n = -n
		return n, errors.New("EOF decoding varint")
	}
	if length == 0 {
		*res = nil
		return 0, nil
	}
	bz = bz[n:]
	if len(bz) < int(length) {
		*res = nil
		return 0, errors.New("Not enough bytes to read")
	}
	if *res == nil {
		*res = append(*res, bz[:length]...)
	} else {
		*res = append((*res)[:0], bz[:length]...)
	}
	return n+int(length), nil
}
func codonDecodeString(bz []byte, n *int, err *error) string {
	var res []byte
	*n, *err = codonGetByteSlice(&res, bz)
	return string(res)
}

`

var ImportsForBridgeLogic = []string{`"io"`, `"fmt"`, `"reflect"`, `amino "github.com/coinexchain/codon/wrap-amino"`}

var BridgeLogic = `
// ========= BridgeBegin ============
type CodecImp struct {
	sealed          bool
}

var _ amino.Sealer = &CodecImp{}
var _ amino.CodecIfc = &CodecImp{}

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
	found := false
	for _, entry := range GetSupportList() {
		if path == entry {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("%s is not supported", path))
	}
}
func (cdc *CodecImp) RegisterInterface(o interface{}, _ *amino.InterfaceOptions) {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	t := derefPtr(o)
	path := t.PkgPath() + "." + t.Name()
	found := false
	for _, entry := range GetSupportList() {
		if path == entry {
			found = true
			break
		}
	}
	if !found {
		panic(fmt.Sprintf("%s is not supported", path))
	}
}
func (cdc *CodecImp) SealImp() {
	if cdc.sealed {
		panic("Codec is already sealed")
	}
	cdc.sealed = true
}

// ========================================

type CodonStub struct {
}

func (_ *CodonStub) NewCodecImp() amino.CodecIfc {
	return &CodecImp{}
}
func (_ *CodonStub) DeepCopy(o interface{}) (r interface{}) {
	r = DeepCopyAny(o)
	return
}

func (_ *CodonStub) MarshalBinaryBare(o interface{}) ([]byte, error) {
	if _, ok := getMagicNumOfVar(o); !ok {
		return nil, errors.New("Not Supported Type")
	}
	buf := make([]byte, 0, 1024)
	EncodeAny(&buf, o)
	return buf, nil
}
func (s *CodonStub) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	if _, ok := getMagicNumOfVar(o); !ok {
		return nil, errors.New("Not Supported Type")
	}
	bz, err := s.MarshalBinaryBare(o)
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], uint64(len(bz)))
	return append(buf[:n], bz...), err
}
func (_ *CodonStub) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	rv := reflect.ValueOf(ptr)
	if rv.Kind() != reflect.Ptr {
		panic("Unmarshal expects a pointer")
	}

	if len(bz) <= 4 {
		return fmt.Errorf("Byte slice is too short: %d", len(bz))
	}
	o, _, err := DecodeAny(bz)
	if rv.Elem().Kind() == reflect.Interface {
		AssignIfcPtrFromStruct(ptr, o)
	} else {
		rv.Elem().Set(reflect.ValueOf(o))
	}
	return err
}
func (s *CodonStub) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
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
	_, err := w.Write(ByteSliceWithLengthPrefix(bz))
	return err
}
func (s *CodonStub) ByteSliceSize(bz []byte) int {
	return s.UvarintSize(uint64(len(bz))) + len(bz)
}
func (_ *CodonStub) EncodeVarint(w io.Writer, i int64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutVarint(buf[:], i)
	_, err := w.Write(buf[:n])
	return err
}
func (s *CodonStub) EncodeInt8(w io.Writer, i int8) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt16(w io.Writer, i int16) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt32(w io.Writer, i int32) error {
	return s.EncodeVarint(w, int64(i))
}
func (s *CodonStub) EncodeInt64(w io.Writer, i int64) error {
	return s.EncodeVarint(w, int64(i))
}
func (_ *CodonStub) EncodeUvarint(w io.Writer, u uint64) error {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], u)
	_, err := w.Write(buf[:n])
	return err
}
func (s *CodonStub) EncodeByte(w io.Writer, b byte) error {
	return s.EncodeUvarint(w, uint64(b))
}
func (s *CodonStub) EncodeUint8(w io.Writer, u uint8) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint16(w io.Writer, u uint16) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint32(w io.Writer, u uint32) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (s *CodonStub) EncodeUint64(w io.Writer, u uint64) error {
	return s.EncodeUvarint(w, uint64(u))
}
func (_ *CodonStub) EncodeBool(w io.Writer, b bool) error {
	u := byte(0)
	if b {
		u = byte(1)
	}
	_, err := w.Write([]byte{u})
	return err
}
func (s *CodonStub) EncodeString(w io.Writer, str string) error {
	return s.EncodeByteSlice(w, []byte(str))
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
func (_ *CodonStub) DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	m, err := codonGetByteSlice(&bz2, bz)
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

