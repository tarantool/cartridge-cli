local t = require('luatest')
local app = require('cartridge-cli')

local g = t.group('ignore_matching')

g.test_plain = function()
    local pattern = 'simple.test'
    t.assert(app.matching('simple.test', pattern), 'plain file')
    t.assert_not(app.matching('sample.test', pattern), 'plain file mismatch')

    pattern = '/simple/route/simple.test'
    t.assert(app.matching('/simple/route/simple.test', pattern), 'plain folder')
    t.assert_not(app.matching('/simple/way/simple.test', pattern), 'plain folder mismatch')
end

g.test_asterisk = function()
    local pattern = 'folder/*'
    t.assert(app.matching('folder/simple.test', pattern), 'one *')
    t.assert_not(app.matching('folder', pattern), 'one * mismatch')

    pattern = '*.test'
    t.assert(app.matching('simple.test', pattern), 'one * combine with str')
    t.assert_not(app.matching('simple.text', pattern), 'one * combine with str')

    pattern = '*/*/*/d'
    t.assert(app.matching('a/b/c/d', pattern), 'path composed of *')
    t.assert_not(app.matching('a/b/c/d/e', pattern), 'path composed of * mismatch')

    pattern = '*'
    t.assert(app.matching('1/b/c/d', pattern), 'all ignore-1')
    t.assert(app.matching('1', pattern), 'all ignore-2')
    t.assert(app.matching('2/text.text', pattern), 'all ignore-3')

    pattern = '/*'
    t.assert(app.matching('1', pattern), 'all in current dir')
    t.assert_not(app.matching('1/2', pattern), 'all in current dir')

    pattern = 'a/b/**'
    t.assert(app.matching('a/b/c/d/sample.test', pattern), 'end with **')
    t.assert_not(app.matching('a/c/b/e/sample.test', pattern), 'end with ** mismatch')

    pattern = 'a/b/**/f/g/**'
    t.assert(app.matching('a/b/c/d/e/f/g/sample.test', pattern), 'with **')
    t.assert_not(app.matching('a/b/c/d/e/g/sample.test', pattern), 'with ** mismatch')

    pattern = '**/*.tar.gz'
    t.assert(app.matching('folder/archive.tar.gz', pattern), 'start with **')
    t.assert_not(app.matching('archive.tar.gz', pattern), 'start with ** mismatch')
end


g.test_rec = function()
    local pattern = 'simple'

    t.assert(app.matching('sample/simple', pattern), 'search file or dir recursive')
    t.assert(app.matching('sample/simple/', pattern), 'search dir recursive')
end

g.test_more = function()
    local pattern = 'dir/file1.txt'

    t.assert(app.matching(pattern, 'dir/file?.txt'), 'single character wildcard')
    t.assert_not(app.matching(pattern, 'dir/files?.txt'), 'single character wildcards mismatch')

    pattern = 'dir/file'
    t.assert(app.matching(pattern .. '1', 'dir/*[0-9]'), 'character ranges')
    t.assert_not(app.matching(pattern .. 'a', 'dir/*[0-9]'), 'character ranges mismatch')

    pattern = 'sample/'
    t.assert(app.matching(pattern, 'sample'), 'dir and pattern without backslash')
    t.assert(app.matching(pattern, 'sample/'), 'dir and pattern with backslash')

    pattern = 'sample'
    t.assert_not(app.matching(pattern, 'sample/'), 'file and pattern with backslash mismatch')
    t.assert(app.matching(pattern .. '.pyc', '*.py[co]'), 'charset first')
    t.assert(app.matching(pattern .. '.pyo', '*.py[co]'), 'charset second')
    t.assert_not(app.matching(pattern .. '.py', '*.py?[co]'), 'charset mismatch')

    pattern = 'qwerty'
    t.assert(app.matching(pattern, '*[!a-d]'), 'charset negate')
    t.assert_not(app.matching(pattern, '*[!a-z]'), 'charset negate mismatch')
end
