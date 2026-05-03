-include .env
export

SONAR_URL     ?= http://127.0.0.1:9000
SONAR_PROJECT := nicgen_komunumo
SONAR_NAME    := Komunumo

.PHONY: sonar sonar-start sonar-stop sonar-setup

sonar-start:
	docker compose -f docker-compose.sonar.yml up -d
	@echo "SonarQube disponible sur $(SONAR_URL) (attendre ~30s au premier lancement)"

sonar-stop:
	docker compose -f docker-compose.sonar.yml down

sonar-setup: ## Crée le projet dans SonarQube local (une seule fois)
	@if [ -z "$(SONAR_TOKEN)" ]; then echo "SONAR_TOKEN manquant dans .env"; exit 1; fi
	@echo "Création du projet $(SONAR_PROJECT) dans SonarQube..."
	@curl -s -u "$(SONAR_TOKEN):" -X POST \
		"$(SONAR_URL)/api/projects/create" \
		-d "name=$(SONAR_NAME)&project=$(SONAR_PROJECT)" | grep -q '"key"' \
		&& echo "Projet créé." || echo "Projet déjà existant ou erreur (vérifier $(SONAR_URL))."

sonar: ## Analyse complète du projet (SonarQube doit être démarré)
	@if [ -z "$(SONAR_TOKEN)" ]; then \
		echo "SONAR_TOKEN manquant. Ajoute SONAR_TOKEN=<token> dans .env"; \
		exit 1; \
	fi
	@$(MAKE) sonar-setup
	@echo "Génération de la couverture backend..."
	@cd backend && go test -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "Lancement du scan SonarQube..."
	docker run --rm \
		--network sonar-net \
		-e SONAR_HOST_URL=http://komunumo-sonar:9000 \
		-e SONAR_TOKEN=$(SONAR_TOKEN) \
		-v $(PWD):/usr/src \
		sonarsource/sonar-scanner-cli \
		-Dsonar.projectKey=$(SONAR_PROJECT) \
		-Dsonar.sources=backend,frontend/app,frontend/components,frontend/lib \
		-Dsonar.exclusions=**/*_test.go,frontend/.next/**,frontend/node_modules/**,docs/**,backend/scripts/** \
		-Dsonar.tests=backend \
		-Dsonar.test.inclusions=**/*_test.go \
		-Dsonar.go.coverage.reportPaths=backend/coverage.out
	@echo "Résultats : $(SONAR_URL)/dashboard?id=$(SONAR_PROJECT)"
