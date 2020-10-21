package json

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/bits"
	"strconv"
	"strings"

	"github.com/Xuanwo/serde-go"
)

func DeserializeFromReader(r io.Reader, v serde.Deserializable) error {
	de := De{
		scanner: newScanner(r),
		state:   (*De).stateValue,
	}

	return v.Deserialize(&de)
}

func DeserializeFromString(s string, v serde.Deserializable) error {
	return DeserializeFromReader(strings.NewReader(s), v)
}

func DeserializeFromBytes(s []byte, v serde.Deserializable) error {
	return DeserializeFromReader(bytes.NewReader(s), v)
}

type stack []bool

func (s *stack) push(v bool) {
	*s = append(*s, v)
}

func (s *stack) pop() bool {
	*s = (*s)[:len(*s)-1]
	if len(*s) == 0 {
		return false
	}
	return (*s)[len(*s)-1]
}

func (s *stack) len() int { return len(*s) }

type De struct {
	scanner *scanner
	state   func(*De) ([]byte, error)
	tok     []byte // Peeked tok
	peeked  bool
	stack
}

func (d *De) peek() (_ []byte, err error) {
	if d.peeked {
		return d.tok, nil
	}
	d.tok, err = d.state(d)
	if err != nil {
		return
	}
	d.peeked = true
	return d.tok, nil
}

func (d *De) next() (_ []byte, err error) {
	if d.peeked {
		d.peeked = false
		return d.tok, nil
	}
	return d.state(d)
}

func (d *De) stateObjectString() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*De).stateEnd
		case inObj:
			d.state = (*De).stateObjectComma
		case !inObj:
			d.state = (*De).stateArrayComma
		}
		return tok, nil
	case '"':
		d.state = (*De).stateObjectColon
		return tok, nil
	default:
		return nil, fmt.Errorf("stateObjectString: missing string key")
	}
}

func (d *De) stateObjectColon() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case Colon:
		d.state = (*De).stateObjectValue
		return d.next()
	default:
		return tok, fmt.Errorf("stateObjectColon: expecting colon")
	}
}

func (d *De) stateObjectValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*De).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*De).stateArrayValue
		d.push(false)
		return tok, nil
	default:
		d.state = (*De).stateObjectComma
		return tok, nil
	}
}

func (d *De) stateObjectComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '}':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*De).stateEnd
		case inObj:
			d.state = (*De).stateObjectComma
		case !inObj:
			d.state = (*De).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*De).stateObjectString
		return d.next()
	default:
		return tok, fmt.Errorf("stateObjectComma: expecting comma")
	}
}

func (d *De) stateArrayValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*De).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*De).stateArrayValue
		d.push(false)
		return tok, nil
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*De).stateEnd
		case inObj:
			d.state = (*De).stateObjectComma
		case !inObj:
			d.state = (*De).stateArrayComma
		}
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateArrayValue: unexpected comma")
	default:
		d.state = (*De).stateArrayComma
		return tok, nil
	}
}

func (d *De) stateArrayComma() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case ']':
		inObj := d.pop()
		switch {
		case d.len() == 0:
			d.state = (*De).stateEnd
		case inObj:
			d.state = (*De).stateObjectComma
		case !inObj:
			d.state = (*De).stateArrayComma
		}
		return tok, nil
	case Comma:
		d.state = (*De).stateArrayValue
		return d.next()
	default:
		return nil, fmt.Errorf("stateArrayComma: expected comma, %v", d.stack)
	}
}

func (d *De) stateValue() ([]byte, error) {
	tok := d.scanner.Next()
	if len(tok) < 1 {
		return nil, io.ErrUnexpectedEOF
	}
	switch tok[0] {
	case '{':
		d.state = (*De).stateObjectString
		d.push(true)
		return tok, nil
	case '[':
		d.state = (*De).stateArrayValue
		d.push(false)
		return tok, nil
	case ',':
		return nil, fmt.Errorf("stateValue: unexpected comma")
	default:
		d.state = (*De).stateEnd
		return tok, nil
	}
}

func (d *De) stateEnd() ([]byte, error) { return nil, io.EOF }

func (d *De) parseBool() (bool, error) {
	tok, err := d.next()
	if err != nil {
		return false, err
	}
	switch string(tok) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, errors.New("expect bool")
	}
}
func (d *De) parseInt(bisSize int) (int64, error) {
	tok, err := d.next()
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseInt(string(tok), 10, bisSize)
	if err != nil {
		return 0, err
	}
	return v, nil
}
func (d *De) parseUint(bisSize int) (uint64, error) {
	tok, err := d.next()
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseUint(string(tok), 10, bisSize)
	if err != nil {
		return 0, err
	}
	return v, nil
}
func (d *De) parseFloat(bisSize int) (float64, error) {
	tok, err := d.next()
	if err != nil {
		return 0, err
	}

	v, err := strconv.ParseFloat(string(tok), bisSize)
	if err != nil {
		return 0, err
	}
	return v, nil
}

