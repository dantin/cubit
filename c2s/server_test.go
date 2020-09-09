package c2s

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/dantin/cubit/component"
	"github.com/dantin/cubit/module"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/transport"
	"github.com/stretchr/testify/require"
)

func TestC2SSocketServer(t *testing.T) {
	r, _, _ := setupTest("localhost")

	errCh := make(chan error)
	cfg := Config{
		ID:               "srv-123",
		ConnectTimeout:   time.Second * time.Duration(5),
		MaxStanzaSize:    8192,
		ResourceConflict: Reject,
		Transport: TransportConfig{
			Type: transport.Socket,
			Port: 9998,
		},
	}
	srv := server{
		cfg:           &cfg,
		router:        r,
		mods:          &module.Modules{},
		comps:         &component.Components{},
		inConnections: make(map[string]stream.C2S),
	}
	go srv.start()

	go func() {
		time.Sleep(time.Millisecond * 150)

		// test XMPP port...
		conn, err := net.Dial("tcp", "127.0.0.1:9998")
		if err != nil {
			errCh <- err
			return
		}

		xmlHdr := []byte(`<?xml version="1.0" encoding="UTF-8">`)
		_, err = conn.Write(xmlHdr)
		if err != nil {
			errCh <- err
			return
		}

		time.Sleep(time.Millisecond * 150) // wait until disconnected

		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second*5))
		defer cancel()

		_ = srv.shutdown(ctx)
		errCh <- nil
	}()
	err := <-errCh
	require.Nil(t, err)
}
