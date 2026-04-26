# ADR-0002 - Next.js 16 App Router et React Server Components

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le frontend doit être performant, SEO-friendly (les pages publiques d'asso doivent être indexables), accessible, et envoyer le moins de JavaScript possible (cohérent avec l'argument éco-conception). Next.js 16 (sortie 2025) stabilise l'App Router, les Server Components, le Partial Prerendering, et renomme `middleware.ts` en `proxy.ts`.

## Décision

Utiliser **Next.js 16 avec App Router et React Server Components (RSC) par défaut**. Composants client (`'use client'`) uniquement où nécessaire (interactivité, état, WebSocket). TypeScript strict mode. Routing par fichiers dans `app/`. Server Actions pour les mutations simples, API routes pour les endpoints complexes consommés aussi par d'autres clients.

## Alternatives écartées

- **Next.js Pages Router** : legacy en 2026, moins de fonctionnalités RSC.
- **Remix / React Router 7** : excellente alternative, mais écosystème SSR moins mature pour notre cas (pas de Server Actions équivalentes aussi simples).
- **SvelteKit** : très bon, mais ajoute une courbe d'apprentissage non-justifiée vu la maîtrise React du candidat.
- **Astro** : excellent pour contenu statique, mal adapté à un réseau social interactif.

## Conséquences

- (+) JS envoyé minimisé (Server Components ne génèrent que du HTML). Argument eco direct.
- (+) SEO natif sur pages publiques (profils asso, événements).
- (+) Accessibilité : HTML sémantique côté serveur, fonctionne sans JS pour la lecture.
- (+) Streaming et Suspense pour des LCP rapides.
- (-) Verrouillage modéré sur l'écosystème Next.js. Mitigation : Next.js peut tourner sur Node standard, Docker, ou Clever Cloud sans Vercel.
- (-) Mental model RSC vs client à apprendre. Mitigation : règle simple "tout est serveur sauf si interaction".

## Références

- Next.js docs - https://nextjs.org/docs
- React Server Components - https://react.dev/reference/rsc/server-components
