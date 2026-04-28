import { ResendVerificationForm } from "@/components/auth/resend-verification-form";

export const metadata = {
  title: "Vérifiez votre messagerie",
  description: "Confirmez votre adresse email pour activer votre compte",
};

export default function VerifyEmailSentPage() {
  return <ResendVerificationForm />;
}
