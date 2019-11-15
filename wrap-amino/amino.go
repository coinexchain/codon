package amino

import (
	"io"
	aminoOrig "github.com/tendermint/go-amino"
)

const Typ3_ByteLength = aminoOrig.Typ3_ByteLength

type ConcreteOptions = aminoOrig.ConcreteOptions
type InterfaceOptions = aminoOrig.InterfaceOptions

type CodecIfc interface {
	MarshalBinaryBare(o interface{}) ([]byte, error)
	MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error)
	MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error)
	MustMarshalBinaryBare(o interface{}) []byte
	MustMarshalBinaryLengthPrefixed(o interface{}) []byte
	MustUnmarshalBinaryBare(bz []byte, ptr interface{})
	MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{})
	PrintTypes(out io.Writer) error
	RegisterConcrete(o interface{}, name string, copts *ConcreteOptions)
	RegisterInterface(ptr interface{}, iopts *InterfaceOptions)
	UnmarshalBinaryBare(bz []byte, ptr interface{}) error
	UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error
	UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{}, maxSize int64) (n int64, err error)
}

type Sealer interface {
	SealImp()
}

type Codec struct {
	onlyOrig bool
	imp CodecIfc
	cdc *aminoOrig.Codec
}

type StubIfc interface {
	NewCodecImp() CodecIfc
	DeepCopy(o interface{}) (r interface{})
	MarshalBinaryBare(o interface{}) ([]byte, error)
	MustMarshalBinaryLengthPrefixed(o interface{}) []byte
	UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error
	UvarintSize(u uint64) int
	EncodeByteSlice(w io.Writer, bz []byte) (err error)
	EncodeUvarint(w io.Writer, u uint64) (err error)
	ByteSliceSize(bz []byte) int
	EncodeInt8(w io.Writer, i int8) (err error)
	EncodeInt16(w io.Writer, i int16) (err error)
	EncodeInt32(w io.Writer, i int32) (err error)
	EncodeInt64(w io.Writer, i int64) (err error)
	EncodeVarint(w io.Writer, i int64) (err error)
	EncodeByte(w io.Writer, b byte) (err error)
	EncodeUint8(w io.Writer, u uint8) (err error)
	EncodeUint16(w io.Writer, u uint16) (err error)
	EncodeUint32(w io.Writer, u uint32) (err error)
	EncodeUint64(w io.Writer, u uint64) (err error)
	EncodeBool(w io.Writer, b bool) (err error)
	EncodeFloat32(w io.Writer, f float32) (err error)
	EncodeFloat64(w io.Writer, f float64) (err error)
	EncodeString(w io.Writer, s string) (err error)
	DecodeInt8(bz []byte) (i int8, n int, err error)
	DecodeInt16(bz []byte) (i int16, n int, err error)
	DecodeInt32(bz []byte) (i int32, n int, err error)
	DecodeInt64(bz []byte) (i int64, n int, err error)
	DecodeVarint(bz []byte) (i int64, n int, err error)
	DecodeByte(bz []byte) (b byte, n int, err error)
	DecodeUint8(bz []byte) (u uint8, n int, err error)
	DecodeUint16(bz []byte) (u uint16, n int, err error)
	DecodeUint32(bz []byte) (u uint32, n int, err error)
	DecodeUint64(bz []byte) (u uint64, n int, err error)
	DecodeUvarint(bz []byte) (u uint64, n int, err error)
	DecodeBool(bz []byte) (b bool, n int, err error)
	DecodeFloat32(bz []byte) (f float32, n int, err error)
	DecodeFloat64(bz []byte) (f float64, n int, err error)
	DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error)
	DecodeString(bz []byte) (s string, n int, err error)
	VarintSize(i int64) int
}

var Stub StubIfc

/////////////////////////////

func NewCodec() *Codec {
	if Stub == nil { // use the orignal amino
		cdc := aminoOrig.NewCodec()
		return &Codec{
			imp: cdc,
			cdc: cdc,
			onlyOrig: true,
		}
	}
	return &Codec{
		imp: Stub.NewCodecImp(),
		cdc: aminoOrig.NewCodec(),
		onlyOrig: false,
	}
}

