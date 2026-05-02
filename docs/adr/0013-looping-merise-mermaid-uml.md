# ADR-0013 - Looping pour Merise (MCD/MLD/MPD), Mermaid pour UML (supersède partiellement ADR-0007)

- Statut : Accepté
- Date : 2026-04-26
- Décideur : nic
- Supersède partiellement : [ADR-0007](./0007-mermaid-modelisation.md) (volet Merise uniquement)

## Contexte

ADR-0007 retenait Mermaid pour l'ensemble des modélisations, y compris ER pour le MCD/MLD. Vérification du référentiel CDA (RNCP37873) : le jury évalue la **conformité Merise IE** (cardinalités 0,N / 1,1 / 0,1 / 1,N standardisées, distinction stricte MCD / MLD / MPD, dérivation algorithmique). Le `erDiagram` de Mermaid produit une notation Crow's Foot proche d'UML, **non strictement Merise**. Risque de note diminuée sur l'axe "modélisation" du dossier.

## Décision

**Séparer les notations selon leur usage :**

| Type de modèle | Outil | Format fichier | Rendu |
| ---------------- | ------- | ---------------- | ------- |
| MCD (Merise) | **Looping** | `.lpg` (proprio binaire) + export PNG/SVG | image dans dossier |
| MLD (Merise) | **Looping** (dérivation auto depuis MCD) | export PNG/SVG + DDL SQL généré | image + extrait SQL |
| MPD (Merise) | **Looping** + post-édition manuelle | export PNG + DDL SQL final | image dans dossier |
| UML cas d'usage | **Mermaid** (`flowchart`) | `.mmd` versionné | rendu GitHub natif + SVG dans dossier |
| UML classes | **Mermaid** (`classDiagram`) | `.mmd` versionné | idem |
| UML séquence | **Mermaid** (`sequenceDiagram`) | `.mmd` versionné | idem |
| ER technique (vue dev) | **Mermaid** (`erDiagram`) gardé | `.mmd` versionné | idem (vue technique en complément du MCD Merise) |

### Looping en pratique

- Outil français gratuit (https://www.looping-mcd.fr/), desktop Windows/Mac/Linux.
- Auteur : Yvan Caron, utilisé en formation française. Reconnu jury.
- Génère MCD Merise IE conforme, dérive le MLD automatiquement, exporte en PNG/SVG/PDF, **génère le DDL SQL** (PostgreSQL/MySQL/SQLite).
- Limite Git : format `.lpg` binaire, pas de diff lisible. Mitigation : commits propres ("docs(specs): regen MCD after adding Notification entity"), captures PNG datées, et MLD textuel maintenu en parallèle dans `docs/specs/04-data/mld.md`.

### Workflow proposé

```
docs/specs/04-data/
  merise/
    komunumo.lpg              # source Looping (binaire, versionné)
    mcd.png                   # export pour dossier (regénéré à chaque modif)
    mld.png                   # idem
    mpd.png                   # idem (avec types et contraintes finales)
    schema.sql                # DDL SQL généré par Looping (référence)
  mcd.mmd                     # ER Mermaid technique (déjà existant, gardé en double)
  mld.md                      # MLD textuel détaillé (déjà existant, gardé)
```

Le **MLD textuel** dans `mld.md` reste la source autoritaire pour les colonnes, types, contraintes, index FTS5 et triggers. Le **MPD** Looping est dérivé une fois le MLD stabilisé. La cohérence est vérifiée manuellement à chaque jalon (revue dossier en S4).

## Alternatives écartées

- **Tout Mermaid (ADR-0007 initial)** : moins canonique côté Merise. Risque jury.
- **Mocodo** (https://mocodo.net/) : plus rigoureux que Looping côté typographie Merise stricte (notation OMT possible), 100 % textuel donc Git-friendly. **Sérieusement envisagé.** Écarté car : (a) sortie graphique moins flatteuse, (b) courbe d'apprentissage de la syntaxe Mocodo, (c) Looping est plus connu en France et fait foi côté formation. Reste **plan B** si Looping pose problème.
- **DBML + dbdiagram.io** : excellent pour MLD orienté dev, **mais** notation ER non strictement Merise. Conservé éventuellement comme vue dev complémentaire.
- **PlantUML avec Merise IE custom** : possible mais demande syntaxe custom et serveur de rendu, perte de l'avantage rapidité Mermaid sans gain canonique.
- **draw.io** : pas versionnable proprement (XML lourd), pas de génération SQL, pas de dérivation auto.

## Conséquences

- (+) **Conformité Merise IE** assurée pour le dossier (axe noté du référentiel CDA).
- (+) DDL SQL **généré automatiquement** depuis le MCD : preuve de cohérence modèle <-> base.
- (+) UML reste rapide à éditer en Mermaid (versionnable, LLM-friendly).
- (+) ER Mermaid technique conservé en doublon : vue dev pratique, complète sans remplacer le MCD.
- (-) Format `.lpg` binaire, diff Git pauvre. Mitigation : commits ciblés + PNG dans Git pour traçabilité visuelle.
- (-) Outil desktop, pas de génération en CI (contrairement à Mermaid). Mitigation : régénération manuelle au plus tard à chaque jalon, exports figés en PNG dans Git.
- (-) Une légère duplication entre `mcd.mmd` (Mermaid) et `merise/mcd.png` (Looping). Assumée : audiences différentes (devs sur GitHub vs jury sur PDF).

## Plan de migration depuis l'existant

1. Installer Looping (téléchargement direct, multi-plateforme).
2. Reproduire le contenu de `04-data/mcd.mmd` dans Looping (entités, attributs, relations).
3. Vérifier que le MLD dérivé correspond à `04-data/mld.md`. Ajuster.
4. Exporter PNG + DDL SQL.
5. Commit `docs(specs): add Merise MCD/MLD/MPD via Looping`.

Estimation : 2-3 heures, à faire en S0 ou début S1.

## Références

- Looping (Yvan Caron) - https://www.looping-mcd.fr/
- Mocodo (alternative textuelle) - https://mocodo.net/
- Méthode Merise (ressource pédagogique) - https://www.commentcamarche.net/contents/665-merise-modele-conceptuel-des-donnees
- ADR-0007 (Mermaid général, partiellement remplacé par celui-ci).
