import SwiftUI
import Observation

/// 頻道列表的狀態容器。
@MainActor
@Observable
final class ChannelStore {
    private let backend: BackendService

    var channels: [Channel] = []
    var isLoading = false
    var errorMessage: String?

    init(backend: BackendService) {
        self.backend = backend
    }

    func load() async {
        isLoading = true
        errorMessage = nil
        do {
            channels = try await backend.fetchChannels()
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    func createChannel(name: String) async {
        do {
            let ch = try await backend.createChannel(name: name)
            channels.insert(ch, at: 0)
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}

/// 單一頻道內的聊天狀態容器:訊息流、發訊息(樂觀更新 + LLM 標注回填)。
@MainActor
@Observable
final class ChatStore {
    private let backend: BackendService
    let channel: Channel

    var messages: [Message] = []
    var isLoading = false
    var errorMessage: String?

    init(backend: BackendService, channel: Channel) {
        self.backend = backend
        self.channel = channel
    }

    var currentUserID: String { backend.currentUser.id }

    func load() async {
        isLoading = true
        do {
            messages = try await backend.fetchMessages(channelID: channel.id)
        } catch {
            errorMessage = error.localizedDescription
        }
        isLoading = false
    }

    /// 發送訊息:先樂觀插入「處理中」的訊息,再等後端 LLM 標注回傳後就地替換。
    func send(_ text: String) async {
        let trimmed = text.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }

        let tempID = "temp_\(UUID().uuidString.prefix(6))"
        let optimistic = Message(
            id: tempID,
            channelID: channel.id,
            authorID: backend.currentUser.id,
            authorName: backend.currentUser.name,
            text: trimmed,
            isProcessing: true
        )
        messages.append(optimistic)

        do {
            let saved = try await backend.postMessage(channelID: channel.id, text: trimmed)
            if let idx = messages.firstIndex(where: { $0.id == tempID }) {
                messages[idx] = saved
            }
        } catch {
            // 失敗:標記該樂觀訊息,並帶上真正的失敗原因方便排查。
            if let idx = messages.firstIndex(where: { $0.id == tempID }) {
                messages[idx].isProcessing = false
                messages[idx].text += "\n⚠️ 傳送失敗:\(error.localizedDescription)"
            }
            errorMessage = error.localizedDescription
        }
    }

    /// 助手回答用的固定作者 ID(本地顯示,不存後端)。
    static let assistantID = "usr_assistant"

    /// 成員用:把問題以自然語言查詢頻道,回答顯示在訊息流(本地,不寫入頻道)。
    func ask(_ question: String) async {
        let trimmed = question.trimmingCharacters(in: .whitespacesAndNewlines)
        guard !trimmed.isEmpty else { return }

        // 我的提問氣泡。
        messages.append(Message(
            id: "ask_\(UUID().uuidString.prefix(6))",
            channelID: channel.id,
            authorID: backend.currentUser.id,
            authorName: backend.currentUser.name,
            text: trimmed))

        // 助手「思考中」氣泡。
        let pendingID = "ans_\(UUID().uuidString.prefix(6))"
        messages.append(Message(
            id: pendingID,
            channelID: channel.id,
            authorID: ChatStore.assistantID,
            authorName: "助手",
            text: "",
            isProcessing: true))

        do {
            let answer = try await backend.semanticQuery(channelID: channel.id, question: trimmed)
            if let idx = messages.firstIndex(where: { $0.id == pendingID }) {
                messages[idx].text = answer.answer
                messages[idx].isProcessing = false
            }
        } catch {
            if let idx = messages.firstIndex(where: { $0.id == pendingID }) {
                messages[idx].text = "查詢失敗:\(error.localizedDescription)"
                messages[idx].isProcessing = false
            }
            errorMessage = error.localizedDescription
        }
    }
}
