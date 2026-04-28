import { ForgotPasswordForm } from "@/components/auth/forgot-password-form";

export const metadata = {
  title: "Mot de passe oublié",
  description: "Réinitialisez votre mot de passe Assolink",
};

export default function ForgotPasswordPage() {
  return <ForgotPasswordForm />;
}
