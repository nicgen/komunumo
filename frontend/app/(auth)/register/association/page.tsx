import { RegisterAssociationForm } from "@/components/auth/register-association-form";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Inscription Association | AssoLink",
  description: "Inscrivez votre association sur AssoLink pour gérer vos membres et activités.",
};

export default function RegisterAssociationPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-10rem)] py-8">
      <RegisterAssociationForm />
    </div>
  );
}
