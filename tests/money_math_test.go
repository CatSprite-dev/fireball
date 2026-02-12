package tests

import (
	"fmt"
	"testing"

	"github.com/CatSprite-dev/fireball/internal/domain"
	"github.com/CatSprite-dev/fireball/internal/service"
)

func TestAddMoneyValue(t *testing.T) {
	a := domain.MoneyValue{Currency: "RUB", Units: "10", Nano: 0000000000}       // 10 RUB
	b := domain.MoneyValue{Currency: "RUB", Units: "5", Nano: 250000000}         // 5.25 RUB
	expected := domain.MoneyValue{Currency: "RUB", Units: "15", Nano: 250000000} // 15.25 RUB
	result := service.AddMoneyValue(a, b)

	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "20", Nano: 500000000}        // 20.5 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "15", Nano: 750000000}        // 15.75 USD
	expected2 := domain.MoneyValue{Currency: "USD", Units: "36", Nano: 250000000} // 36.25 USD
	result = service.AddMoneyValue(a2, b2)

	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.MoneyValue{Currency: "EUR", Units: "5", Nano: 900000000}          // 5.9 EUR
	b3 := domain.MoneyValue{Currency: "EUR", Units: "-6", Nano: -200000000}        // -6.2 EUR
	expected3 := domain.MoneyValue{Currency: "EUR", Units: "-0", Nano: -300000000} // -0.3 EUR
	result = service.AddMoneyValue(a3, b3)

	if result.Currency != expected3.Currency || result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}
}

func TestSubtractMoneyValue(t *testing.T) {
	a := domain.MoneyValue{Currency: "RUB", Units: "10", Nano: 0000000000}      // 10 RUB
	b := domain.MoneyValue{Currency: "RUB", Units: "5", Nano: 250000000}        // 5.25 RUB
	expected := domain.MoneyValue{Currency: "RUB", Units: "4", Nano: 750000000} // 4.75 RUB
	result := service.SubtractMoneyValue(a, b)

	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "20", Nano: 500000000}        // 20.5 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "21", Nano: 500000000}        // 21.5 USD
	expected2 := domain.MoneyValue{Currency: "USD", Units: "-1", Nano: 000000000} // -1 USD
	result = service.SubtractMoneyValue(a2, b2)

	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.MoneyValue{Currency: "EUR", Units: "-5", Nano: -900000000}      // -5.9 EUR
	b3 := domain.MoneyValue{Currency: "EUR", Units: "-6", Nano: -200000000}      // -6.2 EUR
	expected3 := domain.MoneyValue{Currency: "EUR", Units: "0", Nano: 300000000} // 0.3 EUR
	result = service.SubtractMoneyValue(a3, b3)
	if result.Currency != expected3.Currency || result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}

	a4 := domain.MoneyValue{Currency: "JPY", Units: "100", Nano: 0}                // 100 JPY
	b4 := domain.MoneyValue{Currency: "JPY", Units: "-101", Nano: -500000000}      // -101.5 JPY
	expected4 := domain.MoneyValue{Currency: "JPY", Units: "201", Nano: 500000000} // 201.5 JPY
	result = service.SubtractMoneyValue(a4, b4)
	if result.Currency != expected4.Currency || result.Units != expected4.Units || result.Nano != expected4.Nano {
		t.Errorf("AddMoneyValue(%v, %v) = %v; expected %v", a4, b4, result, expected4)
	}
}