func (cdc *Codec) MarshalBinaryBare(o interface{}) ([]byte, error) {
	return cdc.imp.MarshalBinaryBare(o)
}
func (cdc *Codec) MarshalBinaryLengthPrefixed(o interface{}) ([]byte, error) {
	return cdc.imp.MarshalBinaryLengthPrefixed(o)
}
func (cdc *Codec) MarshalBinaryLengthPrefixedWriter(w io.Writer, o interface{}) (n int64, err error) {
	return cdc.imp.MarshalBinaryLengthPrefixedWriter(w, o)
}
func (cdc *Codec) MarshalJSON(o interface{}) ([]byte, error) {
	return cdc.cdc.MarshalJSON(o)
}
func (cdc *Codec) MarshalJSONIndent(o interface{}, prefix, indent string) ([]byte, error) {
	return cdc.cdc.MarshalJSONIndent(o, prefix, indent)
}
func (cdc *Codec) MustMarshalBinaryBare(o interface{}) []byte {
	return cdc.imp.MustMarshalBinaryBare(o)
}
func (cdc *Codec) MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	return cdc.imp.MustMarshalBinaryLengthPrefixed(o)
}
func (cdc *Codec) MustMarshalJSON(o interface{}) []byte {
	return cdc.cdc.MustMarshalJSON(o)
}
func (cdc *Codec) MustUnmarshalBinaryBare(bz []byte, ptr interface{}) {
	cdc.imp.MustUnmarshalBinaryBare(bz, ptr)
}
func (cdc *Codec) MustUnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) {
	cdc.imp.MustUnmarshalBinaryLengthPrefixed(bz, ptr)
}
func (cdc *Codec) MustUnmarshalJSON(bz []byte, ptr interface{}) {
	cdc.cdc.MustUnmarshalJSON(bz, ptr)
}
func (cdc *Codec) PrintTypes(out io.Writer) error {
	return cdc.imp.PrintTypes(out)
}
func (cdc *Codec) RegisterConcrete(o interface{}, name string, copts *ConcreteOptions) {
	if cdc.onlyOrig {
		cdc.cdc.RegisterConcrete(o, name, copts)
	} else {
		cdc.imp.RegisterConcrete(o, name, copts)
		cdc.cdc.RegisterConcrete(o, name, copts)
	}
}
func (cdc *Codec) RegisterInterface(ptr interface{}, iopts *InterfaceOptions) {
	if cdc.onlyOrig {
		cdc.cdc.RegisterInterface(ptr, iopts)
	} else {
		cdc.imp.RegisterInterface(ptr, iopts)
		cdc.cdc.RegisterInterface(ptr, iopts)
	}
}
func (cdc *Codec) Seal() *Codec {
	s, ok := cdc.imp.(Sealer)
	if ok {
		s.SealImp()
	} else {
		o := cdc.imp.(*aminoOrig.Codec)
		o.Seal()
	}
	cdc.cdc.Seal()
	return cdc
}
func (cdc *Codec) UnmarshalBinaryBare(bz []byte, ptr interface{}) error {
	return cdc.imp.UnmarshalBinaryBare(bz, ptr)
}
func (cdc *Codec) UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return cdc.imp.UnmarshalBinaryLengthPrefixed(bz, ptr)
}
func (cdc *Codec) UnmarshalBinaryLengthPrefixedReader(r io.Reader, ptr interface{}, maxSize int64) (n int64, err error) {
	return cdc.imp.UnmarshalBinaryLengthPrefixedReader(r, ptr, maxSize)
}
func (cdc *Codec) UnmarshalJSON(bz []byte, ptr interface{}) error {
	//fmt.Printf("cdc : %v\n", cdc)
	//fmt.Printf("cdc.cdc : %v\n", cdc.cdc)
	return cdc.cdc.UnmarshalJSON(bz, ptr)
}

////////////////////////////////////////////////////////////////////////////////////

func DeepCopy(o interface{}) (r interface{}) {
	r = Stub.DeepCopy(o)
	return
}

func MarshalBinaryBare(o interface{}) ([]byte, error) {
	return Stub.MarshalBinaryBare(o)
}

func MustMarshalBinaryLengthPrefixed(o interface{}) []byte {
	return Stub.MustMarshalBinaryLengthPrefixed(o)
}

func UnmarshalBinaryLengthPrefixed(bz []byte, ptr interface{}) error {
	return Stub.UnmarshalBinaryLengthPrefixed(bz, ptr)
}

func UvarintSize(u uint64) int {
	return Stub.UvarintSize(u)
}


func EncodeByteSlice(w io.Writer, bz []byte) (err error) {
	err = Stub.EncodeByteSlice(w, bz)
	return
}

