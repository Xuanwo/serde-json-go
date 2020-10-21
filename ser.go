package json

import (
	"time"

	"github.com/Xuanwo/go-bufferpool"
	"github.com/Xuanwo/serde-go"
)

var pool = bufferpool.New(1024)

type Ser struct {
	buf *bufferpool.Buffer
}

func NewSer() *Ser {
	return &Ser{buf: pool.Get()}
}

func (s *Ser) SerializeBool(v bool) (err error) {
	s.buf.AppendBool(v)
	return
}

func (s *Ser) SerializeInt(v int) (err error) {
	s.buf.AppendInt(int64(v))
	return
}

func (s *Ser) SerializeInt8(v int8) (err error) {
	s.buf.AppendInt(int64(v))
	return
}

func (s *Ser) SerializeInt16(v int16) (err error) {
	s.buf.AppendInt(int64(v))
	return
}

func (s *Ser) SerializeInt32(v int32) (err error) {
	s.buf.AppendInt(int64(v))
	return
}

func (s *Ser) SerializeInt64(v int64) (err error) {
	s.buf.AppendInt(v)
	return
}

func (s *Ser) SerializeUint(v uint) (err error) {
	s.buf.AppendUint(uint64(v))
	return
}

func (s *Ser) SerializeUint8(v uint8) (err error) {
	s.buf.AppendUint(uint64(v))
	return
}

func (s *Ser) SerializeUint16(v uint16) (err error) {
	s.buf.AppendUint(uint64(v))
	return
}

func (s *Ser) SerializeUint32(v uint32) (err error) {
	s.buf.AppendUint(uint64(v))
	return
}

func (s *Ser) SerializeUint64(v uint64) (err error) {
	s.buf.AppendUint(v)
	return
}

// TODO: we need to support different "fmt byte, prec, bitSize int" supported in JSON
func (s *Ser) SerializeFloat32(v float32) (err error) {
	s.buf.AppendFloat(float64(v))
	return
}

func (s *Ser) SerializeFloat64(v float64) (err error) {
	s.buf.AppendFloat(v)
	return
}

func (s *Ser) SerializeComplex64(v complex64) (err error) {
	s.buf.AppendFloat(float64(real(v)))
	s.buf.AppendByte('+')
	s.buf.AppendFloat(float64(imag(v)))
	s.buf.AppendByte('i')
	return
}

func (s *Ser) SerializeComplex128(v complex128) (err error) {
	s.buf.AppendFloat(real(v))
	s.buf.AppendByte('+')
	s.buf.AppendFloat(imag(v))
	s.buf.AppendByte('i')
	return
}

func (s *Ser) SerializeRune(v rune) (err error) {
	s.buf.AppendRune(v)
	return
}

func (s *Ser) SerializeString(v string) (err error) {
	s.buf.AppendRune('"')
	s.buf.AppendString(v)
	s.buf.AppendRune('"')
	return
}

func (s *Ser) SerializeByte(v byte) (err error) {
	s.buf.AppendByte(v)
	return
}

func (s *Ser) SerializeBytes(v []byte) (err error) {
	s.buf.AppendBytes(v)
	return
}

// TODO: need support different time layout.
func (s *Ser) SerializeTime(v time.Time) (err error) {
	s.buf.AppendTime(v, time.RFC822)
	return
}

func (s *Ser) SerializeSlice(length int) (m serde.SliceSerializer, err error) {
	s.buf.AppendByte('[')
	return s, err
}

func (s *Ser) SerializeMap(length int) (m serde.MapSerializer, err error) {
	s.buf.AppendByte('{')
	return s, nil
}

func (s *Ser) SerializeStruct(name string, length int) (st serde.StructSerializer, err error) {
	s.buf.AppendByte('{')
	return s, nil
}

func (s *Ser) SerializeElement(v serde.Serializable) (err error) {
	if s.buf.Bytes()[len(s.buf.Bytes())-1] != '[' {
		s.buf.AppendByte(',')
	}
	err = v.Serialize(s)
	if err != nil {
		return
	}
	return nil
}

func (s *Ser) EndSlice() (err error) {
	s.buf.AppendByte(']')
	return nil
}

func (s *Ser) SerializeEntry(k, v serde.Serializable) (err error) {
	// TODO: could be optimized
	if s.buf.Bytes()[len(s.buf.Bytes())-1] != '{' {
		s.buf.AppendByte(',')
	}
	err = k.Serialize(s)
	if err != nil {
		return
	}
	s.buf.AppendByte(':')
	err = v.Serialize(s)
	if err != nil {
		return
	}
	return nil
}

func (s *Ser) EndMap() (err error) {
	s.buf.AppendByte('}')
	return nil
}

func (s *Ser) SerializeField(k, v serde.Serializable) (err error) {
	// TODO: could be optimized
	if s.buf.Bytes()[len(s.buf.Bytes())-1] != '{' {
		s.buf.AppendByte(',')
	}
	err = k.Serialize(s)
	if err != nil {
		return
	}
	s.buf.AppendByte(':')
	err = v.Serialize(s)
	if err != nil {
		return
	}
	return nil
}

func (s *Ser) EndStruct() (err error) {
	s.buf.AppendByte('}')
	return nil
}
