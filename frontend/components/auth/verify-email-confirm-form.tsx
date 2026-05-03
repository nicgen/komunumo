"use client";

import { useState } from "react";
import { Button } from "@/components/ui/button";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Loader2, CheckCircle2, XCircle, MailCheck } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthCard } from "./auth-card";

interface VerifyEmailConfirmFormProps {
  token: string;
}

export function VerifyEmailConfirmForm({ token }: VerifyEmailConfirmFormProps) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  async function onConfirm() {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/v1/auth/verify-email", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ token }),
      });

      if (response.ok) {
        setSuccess(true);
        setTimeout(() => {
          router.push("/login?verified=1");
        }, 2000);
      } else {
        const errorData = await response.json();
        setError(errorData.error || "Une erreur est survenue lors de la vérification.");
      }
    } catch {
      setError("Erreur de connexion au serveur.");
    } finally {
      setIsLoading(false);
    }
  }

  if (success) {
    return (
      <AuthCard title="Email vérifié">
        <div className="flex flex-col items-center justify-center space-y-4 py-4">
          <CheckCircle2 className="h-12 w-12 text-green-500" />
          <p className="text-center text-sm text-muted-foreground">
            Votre adresse email a été vérifiée avec succès. Vous allez être redirigé vers la page de connexion.
          </p>
        </div>
      </AuthCard>
    );
  }

  return (
    <AuthCard
      title="Confirmer la vérification"
      description="Veuillez confirmer votre adresse email pour activer votre compte."
    >
      <div className="space-y-6">
        <div className="flex flex-col items-center justify-center space-y-4 py-4">
          <div className="rounded-full bg-primary/10 p-4">
            <MailCheck className="h-8 w-8 text-primary" />
          </div>
        </div>

        {error && (
          <Alert variant="destructive">
            <XCircle className="h-4 w-4" />
            <AlertTitle>Erreur</AlertTitle>
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <Button onClick={onConfirm} className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Vérification...
            </>
          ) : (
            "Confirmer mon adresse email"
          )}
        </Button>

        <div className="text-center text-xs text-muted-foreground">
          Problème de lien?{" "}
          <a href="/verify-email/sent" className="font-medium text-primary hover:underline">
            Demander un nouvel email
          </a>
        </div>
      </div>
    </AuthCard>
  );
}
