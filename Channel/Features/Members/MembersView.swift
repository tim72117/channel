import SwiftUI

/// 成員管理:顯示頻道現有成員,搜尋並邀請朋友加入。
struct MembersView: View {
    let channel: Channel
    @Environment(AppState.self) private var app
    @Environment(\.dismiss) private var dismiss

    @State private var members: [User] = []
    @State private var showingInvite = false

    var body: some View {
        List {
            Section("成員(\(members.count))") {
                ForEach(members) { user in
                    HStack(spacing: 12) {
                        AvatarView(user: user)
                        Text(user.name)
                        if user.id == app.currentUser.id {
                            Text("你").font(.caption).foregroundStyle(.secondary)
                        }
                    }
                }
            }
        }
        .navigationTitle("成員")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar {
            ToolbarItem(placement: .topBarTrailing) {
                Button { showingInvite = true } label: { Label("加朋友", systemImage: "person.badge.plus") }
            }
            ToolbarItem(placement: .topBarLeading) {
                Button("完成") { dismiss() }
            }
        }
        .sheet(isPresented: $showingInvite) {
            NavigationStack {
                InviteFriendView(channel: channel, existing: members) { updated in
                    members = updated
                }
            }
        }
        .task { await load() }
    }

    private func load() async {
        do { members = try await app.backend.fetchMembers(channelID: channel.id) }
        catch { }
    }
}

/// 輸入 email 邀請使用者加入頻道。
private struct InviteFriendView: View {
    let channel: Channel
    let existing: [User]
    let onAdded: ([User]) -> Void

    @Environment(AppState.self) private var app
    @Environment(\.dismiss) private var dismiss
    @State private var email = ""
    @State private var isAdding = false
    @State private var errorMessage: String?
    @FocusState private var focused: Bool

    var body: some View {
        Form {
            Section {
                TextField("輸入對方的 Email", text: $email)
                    .textContentType(.emailAddress)
                    .keyboardType(.emailAddress)
                    .textInputAutocapitalization(.never)
                    .autocorrectionDisabled()
                    .focused($focused)
            } footer: {
                Text("輸入已註冊使用者的 Email,即可邀請加入此頻道。")
            }

            Section {
                Button {
                    Task { await invite() }
                } label: {
                    HStack {
                        Spacer()
                        if isAdding { ProgressView() } else { Text("邀請加入") }
                        Spacer()
                    }
                }
                .disabled(!canInvite || isAdding)
            }

            if let errorMessage {
                Section { Text(errorMessage).foregroundStyle(.red).font(.callout) }
            }
        }
        .navigationTitle("加入成員")
        .navigationBarTitleDisplayMode(.inline)
        .toolbar { ToolbarItem(placement: .topBarTrailing) { Button("完成") { dismiss() } } }
        .onAppear { focused = true }
    }

    private var canInvite: Bool { email.contains("@") }

    private func invite() async {
        let e = email.trimmingCharacters(in: .whitespaces).lowercased()
        isAdding = true
        errorMessage = nil
        defer { isAdding = false }
        do {
            let updated = try await app.backend.addMember(channelID: channel.id, email: e)
            onAdded(updated)
            dismiss()
        } catch {
            errorMessage = error.localizedDescription
        }
    }
}
