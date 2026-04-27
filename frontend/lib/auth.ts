import "server-only";
import { cookies, headers } from "next/headers";

const SESSION_COOKIE = "__Host-session";

export type CurrentUser = {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  status: "pending_verification" | "verified" | "disabled";
};

/**
 * Server-side helper. Reads the session cookie and asks the backend who is
 * logged in. Returns null on missing/expired session.
 *
 * Use only in server components or route handlers.
 */
export async function getCurrentUser(): Promise<CurrentUser | null> {
  const cookieStore = await cookies();
  const session = cookieStore.get(SESSION_COOKIE);
  if (!session) return null;

  const internal = process.env.KOMUNUMO_API_INTERNAL_URL;
  if (!internal) {
    throw new Error("KOMUNUMO_API_INTERNAL_URL is required for getCurrentUser()");
  }

  const fwd = await headers();
  const res = await fetch(`${internal}/api/v1/auth/me`, {
    method: "GET",
    headers: {
      Cookie: `${SESSION_COOKIE}=${session.value}`,
      "X-Forwarded-For": fwd.get("x-forwarded-for") ?? "",
      "User-Agent": fwd.get("user-agent") ?? "",
      Accept: "application/json",
    },
    cache: "no-store",
  });

  if (res.status === 401 || res.status === 404) return null;
  if (!res.ok) {
    throw new Error(`auth.getCurrentUser: backend ${res.status}`);
  }
  return (await res.json()) as CurrentUser;
}
