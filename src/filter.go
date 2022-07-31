package png

type Filterer interface {
	Filter([]byte) ([]byte, error) // TODO: figure out signature
	Unfilter([]byte) ([]byte, error)
}

type AdaptiveFilter struct{}

func (*AdaptiveFilter) Filter(b []byte) ([]byte, error) {
	return b, nil
}

func (*AdaptiveFilter) Unfilter(b []byte) ([]byte, error) {
	return b, nil
}

var _ Filterer = (*AdaptiveFilter)(nil)
