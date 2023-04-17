// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0 .0;

import {ERC20} from "./ERC20.sol";

// ERC20 token for testing purposes
contract TestERC20 is ERC20 {
    uint8 private _decimals;

    constructor(
        string memory name,
        string memory symbol,
        uint8 numDecimals,
        address initialAccount,
        uint256 initialBalance
    ) payable ERC20(name, symbol) {
        _decimals = numDecimals;
        _mint(initialAccount, initialBalance);
    }

    function decimals() public view virtual override returns (uint8) {
        return _decimals;
    }

    function mint(address account, uint256 amount) public {
        _mint(account, amount);
    }

    function burn(address account, uint256 amount) public {
        _burn(account, amount);
    }

    function transferInternal(address from, address to, uint256 value) public {
        _transfer(from, to, value);
    }

    function approveInternal(address owner, address spender, uint256 value) public {
        _approve(owner, spender, value);
    }
}
