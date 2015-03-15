package drum

import (
	"fmt"
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

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []*Track
}

func (pat *Pattern) String() string {
	var prefix string
	// dirty hack to truncate to no decimal place in case of whole numbers
	if pat.tempo == float32(int32(pat.tempo)) {
		prefix = fmt.Sprintf("Saved with HW Version: %s\nTempo: %.0f\n", pat.version, pat.tempo)
	} else {
		prefix = fmt.Sprintf("Saved with HW Version: %s\nTempo: %.1f\n", pat.version, pat.tempo)
	}
	trackStr := ""
	for _, track := range pat.tracks {
		trackStr += fmt.Sprint(track) + "\n"
	}
	return prefix + trackStr
}

type Track struct {
	trackID int
	name    string
	steps   [16]byte
}

func (track *Track) String() string {
	prefix := fmt.Sprintf("(%d) %s\t", track.trackID, track.name)
	measure := "|"
	for i, step := range track.steps {
		if step == 0 {
			measure += "-"
		} else if step == 1 {
			measure += "x"
		}

		if (i+1)%4 == 0 {
			measure += "|"
		}
	}
	return prefix + measure
}

type parseError struct {
	msg    string
	offset int
}

func (e *parseError) Error() string {
	return e.msg
}
