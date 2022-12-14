package relayer

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/vishruthsk/viper-go/signer"

	"github.com/vishruthsk/viper-go/provider"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
	"github.com/vishruthsk/utils-go/mock-client"
)

func TestRelayer_Relay(t *testing.T) {
	c := require.New(t)

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	relayer := NewRelayer(nil, nil)
	input := &Input{}

	relay, err := relayer.Relay(input, nil)
	c.Equal(ErrNoSigner, err)
	c.Empty(relay)

	signer, err := signer.NewRandomSigner()
	c.NoError(err)

	relayer.signer = signer

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrNoProvider, err)
	c.Empty(relay)

	relayer.provider = provider.NewProvider("https://dummy.com", []string{"https://dummy.com"})

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrNoSession, err)
	c.Empty(relay)

	input.Session = &provider.Session{}

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrNoViperAAT, err)
	c.Empty(relay)

	input.ViperAAT = &provider.ViperAAT{}

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrSessionHasNoNodes, err)
	c.Empty(relay)

	input.Session.Nodes = []*provider.Node{{PublicKey: "AOG"}}

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrNoSessionHeader, err)
	c.Empty(relay)

	input.Session.Header = &provider.SessionHeader{}
	input.Node = &provider.Node{PublicKey: "PJOG"}

	relay, err = relayer.Relay(input, nil)
	c.Equal(ErrNodeNotInSession, err)
	c.Empty(relay)

	input.Node = nil

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.ClientRelayRoute),
		http.StatusInternalServerError, "../provider/samples/client_relay.json")

	relay, err = relayer.Relay(input, nil)
	c.Equal(provider.Err5xxOnConnection, err)
	c.Empty(relay)

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.ClientRelayRoute),
		http.StatusBadRequest, "../provider/samples/client_relay_error.json")

	var relayError *provider.RelayError

	relay, err = relayer.Relay(input, nil)
	c.ErrorAs(err, &relayError)
	c.Empty(relay)

	mock.AddMockedResponseFromFile(http.MethodPost, fmt.Sprintf("%s%s", "https://dummy.com", provider.ClientRelayRoute),
		http.StatusOK, "../provider/samples/client_relay.json")

	relay, err = relayer.Relay(input, nil)
	c.NoError(err)
	c.NotEmpty(relay)

	input.Node = &provider.Node{PublicKey: "AOG"}

	relay, err = relayer.Relay(input, nil)
	c.NoError(err)
	c.NotEmpty(relay)
}
