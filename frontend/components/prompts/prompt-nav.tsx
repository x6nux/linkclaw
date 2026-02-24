"use client";

import { cn } from "@/lib/utils";
import {
  DEPARTMENTS,
  DEPARTMENT_POSITIONS,
  POSITION_LABELS,
  type PromptListResponse,
} from "@/lib/types";

export type PromptSelection =
  | { type: "global"; key: "_" }
  | { type: "department"; key: string }
  | { type: "position"; key: string }
  | { type: "agent"; key: string; name: string };

interface Props {
  data: PromptListResponse | null;
  selected: PromptSelection | null;
  onSelect: (sel: PromptSelection) => void;
}

function HasContent({ has }: { has: boolean }) {
  if (!has) return null;
  return <span className="w-1.5 h-1.5 rounded-full bg-blue-400 flex-shrink-0" />;
}

function NavItem({
  label,
  active,
  has,
  onClick,
  indent,
}: {
  label: string;
  active: boolean;
  has: boolean;
  onClick: () => void;
  indent?: boolean;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "w-full text-left text-sm px-3 py-1.5 rounded-md flex items-center gap-2 transition-colors",
        indent && "pl-7",
        active
          ? "bg-blue-500/10 text-blue-400"
          : "text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60"
      )}
    >
      <span className="truncate flex-1">{label}</span>
      <HasContent has={has} />
    </button>
  );
}

export function PromptNav({ data, selected, onSelect }: Props) {
  const globalHas = !!(data?.global);

  return (
    <div className="h-full border-r border-zinc-800 flex flex-col">
      <div className="px-4 h-12 flex items-center border-b border-zinc-800">
        <h2 className="text-sm font-semibold text-zinc-200">提示词管理</h2>
      </div>

      <div className="flex-1 overflow-y-auto py-2 px-2 space-y-3">
        {/* 全局 */}
        <div>
          <NavItem
            label="全局提示词"
            active={selected?.type === "global"}
            has={globalHas}
            onClick={() => onSelect({ type: "global", key: "_" })}
          />
        </div>

        {/* 部门 */}
        <div>
          <p className="px-3 py-1 text-xs font-medium text-zinc-500 uppercase tracking-wider">部门提示词</p>
          {DEPARTMENTS.map((dept) => (
            <NavItem
              key={dept}
              label={dept}
              indent
              active={selected?.type === "department" && selected.key === dept}
              has={!!(data?.departments[dept])}
              onClick={() => onSelect({ type: "department", key: dept })}
            />
          ))}
        </div>

        {/* 职位 */}
        <div>
          <p className="px-3 py-1 text-xs font-medium text-zinc-500 uppercase tracking-wider">职位提示词</p>
          {Object.entries(DEPARTMENT_POSITIONS).map(([dept, positions]) => (
            <div key={dept}>
              <p className="px-3 pt-1.5 pb-0.5 text-xs text-zinc-600">{dept}</p>
              {positions.map((pos) => (
                <NavItem
                  key={pos}
                  label={POSITION_LABELS[pos] ?? pos}
                  indent
                  active={selected?.type === "position" && selected.key === pos}
                  has={!!(data?.positions[pos])}
                  onClick={() => onSelect({ type: "position", key: pos })}
                />
              ))}
            </div>
          ))}
        </div>

        {/* Agent 专属 */}
        {data?.agents && data.agents.length > 0 && (
          <div>
            <p className="px-3 py-1 text-xs font-medium text-zinc-500 uppercase tracking-wider">Agent 专属</p>
            {data.agents.map((agent) => (
              <NavItem
                key={agent.id}
                label={agent.name || agent.id.slice(0, 8)}
                indent
                active={selected?.type === "agent" && selected.key === agent.id}
                has={!!agent.persona}
                onClick={() =>
                  onSelect({ type: "agent", key: agent.id, name: agent.name })
                }
              />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
