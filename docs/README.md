# AssoLink - Documentation

Documentation technique et fonctionnelle du projet **AssoLink** (réseau social pour associations, TPE/PME et bénévoles locaux).

## Naming

- **AssoLink** = nom de projet (interne, dossier CDA, dépôt local, code).
- **Komunumo** = nom de produit (interface utilisateur, branding, marketing). Esperanto pour "communauté", choix éditorial qui évoque l'universalité associative.
- Dépôt distant : `git@github.com:nicgen/komunumo.git`.
- URLs : `https://app.local.hello-there.net` (frontend Komunumo), `https://api.local.hello-there.net` (backend AssoLink).

Les ADR, specs et le dossier emploient "AssoLink" pour parler du projet d'ingénierie. Les chaînes UI, balises `<title>`, méta OG, et copywriting marketing emploient "Komunumo".

## Index

| Dossier | Contenu |
| --------- | --------- |
| [adr/](./adr/) | Architecture Decision Records (format MADR) |
| [specs/](./specs/) | Specifications Speckit (vision, features, API, data, qualité) |
| [diagrams/](./diagrams/) | Diagrammes Mermaid (cas d'usage, classes, séquence) |
| [mockups/](./mockups/) | Maquettes UI |
| [audits/](./audits/) | Rapports Lighthouse, pa11y, EcoIndex datés |
| [learnings/](./learnings/) | Notes personnelles d'apprentissage |
| [dossier/](./dossier/) | Dossier de soutenance (chapitres + build PDF) |

## Conventions

- Tout en Markdown, versionné Git.
- Liens relatifs uniquement.
- Une décision = un ADR écrit le jour même.
- Une feature = une spec Speckit écrite avant le code.
- Mermaid embarqué dans les `.md` quand pertinent, fichiers `.mmd` séparés sinon.

## Build du dossier PDF

```bash
cd dossier/build && make pdf
```

Produit `dossier-vX.Y.Z.pdf` via Pandoc + template Eisvogel.

## Browse en local

Optionnel mais confortable :

```bash
pip install mkdocs-material
mkdocs serve   # http://localhost:8000
```

## Status

Projet en **Semaine 0** (cadrage). Voir `/home/nic/dev/certif/resources/studies/08-plan-5-semaines.md` pour le planning détaillé.
