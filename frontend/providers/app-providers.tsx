"use client";

import { ReactNode, useEffect, useState } from "react";
import { ThemeProvider } from "next-themes";
import { NextIntlClientProvider, type AbstractIntlMessages } from "next-intl";
import { Toaster } from "sonner";
import { AuthGuard } from "@/components/auth/auth-guard";
import zhMessages from "../messages/zh.json";

async function loadMessages(locale: string): Promise<AbstractIntlMessages> {
  try {
    const mod = await import(`../messages/${locale}.json`);
    return mod.default as AbstractIntlMessages;
  } catch {
    return zhMessages as AbstractIntlMessages;
  }
}

export function AppProviders({ children }: { children: ReactNode }) {
  const [locale, setLocale] = useState("zh");
  const [messages, setMessages] = useState<AbstractIntlMessages>(zhMessages as AbstractIntlMessages);

  useEffect(() => {
    const saved = localStorage.getItem("lc_locale") || "zh";
    setLocale(saved);
    loadMessages(saved).then(setMessages);
  }, []);

  return (
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem={false}>
      <NextIntlClientProvider locale={locale} messages={messages}>
        <Toaster
          theme="dark"
          position="top-right"
          toastOptions={{
            className: "!bg-zinc-900 !border-zinc-700 !text-zinc-50",
          }}
        />
        <AuthGuard>{children}</AuthGuard>
      </NextIntlClientProvider>
    </ThemeProvider>
  );
}
