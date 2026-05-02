import { LoginForm } from "@/components/auth/login-form";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { CheckCircle2 } from "lucide-react";

export const metadata = {
  title: "Connexion",
  description: "Connectez-vous à votre compte Assolink",
};

type PageProps = {
  searchParams: Promise<{ verified?: string; password_reset?: string }>;
};

export default async function LoginPage(props: PageProps) {
  const searchParams = await props.searchParams;
  const verified = searchParams.verified === "1";
  const passwordReset = searchParams.password_reset === "1";

  return (
    <div className="space-y-4">
      {verified && (
        <Alert className="border-green-200 bg-green-50/50 text-green-800">
          <CheckCircle2 className="h-4 w-4 text-green-600" />
          <AlertDescription>
            Votre adresse email a été vérifiée. Vous pouvez maintenant vous connecter.
          </AlertDescription>
        </Alert>
      )}

      {passwordReset && (
        <Alert className="border-green-200 bg-green-50/50 text-green-800">
          <CheckCircle2 className="h-4 w-4 text-green-600" />
          <AlertDescription>
            Votre mot de passe a été réinitialisé. Vous pouvez maintenant vous connecter.
          </AlertDescription>
        </Alert>
      )}

      <LoginForm />
    </div>
  );
}
