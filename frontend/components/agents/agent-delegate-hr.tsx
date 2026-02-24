"use client";

interface Props {
  hrRequest: string;
  setHRRequest: (v: string) => void;
  hrSent: boolean;
  isSendingToHR: boolean;
  onSend: () => void;
}

export function AgentDelegateHR({ hrRequest, setHRRequest, hrSent, isSendingToHR, onSend }: Props) {
  if (hrSent) {
    return (
      <div className="space-y-3 py-4 text-center">
        <div className="text-4xl">{"\uD83D\uDCCB"}</div>
        <p className="font-medium text-zinc-200">需求已发送给 HR</p>
        <p className="text-sm text-zinc-500">
          Alex Chen 将在 #general 频道处理您的招聘需求，稍后 Agent 会自动出现在列表中。
        </p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-zinc-400">
        描述您需要什么样的 Agent，HR 总监 Alex Chen 将根据需求分配合适的职位并创建。
      </p>
      <textarea
        value={hrRequest} onChange={e => setHRRequest(e.target.value)}
        rows={5}
        placeholder="例如：我们需要一个负责前端开发的工程师，要求熟悉 React 和 TypeScript，需要参与日常 code review 和需求评审…"
        className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 resize-none"
      />
    </div>
  );
}
