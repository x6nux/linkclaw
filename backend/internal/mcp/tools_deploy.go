package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// pluginDir 内置插件目录
const pluginDir = "/app/plugins"

// defaultDockerNetwork 返回当前容器所在的 Docker 网络（用于新容器默认加入）
func defaultDockerNetwork() string {
	if v := os.Getenv("DOCKER_NETWORK"); v != "" {
		return v
	}
	return "deploy_default"
}

// ── SSH 工具 ─────────────────────────────────────────────────────────

func (h *Handler) toolSSHExec(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Host     string `json:"host"`
		Port     string `json:"port"`
		User     string `json:"user"`
		Password string `json:"password"`
		Key      string `json:"key"`
		Command  string `json:"command"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Host == "" || p.User == "" || p.Command == "" {
		return ErrorResult("参数错误：需要 host, user, command")
	}

	client, err := sshDial(p.Host, p.Port, p.User, p.Password, p.Key)
	if err != nil {
		return ErrorResult("SSH 连接失败: " + err.Error())
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return ErrorResult("SSH session 创建失败: " + err.Error())
	}
	defer sess.Close()

	out, err := sess.CombinedOutput(p.Command)
	output := strings.TrimSpace(string(out))
	if err != nil {
		return ErrorResult(fmt.Sprintf("命令执行失败: %v\n输出:\n%s", err, output))
	}
	if output == "" {
		output = "(无输出)"
	}
	return TextResult(output)
}

func (h *Handler) toolSSHUpload(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Host       string `json:"host"`
		Port       string `json:"port"`
		User       string `json:"user"`
		Password   string `json:"password"`
		Key        string `json:"key"`
		RemotePath string `json:"remote_path"`
		Source     string `json:"source"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Host == "" || p.User == "" || p.RemotePath == "" || p.Source == "" {
		return ErrorResult("参数错误：需要 host, user, remote_path, source")
	}

	// 解析 source：内置插件名 → 读取文件；否则视为 base64（暂不支持，只支持插件名）
	pluginPath := fmt.Sprintf("%s/%s.tar.gz", pluginDir, p.Source)
	data, err := os.ReadFile(pluginPath)
	if err != nil {
		return ErrorResult(fmt.Sprintf("无法读取插件文件 %s: %v", pluginPath, err))
	}

	client, err := sshDial(p.Host, p.Port, p.User, p.Password, p.Key)
	if err != nil {
		return ErrorResult("SSH 连接失败: " + err.Error())
	}
	defer client.Close()

	sess, err := client.NewSession()
	if err != nil {
		return ErrorResult("SSH session 创建失败: " + err.Error())
	}
	defer sess.Close()

	sess.Stdin = bytes.NewReader(data)
	cmd := fmt.Sprintf("cat > %s", p.RemotePath)
	if out, err := sess.CombinedOutput(cmd); err != nil {
		return ErrorResult(fmt.Sprintf("上传失败: %v — %s", err, strings.TrimSpace(string(out))))
	}

	return TextResult(fmt.Sprintf("已上传 %d 字节到 %s:%s", len(data), p.Host, p.RemotePath))
}

// ── Docker 工具 ──────────────────────────────────────────────────────

func (h *Handler) toolDockerRun(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Image   string `json:"image"`
		Name    string `json:"name"`
		Env     string `json:"env"`
		Volumes string `json:"volumes"`
		Network string `json:"network"`
		Extra   string `json:"extra"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Image == "" || p.Name == "" {
		return ErrorResult("参数错误：需要 image, name")
	}

	cmdArgs := []string{"run", "-d", "--name", p.Name, "--restart", "unless-stopped"}

	network := p.Network
	if network == "" {
		network = defaultDockerNetwork()
	}
	cmdArgs = append(cmdArgs, "--network", network)
	for _, env := range splitLines(p.Env) {
		cmdArgs = append(cmdArgs, "-e", env)
	}
	for _, vol := range splitLines(p.Volumes) {
		cmdArgs = append(cmdArgs, "-v", vol)
	}
	if p.Extra != "" {
		cmdArgs = append(cmdArgs, strings.Fields(p.Extra)...)
	}
	cmdArgs = append(cmdArgs, p.Image)

	out, err := dockerCmd(cmdArgs...)
	if err != nil {
		return ErrorResult(fmt.Sprintf("docker run 失败: %v\n%s", err, out))
	}
	return TextResult(fmt.Sprintf("容器 %s 已启动\n%s", p.Name, out))
}

func (h *Handler) toolDockerExecCmd(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Container string `json:"container"`
		Command   string `json:"command"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Container == "" || p.Command == "" {
		return ErrorResult("参数错误：需要 container, command")
	}

	out, err := dockerCmd("exec", p.Container, "sh", "-c", p.Command)
	if err != nil {
		return ErrorResult(fmt.Sprintf("docker exec 失败: %v\n%s", err, out))
	}
	if out == "" {
		out = "(无输出)"
	}
	return TextResult(out)
}

