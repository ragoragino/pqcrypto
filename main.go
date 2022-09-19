package main

import (
	"context"
	"net"
	"net/http"
	"time"
)

type TLSStateID string

const (
	ClientHelloStateID TLSStateID = "ClientHelloState"
	ServerHelloStateID TLSStateID = "ServerHelloState"
)

type TLSStateContext struct {
	c net.Conn

	ClientHello string
	ServerHello string
}

type TLSState interface {
	Execute(ctx context.Context) error
}

type stateWithHook struct {
	hook TLSHook

	TLSState
}

func (s *stateWithHook) Execute(ctx context.Context) error {
	if s.hook != nil {
		if err := s.hook(ctx); err != nil {
			return err
		}
	}

	return s.TLSState.Execute(ctx)
}

type ClientHelloState struct {
	tlsCtx *TLSStateContext
}

func (s *ClientHelloState) Execute(ctx context.Context) error {
	// TODO
	clientHelloBytes := []byte{}

	_, err := s.tlsCtx.c.Write(clientHelloBytes)
	if err != nil {
		return err
	}

	return nil
}

type ServerHelloState struct {
	tlsCtx *TLSStateContext
}

func (s *ServerHelloState) Execute(ctx context.Context) error {
	serverHelloBuffer := []byte{}
	_, err := s.tlsCtx.c.Read(serverHelloBuffer)
	if err != nil {
		return err
	}

	return nil
}

type ServerVerifyState struct {
	tlsCtx *TLSStateContext
}

func (s *ServerVerifyState) Execute(ctx context.Context) error {
	return nil
}

type TLSHook func(ctx context.Context) error

type TLSStateMachine struct {
	tlsCtx TLSStateContext

	hooks  map[TLSStateID]TLSHook
	stages map[TLSStateID]TLSState

	lastState TLSState
}

func (sm *TLSStateMachine) Finished() bool {
	return false
}

func (sm *TLSStateMachine) Next() TLSState {
	switch sm.lastState.(type) {
	case *ClientHelloState:
		if stage, ok := sm.stages[ClientHelloStateID]; ok {
			sm.lastState = stage
		} else {
			sm.lastState = &stateWithHook{
				hook: sm.hooks[ClientHelloStateID],
				TLSState: &ServerHelloState{
					tlsCtx: &sm.tlsCtx,
				},
			}
		}

		return sm.lastState
	case *ServerHelloState:
		if stage, ok := sm.stages[ServerHelloStateID]; ok {
			sm.lastState = stage
		} else {
			sm.lastState = &stateWithHook{
				hook: sm.hooks[ServerHelloStateID],
				TLSState: &ServerVerifyState{
					tlsCtx: &sm.tlsCtx,
				},
			}
		}

		return sm.lastState
	}

	return nil
}

type TLSConn struct {
	net.Conn
	network string
	address string

	sm *TLSStateMachine
}

func (c *TLSConn) Read(b []byte) (int, error) {
	return 0, nil
}

func (c *TLSConn) Write(b []byte) (int, error) {
	return 0, nil
}

func (c *TLSConn) Close() error {
	return nil
}

func (c *TLSConn) Handshake(ctx context.Context, hooks map[TLSStateID]TLSHook) error {
	for {
		if c.sm.Finished() {
			break
		}

		n := c.sm.Next()

		err := n.Execute(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func pqTLSWithHooks(hooks map[TLSStateID]TLSHook, stages map[TLSStateID]TLSState) func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return func(ctx context.Context, network string, addr string) (net.Conn, error) {
		c, err := net.Dial(network, addr)
		if err != nil {
			return c, err
		}

		tlsConn := &TLSConn{
			Conn:    c,
			network: network,
			address: addr,
			sm: &TLSStateMachine{
				hooks:  hooks,
				stages: stages,
			},
		}

		err = tlsConn.Handshake(ctx, hooks)
		if err != nil {
			c.Close()

			return tlsConn, nil
		}

		return tlsConn, nil
	}
}

type PQTLSHandshakeHandlerBuilder struct {
	hooks  map[TLSStateID]TLSHook
	stages map[TLSStateID]TLSState
}

func NewPQTLSHandshakeHandlerBuilder() *PQTLSHandshakeHandlerBuilder {
	return &PQTLSHandshakeHandlerBuilder{
		hooks: make(map[TLSStateID]TLSHook),
	}
}

func (b *PQTLSHandshakeHandlerBuilder) Build() func(ctx context.Context, network string, addr string) (net.Conn, error) {
	return pqTLSWithHooks(b.hooks, b.stages)
}

func (b *PQTLSHandshakeHandlerBuilder) AddHook(id TLSStateID, h TLSHook) {
	b.hooks[id] = h
}

func (b *PQTLSHandshakeHandlerBuilder) WithStage(id TLSStateID, s TLSState) {
	b.stages[id] = s
}

func main() {
	builder := PQTLSHandshakeHandlerBuilder{}

	tr := &http.Transport{
		IdleConnTimeout: 30 * time.Second,
		DialTLSContext:  builder.Build(),
	}
	client := &http.Client{Transport: tr}
	_, err := client.Get("https://example.com")
	if err != nil {
		panic(err)
	}
}
