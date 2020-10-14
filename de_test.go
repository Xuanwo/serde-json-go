package json

import (
	"log"
	"testing"

	"github.com/Xuanwo/serde-go"
)

// serde: Deserialize
type Test struct {
	A string
	B string
	C int
}

type testFieldEnum = int

const (
	testFieldEnumA testFieldEnum = iota + 1
	testFieldEnumB
	testFieldEnumC
)

type testFieldVisitor struct {
	e testFieldEnum

	serde.DummyVisitor
}

func (t *testFieldVisitor) VisitString(v string) (err error) {
	switch v {
	case "A":
		t.e = testFieldEnumA
	case "B":
		t.e = testFieldEnumB
	case "C":
		t.e = testFieldEnumC
	}
	return nil
}

type testVisitor struct {
	v   *Test
	idx int

	serde.DummyVisitor
}

func (t *testVisitor) VisitSlice(s serde.SliceAccess) (err error) {
	_, err = s.NextElement(serde.NewStringVisitor(&t.v.A))
	if err != nil {
		return
	}
	_, err = s.NextElement(serde.NewStringVisitor(&t.v.B))
	if err != nil {
		return
	}
	return nil
}

func (t *testVisitor) VisitMap(m serde.MapAccess) (err error) {
	field := &testFieldVisitor{}
	for {
		ok, err := m.NextKey(field)
		if !ok {
			break
		}
		if err != nil {
			return err
		}

		var v serde.Visitor
		switch field.e {
		case testFieldEnumA:
			v = serde.NewStringVisitor(&t.v.A)
		case testFieldEnumB:
			v = serde.NewStringVisitor(&t.v.B)
		case testFieldEnumC:
			v = serde.NewIntVisitor(&t.v.C)
		}
		err = m.NextValue(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Test) Deserialize(de serde.Deserializer) (err error) {
	return de.DeserializeStruct("Test", []string{"A", "B"}, &testVisitor{v: t})
}

func TestDe_DeserializeAny(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	content := `{
"C": 123,
	"A": "xxx",
	"B": "yyy"
}`

	x := Test{}
	err := DeserializeFromString(content, &x)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%#+v", x)
}
