package service

import (
	"fmt"
	"strconv"

	"github.com/CatSprite-dev/fireball/internal/domain"
)

func AddMoneyValue(a, b domain.MoneyValue) domain.MoneyValue {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalNanoA := unitsA*1_000_000_000 + int64(a.Nano)
	totalNanoB := unitsB*1_000_000_000 + int64(b.Nano)

	resultNano := totalNanoA + totalNanoB

	units := resultNano / 1_000_000_000
	nano := int(resultNano % 1_000_000_000)

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && resultNano < 0 {
		unitsStr = "-0"
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    unitsStr,
		Nano:     nano,
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

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && resultNano < 0 {
		unitsStr = "-0"
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    unitsStr,
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

func DivideMoneyValue(a, b domain.MoneyValue) (domain.MoneyValue, error) {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	valA := float64(unitsA) + float64(a.Nano)/1_000_000_000
	valB := float64(unitsB) + float64(b.Nano)/1_000_000_000

	if valB == 0 {
		return domain.MoneyValue{}, fmt.Errorf("division by zero")
	}

	result := valA / valB

	units := int64(result)
	nano := int((result-float64(units))*1_000_000_000 + 0.5)
	if result < 0 {
		nano = int((result-float64(units))*1_000_000_000 - 0.5)
	}

	return domain.MoneyValue{
		Currency: a.Currency,
		Units:    strconv.FormatInt(units, 10),
		Nano:     nano,
	}, nil
}

func AddQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalNanoA := unitsA*1_000_000_000 + int64(a.Nano)
	totalNanoB := unitsB*1_000_000_000 + int64(b.Nano)

	resultNano := totalNanoA + totalNanoB

	units := resultNano / 1_000_000_000
	nano := int(resultNano % 1_000_000_000)

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && resultNano < 0 {
		unitsStr = "-0"
	}

	return domain.Quotation{
		Units: unitsStr,
		Nano:  nano,
	}
}

func SubtractQuotations(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	totalNanoA := unitsA*1_000_000_000 + int64(a.Nano)
	totalNanoB := unitsB*1_000_000_000 + int64(b.Nano)

	resultNano := totalNanoA - totalNanoB

	units := resultNano / 1_000_000_000
	nano := int(resultNano % 1_000_000_000)

	unitsStr := strconv.FormatInt(units, 10)
	if units == 0 && resultNano < 0 {
		unitsStr = "-0"
	}

	return domain.Quotation{
		Units: unitsStr,
		Nano:  nano,
	}
}

func MultiplyQuotation(a, b domain.Quotation) domain.Quotation {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	valA := float64(unitsA) + float64(a.Nano)/1_000_000_000
	valB := float64(unitsB) + float64(b.Nano)/1_000_000_000

	result := valA * valB

	units := int64(result)
	nano := int((result-float64(units))*1_000_000_000 + 0.5)

	return domain.Quotation{
		Units: strconv.FormatInt(units, 10),
		Nano:  nano,
	}
}

func DivideQuotation(a, b domain.Quotation) (domain.Quotation, error) {
	unitsA, _ := strconv.ParseInt(a.Units, 10, 64)
	unitsB, _ := strconv.ParseInt(b.Units, 10, 64)

	valA := float64(unitsA) + float64(a.Nano)/1_000_000_000
	valB := float64(unitsB) + float64(b.Nano)/1_000_000_000

	if valB == 0 {
		return domain.Quotation{}, fmt.Errorf("division by zero")
	}

	result := valA / valB

	units := int64(result)
	nano := int((result-float64(units))*1_000_000_000 + 0.5)
	if result < 0 {
		nano = int((result-float64(units))*1_000_000_000 - 0.5)
	}

	return domain.Quotation{
		Units: strconv.FormatInt(units, 10),
		Nano:  nano,
	}, nil
}
