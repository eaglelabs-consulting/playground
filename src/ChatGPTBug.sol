// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import {Context} from "suave-std/Context.sol";
import {Suapp} from "suave-std/Suapp.sol";
import {ChatGPT} from "suave-std/protocols/ChatGPT.sol";
import {Suave} from "suave-std/suavelib/Suave.sol";

contract ChatGPTBug is Suapp {
    string internal constant API_KEY_NAMESPACE = "api_key:v0:secret";

    Suave.DataId apiKeyRecord;

    event Response(string s);

    function updateKeyOnchain(Suave.DataId _apiKeyRecord) public {
        apiKeyRecord = _apiKeyRecord;
    }

    function registerKeyOffchain() public returns (bytes memory) {
        bytes memory keyData = Context.confidentialInputs();

        address[] memory peekers = new address[](1);
        peekers[0] = address(this);

        Suave.DataRecord memory record = Suave.newDataRecord(0, peekers, peekers, API_KEY_NAMESPACE);
        Suave.confidentialStore(record.id, API_KEY_NAMESPACE, keyData);

        return abi.encodeWithSelector(this.updateKeyOnchain.selector, record.id);
    }

    function offchain() external returns (bytes memory) {
        bytes memory keyData = Suave.confidentialRetrieve(apiKeyRecord, API_KEY_NAMESPACE);
        string memory apiKey = bytesToString(keyData);
        ChatGPT chatgpt = new ChatGPT(apiKey);

        ChatGPT.Message[] memory messages = new ChatGPT.Message[](1);
        messages[0] = ChatGPT.Message(ChatGPT.Role.User, "Say hello world");

        string memory data = chatgpt.complete(messages);

        emit Response(data);

        return abi.encodeWithSelector(this.onchain.selector);
    }

    function onchain() external payable emitOffchainLogs {}

    function bytesToString(bytes memory data) private pure returns (string memory) {
        uint256 length = data.length;
        bytes memory chars = new bytes(length);

        for (uint256 i = 0; i < length; i++) {
            chars[i] = data[i];
        }

        return string(chars);
    }
}
