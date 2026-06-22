package store

import (
	"errors"
	"time"

	"github.com/channel/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ListChannelsForUser 回傳指定使用者參與(為成員)的頻道,依更新時間新到舊。
// memberCount 與 lastMessagePreview 以子查詢取得。
func (s *Store) ListChannelsForUser(userID string) ([]model.Channel, error) {
	type chanAgg struct {
		ID                 string
		Name               string
		OwnerID            string
		UpdatedAt          time.Time
		MemberCount        int
		LastMessagePreview *string
	}
	var rows []chanAgg
	err := s.db.
		Table("channels c").
		Select(`c.id, c.name, c.owner_id, c.updated_at,
			(SELECT COUNT(*) FROM members m2 WHERE m2.channel_id = c.id) AS member_count,
			(SELECT text FROM messages msg WHERE msg.channel_id = c.id
			 ORDER BY msg.created_at DESC LIMIT 1) AS last_message_preview`).
		Joins("JOIN members m ON m.channel_id = c.id AND m.user_id = ?", userID).
		Order("c.updated_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	out := make([]model.Channel, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.Channel{
			ID:                 r.ID,
			Name:               r.Name,
			OwnerID:            r.OwnerID,
			UpdatedAt:          r.UpdatedAt,
			MemberCount:        r.MemberCount,
			LastMessagePreview: r.LastMessagePreview,
		})
	}
	return out, nil
}

// CreateChannel 建立頻道,建立者即為擁有者(owner),並自動成為成員。
func (s *Store) CreateChannel(id, name string, creator model.User) (model.Channel, error) {
	t := now()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		ch := channelRow{ID: id, Name: name, OwnerID: creator.ID, CreatedAt: t, UpdatedAt: t}
		if err := tx.Create(&ch).Error; err != nil {
			return err
		}
		// 建立者加入成員(中介表)。
		return tx.Create(&memberLink{ChannelID: id, UserID: creator.ID}).Error
	})
	if err != nil {
		return model.Channel{}, err
	}
	return model.Channel{ID: id, Name: name, OwnerID: creator.ID, MemberCount: 1, UpdatedAt: t}, nil
}

// GetChannelOwner 回傳頻道的 owner_id;頻道不存在回 ErrNotFound。
func (s *Store) GetChannelOwner(channelID string) (string, error) {
	var cr channelRow
	err := s.db.Select("owner_id").Where("id = ?", channelID).First(&cr).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	return cr.OwnerID, nil
}

// CountChannels 回傳頻道總數(seed 判斷資料庫是否為空用)。
func (s *Store) CountChannels() (int, error) {
	var n int64
	err := s.db.Model(&channelRow{}).Count(&n).Error
	return int(n), err
}

// channelExists 確認頻道存在。
func (s *Store) channelExists(id string) (bool, error) {
	var n int64
	err := s.db.Model(&channelRow{}).Where("id = ?", id).Count(&n).Error
	return n > 0, err
}

// ----- 成員 -----

// ListMembers 回傳頻道成員(從 users 表撈,依名稱排序)。
func (s *Store) ListMembers(channelID string) ([]model.User, error) {
	var rows []userRow
	err := s.db.
		Joins("JOIN members m ON m.user_id = users.id").
		Where("m.channel_id = ?", channelID).
		Order("users.name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]model.User, 0, len(rows))
	for _, r := range rows {
		out = append(out, toUser(r))
	}
	return out, nil
}

// AddMember 加入成員(冪等),回傳更新後的成員清單。
func (s *Store) AddMember(channelID string, u model.User) ([]model.User, error) {
	ok, err := s.channelExists(channelID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotFound
	}
	// 冪等:衝突時忽略。
	link := memberLink{ChannelID: channelID, UserID: u.ID}
	if err := s.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&link).Error; err != nil {
		return nil, err
	}
	return s.ListMembers(channelID)
}

// ----- 使用者目錄 -----

// UpsertUser 寫入或更新一筆使用者(供 seed)。
func (s *Store) UpsertUser(u model.User) error {
	r := userRow{ID: u.ID, Name: u.Name, AvatarColor: u.AvatarColor}
	return s.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "avatar_color"}),
	}).Create(&r).Error
}

// memberLink 對應 many2many 的中介表 members(用於直接寫入/冪等)。
type memberLink struct {
	ChannelID string `gorm:"primaryKey;column:channel_id"`
	UserID    string `gorm:"primaryKey;column:user_id"`
}

func (memberLink) TableName() string { return "members" }
