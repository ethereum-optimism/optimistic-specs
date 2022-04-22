// xDomainMessageSender: should return the xDomainMsgSender address
// sendMessage: should be able to send a single message
// sendMessage: should be able to send the same message twice
// relayMessage: should revert if the L1 message sender is not the L1CrossDomainMessenger
// relayMessage: should send a call to the target contract
// relayMessage: the xDomainMessageSender is reset to the original value
// relayMessage: should revert if trying to send the same message twice
// relayMessage: should not make a call if the target is the L2 MessagePasser
