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

# .PHONY: test-getting-started
# test-getting-started: bootstrap
# 	cd test/examples/getting-started-app; \
# 		sh test_start.sh ../../../examples/getting-started-app;
# 	cd ./examples/getting-started-app; \
# 		.rocks/bin/luatest -v
# 	.rocks/bin/luacheck ./examples/getting-started-app \
# 		--exclude-files **/.rocks/*
