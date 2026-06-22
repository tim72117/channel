package store

import (
	"github.com/channel/server/internal/model"
	"gorm.io/gorm"
)

// toMessage 把 entity 轉成 API DTO。
func toMessage(r messageRow) model.Message {
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	return model.Message{
		ID:         r.ID,
		ChannelID:  r.ChannelID,
		AuthorID:   r.AuthorID,
		AuthorName: r.AuthorName,
		Text:       r.Text,
		Category:   r.Category,
		Tags:       tags,
		Summary:    r.Summary,
		CreatedAt:  r.CreatedAt,
	}
}

// ListMessages 回傳頻道全部訊息,依時間舊到新(LLM 查詢用)。
func (s *Store) ListMessages(channelID string) ([]model.Message, error) {
	var rows []messageRow
	err := s.db.Where("channel_id = ?", channelID).
		Order("created_at ASC").Find(&rows).Error
	return mapMessages(rows), err
}

// ListMessagesByAuthor 只回傳指定作者在該頻道發的訊息(聊天畫面用)。
func (s *Store) ListMessagesByAuthor(channelID, authorID string) ([]model.Message, error) {
	var rows []messageRow
	err := s.db.Where("channel_id = ? AND author_id = ?", channelID, authorID).
		Order("created_at ASC").Find(&rows).Error
	return mapMessages(rows), err
}

func mapMessages(rows []messageRow) []model.Message {
	out := make([]model.Message, 0, len(rows))
	for _, r := range rows {
		out = append(out, toMessage(r))
	}
	return out
}

// InsertMessage 將訊息(含 LLM 標注結果)寫入資料庫,並更新頻道時間。
func (s *Store) InsertMessage(m model.Message) error {
	r := messageRow{
		ID:         m.ID,
		ChannelID:  m.ChannelID,
		AuthorID:   m.AuthorID,
		AuthorName: m.AuthorName,
		Text:       m.Text,
		Category:   m.Category,
		Tags:       m.Tags,
		Summary:    m.Summary,
		CreatedAt:  m.CreatedAt,
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&r).Error; err != nil {
			return err
		}
		return tx.Model(&channelRow{}).Where("id = ?", m.ChannelID).
			Update("updated_at", m.CreatedAt).Error
	})
}
