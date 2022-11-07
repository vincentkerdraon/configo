ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))


#----- common

.PHONY: all
all: go_test_no_cache clean build 

.PHONY: help
help:
	@grep -hE '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-17s\033[0m %s\n", $$1, $$2}'

.PHONY: fmt
fmt: go_fmt ## Format source code

.PHONY: lint
lint: go_lint ## Lint source code

.PHONY: clean
clean: go_clean ## Remove tmp files

.PHONY: build
build: go_build ## Build what can be

.PHONY: test
test: go_test ## Run unit tests

#----- go


MODULE_NAME=$$(awk -F/ '/module github.com\/vincentkerdraon\// {print tolower($$3)}' go.mod)
MODULE_NAME_FULL=$$(cat go.mod | grep "^module github.com/vincentkerdraon" | cut -d' ' -f2- )
GO_COMPILE_OPTION=CGO_ENABLED=0
GO_COMPILE_LDFLAGS_ADDITIONAL=	
GO_COMPILE_ENV=

.PHONY: go_quality
go_quality: go_fmt go_lint_filtered_non_blocking go_staticcheck go_vet go_vulncheck  ## Run various tools like fmt, lint, staticcheck, vet, vulncheck

.PHONY: go_fmt
go_fmt: ## Format Go source code
	gofmt -w $(shell find . -name '*.go' | grep -v vendor)

.PHONY: go_install_lint
go_install_lint: ## install golint
	go install golang.org/x/lint/golint@latest

.PHONY: go_lint
go_lint: go_install_lint ## format files with go style + report warnings
	## when 1 go.mod: golint -set_exit_status=1 $$(shell find . -name '*.go' | grep -v vendor)
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do echo "-- lint $$file"; cd $$(dirname "$$file") && golint && cd - >/dev/null ; done

.PHONY: go_lint_filtered_non_blocking
go_lint_filtered_non_blocking: go_install_lint ## format files with go style (skip some rules).
	## when 1 go.mod: golint -set_exit_status=1 $$(shell find . -name '*.go' | grep -v vendor)
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do echo "-- lint $$file"; cd $$(dirname "$$file") && golint ./... | grep -v "should have comment or be unexported" | grep -v "which can be annoying to use" || true && cd - >/dev/null ; done

.PHONY: go_install_staticcheck
go_install_staticcheck: ## install staticcheck
	go install honnef.co/go/tools/cmd/staticcheck@latest

.PHONY: go_staticcheck
go_staticcheck: go_install_staticcheck ## linter for go
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do echo "-- lint $$file"; cd $$(dirname "$$file") && staticcheck ./... && cd - >/dev/null ; done

.PHONY: go_vet
go_vet: ## indicate potential errors (not blocking the build)
	go vet $$(go list ./... | grep -v /vendor/)

.PHONY: go_install_vulncheck
go_install_vulncheck: ## install govulncheck
	go install golang.org/x/vuln/cmd/govulncheck@latest

.PHONY: go_vulncheck
go_vulncheck: go_install_vulncheck ## to analyze your codebase and surface known vulnerabilities.
	## when 1 go.mod: govulncheck ./... 
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; govulncheck ./... ; done

.PHONY: go_work
go_work: ## create go workspace. Helps a lot on vscode for example when multiple go modules.
	rm -f "$(ROOT_DIR)/go.work" "$(ROOT_DIR)/go.work.sum"; go work init && go work use -r .

.PHONY: go_clean
go_clean: ## remove tmp files
	go clean $$(go list ./... ) 
	go clean -r -cache -testcache
	rm -rf "$(MODULE_NAME)"

.PHONY: go_clean_dep_cache
go_clean_dep_cache: ## purge go dependencies. This can help to fix Go dependency loop
	## Purging cache of dependencies used in this project
	go clean -modcache $$(go list ./... ) 

.PHONY: go_doc
go_doc: ## show godoc for local code (even private repo).
	@echo "http://localhost:7070/pkg/$(MODULE_NAME_FULL)/"
	godoc -http=:7070 -play=true -notes="BUG|TODO|${FIX_ME}"

.PHONY: go_get
go_get: ## not upgrading to new dep, only using the content of go.sum (deprecated, see go_mod_tidy)
	## when 1 go.mod: go get $$(go list ./... ) 
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; go get $$(go list ./...) ; done

.PHONY: go_mod_tidy
go_mod_tidy: ## not upgrading to new dep, only using the content of go.sum (prefered to go_get)
	## when 1 go.mod: go get $$(go list ./... ) 
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; go mod tidy ; done

.PHONY: go_generate
go_generate: ## find and run go generator
	go generate ./... || exit 1

