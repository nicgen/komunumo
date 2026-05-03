"use client";

import { useEffect, useState } from "react";
import { AuthCard } from "@/components/auth/auth-card";
import { MemberProfileForm } from "@/components/profile/member-profile-form";
import { AssociationProfileForm } from "@/components/profile/association-profile-form";
import { Loader2 } from "lucide-react";

type VisibilityKind = "public" | "members_only" | "private";
type MemberProfile = { kind: "member"; first_name: string; last_name: string; nickname?: string; about_me?: string; visibility: VisibilityKind };
type AssociationProfile = { kind: "association"; legal_name: string; about?: string; postal_code: string; visibility: VisibilityKind };
type UserProfile = MemberProfile | AssociationProfile;

export default function ProfilePage() {
  const [profile, setProfile] = useState<UserProfile | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    async function fetchProfile() {
      try {
        const response = await fetch("/api/v1/me/profile");
        if (response.ok) {
          const data = await response.json();
          setProfile(data);
        } else if (response.status === 401) {
          window.location.href = "/login";
        } else {
          setError("Impossible de charger le profil");
        }
      } catch {
        setError("Erreur de connexion");
      } finally {
        setIsLoading(false);
      }
    }

    fetchProfile();
  }, []);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[calc(100vh-10rem)]">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-[calc(100vh-10rem)]">
        <p className="text-destructive font-medium">{error}</p>
      </div>
    );
  }
  
  if (!profile) {
    return null;
  }

  return (
    <div className="flex flex-col items-center justify-center py-8 px-4">
      <AuthCard
        title="Mon Profil"
        description={
          profile.kind === "member"
            ? `Connecté en tant que ${profile.first_name} ${profile.last_name}`
            : `Gestion de l'association ${profile.legal_name}`
        }
      >
        {profile.kind === "member" ? (
          <MemberProfileForm initialData={profile} />
        ) : (
          <AssociationProfileForm initialData={profile} />
        )}
      </AuthCard>
    </div>
  );
}
