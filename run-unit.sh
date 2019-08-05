# Unit tests
for test_file in $(find tests/unit -type f -prune -name test_\*.lua)
do
	printf '\n[--- Executing test %s ...]\n' "$test_file"
	tarantool $test_file
done
