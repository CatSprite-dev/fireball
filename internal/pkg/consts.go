package pkg

const (
	HTTPMethodGet    = "GET"
	HTTPMethodPost   = "POST"
	HTTPMethodPut    = "PUT"
	HTTPMethodDelete = "DELETE"
	HTTPMethodPatch  = "PATCH"
)

type OperationType string

const (
	OperationTypeUnspecified              OperationType = "OPERATION_TYPE_UNSPECIFIED"
	OperationTypeInput                    OperationType = "OPERATION_TYPE_INPUT"
	OperationTypeBondTax                  OperationType = "OPERATION_TYPE_BOND_TAX"
	OperationTypeOutputSecurities         OperationType = "OPERATION_TYPE_OUTPUT_SECURITIES"
	OperationTypeOvernight                OperationType = "OPERATION_TYPE_OVERNIGHT"
	OperationTypeTax                      OperationType = "OPERATION_TYPE_TAX"
	OperationTypeBondRepaymentFull        OperationType = "OPERATION_TYPE_BOND_REPAYMENT_FULL"
	OperationTypeSellCard                 OperationType = "OPERATION_TYPE_SELL_CARD"
	OperationTypeDividendTax              OperationType = "OPERATION_TYPE_DIVIDEND_TAX"
	OperationTypeOutput                   OperationType = "OPERATION_TYPE_OUTPUT"
	OperationTypeBondRepayment            OperationType = "OPERATION_TYPE_BOND_REPAYMENT"
	OperationTypeTaxCorrection            OperationType = "OPERATION_TYPE_TAX_CORRECTION"
	OperationTypeServiceFee               OperationType = "OPERATION_TYPE_SERVICE_FEE"
	OperationTypeBenefitTax               OperationType = "OPERATION_TYPE_BENEFIT_TAX"
	OperationTypeMarginFee                OperationType = "OPERATION_TYPE_MARGIN_FEE"
	OperationTypeBuy                      OperationType = "OPERATION_TYPE_BUY"
	OperationTypeBuyCard                  OperationType = "OPERATION_TYPE_BUY_CARD"
	OperationTypeInputSecurities          OperationType = "OPERATION_TYPE_INPUT_SECURITIES"
	OperationTypeSellMargin               OperationType = "OPERATION_TYPE_SELL_MARGIN"
	OperationTypeBrokerFee                OperationType = "OPERATION_TYPE_BROKER_FEE"
	OperationTypeBuyMargin                OperationType = "OPERATION_TYPE_BUY_MARGIN"
	OperationTypeDividend                 OperationType = "OPERATION_TYPE_DIVIDEND"
	OperationTypeSell                     OperationType = "OPERATION_TYPE_SELL"
	OperationTypeCoupon                   OperationType = "OPERATION_TYPE_COUPON"
	OperationTypeSuccessFee               OperationType = "OPERATION_TYPE_SUCCESS_FEE"
	OperationTypeDividendTransfer         OperationType = "OPERATION_TYPE_DIVIDEND_TRANSFER"
	OperationTypeAccruingVarmargin        OperationType = "OPERATION_TYPE_ACCRUING_VARMARGIN"
	OperationTypeWritingOffVarmargin      OperationType = "OPERATION_TYPE_WRITING_OFF_VARMARGIN"
	OperationTypeDeliveryBuy              OperationType = "OPERATION_TYPE_DELIVERY_BUY"
	OperationTypeDeliverySell             OperationType = "OPERATION_TYPE_DELIVERY_SELL"
	OperationTypeTrackMFee                OperationType = "OPERATION_TYPE_TRACK_MFEE"
	OperationTypeTrackPFee                OperationType = "OPERATION_TYPE_TRACK_PFEE"
	OperationTypeTaxProgressive           OperationType = "OPERATION_TYPE_TAX_PROGRESSIVE"
	OperationTypeBondTaxProgressive       OperationType = "OPERATION_TYPE_BOND_TAX_PROGRESSIVE"
	OperationTypeDividendTaxProgressive   OperationType = "OPERATION_TYPE_DIVIDEND_TAX_PROGRESSIVE"
	OperationTypeBenefitTaxProgressive    OperationType = "OPERATION_TYPE_BENEFIT_TAX_PROGRESSIVE"
	OperationTypeTaxCorrectionProgressive OperationType = "OPERATION_TYPE_TAX_CORRECTION_PROGRESSIVE"
	OperationTypeTaxRepoProgressive       OperationType = "OPERATION_TYPE_TAX_REPO_PROGRESSIVE"
	OperationTypeTaxRepo                  OperationType = "OPERATION_TYPE_TAX_REPO"
	OperationTypeTaxRepoHold              OperationType = "OPERATION_TYPE_TAX_REPO_HOLD"
	OperationTypeTaxRepoRefund            OperationType = "OPERATION_TYPE_TAX_REPO_REFUND"
	OperationTypeTaxRepoHoldProgressive   OperationType = "OPERATION_TYPE_TAX_REPO_HOLD_PROGRESSIVE"
	OperationTypeTaxRepoRefundProgressive OperationType = "OPERATION_TYPE_TAX_REPO_REFUND_PROGRESSIVE"
	OperationTypeDivExt                   OperationType = "OPERATION_TYPE_DIV_EXT"
	OperationTypeTaxCorrectionCoupon      OperationType = "OPERATION_TYPE_TAX_CORRECTION_COUPON"
	OperationTypeCashFee                  OperationType = "OPERATION_TYPE_CASH_FEE"
	OperationTypeOutFee                   OperationType = "OPERATION_TYPE_OUT_FEE"
	OperationTypeOutStampDuty             OperationType = "OPERATION_TYPE_OUT_STAMP_DUTY"
	OperationTypeOutputSwift              OperationType = "OPERATION_TYPE_OUTPUT_SWIFT"
	OperationTypeInputSwift               OperationType = "OPERATION_TYPE_INPUT_SWIFT"
	OperationTypeOutputAcquiring          OperationType = "OPERATION_TYPE_OUTPUT_ACQUIRING"
	OperationTypeInputAcquiring           OperationType = "OPERATION_TYPE_INPUT_ACQUIRING"
	OperationTypeOutputPenalty            OperationType = "OPERATION_TYPE_OUTPUT_PENALTY"
	OperationTypeAdviceFee                OperationType = "OPERATION_TYPE_ADVICE_FEE"
	OperationTypeTransIisBs               OperationType = "OPERATION_TYPE_TRANS_IIS_BS"
	OperationTypeTransBsBs                OperationType = "OPERATION_TYPE_TRANS_BS_BS"
	OperationTypeOutMulti                 OperationType = "OPERATION_TYPE_OUT_MULTI"
	OperationTypeInpMulti                 OperationType = "OPERATION_TYPE_INP_MULTI"
	OperationTypeOverPlacement            OperationType = "OPERATION_TYPE_OVER_PLACEMENT"
	OperationTypeOverCom                  OperationType = "OPERATION_TYPE_OVER_COM"
	OperationTypeOverIncome               OperationType = "OPERATION_TYPE_OVER_INCOME"
	OperationTypeOptionExpiration         OperationType = "OPERATION_TYPE_OPTION_EXPIRATION"
	OperationTypeFutureExpiration         OperationType = "OPERATION_TYPE_FUTURE_EXPIRATION"
	OperationTypeOtherFee                 OperationType = "OPERATION_TYPE_OTHER_FEE"
	OperationTypeOther                    OperationType = "OPERATION_TYPE_OTHER"
	OperationTypeDfaRedemption            OperationType = "OPERATION_TYPE_DFA_REDEMPTION"
	OperationTypePrimaryOrder             OperationType = "OPERATION_TYPE_PRIMARY_ORDER"
)

