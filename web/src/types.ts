// 與 Go server 的 model.go / docs/API.md 嚴格對齊的型別。
// 任何欄位改動都應同步這裡與後端,測試台才能忠實反映後端回應。

export interface Channel {
  id: string
  name: string
  ownerID: string
  memberCount: number
  lastMessagePreview: string | null
  updatedAt: string // ISO8601
}

export interface Message {
  id: string
  channelID: string
  authorID: string
  authorName: string
  text: string
  category: string | null
  tags: string[]
  summary: string | null
  createdAt: string // ISO8601
}

// User 是公開身分(成員列表、訊息作者等),不含私密資料。
export interface User {
  id: string
  name: string
  avatarColor: string
}

// Profile 是私密資料,只在「自己的帳號」端點回傳。
export interface Profile {
  email: string
}

// Me 是登入後的自己:公開身分 + 私密資料。GET /v1/me 回傳此結構。
export interface Me {
  user: User
  profile: Profile
}

export interface SearchAnswer {
  answer: string
  citedMessageIDs: string[]
  confidence?: number
}

// login / register / apple 的回應:Me + token。
export interface AuthResponse {
  token: string
  user: User
  profile: Profile
}

// 後端統一錯誤格式:{ "error": { "code", "message" } }
export interface APIErrorBody {
  error: {
    code: string
    message: string
  }
}
