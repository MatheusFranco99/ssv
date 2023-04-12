package testing

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"

	specalea "github.com/MatheusFranco99/ssv-spec-AleaBFT/alea"
	spectypes "github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"

	aleastorage "github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/storage"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/types"
	"github.com/MatheusFranco99/ssv/storage/basedb"
	"github.com/MatheusFranco99/ssv/storage/kv"
)

var (
	specModule   = "github.com/MatheusFranco99/ssv-spec"
	specTestPath = "spectest/generate/tests.json"
)

// TODO: add missing tests

// GenerateBLSKeys generates randomly nodes
func GenerateBLSKeys(oids ...spectypes.OperatorID) (map[spectypes.OperatorID]*bls.SecretKey, []*spectypes.Operator) {
	_ = bls.Init(bls.BLS12_381)

	nodes := make([]*spectypes.Operator, 0)
	sks := make(map[spectypes.OperatorID]*bls.SecretKey)

	for i, oid := range oids {
		sk := &bls.SecretKey{}
		sk.SetByCSPRNG()

		nodes = append(nodes, &spectypes.Operator{
			OperatorID: spectypes.OperatorID(i),
			PubKey:     sk.GetPublicKey().Serialize(),
		})
		sks[oid] = sk
	}

	return sks, nodes
}

// MsgGenerator represents a message generator
type MsgGenerator func(height specalea.Height) ([]spectypes.OperatorID, *specalea.Message)

// CreateMultipleStoredInstances enables to create multiple stored instances (with decided messages).
func CreateMultipleStoredInstances(
	sks map[spectypes.OperatorID]*bls.SecretKey,
	start specalea.Height,
	end specalea.Height,
	generator MsgGenerator,
) ([]*aleastorage.StoredInstance, error) {
	results := make([]*aleastorage.StoredInstance, 0)
	for i := start; i <= end; i++ {
		signers, msg := generator(i)
		if msg == nil {
			break
		}
		sm, err := MultiSignMsg(sks, signers, msg)
		if err != nil {
			return nil, err
		}
		results = append(results, &aleastorage.StoredInstance{
			State: &specalea.State{
				ID:                   sm.Message.Identifier,
				Round:                sm.Message.Round,
				Height:               sm.Message.Height,
				LastPreparedRound:    sm.Message.Round,
				LastPreparedValue:    sm.Message.Data,
				Decided:              true,
				DecidedValue:         sm.Message.Data,
				ProposeContainer:     specalea.NewMsgContainer(),
				PrepareContainer:     specalea.NewMsgContainer(),
				CommitContainer:      specalea.NewMsgContainer(),
				RoundChangeContainer: specalea.NewMsgContainer(),
			},
			DecidedMessage: sm,
		})
	}
	return results, nil
}

func signMessage(msg *specalea.Message, sk *bls.SecretKey) (*bls.Sign, error) {
	signatureDomain := spectypes.ComputeSignatureDomain(types.GetDefaultDomain(), spectypes.QBFTSignatureType)
	root, err := spectypes.ComputeSigningRoot(msg, signatureDomain)
	if err != nil {
		return nil, err
	}
	return sk.SignByte(root), nil
}

// MultiSignMsg signs a msg with multiple signers
func MultiSignMsg(sks map[spectypes.OperatorID]*bls.SecretKey, signers []spectypes.OperatorID, msg *specalea.Message) (*messages.SignedMessage, error) {
	_ = bls.Init(bls.BLS12_381)

	var operators = make([]spectypes.OperatorID, 0)
	var agg *bls.Sign
	for _, oid := range signers {
		signature, err := signMessage(msg, sks[oid])
		if err != nil {
			return nil, err
		}
		operators = append(operators, oid)
		if agg == nil {
			agg = signature
		} else {
			agg.Add(signature)
		}
	}

	return &messages.SignedMessage{
		Message:   msg,
		Signature: agg.Serialize(),
		Signers:   operators,
	}, nil
}

