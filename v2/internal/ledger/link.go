package ledger

// InheritInstruments заполняет InstrumentUID у событий без бумаги по их родительской операции.
func InheritInstruments(events []Event) {
	instrumentOf := make(map[string]string, len(events))
	for _, e := range events {
		if e.InstrumentUID != "" {
			instrumentOf[e.OperationID] = e.InstrumentUID
		}
	}
	for i := range events {
		if events[i].InstrumentUID != "" || events[i].ParentID == "" {
			continue
		}
		events[i].InstrumentUID = instrumentOf[events[i].ParentID]
	}
}
