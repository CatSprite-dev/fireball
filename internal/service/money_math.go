package service

import (
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/domain"
)

func AddMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalNano := int64(a.Nano) + int64(b.Nano)
	totalUnits := unitsA + unitsB + totalNano/1_000_000_000
	totalNano %= 1_000_000_000

	return domain.MoneyValue{
		Currency: a.Currency, // валюта одинаковая предполагается
		Units:    strconv.FormatInt(totalUnits, 10),
		Nano:     int(totalNano),
	}
}

func SubtractMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalNanoA := unitsA*1_000_000_000 + int64(a.Nano)
	totalNanoB := unitsB*1_000_000_000 + int64(b.Nano)

	resultNano := totalNanoA - totalNanoB

	units := resultNano / 1_000_000_000
	nano := int(resultNano % 1_000_000_000)

	if nano < 0 {
		units -= 1
		nano += 1_000_000_000
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    strconv.FormatInt(units, 10),
		Nano:     nano,
	}
}

func MultiplyMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	valA := float64(unitsA) + float64(a.Nano)/1_000_000_000
	valB := float64(unitsB) + float64(b.Nano)/1_000_000_000

	result := valA * valB

	units := int64(result)
	nano := int((result-float64(units))*1_000_000_000 + 0.5)

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    strconv.FormatInt(units, 10),
		Nano:     nano,
	}
}

func DivideMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	valA := float64(unitsA) + float64(a.Nano)/1_000_000_000
	valB := float64(unitsB) + float64(b.Nano)/1_000_000_000

	if valB == 0 {
		return domain.MoneyValue{
			Currency: a.Currency,
			Units:    "0",
			Nano:     0,
		}
	}

	result := valA / valB

	units := int64(result)
	nano := int((result-float64(units))*1_000_000_000 + 0.5)

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    strconv.FormatInt(units, 10),
		Nano:     nano,
	}
}

func AddQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalUnits := unitsA + unitsB + int64((a.Nano+b.Nano)/1_000_000_000)
	totalNano := (a.Nano + b.Nano) % 1_000_000_000

	return domain.Quotation{
		Units: strconv.FormatInt(totalUnits, 10),
		Nano:  totalNano,
	}
}

func SubstractQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalUnits := unitsA - unitsB + int64((a.Nano-b.Nano)/1_000_000_000)
	totalNano := (a.Nano - b.Nano) % 1_000_000_000

	return domain.Quotation{
		Units: strconv.FormatInt(totalUnits, 10),
		Nano:  totalNano,
	}
}
