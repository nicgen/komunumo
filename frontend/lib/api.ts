/**
 * Typed fetch wrapper for the Komunumo backend.
 *
 * - Uses relative URLs so requests go through the Next rewrite to the backend.
 * - Sends cookies (credentials: include) for session + CSRF.
 * - On unsafe methods, reads `__Host-csrf` cookie and echoes it as X-CSRF-Token.
 *
 * Server components should NOT use this directly — they read cookies via
 * `next/headers` and call the backend internally; see `lib/auth.ts`.
 */
export type ApiResult<T> =
  | { ok: true; status: number; data: T }
  | { ok: false; status: number; error: { code: string; message: string; fields?: Record<string, string> } };

const UNSAFE = new Set(["POST", "PUT", "PATCH", "DELETE"]);

function readCsrfCookie(): string | null {
  if (typeof document === "undefined") return null;
  const m = document.cookie.match(/(?:^|;\s*)__Host-csrf=([^;]+)/);
  return m ? decodeURIComponent(m[1]) : null;
}

export async function apiFetch<T>(
  path: string,
  init: RequestInit = {},
): Promise<ApiResult<T>> {
  const method = (init.method ?? "GET").toUpperCase();
  const headers = new Headers(init.headers);
  if (!headers.has("Accept")) headers.set("Accept", "application/json");
  if (init.body && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (UNSAFE.has(method)) {
    const csrf = readCsrfCookie();
    if (csrf) headers.set("X-CSRF-Token", csrf);
  }

  let res: Response;
  try {
    res = await fetch(path, { ...init, method, headers, credentials: "include" });
  } catch (cause) {
    return {
      ok: false,
      status: 0,
      error: { code: "network_error", message: cause instanceof Error ? cause.message : "network error" },
    };
  }

  const contentType = res.headers.get("Content-Type") ?? "";
  const isJSON = contentType.includes("application/json");
  const body = isJSON ? await res.json().catch(() => null) : null;

  if (res.ok) {
    return { ok: true, status: res.status, data: (body as T) };
  }
  return {
    ok: false,
    status: res.status,
    error: {
      code: body?.code ?? `http_${res.status}`,
      message: body?.message ?? res.statusText,
      fields: body?.fields,
    },
  };
}
