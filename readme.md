# Sample Hardhat Project

## 
## 合约和 合约测试脚本
存放在： https://github.com/Marco-006/hardhat-test-v2.git

## 
### 什么是 区块链重组回滚
区块链出现“重组回滚”（Reorganization，简称 Reorg）是指：
当网络中同时出现两条（或以上）都合法的临时分叉链时，共识机制最终会选出一条“更优”的链作为唯一有效主链，并放弃（回滚）另一条链上已经确认的区块。被丢弃链上的所有交易会被退回到未确认状态，仿佛从未写入过账本，这就是所谓“回滚”


This project demonstrates a basic Hardhat use case. It comes with a sample contract, a test for that contract, and a Hardhat Ignition module that deploys that contract.

Try running some of the following tasks:

```shell
npx hardhat help
npx hardhat test
REPORT_GAS=true npx hardhat test
npx hardhat node
npx hardhat ignition deploy ./ignition/modules/Lock.js
```

