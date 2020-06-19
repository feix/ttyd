package service

import (
	"fmt"
	"strings"
)

type SSHConfig struct {
	User     string `json:"user"`
	App      string `json:"app"`
	Instance string `json:"instance"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

// 登录串解析成 SSHConfig
func SSHConfigParser(loginStr string, port int) *SSHConfig {
	userHost := strings.Split(loginStr, "@")
	if len(userHost) < 2 {
		return nil
	}
	return &SSHConfig{
		User: userHost[0],
		Host: userHost[1],
		Port: port,
	}
}

func (s *SSHConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
