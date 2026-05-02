import { Button } from "@/components/ui/button";
import { AuthCard } from "@/components/auth/auth-card";
import Link from "next/link";
import { User, Building2 } from "lucide-react";
import { Metadata } from "next";

export const metadata: Metadata = {
  title: "Choisissez votre type de compte | AssoLink",
  description: "Inscrivez-vous en tant que citoyen ou association sur AssoLink.",
};

export default function RegisterChoicePage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-[calc(100vh-10rem)] py-8 px-4">
      <AuthCard
        title="Créer un compte"
        description="Quel type de compte souhaitez-vous créer ?"
        footer={
          <div className="w-full text-center text-sm text-muted-foreground">
            Vous avez déjà un compte ?{" "}
            <Link href="/login" className="font-medium text-primary hover:underline">
              Connectez-vous
            </Link>
          </div>
        }
      >
        <div className="grid grid-cols-1 gap-4">
          <Button asChild variant="outline" className="h-auto py-6 flex flex-col items-center justify-center space-y-2 hover:border-primary hover:bg-primary/5 group transition-all">
            <Link href="/register/member">
              <User className="h-8 w-8 text-muted-foreground group-hover:text-primary transition-colors" />
              <div className="flex flex-col items-center text-center">
                <span className="font-semibold text-base">Je suis un Citoyen</span>
                <span className="text-xs text-muted-foreground font-normal mt-1">
                  Pour participer à la vie associative et découvrir des activités.
                </span>
              </div>
            </Link>
          </Button>

          <Button asChild variant="outline" className="h-auto py-6 flex flex-col items-center justify-center space-y-2 hover:border-primary hover:bg-primary/5 group transition-all">
            <Link href="/register/association">
              <Building2 className="h-8 w-8 text-muted-foreground group-hover:text-primary transition-colors" />
              <div className="flex flex-col items-center text-center">
                <span className="font-semibold text-base">Je suis une Association</span>
                <span className="text-xs text-muted-foreground font-normal mt-1">
                  Pour gérer votre structure, vos membres et vos événements.
                </span>
              </div>
            </Link>
          </Button>
        </div>
      </AuthCard>
    </div>
  );
}
