package json

import (
	"log"
	"testing"
)

// serde: Deserialize
type Test struct {
	A string
	B string
	C int32
	D int64
}

func TestDe_DeserializeAny(t *testing.T) {
	log.SetFlags(log.Lshortfile)

	content := `{
"C": -112323,
	"A": "xxx",
"D": 59583,
	"B": "yyy"
}`

	x := Test{}
	err := DeserializeFromString(content, &x)
	if err != nil {
		t.Error(err)
	}
	log.Printf("%#+v", x)
}

func BenchmarkDe_DeserializeAny(b *testing.B) {
	content := []byte(`{
"C": -112323,
	"A": "xxx",
	"B": "yyy"
}`)

	x := Test{}

	for i := 0; i < b.N; i++ {
		_ = DeserializeFromBytes(content, &x)
	}
	// for i := 0; i < b.N; i++ {
	// 	_ = json.Unmarshal([]byte(content), &x)
	// }
}
