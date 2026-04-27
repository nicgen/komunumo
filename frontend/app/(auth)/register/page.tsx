export const metadata = {
  title: "Créer un compte",
  description: "Inscription pour créer votre compte Assolink",
};

export default function RegisterPage() {
  return (
    <div className="space-y-6">
      <div className="space-y-2">
        <h1 className="text-2xl font-bold text-foreground">Créer mon compte</h1>
        <p className="text-sm text-muted-foreground">
          Remplissez le formulaire ci-dessous pour vous inscrire
        </p>
      </div>

      <form method="POST" action="/api/v1/auth/register" className="space-y-4">
        <div className="space-y-2">
          <label htmlFor="email" className="block text-sm font-medium text-foreground">
            Adresse email
          </label>
          <input
            id="email"
            name="email"
            type="email"
            required
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            placeholder="vous@exemple.com"
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="first_name" className="block text-sm font-medium text-foreground">
            Prénom
          </label>
          <input
            id="first_name"
            name="first_name"
            type="text"
            required
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            placeholder="Jean"
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="last_name" className="block text-sm font-medium text-foreground">
            Nom de famille
          </label>
          <input
            id="last_name"
            name="last_name"
            type="text"
            required
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            placeholder="Dupont"
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="date_of_birth" className="block text-sm font-medium text-foreground">
            Date de naissance
          </label>
          <input
            id="date_of_birth"
            name="date_of_birth"
            type="date"
            required
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
          />
        </div>

        <div className="space-y-2">
          <label htmlFor="password" className="block text-sm font-medium text-foreground">
            Mot de passe
          </label>
          <input
            id="password"
            name="password"
            type="password"
            required
            aria-describedby="password-hint"
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2"
            placeholder="••••••••••••"
          />
          <p id="password-hint" className="text-xs text-muted-foreground">
            Au moins 12 caractères avec majuscules, minuscules, chiffres et caractères spéciaux.
          </p>
        </div>

        <button
          type="submit"
          className="w-full rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 transition-colors"
        >
          Créer mon compte
        </button>
      </form>

      <div className="text-center text-sm text-muted-foreground">
        Vous avez déjà un compte?{" "}
        <a href="/login" className="font-medium text-primary hover:underline">
          Connectez-vous
        </a>
      </div>
    </div>
  );
}
