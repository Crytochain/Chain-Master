# LBR JavaScript API version 0.1.22.

This is the [LBR](https://github.com/LBRChain/LBR-core) compatible JavaScript API which implements the Generic JSON RPC spec as described in the Chain3.md. It's available on npm as a node module, for bower and component as an embeddable js and as a meteor.js package.

chain3 is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

chain3 is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
LICENSE file for more details.

## Updates

* v0.1.22 Fixed a test in contracTest due to bignumber.js lib, update some dependence pull requests.
* v0.1.21 Fixed the previous bug with lodash.
* v0.1.20 Added admin and txpool RPC methods for VNODE, fixed signature error of zero to support microchain DAPP call functions.
* v0.1.19 Added microchain functions to support microchain DAPP call functions.
* v0.1.18 Adopt the eth-lib to perform the message signing and verifying to avoid signature verifying error.
* v0.1.17 Added RLP encode for non-Hex string before signing and verifying to avoid signature verifying error.
* v0.1.16 Added debug APIs and fixed the unit test error.
* v0.1.15 Added local LBR sign and verify signature methods.
* v0.1.14 Fixed document errors.
* v0.1.13 Added scs_getBlockList method.
* v0.1.12 Fixed Readme and missing scs_getMicroChainInfo method.
* v0.1.11 Added scs method getMicroChainInfo to work with subchain explorer.
* v0.1.10 Added vnode and scs methods to work with MicroChains.
* v0.1.9 Added new method to get ip for local node.
* v0.1.8 Moved git repository to https://github.com/LBRChain/chain3 and fixed some bugs.
* v0.1.7 Fixed uneven signature R & S.
* v0.1.6 A complete package with all tests.
* v0.1.4 First compatible version with LBR chain.

Some of the methods require running a local LBR node to use this library.
To use vnode methods, need to enable the vnode APIs in LBR node by:
--rpc --rpcport "8545" --rpcapi "chain3,mc,net,vnode"

To use scs methods, need to enable the SCS RPC ports by:
--rpc

More information is in Chain3.md or [LBR wiki](https://github.com/LBRChain/LBR-core/wiki/Chain3).


## Installation

### Node.js

```bash
npm install chain3
```

### As Browser module
Bower

```bash
bower install chain3
```
### Meteor.js

```bash
meteor add LBRlib:chain3
```


* Include `chain3.min.js` in your html file. (not required for the meteor package)

## Usage
Use the `chain3` object directly from global namespace:

```js
var Chain3 = require('chain3');
var chain3 = new Chain3();
console.log(chain3); // {mc: .., net: ...} // it's here!
```

Set a provider (HttpProvider) with VNODE:

```js
if (typeof chain3 !== 'undefined') {
  chain3 = new Chain3(chain3.currentProvider);
} else {
  // set the provider you want from Chain3.providers
  chain3 = new Chain3(new Chain3.providers.HttpProvider("http://localhost:8545"));
}
```

Set a provider (HttpProvider using [HTTP Basic Authentication](https://en.wikipedia.org/wiki/Basic_access_authentication))

```js
chain3.setProvider(new chain3.providers.HttpProvider('http://host.url', 0, BasicAuthUsername, BasicAuthPassword));
```

There you go, now you can use it:

```js
var coinbase = chain3.mc.coinbase;
var balance = chain3.mc.getBalance(coinbase);
```

To work with SCS servers, need to enable the rpc port on SCS monitors
and set the scsProvider for chain3. 

```js
if (typeof chain3 !== 'undefined') {
  chain3 = new Chain3(chain3.currentProvider);
} else {
  // set the provider you want from Chain3.providers
  chain3 = new Chain3(new Chain3.providers.HttpProvider("http://localhost:8545"));
  chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
}
```

After set the SCS monitor, now you can use it:

```js
mclist=chain3.scs.getMicroChainList();
console.log("SCS MicroChain List:", mclist);

console.log("MicroChain state:", chain3.scs.getDappState(mclist[0]));
console.log("MicroChain blockNumber:", chain3.scs.getBlockNumber(mclist[0]));
```

More examples are under the example directory

## Contribute!

### Requirements

* Node.js
* npm

```bash
sudo apt-get update
sudo apt-get install nodejs
sudo apt-get install npm
sudo apt-get install nodejs-legacy
```

### Building (gulp)
Require install gulp (https://gulpjs.com/) in the system:

```bash
npm run-script build
```


### Testing (mocha)
Test all cases.
May need to install package mocha first.

```bash
mocha
```

Test a singe function.

```bash
mocha test/chain3.mc.coinbase.js 
```

## Some examples

### send_mc

Example codes to send LBR through signed transaction.
```js
var rawTx = {
      from: src.addr,
      nonce: chain3.intToHex(txcount),
      gasPrice: chain3.intToHex(2000000000),
      gasLimit: chain3.intToHex(2000),
      to: '0xf1f5b7a35dff6400af7ab3ea54e4e637059ef909',
      value: chain3.intToHex(chain3.toSha(value, 'mc')), 
      data: '0x00',
      chainId: chainid
    }

var cmd1 = chain3.signTransaction(rawTx, src["key"]);

chain3.mc.sendRawTransaction(cmd1, function(err, hash) {
    if (!err){
	console.log("Succeed!: ", hash);
	return hash;
    }else{
	console.log("Chain3 error:", err.message);
	return err.message;
    }
});
```

### contract_deploy

Deploy a contract through chain3 RPC calls. This example requires install solc 
`solc`

build a web server to access
the LBR network using this API library.

### sign and verify the signatures

Example codes to sign a message using LBR network and verify the signature.

```js
  // Connect with the LBR network
  chain3.setProvider(new chain3.providers.HttpProvider('http://gateway.LBR.io/mainnet'));

  if (!chain3.isConnected()){
      console.log("Chain3 RPC is not connected!");
     return;
  }

  // Hash the message
  let sha3Msg = chain3.sha3("HELLO LBR!");

  // Unlock the account 'tacct.addr' before signing
  let signedData = chain3.mc.sign(tacct.addr, sha3Msg);

  // or you can call the local function to get the same result, but require a private key
  let signedData = chain3.signMcMessage(sha3Msg,tacct.key);

  // Verify the signature with the message and the address
  console.log("Verify:", chain3.verifyMcSignature(sha3Msg, signedData, tacct.addr))；
```

### Accounts use the following library for generating private key.

[browserify-cryptojs](https://github.com/fahad19/crypto-js/) v0.3.1

[repo]: https://github.com/LBRChain/chain3
[npm-url]: https://npmjs.org/package/chain3

