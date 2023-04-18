package ignorefile

type Matcher interface {
	MatchPath(path string) bool
}

type NopMatcher struct{}

//nolint:revive
func (n NopMatcher) MatchPath(path string) bool {
	return false
}

var _ Matcher = &NopMatcher{}
