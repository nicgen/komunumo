export const metadata = {
  title: "Vérifiez votre messagerie",
  description: "Confirmez votre adresse email pour activer votre compte",
};

export default function VerifyEmailSentPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold text-foreground">Vérifiez votre messagerie</h1>
        <p className="text-sm text-muted-foreground">
          Nous avons envoyé un lien de vérification à votre adresse email. Veuillez cliquer sur ce
          lien pour activer votre compte.
        </p>
      </div>

      <div className="rounded-md border border-border bg-muted/50 p-4">
        <p className="text-sm text-muted-foreground">
          Si vous ne recevez pas l'email dans les prochaines minutes, vérifiez votre dossier de
          spam ou utilisez le formulaire ci-dessous pour renvoyer l'email de vérification.
        </p>
      </div>

      <form method="POST" action="/api/v1/auth/resend-verification" className="space-y-4">
        <div className="space-y-2">
          <label htmlFor="resend-email" className="block text-sm font-medium text-foreground">
            Renvoyer l'email de vérification à
          </label>
          <input
            id="resend-email"
            name="email"
            type="email"
            required
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            placeholder="vous@exemple.com"
          />
        </div>

        <button
          type="submit"
          className="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-colors"
        >
          Renvoyer l'email de vérification
        </button>
      </form>

      <div className="text-center text-sm text-muted-foreground">
        <a href="/register" className="font-medium text-primary hover:underline">
          Créer un nouveau compte
        </a>
      </div>
    </div>
  );
}
