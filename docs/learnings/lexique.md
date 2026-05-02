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
