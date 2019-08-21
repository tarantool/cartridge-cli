# Cartridge Command Line Interface

## Building

Simply run

```sh
tarantoolctl rocks install cartridge-cli
```

## Running end-to-end tests

```sh
vagrant up
vagrant ssh 1_10 < tests/end-to-end.sh
vagrant halt
```
