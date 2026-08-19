"use client";

import { legend } from "@/lib/copy";

export function Legend() {
  return (
    <div className="hidden shrink-0 flex-wrap gap-x-2.5 gap-y-1 px-3 pt-1 text-[11px] text-muted sm:flex sm:px-4">
      {legend.map((item) => (
        <span key={item.label} className="inline-flex items-center gap-1.5">
          <i className="inline-block h-2.5 w-2.5 rounded-sm" style={{ background: item.color }} />
          {item.label}
        </span>
      ))}
    </div>
  );
}
