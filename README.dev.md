## How to run tests

### Requirements

1. Install [Go](https://go.dev/doc/install) 1.18.
   ```bash
   go version
   ```

2. Install [Python](https://www.python.org/downloads/) 3.x and [pip](https://pypi.org/project/pip/).
   ```bash
   python3 --version
   pip3 --version
   ```

3. Install `unzip`, `rpm` and `cpio` packages.
   ```bash
   unzip -v
   rpm --version
   cpio --version
   ```
   
4. Install [git](https://git-scm.com/downloads).
   ```bash
   git --version
   ```
      
5. Install [docker](https://www.docker.com/get-started).
   ```bash
   docker --version
   ```
    
6. Install [Tarantool](https://www.tarantool.io/en/download/os-installation/) (1.10 or 2.x version).
   ```bash
   tarantool --version
   ```

7. Install [mage](https://github.com/magefile/mage).
   ```bash
   mage --version
   ```
   If something went wrong, this may help.
   ```bash
   export PATH=$(go env GOPATH)/bin:$PATH
   ```
     
8. Install [pytest](https://docs.pytest.org/en/6.2.x/getting-started.html).
   ```bash
   python3 -m pytest --version
   ```

9. Clone this repo.
   ```bash
   git clone git@github.com:tarantool/cartridge-cli.git
   cd ./cartridge-cli
   ```
   
10. To run tests, git user must be configured. For example,
    ```bash
    git config --global user.email "test@tarantool.io"
    git config --global user.name "Tar Antool"
    ```
   
11. Install pytest dependencies.
    ```bash
    pip3 install -r test/requirements.txt
    ```
    
12. Install luacheck.
    ```bash
    tarantoolctl rocks install luacheck
    ```

All remaining dependencies (like code generation) will be invoked with mage if needed.

### Test run

To run all tests, call
```bash
mage test
```

You can run specific test sections.
```bash
# Static code analysis (including tests code).
mage lint
# Go unit tests.
mage unit
# pytest integration tests.
mage integration
# Run test example with pytest.
mage testExamples
# pytest end-to-end tests for packages.
mage e2e
```
