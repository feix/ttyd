package proxy

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/feix/ttyd/config"
	"github.com/feix/ttyd/service"
	"golang.org/x/crypto/ssh"
)

var signer ssh.Signer

func init() {
	key, err := ioutil.ReadFile(config.GetEnv("PRIVATE_KEY", ""))
	if err != nil {
		panic(err)
	}

	// 需要口令的私钥 用于 ssh 认证
	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(config.GetEnv("PRIVATE_KEY_PASS", "")))
	if err != nil {
		panic(err)
	}
}

func SSH(ctx context.Context, rw io.ReadWriter, sshConfig service.SSHConfig, winSizeCh <-chan *service.WinSize) error {
	addr := fmt.Sprintf("%s:%d", sshConfig.Host, sshConfig.Port)
	clientConfig := &ssh.ClientConfig{
		User: sshConfig.User,
		// 需要口令的私钥 用于 ssh 认证
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return err
	}
	defer sshClient.Close()

	sshSession, err := sshClient.NewSession()
	if err != nil {
		return err
	}
	defer sshSession.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := sshSession.RequestPty("xterm", 0, 0, modes); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case winSize := <-winSizeCh:
				// 监听 window size 变化
				if winSize != nil {
					_ = sshSession.WindowChange(winSize.Height, winSize.Width)
				}
			case <-ctx.Done():
				// 监听 ctx 变化
				_ = sshSession.Signal(ssh.SIGQUIT)
				return
			}
		}
	}()

	// 绑定 stdin stdout stderr
	sshSession.Stdin = rw
	sshSession.Stdout = rw
	sshSession.Stderr = rw

	if err := sshSession.Shell(); err != nil {
		return err
	}
	return sshSession.Wait()
}
