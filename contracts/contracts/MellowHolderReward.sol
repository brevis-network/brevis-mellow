// SPDX-License-Identifier: MIT
pragma solidity ^0.8.18;

import "@openzeppelin/contracts/access/Ownable.sol";

import "./lib/BrevisAppZkOnly.sol";

// Only accept ZK-attested results.
contract MellowHolderReward is BrevisAppZkOnly, Ownable {
    event RewardAttested(uint64 startBlockNum, uint64 endBlockNum, address account, uint248 reward);

    mapping (bytes32 => bool) vkHashes;

    constructor(address _brevisRequest) BrevisAppZkOnly(_brevisRequest) Ownable(msg.sender) {}

    // BrevisQuery contract will call our callback once Brevis backend submits the proof.
    // This method is called with once the proof is verified.
    function handleProofResult(bytes32 _vkHash, bytes calldata _circuitOutput) internal override {
        // We need to check if the verifying key that Brevis used to verify the proof
        // generated by our circuit is indeed our designated verifying key. This proves
        // that the _circuitOutput is authentic
        require(vkHashes[_vkHash], "invalid vk");
        (uint64 startBlockNum, uint64 endBlockNum, address[] memory accounts, uint248[] memory rewards) = decodeOutput(_circuitOutput);

        for (uint256 i = 0; i < accounts.length; i++) {
            emit RewardAttested(startBlockNum, endBlockNum, accounts[i], rewards[i]);
        }
    }


    function decodeOutput(bytes calldata o) internal pure returns (uint64 startBlockNum, uint64 endBlockNum, address[] memory accounts, uint248[] memory rewards) {
        require(o.length > 16, "invalid output");
        startBlockNum = uint64(bytes8(o[0:8]));
        endBlockNum = uint64(bytes8(o[8:16]));
        require((o.length - 16) % 51 == 0, "invalid account output");
        uint256 accountsLength = (o.length - 16) / 51;
        accounts = new address[](accountsLength);
        rewards = new uint248[](accountsLength);
        for (uint256 i = 0; i < accountsLength; i++) {
            accounts[i] = address(bytes20(o[16+51*i:36+51*i]));
            rewards[i] = uint248(bytes31(o[36+51*i:67+51*i]));
        }
    }

    function setVkHash(bytes32 _vkHash) external onlyOwner {
        vkHashes[_vkHash] = true;
    }

    function deprecateVkHash(bytes32 _vkHash) external onlyOwner {
        vkHashes[_vkHash] = false;
    }

    function mockDecode(bytes calldata o) external pure returns (uint64 startBlockNum, uint64 endBlockNum, address[] memory accounts, uint248[] memory rewards) {
        require(o.length > 16, "invalid output");
        startBlockNum = uint64(bytes8(o[0:8]));
        endBlockNum = uint64(bytes8(o[8:16]));
        require((o.length - 16) % 51 == 0, "invalid account output");
        uint256 accountsLength = (o.length - 16) / 51;
        accounts = new address[](accountsLength);
        rewards = new uint248[](accountsLength);
        for (uint256 i = 0; i < accountsLength; i++) {
            accounts[i] = address(bytes20(o[16+51*i:36+51*i]));
            if (accounts[i] == address(0)) {
                break;
            }
            rewards[i] = uint248(bytes31(o[36+51*i:67+51*i]));
        }
    }
}
