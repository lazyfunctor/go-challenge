// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information

package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

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

type decodeState struct {
	data    []byte
	offset  int
	datalen int
}

func (d *decodeState) init(data []byte) *decodeState {
	d.data = data
	d.offset = 0
	d.datalen = len(data)
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
	if d.datalen < dataLen {
		d.error(&parseError{msg: "Incomplete file", offset: d.offset})
	} else {
		d.datalen = dataLen
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
	for d.offset < d.datalen {
		track := d.parseTrack()
		pat.tracks = append(pat.tracks, track)
	}
	return
}

type encodeState struct {
	pat     *Pattern
	offset  int
	data    []byte
	dataLen int
}

func (e *encodeState) init(pat *Pattern) {
	e.pat = pat
	e.offset = 0
	e.dataLen = e.calcLength()
	totalLen := e.dataLen + 14 //6 bytes for main header + 7 empty bytes + 1 length byte
	e.data = make([]byte, totalLen)
}

// make encddeState an io.Writer
func (e *encodeState) Write(p []byte) (n int, err error) {
	for idx, val := range p {
		e.data[e.offset+idx] = val
		n += 1
	}
	e.offset += n
	return
}

func (e *encodeState) writeHeader() (err error) {
	var buf bytes.Buffer
	buf.Write([]byte("SPLICE"))
	buf.Write(bytes.Repeat([]byte{0}, 7))
	buf.WriteByte(byte(e.dataLen))
	ver_len := len(e.pat.version)
	buf.Write([]byte(e.pat.version))
	buf.Write(bytes.Repeat([]byte{0}, (32 - ver_len)))
	bits := math.Float32bits(e.pat.tempo)
	tempo := make([]byte, 4)
	binary.LittleEndian.PutUint32(tempo, bits)
	buf.Write(tempo)
	_, err = buf.WriteTo(e)
	return
}

func (e *encodeState) writeTrack(track *Track) (err error) {
	var buf bytes.Buffer
	buf.WriteByte(byte(track.trackID))
	buf.Write(bytes.Repeat([]byte{0}, 3))
	buf.WriteByte(byte(len(track.name)))
	buf.Write([]byte(track.name))
	buf.Write(track.steps[:])
	_, err = buf.WriteTo(e)
	return
}

func (e *encodeState) calcLength() int {
	dataLen := 36 //version + tempo
	for _, track := range e.pat.tracks {
		dataLen += 5               // trackid + len byte
		dataLen += len(track.name) // track name
		dataLen += 16              // steps
	}
	return dataLen
}

func (e *encodeState) encode() (data []byte, err error) {
	err = e.writeHeader()
	if err != nil {
		return
	}
	for _, track := range e.pat.tracks {
		e.writeTrack(track)
	}
	data = e.data
	return
}
