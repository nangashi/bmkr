import { redirect } from "next/navigation";
import { auth, signOut } from "@/auth";

export default async function HomePage() {
	const session = await auth();
	if (!session) {
		redirect("/sign-in");
	}

	return (
		<div>
			<h1>bmkr</h1>
			<p>{session.user?.email}</p>
			<form
				action={async () => {
					"use server";
					await signOut({ redirectTo: "/sign-in" });
				}}
			>
				<button type="submit">ログアウト</button>
			</form>
		</div>
	);
}