func (h *Handler) toolDockerCp(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Source string `json:"source"`
		Dest   string `json:"dest"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Source == "" || p.Dest == "" {
		return ErrorResult("参数错误：需要 source, dest")
	}

	// 支持内置插件路径替换
	src := expandPluginPath(p.Source)
	out, err := dockerCmd("cp", src, p.Dest)
	if err != nil {
		return ErrorResult(fmt.Sprintf("docker cp 失败: %v\n%s", err, out))
	}
	return TextResult(fmt.Sprintf("已复制 %s → %s", p.Source, p.Dest))
}

func (h *Handler) toolDockerRm(_ context.Context, _ *Session, args json.RawMessage) ToolCallResult {
	var p struct {
		Container string `json:"container"`
	}
	if err := json.Unmarshal(args, &p); err != nil || p.Container == "" {
		return ErrorResult("参数错误：需要 container")
	}

	out, err := dockerCmd("rm", "-f", p.Container)
	if err != nil {
		return ErrorResult(fmt.Sprintf("docker rm 失败: %v\n%s", err, out))
	}
	return TextResult(fmt.Sprintf("容器 %s 已删除", p.Container))
}

func (h *Handler) toolGetPluginInfo(_ context.Context, _ *Session, _ json.RawMessage) ToolCallResult {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return TextResult(fmt.Sprintf("插件目录 %s 不存在或无法读取", pluginDir))
	}
	var lines []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		lines = append(lines, fmt.Sprintf("- %s (%d KB)", e.Name(), size/1024))
	}
	if len(lines) == 0 {
		return TextResult("插件目录为空")
	}
	return TextResult(fmt.Sprintf("可用插件（%s）：\n%s", pluginDir, strings.Join(lines, "\n")))
}

// ── SSH 辅助 ─────────────────────────────────────────────────────────

func sshDial(host, port, user, password, key string) (*ssh.Client, error) {
	if port == "" {
		port = "22"
	}

	var authMethods []ssh.AuthMethod
	if key != "" {
		signer, err := ssh.ParsePrivateKey([]byte(key))
		if err != nil {
			return nil, fmt.Errorf("解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if password != "" {
		authMethods = append(authMethods, ssh.Password(password))
	}
	if len(authMethods) == 0 {
		if signer, err := loadDefaultSSHKey(); err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	return ssh.Dial("tcp",
		net.JoinHostPort(host, port),
		&ssh.ClientConfig{
			User:            user,
			Auth:            authMethods,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
			Timeout:         15 * time.Second,
		},
	)
}

func loadDefaultSSHKey() (ssh.Signer, error) {
	for _, path := range []string{"/root/.ssh/id_ed25519", "/root/.ssh/id_rsa"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		return ssh.ParsePrivateKey(data)
	}
	return nil, fmt.Errorf("no default SSH key found")
}

// ── Docker 辅助 ──────────────────────────────────────────────────────

func dockerCmd(args ...string) (string, error) {
	out, err := exec.Command("docker", args...).CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func expandPluginPath(path string) string {
	// 如果路径以 @plugin/ 开头，替换为实际插件目录
	if strings.HasPrefix(path, "@plugin/") {
		name := strings.TrimPrefix(path, "@plugin/")
		return fmt.Sprintf("%s/%s", pluginDir, name)
	}
	return path
}
