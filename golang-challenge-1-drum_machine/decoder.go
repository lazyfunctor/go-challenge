package drum

import (
	_ "fmt"
	"io/ioutil"
	"runtime"
)

func DecodeFile(path string) (pat *Pattern, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()
	data, err := ioutil.ReadFile(path)
	var d decodeState
	d.init(data)
	pat = d.decode()
	return
}
