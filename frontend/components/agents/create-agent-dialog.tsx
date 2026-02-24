"use client";

import { useState } from "react";
import { toast } from "sonner";
import { type Position, type Agent } from "@/lib/types";
import { useDeployment } from "@/hooks/use-deployment";
import { AgentStepBasics } from "./agent-step-basics";
import { AgentStepType, type AgentImageType } from "./agent-step-type";
import { AgentStepDeploy, type DeployMode } from "./agent-step-deploy";
import { AgentStepResult } from "./agent-step-result";
import { AgentDelegateHR } from "./agent-delegate-hr";

interface Props {
  open: boolean;
  onClose: () => void;
  onCreated: (agent: Agent, apiKey: string) => void;
  hasHR?: boolean;
}

type CreateMode = "self" | "delegate_hr";

// 按部门分组职位（排除 chairman）
const ALL_DEPT_POSITIONS: Record<string, Position[]> = {
  "高管": ["cto", "cfo", "coo", "cmo"],
  "人力": ["hr_director", "hr_manager"],
  "产品": ["product_manager", "ux_designer"],
  "工程": ["frontend_dev", "backend_dev", "fullstack_dev", "mobile_dev", "devops", "qa_engineer", "data_engineer"],
  "商务": ["sales_manager", "bd_manager", "customer_success"],
  "市场": ["marketing_manager", "content_creator"],
  "财务": ["accountant", "financial_analyst"],
};

const HR_ONLY_POSITIONS: Record<string, Position[]> = {
  "人力": ["hr_director", "hr_manager"],
};