type OperationState string

const (
	OperationStateUnspecified OperationState = "OPERATION_STATE_UNSPECIFIED"
	OperationStateExecuted    OperationState = "OPERATION_STATE_EXECUTED"
	OperationStateCanceled    OperationState = "OPERATION_STATE_CANCELED"
	OperationStateProgress    OperationState = "OPERATION_STATE_PROGRESS"
)

type CandleInterval string

const (
	CandleIntervalUnspecified CandleInterval = "CANDLE_INTERVAL_UNSPECIFIED"
	CandleInterval1Min        CandleInterval = "CANDLE_INTERVAL_1_MIN"
	CandleInterval5Min        CandleInterval = "CANDLE_INTERVAL_5_MIN"
	CandleInterval15Min       CandleInterval = "CANDLE_INTERVAL_15_MIN"
	CandleIntervalHour        CandleInterval = "CANDLE_INTERVAL_HOUR"
	CandleIntervalDay         CandleInterval = "CANDLE_INTERVAL_DAY"
	CandleInterval2Min        CandleInterval = "CANDLE_INTERVAL_2_MIN"
	CandleInterval3Min        CandleInterval = "CANDLE_INTERVAL_3_MIN"
	CandleInterval10Min       CandleInterval = "CANDLE_INTERVAL_10_MIN"
	CandleInterval30Min       CandleInterval = "CANDLE_INTERVAL_30_MIN"
	CandleInterval2Hour       CandleInterval = "CANDLE_INTERVAL_2_HOUR"
	CandleInterval4Hour       CandleInterval = "CANDLE_INTERVAL_4_HOUR"
	CandleIntervalWeek        CandleInterval = "CANDLE_INTERVAL_WEEK"
	CandleIntervalMonth       CandleInterval = "CANDLE_INTERVAL_MONTH"
	CandleInterval5Sec        CandleInterval = "CANDLE_INTERVAL_5_SEC"
	CandleInterval10Sec       CandleInterval = "CANDLE_INTERVAL_10_SEC"
	CandleInterval30Sec       CandleInterval = "CANDLE_INTERVAL_30_SEC"
)

type CandleSource string

const (
	CandleSourceUnspecified    CandleSource = "CANDLE_SOURCE_UNSPECIFIED"
	CandleSourceExchange       CandleSource = "CANDLE_SOURCE_EXCHANGE"
	CandleSourceIncludeWeekend CandleSource = "CANDLE_SOURCE_INCLUDE_WEEKEND"
)

type AccountStatus string

const (
	AccountStatusUnspecified AccountStatus = "ACCOUNT_STATUS_UNSPECIFIED"
	AccountStatusNew         AccountStatus = "ACCOUNT_STATUS_NEW"
	AccountStatusOpen        AccountStatus = "ACCOUNT_STATUS_OPEN"
	AccountStatusClosed      AccountStatus = "ACCOUNT_STATUS_CLOSED"
	AccountStatusAll         AccountStatus = "ACCOUNT_STATUS_ALL"
)
