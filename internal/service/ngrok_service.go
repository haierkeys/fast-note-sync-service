package service

import (
	"context"
	"fmt"
	"io"
	"net"

	"go.uber.org/zap"
	"golang.ngrok.com/ngrok"
	"golang.ngrok.com/ngrok/config"
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
	session   ngrok.Session
	tunnel    ngrok.Tunnel
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

	opts := []ngrok.ConnectOption{
		ngrok.WithAuthtoken(s.authToken),
	}

	session, err := ngrok.Connect(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to ngrok: %w", err)
	}
	s.session = session

	var tunnel ngrok.Tunnel
	if s.domain != "" {
		tunnel, err = session.Listen(ctx, config.HTTPEndpoint(config.WithDomain(s.domain)))
	} else {
		tunnel, err = session.Listen(ctx, config.HTTPEndpoint())
	}
	if err != nil {
		return fmt.Errorf("failed to listen for tunnel: %w", err)
	}
	s.tunnel = tunnel

	s.logger.Info("ngrok tunnel established", zap.String("url", tunnel.URL()))

	// Start forwarding
	go func() {
		for {
			conn, err := tunnel.Accept()
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
		io.Copy(localConn, conn)
		done <- struct{}{}
	}()
	go func() {
		io.Copy(conn, localConn)
		done <- struct{}{}
	}()
	<-done
}

// Stop stops the ngrok tunnel and session
// Stop 停止 ngrok 隧道和会话
func (s *ngrokService) Stop(ctx context.Context) error {
	if s.tunnel != nil {
		if err := s.tunnel.CloseWithContext(ctx); err != nil {
			s.logger.Warn("failed to close ngrok tunnel", zap.Error(err))
		}
	}
	if s.session != nil {
		if err := s.session.Close(); err != nil {
			s.logger.Warn("failed to close ngrok session", zap.Error(err))
		}
	}
	return nil
}

// TunnelURL returns the current tunnel URL
// TunnelURL 返回当前隧道 URL
func (s *ngrokService) TunnelURL() string {
	if s.tunnel != nil {
		return s.tunnel.URL()
	}
	return ""
}
