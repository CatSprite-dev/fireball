package service

import (
	"fmt"
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/shopspring/decimal"
)

func AddMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	decA := decimal.NewFromInt(unitsA*1_000_000_000 + int64(a.Nano))
	decB := decimal.NewFromInt(unitsB*1_000_000_000 + int64(b.Nano))

	result := decA.Add(decB)

	units := result.Div(decimal.NewFromInt(1_000_000_000)).IntPart()
	nano := result.Mod(decimal.NewFromInt(1_000_000_000)).IntPart()

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0" // Tinkoff API требует "-0" для отрицательных чисел меньше 1
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    unitsStr,
		Nano:     int(nano),
	}
}

func SubtractMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	decA := decimal.NewFromInt(unitsA*1_000_000_000 + int64(a.Nano))
	decB := decimal.NewFromInt(unitsB*1_000_000_000 + int64(b.Nano))

	result := decA.Sub(decB)

	units := result.Div(decimal.NewFromInt(1_000_000_000)).IntPart()
	nano := result.Mod(decimal.NewFromInt(1_000_000_000)).IntPart()

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0" // Tinkoff API требует "-0" для отрицательных чисел меньше 1
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    unitsStr,
		Nano:     int(nano),
	}
}

func MultiplyMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	aRub := decimal.NewFromInt(unitsA).Add(decimal.NewFromInt(int64(a.Nano)).Div(decimal.NewFromInt(1e9)))
	bRub := decimal.NewFromInt(unitsB).Add(decimal.NewFromInt(int64(b.Nano)).Div(decimal.NewFromInt(1e9)))

	result := aRub.Mul(bRub)

	units := result.IntPart()
	nano := result.Sub(result.Floor()).Mul(decimal.NewFromInt(1e9)).IntPart()

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    strconv.FormatInt(units, 10),
		Nano:     int(nano),
	}
}

func DivideMoneyValue(a, b domain.MoneyValue) (domain.MoneyValue, error) {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	aRub := decimal.NewFromInt(unitsA).Add(decimal.NewFromInt(int64(a.Nano)).Div(decimal.NewFromInt(1e9)))
	bRub := decimal.NewFromInt(unitsB).Add(decimal.NewFromInt(int64(b.Nano)).Div(decimal.NewFromInt(1e9)))

	if bRub.IsZero() {
		return domain.MoneyValue{}, fmt.Errorf("division by zero")
	}

	result := aRub.Div(bRub)

	var units int64
	var nano int64

	if result.IsNegative() {
		// Округляем вверх для отрицательных
		units = result.Ceil().IntPart()
		nano = result.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()
	} else {
		// Округляем вниз для положительных
		units = result.Floor().IntPart()
		nano = result.Sub(result.Floor()).Mul(decimal.NewFromInt(1e9)).IntPart()
	}

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0"
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    unitsStr,
		Nano:     int(nano),
	}, nil
}

func AddQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	decA := decimal.NewFromInt(unitsA*1_000_000_000 + int64(a.Nano))
	decB := decimal.NewFromInt(unitsB*1_000_000_000 + int64(b.Nano))

	result := decA.Add(decB)

	units := result.Div(decimal.NewFromInt(1_000_000_000)).IntPart()
	nano := result.Mod(decimal.NewFromInt(1_000_000_000)).IntPart()

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0" // Tinkoff API требует "-0" для отрицательных чисел меньше 1
	}

	return domain.Quotation{
		Units: unitsStr,
		Nano:  int(nano),
	}
}

func SubtractQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	decA := decimal.NewFromInt(unitsA*1_000_000_000 + int64(a.Nano))
	decB := decimal.NewFromInt(unitsB*1_000_000_000 + int64(b.Nano))

	result := decA.Sub(decB)

	units := result.Div(decimal.NewFromInt(1_000_000_000)).IntPart()
	nano := result.Mod(decimal.NewFromInt(1_000_000_000)).IntPart()

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0" // Tinkoff API требует "-0" для отрицательных чисел меньше 1
	}

	return domain.Quotation{
		Units: unitsStr,
		Nano:  int(nano),
	}
}

func MultiplyQuotation(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	aRub := decimal.NewFromInt(unitsA).Add(decimal.NewFromInt(int64(a.Nano)).Div(decimal.NewFromInt(1e9)))
	bRub := decimal.NewFromInt(unitsB).Add(decimal.NewFromInt(int64(b.Nano)).Div(decimal.NewFromInt(1e9)))

	result := aRub.Mul(bRub)

	units := result.IntPart()
	nano := result.Sub(result.Floor()).Mul(decimal.NewFromInt(1e9)).IntPart()

	return domain.Quotation{
		Units: strconv.FormatInt(units, 10),
		Nano:  int(nano),
	}
}

func DivideQuotation(a, b domain.Quotation) (domain.Quotation, error) {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	aRub := decimal.NewFromInt(unitsA).Add(decimal.NewFromInt(int64(a.Nano)).Div(decimal.NewFromInt(1e9)))
	bRub := decimal.NewFromInt(unitsB).Add(decimal.NewFromInt(int64(b.Nano)).Div(decimal.NewFromInt(1e9)))

	if bRub.IsZero() {
		return domain.Quotation{}, fmt.Errorf("division by zero")
	}

	result := aRub.Div(bRub)

	var units int64
	var nano int64

	if result.IsNegative() {
		// Округляем вверх для отрицательных
		units = result.Ceil().IntPart()
		nano = result.Sub(decimal.NewFromInt(units)).Mul(decimal.NewFromInt(1e9)).IntPart()
	} else {
		// Округляем вниз для положительных
		units = result.Floor().IntPart()
		nano = result.Sub(result.Floor()).Mul(decimal.NewFromInt(1e9)).IntPart()
	}

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && nano < 0 {
		unitsStr = "-0"
	}

	return domain.Quotation{
		Units: unitsStr,
		Nano:  int(nano),
	}, nil
}
