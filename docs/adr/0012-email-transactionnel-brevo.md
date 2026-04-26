# ADR-0012 - Email transactionnel via Brevo

- Statut : Accepté
- Date : 2026-04-26
- Révisé : 2026-04-26 (note ajoutée sur le footer Brevo en plan gratuit)
- Décideur : nic

## Contexte

L'application doit envoyer des emails transactionnels :
- Vérification d'adresse à l'inscription (UC4).
- Réinitialisation de mot de passe (UC5).
- Notifications agrégées en digest (V2 ou si capacité S2).
- Lien de téléchargement export RGPD (UC23).

Volume estimé MVP : < 500 emails/mois. Délivrabilité critique (un email vérification non délivré = blocage onboarding). Souveraineté préférée.

## Décision

**Brevo** (anciennement Sendinblue, **société française**) via API HTTP transactionnelle.

- Plan gratuit : 300 emails/jour (~9 000/mois), suffisant en MVP. **Note** : le plan gratuit insère un footer "Sent via Brevo" sur les emails. Acceptable pour le MVP, suppressible en plan Lite (~25 €/mois) le jour où la marque doit être 100 % maîtrisée.
- API : `POST https://api.brevo.com/v3/smtp/email`.
- Authentification : clé API dans 1Password, injectée via `op run`.
- Domaine d'envoi : `noreply@hello-there.net`, SPF/DKIM/DMARC configurés côté DNS Cloudflare.
- Templates : stockés en base Brevo (modèles versionnés via export JSON dans `assolink/ops/email-templates/`).

Adapter Go : `internal/adapters/email/brevo.go` implémente le port `Mailer` défini dans `internal/ports/email.go`. Mock disponible pour tests.

## Alternatives écartées

- **Mailjet** (français aussi, racheté par Sinch) : équivalent fonctionnel, plan gratuit comparable. Brevo retenu pour son interface plus moderne et sa documentation.
- **AWS SES** : excellent rapport qualité/prix mais entreprise US (CLOUD Act) et configuration DKIM plus pénible.
- **Postmark** : meilleur en délivrabilité mais payant dès le 1er email (15 $/mois 10k).
- **SMTP self-hosted** : trop fragile (réputation IP, blacklists, maintenance), incompatible avec le délai MVP.
- **Resend** : très bon DX mais entreprise US.

## Conséquences

- (+) Souveraineté maintenue (Brevo = entreprise française, datacenters EU).
- (+) Coût zéro en MVP et en V1 raisonnable.
- (+) Tableau de bord Brevo donne taux d'ouverture / bounce, utile pour audit délivrabilité.
- (+) Webhooks Brevo (bounce, complaint) consommables côté Go pour invalider des emails morts.
- (-) Lock-in modéré : l'adapter `Mailer` permet de basculer si besoin (Mailjet, SES, etc.) en quelques heures.
- (-) Emails marketing non couverts. Hors scope MVP, prévu V2 si besoin newsletter.

## Argumentaire jury

> "Brevo est un fournisseur d'email transactionnel français, proposant un plan gratuit de 300 emails/jour. Ce choix répond à deux critères : la souveraineté (entreprise et datacenters européens, conforme RGPD sans transfert hors UE) et la viabilité économique pour un projet MVP. L'intégration passe par un port `Mailer` côté hexagonal Go, permettant de substituer le fournisseur sans toucher au domaine."

## Références

- Brevo Transactional API - https://developers.brevo.com/reference/sendtransacemail
- Brevo plan gratuit - https://www.brevo.com/fr/pricing/
- ADR-0001 (port `Mailer`).
- `02-features/auth.md` (déclencheurs UC1, UC4, UC5).
