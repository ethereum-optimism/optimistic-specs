// pause: should pause the contract when called by the current owner
// pause: should not pause the contract when called by account other than the owner

// sendMessage: should be able to send a single message
// sendMessage: should be able to send the same message twice

// replayMessage: when giving some incorrect input value: should revert if given the wrong target
// replayMessage: when giving some incorrect input value: should revert if given the wrong sender
// replayMessage: when giving some incorrect input value: should revert if given the wrong message
// replayMessage: when giving some incorrect input value: should revert if given the wrong queue index
// replayMessage: when giving some incorrect input value: should revert if given the wrong old gas limit

// replayMessage: when all input values are the same as the existing message: should succeed
// replayMessage: when all input values are the same as the existing message: should emit the TransactionEnqueued event
// replayMessage: when all input values are the same as the existing message: should succeed if all inputs are the same as the existing message

// xDomainMessageSender: should return the xDomainMsgSender address


// relayMessage: should revert if still inside the fraud proof window
// relayMessage: should revert if attempting to relay a message sent to an L1 system contract
// relayMessage: should revert if provided an invalid output root proof
// relayMessage: should revert if provided an invalid storage trie witness
// relayMessage: should send a successful call to the target contract
// relayMessage: the xDomainMessageSender is reset to the original value
// relayMessage: should revert if trying to send the same message twice
// relayMessage: should revert if paused

// blockMessage and allowMessage: should revert if called by an account other than the owner
// blockMessage and allowMessage: should revert if the message is blocked
// blockMessage and allowMessage: should succeed if the message is blocked, then unblocked
