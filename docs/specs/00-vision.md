# 00 - Vision

## Pitch

AssoLink est un **réseau social opérationnel** dédié aux associations loi 1901, TPE/PME et bénévoles d'un même territoire. Il permet d'annoncer des besoins, recruter des bénévoles, mutualiser des ressources, organiser des événements et communiquer sans dépendance aux réseaux sociaux dominants.

## Mission

Offrir un outil **souverain, accessible et sobre** au tissu associatif et économique local, pour qu'il puisse communiquer sans s'exposer aux dérives des plateformes commerciales (publicité ciblée, modération opaque, données monétisées).

## Personas

### P1 - Anne, présidente d'association (45 ans)

- Préside une asso culturelle de 80 membres.
- Cherche des bénévoles ponctuels pour des événements.
- Utilise Facebook par dépit, déteste y voir ses adhérents exposés à la pub.
- Compétences numériques moyennes.
- **Besoins** : publier des annonces, gérer les inscriptions à un événement, échanger en groupe.

### P2 - Karim, gérant TPE (38 ans)

- Gère un FabLab associatif (10 salariés).
- Veut être visible auprès des assos locales pour partenariats et prestations.
- **Besoins** : profil pro, posts, candidater aux annonces, messagerie directe.

### P3 - Léa, bénévole occasionnelle (24 ans)

- Étudiante, veut s'engager localement quelques heures par semaine.
- Mobile-first, attentive à l'accessibilité (proche d'un parent malvoyant).
- **Besoins** : trouver des annonces près de chez elle, postuler en 2 clics, suivre ses assos préférées.

### P4 - Maurice, bénévole senior (68 ans)

- Retraité, donne 5h/semaine à plusieurs assos.
- Utilise un PC fixe avec un navigateur ancien, lecteur d'écran ponctuel.
- **Besoins** : navigation simple, gros caractères, pas d'animations.

## Mesure du succès

| Indicateur | Cible MVP (démo jury) | Cible 6 mois post-soutenance |
| ------------ | ---------------------- | ------------------------------- |
| Associations inscrites | 20 (seed data) | 100 réelles |
| Comptes utilisateurs | 50 (seed) | 500 réels |
| Événements créés | 10 | 50/mois |
| Score Lighthouse Accessibilité | 100 | Maintenu |
| Score Lighthouse Performance | ≥ 95 | Maintenu |
| Critères RGAA AAA passés sur parcours critiques | 100% | 100% |
| Note EcoIndex page d'accueil | A | A |
| Empreinte CO2 page d'accueil | < 0.3 gCO2eq | < 0.3 gCO2eq |

## Contraintes

- **Délai** : MVP livré en 4 semaines + 1 semaine finition + 1 semaine soutenance.
- **Équipe** : 1 développeur (assisté LLM).
- **Stack imposée** : Go, coder/websocket, SQLite, Next.js, React, Tailwind.
- **Conformité** : RGPD (article 20 portabilité), RGAA AAA parcours critiques, RGESN niveau "appliqué".
- **Hébergement** : frontend Vercel (Paris/Francfort), backend Scaleway DEV1-S (Paris) pour la démo, Traefik local pour le dev.
- **Budget** : 0 EUR de licences logicielles (open-source partout).

## Hors scope V1

- Fédération ActivityPub (modèle compatible mais endpoints non livrés).
- Application mobile native (PWA suffit).
- Visioconférence WebRTC.
- Marketplace / paiements.
- Modération assistée par IA (modération humaine en V1).
- Carte interactive (filtre par code postal en V1, carte en V2).
- Catalogue de prêt entre associations (mention V2).
- Édition collaborative CRDT.

## Vision long terme (5 ans)

- Devenir le réseau de référence pour le tissu associatif francophone.
- Fédérer avec Mobilizon, Mastodon, Pixelfed via ActivityPub.
- Modèle économique : freemium pour les associations (gratuit), payant léger pour les TPE/PME (~5 EUR/mois).
- Code open source AGPL, gouvernance associative.
