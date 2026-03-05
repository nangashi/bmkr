import { signIn } from "@/auth";

export default function SignInPage() {
	return (
		<div>
			<h1>ログイン</h1>
			<form
				action={async () => {
					"use server";
					await signIn("google", { redirectTo: "/" });
				}}
			>
				<button type="submit">Google でログイン</button>
			</form>
		</div>
	);
}
