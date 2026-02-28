package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/haierkeys/fast-note-sync-service/pkg/code"
	"go.uber.org/zap"
)

// CloudflareService provides Cloudflare Tunnel service
// CloudflareService 提供 Cloudflare Tunnel 隧道服务
type CloudflareService interface {
	Start(ctx context.Context, token string, logEnabled bool) error
	Stop(ctx context.Context) error
	TunnelURL() string
	// DownloadBinary downloads the cloudflared binary and returns the path or a detailed error
	// DownloadBinary 下载 cloudflared 二进制文件，返回路径或包含手动下载建议的详细错误
	DownloadBinary() (string, error)
}

type cloudflareService struct {
	logger     zap.Logger
	token      string
	logEnabled bool
	url        string
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	cmd        *exec.Cmd
}

// NewCloudflareService creates a new Cloudflare service
// NewCloudflareService 创建一个新的 Cloudflare 服务
func NewCloudflareService(logger *zap.Logger) CloudflareService {
	return &cloudflareService{
		logger: *logger,
	}
}

// Start starts the Cloudflare tunnel
func (s *cloudflareService) Start(ctx context.Context, token string, logEnabled bool) error {
	if token == "" {
		return fmt.Errorf("cloudflare tunnel token is required")
	}
	s.token = token
	s.logEnabled = logEnabled

	s.ctx, s.cancel = context.WithCancel(ctx)

	s.logger.Info("Starting Cloudflare Tunnel service...")

	// 确保二进制文件存在
	binPath, err := s.DownloadBinary()
	if err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := s.runTunnelProcess(s.ctx, binPath, token); err != nil {
			s.logger.Error("Cloudflare Tunnel process failed", zap.Error(err))
		}
	}()

	return nil
}

// DownloadBinary 实现主动下载逻辑
func (s *cloudflareService) DownloadBinary() (string, error) {
	storageDir := "storage/cloudflared_tunnel"
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %w", err)
	}

	ext := ""
	if runtime.GOOS == "windows" {
		ext = ".exe"
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	fileName := fmt.Sprintf("cloudflared-%s-%s%s", goos, goarch, ext)
	binPath := filepath.Join(storageDir, fileName)

	// 如果文件已存在且可执行
	if _, err := os.Stat(binPath); err == nil {
		return binPath, nil
	}

	// 构造下载链接
	downloadURL := "https://github.com/cloudflare/cloudflared/releases/latest/download/" + fileName

	s.logger.Info("Cloudflared binary not found, attempting to download...", zap.String("url", downloadURL))

	// 执行下载
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(downloadURL)
	if err != nil {
		if code.GetGlobalDefaultLang() == "zh_cn" {
			return "", fmt.Errorf("下载失败:\n%v。 \n[💡 建议] 请手动下载: %s \n并放置于: %s", err, downloadURL, storageDir)
		}
		return "", fmt.Errorf("download failed:\n%v. \n[💡 Suggestion] Please manually download from: %s \nAnd place it in: %s", err, downloadURL, storageDir)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if code.GetGlobalDefaultLang() == "zh_cn" {
			return "", fmt.Errorf("下载服务器返回状态 %s。 \n[💡 建议] 请手动下载: \n %s \n并放置于: %s", resp.Status, downloadURL, storageDir)
		}
		return "", fmt.Errorf("download server returned %s. \n[💡 Suggestion] Please manually download from:\n %s \nAnd place it in: %s", resp.Status, downloadURL, storageDir)
	}

	// 保存到文件
	out, err := os.Create(binPath)
	if err != nil {
		return "", fmt.Errorf("failed to create binary file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("failed to save binary: %w", err)
	}

	// 赋予执行权限 (Unix)
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return "", err
		}
	}

	s.logger.Info("Cloudflared binary downloaded successfully", zap.String("path", binPath))
	return binPath, nil
}

// runTunnelProcess 运行外部进程
func (s *cloudflareService) runTunnelProcess(ctx context.Context, binPath, token string) error {
	var writers []io.Writer
	writers = append(writers, os.Stdout)

	var errWriters []io.Writer
	errWriters = append(errWriters, os.Stderr)

	if s.logEnabled {
		// 确保统一日志目录存在
		logDir := "storage/logs"
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		logPath := filepath.Join(logDir, "cloudflared_tunnel.log")

		logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open cloudflared log file: %w", err)
		}
		defer logFile.Close()

		writers = append(writers, logFile)
		errWriters = append(errWriters, logFile)
		s.logger.Info("Cloudflare Tunnel logging enabled", zap.String("logPath", logPath))
	}

	// 使用 context.WithCancel 可以确保主 Context 取消时，子进程也被杀死
	s.cmd = exec.CommandContext(ctx, binPath, "tunnel", "--no-autoupdate", "run", "--token", token)

	s.cmd.Stdout = io.MultiWriter(writers...)
	s.cmd.Stderr = io.MultiWriter(errWriters...)

	s.logger.Info("Lauching cloudflared process...")
	if err := s.cmd.Start(); err != nil {
		return err
	}

	// 等待进程结束
	if err := s.cmd.Wait(); err != nil && ctx.Err() == nil {
		return fmt.Errorf("cloudflared exited unexpectedly: %w", err)
	}

	return nil
}

// Stop stops the Cloudflare tunnel
// Stop 停止 Cloudflare 隧道
func (s *cloudflareService) Stop(ctx context.Context) error {
	s.logger.Info("Shutting down Cloudflare Tunnel service...")
	if s.cancel != nil {
		s.cancel()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("Cloudflare Tunnel process terminated")
	case <-ctx.Done():
		s.logger.Warn("Cloudflare Tunnel shutdown timed out")
	}

	return nil
}

// TunnelURL returns the current tunnel URL
// TunnelURL 返回当前隧道 URL
func (s *cloudflareService) TunnelURL() string {
	return s.url
}
