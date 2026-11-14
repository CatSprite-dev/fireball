package ledger

import (
	"time"

	"github.com/CatSprite-dev/fireball/v2/internal/money"
	"github.com/shopspring/decimal"
)

type OperationKind string

const (
	KindDeposit         OperationKind = "deposit"      // пополнение
	KindWithdrawal      OperationKind = "withdrawal"   // вывод
	KindTransferIn      OperationKind = "transfer_in"  // перевод со своего счёта
	KindTransferOut     OperationKind = "transfer_out" // перевод на свой счёт
	KindBuy             OperationKind = "buy"
	KindSell            OperationKind = "sell"
	KindDividend        OperationKind = "dividend"
	KindCoupon          OperationKind = "coupon"
	KindRedemption      OperationKind = "redemption"       // погашение облигации
	KindAmortization    OperationKind = "amortization"     // частичное погашение номинала облигации
	KindFee             OperationKind = "fee"              // любая комиссия
	KindTax             OperationKind = "tax"              // любой налог, включая коррекцию
	KindVariationMargin OperationKind = "variation_margin" // начисление и списание вариационной маржи
	KindExpiration      OperationKind = "expiration"       // экспирация фьючерса или опциона
	KindExchange        OperationKind = "exchange"         // вторая нога обмена валюты
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
	OperationKind   OperationKind
	InstrumentUID   string
	AssetUID        string
	InstrumentKind  InstrumentKind
	Quantity        decimal.Decimal
	Amount          money.Money
	BaseAmount      money.Money
	AccruedInterest money.Money
}