func EncodeUvarint(w io.Writer, u uint64) (err error) {
	err = Stub.EncodeUvarint(w, u)
	return
}

func ByteSliceSize(bz []byte) int {
	return Stub.ByteSliceSize(bz)
}

func EncodeInt8(w io.Writer, i int8) (err error) {
	err = Stub.EncodeInt8(w, i)
	return
}
func EncodeInt16(w io.Writer, i int16) (err error) {
	err = Stub.EncodeInt16(w, i)
	return
}
func EncodeInt32(w io.Writer, i int32) (err error) {
	err = Stub.EncodeInt32(w, i)
	return
}
func EncodeInt64(w io.Writer, i int64) (err error) {
	err = Stub.EncodeInt64(w, i)
	return
}
func EncodeVarint(w io.Writer, i int64) (err error) {
	err = Stub.EncodeVarint(w, i)
	return
}
func EncodeByte(w io.Writer, b byte) (err error) {
	err = Stub.EncodeByte(w, b)
	return
}
func EncodeUint8(w io.Writer, u uint8) (err error) {
	err = Stub.EncodeUint8(w, u)
	return
}
func EncodeUint16(w io.Writer, u uint16) (err error) {
	err = Stub.EncodeUint16(w, u)
	return
}
func EncodeUint32(w io.Writer, u uint32) (err error) {
	err = Stub.EncodeUint32(w, u)
	return
}
func EncodeUint64(w io.Writer, u uint64) (err error) {
	err = Stub.EncodeUint64(w, u)
	return
}
func EncodeBool(w io.Writer, b bool) (err error) {
	err = Stub.EncodeBool(w, b)
	return
}
func EncodeFloat32(w io.Writer, f float32) (err error) {
	err = Stub.EncodeFloat32(w, f)
	return
}
func EncodeFloat64(w io.Writer, f float64) (err error) {
	err = Stub.EncodeFloat64(w, f)
	return
}
func EncodeString(w io.Writer, s string) (err error) {
	err = Stub.EncodeString(w, s)
	return
}
func DecodeInt8(bz []byte) (i int8, n int, err error) {
	i, n, err = Stub.DecodeInt8(bz)
	return
}
func DecodeInt16(bz []byte) (i int16, n int, err error) {
	i, n, err = Stub.DecodeInt16(bz)
	return
}
func DecodeInt32(bz []byte) (i int32, n int, err error) {
	i, n, err = Stub.DecodeInt32(bz)
	return
}
func DecodeInt64(bz []byte) (i int64, n int, err error) {
	i, n, err = Stub.DecodeInt64(bz)
	return
}
func DecodeVarint(bz []byte) (i int64, n int, err error) {
	i, n, err = Stub.DecodeVarint(bz)
	return
}
func DecodeByte(bz []byte) (b byte, n int, err error) {
	b, n, err = Stub.DecodeByte(bz)
	return
}
func DecodeUint8(bz []byte) (u uint8, n int, err error) {
	u, n, err = Stub.DecodeUint8(bz)
	return
}
func DecodeUint16(bz []byte) (u uint16, n int, err error) {
	u, n, err = Stub.DecodeUint16(bz)
	return
}
func DecodeUint32(bz []byte) (u uint32, n int, err error) {
	u, n, err = Stub.DecodeUint32(bz)
	return
}
func DecodeUint64(bz []byte) (u uint64, n int, err error) {
	u, n, err = Stub.DecodeUint64(bz)
	return
}
func DecodeUvarint(bz []byte) (u uint64, n int, err error) {
	u, n, err = Stub.DecodeUvarint(bz)
	return
}
func DecodeBool(bz []byte) (b bool, n int, err error) {
	b, n, err = Stub.DecodeBool(bz)
	return
}
func DecodeFloat32(bz []byte) (f float32, n int, err error) {
	f, n, err = Stub.DecodeFloat32(bz)
	return
}
func DecodeFloat64(bz []byte) (f float64, n int, err error) {
	f, n, err = Stub.DecodeFloat64(bz)
	return
}
func DecodeByteSlice(bz []byte) (bz2 []byte, n int, err error) {
	bz2, n, err = Stub.DecodeByteSlice(bz)
	return
}
func DecodeString(bz []byte) (s string, n int, err error) {
	s, n, err = Stub.DecodeString(bz)
	return
}
func VarintSize(i int64) int {
	return Stub.VarintSize(i)
}
