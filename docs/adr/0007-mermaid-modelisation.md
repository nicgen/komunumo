# ADR-0007 - Mermaid pour la modélisation UML et ER

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le brief CDA exige des modèles **Merise** (MCD/MLD/MPD) et **UML** (cas d'usage, classes, séquence minimum). Les diagrammes doivent être versionnables Git, modifiables rapidement (révisions multiples en 5 semaines), affichables dans le dossier PDF et le repo, et générables/corrigeables par LLM.

## Décision

Utiliser **Mermaid** comme outil principal de modélisation, avec fichiers `.mmd` versionnés dans `docs/diagrams/` et embarqués dans le dossier final via Pandoc. Les diagrammes couverts :

- Cas d'usage (UML use case via `flowchart`).
- Classes (UML class).
- Séquence (UML sequence) sur au moins 2 parcours critiques (auth, envoi message WS).
- ER pour MCD (Mermaid `erDiagram`), complété au besoin par DBML (dbdiagram.io) pour le MLD si une vue plus académique est demandée.

## Alternatives écartées

- **PlantUML** : plus formel et fidèle à la notation Merise IE, mais demande un serveur de rendu (plantuml-server) ou une JVM en local. Friction supplémentaire pour rien sur notre cas.
- **draw.io / diagrams.net** : visuel agréable mais XML peu lisible, non générable par LLM, pas de rendu natif GitHub.
- **Lucidchart, Miro** : SaaS, payant au-delà du free, non versionnable Git.
- **dbdiagram.io / DBML** : excellent pour MCD/MLD spécifiquement, gardé en complément optionnel uniquement.

## Conséquences

- (+) Versionnable, diffable, rendable par GitHub nativement.
- (+) LLM (Gemini, Claude) génèrent et corrigent du Mermaid en un prompt.
- (+) Embed dans dossier Pandoc via filtre `mermaid-filter` ou pré-rendu en SVG.
- (+) MCP serveurs Mermaid existent (validation, rendu).
- (-) Notation Merise IE moins canonique qu'avec PlantUML. Mitigation : `erDiagram` reste lisible et largement accepté ; le jury évalue la rigueur du modèle, pas la stricte typographie.
- (-) Diagrammes très grands deviennent illisibles. Mitigation : split par sous-domaine fonctionnel.

## Références

- Mermaid docs - https://mermaid.js.org/
- DBML (complément MCD/MLD) - https://dbml.dbdiagram.io/
