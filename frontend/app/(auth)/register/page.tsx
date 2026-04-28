import { RegisterForm } from "@/components/auth/register-form";

export const metadata = {
  title: "Créer un compte",
  description: "Inscription pour créer votre compte Assolink",
};

export default function RegisterPage() {
  return <RegisterForm />;
}
