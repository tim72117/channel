package store

import "time"

// 以下 entity 是 GORM 的資料表映射(帶 gorm tag),與 API DTO(model.*)分離。
// store 方法負責 entity <-> model 的轉換。

type userRow struct {
	ID           string  `gorm:"primaryKey;column:id"`
	Name         string  `gorm:"column:name;not null"`
	AvatarColor  string  `gorm:"column:avatar_color;not null"`
	AppleSub     *string `gorm:"column:apple_sub;uniqueIndex"` // 可為 NULL
	Email        *string `gorm:"column:email;uniqueIndex"`     // 可為 NULL
	PasswordHash *string `gorm:"column:password_hash"`         // 可為 NULL

	// 多對多:此使用者參與的頻道(透過 members 中介表)。
	Channels []channelRow `gorm:"many2many:members;joinForeignKey:user_id;joinReferences:channel_id"`
}

func (userRow) TableName() string { return "users" }

type channelRow struct {
	ID        string    `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name;not null"`
	OwnerID   string    `gorm:"column:owner_id;not null;default:''"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	// Has Many:頻道的訊息(刪頻道時級聯刪訊息)。
	Messages []messageRow `gorm:"foreignKey:ChannelID;constraint:OnDelete:CASCADE"`
	// 多對多:頻道成員(透過 members 中介表)。
	Members []userRow `gorm:"many2many:members;joinForeignKey:channel_id;joinReferences:user_id"`
}

func (channelRow) TableName() string { return "channels" }

type messageRow struct {
	ID         string    `gorm:"primaryKey;column:id"`
	ChannelID  string    `gorm:"column:channel_id;not null;index"`
	AuthorID   string    `gorm:"column:author_id;not null"`
	AuthorName string    `gorm:"column:author_name;not null"`
	Text       string    `gorm:"column:text;not null"`
	Category   *string   `gorm:"column:category"`
	Tags       []string  `gorm:"column:tags;serializer:json"` // JSON 陣列存單一 TEXT 欄位
	Summary    *string   `gorm:"column:summary"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`

	// Has Many:此訊息衍生的條目(刪訊息級聯刪條目)。
	Entries []entryRow `gorm:"foreignKey:MessageID;constraint:OnDelete:CASCADE"`
}

func (messageRow) TableName() string { return "messages" }

// entryRow 是 LLM 解析出的日期/事件條目,關聯到觸發的訊息(刪訊息級聯刪條目)。
type entryRow struct {
	ID        string    `gorm:"primaryKey;column:id"`
	MessageID string    `gorm:"column:message_id;not null;index"`
	ChannelID string    `gorm:"column:channel_id;not null;index"`
	Item      string    `gorm:"column:item;not null"`
	Start     string    `gorm:"column:start"`
	End       string    `gorm:"column:end_at"` // end 是 SQL 保留字,欄位改名 end_at
	AllDay    bool      `gorm:"column:all_day"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (entryRow) TableName() string { return "entries" }
