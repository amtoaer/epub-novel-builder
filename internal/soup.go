package internal

import "github.com/anaskhan96/soup"

type SafeSoup struct {
	soup.Root
}

func (s SafeSoup) SafeFind(args ...string) SafeSoup {
	if s.Pointer != nil {
		s = SafeSoup{s.Find(args...)}
	}
	return s
}

func (s SafeSoup) SafeFindAll(args ...string) ([]SafeSoup, error) {
	if s.Pointer != nil {
		tmp := s.FindAll(args...)
		result := make([]SafeSoup, len(tmp))
		for idx := range tmp {
			result[idx] = SafeSoup{tmp[idx]}
		}
		return result, nil
	}
	return nil, s.Error
}
