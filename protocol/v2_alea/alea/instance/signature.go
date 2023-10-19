package instance

import (
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"

	"github.com/MatheusFranco99/ssv-spec-AleaBFT/types"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea"
	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/pkg/errors"
)

var (
	data = []byte{1, 2, 3, 4}

	Ed25519PrivateKey = "965779ca218caa45325d369315d363d4a6dd43bb1367e41846c23c6cdf469ebe47755444af11ec67776315f8bffce683b2b3afdc49c4f44c88843c73863c8572"
	Ed25519PublicKey  = "3518bc897097134f5d764896d12c3541895e116bf4f41f3ef817972c7f5f79e1"
	Ed25519Signature  = "5bb83ab6b85bed3e1e4d1d0e32453559c643f5e291be0829c83faa4326b6e2f9830673c1cae85dbfcfbeb6eeffbcd67ac02ac796eae68b55d833095f8d2a2502"

	encodedSig    = "3241a7e1bbb911463000e487602051b5c839ddb486c594ef974eb4f8c0c6b750ff04a34aa4d38ada7535c0d24596efe5e42339b07c7d76468742e9695604f24a"
	rawSig, _     = hex.DecodeString(encodedSig)
	digest        = "9f64a747e1b97f131fabb6b447296c9b6f0201e79fb3c5356e6c77e89b6a806a"
	rawDigest, _  = hex.DecodeString(digest)
	rsaPrivateKey = parseKey(`-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBALKZD0nEffqM1ACuak0bijtqE2QrI/KLADv7l3kK3ppMyCuLKoF0
fd7Ai2KW5ToIwzFofvJcS/STa6HA5gQenRUCAwEAAQJBAIq9amn00aS0h/CrjXqu
/ThglAXJmZhOMPVn4eiu7/ROixi9sex436MaVeMqSNf7Ex9a8fRNfWss7Sqd9eWu
RTUCIQDasvGASLqmjeffBNLTXV2A5g4t+kLVCpsEIZAycV5GswIhANEPLmax0ME/
EO+ZJ79TJKN5yiGBRsv5yvx5UiHxajEXAiAhAol5N4EUyq6I9w1rYdhPMGpLfk7A
IU2snfRJ6Nq2CQIgFrPsWRCkV+gOYcajD17rEqmuLrdIRexpg8N1DOSXoJ8CIGlS
tAboUGBxTDq3ZroNism3DaMIbKPyYrAqhKov1h5V
-----END RSA PRIVATE KEY-----`)

	RSAPublicKey  = ""
	RSAPrivateKey = ""
	RSASignature  = "8351441e84e57ed96080df8d19e5755901879a1672e791d63b90de872e609f2e7cf578813dda3a2fcf27474ee978ac5a8df5c29d65b69fb9ac244cf543627646500da0a4f2d825ca61fadf14dc7407e13cbebafc125a05ac5a321ab4210bff11063a846c1a587a981a43340e3599a2a51cf9c895854175e4b3b84a23c83d2b03a3831196b679669399c971afd1296a8bfc9e29bae66013f1bdb37872e504233beeaf3f068649a29c93b3e47a790e07736950ff34b674973e23ec965ba48d54bb904b41100665e9f40baad34960fb17f18de4115b0d0de59fe3830fb361af59c2e01b1ba886e58e3082cdc510a50a51da4ad6f7cbb25d5b36af3e166220a603ff"
)

func Sign(state *messages.State, config alea.IConfig, msg *messages.Message) ([]byte, map[types.OperatorID][32]byte, error) {
	sig := []byte{}
	hash_map := make(map[types.OperatorID][32]byte)
	var err error
	if state.UseBLS {
		sig, err = config.GetSigner().SignRoot(msg, types.QBFTSignatureType, state.Share.SharePubKey)
		if err != nil {
			return sig, hash_map, errors.Wrap(err, "CreateVCBCSend: failed signing filler msg")
		}
	} else if state.UseDiffieHellman {
		msg_bytes, err := msg.Encode()
		if err != nil {
			return sig, hash_map, errors.Wrap(err, "CreateVCBCSend: failed to encode message")
		}
		hash_map = state.DiffieHellmanContainerOneTimeCost.GetHashMap(msg_bytes)
	} else if state.UseEDDSA {
		// sig = MockEDDSASignature()
		sig, _ = MockEDDSASignature()
	} else if state.UseRSA {
		sig, _ = MockRSASigning()
	}
	return sig, hash_map, nil
}

