package h2

import (
	"io"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/hpack"
)

const (
	defaultMaxHeaderListSize = uint32(16 << 20)
	defaultMaxFramSize       = uint32(16384)
	defaultHeaderTableSize   = uint32(4096)
	defaultWindowSize        = 65535
)

// ConnectOptions options for http2 connect with the server.
type ConnectOptions struct {
	InitialWindowSize         uint32
	InitialConnWindowSize     uint32
	MaxFrameSize              uint32
	MaxHeaderListSize         *uint32
	MaxConcurrentStreams      uint32
	InitialMaxSendStreams     uint32
	MaxConcurrentResetStreams uint32
	EnablePush                bool
}

type framer struct {
	fr *http2.Framer
}

func newFramer(conn io.ReadWriteCloser, maxFramSize, maxHeaderListSize uint32) *framer {
	f := &framer{
		fr: http2.NewFramer(conn, conn),
	}
	f.fr.SetMaxReadFrameSize(maxFramSize)
	f.fr.MaxHeaderListSize = maxHeaderListSize
	f.fr.ReadMetaHeaders = hpack.NewDecoder(defaultHeaderTableSize, nil)
	return f
}
