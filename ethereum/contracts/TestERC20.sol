// SPDX-License-Identifier: MIT

pragma solidity ^0.8.0;

import {ERC20} from "@openzeppelin/contracts/token/ERC20/ERC20.sol";

// ERC20 token for testing purposes
contract TestERC20 is ERC20 {
    uint8 private immutable _decimals;

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

    function approve(address spender, uint256 amount) public virtual override returns (bool) {
        address owner = _msgSender();

        // This next checks is performed by the USDT contract, that we want to
        // be compatible with:
        // https://etherscan.io/token/0xdac17f958d2ee523a2206206994597c13d831ec7#code
        //
        // To change the approve amount you first have to reduce the addresses
        // allowance to zero to prevent an attack described here:
        // https://docs.google.com/document/d/1YLPtQxZu1UAvO9cZ1O2RPXBbT0mooh4DYKjA_jp-RLM/edit
        require(
            amount == 0 || allowance(owner, spender) == 0,
            "approve allowance must be set to zero before updating"
        );

        _approve(owner, spender, amount);
        return true;
    }

    function transferInternal(address from, address to, uint256 value) public {
        _transfer(from, to, value);
    }

    function approveInternal(address owner, address spender, uint256 value) public {
        _approve(owner, spender, value);
    }

    // You can send a zero-value transfer directly to the contract address to
    // get a 100 standard unit tokens.
    receive() external payable {
        mint(msg.sender, 100 * 10 ** uint(_decimals));
        if (msg.value > 0) {
            payable(msg.sender).transfer(msg.value);
        }
    }
}
