"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { toast } from "sonner";
import { Loader2 } from "lucide-react";

const profileSchema = z.object({
  about: z.string().max(2000, { message: "Maximum 2000 caractères" }).optional(),
  postal_code: z.string().regex(/^\d{5}$/, { message: "Code postal invalide (5 chiffres)" }),
  visibility: z.enum(["public", "members_only", "private"]),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

interface AssociationProfileFormProps {
  initialData: {
    about?: string;
    postal_code: string;
    visibility: string;
  };
}

export function AssociationProfileForm({ initialData }: AssociationProfileFormProps) {
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isDirty },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      about: initialData.about || "",
      postal_code: initialData.postal_code,
      visibility: (initialData.visibility as any) || "public",
    },
  });

  async function onSubmit(data: ProfileFormValues) {
    setIsLoading(true);
    try {
      const response = await fetch("/api/v1/me/profile", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
      });

      if (response.ok) {
        toast.success("Profil mis à jour");
      } else {
        const err = await response.json();
        toast.error(err.error || "Une erreur est survenue");
      }
    } catch (err) {
      toast.error("Erreur de connexion");
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <div className="space-y-2">
        <Label htmlFor="about">À propos de l'association</Label>
        <textarea
          id="about"
          placeholder="Décrivez l'objet de votre association..."
          className="flex min-h-[150px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          disabled={isLoading}
          {...register("about")}
        />
        <p className="text-[10px] text-muted-foreground text-right">
          {watch("about")?.length || 0}/2000
        </p>
        {errors.about && (
          <p className="text-xs text-destructive font-medium mt-1">{errors.about.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="postal_code">Code Postal</Label>
        <Input
          id="postal_code"
          placeholder="75011"
          disabled={isLoading}
          {...register("postal_code")}
        />
        {errors.postal_code && (
          <p className="text-xs text-destructive font-medium mt-1">{errors.postal_code.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="visibility">Visibilité du profil</Label>
        <select
          id="visibility"
          className="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          {...register("visibility")}
          disabled={isLoading}
        >
          <option value="public">Public (tout le monde)</option>
          <option value="members_only">Membres seulement</option>
          <option value="private">Privé (moi uniquement)</option>
        </select>
      </div>

      <Button type="submit" disabled={isLoading || !isDirty} className="w-full">
        {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        Enregistrer les modifications
      </Button>
    </form>
  );
}
