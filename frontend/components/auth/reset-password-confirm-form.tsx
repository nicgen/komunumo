"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, CheckCircle2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthCard } from "./auth-card";

const resetPasswordConfirmSchema = z.object({
  new_password: z
    .string()
    .min(12, { message: "Le mot de passe doit faire au moins 12 caractères" })
    .regex(/[A-Z]/, { message: "Le mot de passe doit contenir au moins une majuscule" })
    .regex(/[a-z]/, { message: "Le mot de passe doit contenir au moins une minuscule" })
    .regex(/[0-9]/, { message: "Le mot de passe doit contenir au moins un chiffre" })
    .regex(/[^A-Za-z0-9]/, { message: "Le mot de passe doit contenir au moins un caractère spécial" }),
  confirm_password: z.string(),
}).refine((data) => data.new_password === data.confirm_password, {
  message: "Les mots de passe ne correspondent pas",
  path: ["confirm_password"],
});

type ResetPasswordConfirmFormValues = z.infer<typeof resetPasswordConfirmSchema>;

interface ResetPasswordConfirmFormProps {
  token: string;
}

export function ResetPasswordConfirmForm({ token }: ResetPasswordConfirmFormProps) {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetPasswordConfirmFormValues>({
    resolver: zodResolver(resetPasswordConfirmSchema),
    defaultValues: {
      new_password: "",
      confirm_password: "",
    },
  });

  async function onSubmit(data: ResetPasswordConfirmFormValues) {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/v1/auth/password-reset/confirm", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          token,
          new_password: data.new_password,
        }),
      });

      if (response.ok) {
        setSuccess(true);
        setTimeout(() => {
          router.push("/login?password_reset=1");
        }, 2000);
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

  if (success) {
    return (
      <AuthCard title="Mot de passe réinitialisé">
        <div className="flex flex-col items-center justify-center space-y-4 py-4">
          <CheckCircle2 className="h-12 w-12 text-green-500" />
          <p className="text-center text-sm text-muted-foreground">
            Votre mot de passe a été réinitialisé avec succès. Vous allez être redirigé vers la page de connexion.
          </p>
        </div>
      </AuthCard>
    );
  }

  return (
    <AuthCard
      title="Nouveau mot de passe"
      description="Choisissez un nouveau mot de passe pour votre compte."
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <Alert variant="destructive" className="py-2">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="space-y-2">
          <Label htmlFor="new_password">Nouveau mot de passe</Label>
          <Input
            id="new_password"
            type="password"
            placeholder="••••••••••••"
            disabled={isLoading}
            aria-describedby={errors.new_password ? "new_password-error" : "new_password-hint"}
            {...register("new_password")}
            className={errors.new_password ? "border-destructive" : ""}
          />
          {errors.new_password ? (
            <p id="new_password-error" className="text-xs text-destructive">{errors.new_password.message}</p>
          ) : (
            <p id="new_password-hint" className="text-[10px] text-muted-foreground">
              Au moins 12 caractères avec majuscules, minuscules, chiffres et caractères spéciaux.
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="confirm_password">Confirmer le mot de passe</Label>
          <Input
            id="confirm_password"
            type="password"
            placeholder="••••••••••••"
            disabled={isLoading}
            aria-describedby={errors.confirm_password ? "confirm_password-error" : undefined}
            {...register("confirm_password")}
            className={errors.confirm_password ? "border-destructive" : ""}
          />
          {errors.confirm_password && (
            <p id="confirm_password-error" className="text-xs text-destructive">{errors.confirm_password.message}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Réinitialisation...
            </>
          ) : (
            "Réinitialiser le mot de passe"
          )}
        </Button>
      </form>
    </AuthCard>
  );
}
