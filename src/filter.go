package png

import (
	"fmt"
	"math"
)

var filts = []func(orig, a, b, c int) int{
	// None
	func(orig, a, b, c int) int {
		return orig
	},
	// Sub
	func(orig, a, b, c int) int {
		return orig - a
	},
	// Up
	func(orig, a, b, c int) int {
		return orig - b
	},
	// Average
	func(orig, a, b, c int) int {
		return (orig - int(math.Floor((float64(a+b))/2))) % 0xFF
	},
	// Paeth
	func(orig, a, b, c int) int {
		p := a + b - c
		pa := abs(p - a)
		pb := abs(p - b)
		pc := abs(p - c)

		if pa <= pb && pa <= pc {
			return a
		}
		if pc <= pc {
			return b
		}
		return c
	},
}

type Filterer interface {
	Filter(*ImageData) error // TODO: figure out signature
	Unfilter(*ImageData) error
}

type AdaptiveFilter struct {
	Width uint32
}

func (af *AdaptiveFilter) Filter(id *ImageData) error {
	if af.Width == 0 {
		return fmt.Errorf("invalid adaptive filter image width of 0")
	}

	idOut := make([]byte, len(id.data))

	var filt int
	for i, orig := range id.data {
		// check if byte is filter designator
		if i%int(af.Width+1) == 0 {
			filt = int(orig) % len(filts)
			idOut[i] = orig
			continue
		}

		upper := uint32(i) <= 4*af.Width
		left := uint32(i)%(4*af.Width) == 0

		a, b, c := 0, 0, 0
		if !left {
			a = int(id.data[i-4])
		}
		if !upper && !left {
			b = int(id.data[i-int(af.Width)*4])
		}
		if !upper {
			c = int(id.data[i-int(af.Width)*4-4])
		}

		idOut[i] = byte(filts[filt](int(orig), a, b, c))
	}

	return nil
}

func (af *AdaptiveFilter) Unfilter(id *ImageData) error { return nil }

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

var _ Filterer = (*AdaptiveFilter)(nil)
