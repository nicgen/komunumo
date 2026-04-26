# Qualité - Éco-conception (RGESN)

## Référentiels

- **RGESN** (DINUM/ARCEP/ADEME) - 78 critères.
- **W3C Sustainable Web Design** - https://w3c.github.io/sustyweb/
- **EcoIndex** (Green IT Analysis) - https://www.ecoindex.fr/

## Budgets de performance par page

| Métrique | Cible homepage | Cible page interne | Cible feed authentifié |
|----------|----------------|--------------------|------------------------|
| Poids transféré | < 300 KB | < 500 KB | < 700 KB (dont images) |
| Nombre de requêtes | < 25 | < 35 | < 50 |
| LCP | < 2.0s | < 2.5s | < 2.5s |
| TBT | < 100ms | < 200ms | < 200ms |
| CLS | < 0.05 | < 0.1 | < 0.1 |
| JS premier load | < 80 KB gz | < 100 KB gz | < 120 KB gz |
| Score Lighthouse Perf | ≥ 95 | ≥ 90 | ≥ 85 |
| Note EcoIndex | A | A ou B | B |

Tout dépassement bloque le merge en CI (lighthouse-ci avec assertions).

## Choix techniques pro-éco déjà décidés

- Server Components par défaut (pas de JS si pas d'interactivité).
- `next/image` avec AVIF (-30% vs WebP, -50% vs JPEG).
- `next/font` self-hosted (pas de requête Google Fonts).
- Tailwind v4 atomique (CSS proportionnel à l'usage).
- Pas de tracking analytics tiers (ou Plausible/Umami self-hosted en V2).
- Pas d'auto-play vidéo, pas d'animations cosmétiques sans valeur produit.
- Compression Brotli sur tous les assets texte côté Traefik.
- Cache HTTP `max-age=31536000, immutable` sur assets hashés.
- Pagination ou bouton "voir plus" explicite (jamais d'infinite scroll).
- Limite upload : 2 Mo / image, redimensionné serveur à 1280px max.

## Mesure et affichage à l'utilisateur

Footer global affiche :
```
Cette page : 0.12 gCO2eq • Voir détails
```

Calcul : modèle Sustainable Web Design Model v4
```
gCO2eq ≈ KB_transferred × 0.81 × carbon_intensity_FR
        (carbon_intensity_FR ≈ 56 gCO2eq/kWh - 2025 ADEME)
```

Page `/eco` détaille la méthode, les sources, les limites de l'estimation.

## Mode sobriété

Cumulé avec a11y (cf. `a11y.md`), réduit la consommation à :
- ~50 KB / page (HTML + CSS critique inline + 0 image).
- 0 WebSocket (polling minimal ou notif banner only).
- Argument oral fort sur le double bénéfice eco + a11y.

## Audit en CI

Workflow `eco.yml` quotidien :
1. Lighthouse-ci sur 5 pages clés.
2. EcoIndex via API (snapshots datés dans `docs/audits/`).
3. Tendances trackées dans GitHub Actions summary.

## Carbon budget par feature

Avant d'ajouter une feature, estimer son coût :
- Bytes additionnels par page concernée.
- Requêtes additionnelles.
- CPU additionnel (si client-heavy).

Si la feature pousse la page au-dessus du budget : refactor ou rejet.

## Décisions documentées

| Décision | Gain estimé |
|----------|-------------|
| Pas de Material Icons (font 130 KB), SVG inline à la place | -120 KB |
| Pas de jQuery / Lodash | -50 KB chacun évité |
| RSC pour pages publiques | -40 KB JS / page |
| AVIF au lieu de JPEG | -50% poids images |
| Pas de polyfills inutiles (cible navigateurs modernes) | -20 KB |

Total estimé sur la home : **~280 KB économisés** vs un projet "standard" Next.js typique.

## Mention dans le dossier

> "Notre application affiche en bas de chaque page l'empreinte carbone estimée de la requête servie. Cette transparence est un parti pris militant et éducatif. Nous reconnaissons les limites du calcul (variables datacenter, mix énergétique, terminal client) ; le détail de la méthode est exposé sur la page /eco."
