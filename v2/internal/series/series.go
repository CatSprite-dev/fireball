// Package series — значения во времени: курс, цена, количество по датам.
package series

import (
	"slices"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

var moscow = time.FixedZone("MSK", 3*60*60)

// Day — метка московского дня: полночь по UTC той же даты.
func Day(t time.Time) time.Time {
	y, m, d := t.In(moscow).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// EndOfDay — последний момент московского дня с меткой d.
func EndOfDay(d time.Time) time.Time {
	return time.Date(d.Year(), d.Month(), d.Day()+1, 0, 0, 0, 0, moscow).Add(-time.Nanosecond)
}

type Point struct {
	Date  time.Time
	Value decimal.Decimal
}

// Series — точки по возрастанию даты.
type Series []Point

// At — последняя точка не позже t; false — до t точек нет.
func (s Series) At(t time.Time) (Point, bool) {
	i := sort.Search(len(s), func(i int) bool { return s[i].Date.After(t) })
	if i == 0 {
		return Point{}, false
	}
	return s[i-1], true
}

// Sorted — копия ряда, упорядоченная по дате.
func Sorted(s Series) Series {
	s = slices.Clone(s)
	slices.SortFunc(s, func(a, b Point) int { return a.Date.Compare(b.Date) })
	return s
}
