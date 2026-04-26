# 01 - Domaine et glossaire

## Glossaire

| Terme | Définition |
|-------|------------|
| **Compte** (Account) | Une identité authentifiable. Soit un compte personnel (Member), soit un compte association (Association). Un compte = un email unique. |
| **Personne** (Member) | Compte représentant un individu (bénévole, adhérent, salarié). Possède un nom, prénom, date de naissance, etc. |
| **Association** | Compte représentant une entité morale (asso loi 1901, TPE, collectif). Possède un nom moral, optionnellement SIREN/RNA, code postal de référence. |
| **Adhésion** (Membership) | Lien entre une Personne et une Association, qualifié par un rôle (`owner`, `admin`, `member`). Une Personne peut avoir N adhésions ; une Association peut avoir N membres. |
| **Profil** | Page publique d'un compte. Peut être public ou privé. Si privé, contenu visible uniquement aux abonnés acceptés. |
| **Suivi** (Follow) | Relation asymétrique entre deux comptes. Si la cible est privée, demande nécessaire avec acceptation. |
| **Post** | Contenu publié par un compte. Visibilité : `public`, `followers`, `members` (asso), `private`. |
| **Commentaire** | Réponse sous un post. Hérite des règles de visibilité du post parent. |
| **Groupe** | Synonyme d'Association dans l'UI ; les associations **sont** les groupes. Une Association peut avoir des canaux internes (chat groupe). |
| **Événement** | Action organisée par une Association : titre, description, date/heure, lieu (code postal), capacité. Les comptes peuvent répondre `going` / `not_going` / `maybe`. |
| **Conversation** | Fil de messages entre 2 ou N comptes. Soit directe (DM 1-1), soit groupe (canal d'asso). |
| **Message** | Unité d'échange dans une conversation. Texte + emojis. Pas de fichier en V1. |
| **Notification** | Événement signalé à un utilisateur. Agrégeable (cf. F2). |
| **Audit log** | Journal append-only signé HMAC chaîné, réservé aux actions sensibles (suppression compte, suppression contenu, modération, export RGPD). |

## Invariants métier

1. Un email correspond à **au plus** un compte (Member OU Association).
2. Un compte Association a **au moins** un Member avec rôle `owner` à tout instant.
3. Un Member ne peut pas s'auto-inviter dans une Association ; il doit être invité ou demander.
4. La suppression d'un compte est **soft-delete** + audit log + tâche de purge à 30 jours.
5. Un Post privé n'est visible que de son auteur et des followers explicitement listés.
6. Une conversation directe nécessite que **au moins une** des deux personnes suive l'autre OU que le compte cible soit public.
7. Un Member ne peut envoyer un message dans un canal d'asso que s'il a une `Membership` active dans cette asso.
8. Tout enregistrement modifiable porte `created_at` et `updated_at`.
9. Tout export RGPD (article 20) est tracé dans l'audit log.
10. Aucune donnée personnelle n'est exposée dans les logs applicatifs (PII redaction obligatoire).

## Permissions de référence

| Action | Member | Asso owner | Asso admin | Asso member |
|--------|--------|------------|------------|-------------|
| Publier post personnel | Oui | - | - | - |
| Publier post asso | - | Oui | Oui | Non |
| Inviter dans asso | - | Oui | Oui | Non |
| Accepter demande adhésion | - | Oui | Oui | Non |
| Créer événement asso | - | Oui | Oui | Non |
| Modérer commentaires asso | - | Oui | Oui | Non |
| Supprimer compte asso | - | Oui | Non | Non |
| Voir audit log asso | - | Oui | Oui | Non |
| Exporter ses données (RGPD) | Oui | Oui | Oui | Oui |

## Bornes de domaine

- Un Post fait au plus **5000 caractères**.
- Un Member peut suivre au plus **5000 comptes** (limite anti-spam).
- Une Association peut avoir au plus **5000 membres** en V1 (révisable).
- Un événement a au plus **500 RSVP** en V1.
- Une conversation directe contient au plus **2 participants** ; un canal d'asso au plus **5000**.
- Taille upload avatar : **2 Mo max**, formats JPEG/PNG/WEBP/AVIF, redimensionné serveur à 256x256.

## Règles RGPD intégrées au domaine

- Article 17 (effacement) : suppression compte = anonymisation des contributions (auteur remplacé par "Compte supprimé"), pas perte des contenus utiles à la communauté.
- Article 20 (portabilité) : export complet en JSON + PDF lisible humain, livré en moins de 24h (synchrone si possible).
- Article 13 (information) : page `/legal/confidentialite` détaillant traitements et durées de conservation.
- Pseudonymisation possible : pseudo distinct du nom légal, affiché par défaut dans les fils publics.
