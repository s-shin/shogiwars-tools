package main

import "time"

type RecordID string

type GameType string

const (
	TenMinutes        GameType = "10m"
	ThreeMinutes               = "3m"
	TenSecondsPerMove          = "10s"
)

func (t GameType) ParamString() string {
	switch t {
	case TenMinutes:
		return ""
	case ThreeMinutes:
		return "sb"
	case TenSecondsPerMove:
		return "s1"
	}
	return ""
}

type Player struct {
	UserName string `json:"user_name"`
	Rank     string `json:"rank"`
}

type Players struct {
	Black Player `json:"black"`
	White Player `json:"white"`
}

type RecordItem struct {
	GameType GameType  `json:"game_type"`
	Date     time.Time `json:"date"`
	Winner   string    `json:"winner"`
	Players  Players   `json:"players"`
	Tags     []string  `json:"tags"`
	RecordID RecordID  `json:"record_id"`
}

func NewRecordItem() *RecordItem {
	return &RecordItem{
		Tags: make([]string, 0),
	}
}

type AppData struct {
	RecordItems []*RecordItem `json:"record_items"`
}

//---

type AppDataManager struct {
	*AppData
	RecordItemByID map[RecordID]*RecordItem
}

func NewAppDataManager(appData *AppData) *AppDataManager {
	return &AppDataManager{AppData: appData}
}

func (m *AppDataManager) UpdateRecordItemIndex() {
	if len(m.RecordItemByID) == len(m.AppData.RecordItems) {
		return
	}
	m.RecordItemByID = make(map[RecordID]*RecordItem, len(m.AppData.RecordItems))
	for _, item := range m.AppData.RecordItems {
		m.RecordItemByID[item.RecordID] = item
	}
}

func (m *AppDataManager) GetRecordItem(id RecordID) (*RecordItem, bool) {
	m.UpdateRecordItemIndex()
	item, ok := m.RecordItemByID[id]
	return item, ok
}

func (m *AppDataManager) AppendRecordItems(items ...*RecordItem) int {
	n := 0
	for _, item := range items {
		if _, ok := m.GetRecordItem(item.RecordID); !ok {
			n++
			m.RecordItems = append(m.RecordItems, item)
		}
	}
	return n
}
