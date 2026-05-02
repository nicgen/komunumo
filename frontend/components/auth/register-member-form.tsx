"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2 } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthCard } from "./auth-card";

const registerSchema = z.object({
  email: z.string().email({ message: "Adresse email invalide" }),
  first_name: z.string().min(2, { message: "Le prénom doit faire au moins 2 caractères" }),
  last_name: z.string().min(2, { message: "Le nom doit faire au moins 2 caractères" }),
  birth_date: z.string().min(1, { message: "La date de naissance est requise" }),
  password: z
    .string()
    .min(12, { message: "Le mot de passe doit faire au moins 12 caractères" })
    .regex(/[A-Z]/, { message: "Le mot de passe doit contenir au moins une majuscule" })
    .regex(/[a-z]/, { message: "Le mot de passe doit contenir au moins une minuscule" })
    .regex(/[0-9]/, { message: "Le mot de passe doit contenir au moins un chiffre" })
    .regex(/[^A-Za-z0-9]/, { message: "Le mot de passe doit contenir au moins un caractère spécial" }),
});

type RegisterMemberFormValues = z.infer<typeof registerSchema>;

export function RegisterMemberForm() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterMemberFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      first_name: "",
      last_name: "",
      birth_date: "",
      password: "",
    },
  });

  async function onSubmit(data: RegisterMemberFormValues) {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/v1/auth/register/member", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      if (response.ok) {
        router.push("/verify-email/sent");
      } else {
        const errorData = await response.json();
        setError(errorData.error || "Une erreur est survenue lors de l'inscription.");
      }
    } catch (err) {
      setError("Erreur de connexion au serveur.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthCard
      title="Créer mon compte citoyen"
      description="Inscrivez-vous pour rejoindre la communauté AssoLink"
      footer={
        <div className="w-full text-center text-sm text-muted-foreground">
          Vous avez déjà un compte ?{" "}
          <a href="/login" className="font-medium text-primary hover:underline">
            Connectez-vous
          </a>
        </div>
      }
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        {error && (
          <Alert variant="destructive" className="py-2">
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="first_name">Prénom</Label>
            <Input
              id="first_name"
              placeholder="Jean"
              disabled={isLoading}
              aria-describedby={errors.first_name ? "first_name-error" : undefined}
              {...register("first_name")}
              className={errors.first_name ? "border-destructive" : ""}
            />
            {errors.first_name && (
              <p id="first_name-error" className="text-xs text-destructive font-medium mt-1">
                {errors.first_name.message}
              </p>
            )}
          </div>
          <div className="space-y-2">
            <Label htmlFor="last_name">Nom de famille</Label>
            <Input
              id="last_name"
              placeholder="Dupont"
              disabled={isLoading}
              aria-describedby={errors.last_name ? "last_name-error" : undefined}
              {...register("last_name")}
              className={errors.last_name ? "border-destructive" : ""}
            />
            {errors.last_name && (
              <p id="last_name-error" className="text-xs text-destructive font-medium mt-1">
                {errors.last_name.message}
              </p>
            )}
          </div>
        </div>

        <div className="space-y-2">
          <Label htmlFor="email">Adresse email</Label>
          <Input
            id="email"
            type="email"
            placeholder="vous@exemple.com"
            disabled={isLoading}
            aria-describedby={errors.email ? "email-error" : undefined}
            {...register("email")}
            className={errors.email ? "border-destructive" : ""}
          />
          {errors.email && (
            <p id="email-error" className="text-xs text-destructive font-medium mt-1">
              {errors.email.message}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="birth_date">Date de naissance</Label>
          <Input
            id="birth_date"
            type="date"
            disabled={isLoading}
            aria-describedby={errors.birth_date ? "birth_date-error" : undefined}
            {...register("birth_date")}
            className={errors.birth_date ? "border-destructive" : ""}
          />
          {errors.birth_date && (
            <p id="birth_date-error" className="text-xs text-destructive font-medium mt-1">
              {errors.birth_date.message}
            </p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="password">Mot de passe</Label>
          <Input
            id="password"
            type="password"
            placeholder="••••••••••••"
            disabled={isLoading}
            aria-describedby={errors.password ? "password-error" : "password-hint"}
            {...register("password")}
            className={errors.password ? "border-destructive" : ""}
          />
          {errors.password ? (
            <p id="password-error" className="text-xs text-destructive font-medium mt-1">
              {errors.password.message}
            </p>
          ) : (
            <p id="password-hint" className="text-[10px] text-muted-foreground">
              Au moins 12 caractères avec majuscules, minuscules, chiffres et caractères spéciaux.
            </p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Création en cours...
            </>
          ) : (
            "Créer mon compte citoyen"
          )}
        </Button>
      </form>
    </AuthCard>
  );
}
