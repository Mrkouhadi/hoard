package hoard

import (
	"bytes"
	"sync"

	"github.com/vmihailenco/msgpack/v5"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func serialize(value interface{}) ([]byte, error) {
	buf := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buf)
	buf.Reset()

	return msgpack.Marshal(value)
}

func deserialize(data []byte) (interface{}, error) {
	var value interface{}
	err := msgpack.Unmarshal(data, &value)
	return value, err
}
