default:

clean:
	@rm -rf bin

prerelease: clean
ifndef tag
	$(error tag is undefined)
endif
	@for os in linux darwin windows; do \
		if [ $$os = "windows" ]; then ext=".exe"; fi ;\
		arch=amd64 ;\
		name="ynab-simplefin-sync-$$os-$$arch$$ext" ;\
		echo "Building $$name" ;\
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -ldflags='-w -s' -o bin/release/$$name ;\
	done
	@upx -6 bin/release/*

release: prerelease
	@if ! git diff --quiet; then \
		echo "Can't release... dirty worktree" ;\
		echo ;\
		git status ;\
		exit 1 ;\
	fi
	if ! gh release view $(tag); then \
		gh release create $(tag) --generate-notes --draft ;\
	fi
	while ! gh release view $(tag); do sleep 1; done
	gh release upload $(tag) --clobber bin/release/*
