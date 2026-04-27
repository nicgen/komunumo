import type { ReactNode } from "react";

export default function AuthLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:top-2 focus:left-2 focus:z-50 focus:rounded focus:bg-foreground focus:px-3 focus:py-2 focus:text-background focus:outline-none focus:ring-2 focus:ring-ring"
      >
        Aller au contenu principal
      </a>
      <main
        id="main-content"
        className="flex min-h-svh flex-col items-center justify-center px-4 py-12"
      >
        <div className="w-full max-w-sm">{children}</div>
      </main>
    </>
  );
}
