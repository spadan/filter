package filter

type StringSet map[string]struct{}

func NewStringSet(keys ...string) StringSet {
	r := StringSet{}
	for _, key := range keys {
		r[key] = struct{}{}
	}
	return r
}

func (s StringSet) Add(keys StringSet) {
	for key := range keys {
		s[key] = struct{}{}
	}
}

func (s StringSet) Contains(key string) bool {
	_, exist := s[key]
	return exist
}

// IsIntersection 是否存在交集
func (s StringSet) IsIntersection(other StringSet) bool {
	for key := range other {
		if s.Contains(key) {
			return true
		}
	}
	return false
}