// SignMsg handle MultiSignMsg error and return just messages.SignedMessage
func SignMsg(t *testing.T, sks map[spectypes.OperatorID]*bls.SecretKey, signers []spectypes.OperatorID, msg *specalea.Message) *messages.SignedMessage {
	res, err := MultiSignMsg(sks, signers, msg)
	require.NoError(t, err)
	return res
}

// AggregateSign sign specalea.Message and then aggregate
func AggregateSign(t *testing.T, sks map[spectypes.OperatorID]*bls.SecretKey, signers []spectypes.OperatorID, consensusMessage *specalea.Message) *messages.SignedMessage {
	signedMsg := SignMsg(t, sks, signers, consensusMessage)
	// TODO: use SignMsg instead of AggregateSign
	// require.NoError(t, sigSignMsgnedMsg.Aggregate(signedMsg))
	return signedMsg
}

// AggregateInvalidSign sign specalea.Message and then change the signer id to mock invalid sig
func AggregateInvalidSign(t *testing.T, sks map[spectypes.OperatorID]*bls.SecretKey, consensusMessage *specalea.Message) *messages.SignedMessage {
	sigend := SignMsg(t, sks, []spectypes.OperatorID{1}, consensusMessage)
	sigend.Signers = []spectypes.OperatorID{2}
	return sigend
}

// NewInMemDb returns basedb.IDb with in-memory type
func NewInMemDb() basedb.IDb {
	db, _ := kv.New(basedb.Options{
		Type:   "badger-memory",
		Path:   "",
		Logger: zap.L(),
	})
	return db
}

// CommitDataToBytes encode commit data and handle error if exist
func CommitDataToBytes(t *testing.T, input *specalea.CommitData) []byte {
	ret, err := json.Marshal(input)
	require.NoError(t, err)
	return ret
}

func GetSpecTestJSON(path string, module string) ([]byte, error) {
	goModFile, err := getGoModFile(path)
	if err != nil {
		return nil, errors.New("could not get go.mod file")
	}

	// check if there is a replace
	var modPath, modVersion string
	var replace *modfile.Replace
	for _, r := range goModFile.Replace {
		if strings.EqualFold(specModule, r.Old.Path) {
			replace = r
			break
		}
	}

	if replace != nil {
		modPath = replace.New.Path
		modVersion = replace.New.Version
	} else {
		// get from require
		var req *modfile.Require
		for _, r := range goModFile.Require {
			if strings.EqualFold(specModule, r.Mod.Path) {
				req = r
				break
			}
		}
		if req == nil {
			return nil, errors.Errorf("could not find %s module", specModule)
		}
		modPath = req.Mod.Path
		modVersion = req.Mod.Version
	}

	// get module path
	p, err := GetModulePath(modPath, modVersion)
	if err != nil {
		return nil, errors.Wrap(err, "could not get module path")
	}

	if _, err := os.Stat(p); os.IsNotExist(err) {
		return nil, errors.Wrapf(err, "you don't have this module-%s/version-%s installed", modPath, modVersion)
	}

	return os.ReadFile(filepath.Join(filepath.Clean(p), filepath.Clean(module), filepath.Clean(specTestPath)))
}

func GetModulePath(name, version string) (string, error) {
	// first we need GOMODCACHE
	cache, ok := os.LookupEnv("GOMODCACHE")
	if !ok {
		cache = path.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	// then we need to escape path
	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", err
	}

	// version also
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", err
	}

	return path.Join(cache, escapedPath+"@"+escapedVersion), nil
}

func getGoModFile(path string) (*modfile.File, error) {
	// find project root path
	for {
		if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
			break
		}
		path = filepath.Dir(path)
		if path == "/" {
			return nil, errors.New("could not find go.mod file")
		}
	}

	// read go.mod
	buf, err := os.ReadFile(filepath.Join(filepath.Clean(path), "go.mod"))
	if err != nil {
		return nil, errors.New("could not read go.mod")
	}

	// parse go.mod
	return modfile.Parse("go.mod", buf, nil)
}
