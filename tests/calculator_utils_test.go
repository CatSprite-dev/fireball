package tests

import (
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
}

func TestSubstactMoneyValue(t *testing.T) {
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
}

func TestMultiplyMoneyValue(t *testing.T) {
	a := domain.MoneyValue{Currency: "RUB", Units: "10", Nano: 500000000}        // 10.5 RUB
	b := domain.MoneyValue{Currency: "RUB", Units: "2", Nano: 250000000}         // 2.25
	expected := domain.MoneyValue{Currency: "RUB", Units: "23", Nano: 625000000} // 23.625 RUB
	result := service.MultiplyMoneyValue(a, b)
	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("MultiplyMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "10", Nano: 000000000}        // 10 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "3", Nano: 500000000}         // 3.5
	expected2 := domain.MoneyValue{Currency: "USD", Units: "35", Nano: 000000000} // 35 USD
	result = service.MultiplyMoneyValue(a2, b2)
	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("MultiplyMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}
}

func TestDivideMoneyValue(t *testing.T) {
	a := domain.MoneyValue{Currency: "RUB", Units: "7", Nano: 500000000}        // 7.5 RUB
	b := domain.MoneyValue{Currency: "RUB", Units: "2", Nano: 500000000}        // 2.5
	expected := domain.MoneyValue{Currency: "RUB", Units: "3", Nano: 000000000} // 3 RUB
	result := service.DivideMoneyValue(a, b)
	if result.Currency != expected.Currency || result.Units != expected.Units || result.Nano != expected.Nano {
		t.Errorf("DivideMoneyValue(%v, %v) = %v; expected %v", a, b, result, expected)
	}

	a2 := domain.MoneyValue{Currency: "USD", Units: "10", Nano: 000000000}       // 10 USD
	b2 := domain.MoneyValue{Currency: "USD", Units: "4", Nano: 000000000}        // 4 USD
	expected2 := domain.MoneyValue{Currency: "USD", Units: "2", Nano: 500000000} // 2.5 USD
	result = service.DivideMoneyValue(a2, b2)
	if result.Currency != expected2.Currency || result.Units != expected2.Units || result.Nano != expected2.Nano {
		t.Errorf("DivideMoneyValue(%v, %v) = %v; expected %v", a2, b2, result, expected2)
	}
}
