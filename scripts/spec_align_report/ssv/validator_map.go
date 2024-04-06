package ssv

import "github.com/MatheusFranco99/ssv/scripts/spec_align_report/utils"

func validatorSet() []utils.KeyValue {
	var mapSet = utils.NewMap()
	mapSet.Set("package validator", "package ssv")
	mapSet.Set("\"context\"\n", "")
	mapSet.Set("\"encoding/hex\"\n", "")
	mapSet.Set("\"github.com/MatheusFranco99/ssv/protocol/v2_alea/message\"\n", "")
	mapSet.Set("specssv \"github.com/MatheusFranco99/ssv-spec-AleaBFT/ssv\"\n", "")
	mapSet.Set("spectypes \"github.com/MatheusFranco99/ssv-spec-AleaBFT/types\"", "\"github.com/MatheusFranco99/ssv-spec-AleaBFT/types\"")
	mapSet.Set("logging \"github.com/ipfs/go-log\"\n", "")
	mapSet.Set("specalea \"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea\"", "\"github.com/MatheusFranco99/ssv-spec-AleaBFT/alea\"")
	mapSet.Set("\"go.uber.org/zap\"\n\n", "")
	mapSet.Set("\"github.com/MatheusFranco99/ssv/ibft/storage\"\n", "")
	mapSet.Set("\"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/msgqueue\"\n", "")
	mapSet.Set("\"github.com/MatheusFranco99/ssv/protocol/v2_alea/ssv/runner\"\n", "")
	mapSet.Set("\"github.com/MatheusFranco99/ssv/protocol/v2_alea/types\"\n", "")

	mapSet.Set("ctx    context.Context\n", "")
	mapSet.Set("cancel context.CancelFunc\n", "")
	mapSet.Set("logger *zap.Logger\n", "")
	mapSet.Set("var logger = logging.Logger(\"ssv/protocol/ssv/validator\").Desugar()\n", "")
	mapSet.Set("specalea.Network", "Network")
	mapSet.Set("specssv.", "")
	mapSet.Set("specalea.", "qbft.")
	mapSet.Set("spectypes.", "types.")
	mapSet.Set("runner.", "")
	mapSet.Set("*types.SSVShare", "*types.Share")

	mapSet.Set("Storage *storage.ALEAStores\n", "")
	mapSet.Set("Queues  map[types.BeaconRole]msgqueue.MsgQueue\n", "")
	mapSet.Set("state uint32\n", "")

	// not aligned to spec due to use of options and queue
	mapSet.Set("func NewValidator(pctx context.Context, options Options) *Validator {\n\toptions.defaults()\n\tctx, cancel := context.WithCancel(pctx)\n\n\tlogger = logger.With(zap.String(\"validator\", hex.EncodeToString(options.SSVShare.ValidatorPubKey)))\n\n\tv := &Validator{\n\t\tctx:         ctx,\n\t\tcancel:      cancel,\n\t\tlogger:      logger,\n\t\tDutyRunners: options.DutyRunners,\n\t\tNetwork:     options.Network,\n\t\tBeacon:      options.Beacon,\n\t\tStorage:     options.Storage,\n\t\tShare:       options.SSVShare,\n\t\tSigner:      options.Signer,\n\t\tQueues:      make(map[types.BeaconRole]msgqueue.MsgQueue), // populate below\n\t\tstate:       uint32(NotStarted),\n\t}\n\n\tindexers := msgqueue.WithIndexers(msgqueue.SignedMsgIndexer(), msgqueue.DecidedMsgIndexer(), msgqueue.SignedPostConsensusMsgIndexer(), msgqueue.EventMsgMsgIndexer())\n\tfor _, dutyRunner := range options.DutyRunners {\n\t\t// set timeout F\n\t\tdutyRunner.GetBaseRunner().TimeoutF = v.onTimeout\n\n\t\tq, _ := msgqueue.New(logger, indexers) // TODO: handle error\n\t\tv.Queues[dutyRunner.GetBaseRunner().BeaconRoleType] = q\n\t}\n\n\treturn v\n}",
		"func NewValidator(\n\tnetwork Network,\n\tbeacon BeaconNode,\n\tshare *types.Share,\n\tsigner types.KeyManager,\n\trunners map[types.BeaconRole]Runner,\n) *Validator {\n\treturn &Validator{\n\t\tDutyRunners: runners,\n\t\tNetwork:     network,\n\t\tBeacon:      beacon,\n\t\tShare:       share,\n\t\tSigner:      signer,\n\t}\n}")

	// We use share as we don't have runners in non committee validator
	mapSet.Set("validateMessage(v.Share.Share,", "v.validateMessage(dutyRunner,")
	mapSet.Set("func validateMessage(share types.Share,", "func (v *Validator) validateMessage(runner Runner,")
	mapSet.Set("!share.ValidatorPubKey", "!v.Share.ValidatorPubKey")

	mapSet.Set("case message.SSVEventMsgType:\n\t\treturn v.handleEventMessage(msg, dutyRunner)\n", "")

	return mapSet.Range()
}
func specValidatorSet() []utils.KeyValue {
	var mapSet = utils.NewMap()
	// TODO add comment to spec
	mapSet.Set("func NewValidator(", "// NewValidator creates a new instance of Validator.\nfunc NewValidator(")
	return mapSet.Range()
}
