import NextAuth from "next-auth";
import Google from "next-auth/providers/google";

export const { handlers, auth, signIn, signOut } = NextAuth({
	providers: [Google],
	callbacks: {
		signIn({ profile }) {
			if (profile?.email === process.env.ALLOWED_EMAIL) {
				return true;
			}
			console.error(
				`Sign-in rejected: ${profile?.email ?? "unknown"} is not allowed`,
			);
			return false;
		},
	},
	pages: {
		signIn: "/sign-in",
	},
});
