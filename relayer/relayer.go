// Package relayer is a helper for doing relays with simpler input that just using the package Provider
// Underneath uses the package Provider
package relayer

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"math/big"

	"github.com/vishruthsk/viper-go/provider"

	"golang.org/x/crypto/sha3"
)

var (
	// ErrNoSigner error when no signer is provided
	ErrNoSigner = errors.New("no signer provided")
	// ErrNoSession error when no session is provided
	ErrNoSession = errors.New("no session provided")
	// ErrNoSessionHeader error when no session header is provided
	ErrNoSessionHeader = errors.New("no session header provided")
	// ErrNoProvider error when no provider is provided
	ErrNoProvider = errors.New("no provider provided")
	// ErrNoViperAAT error when no Viper AAT is provided
	ErrNoViperAAT = errors.New("no Viper AAT provided")
	// ErrSessionHasNoNodes error when provided session has no nodes
	ErrSessionHasNoNodes = errors.New("session has no nodes")
	// ErrNodeNotInSession error when given node is not in session
	ErrNodeNotInSession = errors.New("node not in session")
)

// Provider interface representing provider functions necessary for Relayer Package
type Provider interface {
	Relay(rpcURL string, input *provider.RelayInput, options *provider.RelayRequestOptions) (*provider.RelayOutput, error)
}

// Signer interface representing signer functions necessary for Relayer Package
type Signer interface {
	Sign(payload []byte) (string, error)
}

// Relayer implementation of relayer interface
type Relayer struct {
	signer   Signer
	provider Provider
}

// NewRelayer returns instance of Relayer with given input
func NewRelayer(signer Signer, provider Provider) *Relayer {
	return &Relayer{
		signer:   signer,
		provider: provider,
	}
}

func (r *Relayer) validateRelayRequest(input *Input) error {
	if r.signer == nil {
		return ErrNoSigner
	}

	if r.provider == nil {
		return ErrNoProvider
	}

	if input.Session == nil {
		return ErrNoSession
	}

	if input.ViperAAT == nil {
		return ErrNoViperAAT
	}

	if len(input.Session.Nodes) == 0 {
		return ErrSessionHasNoNodes
	}

	if input.Session.Header == nil {
		return ErrNoSessionHeader
	}

	return nil
}

func getNode(input *Input) (*provider.Node, error) {
	if input.Node == nil {
		return GetRandomSessionNode(input.Session)
	}

	if !IsNodeInSession(input.Session, input.Node) {
		return nil, ErrNodeNotInSession
	}

	return input.Node, nil
}

func (r *Relayer) getSignedProofBytes(proof *provider.RelayProof) (string, error) {
	proofBytes, err := GenerateProofBytes(proof)
	if err != nil {
		return "", err
	}

	return r.signer.Sign(proofBytes)
}

// Relay does relay request with given input
func (r *Relayer) Relay(input *Input, options *provider.RelayRequestOptions) (*Output, error) {
	err := r.validateRelayRequest(input)
	if err != nil {
		return nil, err
	}

	node, err := getNode(input)
	if err != nil {
		return nil, err
	}

	relayPayload := &provider.RelayPayload{
		Data:    input.Data,
		Method:  input.Method,
		Path:    input.Path,
		Headers: input.Headers,
	}

	relayMeta := &provider.RelayMeta{
		BlockHeight: input.Session.Header.SessionHeight,
	}

	hashedReq, err := HashRequest(&RequestHash{
		Payload: relayPayload,
		Meta:    relayMeta,
	})
	if err != nil {
		return nil, err
	}

	entropy, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return nil, err
	}

	signedProofBytes, err := r.getSignedProofBytes(&provider.RelayProof{
		RequestHash:        hashedReq,
		Entropy:            entropy.Int64(),
		SessionBlockHeight: input.Session.Header.SessionHeight,
		ServicerPubKey:     node.PublicKey,
		Blockchain:         input.Blockchain,
		AAT:                input.ViperAAT,
	})
	if err != nil {
		return nil, err
	}

	relayProof := &provider.RelayProof{
		RequestHash:        hashedReq,
		Entropy:            entropy.Int64(),
		SessionBlockHeight: input.Session.Header.SessionHeight,
		ServicerPubKey:     node.PublicKey,
		Blockchain:         input.Blockchain,
		AAT:                input.ViperAAT,
		Signature:          signedProofBytes,
	}

	relay := &provider.RelayInput{
		Payload: relayPayload,
		Meta:    relayMeta,
		Proof:   relayProof,
	}

	relayOutput, err := r.provider.Relay(node.ServiceURL, relay, options)
	if err != nil {
		return nil, err
	}

	return &Output{
		RelayOutput: relayOutput,
		Proof:       relayProof,
		Node:        node,
	}, nil
}

// GetRandomSessionNode returns a random node from given session
func GetRandomSessionNode(session *provider.Session) (*provider.Node, error) {
	index, err := rand.Int(rand.Reader, big.NewInt(int64(len(session.Nodes))))
	if err != nil {
		return nil, err
	}

	return session.Nodes[index.Int64()], nil
}

// IsNodeInSession verifies if given node is in given session
func IsNodeInSession(session *provider.Session, node *provider.Node) bool {
	for _, sessionNode := range session.Nodes {
		if sessionNode.PublicKey == node.PublicKey {
			return true
		}
	}

	return false
}

// GenerateProofBytes returns relay proof as encoded bytes
func GenerateProofBytes(proof *provider.RelayProof) ([]byte, error) {
	token, err := HashAAT(proof.AAT)
	if err != nil {
		return nil, err
	}

	proofMap := &relayProofForSignature{
		RequestHash:        proof.RequestHash,
		Entropy:            proof.Entropy,
		SessionBlockHeight: proof.SessionBlockHeight,
		ServicerPubKey:     proof.ServicerPubKey,
		Blockchain:         proof.Blockchain,
		Token:              token,
		Signature:          "",
	}

	marshaledProof, err := json.Marshal(proofMap)
	if err != nil {
		return nil, err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledProof)
	if err != nil {
		return nil, err
	}

	return hasher.Sum(nil), nil
}

// HashAAT returns Viper AAT as hashed string
func HashAAT(aat *provider.ViperAAT) (string, error) {
	tokenToSend := *aat
	tokenToSend.Signature = ""

	marshaledAAT, err := json.Marshal(tokenToSend)
	if err != nil {
		return "", err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledAAT)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// HashRequest creates the request hash from its structure
func HashRequest(reqHash *RequestHash) (string, error) {
	marshaledReqHash, err := json.Marshal(reqHash)
	if err != nil {
		return "", err
	}

	hasher := sha3.New256()

	_, err = hasher.Write(marshaledReqHash)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
