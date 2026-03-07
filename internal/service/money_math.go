package service

import (
	"fmt"
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/shopspring/decimal"
)

const nanoScale = int64(1_000_000_000)

var decNano = decimal.NewFromInt(nanoScale)

func parseNano(units string, nano int) decimal.Decimal {
	u, _ := strconv.ParseInt(units, 10, 64)
	return decimal.NewFromInt(u*nanoScale + int64(nano))
}

func parseDecimal(units string, nano int) decimal.Decimal {
	u, _ := strconv.ParseInt(units, 10, 64)
	return decimal.NewFromInt(u).Add(decimal.NewFromInt(int64(nano)).Div(decNano))
}

func unitsStr(units, nano int64) string {
	if units == 0 && nano < 0 {
		return "-0"
	}
	return strconv.FormatInt(units, 10)
}

func AddMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	result := parseNano(a.Units, a.Nano).Add(parseNano(b.Units, b.Nano))
	units := result.Div(decNano).IntPart()
	nano := result.Mod(decNano).IntPart()
	return domain.MoneyValue{Currency: a.Currency, Units: unitsStr(units, nano), Nano: int(nano)}
}

func SubtractMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	result := parseNano(a.Units, a.Nano).Sub(parseNano(b.Units, b.Nano))
	units := result.Div(decNano).IntPart()
	nano := result.Mod(decNano).IntPart()
	return domain.MoneyValue{Currency: a.Currency, Units: unitsStr(units, nano), Nano: int(nano)}
}

func MultiplyMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	result := parseDecimal(a.Units, a.Nano).Mul(parseDecimal(b.Units, b.Nano))
	units := result.IntPart()
	nano := result.Sub(result.Floor()).Mul(decNano).IntPart()
	return domain.MoneyValue{Currency: a.Currency, Units: strconv.FormatInt(units, 10), Nano: int(nano)}
}

func DivideMoneyValue(a, b domain.MoneyValue) (domain.MoneyValue, error) {
	divisor := parseDecimal(b.Units, b.Nano)
	if divisor.IsZero() {
		return domain.MoneyValue{}, fmt.Errorf("division by zero")
	}
	result := parseDecimal(a.Units, a.Nano).Div(divisor)

	var units, nano int64
	if result.IsNegative() {
		units = result.Ceil().IntPart()
		nano = result.Sub(decimal.NewFromInt(units)).Mul(decNano).IntPart()
	} else {
		units = result.Floor().IntPart()
		nano = result.Sub(result.Floor()).Mul(decNano).IntPart()
	}
	return domain.MoneyValue{Currency: a.Currency, Units: unitsStr(units, nano), Nano: int(nano)}, nil
}

func AddQuotations(a, b domain.Quotation) domain.Quotation {
	result := parseNano(a.Units, a.Nano).Add(parseNano(b.Units, b.Nano))
	units := result.Div(decNano).IntPart()
	nano := result.Mod(decNano).IntPart()
	return domain.Quotation{Units: unitsStr(units, nano), Nano: int(nano)}
}

func SubtractQuotations(a, b domain.Quotation) domain.Quotation {
	result := parseNano(a.Units, a.Nano).Sub(parseNano(b.Units, b.Nano))
	units := result.Div(decNano).IntPart()
	nano := result.Mod(decNano).IntPart()
	return domain.Quotation{Units: unitsStr(units, nano), Nano: int(nano)}
}

func MultiplyQuotation(a, b domain.Quotation) domain.Quotation {
	result := parseDecimal(a.Units, a.Nano).Mul(parseDecimal(b.Units, b.Nano))
	units := result.IntPart()
	nano := result.Sub(result.Floor()).Mul(decNano).IntPart()
	return domain.Quotation{Units: strconv.FormatInt(units, 10), Nano: int(nano)}
}

func DivideQuotation(a, b domain.Quotation) (domain.Quotation, error) {
	divisor := parseDecimal(b.Units, b.Nano)
	if divisor.IsZero() {
		return domain.Quotation{}, fmt.Errorf("division by zero")
	}
	result := parseDecimal(a.Units, a.Nano).Div(divisor)

	var units, nano int64
	if result.IsNegative() {
		units = result.Ceil().IntPart()
		nano = result.Sub(decimal.NewFromInt(units)).Mul(decNano).IntPart()
	} else {
		units = result.Floor().IntPart()
		nano = result.Sub(result.Floor()).Mul(decNano).IntPart()
	}
	return domain.Quotation{Units: unitsStr(units, nano), Nano: int(nano)}, nil
}
