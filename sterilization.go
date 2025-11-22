package hoard

import (
	"bytes"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

// Serialization helpers

var bufferPool = sync.Pool{
	New: func() interface{} { return new(bytes.Buffer) },
}

func Serialize(value interface{}) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufferPool.Put(buf)
	return msgpack.Marshal(value)
}

func Deserialize(data []byte) (interface{}, error) {
	var v interface{}
	err := msgpack.Unmarshal(data, &v)
	return v, err
}
