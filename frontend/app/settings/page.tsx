"use client";

import { Shell } from "@/components/layout/shell";
import { CompanyInfo } from "@/components/settings/company-info";
import { Preferences } from "@/components/settings/preferences";
import { ChangePassword } from "@/components/settings/change-password";
import { SystemSettings } from "@/components/settings/system-settings";

export default function SettingsPage() {
  return (
    <Shell>
      <div className="space-y-6">
        <div>
          <h1 className="text-2xl font-semibold text-zinc-50">设置</h1>
          <p className="text-zinc-400 text-sm mt-1">平台配置与公司信息</p>
        </div>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <CompanyInfo />
          <Preferences />
          <ChangePassword />
          <div className="lg:col-span-2">
            <SystemSettings />
          </div>
        </div>
      </div>
    </Shell>
  );
}
