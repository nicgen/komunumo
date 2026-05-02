import { RegisterMemberForm } from "@/components/auth/register-member-form";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Inscription Citoyen | AssoLink",
  description: "Rejoignez AssoLink en tant que citoyen pour participer à la vie associative.",
};

export default function RegisterMemberPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-10rem)] py-8">
      <RegisterMemberForm />
    </div>
  );
}
