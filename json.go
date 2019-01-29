package main

import (
	"bytes"
	"encoding/json"
)

func toJSON(v interface{}) bytes.Buffer {
	var b bytes.Buffer

	err := json.NewEncoder(&b).Encode(v)
	if err != nil {
		panic(err)
	}

	return b
}

func fromJSON(b []byte, v interface{}) interface{} {

	r := bytes.NewReader(b)

	err := json.NewDecoder(r).Decode(v)
	if err != nil {
		panic(err)
	}

	return v
}
