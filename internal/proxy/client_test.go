package proxy

import (
	"testing"
	"time"
)

func TestNewProxyClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *ProxyConfig
		wantErr bool
	}{
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
		{
			name: "direct connection",
			config: &ProxyConfig{
				Type:                ProxyTypeNone,
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
			wantErr: false,
		},
		{
			name: "socks5 with address",
			config: &ProxyConfig{
				Type:                ProxyTypeSOCKS5,
				Address:             "localhost:1080",
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
			wantErr: false,
		},
		{
			name: "http with address",
			config: &ProxyConfig{
				Type:                ProxyTypeHTTP,
				Address:             "localhost:8080",
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
			wantErr: false,
		},
		{
			name: "socks5 without address",
			config: &ProxyConfig{
				Type:                ProxyTypeSOCKS5,
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
			wantErr: true,
		},
		{
			name: "invalid proxy type",
			config: &ProxyConfig{
				Type:                "invalid",
				Timeout:             30 * time.Second,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewProxyClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewProxyClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewProxyClient() returned nil client without error")
			}
			if client != nil {
				client.Close()
			}
		})
	}
}

func TestDefaultProxyConfig(t *testing.T) {
	cfg := DefaultProxyConfig()
	if cfg == nil {
		t.Fatal("DefaultProxyConfig() returned nil")
	}
	if cfg.Type != ProxyTypeNone {
		t.Errorf("Expected type %s, got %s", ProxyTypeNone, cfg.Type)
	}
	if cfg.MaxIdleConns != 100 {
		t.Errorf("Expected MaxIdleConns 100, got %d", cfg.MaxIdleConns)
	}
	if cfg.MaxIdleConnsPerHost != 10 {
		t.Errorf("Expected MaxIdleConnsPerHost 10, got %d", cfg.MaxIdleConnsPerHost)
	}
}

func TestProxyClient_Config(t *testing.T) {
	cfg := &ProxyConfig{
		Type:                ProxyTypeNone,
		Timeout:             30 * time.Second,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
	}
	client, err := NewProxyClient(cfg)
	if err != nil {
		t.Fatalf("NewProxyClient() error = %v", err)
	}
	defer client.Close()

	returnedCfg := client.Config()
	if returnedCfg != cfg {
		t.Error("Config() returned different config than provided")
	}
}

func TestProxyClient_Client(t *testing.T) {
	client, err := NewProxyClient(nil)
	if err != nil {
		t.Fatalf("NewProxyClient() error = %v", err)
	}
	defer client.Close()

	httpClient := client.Client()
	if httpClient == nil {
		t.Error("Client() returned nil")
	}
}
