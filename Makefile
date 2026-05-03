-include .env
export

SONAR_URL        ?= http://127.0.0.1:9000
SONAR_CLOUD_URL  := https://sonarcloud.io
SONAR_CLOUD_ORG  := nicgen
SONAR_PROJECT    := assolink
SONAR_NAME       := Komunumo

.PHONY: sonar sonar-start sonar-stop sonar-setup sonar-sync-rules

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

sonar-sync-rules: ## Importe les quality profiles SonarCloud → SonarQube local
	@if [ -z "$(SONAR_CLOUD_TOKEN)" ]; then \
		echo "SONAR_CLOUD_TOKEN manquant. Ajoute dans .env le token SonarCloud (celui du secret GitHub)."; \
		exit 1; \
	fi
	@if [ -z "$(SONAR_TOKEN)" ]; then echo "SONAR_TOKEN manquant dans .env"; exit 1; fi
	@echo "Export des quality profiles depuis SonarCloud (org: $(SONAR_CLOUD_ORG))..."
	@for lang in go js ts; do \
		echo "  → $$lang"; \
		curl -sf -u "$(SONAR_CLOUD_TOKEN):" \
			"$(SONAR_CLOUD_URL)/api/qualityprofiles/export?organization=$(SONAR_CLOUD_ORG)&language=$$lang" \
			-o /tmp/sonar-profile-$$lang.xml || echo "    Pas de profil custom pour $$lang (profil Sonar Way par défaut)"; \
	done
	@echo "Import dans SonarQube local..."
	@for lang in go js ts; do \
		if [ -f /tmp/sonar-profile-$$lang.xml ] && [ -s /tmp/sonar-profile-$$lang.xml ]; then \
			curl -sf -u "$(SONAR_TOKEN):" \
				-F "backup=@/tmp/sonar-profile-$$lang.xml" \
				"$(SONAR_URL)/api/qualityprofiles/restore" \
				&& echo "  ✓ $$lang importé" || echo "  ✗ $$lang échec import"; \
		fi; \
	done
	@echo "Profils synchronisés. Vérifie : $(SONAR_URL)/profiles"

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
