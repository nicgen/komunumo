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

const loginSchema = z.object({
  email: z.string().email({ message: "Adresse email invalide" }),
  password: z.string().min(1, { message: "Le mot de passe est requis" }),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginForm() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  async function onSubmit(data: LoginFormValues) {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/v1/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      if (response.ok) {
        router.push("/home");
        router.refresh();
      } else {
        const errorData = await response.json();
        setError(errorData.error || "Une erreur est survenue lors de la connexion.");
      }
    } catch {
      setError("Erreur de connexion au serveur.");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <AuthCard
      title="Connexion"
      description="Entrez vos identifiants pour accéder à votre compte"
      footer={
        <div className="w-full text-center text-sm text-muted-foreground">
          Pas encore de compte?{" "}
          <a href="/register" className="font-medium text-primary hover:underline">
            Créer un compte
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
            <p id="email-error" className="text-xs text-destructive">{errors.email.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <Label htmlFor="password">Mot de passe</Label>
            <a
              href="/forgot-password"
              className="text-xs font-medium text-primary hover:underline"
            >
              Mot de passe oublié?
            </a>
          </div>
          <Input
            id="password"
            type="password"
            placeholder="••••••••••••"
            disabled={isLoading}
            aria-describedby={errors.password ? "password-error" : undefined}
            {...register("password")}
            className={errors.password ? "border-destructive" : ""}
          />
          {errors.password && (
            <p id="password-error" className="text-xs text-destructive">{errors.password.message}</p>
          )}
        </div>

        <Button type="submit" className="w-full" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Connexion en cours...
            </>
          ) : (
            "Se connecter"
          )}
        </Button>
      </form>
    </AuthCard>
  );
}
