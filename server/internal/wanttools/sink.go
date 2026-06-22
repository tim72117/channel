package wanttools

import "sync"

// RecordedEntry 是 record_entry 工具解析出的一筆條目(不含關聯 ID)。
type RecordedEntry struct {
	Item   string
	Start  string
	End    string
	AllDay bool
}

// EntrySink 接收一筆條目與其關聯的 messageID / channelID,負責持久化。
// server 啟動時用 BindSink 注入(寫進 store)。
type EntrySink func(messageID, channelID string, e RecordedEntry) error

var (
	// recordMu 序列化整個「記錄一則訊息」的流程,確保 context 與工具呼叫不交錯。
	recordMu sync.Mutex
	sink     EntrySink
	curMsgID string
	curChnID string
)

// BindSink 注入條目持久化實作(server 啟動時呼叫)。
func BindSink(fn EntrySink) { sink = fn }

// RecordLock / RecordUnlock 包住一次完整的記錄流程(設定 context → 跑 agent → 清除)。
func RecordLock()   { recordMu.Lock() }
func RecordUnlock() { recordMu.Unlock() }

// SetContext 設定本次記錄對應的訊息(在 RecordLock 之後、Submit 之前呼叫)。
func SetContext(messageID, channelID string) {
	curMsgID, curChnID = messageID, channelID
}

// ClearContext 清除 context(本次記錄結束後)。
func ClearContext() { curMsgID, curChnID = "", "" }

// emit 由工具呼叫,把條目交給 sink(帶上當前 context 的關聯 ID)。
func emit(e RecordedEntry) error {
	if sink == nil {
		return nil // 未注入 sink(例如測試)時不持久化。
	}
	return sink(curMsgID, curChnID, e)
}
