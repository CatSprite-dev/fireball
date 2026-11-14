package ledger

// currentAssetUIDByOldAssetUID — конвертации, которые брокер не связывает общим признаком.
var currentAssetUIDByOldAssetUID = map[string]string{
	// Расписки TCS Group (US87238U2033) → акции МКПАО «Т-Технологии» (T).
	"4ac39e69-fff7-406c-8324-a869d97ec022": "4393bdb6-9017-466e-a3bb-ddb55d24159f",
}
