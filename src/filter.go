package png

type Filterer interface {
	Filter(string) string // TODO: figure out signature
	Unfilter(string) string
}

type AdaptiveFilter struct{}

func (*AdaptiveFilter) Filter(string) string {
	return ""
}

func (*AdaptiveFilter) Unfilter(string) string {
	return ""
}

var _ Filterer = (*AdaptiveFilter)(nil)
