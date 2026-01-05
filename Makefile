.PHONY: run test clean release

# Release: bump version and create git tag
# Usage: make release BUMP=patch|minor|major
release:
ifndef BUMP
	$(error BUMP is required. Usage: make release BUMP=patch|minor|major)
endif
	@VERSION=$$(grep -E '^version = ' pyproject.toml | sed 's/version = "\(.*\)"/\1/'); \
	IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
	case "$(BUMP)" in \
		patch) PATCH=$$((PATCH + 1));; \
		minor) MINOR=$$((MINOR + 1)); PATCH=0;; \
		major) MAJOR=$$((MAJOR + 1)); MINOR=0; PATCH=0;; \
		*) echo "Invalid BUMP value: $(BUMP). Use patch, minor, or major"; exit 1;; \
	esac; \
	NEW_VERSION="$$MAJOR.$$MINOR.$$PATCH"; \
	sed -i '' "s/^version = \".*\"/version = \"$$NEW_VERSION\"/" pyproject.toml; \
	echo "Bumped version: $$VERSION -> $$NEW_VERSION"; \
	jj commit -m "chore(release): v$$NEW_VERSION"; \
	echo "Created jj commit: chore(release): v$$NEW_VERSION"; \
	git tag "v$$NEW_VERSION"; \
	echo "Created tag: v$$NEW_VERSION"

# Run the CLI
run:
	uv run kekkai

# Run tests
test:
	uv run --with pytest pytest tests/ -v

# Clean build artifacts
clean:
	rm -rf .venv .pytest_cache __pycache__ src/kekkai/__pycache__ tests/__pycache__
