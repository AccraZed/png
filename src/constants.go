package png

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

var (
	PNG_SIGNATURE = []byte{137, 80, 78, 71, 13, 10, 26, 10}
)
