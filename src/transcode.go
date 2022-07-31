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

	// seenset keeps track of which index in data the ChunkTypes begin
	// no support for multiple chunks of the same type yet
	SeenSet   map[string]uint32
	Data      []byte
	DataState DataState

	Filterer   *AdaptiveFilter
	compressor Compresser
	interlacer Interlacer
}

func (t *Transcoder) String() string {
	return fmt.Sprintf("img W/H %vx%v BD/CT %v/%v SeenSet %v",
		t.Width, t.Height, t.BitDepth, t.ColorType, t.SeenSet)
}

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
	t := &Transcoder{
		SeenSet:   make(map[string]uint32),
		DataState: DataStateCompressed,
	}
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

func (t *Transcoder) initChunk(loc uint32, typ, data, crc []byte) error {
	if crc32.ChecksumIEEE(append(typ, data...)) != binary.BigEndian.Uint32(crc) {
		return fmt.Errorf("crc32 failed for chunk %s (byte %v)", string(typ), loc)
	}

	switch string(typ) {
	case "IHDR":
		t.Width = binary.BigEndian.Uint32(data[:4])
		t.Height = binary.BigEndian.Uint32(data[4:8])

		t.BitDepth = BitDepth(data[8])
		t.ColorType = ColorType(data[9])
		if err := verifyBitDepthAndColorType(t.BitDepth, t.ColorType); err != nil {
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
			t.Filterer = &AdaptiveFilter{}
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
		if _, ok := t.SeenSet["IHDR"]; !ok {
			return fmt.Errorf("IDAT header declared before IHDR")
		}
		t.Data = data
	case "IEND":
		// only seenset is updated
	default:
		fmt.Printf("WARNING: unimplemented type %s\n", string(typ))
	}

	t.SeenSet[string(typ)] = loc
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

// TODO: Look into how to make this state changy shit not as gross
// Transcode converts the current image Data to the target state data
func (t *Transcoder) Transcode(target DataState) error {
	if target == DataStateUnknown {
		return fmt.Errorf("target DataState is unknown")
	}

	diff := int(target) - int(t.DataState)
	if diff == 0 {
		return nil
	}

	var err error
	switch t.DataState {
	case DataStateRaw:
		if diff > 0 {
			t.Data, err = t.Filterer.Filter(t.Data)
			if err != nil {
				return fmt.Errorf("error filtering: %w", err)
			}
			t.DataState = DataStateFiltered
		} else {
			return fmt.Errorf("cannot go down from DataStateRaw")
		}
	case DataStateFiltered:
		if diff > 0 {
			t.Data, err = t.interlacer.Interlace(t.Data)
			if err != nil {
				return fmt.Errorf("error interlacing: %w", err)
			}
			t.DataState = DataStateInterlaced
		} else {
			t.Data, err = t.Filterer.Unfilter(t.Data)
			if err != nil {
				return fmt.Errorf("error unfiltering: %w", err)
			}
			t.DataState = DataStateRaw
		}
	case DataStateInterlaced:
		if diff > 0 {
			t.Data, err = t.compressor.Compress(t.Data)
			if err != nil {
				return fmt.Errorf("error compressing: %w", err)
			}
			t.DataState = DataStateCompressed
		} else {
			t.Data, err = t.interlacer.Unterlace(t.Data)
			if err != nil {
				return fmt.Errorf("error unterlacing: %w", err)
			}
			t.DataState = DataStateFiltered
		}
	case DataStateCompressed:
		if diff > 0 {
			return fmt.Errorf("cannot go up from DataStateCompressed")
		} else {
			t.Data, err = t.compressor.Uncompress(t.Data)
			if err != nil {
				return fmt.Errorf("error uncompressing: %w", err)
			}
			t.DataState = DataStateInterlaced
		}
	default:
		return fmt.Errorf("unknown DataState")
	}

	return t.Transcode(target)
}
