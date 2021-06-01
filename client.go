package h2

import (
	"context"
	"fmt"
	"io"
	"sync"

	"golang.org/x/net/http2"
)

var clientPreface []byte = []byte(http2.ClientPreface)

type Client struct {
	conn   io.ReadWriteCloser
	framer *framer

	mu            sync.Mutex // guard the following variables
	state         State
	activeStreams map[uint32]*Stream
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
	if opts.InitialWindowSize >= defaultWindowSize {
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
	if delta := opts.InitialConnWindowSize - defaultConnWindowSize; delta > 0 {
		if err := client.framer.fr.WriteWindowUpdate(0, delta); err != nil {
			return nil, fmt.Errorf("hankshake: failed to write window update: %v", err)
		}
	}
	return client, nil
}

func (c *Client) reader() {
	frame, err := c.framer.fr.ReadFrame()
	if err != nil {
		c.Close()
		return
	}
	sf, ok := frame.(*http2.SettingsFrame)
	if !ok {
		c.Close()
		return
	}
}

func (c *Client) handleSettings(f *http2.SettingsFrame, isFirst bool) {
	if f.IsAck() {
		return
	}
}

func (c *Client) Close() error {
	c.mu.Lock()
	if c.state == closing {
		c.mu.Unlock()
		return nil
	}

	c.state = closing

	return nil
}
