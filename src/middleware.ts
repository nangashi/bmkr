export { auth as middleware } from "@/auth";

export const config = {
	matcher: ["/((?!api/ping|_next/static|_next/image|favicon.ico).*)"],
};
