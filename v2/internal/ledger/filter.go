package ledger

import "slices"

// Filter возвращает события только выбранных счетов.
// Исходный срез не меняется, порядок событий сохраняется.
func Filter(events []Event, accountIDs []string) []Event {
	selected := make(map[string]bool, len(accountIDs))
	for _, id := range accountIDs {
		selected[id] = true
	}

	var result []Event
	for _, e := range events {
		if selected[e.AccountID] {
			result = append(result, e)
		}
	}
	return result
}

// SortByDate упорядочивает события по дате, от старых к новым.
// События с одинаковой датой сохраняют исходный порядок.
func SortByDate(events []Event) {
	slices.SortStableFunc(events, func(a, b Event) int {
		return a.Date.Compare(b.Date)
	})
}
