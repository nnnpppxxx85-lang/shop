package db

import (
	"sort"
	"strconv"
)

// NaturalSortStrings сортирует имена файлов по-человечески: "_2" раньше "_10".
// Обычная лексикографическая сортировка расставила бы "_10" перед "_2",
// что перепутало бы порядок фотографий в карточке товара при 10+ фото в папке.
func NaturalSortStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	sort.SliceStable(out, func(i, j int) bool {
		return naturalLess(out[i], out[j])
	})
	return out
}

func naturalLess(a, b string) bool {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		ca, cb := a[i], b[j]
		if isDigit(ca) && isDigit(cb) {
			ni, ei := readNumber(a, i)
			nj, ej := readNumber(b, j)
			if ni != nj {
				return ni < nj
			}
			i, j = ei, ej
			continue
		}
		if ca != cb {
			return ca < cb
		}
		i++
		j++
	}
	return len(a)-i < len(b)-j
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }

func readNumber(s string, start int) (int, int) {
	end := start
	for end < len(s) && isDigit(s[end]) {
		end++
	}
	n, _ := strconv.Atoi(s[start:end])
	return n, end
}
