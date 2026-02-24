"use client";

import { POSITION_LABELS, type Position } from "@/lib/types";

interface Props {
  name: string;
  setName: (v: string) => void;
  position: Position;
  setPosition: (v: Position) => void;
  persona: string;
  setPersona: (v: string) => void;
  hasHR: boolean;
  deptPositions: Record<string, Position[]>;
}

export function AgentStepBasics({
  name, setName, position, setPosition, persona, setPersona, hasHR, deptPositions,
}: Props) {
  return (
    <>
      <div>
        <label className="block text-sm text-zinc-400 mb-1.5">
          Agent 名称<span className="text-zinc-600 ml-1 text-xs">（可选，留空则由 Agent 自行取名）</span>
        </label>
        <input
          value={name} onChange={e => setName(e.target.value)}
          placeholder="留空则 Agent 启动后自动取名"
          className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500"
        />
      </div>

      {!hasHR && (
        <div className="bg-amber-500/10 border border-amber-500/30 rounded-lg p-3 text-sm text-amber-400">
          尚未创建 HR Agent，请先创建一个 HR 来管理后续的 Agent 招聘与部署。
        </div>
      )}

      <div>
        <label className="block text-sm text-zinc-400 mb-1.5">职位 <span className="text-red-400">*</span></label>
        <div className="space-y-2 max-h-52 overflow-y-auto pr-1">
          {Object.entries(deptPositions).map(([dept, positions]) => (
            <div key={dept}>
              <p className="text-xs text-zinc-600 mb-1">{dept}</p>
              <div className="flex flex-wrap gap-1.5">
                {positions.map(p => (
                  <button
                    key={p}
                    onClick={() => setPosition(p)}
                    className={`px-2.5 py-1 rounded text-xs transition-colors ${
                      position === p
                        ? "bg-blue-600 text-white"
                        : "bg-zinc-800 text-zinc-400 hover:bg-zinc-700"
                    }`}
                  >
                    {POSITION_LABELS[p]}
                  </button>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>

      <div>
        <label className="block text-sm text-zinc-400 mb-1.5">人设描述（可选）</label>
        <textarea
          value={persona} onChange={e => setPersona(e.target.value)}
          rows={3} placeholder="描述 Agent 的性格、专长和行为方式…"
          className="w-full bg-zinc-800 border border-zinc-700 rounded-lg px-3 py-2 text-sm text-zinc-100 placeholder-zinc-600 focus:outline-none focus:border-zinc-500 resize-none"
        />
      </div>
    </>
  );
}
