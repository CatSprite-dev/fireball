package ledger

import (
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

type Kind string

const (
	KindDeposit      Kind = "deposit"      // пополнение
	KindWithdrawal   Kind = "withdrawal"   // вывод
	KindTransferIn   Kind = "transfer_in"  // перевод со своего счёта
	KindTransferOut  Kind = "transfer_out" // перевод на свой счёт
	KindBuy          Kind = "buy"
	KindSell         Kind = "sell"
	KindDividend     Kind = "dividend"
	KindCoupon       Kind = "coupon"
	KindRedemption   Kind = "redemption"   // погашение облигации
	KindAmortization Kind = "amortization" // частичное погашение номинала облигации
	KindFee          Kind = "fee"          // любая комиссия
	KindTax          Kind = "tax"          // любой налог, включая коррекцию
	KindDerivatives  Kind = "derivatives"  // всё по срочному рынку
)

type InstrumentKind string

const (
	InstrumentShare    InstrumentKind = "share"
	InstrumentBond     InstrumentKind = "bond"
	InstrumentETF      InstrumentKind = "etf"
	InstrumentCurrency InstrumentKind = "currency"
	InstrumentFutures  InstrumentKind = "futures"
	InstrumentOption   InstrumentKind = "option"
	InstrumentOther    InstrumentKind = "other"
)

// Event — одна операция брокера
type Event struct {
	AccountID       string
	OperationID     string
	ParentID        string
	Date            time.Time
	Kind            Kind
	InstrumentUID   string
	InstrumentKind  InstrumentKind
	Quantity        decimal.Decimal
	Amount          money.Money
	BaseAmount      money.Money
	AccruedInterest money.Money
}
