# ADR-0006 - Tailwind v4 et shadcn/ui sur Radix UI

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

L'UI doit être responsive, **accessible RGAA AAA** sur les parcours critiques, livrable rapidement (4 semaines MVP) et défendable techniquement face au jury. La maintenance par un dev solo impose de minimiser le code custom.

## Décision

- **Tailwind CSS v4** comme système de styles utilitaires (engine Rust, runtime nul).
- **shadcn/ui** comme bibliothèque de composants. Les composants sont **copiés** dans `components/ui/`, pas importés en npm. Le projet possède le code et peut le modifier.
- shadcn/ui s'appuie sur **Radix UI** dont les primitives gèrent par construction le focus, ARIA, navigation clavier, lecteur d'écran.
- Tokens de design centralisés dans `app/globals.css` via `@theme` directive.

## Alternatives écartées

- **Material UI / MUI** : poids JS élevé, customisation lourde, choix de design imposé.
- **Chakra UI** : excellente accessibilité mais runtime CSS plus lourd que Tailwind v4.
- **Mantine** : alternative valable, écosystème plus restreint.
- **Vanilla CSS Modules** : viable mais ralentit la livraison en MVP.
- **Composants maison** : risque accessibilité élevé sur 4 semaines.

## Conséquences

- (+) Accessibilité gratuite via Radix (focus trap, escape handling, aria-*).
- (+) Bundle CSS proportionnel à l'usage réel (atomique).
- (+) Pas de dépendance UI verrouillante (code à toi).
- (+) shadcn/ui est devenu un standard de fait en 2025-2026.
- (-) Maintenance manuelle si shadcn évolue. Mitigation : MAJ ponctuelles, pas auto.
- (-) Tailwind classes peuvent rendre le JSX dense. Mitigation : extractions en composants dès qu'un pattern se répète.
- (-) Tokens de couleur à valider au contraste WCAG AA dès la première itération. Mitigation : palette validée Tailwind a11y au S0.

## Références

- shadcn/ui - https://ui.shadcn.com/
- Radix UI accessibility - https://www.radix-ui.com/primitives/docs/overview/accessibility
- Tailwind CSS v4 announcement - https://tailwindcss.com/blog/tailwindcss-v4
