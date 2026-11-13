package convert

import "github.com/CatSprite-dev/fireball/v2/internal/ledger"

var operationKinds = map[string]ledger.OperationKind{
	// деньги на счёт и со счёта
	"OPERATION_TYPE_INPUT":            ledger.KindDeposit,
	"OPERATION_TYPE_INPUT_SWIFT":      ledger.KindDeposit,
	"OPERATION_TYPE_INPUT_ACQUIRING":  ledger.KindDeposit,
	"OPERATION_TYPE_OUTPUT":           ledger.KindWithdrawal,
	"OPERATION_TYPE_OUTPUT_SWIFT":     ledger.KindWithdrawal,
	"OPERATION_TYPE_OUTPUT_ACQUIRING": ledger.KindWithdrawal,

	// переводы между своими счетами
	"OPERATION_TYPE_INP_MULTI": ledger.KindTransferIn,
	"OPERATION_TYPE_OUT_MULTI": ledger.KindTransferOut,

	// сделки
	"OPERATION_TYPE_BUY":         ledger.KindBuy,
	"OPERATION_TYPE_BUY_CARD":    ledger.KindBuy,
	"OPERATION_TYPE_BUY_MARGIN":  ledger.KindBuy,
	"OPERATION_TYPE_SELL":        ledger.KindSell,
	"OPERATION_TYPE_SELL_CARD":   ledger.KindSell,
	"OPERATION_TYPE_SELL_MARGIN": ledger.KindSell,

	// доходы по бумагам
	"OPERATION_TYPE_DIVIDEND":            ledger.KindDividend,
	"OPERATION_TYPE_DIV_EXT":             ledger.KindDividend,
	"OPERATION_TYPE_COUPON":              ledger.KindCoupon,
	"OPERATION_TYPE_BOND_REPAYMENT_FULL": ledger.KindRedemption,
	"OPERATION_TYPE_BOND_REPAYMENT":      ledger.KindAmortization,

	// комиссии
	"OPERATION_TYPE_BROKER_FEE":     ledger.KindFee,
	"OPERATION_TYPE_SERVICE_FEE":    ledger.KindFee,
	"OPERATION_TYPE_MARGIN_FEE":     ledger.KindFee,
	"OPERATION_TYPE_SUCCESS_FEE":    ledger.KindFee,
	"OPERATION_TYPE_TRACK_MFEE":     ledger.KindFee,
	"OPERATION_TYPE_TRACK_PFEE":     ledger.KindFee,
	"OPERATION_TYPE_CASH_FEE":       ledger.KindFee,
	"OPERATION_TYPE_OUT_FEE":        ledger.KindFee,
	"OPERATION_TYPE_OUT_STAMP_DUTY": ledger.KindFee,
	"OPERATION_TYPE_ADVICE_FEE":     ledger.KindFee,
	"OPERATION_TYPE_OTHER_FEE":      ledger.KindFee,
	"OPERATION_TYPE_OUTPUT_PENALTY": ledger.KindFee,

	// налоги, включая возвраты и коррекции
	"OPERATION_TYPE_TAX":                         ledger.KindTax,
	"OPERATION_TYPE_BOND_TAX":                    ledger.KindTax,
	"OPERATION_TYPE_DIVIDEND_TAX":                ledger.KindTax,
	"OPERATION_TYPE_BENEFIT_TAX":                 ledger.KindTax,
	"OPERATION_TYPE_TAX_CORRECTION":              ledger.KindTax,
	"OPERATION_TYPE_TAX_CORRECTION_COUPON":       ledger.KindTax,
	"OPERATION_TYPE_TAX_PROGRESSIVE":             ledger.KindTax,
	"OPERATION_TYPE_BOND_TAX_PROGRESSIVE":        ledger.KindTax,
	"OPERATION_TYPE_DIVIDEND_TAX_PROGRESSIVE":    ledger.KindTax,
	"OPERATION_TYPE_BENEFIT_TAX_PROGRESSIVE":     ledger.KindTax,
	"OPERATION_TYPE_TAX_CORRECTION_PROGRESSIVE":  ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO":                    ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO_HOLD":               ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO_REFUND":             ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO_PROGRESSIVE":        ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO_HOLD_PROGRESSIVE":   ledger.KindTax,
	"OPERATION_TYPE_TAX_REPO_REFUND_PROGRESSIVE": ledger.KindTax,

	// срочный рынок
	"OPERATION_TYPE_ACCRUING_VARMARGIN":    ledger.KindDerivatives,
	"OPERATION_TYPE_WRITING_OFF_VARMARGIN": ledger.KindDerivatives,
	"OPERATION_TYPE_OPTION_EXPIRATION":     ledger.KindDerivatives,
	"OPERATION_TYPE_FUTURE_EXPIRATION":     ledger.KindDerivatives,
}

var instrumentKinds = map[string]ledger.InstrumentKind{
	"INSTRUMENT_TYPE_SHARE":    ledger.InstrumentShare,
	"INSTRUMENT_TYPE_BOND":     ledger.InstrumentBond,
	"INSTRUMENT_TYPE_ETF":      ledger.InstrumentETF,
	"INSTRUMENT_TYPE_CURRENCY": ledger.InstrumentCurrency,
	"INSTRUMENT_TYPE_FUTURES":  ledger.InstrumentFutures,
	"INSTRUMENT_TYPE_OPTION":   ledger.InstrumentOption,
}

const instrumentUnspecified = "INSTRUMENT_TYPE_UNSPECIFIED"

func toInstrumentKind(s string) ledger.InstrumentKind {
	if s == "" || s == instrumentUnspecified {
		return ""
	}
	if kind, ok := instrumentKinds[s]; ok {
		return kind
	}
	return ledger.InstrumentOther
}

// positionKinds — тип бумаги в позиции портфеля брокера → наш тип.
// В портфеле брокер отдаёт только строку instrumentType, без instrumentKind.
var positionKinds = map[string]ledger.InstrumentKind{
	"share":    ledger.InstrumentShare,
	"bond":     ledger.InstrumentBond,
	"etf":      ledger.InstrumentETF,
	"currency": ledger.InstrumentCurrency,
	"futures":  ledger.InstrumentFutures,
	"option":   ledger.InstrumentOption,
}

func positionKind(s string) ledger.InstrumentKind {
	if s == "" {
		return ""
	}
	if kind, ok := positionKinds[s]; ok {
		return kind
	}
	return ledger.InstrumentOther
}
