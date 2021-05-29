package h2s

import (
	"context"
	"fmt"
	"io"

	"golang.org/x/net/http2"
)

var clientPreface []byte = []byte(http2.ClientPreface)

type Client struct {
	conn   io.ReadWriteCloser
	framer *framer
}

func Handshake(ctx context.Context, conn io.ReadWriteCloser, opts *ConnectOptions) (*Client, error) {
	maxFrameSize := defaultMaxFramSize
	if opts.MaxFrameSize > 0 {
		maxFrameSize = opts.MaxFrameSize
	}
	maxHeaderListSize := defaultMaxHeaderListSize
	if opts.MaxHeaderListSize != nil {
		maxHeaderListSize = *opts.MaxHeaderListSize
	}
	client := &Client{
		conn:   conn,
		framer: newFramer(conn, maxFrameSize, maxHeaderListSize),
	}

	n, err := client.conn.Write(clientPreface)
	if err != nil {
		return nil, err
	}
	if n != len(clientPreface) {
		return nil, fmt.Errorf("preface mismatch, wrote %d bytes; want %d", n, len(clientPreface))
	}

	var settings []http2.Setting
	if opts.InitialWindowSize != defaultWindowSize {
		settings = append(settings, http2.Setting{
			ID:  http2.SettingInitialWindowSize,
			Val: opts.InitialWindowSize,
		})
	}
	if opts.MaxHeaderListSize != nil {
		settings = append(settings, http2.Setting{
			ID:  http2.SettingMaxHeaderListSize,
			Val: maxHeaderListSize,
		})
	}
	if err := client.framer.fr.WriteSettings(settings...); err != nil {
		return nil, err
	}

	return client, nil
}
