bootstrap: .rocks

.rocks:
	tarantoolctl rocks install luatest 0.5.0
	tarantoolctl rocks install luacov 0.13.0
	tarantoolctl rocks install luacheck 0.25.0

tmp/sdk-1.10:
	echo "Using tarantool-enterprise-bundle ${BUNDLE_VERSION}"
	curl -O -L https://tarantool:${DOWNLOAD_TOKEN}@download.tarantool.io/enterprise/tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz
	tar -xzf tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz -C ./tmp
	mv tmp/tarantool-enterprise tmp/sdk-1.10
	rm -f tarantool-enterprise-bundle-${BUNDLE_VERSION}.tar.gz

tmp/cache-image.tar:
	docker build \
		--tag cache-image \
		--target ${CACHE_IMAGE_TARGET} \
		- < Dockerfile.cache
	docker save -o tmp/cache-image.tar cache-image

.PHONY: lint
lint: bootstrap
	.rocks/bin/luacheck ./
	flake8

.PHONY: test
test: unit integration test-examples e2e

python_deps:
	pip3 install -r test/requirements.txt

.PHONY: integration
integration:
	python3 -m pytest test/integration

.PHONY: e2e
e2e:
	python3 -m pytest test/e2e

.PHONY: unit
unit: bootstrap
	rm -f tmp/luacov.*
	.rocks/bin/luatest -v --coverage && .rocks/bin/luacov .
	grep -A999 '^Summary' tmp/luacov.report.out

.PHONY: test-examples
test-examples:
	python3 -m pytest test/examples

.PHONY: ci_prepare
ci_prepare: python_deps
	git config --global user.email "test@tarantool.io"
	git config --global user.name "Test Tarantool"

.PHONY: clean
clean:
	rm -rf .rocks build build.luarocks .cache CMakeCache.txt CMakeFiles
