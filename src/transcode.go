package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
)

type Transcoder struct {
	width, height uint32
	bitDepth      BitDepth
	colorType     ColorType

	filterer *AdaptiveFilter

	compressor Compresser
	interlacer Interlacer
}

func (t *Transcoder) Transcode(raw, target string) string {
	filtered := t.filterer.Filter(raw)
	compressed := t.compressor.Compress(filtered)

	return compressed
}

var (
	PNG_SIGNATURE = []byte{137, 80, 78, 71, 13, 10, 26, 10}
)

type ColorType uint16

const (
	CTUnknown ColorType = 0
	CT2       ColorType = 2
	CT3       ColorType = 3
	CT4       ColorType = 4
	CT6       ColorType = 6
)

type BitDepth uint16

const (
	BDUnknown BitDepth = 0
	BD1       BitDepth = 1
	BD2       BitDepth = 2
	BD4       BitDepth = 4
	BD8       BitDepth = 8
	BD16      BitDepth = 16
)

func NewTranscoder(path string) (*Transcoder, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unable to open file: %v", err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}

	if !bytes.Equal(b[:8], PNG_SIGNATURE) {
		return nil, fmt.Errorf("invalid png header")
	}

	//? todo: consider the very chaotic idea of making chunk processing
	//? concurrent for the memes hehehe
	t := &Transcoder{}
	var pos uint32 = 8
	for {
		length := binary.BigEndian.Uint32(b[pos : pos+4])
		pos += 4
		typ := b[pos : pos+4]
		pos += 4
		data := b[pos : pos+length]
		pos += length
		crc := b[pos : pos+4]
		pos += 4

		t.processChunk(typ, data, crc)
	}
}

func (t *Transcoder) processChunk(typ, data, crc []byte) error {
	switch string(typ) {
	case "IHDR":
		t.width = binary.BigEndian.Uint32(data[:4])
		t.height = binary.BigEndian.Uint32(data[4:8])

		t.bitDepth = BitDepth(data[8])
		t.colorType = ColorType(data[9])
		if err := verifyBitDepthAndColorType(t.bitDepth, t.colorType); err != nil {
			return err
		}

		// Compression method
		switch data[10] {
		case 0:
			t.compressor = NewFlater()
		default:
			return fmt.Errorf("unsupported compressor type: %v", data[10])
		}

		// Filter Method
		switch data[11] {
		case 0:
			t.filterer = &AdaptiveFilter{}
		default:
			return fmt.Errorf("unsupported filter method: %v", data[11])
		}

		// Interlace method
		switch data[12] {
		case 0:
			t.interlacer = NewNoInterlacer()
		case 1:
			t.interlacer = NewAdam7()
		default:
			return fmt.Errorf("unsupported interlace method : %v", data[12])
		}
	case "IDAT":
	case "PLTE":
	}

	return nil
}

func verifyBitDepthAndColorType(bd BitDepth, ct ColorType) error {
	if ct == CTUnknown || bd == BDUnknown {
		return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
	}
	switch ct {
	case CT2, CT4, CT6:
		if bd < BD8 {
			return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
		}
	case CT3:
		if bd > BD8 {
			return fmt.Errorf("invalid ColorType/BitDepth: %v/%v", ct, bd)
		}
	}

	return nil
}
