package store

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/channel/server/internal/model"
	"gorm.io/gorm"
)

// tripIDPrefix + 隨機 hex = trip ID(對齊 ch_/ent_ 風格)。
func newTripID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "trip_" + hex.EncodeToString(b)
}

func toTrip(r tripRow) model.Trip {
	return model.Trip{
		ID:        r.ID,
		ChannelID: r.ChannelID,
		Title:     r.Title,
		Start:     r.Start,
		End:       r.End,
		CreatedAt: r.CreatedAt,
	}
}

// InsertTrip 寫入一筆行程。
func (s *Store) InsertTrip(t model.Trip) error {
	r := tripRow{
		ID:        t.ID,
		ChannelID: t.ChannelID,
		Title:     t.Title,
		Start:     t.Start,
		End:       t.End,
		CreatedAt: t.CreatedAt,
	}
	return s.db.Create(&r).Error
}

// ListTripsByChannel 回傳頻道所有行程,依開始時間排序(字典序即時間序)。
func (s *Store) ListTripsByChannel(channelID string) ([]model.Trip, error) {
	var rows []tripRow
	err := s.db.Where("channel_id = ?", channelID).
		Order("start ASC, created_at ASC").Find(&rows).Error
	out := make([]model.Trip, 0, len(rows))
	for _, r := range rows {
		out = append(out, toTrip(r))
	}
	return out, err
}

// ListEntriesByTrip 回傳某行程的 entries,依開始時間排序。
func (s *Store) ListEntriesByTrip(channelID, tripID string) ([]model.Entry, error) {
	var rows []entryRow
	err := s.db.Where("channel_id = ? AND trip_id = ?", channelID, tripID).
		Order("start ASC, created_at ASC").Find(&rows).Error
	return mapEntries(rows), err
}

// SetEntryTrip 設定某 entry 的所屬行程(供重組/誤組修正,可傳 nil 解除歸組)。
func (s *Store) SetEntryTrip(entryID string, tripID *string) error {
	return s.db.Model(&entryRow{}).Where("id = ?", entryID).
		Update("trip_id", tripID).Error
}

// FindOrCreateTrip 是歸組核心:依時間把新 entry 歸入現有行程或新建。
//
// 歸組邏輯(以「區間事件為骨架」):
//   - entryStart 為空(無時間)→ 不歸組,回 (nil, nil)。
//   - 掃頻道現有 trips,若新 entry 的時間區間 [entryStart, entryEnd] 與某 trip 的
//     [trip.Start, trip.End] 重疊 → 歸入該 trip,並擴張 trip 範圍(取聯集)。
//     重疊判定即「有跨度的住宿/出差等事件框出的行程範圍」涵蓋了單點事件。
//   - 無命中 → 新建 trip(以此 entry 的起訖與 item 當初值)。
//
// start/end 以 ISO 字串儲存(字典序即時間序),故區間比較用字串。
// 全程用交易包起,避免併發重複建 trip。
func (s *Store) FindOrCreateTrip(channelID, entryStart, entryEnd, item string) (*string, error) {
	if entryStart == "" {
		return nil, nil // 無時間 entry 不歸組
	}
	// 單點事件 end 為空時,以 start 當訖點。
	eEnd := entryEnd
	if eEnd == "" {
		eEnd = entryStart
	}

	var tripID *string
	err := s.db.Transaction(func(tx *gorm.DB) error {
		var trips []tripRow
		if err := tx.Where("channel_id = ?", channelID).
			Order("start ASC, created_at ASC").Find(&trips).Error; err != nil {
			return err
		}

		// 找第一個時間區間重疊的 trip。
		// 重疊條件:trip.Start <= entryEnd 且 trip.End >= entryStart。
		// trip 的 End 可能為空(單點 trip),空時以 Start 當訖點比較。
		for i := range trips {
			tEnd := trips[i].End
			if tEnd == "" {
				tEnd = trips[i].Start
			}
			if trips[i].Start != "" && trips[i].Start <= eEnd && tEnd >= entryStart {
				// 命中:歸入並擴張 trip 範圍(取 min(start) / max(end))。
				newStart := trips[i].Start
				if entryStart < newStart {
					newStart = entryStart
				}
				newEnd := tEnd
				if eEnd > newEnd {
					newEnd = eEnd
				}
				if newStart != trips[i].Start || newEnd != trips[i].End {
					if err := tx.Model(&tripRow{}).Where("id = ?", trips[i].ID).
						Updates(map[string]interface{}{"start": newStart, "end_at": newEnd}).Error; err != nil {
						return err
					}
				}
				id := trips[i].ID
				tripID = &id
				return nil
			}
		}

		// 無命中:新建 trip。
		id := newTripID()
		nt := tripRow{
			ID:        id,
			ChannelID: channelID,
			Title:     item,
			Start:     entryStart,
			End:       eEnd,
			CreatedAt: now(),
		}
		if err := tx.Create(&nt).Error; err != nil {
			return err
		}
		tripID = &id
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tripID, nil
}
