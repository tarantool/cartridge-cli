package.path = package.path .. ";../?;?"

local tap = require('tap')
local app = require('tarantoolapp')

local test = tap.test('tarantoolapp.ignore matching')

test:plan(4)

local function plain_test(test)
    test:plan(4)

    local pattern = 'simple.test'
    test:ok(app.matching('simple.test', pattern),
        'plain file')
    test:ok(not app.matching('sample.test', pattern),
        'plain file mismatch')

    pattern = '/simple/route/simple.test'
    test:ok(app.matching('/simple/route/simple.test', pattern),
        'plain folder')
    test:ok(not app.matching('/simple/way/simple.test', pattern),
        'plain folder mismatch')
end

test:test(
    'plain_test',
    plain_test)

local function asterisk_test(test)
    test:plan(17)
    local pattern = 'folder/*'
    test:ok(app.matching('folder/simple.test', pattern),
        'one *')
    test:ok(not app.matching('folder', pattern),
        'one * mismatch')

    pattern = '*.test'
    test:ok(app.matching('simple.test', pattern),
        'one * combine with str')
    test:ok(not app.matching('simple.text', pattern),
        'one * combine with str')

    pattern = '*/*/*/d'
    test:ok(app.matching('a/b/c/d', pattern),
        'path composed of *')
    test:ok(not app.matching('a/b/c/d/e', pattern),
        'path composed of * mismatch')
    
    pattern = '*'
    test:ok(app.matching('1/b/c/d', pattern),
        'all ignore-1')
    test:ok(app.matching('1', pattern),
        'all ignore-2')
    test:ok(app.matching('2/text.text', pattern),
        'all ignore-3')

    pattern = '/*'
    test:ok(app.matching('1', pattern),
        'all in current dir')
    test:ok(not app.matching('1/2', pattern),
        'all in current dir')
    
    pattern = 'a/b/**'
    test:ok(app.matching('a/b/c/d/sample.test', pattern),
        'end with **')
    test:ok(not app.matching('a/c/b/e/sample.test', pattern),
        'end with ** mismatch')

    pattern = 'a/b/**/f/g/**'
    test:ok(app.matching('a/b/c/d/e/f/g/sample.test', pattern),
        'with **')
    test:ok(not app.matching('a/b/c/d/e/g/sample.test', pattern),
        'with ** mismatch')

    pattern = '**/*.tar.gz'
    test:ok(app.matching('folder/archive.tar.gz', pattern),
        'start with **')
    test:ok(not app.matching('archive.tar.gz', pattern),
        'start with ** mismatch')
end


test:test(
    'asterisk_test',
    asterisk_test)

local function rec_test(test)
    test:plan(2)
    local pattern = 'simple'

    test:ok(app.matching('sample/simple', pattern),
        'search file or dir recursive')
    test:ok(app.matching('sample/simple/', pattern),
        'search dir recursive')

end

test:test(
    'rec_test',
    rec_test)

local function more_test(test)
    test:plan(12)
    local pattern = 'dir/file1.txt'
    
    test:ok(app.matching(pattern, 'dir/file?.txt'),
        'single character wildcard')
    
    test:ok(not app.matching(pattern, 'dir/files?.txt'),
        'single character wildcards mismatch')

    pattern = 'dir/file'
    test:ok(app.matching(pattern .. '1', 'dir/*[0-9]'),
        'character ranges')
    
    test:ok(not app.matching(pattern .. 'a', 'dir/*[0-9]'),
        'character ranges mismatch')

    pattern = 'sample/'
    test:ok(app.matching(pattern, 'sample'),
        'dir and pattern without backslash')
    
    test:ok(app.matching(pattern, 'sample/'),
        'dir and pattern with backslash')

    pattern = 'sample'
    test:ok(not app.matching(pattern, 'sample/'),
        'file and pattern with backslash mismatch')

    test:ok(app.matching(pattern .. '.pyc', '*.py[co]'),
        'charset first')
    
    test:ok(app.matching(pattern .. '.pyo', '*.py[co]'),
        'charset second')

    test:ok(not app.matching(pattern .. '.py', '*.py?[co]'),
        'charset mismatch')

    pattern = 'qwerty'
    test:ok(app.matching(pattern, '*[!a-d]'),
        'charset negate')
    test:ok(not app.matching(pattern, '*[!a-z]'),
        'charset negate mismatch')
end

test:test(
    'more_test',
    more_test)
