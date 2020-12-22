package connect

import (
	"sort"
	"strings"

	"github.com/adam-hanna/arrayOperations"
	"github.com/c-bata/go-prompt"
	"github.com/tarantool/cartridge-cli/cli/common"
	"github.com/tarantool/cartridge-cli/cli/context"
)

const (
	MaxHistoryLines = 10000

	TCPNetwork  = "tcp"
	UnixNetwork = "unix"
)

type ConnOpts struct {
	Network  string
	Address  string
	Username string
	Password string
}

type GetRawSuggestionsFunc func(console *Console, lastWord string) interface{}

func getConnOpts(connString string, ctx *context.Ctx) (*ConnOpts, error) {
	connOpts := ConnOpts{
		Username: ctx.Connect.Username,
		Password: ctx.Connect.Password,
	}

	connStringParts := strings.SplitN(connString, "@", 2)
	address := connStringParts[len(connStringParts)-1]

	if len(connStringParts) > 1 {
		authString := connStringParts[0]
		authStringParts := strings.SplitN(authString, ":", 2)

		if connOpts.Username == "" {
			connOpts.Username = authStringParts[0]
		}
		if len(authStringParts) > 1 && connOpts.Password == "" {
			connOpts.Password = authStringParts[1]
		}
	}

	addrLen := len(address)
	switch {
	case addrLen > 0 && (address[0] == '.' || address[0] == '/'):
		connOpts.Network = UnixNetwork
		connOpts.Address = address
	case addrLen >= 7 && address[0:7] == "unix://":
		connOpts.Network = UnixNetwork
		connOpts.Address = address[7:]
	case addrLen >= 5 && address[0:5] == "unix:":
		connOpts.Network = UnixNetwork
		connOpts.Address = address[5:]
	case addrLen >= 6 && address[0:6] == "unix/:":
		connOpts.Network = UnixNetwork
		connOpts.Address = address[6:]
	case addrLen >= 6 && address[0:6] == "tcp://":
		connOpts.Network = TCPNetwork
		connOpts.Address = address[6:]
	case addrLen >= 4 && address[0:4] == "tcp:":
		connOpts.Network = TCPNetwork
		connOpts.Address = address[4:]
	default:
		connOpts.Network = TCPNetwork
		connOpts.Address = address
	}

	return &connOpts, nil
}

func getSuggestions(console *Console, in prompt.Document,
	getRawSuggestionsFunc GetRawSuggestionsFunc) []prompt.Suggest {

	if len(in.Text) == 0 {
		return nil
	}

	lastWordStart := in.FindStartOfPreviousWordUntilSeparator(tarantoolWordSeparators)
	lastWord := in.Text[lastWordStart:]

	if len(lastWord) == 0 {
		return nil
	}

	suggestionsRaw := getRawSuggestionsFunc(console, lastWord)

	suggestionsTexts, err := common.ConvertToStringsSlice(suggestionsRaw)
	if err != nil {
		return nil
	}

	suggestionsTexts = arrayOperations.DifferenceString(suggestionsTexts)
	if len(suggestionsTexts) == 0 {
		return nil
	}

	sort.Strings(suggestionsTexts)

	suggestions := make([]prompt.Suggest, len(suggestionsTexts))
	for i, suggestionText := range suggestionsTexts {
		suggestions[i] = prompt.Suggest{
			Text: suggestionText,
		}
	}

	return suggestions
}
