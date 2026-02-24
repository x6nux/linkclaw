package mcp

// deployToolDefs 入职运维工具定义（HR / Chairman 专用）
var deployToolDefs = []ToolDef{
	{Perm: PermOnboard, Tool: Tool{
		Name:        "ssh_exec",
		Description: "在远程服务器上通过 SSH 执行命令。返回命令输出。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"host", "user", "command"},
			Properties: map[string]PropSchema{
				"host":     {Type: "string", Description: "SSH 主机地址"},
				"port":     {Type: "string", Description: "SSH 端口（默认 22）"},
				"user":     {Type: "string", Description: "SSH 用户名"},
				"password": {Type: "string", Description: "SSH 密码（与 key 二选一）"},
				"key":      {Type: "string", Description: "SSH 私钥内容（与 password 二选一）"},
				"command":  {Type: "string", Description: "要执行的 shell 命令"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "ssh_upload",
		Description: "通过 SSH 上传文件到远程服务器。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"host", "user", "remote_path", "source"},
			Properties: map[string]PropSchema{
				"host":        {Type: "string", Description: "SSH 主机地址"},
				"port":        {Type: "string", Description: "SSH 端口（默认 22）"},
				"user":        {Type: "string", Description: "SSH 用户名"},
				"password":    {Type: "string", Description: "SSH 密码（与 key 二选一）"},
				"key":         {Type: "string", Description: "SSH 私钥内容（与 password 二选一）"},
				"remote_path": {Type: "string", Description: "远程目标路径（如 /tmp/file.tar.gz）"},
				"source":      {Type: "string", Description: "上传源：内置插件名（如 openclaw-linkclaw）或 base64 编码内容"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "docker_run",
		Description: "在本地 Docker 运行容器。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"image", "name"},
			Properties: map[string]PropSchema{
				"image":   {Type: "string", Description: "Docker 镜像名（如 ghcr.io/qwibitai/openclaw:latest）"},
				"name":    {Type: "string", Description: "容器名称"},
				"env":     {Type: "string", Description: "环境变量，格式：KEY=VALUE，多个用换行分隔"},
				"volumes": {Type: "string", Description: "挂载卷，格式：host:container[:ro]，多个用换行分隔"},
				"network": {Type: "string", Description: "Docker 网络名称（默认自动加入项目网络 deploy_default）"},
				"extra":   {Type: "string", Description: "额外 docker run 参数"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "docker_exec_cmd",
		Description: "在本地 Docker 容器内执行命令。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"container", "command"},
			Properties: map[string]PropSchema{
				"container": {Type: "string", Description: "容器名称或 ID"},
				"command":   {Type: "string", Description: "要在容器内执行的命令"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "docker_cp",
		Description: "在本地 Docker 容器和宿主机之间复制文件。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"source", "dest"},
			Properties: map[string]PropSchema{
				"source": {Type: "string", Description: "源路径（宿主机路径或 container:/path）"},
				"dest":   {Type: "string", Description: "目标路径（宿主机路径或 container:/path）"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "docker_rm",
		Description: "强制删除本地 Docker 容器。",
		InputSchema: InputSchema{
			Type:     "object",
			Required: []string{"container"},
			Properties: map[string]PropSchema{
				"container": {Type: "string", Description: "容器名称或 ID"},
			},
		},
	}},
	{Perm: PermOnboard, Tool: Tool{
		Name:        "get_plugin_info",
		Description: "获取可用的内置插件信息（如 openclaw-linkclaw 插件包是否存在、路径等）。",
		InputSchema: InputSchema{Type: "object"},
	}},
}
