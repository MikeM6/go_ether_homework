pragma solidity ^0.8.24;

contract Counter {
    uint256 public value;

    function increment(uint256 by) external {
        value += by;
    }
}
