// Package service 实现业务逻辑层
// 本文件保留包级别的通道和消息类型定义
package service

// NoteMigrateChannel 迁移任务通道
var NoteMigrateChannel = make(chan NoteMigrateMsg, 1000)

// NoteMigrateMsg 笔记迁移消息
type NoteMigrateMsg struct {
	OldNoteID int64
	NewNoteID int64
	UID       int64
}

// NoteHistoryMsg 笔记历史记录延时处理消息
type NoteHistoryMsg struct {
	NoteID int64
	UID    int64
}

// NoteHistoryChannel 延时任务通道，后台 task 会监听此通道
var NoteHistoryChannel = make(chan NoteHistoryMsg, 1000)

// NoteHistoryDelayPush 将笔记推送至延时处理队列
func NoteHistoryDelayPush(noteID int64, uid int64) {
	NoteHistoryChannel <- NoteHistoryMsg{NoteID: noteID, UID: uid}
}
