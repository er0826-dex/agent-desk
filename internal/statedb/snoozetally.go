package statedb

import (
	"encoding/json"
	"sort"
)

// The snooze tally counts which picker rows get used (by label, with "typed"
// for a hand-entered time), so the suggestion ladder can be reshaped from
// evidence. One JSON object in the metadata table.

const snoozeTallyKey = "snooze_tally"

// SnoozeTally returns pick counts by label.
func (s *StateDB) SnoozeTally() (map[string]int, error) {
	raw, err := s.GetMeta(snoozeTallyKey)
	if err != nil || raw == "" {
		return map[string]int{}, err
	}
	var m map[string]int
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return map[string]int{}, nil // a damaged blob starts over rather than blocking picks
	}
	return m, nil
}

// BumpSnoozeTally adds one pick for the label.
func (s *StateDB) BumpSnoozeTally(label string) error {
	m, err := s.SnoozeTally()
	if err != nil {
		return err
	}
	m[label]++
	raw, _ := json.Marshal(m)
	return s.SetMeta(snoozeTallyKey, string(raw))
}

// TallyRow is one line of the sorted tally.
type TallyRow struct {
	Label string
	Count int
}

// SortedTally orders a tally by count, highest first, then label.
func SortedTally(m map[string]int) []TallyRow {
	rows := make([]TallyRow, 0, len(m))
	for k, v := range m {
		rows = append(rows, TallyRow{k, v})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Count != rows[j].Count {
			return rows[i].Count > rows[j].Count
		}
		return rows[i].Label < rows[j].Label
	})
	return rows
}
