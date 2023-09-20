package testing

import (
	"os"
	"runtime/pprof"
	"testing"

	"github.com/MatheusFranco99/ssv/protocol/v2_alea/alea/messages"
	"github.com/stretchr/testify/require"
)

func TestAlea(t *testing.T) {

	inst := BaseInstance(1)

	inst2 := BaseInstance(2)

	inst.Start(TestingContent, TestingHeight)
	inst2.Start(TestingContent, TestingHeight)

	// VCBC from 1
	vcbc_send_message_1 := VCBCSendMessage(1, TestingHeight, 1, TestingContent, inst.State)

	vcbc_ready_message_1_1 := VCBCReadyMessage(1, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_2 := VCBCReadyMessage(2, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_3 := VCBCReadyMessage(3, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_4 := VCBCReadyMessage(4, TestingHeight, 1, 1, TestingContentHash, inst.State)
	agg_msg_1, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_1_1, vcbc_ready_message_1_2, vcbc_ready_message_1_3, vcbc_ready_message_1_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_1 := VCBCFinalMessage(1, TestingHeight, 1, agg_msg_1, TestingContentHash, inst.State)

	// VCBC from 2
	vcbc_send_message_2 := VCBCSendMessage(2, TestingHeight, 1, TestingContent, inst2.State)

	vcbc_ready_message_2_1 := VCBCReadyMessage(1, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_2 := VCBCReadyMessage(2, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_3 := VCBCReadyMessage(3, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_4 := VCBCReadyMessage(4, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	agg_msg_2, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_2_1, vcbc_ready_message_2_2, vcbc_ready_message_2_3, vcbc_ready_message_2_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_2 := VCBCFinalMessage(2, TestingHeight, 1, agg_msg_2, TestingContentHash, inst2.State)

	_, _, _, err = inst.ProcessMsg(vcbc_send_message_1)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_1)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_2)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_3)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_4)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_final_message_1)
	require.NoError(t, err)

	_, _, _, err = inst.ProcessMsg(vcbc_send_message_2)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_1)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_2)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_3)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_4)
	require.NoError(t, err)
	_, _, _, err = inst.ProcessMsg(vcbc_final_message_2)
	require.NoError(t, err)
}

func BenchmarkAlea(t *testing.B) {

	inst := BaseInstance(1)
	inst2 := BaseInstance(2)
	inst3 := BaseInstance(3)
	inst4 := BaseInstance(4)

	inst.Start(TestingContent, TestingHeight)
	inst2.Start(TestingContent, TestingHeight)
	inst3.Start(TestingContent, TestingHeight)
	inst4.Start(TestingContent, TestingHeight)

	// VCBC from 1
	vcbc_send_message_1 := VCBCSendMessage(1, TestingHeight, 1, TestingContent, inst.State)

	vcbc_ready_message_1_1 := VCBCReadyMessage(1, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_2 := VCBCReadyMessage(2, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_3 := VCBCReadyMessage(3, TestingHeight, 1, 1, TestingContentHash, inst.State)
	vcbc_ready_message_1_4 := VCBCReadyMessage(4, TestingHeight, 1, 1, TestingContentHash, inst.State)
	agg_msg_1, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_1_1, vcbc_ready_message_1_2, vcbc_ready_message_1_3, vcbc_ready_message_1_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_1 := VCBCFinalMessage(1, TestingHeight, 1, agg_msg_1, TestingContentHash, inst.State)

	// VCBC from 2
	vcbc_send_message_2 := VCBCSendMessage(2, TestingHeight, 1, TestingContent, inst2.State)

	vcbc_ready_message_2_1 := VCBCReadyMessage(1, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_2 := VCBCReadyMessage(2, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_3 := VCBCReadyMessage(3, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	vcbc_ready_message_2_4 := VCBCReadyMessage(4, TestingHeight, 1, 2, TestingContentHash, inst2.State)
	agg_msg_2, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_2_1, vcbc_ready_message_2_2, vcbc_ready_message_2_3, vcbc_ready_message_2_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_2 := VCBCFinalMessage(2, TestingHeight, 1, agg_msg_2, TestingContentHash, inst2.State)

	// VCBC from 3
	vcbc_send_message_3 := VCBCSendMessage(3, TestingHeight, 1, TestingContent, inst3.State)

	vcbc_ready_message_3_1 := VCBCReadyMessage(1, TestingHeight, 1, 3, TestingContentHash, inst3.State)
	vcbc_ready_message_3_2 := VCBCReadyMessage(2, TestingHeight, 1, 3, TestingContentHash, inst3.State)
	vcbc_ready_message_3_3 := VCBCReadyMessage(3, TestingHeight, 1, 3, TestingContentHash, inst3.State)
	vcbc_ready_message_3_4 := VCBCReadyMessage(4, TestingHeight, 1, 3, TestingContentHash, inst3.State)
	agg_msg_3, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_3_1, vcbc_ready_message_3_2, vcbc_ready_message_3_3, vcbc_ready_message_3_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_3 := VCBCFinalMessage(3, TestingHeight, 1, agg_msg_3, TestingContentHash, inst3.State)

	// VCBC from 4
	vcbc_send_message_4 := VCBCSendMessage(4, TestingHeight, 1, TestingContent, inst4.State)

	vcbc_ready_message_4_1 := VCBCReadyMessage(1, TestingHeight, 1, 4, TestingContentHash, inst4.State)
	vcbc_ready_message_4_2 := VCBCReadyMessage(2, TestingHeight, 1, 4, TestingContentHash, inst4.State)
	vcbc_ready_message_4_3 := VCBCReadyMessage(3, TestingHeight, 1, 4, TestingContentHash, inst4.State)
	vcbc_ready_message_4_4 := VCBCReadyMessage(4, TestingHeight, 1, 4, TestingContentHash, inst4.State)
	agg_msg_4, err := AggregateMsgs([]*messages.SignedMessage{vcbc_ready_message_4_1, vcbc_ready_message_4_2, vcbc_ready_message_4_3, vcbc_ready_message_4_4})
	if err != nil {
		panic(err)
	}
	vcbc_final_message_4 := VCBCFinalMessage(4, TestingHeight, 1, agg_msg_4, TestingContentHash, inst4.State)

	cpuProfileFile, err := os.Create("profile.out")
	if err != nil {
		panic(err)
	}

	err = pprof.StartCPUProfile(cpuProfileFile)
	if err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		_, _, _, err := inst.ProcessMsg(vcbc_send_message_1)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_1)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_2)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_3)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_1_4)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_final_message_1)
		require.NoError(t, err)

		_, _, _, err = inst.ProcessMsg(vcbc_send_message_2)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_1)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_2)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_3)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_2_4)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_final_message_2)
		require.NoError(t, err)

		_, _, _, err = inst.ProcessMsg(vcbc_send_message_3)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_3_1)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_3_2)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_3_3)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_3_4)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_final_message_3)
		require.NoError(t, err)

		_, _, _, err = inst.ProcessMsg(vcbc_send_message_4)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_4_1)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_4_2)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_4_3)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_ready_message_4_4)
		require.NoError(t, err)
		_, _, _, err = inst.ProcessMsg(vcbc_final_message_4)
		require.NoError(t, err)

		require.True(t, inst.State.Decided)
	}
	t.StopTimer()
}
