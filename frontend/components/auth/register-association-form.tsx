"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Loader2, Info } from "lucide-react";
import { useRouter } from "next/navigation";
import { AuthCard } from "./auth-card";
import { Separator } from "@/components/ui/separator";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

const registerSchema = z.object({
  email: z.string().email({ message: "Adresse email invalide" }),
  password: z
    .string()
    .min(12, { message: "Le mot de passe doit faire au moins 12 caractères" })
    .regex(/[A-Z]/, { message: "Le mot de passe doit contenir au moins une majuscule" })
    .regex(/[a-z]/, { message: "Le mot de passe doit contenir au moins une minuscule" })
    .regex(/[0-9]/, { message: "Le mot de passe doit contenir au moins un chiffre" })
    .regex(/[^A-Za-z0-9]/, { message: "Le mot de passe doit contenir au moins un caractère spécial" }),
  legal_name: z.string().min(2, { message: "Le nom légal est requis" }),
  postal_code: z.string().regex(/^\d{5}$/, { message: "Code postal invalide (5 chiffres)" }),
  siren: z.string().regex(/^\d{9}$/, { message: "SIREN doit faire 9 chiffres" }).optional().or(z.literal("")),
  rna: z.string().regex(/^W\d{9}$/, { message: "RNA doit commencer par W suivi de 9 chiffres" }).optional().or(z.literal("")),
  first_name: z.string().min(2, { message: "Le prénom est requis" }),
  last_name: z.string().min(2, { message: "Le nom est requis" }),
  birth_date: z.string().min(1, { message: "La date de naissance est requise" }),
});

type RegisterAssociationFormValues = z.infer<typeof registerSchema>;

export function RegisterAssociationForm() {
  const router = useRouter();
  const [error, setError] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterAssociationFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
      legal_name: "",
      postal_code: "",
      siren: "",
      rna: "",
      first_name: "",
      last_name: "",
      birth_date: "",
    },
  });

  async function onSubmit(data: RegisterAssociationFormValues) {
    setIsLoading(true);
    setError(null);

    // Filter out empty strings for optional fields
    const payload = {
      ...data,
      siren: data.siren || undefined,
      rna: data.rna || undefined,
    };

    try {
      const response = await fetch("/api/v1/auth/register/association", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
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
      title="Inscrire mon association"
      description="Créez un compte pour votre structure sur AssoLink"
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

        <div className="space-y-4">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[10px] text-primary">1</span>
            Informations sur l'association
          </h3>
          
          <div className="space-y-2">
            <Label htmlFor="legal_name">Nom légal de l'association</Label>
            <Input
              id="legal_name"
              placeholder="Ex: Les Amis du Code"
              disabled={isLoading}
              aria-describedby={errors.legal_name ? "legal_name-error" : undefined}
              {...register("legal_name")}
              className={errors.legal_name ? "border-destructive" : ""}
            />
            {errors.legal_name && (
              <p id="legal_name-error" className="text-xs text-destructive font-medium mt-1">
                {errors.legal_name.message}
              </p>
            )}
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="postal_code">Code Postal</Label>
              <Input
                id="postal_code"
                placeholder="75011"
                disabled={isLoading}
                aria-describedby={errors.postal_code ? "postal_code-error" : undefined}
                {...register("postal_code")}
                className={errors.postal_code ? "border-destructive" : ""}
              />
              {errors.postal_code && (
                <p id="postal_code-error" className="text-xs text-destructive font-medium mt-1">
                  {errors.postal_code.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <Label htmlFor="email">Email de l'association</Label>
              <Input
                id="email"
                type="email"
                placeholder="contact@asso.fr"
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
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <div className="flex items-center gap-1.5">
                <Label htmlFor="siren">Numéro SIREN (optionnel)</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="text-xs">9 chiffres identifiant votre structure au répertoire SIRENE</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <Input
                id="siren"
                placeholder="123456789"
                disabled={isLoading}
                aria-describedby={errors.siren ? "siren-error" : undefined}
                {...register("siren")}
                className={errors.siren ? "border-destructive" : ""}
              />
              {errors.siren && (
                <p id="siren-error" className="text-xs text-destructive font-medium mt-1">
                  {errors.siren.message}
                </p>
              )}
            </div>
            <div className="space-y-2">
              <div className="flex items-center gap-1.5">
                <Label htmlFor="rna">Numéro RNA (optionnel)</Label>
                <TooltipProvider>
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Info className="h-3 w-3 text-muted-foreground cursor-help" />
                    </TooltipTrigger>
                    <TooltipContent>
                      <p className="text-xs">Identifiant commençant par W suivi de 9 chiffres</p>
                    </TooltipContent>
                  </Tooltip>
                </TooltipProvider>
              </div>
              <Input
                id="rna"
                placeholder="W123456789"
                disabled={isLoading}
                aria-describedby={errors.rna ? "rna-error" : undefined}
                {...register("rna")}
                className={errors.rna ? "border-destructive" : ""}
              />
              {errors.rna && (
                <p id="rna-error" className="text-xs text-destructive font-medium mt-1">
                  {errors.rna.message}
                </p>
              )}
            </div>
          </div>
        </div>

        <Separator className="my-6" />

        <div className="space-y-4">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[10px] text-primary">2</span>
            Le représentant légal
          </h3>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="first_name">Prénom</Label>
              <Input
                id="first_name"
                placeholder="Anne"
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
              <Label htmlFor="last_name">Nom</Label>
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
        </div>

        <Separator className="my-6" />

        <div className="space-y-4">
          <h3 className="text-sm font-semibold flex items-center gap-2">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/10 text-[10px] text-primary">3</span>
            Sécurité du compte
          </h3>

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
        </div>

        <Button type="submit" className="w-full mt-6" disabled={isLoading}>
          {isLoading ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Création en cours...
            </>
          ) : (
            "Inscrire l'association"
          )}
        </Button>
      </form>
    </AuthCard>
  );
}
