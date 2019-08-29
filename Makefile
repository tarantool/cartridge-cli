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

.PHONY: clean
clean:
	rm -rf .rocks
