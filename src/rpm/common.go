package rpm

import (
	"bytes"
	"encoding/binary"
)

func packValues(values ...interface{}) []byte {
	buf := &bytes.Buffer{}

	for _, v := range values {
		binary.Write(buf, binary.BigEndian, v)
	}

	return buf.Bytes()
}

const (
	defaultFileUser   = "root"
	defaultFileGroup  = "root"
	defaultFileLang   = ""
	defaultFileLinkTo = ""
	emptyDigest       = ""
)

var (
	knownFiles = map[string]struct{}{
		".":                    struct{}{},
		"bin":                  struct{}{},
		"usr":                  struct{}{},
		"usr/bin":              struct{}{},
		"usr/local":            struct{}{},
		"usr/local/bin":        struct{}{},
		"usr/share":            struct{}{},
		"usr/share/tarantool":  struct{}{},
		"usr/lib":              struct{}{},
		"usr/lib/tmpfiles.d":   struct{}{},
		"var":                  struct{}{},
		"var/lib":              struct{}{},
		"var/lib/tarantool":    struct{}{},
		"var/run":              struct{}{},
		"var/log":              struct{}{},
		"etc":                  struct{}{},
		"etc/tarantool":        struct{}{},
		"etc/tarantool/conf.d": struct{}{},
		"etc/systemd":          struct{}{},
		"etc/systemd/system":   struct{}{},
	}
)