export function CreateAgentDialog({ open, onClose, onCreated, hasHR = false }: Props) {
  const [step, setStep] = useState(1);
  const [createMode, setCreateMode] = useState<CreateMode>("self");

  // 委托 HR 模式
  const [hrRequest, setHRRequest] = useState("");
  const [hrSent, setHRSent] = useState(false);
  const [isSendingToHR, setIsSendingToHR] = useState(false);

  // Step 1: 基础信息
  const [name, setName] = useState("");
  const [position, setPosition] = useState<Position>(hasHR ? "backend_dev" : "hr_director");
  const [persona, setPersona] = useState("");

  // Step 2: Agent 类型 + 模型
  const [agentImage, setAgentImage] = useState<AgentImageType>("nanoclaw");
  const [model, setModel] = useState("");

  // Step 3: 部署方式
  const [deployMode, setDeployMode] = useState<DeployMode>("manual");
  const [sshHost, setSSHHost] = useState("");
  const [sshPort, setSSHPort] = useState("22");
  const [sshUser, setSSHUser] = useState("root");
  const [sshPassword, setSSHPassword] = useState("");
  const [sshKey, setSSHKey] = useState("");
  const [sshAuthMethod, setSSHAuthMethod] = useState<"password" | "key">("password");

  // Step 4: 结果
  const [isCreating, setIsCreating] = useState(false);
  const [createdAgent, setCreatedAgent] = useState<Agent | null>(null);
  const [apiKey, setApiKey] = useState("");
  const [deployResult, setDeployResult] = useState<{ status: string; error?: string } | null>(null);

  const { deploy, isDeploying, error: deployError } = useDeployment();
  const deptPositions = hasHR ? ALL_DEPT_POSITIONS : HR_ONLY_POSITIONS;

  async function handleDelegateHR() {
    if (!hrRequest.trim()) return;
    setIsSendingToHR(true);
    try {
      const token = localStorage.getItem("lc_token");
      await fetch("/api/v1/messages", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ channel: "general", content: `[招聘需求] ${hrRequest}`, msg_type: "text" }),
      });
      setHRSent(true);
    } finally {
      setIsSendingToHR(false);
    }
  }

  async function handleCreate() {
    setIsCreating(true);
    try {
      const token = localStorage.getItem("lc_token");
      const res = await fetch("/api/v1/agents", {
        method: "POST",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}` },
        body: JSON.stringify({ name, position, persona, model }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || "创建失败");
      setCreatedAgent(data.agent);
      setApiKey(data.api_key || "");
      setStep(4);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : "创建失败");
    } finally {
      setIsCreating(false);
    }
  }

  async function handleDeploy() {
    if (!createdAgent || !apiKey) return;
    const result = await deploy(createdAgent.id, {
      deployType: deployMode as "local_docker" | "ssh_docker",
      agentImage, apiKey,
      sshHost: sshHost || undefined,
      sshPort: sshPort ? parseInt(sshPort) : undefined,
      sshUser: sshUser || undefined,
      sshPassword: sshAuthMethod === "password" ? sshPassword : undefined,
      sshKey: sshAuthMethod === "key" ? sshKey : undefined,
    });
    setDeployResult(result ? { status: result.status, error: result.errorMsg } : { status: "failed", error: deployError || "未知错误" });
  }

  function handleFinish() {
    if (createdAgent) onCreated(createdAgent, apiKey);
    handleReset();
    onClose();
  }

  function handleReset() {
    setStep(1); setCreateMode("self"); setHRRequest(""); setHRSent(false);
    setName(""); setPosition(hasHR ? "backend_dev" : "hr_director"); setPersona("");
    setAgentImage("nanoclaw"); setModel(""); setDeployMode("manual");
    setSSHHost(""); setSSHPort("22"); setSSHUser("root");
    setSSHPassword(""); setSSHKey(""); setSSHAuthMethod("password");
    setIsCreating(false); setIsSendingToHR(false);
    setCreatedAgent(null); setApiKey(""); setDeployResult(null);
  }

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="bg-zinc-900 border border-zinc-700 rounded-xl w-full max-w-lg mx-4 shadow-2xl">
        {/* Header */}
        <div className="border-b border-zinc-800">
          <div className="flex items-center justify-between px-5 pt-5 pb-3">
            <h2 className="text-lg font-semibold text-zinc-50">创建 Agent</h2>
            <button onClick={() => { handleReset(); onClose(); }} className="text-zinc-500 hover:text-zinc-300 transition-colors text-xl leading-none">&times;</button>
          </div>
          {hasHR && (
            <div className="flex px-5 pb-0 gap-0">
              {([["self", "自选职位"], ["delegate_hr", "委托 HR 分配"]] as [CreateMode, string][]).map(([mode, label]) => (
                <button key={mode} onClick={() => { setCreateMode(mode); setStep(1); }}
                  className={`px-4 py-2 text-sm border-b-2 transition-colors ${
                    createMode === mode ? "border-blue-500 text-blue-400" : "border-transparent text-zinc-500 hover:text-zinc-300"
                  }`}
                >{label}</button>
              ))}
            </div>
          )}
          {createMode === "self" && step < 4 && (
            <p className="text-xs text-zinc-600 px-5 pb-3">
              步骤 {step} / 3 — {["基础信息", "Agent 类型", "部署方式"][step - 1]}
            </p>
          )}
        </div>

        {/* Step Content */}
        <div className="p-5 space-y-4 max-h-[70vh] overflow-y-auto">
          {createMode === "delegate_hr" && (
            <AgentDelegateHR hrRequest={hrRequest} setHRRequest={setHRRequest} hrSent={hrSent} isSendingToHR={isSendingToHR} onSend={handleDelegateHR} />
          )}
          {createMode === "self" && step === 1 && (
            <AgentStepBasics name={name} setName={setName} position={position} setPosition={setPosition} persona={persona} setPersona={setPersona} hasHR={hasHR} deptPositions={deptPositions} />
          )}
          {createMode === "self" && step === 2 && (
            <AgentStepType agentImage={agentImage} setAgentImage={setAgentImage} model={model} setModel={setModel} />
          )}
          {createMode === "self" && step === 3 && (
            <AgentStepDeploy deployMode={deployMode} setDeployMode={setDeployMode}
              sshHost={sshHost} setSSHHost={setSSHHost} sshPort={sshPort} setSSHPort={setSSHPort}
              sshUser={sshUser} setSSHUser={setSSHUser} sshPassword={sshPassword} setSSHPassword={setSSHPassword}
              sshKey={sshKey} setSSHKey={setSSHKey} sshAuthMethod={sshAuthMethod} setSSHAuthMethod={setSSHAuthMethod} />
          )}
          {createMode === "self" && step === 4 && createdAgent && (
            <AgentStepResult createdAgent={createdAgent} apiKey={apiKey} deployMode={deployMode}
              agentImage={agentImage} deployResult={deployResult} isDeploying={isDeploying} onDeploy={handleDeploy} />
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-5 border-t border-zinc-800">
          {createMode === "delegate_hr" && !hrSent && (
            <><span /><button onClick={handleDelegateHR} disabled={isSendingToHR || !hrRequest.trim()}
              className="px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors disabled:opacity-40">
              {isSendingToHR ? "发送中…" : "发送给 HR"}</button></>
          )}
          {createMode === "delegate_hr" && hrSent && (
            <><span /><button onClick={() => { handleReset(); onClose(); }}
              className="px-5 py-2 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-sm font-medium text-zinc-200 transition-colors">关闭</button></>
          )}
          {createMode === "self" && (
            <>
              <button onClick={() => step > 1 && step < 4 ? setStep(s => s - 1) : null}
                className={`px-4 py-2 rounded-lg text-sm text-zinc-400 hover:text-zinc-200 transition-colors ${step <= 1 || step === 4 ? "invisible" : ""}`}>
                上一步
              </button>
              {step < 3 && (
                <button onClick={() => setStep(s => s + 1)}
                  disabled={step === 2 && !model}
                  className="px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-700 text-sm font-medium text-white transition-colors disabled:opacity-40">下一步</button>
              )}
              {step === 3 && (
                <button onClick={handleCreate} disabled={isCreating}
                  className="px-5 py-2 rounded-lg bg-green-600 hover:bg-green-700 text-sm font-medium text-white transition-colors disabled:opacity-40">
                  {isCreating ? "创建中…" : "确认创建"}</button>
              )}
              {step === 4 && (
                <button onClick={handleFinish}
                  className="px-5 py-2 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-sm font-medium text-zinc-200 transition-colors">完成</button>
              )}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
