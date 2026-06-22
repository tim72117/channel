// Package model 定義 API 與資料層共用的資料結構。
// JSON 欄位對齊 docs/API.md,讓 iOS App 的 Codable 模型可直接解析。
package model

import "time"

type Channel struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	OwnerID            string    `json:"ownerID"`
	MemberCount        int       `json:"memberCount"`
	LastMessagePreview *string   `json:"lastMessagePreview"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

type Message struct {
	ID         string    `json:"id"`
	ChannelID  string    `json:"channelID"`
	AuthorID   string    `json:"authorID"`
	AuthorName string    `json:"authorName"`
	Text       string    `json:"text"`
	Category   *string   `json:"category"`
	Tags       []string  `json:"tags"`
	Summary    *string   `json:"summary"`
	CreatedAt  time.Time `json:"createdAt"`
}

// User 是公開身分(成員列表、訊息作者等到處可見),不含私密資料。
type User struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	AvatarColor string `json:"avatarColor"`
}

// Profile 是使用者的私密資料,只在「自己的帳號」端點回傳。
type Profile struct {
	Email string `json:"email"`
}

// Me 代表登入後的自己:公開身分(user)+ 私密資料(profile)。
// /me、login、register、apple 回傳此結構。
type Me struct {
	User    User    `json:"user"`
	Profile Profile `json:"profile"`
}

// SearchAnswer 對應語意查詢回應。
type SearchAnswer struct {
	Answer          string   `json:"answer"`
	CitedMessageIDs []string `json:"citedMessageIDs"`
	Confidence      *float64 `json:"confidence,omitempty"`
}

// Entry 是 LLM(record_entry 工具)從訊息解析出的日期/事件條目,關聯到觸發的訊息。
type Entry struct {
	ID        string    `json:"id"`
	MessageID string    `json:"messageID"` // 觸發此條目的訊息
	ChannelID string    `json:"channelID"`
	Item      string    `json:"item"`             // 事項描述
	Start     string    `json:"start"`            // 'YYYY-MM-DD HH:MM' 或全日 'YYYY-MM-DD';可空
	End       string    `json:"end,omitempty"`    // 範圍結束;可空
	AllDay    bool      `json:"allDay"`           // 全日事件
	CreatedAt time.Time `json:"createdAt"`
}
