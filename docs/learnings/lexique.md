# Lexique

Termes techniques et métier à maîtriser pour la soutenance.

---

## Architecture & Design

### Architecture hexagonale (Ports & Adapters)

**Définition :** Pattern d'architecture qui isole le domaine métier de ses dépendances externes (base de données, HTTP, email…) via des interfaces (ports) implémentées par des adaptateurs.

**Termes associés :** port, adapter, domaine, use case, infrastructure

**Dans le projet :** `backend/internal/domain/`, `backend/internal/ports/`, `backend/internal/adapters/`

---

### Domain-Driven Design (DDD)

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Backend Go

### Goroutine

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### Middleware

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Base de données

### WAL (Write-Ahead Logging)

**Définition :** Mode de journalisation SQLite où les écritures sont d'abord enregistrées dans un fichier log séparé avant d'être appliquées à la base. Permet les lectures concurrentes pendant une écriture.

**Termes associés :** SQLite, concurrence, PRAGMA journal_mode

**Dans le projet :** activé à l'ouverture de la connexion dans `backend/internal/adapters/db/`

---

### Migration

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Sécurité

### bcrypt

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### HttpOnly / SameSite / Secure (cookie)

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### CSRF

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Frontend

### Server Component / Client Component (React)

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### Hydration

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Qualité & Tests

### TDD (Test-Driven Development)

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### Couverture de code

**Définition :**

**Termes associés :**

**Dans le projet :**

---

## Standards & Spécifications

### Versionnement d'une spécification (norme)

**Définition :** Le versionnement d'une norme (ex. OpenAPI, HTTP, JSON Schema) suit une logique différente du versionnement logiciel. Les trois niveaux ont des significations précises :

- **Version majeure** (ex. 3 → 4) : rupture de compatibilité, changements de paradigme
- **Version mineure** (ex. 3.1 → 3.2) : ajout de fonctionnalités, rétrocompatible
- **Patch** (ex. 3.1.0 → 3.1.1) : corrections **éditoriales uniquement** — fautes de frappe, ambiguïtés clarifiées, exemples incorrects, liens cassés. Aucun changement de comportement ou de validation. Un document valide en 3.1.0 est valide en 3.1.2 sans aucune modification.

**Conséquence pratique :** déclarer `openapi: "3.1.0"` dans un fichier reste valide face à un outil qui supporte 3.1.2. La version dans l'en-tête est informative, pas contractuelle au niveau patch.

**Termes associés :** SemVer, rétrocompatibilité, spécification, norme

**Dans le projet :** `docs/specs/03-api/openapi.yaml` — déclaré en OpenAPI 3.1.0

---

### OpenAPI

**Définition :** Spécification standard (anciennement Swagger) pour décrire des API REST de manière lisible par les humains et les machines. Permet de générer de la documentation interactive, des clients, et de valider les contrats d'API.

**Versions clés :** 3.0.x (courante), 3.1.0 (alignée JSON Schema Draft 2020-12), 3.2.0 (sept. 2025)

**Termes associés :** contrat d'API, JSON Schema, Swagger, Redocly

**Dans le projet :** `docs/specs/03-api/openapi.yaml`, validé en CI par `@redocly/cli`

---

## Méthodes & Processus

### Conventional Commits

**Définition :**

**Termes associés :**

**Dans le projet :**

---

### ADR (Architecture Decision Record)

**Définition :** Document qui capture une décision architecturale importante : le contexte, les options envisagées, la décision retenue et ses conséquences.

**Termes associés :** décision, trade-off, conséquences

**Dans le projet :** `docs/adr/`

---

<!-- Template pour ajouter un terme :

### Terme

**Définition :**

**Termes associés :**

**Dans le projet :**

---
-->
