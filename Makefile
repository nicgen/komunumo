-include .env
export

SONAR_URL     ?= http://localhost:9000
SONAR_PROJECT := nicgen_komunumo

.PHONY: sonar sonar-start sonar-stop

sonar-start:
	docker compose -f docker-compose.sonar.yml up -d
	@echo "SonarQube disponible sur $(SONAR_URL) (attendre ~30s au premier lancement)"

sonar-stop:
	docker compose -f docker-compose.sonar.yml down

sonar: ## Analyse complète du projet (SonarQube doit être démarré)
	@if [ -z "$(SONAR_TOKEN)" ]; then \
		echo "SONAR_TOKEN manquant. Créer un token sur $(SONAR_URL)/account/security puis :"; \
		echo "  Ajoute SONAR_TOKEN=<token> dans .env"; \
		exit 1; \
	fi
	@echo "Génération de la couverture backend..."
	cd backend && go test -coverprofile=coverage.out ./cmd/... ./internal/...
	@echo "Lancement du scan SonarQube..."
	docker run --rm \
		--network host \
		-e SONAR_HOST_URL=$(SONAR_URL) \
		-e SONAR_TOKEN=$(SONAR_TOKEN) \
		-v $(PWD):/usr/src \
		sonarsource/sonar-scanner-cli \
		-Dsonar.projectKey=$(SONAR_PROJECT) \
		-Dsonar.sources=backend,frontend/app,frontend/components,frontend/lib \
		-Dsonar.exclusions=**/*_test.go,frontend/.next/**,frontend/node_modules/**,docs/**,backend/scripts/** \
		-Dsonar.tests=backend \
		-Dsonar.test.inclusions=**/*_test.go \
		-Dsonar.go.coverage.reportPaths=backend/coverage.out \
		-Dsonar.host.url=$(SONAR_URL)
	@echo "Résultats : $(SONAR_URL)/dashboard?id=$(SONAR_PROJECT)"
