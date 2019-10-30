bootstrap: .rocks

.rocks:
	tarantoolctl rocks install luatest 0.3.0
	tarantoolctl rocks install luacheck 0.25.0

tmp/sdk-1.10:
	echo "Using tarantool-enterprise-bundle ${BUNDLE_VERSION}"
	curl -O -L https://tarantool:${DOWNLOAD_TOKEN}@download.tarantool.io/enterprise/tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz
	tar -xzf tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz -C ./tmp
	mv tmp/tarantool-enterprise tmp/sdk-1.10
	rm -f tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz

.PHONY: lint
lint: bootstrap
	.rocks/bin/luacheck ./

.PHONY: test
test: luatest pytest test-getting-started

.PHONY: luatest
luatest: bootstrap
	.rocks/bin/luatest

.PHONY: pytest
pytest: bootstrap
	python3.6 -m pytest -vvl

.PHONY: test-getting-started
test-getting-started: bootstrap
	cd test/examples/getting-started-app; \
	sh test_start.sh ../../../examples/getting-started-app

.PHONY: ci_prepare
ci_prepare:
	git config --global user.email "test@tarantool.io"
	git config --global user.name "Test Tarantool"

.PHONY: clean
clean:
	rm -rf .rocks tmp/sdk-1.10
