package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io/ioutil"
	"os"
)

type Transcoder struct {
	Width, Height uint32
	BitDepth      BitDepth
	ColorType     ColorType
	Interlace     InterlaceMethod // whether Adam7 Interlacer is used

	// seenset keeps track of which index in data the ChunkTypes begin
	// no support for multiple chunks of the same type yet
	SeenSet    map[string][]uint32
	DataChunks []*ImageData
	DataState  DataState

	Filterer   *AdaptiveFilter
	compressor Compresser
}
type ImageData struct {
	DataState DataState
	data      []byte
	Scanlines []Scanline
}

type Scanline struct {
	pos       uint
	numPixels uint
	filter    FilterType
}

func (t *Transcoder) String() string {
	return fmt.Sprintf("img W/H %vx%v BD/CT %v/%v SeenSet %v",
		t.Width, t.Height, t.BitDepth, t.ColorType, t.SeenSet)
}

func NewTranscoder(f *os.File) (*Transcoder, error) {
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("unable to read file: %v", err)
	}

	if !bytes.Equal(b[:8], PNG_SIGNATURE) {
		return nil, fmt.Errorf("invalid png header")
	}

	//? todo: consider the very chaotic idea of making chunk processing
	//? concurrent for the memes hehehe
	t := &Transcoder{
		SeenSet:   make(map[string][]uint32),
		DataState: DataStateCompressed,
	}
	// Process raw data
	var pos uint32 = 8
	for {
		if _, ok := t.SeenSet["IEND"]; ok {
			break
		}

		loc := pos
		length := binary.BigEndian.Uint32(b[pos : pos+4])
		pos += 4
		typ := b[pos : pos+4]
		pos += 4
		chunk := b[pos : pos+length]
		pos += length
		crc := b[pos : pos+4]
		pos += 4

		err := t.initChunk(loc, typ, chunk, crc)
		if err != nil {
			return nil, err
		}
	}

	return t, err
}

func (t *Transcoder) initChunk(loc uint32, typ, rawData, crc []byte) error {
	if crc32.ChecksumIEEE(append(typ, rawData...)) != binary.BigEndian.Uint32(crc) {
		return fmt.Errorf("crc32 failed for chunk %s (byte %v)", string(typ), loc)
	}

	switch string(typ) {
	case "IHDR":
		t.Width = binary.BigEndian.Uint32(rawData[:4])
		t.Height = binary.BigEndian.Uint32(rawData[4:8])

		t.BitDepth = BitDepth(rawData[8])
		t.ColorType = ColorType(rawData[9])
		if err := verifyBitDepthAndColorType(t.BitDepth, t.ColorType); err != nil {
			return err
		}

		// Compression method
		switch rawData[10] {
		case 0:
			t.compressor = NewFlater()
		default:
			return fmt.Errorf("unsupported compressor type: %v", rawData[10])
		}

		// Filter Method
		switch rawData[11] {
		case 0:
			t.Filterer = &AdaptiveFilter{Width: t.Width}
		default:
			return fmt.Errorf("unsupported filter method: %v", rawData[11])
		}

		// Interlace method
		t.Interlace = InterlaceMethod(rawData[12])

	case "IDAT":
		if _, ok := t.SeenSet["IHDR"]; !ok {
			return fmt.Errorf("IDAT header declared before IHDR")
		}
		t.DataChunks = append(t.DataChunks, t.initImageData(rawData))
	case "IEND":
		// only seenset is updated
	default:
		fmt.Printf("WARNING: unimplemented type %s\n", string(typ))
	}

	if _, ok := t.SeenSet[string(typ)]; !ok {
		t.SeenSet[string(typ)] = make([]uint32, 0)
	}
	t.SeenSet[string(typ)] = append(t.SeenSet[string(typ)], loc)

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

func (t *Transcoder) initImageData(rawData []byte) *ImageData {

	return &ImageData{
		DataState: DataStateUnfiltered,
		data:      rawData,
		Scanlines: nil,
	}
}
