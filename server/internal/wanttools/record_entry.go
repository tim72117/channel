// Package wanttools 提供註冊到 want 引擎的自訂 agent 工具。
// 被 blank import 時,init() 會把工具註冊進 want 的全域 toolbox,
// 之後 LLM agent 在推論時即可呼叫(general role 的 Tools: ["*"] 允許所有工具)。
package wanttools

import (
	"context"
	"fmt"

	"want/types"
)

func init() {
	types.RegisterTool(RecordEntryDeclaration, func() types.ToolInterface {
		return &RecordEntryTool{}
	})
}

// RecordEntryDeclaration 是給 LLM 看的工具宣告。
// 事件時間由 LLM 從訊息解析,支援單一時間點、時間範圍與全日事件;
// 系統另記錄 recorded_at(寫入當下時間)作為審計用。
var RecordEntryDeclaration = types.ToolDeclaration{
	Name: "record_entry",
	Description: "將一則項目記錄成帶有日期時間的條目,寫入記事檔。" +
		"當使用者想把訊息存成待辦、行程、備忘或日誌條目時使用。每呼叫一次新增一筆。" +
		"請從訊息解析出事件的時間,可以是單一時間點、時間範圍或全日事件。",
	Type: "sync",
	Parameters: map[string]interface{}{
		"type": "OBJECT",
		"properties": map[string]interface{}{
			"item": map[string]interface{}{
				"type":        "STRING",
				"description": "要記錄的事項內容(去掉時間後的描述),例如:'開會討論 Q3 預算'",
			},
			"start": map[string]interface{}{
				"type": "STRING",
				"description": "事件開始時間。有明確時刻時用 'YYYY-MM-DD HH:MM';" +
					"全日事件(allDay=true)時用 'YYYY-MM-DD'(不含時刻)。" +
					"相對時間(如「明天」「週五早上十點」)請依提供的今天日期換算成絕對日期。" +
					"訊息完全沒提到時間就留空字串。",
			},
			"end": map[string]interface{}{
				"type": "STRING",
				"description": "事件結束時間,格式同 start。" +
					"只有當訊息表達時間範圍(如「三點到五點」「6/30 到 7/2」)時才填,否則留空字串。",
			},
			"allDay": map[string]interface{}{
				"type": "BOOLEAN",
				"description": "是否為全日事件(只有日期、沒有特定時刻,如「6月30號休假」)。" +
					"有明確時刻時為 false。",
			},
		},
		"required": []string{"item"},
	},
}

type RecordEntryTool struct {
	types.BaseToolConfig
}

func (t *RecordEntryTool) Call(args types.ToolArguments, ctx types.ToolContext) (types.ToolCallResult, error) {
	return t.Execute(context.Background(), args, ctx)
}

func (t *RecordEntryTool) RenderToolUse(args types.ToolArguments) string {
	return fmt.Sprintf("正在記錄條目:%s", args.GetString("item"))
}

func (t *RecordEntryTool) RenderToolUseError(err error) string {
	return fmt.Sprintf("記錄條目失敗:%v", err)
}

func (t *RecordEntryTool) RenderToolResult(data map[string]interface{}) string {
	if msg, ok := data["message"].(string); ok {
		return msg
	}
	return "已記錄條目"
}

func (t *RecordEntryTool) Execute(_ context.Context, args types.ToolArguments, _ types.ToolContext) (types.ToolCallResult, error) {
	item := args.GetString("item")
	if item == "" {
		return types.ToolCallResult{}, fmt.Errorf("item 不可為空")
	}

	// 事件時間由 LLM 從訊息解析。
	start := args.GetString("start")
	end := args.GetString("end")
	allDay := args.GetBool("allDay")

	// 交給 sink 持久化(帶上當前記錄 context 的 messageID / channelID)。
	if err := emit(RecordedEntry{Item: item, Start: start, End: end, AllDay: allDay}); err != nil {
		return types.ToolCallResult{}, fmt.Errorf("寫入條目失敗: %w", err)
	}

	resultMsg := fmt.Sprintf("已記錄:%s %s", describeTime(start, end, allDay), item)
	return types.ToolCallResult{
		Content: []types.ResultContentBlock{types.TextBlock(resultMsg)},
		ToolUseResult: map[string]interface{}{
			"message": resultMsg,
			"start":   start,
			"end":     end,
			"allDay":  allDay,
			"item":    item,
		},
	}, nil
}

// describeTime 把時間描述成人類可讀字串。
func describeTime(start, end string, allDay bool) string {
	switch {
	case start == "":
		return "(未指定時間)"
	case allDay && end != "":
		return start + " ~ " + end + "(全日)"
	case allDay:
		return start + "(全日)"
	case end != "":
		return start + " ~ " + end
	default:
		return start
	}
}