.PHONY: go_update_dep
go_upgrade_dep: ## find newest version
	## This will fail if one of the modules fail.
	## when 1 go.mod: go get -u $$(go list ./... ); go mod tidy
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do \
		cd $$(dirname "$$file"); \
		echo "-- in $$(pwd)"; \
		go get -u -t $$(go list ./...); \
		if [ $$? -ne 0 ]; then \
			echo && echo "fail update (1)"; \
			exit 0; \
		fi; \
		go mod tidy ; \
		if [ $$? -ne 0 ]; then \
			echo && echo "fail tidy (1)"; \
			exit 0; \
		fi; \
	done

.PHONY: go_update_dep
go_upgrade_dep_ignore_errors:
	# this will keep upgrading even if one of the modules fails
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do \
		cd $$(dirname "$$file"); \
		echo "-- in $$(pwd)"; \
		go get -u -t $$(go list ./...); \
		go mod tidy ; \
	done

.PHONY: go_upgrade_dep_purge
go_upgrade_dep_purge:
	find "$(ROOT_DIR)" -type f -name 'go.sum' -exec rm -f {} +
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do echo "-- purge $$file"; sed -n '/require (/q;p' $$file > $$file.tmp; mv $$file.tmp $$file; done

.PHONY: go_upgrade_dep_force
go_upgrade_dep_force: go_upgrade_dep_purge ## find dependencies from scatch
	$(MAKE) go_upgrade_dep_ignore_errors && \
	## T O D O: change go_upgrade_dep_ignore_errors to return an error if at least 1 of the module failed.\
	if [ $$? -eq 0 ]; then \
		echo && echo "success force upgrade (1)"; \
		exit 0; \
	fi; \
	echo && echo "-- missing some entries, retrying" && echo; \
	$(MAKE) go_upgrade_dep && \
	if [ $$? -eq 0 ]; then \
		echo && echo "success force upgrade (2)"; \
		exit 0; \
	fi; \
	echo && echo "-- still missing some entries, you probably need to run \"go get [package]\" in a specific order"; \
	exit 1 \
	#Makefile code: of course `$(MAKE) go_upgrade_dep || echo ...` won't work.

.PHONY: go_test
go_test: ## run unit tests
	## when 1 go.mod: $(GO_COMPILE_OPTION) go test $$(go list ./... ) 
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; $(GO_COMPILE_OPTION) go test $$(go list ./... ) || exit 10 ; done


.PHONY: go_test_no_cache
go_test_no_cache: ## force run unit tests
	# Running all the tests without cache (and in parallel). 
	# Hence, some tests can take longer than usual to exec
	# Also, some borderline concurrency errors migh show up.
	## when 1 go.mod: $(GO_COMPILE_OPTION) go test $$(go list ./... ) -count=1
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; $(GO_COMPILE_OPTION) go test $$(go list ./... ) -count=1 || exit 10 ; done

.PHONY: go_test_race
go_test_race: ## force run unit tests, need CGO for -race.
	# same as go_test_no_cache, but with CGO activated.
	find "$(ROOT_DIR)" -type f -name 'go.mod' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; CGO_ENABLED=1 go test -race $$(go list ./... ) -count=1 || exit 10 ; done

.PHONY: go_build
go_build: ## run go build
	## LDFlags can be used even if not declared in the app => ignored.
	## Note: no need to pass go version here, available from code directly.
	## Note: using -trimpath improves build reproducibility but is too long.
	BUILD_GO_LDFLAGS="$$BUILD_GO_LDFLAGS $(GO_COMPILE_LDFLAGS_ADDITIONAL) " && \
	## when 1 go.mod: $(GO_COMPILE_OPTION) go build -v $$(BUILD_GO_LDFLAGS) \
	find "$(ROOT_DIR)" -type f -name 'main.go' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; $(GO_COMPILE_OPTION) $(GO_COMPILE_ENV) go build -ldflags="$$BUILD_GO_LDFLAGS" ; done

.PHONY: go_build_race
go_build_race: ## run go build, detect races. need CGO for -race.
	## go_build_race is similar to go_build, but needs CGO to run -race
	BUILD_GO_LDFLAGS="$$BUILD_GO_LDFLAGS $(GO_COMPILE_LDFLAGS_ADDITIONAL) " && \
	find "$(ROOT_DIR)" -type f -name 'main.go' | while IFS= read -r file; do cd $$(dirname "$$file"); echo "-- in $$(pwd)" ; CGO_ENABLED=1 $(GO_COMPILE_ENV) go build -race -ldflags="$$BUILD_GO_LDFLAGS" ; done
