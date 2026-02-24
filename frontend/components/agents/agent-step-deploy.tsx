"use client";

export type DeployMode = "manual" | "local_docker" | "ssh_docker";

const DEPLOY_MODE_LABELS: Record<DeployMode, string> = {
  manual:       "仅创建（手动连接）",
  local_docker: "本地 Docker（自动启动）",
  ssh_docker:   "SSH 远程 Docker",
};

const DEPLOY_MODE_DESCS: Record<DeployMode, string> = {
  manual:       "创建 Agent 记录，获取 API Key 后自行部署",
  local_docker: "在当前服务器上通过 Docker 自动拉取并启动",
  ssh_docker:   "SSH 连接到目标服务器，远程执行 docker run",
};

interface Props {
  deployMode: DeployMode;
  setDeployMode: (v: DeployMode) => void;
  sshHost: string;
  setSSHHost: (v: string) => void;
  sshPort: string;
  setSSHPort: (v: string) => void;
  sshUser: string;
  setSSHUser: (v: string) => void;
  sshPassword: string;
  setSSHPassword: (v: string) => void;
  sshKey: string;
  setSSHKey: (v: string) => void;
  sshAuthMethod: "password" | "key";
  setSSHAuthMethod: (v: "password" | "key") => void;
}

export function AgentStepDeploy({
  deployMode, setDeployMode,
  sshHost, setSSHHost, sshPort, setSSHPort, sshUser, setSSHUser,
  sshPassword, setSSHPassword, sshKey, setSSHKey,
  sshAuthMethod, setSSHAuthMethod,
}: Props) {
  return (
    <div className="space-y-3">
      <p className="text-sm text-zinc-400">选择如何启动这个 Agent：</p>

      {(["manual", "local_docker", "ssh_docker"] as DeployMode[]).map(mode => (
        <button
          key={mode}
          onClick={() => setDeployMode(mode)}
          className={`w-full p-4 rounded-lg border text-left transition-colors ${
            deployMode === mode
              ? "border-blue-500 bg-blue-500/10"
              : "border-zinc-700 bg-zinc-800 hover:border-zinc-600"
          }`}
        >
          <div className="flex items-center gap-3">
            <div className={`w-4 h-4 rounded-full border-2 flex-shrink-0 ${deployMode === mode ? "border-blue-500 bg-blue-500" : "border-zinc-600"}`} />
            <div>
              <p className="font-medium text-zinc-100">{DEPLOY_MODE_LABELS[mode]}</p>
              <p className="text-xs text-zinc-500 mt-0.5">{DEPLOY_MODE_DESCS[mode]}</p>
            </div>
          </div>
        </button>
      ))}

      {/* SSH 配置 */}
      {deployMode === "ssh_docker" && (
        <div className="mt-3 p-4 bg-zinc-800/50 rounded-lg border border-zinc-700 space-y-3">
          <p className="text-xs text-zinc-500 font-medium uppercase tracking-wide">SSH 配置</p>
          <div className="grid grid-cols-3 gap-2">
            <div className="col-span-2">
              <label className="block text-xs text-zinc-500 mb-1">主机地址</label>
              <input value={sshHost} onChange={e => setSSHHost(e.target.value)}
                placeholder="192.168.1.100"
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2.5 py-1.5 text-xs text-zinc-100 focus:outline-none focus:border-zinc-500" />
            </div>
            <div>
              <label className="block text-xs text-zinc-500 mb-1">端口</label>
              <input value={sshPort} onChange={e => setSSHPort(e.target.value)}
                placeholder="22"
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2.5 py-1.5 text-xs text-zinc-100 focus:outline-none focus:border-zinc-500" />
            </div>
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">用户名</label>
            <input value={sshUser} onChange={e => setSSHUser(e.target.value)}
              placeholder="root"
              className="w-full bg-zinc-800 border border-zinc-700 rounded px-2.5 py-1.5 text-xs text-zinc-100 focus:outline-none focus:border-zinc-500" />
          </div>
          <div>
            <label className="block text-xs text-zinc-500 mb-1">认证方式</label>
            <div className="flex gap-2">
              {(["password", "key"] as const).map(m => (
                <button key={m} onClick={() => setSSHAuthMethod(m)}
                  className={`px-3 py-1.5 rounded text-xs transition-colors ${sshAuthMethod === m ? "bg-blue-600 text-white" : "bg-zinc-700 text-zinc-400 hover:bg-zinc-600"}`}>
                  {m === "password" ? "密码" : "私钥"}
                </button>
              ))}
            </div>
          </div>
          {sshAuthMethod === "password" ? (
            <div>
              <label className="block text-xs text-zinc-500 mb-1">SSH 密码</label>
              <input type="password" value={sshPassword} onChange={e => setSSHPassword(e.target.value)}
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2.5 py-1.5 text-xs text-zinc-100 focus:outline-none focus:border-zinc-500" />
            </div>
          ) : (
            <div>
              <label className="block text-xs text-zinc-500 mb-1">私钥内容（留空使用服务器默认密钥）</label>
              <textarea value={sshKey} onChange={e => setSSHKey(e.target.value)}
                rows={4} placeholder="-----BEGIN OPENSSH PRIVATE KEY-----&#10;..."
                className="w-full bg-zinc-800 border border-zinc-700 rounded px-2.5 py-1.5 text-xs text-zinc-100 font-mono focus:outline-none focus:border-zinc-500 resize-none" />
            </div>
          )}
        </div>
      )}
    </div>
  );
}
