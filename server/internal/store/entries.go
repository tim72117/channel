package store

import "github.com/channel/server/internal/model"

func toEntry(r entryRow) model.Entry {
	return model.Entry{
		ID:        r.ID,
		MessageID: r.MessageID,
		ChannelID: r.ChannelID,
		Item:      r.Item,
		Start:     r.Start,
		End:       r.End,
		AllDay:    r.AllDay,
		CreatedAt: r.CreatedAt,
	}
}

// InsertEntry 寫入一筆條目(關聯到觸發的訊息)。
func (s *Store) InsertEntry(e model.Entry) error {
	r := entryRow{
		ID:        e.ID,
		MessageID: e.MessageID,
		ChannelID: e.ChannelID,
		Item:      e.Item,
		Start:     e.Start,
		End:       e.End,
		AllDay:    e.AllDay,
		CreatedAt: e.CreatedAt,
	}
	return s.db.Create(&r).Error
}

// ListEntriesByChannel 回傳頻道的所有條目,依開始時間排序。
func (s *Store) ListEntriesByChannel(channelID string) ([]model.Entry, error) {
	var rows []entryRow
	err := s.db.Where("channel_id = ?", channelID).
		Order("start ASC, created_at ASC").Find(&rows).Error
	return mapEntries(rows), err
}

// ListEntriesByMessage 回傳某則訊息衍生的條目。
func (s *Store) ListEntriesByMessage(messageID string) ([]model.Entry, error) {
	var rows []entryRow
	err := s.db.Where("message_id = ?", messageID).
		Order("created_at ASC").Find(&rows).Error
	return mapEntries(rows), err
}

func mapEntries(rows []entryRow) []model.Entry {
	out := make([]model.Entry, 0, len(rows))
	for _, r := range rows {
		out = append(out, toEntry(r))
	}
	return out
}
