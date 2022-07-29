package png

type Interlacer interface {
	Interlace(string) string
	Unterlace(string) string
}

type Adam7 struct{}

func NewAdam7() Interlacer {
	return &Adam7{}
}

func (*Adam7) Interlace(string) string { return "unimplemented" }
func (*Adam7) Unterlace(string) string { return "unimplemented" }

type NoInterlacer struct{}

func NewNoInterlacer() Interlacer {
	return &NoInterlacer{}
}
func (*NoInterlacer) Interlace(a string) string { return a }
func (*NoInterlacer) Unterlace(a string) string { return a }

var _ Interlacer = (*Adam7)(nil)
var _ Interlacer = (*NoInterlacer)(nil)
