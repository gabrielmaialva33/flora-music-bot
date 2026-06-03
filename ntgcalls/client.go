package ntgcalls

import "sync"

type Client struct {
	ptr                         uintptr
	mu                          sync.RWMutex // Protects all callback slices
	connectionChangeCallbacks   []ConnectionChangeCallback
	streamEndCallbacks          []StreamEndCallback
	upgradeCallbacks            []UpgradeCallback
	signalCallbacks             []SignalCallback
	frameCallbacks              []FrameCallback
	remoteSourceCallbacks       []RemoteSourceCallback
	broadcastTimestampCallbacks []BroadcastTimestampCallback
	broadcastPartCallbacks      []BroadcastPartCallback
}

// appendLocked appends v to *s under mu's write lock.
func appendLocked[T any](mu *sync.RWMutex, s *[]T, v T) {
	mu.Lock()
	defer mu.Unlock()
	*s = append(*s, v)
}

// snapshot returns a copy of *s taken under mu's read lock, or nil if empty.
// Taking the slice by pointer ensures the header is read inside the lock,
// avoiding a data race with appendLocked which reassigns the header.
func snapshot[T any](mu *sync.RWMutex, s *[]T) []T {
	mu.RLock()
	defer mu.RUnlock()
	if len(*s) == 0 {
		return nil
	}
	cbs := make([]T, len(*s))
	copy(cbs, *s)
	return cbs
}

func (ctx *Client) OnStreamEnd(callback StreamEndCallback) {
	appendLocked(&ctx.mu, &ctx.streamEndCallbacks, callback)
}

func (ctx *Client) OnUpgrade(callback UpgradeCallback) {
	appendLocked(&ctx.mu, &ctx.upgradeCallbacks, callback)
}

func (ctx *Client) OnConnectionChange(callback ConnectionChangeCallback) {
	appendLocked(&ctx.mu, &ctx.connectionChangeCallbacks, callback)
}

func (ctx *Client) OnSignal(callback SignalCallback) {
	appendLocked(&ctx.mu, &ctx.signalCallbacks, callback)
}

func (ctx *Client) OnFrame(callback FrameCallback) {
	appendLocked(&ctx.mu, &ctx.frameCallbacks, callback)
}

func (ctx *Client) OnRemoteSourceChange(callback RemoteSourceCallback) {
	appendLocked(&ctx.mu, &ctx.remoteSourceCallbacks, callback)
}

func (ctx *Client) OnRequestBroadcastTimestamp(
	callback BroadcastTimestampCallback,
) {
	appendLocked(&ctx.mu, &ctx.broadcastTimestampCallbacks, callback)
}

func (ctx *Client) OnRequestBroadcastPart(callback BroadcastPartCallback) {
	appendLocked(&ctx.mu, &ctx.broadcastPartCallbacks, callback)
}

func (ctx *Client) getStreamEndCallbacks() []StreamEndCallback {
	return snapshot(&ctx.mu, &ctx.streamEndCallbacks)
}

func (ctx *Client) getUpgradeCallbacks() []UpgradeCallback {
	return snapshot(&ctx.mu, &ctx.upgradeCallbacks)
}

func (ctx *Client) getConnectionChangeCallbacks() []ConnectionChangeCallback {
	return snapshot(&ctx.mu, &ctx.connectionChangeCallbacks)
}

func (ctx *Client) getSignalCallbacks() []SignalCallback {
	return snapshot(&ctx.mu, &ctx.signalCallbacks)
}

func (ctx *Client) getFrameCallbacks() []FrameCallback {
	return snapshot(&ctx.mu, &ctx.frameCallbacks)
}

func (ctx *Client) getRemoteSourceCallbacks() []RemoteSourceCallback {
	return snapshot(&ctx.mu, &ctx.remoteSourceCallbacks)
}

func (ctx *Client) getBroadcastTimestampCallbacks() []BroadcastTimestampCallback {
	return snapshot(&ctx.mu, &ctx.broadcastTimestampCallbacks)
}

func (ctx *Client) getBroadcastPartCallbacks() []BroadcastPartCallback {
	return snapshot(&ctx.mu, &ctx.broadcastPartCallbacks)
}
