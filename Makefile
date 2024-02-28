.PHONY: bump_version major minor patch

major:
	@$(MAKE) TYPE=major bump_version

minor:
	@$(MAKE) TYPE=minor bump_version

patch:
	@$(MAKE) TYPE=patch bump_version

CURRENT=$(shell git describe --tags --abbrev=0)
MAJOR=$(shell echo $(CURRENT) | cut -d. -f1)
MINOR=$(shell echo $(CURRENT) | cut -d. -f2)
PATCH=$(shell echo $(CURRENT) | cut -d. -f3)
ifeq ($(TYPE),major)
  NEW_VERSION := $(shell echo $(MAJOR)+1 | bc).0.0
else ifeq ($(TYPE),minor)
  NEW_VERSION := $(MAJOR).$(shell echo $(MINOR)+1 | bc).0
else ifeq ($(TYPE),patch)
  NEW_VERSION := $(MAJOR).$(MINOR).$(shell echo $(PATCH)+1 | bc)
endif
bump_version:
	@echo "Bumping version to $(NEW_VERSION)"
	@sed -i '' -e 's/"version": "jorge v.*"/"version": "jorge v$(NEW_VERSION)"/' main.go
	git add main.go
	git commit -m "v$(NEW_VERSION)"
	git tag -a $(NEW_VERSION) -m "v$(NEW_VERSION)"
	git push origin
	git push origin --tags
	make docs

docs:
	jorge build docs
	rsync -vPrz --delete docs/target/ root@olano.dev:/var/www/jorge
