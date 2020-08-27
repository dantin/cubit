package serializer

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockType struct {
	someField string
}

func (t *mockType) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	return dec.Decode(&t.someField)
}

func (t *mockType) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	return enc.Encode(&t.someField)
}

func TestSerialize(t *testing.T) {
	var v1, v2 mockType
	v1.someField = "foo value"

	b, err := Serialize(&v1)
	require.NotNil(t, b)
	require.Nil(t, err)

	err = Deserialize(b, &v2)

	require.Nil(t, err)
	require.True(t, reflect.DeepEqual(&v1, &v2))
}

func TestSerializeSlice(t *testing.T) {
	var v1, v2 []mockType
	v1 = append(v1, mockType{someField: "foo value 1"})
	v1 = append(v1, mockType{someField: "foo value 2"})

	b, err := SerializeSlice(&v1)
	require.NotNil(t, b)
	require.Nil(t, err)

	err = DeserializeSlice(b, &v2)

	require.Nil(t, err)
	require.True(t, reflect.DeepEqual(&v1, &v2))
}
