package service

// staleReviveToleranceMs is the clock-skew tolerance (5 minutes) applied when
// judging whether an upload predates a server-side deletion. Devices are
// expected to be NTP-synced; this absorbs reasonable skew without weakening
// stale detection — stale re-reports are typically days old.
//
// staleReviveToleranceMs 判定上传是否早于服务端删除时刻时容忍的时钟偏差（5 分钟）。
const staleReviveToleranceMs = 5 * 60 * 1000

// isStaleRevive reports whether an upload that would revive a server-deleted
// resource is a "stale revive": both client timestamps (mtime and ctime) fail
// to prove the upload is newer than the server-side deletion moment.
//
// Rationale: clients that lost their local sync state re-report entire vaults;
// without this guard, any such report flips delete tombstones back to create
// and broadcasts the resurrection to every device (the delete/resurrect
// tug-of-war). An upload is treated as fresh — and therefore allowed to revive
// — when either timestamp falls within the tolerance window of the deletion
// moment or later, which keeps genuine re-creations, recycle-bin restores and
// post-deletion edits working.
//
// isStaleRevive 判断一次"想复活已删资源"的上传是否为陈旧重报：
// 当 mtime 与 ctime 都无法证明上传晚于删除时刻（含容差）时返回 true。
// 任一时间戳落在「删除时刻 - 容差」之后即视为真实的新意图，放行复活。
func isStaleRevive(mtime, ctime, deletedAt int64) bool {
	if deletedAt <= 0 {
		return false
	}
	fresh := func(ts int64) bool { return ts > 0 && ts >= deletedAt-staleReviveToleranceMs }
	return !fresh(mtime) && !fresh(ctime)
}
