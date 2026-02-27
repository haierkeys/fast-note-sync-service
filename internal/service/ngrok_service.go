package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"

	"go.uber.org/zap"
	"golang.ngrok.com/ngrok/v2"
)

// NgrokService provides ngrok tunnel service
// NgrokService 提供 ngrok 隧道服务
type NgrokService interface {
	Start(ctx context.Context, addr string) error
	Stop(ctx context.Context) error
	TunnelURL() string
}

type ngrokService struct {
	logger    *zap.Logger
	authToken string
	domain    string
	listener  net.Listener
	url       string
	agent     ngrok.Agent
}

// NewNgrokService creates a new ngrok service
// NewNgrokService 创建一个新的 ngrok 服务
func NewNgrokService(logger *zap.Logger, authToken, domain string) NgrokService {
	return &ngrokService{
		logger:    logger,
		authToken: authToken,
		domain:    domain,
	}
}

// Start starts the ngrok tunnel
// Start 启动 ngrok 隧道
func (s *ngrokService) Start(ctx context.Context, addr string) error {
	if s.authToken == "" {
		return fmt.Errorf("ngrok auth token is required")
	}

	// 1. Create an agent instance to hold AgentOption
	agent, err := ngrok.NewAgent(ngrok.WithAuthtoken(s.authToken))
	if err != nil {
		return fmt.Errorf("failed to create ngrok v2 agent: %w", err)
	}
	s.agent = agent

	// 2. Configure endpoint options
	var endpointOpts []ngrok.EndpointOption
	if s.domain != "" {
		endpointOpts = append(endpointOpts, ngrok.WithURL("https://"+s.domain))
	}

	// 3. Listen creates the endpoint
	ln, err := agent.Listen(ctx, endpointOpts...)
	if err != nil {
		return fmt.Errorf("failed to start ngrok v2 tunnel: %w", err)
	}
	s.listener = ln

	if u, ok := ln.(interface{ URL() *url.URL }); ok {
		s.url = u.URL().String()
	} else if u, ok := ln.(interface{ URL() string }); ok {
		s.url = u.URL()
	} else {
		s.url = ln.Addr().String()
	}

	s.logger.Info("ngrok v2 tunnel established", zap.String("url", s.url))

	// Start forwarding
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				s.logger.Debug("ngrok tunnel accept error (likely closed)", zap.Error(err))
				return
			}
			go s.handleConn(conn, addr)
		}
	}()

	return nil
}

func (s *ngrokService) handleConn(conn net.Conn, addr string) {
	defer conn.Close()
	localConn, err := net.Dial("tcp", addr)
	if err != nil {
		s.logger.Error("failed to dial local address", zap.String("addr", addr), zap.Error(err))
		return
	}
	defer localConn.Close()

	done := make(chan struct{}, 2)
	go func() {
		_, _ = io.Copy(localConn, conn)
		done <- struct{}{}
	}()
	go func() {
		_, _ = io.Copy(conn, localConn)
		done <- struct{}{}
	}()
	<-done
}

// Stop stops the ngrok tunnel
// Stop 停止 ngrok 隧道
func (s *ngrokService) Stop(ctx context.Context) error {
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.Warn("failed to close ngrok tunnel", zap.Error(err))
		}
	}
	if s.agent != nil {
		if err := s.agent.Disconnect(); err != nil {
			s.logger.Warn("failed to disconnect ngrok agent", zap.Error(err))
		}
	}
	return nil
}

// TunnelURL returns the current tunnel URL
// TunnelURL 返回当前隧道 URL
func (s *ngrokService) TunnelURL() string {
	return s.url
}
