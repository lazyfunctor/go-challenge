package drum

import (
	_ "fmt"
	"io/ioutil"
)

func Encode(pat *Pattern) (data []byte, err error) {
	var e encodeState
	e.init(pat)
	data, err = e.encode()
	return
}

func EncodeToFile(path string, pat *Pattern) (err error) {
	var e encodeState
	e.init(pat)
	data, err := e.encode()

	if err != nil {
		return
	}
	err = ioutil.WriteFile(path, data, 0644)
	return
}
