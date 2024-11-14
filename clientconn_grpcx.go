package grpc

import (
	"context"
	"google.golang.org/grpc/internal/channelz"
	"google.golang.org/grpc/resolver"
	"time"
)

// tryConnectAllAddrs tries to create multiple connection to the addresses.
// It returns an error if no address was successfully
// connected, or updates ac appropriately with the new transport.
func (ac *addrConn) tryConnectAllAddrs(ctx context.Context, addrs []resolver.Address, connectDeadline time.Time) error {
	var firstConnErr error
	connected := false
	for _, addr := range addrs {
		ac.channelz.ChannelMetrics.Target.Store(&addr.Addr)
		if ctx.Err() != nil {
			return errConnClosing
		}
		ac.mu.Lock()

		ac.cc.mu.RLock()
		ac.dopts.copts.KeepaliveParams = ac.cc.mkp
		ac.cc.mu.RUnlock()

		copts := ac.dopts.copts
		if ac.scopts.CredsBundle != nil {
			copts.CredsBundle = ac.scopts.CredsBundle
		}
		ac.mu.Unlock()

		channelz.Infof(logger, ac.channelz, "Subchannel picks a new address %q to connect", addr.Addr)

		err := ac.createTransport(ctx, addr, copts, connectDeadline)
		if err == nil {
			connected = true
			// continue connecting to other addrs
			continue
		}
		if firstConnErr == nil {
			firstConnErr = err
		}
		ac.cc.updateConnectionError(err)
	}

	if connected {
		return nil
	}

	// Couldn't connect to any address.
	return firstConnErr
}
