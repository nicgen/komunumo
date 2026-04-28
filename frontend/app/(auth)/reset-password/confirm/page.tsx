import { ResetPasswordConfirmForm } from "@/components/auth/reset-password-confirm-form";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { XCircle, ArrowLeft } from "lucide-react";

export const metadata = {
  title: "Réinitialiser le mot de passe",
  description: "Choisissez un nouveau mot de passe pour votre compte Assolink",
};

type PageProps = {
  searchParams: Promise<{ token?: string }>;
};

export default async function ResetPasswordConfirmPage(props: PageProps) {
  const searchParams = await props.searchParams;
  const token = searchParams.token;

  if (!token) {
    return (
      <div className="space-y-4">
        <Alert variant="destructive">
          <XCircle className="h-4 w-4" />
          <AlertTitle>Lien invalide ou expiré</AlertTitle>
          <AlertDescription>
            Le lien de réinitialisation n'est pas valide ou a expiré. Veuillez faire une nouvelle
            demande.
          </AlertDescription>
        </Alert>

        <div className="text-center">
          <Button asChild variant="outline">
            <a href="/forgot-password" className="inline-flex items-center">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Nouvelle demande
            </a>
          </Button>
        </div>
      </div>
    );
  }

  return <ResetPasswordConfirmForm token={token} />;
}
