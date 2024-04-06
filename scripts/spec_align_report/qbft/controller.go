package qbft

import (
	"github.com/MatheusFranco99/ssv/scripts/spec_align_report/utils"
)

func ProcessController() {
	if err := utils.Mkdir(utils.DataPath+"/controller", true); err != nil {
		panic(err)
	}

	controllerCompareStruct := initControllerCompareStruct()
	if err := controllerCompareStruct.ReplaceMap(); err != nil {
		panic(err)
	}
	_ = controllerCompareStruct.Run()

	decidedCompareStruct := initDecidedCompareStruct()
	if err := decidedCompareStruct.ReplaceMap(); err != nil {
		panic(err)
	}
	_ = decidedCompareStruct.Run()

	futureMsgCompareStruct := initFutureMsgCompareStruct()
	if err := futureMsgCompareStruct.ReplaceMap(); err != nil {
		panic(err)
	}
	_ = futureMsgCompareStruct.Run()

}
func initControllerCompareStruct() *utils.Compare {
	c := &utils.Compare{
		Name:        "controller",
		Replace:     ControllerSet(),
		SpecReplace: SpecControllerSet(),
		SSVPath:     utils.DataPath + "/controller/controller.go",
		SpecPath:    utils.DataPath + "/controller/controller_spec.go",
	}
	if err := utils.Copy("./protocol/v2_alea/alea/controller/controller.go", c.SSVPath); err != nil {
		panic(err)
	}
	if err := utils.Copy("./scripts/spec_align_report/ssv-spec/qbft/controller.go", c.SpecPath); err != nil {
		panic(err)
	}
	return c
}
func initDecidedCompareStruct() *utils.Compare {
	c := &utils.Compare{
		Name:        "decided",
		Replace:     DecidedSet(),
		SpecReplace: SpecDecidedSet(),
		SSVPath:     utils.DataPath + "/controller/decided.go",
		SpecPath:    utils.DataPath + "/controller/decided_spec.go",
	}
	if err := utils.Copy("./protocol/v2_alea/alea/controller/decided.go", c.SSVPath); err != nil {
		panic(err)
	}
	if err := utils.Copy("./scripts/spec_align_report/ssv-spec/qbft/decided.go", c.SpecPath); err != nil {
		panic(err)
	}
	return c
}
func initFutureMsgCompareStruct() *utils.Compare {
	c := &utils.Compare{
		Name:        "future_msg",
		Replace:     FutureMessageSet(),
		SpecReplace: SpecFutureMessageSet(),
		SSVPath:     utils.DataPath + "/controller/future_msg.go",
		SpecPath:    utils.DataPath + "/controller/future_msg_spec.go",
	}
	if err := utils.Copy("./protocol/v2_alea/alea/controller/future_msg.go", c.SSVPath); err != nil {
		panic(err)
	}
	if err := utils.Copy("./scripts/spec_align_report/ssv-spec/qbft/future_msg.go", c.SpecPath); err != nil {
		panic(err)
	}
	return c
}
