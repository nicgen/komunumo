# ADR-0010 - Markdown + Pandoc + Eisvogel pour le dossier de soutenance

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic

## Contexte

Le dossier de soutenance CDA doit être un PDF professionnel, lisible, paginé, avec sommaire, numérotation, en-têtes, bibliographie. Il doit pouvoir être versionné Git, généré reproductiblement en CI, modifié rapidement et corrigé par LLM. Le candidat assume que la majorité du contenu sera générée puis relue.

## Décision

Rédiger le dossier en **Markdown** dans `docs/dossier/0X-titre.md`, compiler en PDF via **Pandoc + template Eisvogel** :

```bash
pandoc dossier/0*.md \
  -o dossier-vX.Y.Z.pdf \
  --template eisvogel \
  --listings \
  --toc --toc-depth=3 \
  --number-sections \
  --metadata-file dossier/build/metadata.yml \
  --filter pandoc-mermaid
```

Build via `Makefile` dans `docs/dossier/build/`. Pipeline CI dédiée (`.github/workflows/dossier.yml`) qui produit le PDF en artefact à chaque push.

## Alternatives écartées

- **LaTeX direct** : contrôle typographique maximal, mais courbe d'apprentissage et debug pénible. ROI négatif sur 2 semaines de rédaction. Le rendu Pandoc + Eisvogel est indistinguable d'un LaTeX manuel pour un jury non-typographe.
- **Microsoft Word / Google Docs** : non versionnable Git, peu friendly LLM, pas reproductible en CI. Disqualifié.
- **Notion / Outline export PDF** : verrouillage SaaS, mise en page peu maîtrisable.
- **Typst** : prometteur, mais écosystème encore jeune et templates moins matures que Eisvogel en 2026.
- **AsciiDoc + AsciiDoctor PDF** : alternative valable, mais Markdown reste plus universellement supporté par l'outillage et les LLM.

## Conséquences

- (+) Rédaction LLM-friendly, relecture rapide.
- (+) Versionnage Git natif, diff lisible.
- (+) Build reproductible en CI : `make pdf` fonctionne identiquement partout.
- (+) Argument oral : "pipeline DevOps même pour la documentation - le PDF est compilé par GitHub Actions et stocké comme artefact à chaque tag."
- (+) Mermaid embarqué via filtre Pandoc.
- (-) Mise en page très fine (espace blanc, kerning) moins contrôlable qu'en LaTeX direct. Mitigation : Eisvogel offre 95% du résultat d'un LaTeX maison.
- (-) Filtres Pandoc tiers (mermaid-filter, pandoc-crossref) à installer en CI. Mitigation : image Docker `pandoc/extra` les contient.

## Références

- Pandoc - https://pandoc.org/
- Eisvogel template - https://github.com/Wandmalfarbe/pandoc-latex-template
- pandoc-mermaid filter - https://github.com/raghur/mermaid-filter