func Verify(state *messages.State, config alea.IConfig, signedMsg *messages.SignedMessage, operators []*types.Operator) error {

	if state.UseBLS {
		if err := signedMsg.Signature.VerifyByOperators(signedMsg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
			return errors.Wrap(err, "msg signature invalid")
		}
	} else if state.UseDiffieHellman {
		msg_bytes, err := signedMsg.Message.Encode()
		if err != nil {
			return errors.Wrap(err, "Could not encode message")
		}
		if !state.DiffieHellmanContainerOneTimeCost.VerifyHash(msg_bytes, signedMsg.GetSigners()[0], signedMsg.DiffieHellmanProof[state.Share.OperatorID]) {
			return errors.New("Failed Diffie Hellman verification")
		}
	} else if state.UseEDDSA {
		if !MockEDDSAVerification() {
			return errors.New("Could not verify EDDSA signature")
		}
	} else if state.UseRSA {
		return MockRSAVerification()
	}
	return nil
}

func VerifyVCBCFinal(state *messages.State, config alea.IConfig, signedMsg *messages.SignedMessage, operators []*types.Operator) error {


	vcbcFinalData, err := signedMsg.Message.GetVCBCFinalData()
	if err != nil {
		return errors.Wrap(err, "VerifyVCBCFinal: could not get vcbcFinalData data from signedMessage")
	}

	// get sender ID
	aggregated_msg := vcbcFinalData.AggregatedMessage

	if err := aggregated_msg.Signature.VerifyByOperators(aggregated_msg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return errors.Wrap(err, "msg signature invalid")
	}
	return nil
}

func VerifyBLSAggregate(state *messages.State, config alea.IConfig, signedMsgs []*messages.SignedMessage, operators []*types.Operator) error {
	// aggregated_msg, err := aggregateMsgs(signedMsgs)
	// if err != nil {
	// 	return err
	// }

	// verify signature
	if err := signedMsgs[0].Signature.VerifyByOperators(signedMsgs[0], config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return nil
		return errors.Wrap(err, "aggregated msg signature invalid")
	}
	return nil
}


func VerifyBLSAggregateFinals(state *messages.State, config alea.IConfig, signedMsgs []*messages.SignedMessage, operators []*types.Operator) error {
	aggregate_msgs := make([]*messages.SignedMessage, len(signedMsgs))
	for i, _ := range signedMsgs {

		vcbcFinalData, err := signedMsgs[i].Message.GetVCBCFinalData()
		if err != nil {
			return errors.Wrap(err, "VerifyBLSAggregateFinals: could not get vcbcFinalData data from signedMessage")
		}

		// get sender ID
		aggregated_msg_i := vcbcFinalData.AggregatedMessage
		
		aggregate_msgs[i] = aggregated_msg_i
	}
	aggregated_msg, err := aggregateMsgs(aggregate_msgs)
	if err != nil {
		return err
	}

	// verify signature
	if err := aggregated_msg.Signature.VerifyByOperators(aggregated_msg, config.GetSignatureDomainType(), types.QBFTSignatureType, operators); err != nil {
		return nil
		return errors.Wrap(err, "aggregated msg signature invalid")
	}
	return nil
}

func MockEDDSASignature() ([]byte, error) {
	// Parse the private key
	privateKeyBytes, err := hex.DecodeString(Ed25519PrivateKey)
	if err != nil {
		return nil, err
	}
	privateKey := ed25519.PrivateKey(privateKeyBytes)

	// Calculate the signature
	signature := ed25519.Sign(privateKey, data)

	return signature, nil
}

func MockEDDSAVerification() bool {

	publicKeyBytes, err := hex.DecodeString(Ed25519PublicKey)
	if err != nil {
		return false
	}
	signatureBytes, err := hex.DecodeString(Ed25519Signature)
	if err != nil {
		return false
	}
	publicKey := ed25519.PublicKey(publicKeyBytes)

	// Verify the signature
	return ed25519.Verify(publicKey, data, signatureBytes)

}

func MockRSASigning() ([]byte, error) {
	sign, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, rawDigest)
	return sign, err
}

func MockRSAVerification() error {
	return rsa.VerifyPKCS1v15(&rsaPrivateKey.PublicKey, crypto.SHA256, rawDigest, rawSig)
}

func parseKey(s string) *rsa.PrivateKey {
	p, _ := pem.Decode([]byte(s))
	k, err := x509.ParsePKCS1PrivateKey(p.Bytes)
	if err != nil {
		panic(err)
	}
	return k
}
