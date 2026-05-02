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
  nickname: z.string().max(50).optional(),
  about_me: z.string().max(500, { message: "Maximum 500 caractères" }).optional(),
  visibility: z.enum(["public", "members_only", "private"]),
});

type ProfileFormValues = z.infer<typeof profileSchema>;

interface MemberProfileFormProps {
  initialData: {
    nickname?: string;
    about_me?: string;
    visibility: string;
  };
}

export function MemberProfileForm({ initialData }: MemberProfileFormProps) {
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors, isDirty },
  } = useForm<ProfileFormValues>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      nickname: initialData.nickname || "",
      about_me: initialData.about_me || "",
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
        <Label htmlFor="nickname">Pseudo</Label>
        <Input
          id="nickname"
          placeholder="Votre pseudo"
          disabled={isLoading}
          {...register("nickname")}
        />
        {errors.nickname && (
          <p className="text-xs text-destructive font-medium mt-1">{errors.nickname.message}</p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="about_me">À propos de moi</Label>
        <textarea
          id="about_me"
          placeholder="Dites-nous en plus sur vous..."
          className="flex min-h-[120px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
          disabled={isLoading}
          {...register("about_me")}
        />
        <p className="text-[10px] text-muted-foreground text-right">
          {watch("about_me")?.length || 0}/500
        </p>
        {errors.about_me && (
          <p className="text-xs text-destructive font-medium mt-1">{errors.about_me.message}</p>
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
