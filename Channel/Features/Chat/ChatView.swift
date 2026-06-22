import SwiftUI

/// 頻道聊天畫面:訊息流 + 底部輸入列。
/// owner 輸入=發訊息(走 LLM 分類);成員輸入=語意查詢(走 RAG 回答,顯示在訊息流)。
struct ChatView: View {
    let channel: Channel
    @Environment(AppState.self) private var app
    @State private var store: ChatStore?
    @State private var draft = ""
    @State private var showingMembers = false
    @FocusState private var inputFocused: Bool

    /// 目前使用者是否為頻道擁有者。
    private var isOwner: Bool { channel.ownerID == app.currentUser.id }

    var body: some View {
        Group {
            if let store {
                content(store)
            } else {
                ProgressView()
            }
        }
        .task { await setup() }
        .navigationTitle(channel.name)
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .primaryAction) {
                Button { showingMembers = true } label: { Image(systemName: "person.2") }
            }
        }
        .sheet(isPresented: $showingMembers) {
            NavigationStack { MembersView(channel: channel) }
        }
    }

    @ViewBuilder
    private func content(_ store: ChatStore) -> some View {
        VStack(spacing: 0) {
            ScrollViewReader { proxy in
                ScrollView {
                    LazyVStack(spacing: 12) {
                        ForEach(store.messages) { msg in
                            MessageRow(message: msg, isMe: msg.authorID == store.currentUserID)
                                .id(msg.id)
                        }
                    }
                    .padding()
                }
                .onChange(of: store.messages.count) {
                    if let last = store.messages.last {
                        withAnimation { proxy.scrollTo(last.id, anchor: .bottom) }
                    }
                }
            }
            inputBar(store)
        }
    }

    private func inputBar(_ store: ChatStore) -> some View {
        HStack(spacing: 10) {
            TextField(isOwner ? "輸入訊息…" : "用自然語言查詢這個頻道…",
                      text: $draft, axis: .vertical)
                .textFieldStyle(.plain)
                .padding(.horizontal, 14).padding(.vertical, 8)
                .background(Color(.secondarySystemBackground), in: Capsule())
                .focused($inputFocused)
                .lineLimit(1...4)

            Button {
                let text = draft
                draft = ""
                Task {
                    if isOwner {
                        await store.send(text)
                    } else {
                        await store.ask(text)
                    }
                }
            } label: {
                Image(systemName: isOwner ? "arrow.up.circle.fill" : "sparkle.magnifyingglass")
                    .font(.title)
                    .foregroundStyle(draft.trimmingCharacters(in: .whitespaces).isEmpty
                                     ? .gray : (isOwner ? .accentColor : .purple))
            }
            .disabled(draft.trimmingCharacters(in: .whitespaces).isEmpty)
        }
        .padding(.horizontal).padding(.vertical, 8)
        .background(.bar)
    }

    private func setup() async {
        guard store == nil else { return }
        let s = ChatStore(backend: app.backend, channel: channel)
        store = s
        await s.load()
    }
}
