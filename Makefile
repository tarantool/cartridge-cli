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
lint:
	flake8

.PHONY: test
test: unit pytest # test-getting-started

python_deps:
	pip3.6 install -r test/python/requirements.txt

.PHONY: pytest
pytest:
	python3.6 -m pytest -vvl --durations=10

.PHONY: unit
unit:
	go test -v ./src/rpm

# .PHONY: test-getting-started
# test-getting-started: bootstrap
# 	cd test/examples/getting-started-app; \
# 		sh test_start.sh ../../../examples/getting-started-app;
# 	cd ./examples/getting-started-app; \
# 		.rocks/bin/luatest -v
# 	.rocks/bin/luacheck ./examples/getting-started-app \
# 		--exclude-files **/.rocks/*

.PHONY: ci_prepare
ci_prepare: python_deps
	git config --global user.email "test@tarantool.io"
	git config --global user.name "Test Tarantool"

.PHONY: clean
clean:
	rm -rf .rocks build build.luarocks .cache CMakeCache.txt CMakeFiles
