package json

import (
	"log"
	"testing"
)

func TestSer(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	x := Test{
		A: "test",
	}
	s := NewSer()
	err := x.Serialize(s)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%v", string(s.buf.Bytes()))
}