func TestMultiplyMoneyValue(t *testing.T) {
	b := domain.MoneyValue{Currency: "RUB", Units: "2", Nano: 250000000}         // 2.25
	expected := domain.MoneyValue{Currency: "RUB", Units: "23", Nano: 625000000} // 23.625 RUB
	a := domain.MoneyValue{Currency: "RUB", Units: "10", Nano: 500000000}        // 10.5 RUB
	result := service.MultiplyMoneyValue(a, b)
	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("MultiplyMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "10", Nano: 000000000}         // 10 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "-3", Nano: -500000000}        // -3.5
	expected2 := domain.MoneyValue{Currency: "USD", Units: "-35", Nano: 000000000} // -35 USD
	result = service.MultiplyMoneyValue(a2, b2)
	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("MultiplyMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.MoneyValue{Currency: "EUR", Units: "-5", Nano: -900000000}       // -5.9 EUR
	b3 := domain.MoneyValue{Currency: "EUR", Units: "-6", Nano: -200000000}       // -6.2 EUR
	expected3 := domain.MoneyValue{Currency: "EUR", Units: "36", Nano: 580000000} // 36.58 EUR
	result = service.MultiplyMoneyValue(a3, b3)
	if result.Currency != expected3.Currency || result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("MultiplyMoneyValue(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}
}

func TestDivideMoneyValue(t *testing.T) {
	a := domain.MoneyValue{Currency: "RUB", Units: "7", Nano: 500000000}        // 7.5 RUB
	b := domain.MoneyValue{Currency: "RUB", Units: "2", Nano: 500000000}        // 2.5
	expected := domain.MoneyValue{Currency: "RUB", Units: "3", Nano: 000000000} // 3 RUB
	result, _ := service.DivideMoneyValue(a, b)
	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("DivideMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "10", Nano: 000000000}         // 10 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "-4", Nano: 000000000}         // -4 USD
	expected2 := domain.MoneyValue{Currency: "USD", Units: "-2", Nano: -500000000} // -2.5 USD
	result, _ = service.DivideMoneyValue(a2, b2)
	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("DivideMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.MoneyValue{Currency: "EUR", Units: "-5", Nano: -900000000} // -5.9 EUR
	b3 := domain.MoneyValue{Currency: "EUR", Units: "0", Nano: 000000000}   // 0 EUR
	expectedErr := fmt.Errorf("division by zero")
	_, err := service.DivideMoneyValue(a3, b3)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("DivideMoneyValue(%v, %v) expected error %v; got %v", a3, b3, expectedErr, err)
	}
}

func TestAddQuotations(t *testing.T) {
	a := domain.Quotation{Units: "2", Nano: 500000000}        // 2.5
	b := domain.Quotation{Units: "3", Nano: 250000000}        // 3.25
	expected := domain.Quotation{Units: "5", Nano: 750000000} // 5.75
	result := service.AddQuotations(a, b)
	if result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("AddQuotations(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.Quotation{Units: "1", Nano: 750000000}          // 1.75
	b2 := domain.Quotation{Units: "-2", Nano: -500000000}        // -2.5
	expected2 := domain.Quotation{Units: "-0", Nano: -750000000} // -0.75
	result = service.AddQuotations(a2, b2)
	if result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("AddQuotations(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.Quotation{Units: "-3", Nano: -250000000}       // -3.25
	b3 := domain.Quotation{Units: "-1", Nano: -750000000}       // -1.75
	expected3 := domain.Quotation{Units: "-5", Nano: 000000000} // -5
	result = service.AddQuotations(a3, b3)
	if result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("AddQuotations(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}

	a4 := domain.Quotation{Units: "1", Nano: 500000000}        // 1.5
	b4 := domain.Quotation{Units: "-1", Nano: -500000000}      // -1.5
	expected4 := domain.Quotation{Units: "0", Nano: 000000000} // 0
	result = service.AddQuotations(a4, b4)
	if result.Units != expected4.Units || result.Nano != expected4.Nano {
		t.Errorf("AddQuotations(%v, %v) = %v; expected %v", a4, b4, result, expected4)
	}

	a5 := domain.Quotation{Units: "7", Nano: 500000000}        // 7.5
	b5 := domain.Quotation{Units: "-2", Nano: -250000000}      // -2.25
	expected5 := domain.Quotation{Units: "5", Nano: 250000000} // 5.25
	result = service.AddQuotations(a5, b5)
	if result.Units != expected5.Units || result.Nano != expected5.Nano {
		t.Errorf("AddQuotations(%v, %v) = %v; expected %v", a5, b5, result, expected5)
	}
}

func TestSubtractQuotations(t *testing.T) {
	a := domain.Quotation{Units: "2", Nano: 500000000}          // 2.5
	b := domain.Quotation{Units: "3", Nano: 250000000}          // 3.25
	expected := domain.Quotation{Units: "-0", Nano: -750000000} // -0.75
	result := service.SubtractQuotations(a, b)
	if result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("SubtractQuotations(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.Quotation{Units: "1", Nano: 750000000}        // 1.75
	b2 := domain.Quotation{Units: "-2", Nano: -500000000}      // -2.5
	expected2 := domain.Quotation{Units: "4", Nano: 250000000} // 4.25
	result = service.SubtractQuotations(a2, b2)
	if result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("SubtractQuotations(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.Quotation{Units: "3", Nano: 250000000}          // 3.25
	b3 := domain.Quotation{Units: "5", Nano: 750000000}          // 5.75
	expected3 := domain.Quotation{Units: "-2", Nano: -500000000} // -2.5
	result = service.SubtractQuotations(a3, b3)
	if result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("SubtractQuotations(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}

	a4 := domain.Quotation{Units: "1", Nano: 758496539}        // 1.758496539
	b4 := domain.Quotation{Units: "0", Nano: 573867463}        // 0.573867463
	expected4 := domain.Quotation{Units: "1", Nano: 184629076} // 1.184629076
	result = service.SubtractQuotations(a4, b4)
	if result.Units != expected4.Units || result.Nano != expected4.Nano {
		t.Errorf("SubtractQuotations(%v, %v) = %v; expected %v", a4, b4, result, expected4)
	}
}

func TestMultiplyQuotations(t *testing.T) {
	b := domain.Quotation{Units: "2", Nano: 250000000}         // 2.25
	expected := domain.Quotation{Units: "23", Nano: 625000000} // 23.625 RUB
	a := domain.Quotation{Units: "10", Nano: 500000000}        // 10.5 RUB
	result := service.MultiplyQuotation(a, b)
	if result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("MultiplyQuotation(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.Quotation{Units: "10", Nano: 000000000}         // 10 USD
	b2 := domain.Quotation{Units: "-3", Nano: -500000000}        // -3.5
	expected2 := domain.Quotation{Units: "-35", Nano: 000000000} // -35 USD
	result = service.MultiplyQuotation(a2, b2)
	if result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("MultiplyQuotation(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.Quotation{Units: "-5", Nano: -900000000}       // -5.9 EUR
	b3 := domain.Quotation{Units: "-6", Nano: -200000000}       // -6.2 EUR
	expected3 := domain.Quotation{Units: "36", Nano: 580000000} // 36.58 EUR
	result = service.MultiplyQuotation(a3, b3)
	if result.Units != expected3.Units || result.Nano != expected3.Nano {
		t.Errorf("MultiplyQuotation(%v, %v) = %v; expected %v", a3, b3, result, expected3)
	}

	a4 := domain.Quotation{Units: "0", Nano: 530000000}         // 0.53
	b4 := domain.Quotation{Units: "100", Nano: 0}               // 100
	expected4 := domain.Quotation{Units: "53", Nano: 000000000} // 53
	result = service.MultiplyQuotation(a4, b4)
	if result.Units != expected4.Units || result.Nano != expected4.Nano {
		t.Errorf("MultiplyQuotation(%v, %v) = %v; expected %v", a4, b4, result, expected4)
	}
}

func TestDivideQuotations(t *testing.T) {
	a := domain.Quotation{Units: "7", Nano: 500000000}        // 7.5 RUB
	b := domain.Quotation{Units: "2", Nano: 500000000}        // 2.5
	expected := domain.Quotation{Units: "3", Nano: 000000000} // 3 RUB
	result, _ := service.DivideQuotation(a, b)
	if result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("DivideQuotation(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.Quotation{Units: "10", Nano: 000000000}         // 10 USD
	b2 := domain.Quotation{Units: "-4", Nano: 000000000}         // -4 USD
	expected2 := domain.Quotation{Units: "-2", Nano: -500000000} // -2.5 USD
	result, _ = service.DivideQuotation(a2, b2)
	if result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("DivideQuotation(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}

	a3 := domain.Quotation{Units: "-5", Nano: -900000000} // -5.9 EUR
	b3 := domain.Quotation{Units: "0", Nano: 000000000}   // 0 EUR
	expectedErr := fmt.Errorf("division by zero")
	_, err := service.DivideQuotation(a3, b3)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("DivideQuotation(%v, %v) expected error %v; got %v", a3, b3, expectedErr, err)
	}
}
