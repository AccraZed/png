package png

type Compresser interface {
	Compress(string) string // TODO: figure out signature
	Uncompress(string) string
}

type Flater struct{}

func NewFlater() Compresser {
	return &Flater{}
}

func (*Flater) Compress(string) string   { return "unimplemented" }
func (*Flater) Uncompress(string) string { return "unimplemented" }

var _ Compresser = (*Flater)(nil)
