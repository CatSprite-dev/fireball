package pkg

import "time"

type Split struct {
	Date time.Time
	Coef float64
}

var Splits = map[string]Split{
	"T": { // Т-Технологии
		Date: time.Date(2026, 4, 17, 0, 0, 0, 0, time.UTC),
		Coef: 10,
	},
	"PLZL": { // Полюс
		Date: time.Date(2025, 3, 27, 0, 0, 0, 0, time.UTC),
		Coef: 10,
	},
	"VTBR": { // Банк ВТБ
		Date: time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC),
		Coef: float64(1) / 5000,
	},
	"GMKN": { // Норильский Никель
		Date: time.Date(2024, 4, 4, 0, 0, 0, 0, time.UTC),
		Coef: 100,
	},
	"TRNF_p": { // Транснефть (прив.)
		Date: time.Date(2024, 2, 19, 0, 0, 0, 0, time.UTC),
		Coef: 100,
	},
	"ROLO": { // Русолово
		Date: time.Date(2023, 1, 17, 0, 0, 0, 0, time.UTC),
		Coef: 10,
	},
	"RNFT": { // НК РуссНефть
		Date: time.Date(2016, 11, 16, 0, 0, 0, 0, time.UTC),
		Coef: 2000,
	},
	"RGSS": { // СК Росгосстрах
		Date: time.Date(2016, 4, 27, 0, 0, 0, 0, time.UTC),
		Coef: 15.002,
	},
}
