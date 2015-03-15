// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"math"
)

type decodeState struct {
	data     []byte
	offset   int
	totallen int
}

func (d *decodeState) init(data []byte) *decodeState {
	d.data = data
	d.offset = 0
	d.totallen = len(data)
	return d
}

func (d *decodeState) error(err error) {
	panic(err)
}

func (d *decodeState) readBytes(n int) []byte {
	bytesRead := d.data[d.offset : d.offset+n]
	d.offset += n
	return bytesRead
}

func (d *decodeState) readHeader() (version string, tempo float32) {
	header := d.readBytes(6)
	if string(header) != "SPLICE" {
		d.error(&parseError{msg: "header missing", offset: d.offset})
	}
	_ = d.readBytes(7)
	dataLen := int(d.readBytes(1)[0])
	if d.totallen < dataLen {
		d.error(&parseError{msg: "Incomplete file", offset: d.offset})
	} else {
		d.totallen = dataLen
	}

	version_info := d.readBytes(32)
	version = string(bytes.Trim(version_info, "\x00"))

	tempo_bytes := d.readBytes(4)
	bits := binary.LittleEndian.Uint32(tempo_bytes)
	tempo = math.Float32frombits(bits)
	return
}

func (d *decodeState) parseTrack() *Track {
	trackInfo := d.readBytes(4)
	trackID := int(trackInfo[0])
	nameLen := int(d.readBytes(1)[0])
	name := string(d.readBytes(nameLen))
	step_bytes := d.readBytes(16)
	var steps [16]byte
	copy(steps[:], step_bytes)
	track := &Track{trackID: trackID, name: name, steps: steps}
	return track
}

func (d *decodeState) decode() (pat *Pattern) {
	version, tempo := d.readHeader()
	pat = &Pattern{}
	pat.version = version
	pat.tempo = tempo
	var tracks []*Track
	pat.tracks = tracks
	for d.offset < d.totallen {
		track := d.parseTrack()
		pat.tracks = append(pat.tracks, track)
	}
	return
}
