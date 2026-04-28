import { AuthCard } from "@/components/auth/auth-card";
import { Mail, ArrowLeft } from "lucide-react";

export const metadata = {
  title: "Email envoyé",
  description: "Vérifiez votre messagerie pour réinitialiser votre mot de passe",
};

export default function ResetPasswordSentPage() {
  return (
    <AuthCard
      title="Vérifiez votre messagerie"
      description="Si un compte existe avec cette adresse email, vous recevrez un lien de réinitialisation dans les prochaines minutes."
      footer={
        <div className="w-full text-center text-sm text-muted-foreground">
          <a href="/login" className="inline-flex items-center font-medium text-primary hover:underline">
            <ArrowLeft className="mr-2 h-4 w-4" />
            Retour à la connexion
          </a>
        </div>
      }
    >
      <div className="flex flex-col items-center justify-center space-y-4 py-4">
        <div className="rounded-full bg-primary/10 p-3">
          <Mail className="h-6 w-6 text-primary" />
        </div>
        <p className="text-center text-sm text-muted-foreground">
          Vérifiez également votre dossier de spam. Le lien est valable pendant 1 heure.
        </p>
      </div>
    </AuthCard>
  );
}
