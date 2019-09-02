bootstrap: .rocks

.rocks:
	tarantoolctl rocks install luatest 0.2.0
	tarantoolctl rocks install luacheck

.PHONY: lint
lint: bootstrap
	.rocks/bin/luacheck ./

.PHONY: test
test: luatest pytest

.PHONY: luatest
luatest: bootstrap
	.rocks/bin/luatest

.PHONY: pytest
pytest: bootstrap
	python3.6 -m pytest -vvl

.PHONY: ci_prepare
ci_prepare:
	git config --global user.email "test@tarantool.io"
	git config --global user.name "Test Tarantool"

.PHONY: clean
clean:
	rm -rf .rocks
