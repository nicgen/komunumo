export const metadata = {
  title: "Confirmer la vérification",
  description: "Confirmation de votre adresse email",
};

type PageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function VerifyEmailConfirmPage(props: PageProps) {
  const searchParams = await props.searchParams;
  const token = searchParams.token;

  if (!token) {
    return (
      <div className="space-y-4">
        <div className="rounded-md border border-red-200 bg-red-50 p-4">
          <h1 className="text-lg font-semibold text-red-900">Lien invalide ou expiré</h1>
          <p className="mt-2 text-sm text-red-700">
            Le lien de vérification n'est pas valide ou a expiré. Veuillez demander un nouvel
            email de vérification.
          </p>
        </div>

        <div className="text-center">
          <a
            href="/verify-email/sent"
            className="inline-block rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-colors"
          >
            Renvoyer l'email de vérification
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold text-foreground">Confirmer la vérification</h1>
        <p className="text-sm text-muted-foreground">
          Veuillez confirmer votre adresse email en cliquant sur le bouton ci-dessous.
        </p>
      </div>

      <form method="POST" action="/api/v1/auth/verify-email" className="space-y-4">
        <input type="hidden" name="token" value={token} />

        <button
          type="submit"
          className="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-colors"
        >
          Confirmer la vérification
        </button>
      </form>

      <noscript>
        <p className="text-xs text-muted-foreground text-center">
          JavaScript est désactivé. Veuillez cliquer sur le bouton ci-dessus pour continuer.
        </p>
      </noscript>

      <div className="text-center text-sm text-muted-foreground">
        Problème de lien?{" "}
        <a href="/verify-email/sent" className="font-medium text-primary hover:underline">
          Demander un nouvel email
        </a>
      </div>
    </div>
  );
}