func (d *De) DeserializeAny(v serde.Visitor) (err error) {
	tok, err := d.peek()
	if err != nil {
		return
	}

	switch tok[0] {
	case '{':
		return d.DeserializeMap(v)
	case '[':
		return d.DeserializeSlice(v)
	case True, False:
		return d.DeserializeBool(v)
	case Null:
		return v.VisitNil()
	case '"':
		return d.DeserializeString(v)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		return d.DeserializeUint64(v)
	case '-':
		return d.DeserializeInt64(v)
	default:
		panic("unhandled type")
	}

	return nil
}

func (d *De) DeserializeBool(v serde.Visitor) (err error) {
	b, err := d.parseBool()
	if err != nil {
		return
	}
	return v.VisitBool(b)
}

func (d *De) DeserializeInt(v serde.Visitor) (err error) {
	i, err := d.parseInt(bits.UintSize)
	if err != nil {
		return
	}
	return v.VisitInt(int(i))
}

func (d *De) DeserializeInt8(v serde.Visitor) (err error) {
	i, err := d.parseInt(8)
	if err != nil {
		return
	}
	return v.VisitInt8(int8(i))
}

func (d *De) DeserializeInt16(v serde.Visitor) (err error) {
	i, err := d.parseInt(16)
	if err != nil {
		return
	}
	return v.VisitInt16(int16(i))
}

func (d *De) DeserializeInt32(v serde.Visitor) (err error) {
	i, err := d.parseInt(32)
	if err != nil {
		return
	}
	return v.VisitInt32(int32(i))
}

func (d *De) DeserializeInt64(v serde.Visitor) (err error) {
	i, err := d.parseInt(64)
	if err != nil {
		return
	}
	return v.VisitInt64(i)
}

func (d *De) DeserializeUint(v serde.Visitor) (err error) {
	i, err := d.parseUint(bits.UintSize)
	if err != nil {
		return
	}
	return v.VisitUint(uint(i))
}

func (d *De) DeserializeUint8(v serde.Visitor) (err error) {
	i, err := d.parseUint(8)
	if err != nil {
		return
	}
	return v.VisitUint8(uint8(i))
}

func (d *De) DeserializeUint16(v serde.Visitor) (err error) {
	i, err := d.parseUint(16)
	if err != nil {
		return
	}
	return v.VisitUint16(uint16(i))
}

func (d *De) DeserializeUint32(v serde.Visitor) (err error) {
	i, err := d.parseUint(32)
	if err != nil {
		return
	}
	return v.VisitUint32(uint32(i))
}

func (d *De) DeserializeUint64(v serde.Visitor) (err error) {
	i, err := d.parseUint(64)
	if err != nil {
		return
	}
	return v.VisitUint64(i)
}

func (d *De) DeserializeFloat32(v serde.Visitor) (err error) {
	f, err := d.parseFloat(32)
	if err != nil {
		return
	}
	return v.VisitFloat32(float32(f))
}

func (d *De) DeserializeFloat64(v serde.Visitor) (err error) {
	f, err := d.parseFloat(64)
	if err != nil {
		return
	}
	return v.VisitFloat64(f)
}

func (d *De) DeserializeComplex64(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeComplex128(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeRune(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeString(v serde.Visitor) (err error) {
	tok, err := d.next()
	if err != nil {
		return err
	}
	return v.VisitString(string(tok[1 : len(tok)-1]))
}

func (d *De) DeserializeByte(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeBytes(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeTime(v serde.Visitor) (err error) {
	panic("implement me")
}

func (d *De) DeserializeSlice(v serde.Visitor) (err error) {
	tok, err := d.next()
	if err != nil {
		return
	}

	if tok[0] != ArrayStart {
		return errors.New("expect array start")
	}

	err = v.VisitSlice(newCommaSeparated(d))
	if err != nil {
		return
	}

	tok, err = d.next()
	if err != nil {
		return
	}
	if tok[0] != ArrayEnd {
		return errors.New("expect array end")
	}

	return nil
}

func (d *De) DeserializeMap(v serde.Visitor) (err error) {
	tok, err := d.next()
	if err != nil {
		return
	}

	if tok[0] != ObjectStart {
		// TODO: should be extracted.
		return errors.New("expect map start")
	}

	err = v.VisitMap(newCommaSeparated(d))
	if err != nil {
		return
	}

	tok, err = d.next()
	if err != nil {
		return
	}
	if tok[0] != ObjectEnd {
		// TODO: should be extracted.
		return errors.New("expect map end")
	}

	return nil
}

func (d *De) DeserializeStruct(name string, fields []string, v serde.Visitor) (err error) {
	return d.DeserializeMap(v)
}

type commaSeparated struct {
	de *De
}

func newCommaSeparated(de *De) *commaSeparated {
	return &commaSeparated{
		de: de,
	}
}

func (c *commaSeparated) NextKey(v serde.Visitor) (ok bool, err error) {
	tok, err := c.de.peek()
	if err != nil {
		return
	}
	if tok[0] == ObjectEnd {
		return false, nil
	}

	err = c.de.DeserializeString(v)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *commaSeparated) NextValue(v serde.Visitor) (err error) {
	return c.de.DeserializeAny(v)
}

func (c *commaSeparated) NextElement(v serde.Visitor) (ok bool, err error) {
	tok, err := c.de.peek()
	if err != nil {
		return
	}
	if tok[0] == ArrayEnd {
		return false, nil
	}

	err = c.de.DeserializeAny(v)
	if err != nil {
		return false, err
	}
	return true, nil
}
