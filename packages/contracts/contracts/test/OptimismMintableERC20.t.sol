// SPDX-License-Identifier: MIT
pragma solidity 0.8.10;

import { Bridge_Initializer } from "./CommonTest.t.sol";
import { LibRLP } from "./Lib_RLP.t.sol";
import "../universal/SupportedInterfaces.sol";
import { IERC165 } from "@openzeppelin/contracts/utils/introspection/IERC165.sol";

contract OptimismMintableTokenFactory_Test is Bridge_Initializer {
    event Mint(address indexed _account, uint256 _amount);
    event Burn(address indexed _account, uint256 _amount);

    function setUp() public override {
        super.setUp();
    }

    function test_remoteToken() external {
        assertEq(L2Token.remoteToken(), address(L1Token));
    }

    function test_bridge() external {
        assertEq(L2Token.bridge(), address(L2Bridge));
    }

    function test_l1Token() external {
        assertEq(L2Token.l1Token(), address(L1Token));
    }

    function test_l2Bridge() external {
        assertEq(L2Token.l2Bridge(), address(L2Bridge));
    }

    function test_mint() external {
        vm.expectEmit(true, true, true, true);
        emit Mint(alice, 100);

        vm.prank(address(L2Bridge));
        L2Token.mint(alice, 100);

        assertEq(L2Token.balanceOf(alice), 100);
    }

    function test_mintRevertsFromNotBridge() external {
        // NOT the bridge
        vm.expectRevert("Only L2 Bridge can mint and burn");
        vm.prank(address(alice));
        L2Token.mint(alice, 100);
    }

    function test_burn() external {
        vm.prank(address(L2Bridge));
        L2Token.mint(alice, 100);

        vm.expectEmit(true, true, true, true);
        emit Burn(alice, 100);

        vm.prank(address(L2Bridge));
        L2Token.burn(alice, 100);

        assertEq(L2Token.balanceOf(alice), 0);
    }

    function test_burnRevertsFromNotBridge() external {
        // NOT the bridge
        vm.expectRevert("Only L2 Bridge can mint and burn");
        vm.prank(address(alice));
        L2Token.burn(alice, 100);
    }

    function test_erc165_supportsInterface() external {
        // The assertEq calls in this test are comparing the manual calculation of the iface,
        // with what is returned by the solidity's type().interfaceId, just to be safe.
        bytes4 iface1 = bytes4(keccak256("supportsInterface(bytes4)"));
        assertEq(iface1, type(IERC165).interfaceId);
        assert(L2Token.supportsInterface(iface1));
        emit log_bytes32(bytes32(iface1));

        bytes4 iface2 = L2Token.l1Token.selector ^ L2Token.mint.selector ^ L2Token.burn.selector;
        assertEq(iface2, type(IL1Token).interfaceId);
        assert(L2Token.supportsInterface(iface2));
        emit log_bytes32(bytes32(iface2));

        bytes4 iface3 = L2Token.remoteToken.selector ^ L2Token.mint.selector ^ L2Token.burn.selector;
        assertEq(iface3, type(IRemoteToken).interfaceId);
        assert(L2Token.supportsInterface(iface3));
        emit log_bytes32(bytes32(iface3));

    }
}
