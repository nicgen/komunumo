"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, Mail, CheckCircle2 } from "lucide-react";
import { AuthCard } from "./auth-card";

const resendSchema = z.object({
  email: z.string().email({ message: "Adresse email invalide" }),
});

type ResendFormValues = z.infer<typeof resendSchema>;

export function ResendVerificationForm() {
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResendFormValues>({
    resolver: zodResolver(resendSchema),
    defaultValues: {
      email: "",
    },
  });

  async function onSubmit(data: ResendFormValues) {
    setIsLoading(true);
    setError(null);
    setSuccess(false);

    try {
      const response = await fetch("/api/v1/auth/resend-verification", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      if (response.ok) {
        setSuccess(true);
      } else {
        const errorData = await response.json();
        setError(errorData.error || "Une erreur est survenue.");
      }
    } catch (err) {
      setError("Erreur de connexion au serveur.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthCard
      title="Vérifiez votre messagerie"
      description="Nous avons envoyé un lien de vérification à votre adresse email. Veuillez cliquer sur ce lien pour activer votre compte."
      footer={
        <div className="w-full text-center text-sm text-muted-foreground">
          <a href="/login" className="font-medium text-primary hover:underline">
            Retour à la connexion
          </a>
        </div>
      }
    >
      <div className="space-y-6">
        <div className="flex flex-col items-center justify-center space-y-2 py-2">
          <div className="rounded-full bg-primary/10 p-3">
            <Mail className="h-6 w-6 text-primary" />
          </div>
          <p className="text-center text-xs text-muted-foreground">
            Vérifiez également votre dossier de spam.
          </p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 pt-4 border-t">
          {error && (
            <Alert variant="destructive" className="py-2">
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}

          {success && (
            <Alert className="border-green-200 bg-green-50/50 text-green-800 py-2">
              <CheckCircle2 className="h-4 w-4 text-green-600" />
              <AlertDescription>
                Un nouvel email de vérification a été envoyé.
              </AlertDescription>
            </Alert>
          )}

          <div className="space-y-2">
            <Label htmlFor="email" className="text-xs">Renvoyer l'email de vérification à</Label>
            <Input
              id="email"
              type="email"
              placeholder="vous@exemple.com"
              disabled={isLoading}
              {...register("email")}
              className={errors.email ? "border-destructive" : ""}
            />
            {errors.email && (
              <p className="text-xs text-destructive">{errors.email.message}</p>
            )}
          </div>

          <Button type="submit" variant="secondary" className="w-full" disabled={isLoading}>
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Envoi en cours...
              </>
            ) : (
              "Renvoyer l'email"
            )}
          </Button>
        </form>
      </div>
    </AuthCard>
  );
}
