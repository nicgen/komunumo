import { VerifyEmailConfirmForm } from "@/components/auth/verify-email-confirm-form";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { XCircle, ArrowLeft } from "lucide-react";

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
        <Alert variant="destructive">
          <XCircle className="h-4 w-4" />
          <AlertTitle>Lien invalide ou expiré</AlertTitle>
          <AlertDescription>
            Le lien de vérification n'est pas valide ou a expiré. Veuillez demander un nouvel
            email de vérification.
          </AlertDescription>
        </Alert>

        <div className="text-center">
          <Button asChild variant="outline">
            <a href="/verify-email/sent" className="inline-flex items-center">
              <ArrowLeft className="mr-2 h-4 w-4" />
              Renvoyer l'email
            </a>
          </Button>
        </div>
      </div>
    );
  }

  return <VerifyEmailConfirmForm token={token} />;
}
