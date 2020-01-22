---
name: Javascript API

---

# Chain3 JavaScript API
LBR JavaScript API was built for LBR chain. It was developed based on the Ethereum web3.js API routines and made some methods compatiable with the web3.js methods so the users can easily move their Ðapp to LBR chain. In
Chain3 also supported additional methods for LBR platform, such as VNODE and SCS methods.

To make a Ðapp work on LBR network, user should use the `chain3` object provided by the [chain3.js library](https://github.com/LBRChain/chain3). It communicates to a local LBR  node through [JSON RPC](https://github.com/LBRChain/LBR-core/wiki/JSON-RPC). chain3.js works with any LBR VNODE and SCS, which exposes an RPC layer.

`chain3` contains the `mc` object - `chain3.mc` (for specifically LBR Mother blockchain interactions), and the `scs` object - `chain3.scs` (for MicroChain interactions). Over time we'll introduce other objects for each of the other chain3 protocols. Working [examples can be found here](https://github.com/LBRChain/chain3/tree/master/example).


## Using callbacks

As this API is designed to work with a local RPC node and all its functions are by default use synchronous HTTP requests.con

If you want to make an asynchronous request, you can pass an optional callback as the last parameter to most functions.
All callbacks are using an [error first callback](http://fredkschott.com/post/2014/03/understanding-error-first-callbacks-in-node-js/) style:

```js
chain3.mc.getBlock(48, function(error, result){
    if(!error)
        console.log(result)
    else
        console.error(error);
})
```

## Batch requests

Batch requests allow queuing up requests and processing them at once.

```js
var batch = chain3.createBatch();
batch.add(chain3.mc.getBalance.request('0x0000000000000000000000000000000000000000', 'latest', callback));
batch.add(chain3.mc.contract(abi).at(address).balance.request(address, callback2));
batch.execute();
```

## A note on big numbers in chain3.js

You will always get a BigNumber object for balance values as JavaScript is not able to handle big numbers correctly.
Look at the following examples:

```js
"101010100324325345346456456456456456456"
// "101010100324325345346456456456456456456"
101010100324325345346456456456456456456
// 1.0101010032432535e+38
```

chain3.js depends on the [BigNumber Library](https://github.com/MikeMcl/bignumber.js/) and adds it automatically.

```js
var balance = new BigNumber('131242344353464564564574574567456');
// or var balance = chain3.mc.getBalance(someAddress);

balance.plus(21).toString(10); // toString(10) converts it to a number string
// "131242344353464564564574574567477"
```

The next example wouldn't work as we have more than 20 floating points, therefore it is recommended that you always keep your balance  in *sha* and only transform it to other units when presenting to the user:
```js
var balance = new BigNumber('13124.234435346456466666457455567456');

balance.plus(21).toString(10); // toString(10) converts it to a number string, but can only show max 20 floating points 
// "13145.23443534645646666646" // you number would be cut after the 20 floating point
```

## Chain3 Javascript Ðapp API Reference

* [chain3](#chain3)
  * [version](#chain3versionapi)
     * [api](#chain3versionapi)
     * [node](#chain3versionnode)
     * [network](#chain3versionnetwork)
     * [LBR](#chain3versionLBR)
  * [isConnected()](####chain3isconnected)
  * [setProvider(provider)](#chain3setprovider)
  * [currentProvider](#chain3currentprovider)
  * [reset()](#chain3reset)
  * [sha3(string)](#chain3sha3)
  * [toHex(stringOrNumber)](#chain3tohex)
  * [toAscii(hexString)](#chain3toascii)
  * [fromAscii(textString, [padding])](#chain3fromascii)
  * [toDecimal(hexString)](#chain3todecimal)
  * [toChecksumAddress(string)](#chain3tochecksumaddress)
  * [fromDecimal(number)](#chain3fromdecimal)
  * [fromSha(numberStringOrBigNumber, unit)](#chain3fromsha)
  * [toSha(numberStringOrBigNumber, unit)](#chain3tosha)
  * [toBigNumber(numberOrHexString)](#chain3tobignumber)
  * [isAddress(hexString)](#chain3isAddress)
  * [net](#chain3net)
     * [listening/getListening](#chain3netlistening)
     * [peerCount/getPeerCount](#chain3mcpeercount)
  * [mc](#chain3mc)
     * [defaultAccount](#chain3mcdefaultaccount)
     * [defaultBlock](#chain3mcdefaultblock)
     * [syncing/getSyncing](#chain3mcsyncing)
     * [isSyncing](#chain3mcissyncing)
     * [coinbase/getCoinbase](#chain3mccoinbase)
     * [hashrate/getHashrate](#chain3mchashrate)
     * [gasPrice/getGasPrice](#chain3mcgasprice)
     * [accounts/getAccounts](#chain3mcaccounts)
     * [mining/getMining](#chain3mcmining)
     * [blockNumber/getBlockNumber](#chain3mcblocknumber)
     * [getBalance(address)](#chain3mcgetbalance)
     * [getStorageAt(address, position)](#chain3mcgetstorageat)
     * [getCode(address)](#chain3mcgetcode)
     * [getBlock(hash/number)](#chain3mcgetblock)
     * [getBlockTransactionCount(hash/number)](#chain3mcgetblocktransactioncount)
     * [getUncle(hash/number)](#chain3mcgetuncle)
     * [getBlockUncleCount(hash/number)](#chain3mcgetblockunclecount)
     * [getTransaction(hash)](#chain3mcgettransaction)
     * [getTransactionFromBlock(hashOrNumber, indexNumber)](#chain3mcgettransactionfromblock)
     * [getTransactionReceipt(hash)](#chain3mcgettransactionreceipt)
     * [getTransactionCount(address)](#chain3mcgettransactioncount)
     * [sendTransaction(object)](#chain3mcsendtransaction)
     * [call(object)](#chain3mccall)
     * [estimateGas(object)](#chain3mcestimategas)
     * [filter(array (, options) )](#chain3mcfilter)
        - [watch(callback)](#chain3mcfilter)
        - [stopWatching(callback)](#chain3mcfilter)
        - [get()](#chain3mcfilter)
    * [contract(abiArray)](#chain3mccontract)
    * [contract.myMethod()](#contract-methods)
    * [contract.myEvent()](#contract-events)
    * [contract.allEvents()](#contract-allevents)
    * [encodeParams](#chain3encodeParams)
    * [namereg](#chain3mcnamereg)
    * [sendIBANTransaction](#chain3mcsendibantransaction)
    * [iban](#chain3mciban)
      * [fromAddress](#chain3mcibanfromaddress)
      * [fromBban](#chain3mcibanfrombban)
      * [createIndirect](#chain3mcibancreateindirect)
      * [isValid](#chain3mcibanisvalid)
      * [isDirect](#chain3mcibanisdirect)
      * [isIndirect](#chain3mcibanisindirect)
      * [checksum](#chain3mcibanchecksum)
      * [institution](#chain3mcibaninstitution)
      * [client](#chain3mcibanclient)
      * [address](#chain3mcibanaddress)
      * [toString](#chain3mcibantostring)
    * [admin](#chain3admin)
      * [addPeer](#admin_addPeer)
      * [datadir](#admin_datadir)
      * [nodeInfo](#admin_nodeInfo)
      * [peers](#admin_peers)
      * [stopRPC](#admin_stopRPC)
      * [startRPC](#admin_startRPC)
    * [txpool](#chain3txpool)
      * [content](#txpool_content)
      * [status](#txpool_status)
      * [inspect](#txpool_inspect)
  	* [vnode](#chain3vnode)
	  	* [vnodeAddress](#chain3vnode_address)
  		* [scsService](#vnode_scsservice)
  		* [serviceCfg](#vnode_servicecfg)
  		* [showToPublic](#vnode_showtopublic)
  		* [vnodeIP](#vnode_vnodeip)
  	* [scs](#chain3scs)
  	  * [directCall](#scs_directcall)
  	  * [getBlock](#scs_getblock)
  	  * [getBlockNumber](#scs_getblocknumber)
  	  * [getDappState](#scs_getdappstate)
  	  * [getMicroChainList](#scs_getmicrochainlist)
  	  * [getMicroChainInfo](#scs_getmicrochaininfo)
  	  * [getNonce](#scs_getnonce)
  	  * [getSCSId](#scs_getscsid)
  	  * [getTransactionByHash](#scs_gettransactionbyhash)
      * [getTransactionByNonce](#scs_gettransactionbynonce)      
      * [getReceiptByHash](#scs_getreceiptbyhash)
      * [getReceiptByNonce](#scs_getreceiptbynonce)    
      * [getExchangeByAddress](#scs_getexchangebyaddress)
      * [getExchangeInfo](#scs_getexchangeinfo)    
      * [getTxpool](#scs_gettxpool)   


### Usage
***

<h4 id="chain3">chain3.version.api</h4>

The `chain3` object provides all methods. 

##### Example

```js
var Chain3 = require('chain3');
// create an instance of chain3 using the HTTP provider.
var chain3 = new Chain3(new Chain3.providers.HttpProvider("http://localhost:8545"));
```

***

<h4 id="chain3versionapi">chain3.version.api</h4>

```js
chain3.version.api
// or async
chain3.version.getApi(callback(error, result){ ... })
```

##### Returns

`String` - The LBR js api version.

##### Example

```js
var version = chain3.version.api;
console.log(version); // "0.2.0"
```

***

<h4 id="chain3versionnode">chain3.version.node</h4>

    chain3.version.node
    // or async
    chain3.version.getClient(callback(error, result){ ... })


##### Returns

`String` - The client/node version.

##### Example

```js
var version = chain3.version.node;
console.log(version); // "LBR/v0.1.0-develop/darwin-amd64/go1.9"
```

***

<h4 id="chain3versionnetwork">chain3.version.network</h4>


    chain3.version.network
    // or async
    chain3.version.getNetwork(callback(error, result){ ... })


##### Returns

`String` - The network protocol version.

##### Example

```js
var version = chain3.version.network;
console.log(version); // 54
```

***

<h4 id="chain3versionnetLBR">chain3.version.LBR </h4>


    chain3.version.LBR
    // or async
    chain3.version.getLBR(callback(error, result){ ... })


##### Returns

`String` - The LBR protocol version.

##### Example

```js
var version = chain3.version.LBR;
console.log(version); // 0x3f
```

***

<h4 id="chain3isconnected">chain3.isConnected </h4>


    chain3.isConnected()

Should be called to check if a connection to a node exists

##### Parameters
none

##### Returns

`Boolean`

##### Example

```js
if(!chain3.isConnected()) {
  
   // show some dialog to ask the user to start a node

} else {
 
   // start chain3 filters, calls, etc
  
}
```

***

<h4 id="chain3setprovider">chain3.setProvider </h4>


    chain3.setProvider(provider)

Should be called to set provider.

##### Parameters
none

##### Returns

`undefined`

##### Example

```js
chain3.setProvider(new chain3.providers.HttpProvider('http://localhost:8545')); // 8545 for go/LBR
```

***

<h4 id="chain3currentprovider">chain3.currentProvider </h4>

    chain3.currentProvider

Will contain the current provider, if one is set. This can be used to check if LBR etc. set already a provider.


##### Returns

`Object` - The provider set or `null`;

##### Example

```js
// Check if LBR etc. already set a provider
if(!chain3.currentProvider)
    chain3.setProvider(new chain3.providers.HttpProvider("http://localhost:8545"));

```

***

<h4 id="chain3currentreset">chain3.reset </h4>

    chain3.reset(keepIsSyncing)

Should be called to reset state of chain3. Resets everything except manager. Uninstalls all filters. Stops polling.

##### Parameters

1. `Boolean` - If `true` it will uninstall all filters, but will keep the [chain3.mc.isSyncing()](#chain3mcissyncing) polls

##### Returns

`undefined`

##### Example

```js
chain3.reset();
```

***

<h4 id="chain3sha3">chain3.sha3 </h4>

    chain3.sha3(string [, callback])

##### Parameters

1. `String` - The string to hash using the SHA3 algorithm
2. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.


##### Returns

`String` - The SHA3 of the given data.

##### Example

```js
var str = chain3.sha3("Some ASCII string to be hashed in LBR");
console.log(str); // "0xbfa24877cd68e6734710402a2af3e29cf18bd6d2f304aa528ffa3a32fa7652d2"

```

***

<h4 id="chain3tohex">chain3.toHex</h4>

    chain3.toHex(mixed);
 
Converts any value into HEX.

##### Parameters

1. `String|Number|Object|Array|BigNumber` - The value to parse to HEX. If its an object or array it will be `JSON.stringify` first. If its a BigNumber it will make it the HEX value of a number.

##### Returns

`String` - The hex string of `mixed`.

##### Example

```js
var str = chain3.toHex(LBR network);
console.log(str); // '0x6d6f6163206e6574776f726b'

console.log(chain3.toHex({LBR: 'test'}));
//'0x7b226d6f6163223a2274657374227d'
```

***

<h4 id="chain3toascii">chain3.toAscii</h4>

#### chain3.toAscii

    chain3.toAscii(hexString);

Converts a HEX string into a ASCII string.

##### Parameters

1. `String` - A HEX string to be converted to ascii.

##### Returns

`String` - An ASCII string made from the given `hexString`.

##### Example

```js
var str = chain3.toAscii("0x0x6d6f6163206e6574776f726b");
console.log(str); // "LBR network"
```

***

<h4 id="chain3fromascii">chain3.fromAscii</h4>

    chain3.fromAscii(string);

Converts any ASCII string to a HEX string.

##### Parameters

`String` - An ASCII string to be converted to HEX.

##### Returns

`String` - The converted HEX string.

##### Example

```js
var str = chain3.fromAscii('LBR network');
console.log(str); // "0x0x6d6f6163206e6574776f726b"
```

***

<h4 id="chain3tochecksumaddress">chain3.toChecksumAddress </h4>


    chain3.toChecksumAddress(hexString);

Converts a string to the checksummed address equivalent.

##### Parameters

1. `String` - A string to be converted to a checksummed address.


##### Returns

`String` - A string containing the checksummed address.

##### Example

```js
var myAddress = chain3.toChecksumAddress('0xa0c876ec9f2d817c4304a727536f36363840c02c');
console.log(myAddress); // '0xA0C876eC9F2d817c4304A727536f36363840c02c'
```

***

<h4 id="chain3todecimal">chain3.toDecimal </h4>

    chain3.toDecimal(hexString);

Converts a HEX string to its number representation.

##### Parameters

1. `String` - An HEX string to be converted to a number.


##### Returns

`Number` - The number representing the data `hexString`.

##### Example

```js
var number = chain3.toDecimal('0x15');
console.log(number); // 21
```

***

<h4 id="chain3fromdecimal">chain3.fromDecimal</h4>

    chain3.fromDecimal(number);

Converts a number or number string to its HEX representation.

##### Parameters

1. `Number|String` - A number to be converted to a HEX string.

##### Returns

`String` - The HEX string representing of the given `number`.

##### Example

```js
var value = chain3.fromDecimal('21');
console.log(value); // "0x15"
```

***

<h4 id="chain3fromsha">chain3.fromSha</h4>

    chain3.fromSha(number, unit)

Converts a number of sha into the following LBR units:

- `ksha`/`femtomc`
- `msha`/`picomc`
- `gsha`/`nano`/`xiao`
- `micro`/`sand`
- `milli`
- `mc`
- `kmc`/`grand`
- `mmc`
- `gmc`
- `tmc`

##### Parameters

1. `Number|String|BigNumber` - A number or BigNumber instance.
2. `String` - One of the above LBR units.


##### Returns

`String|BigNumber` - Either a number string, or a BigNumber instance, depending on the given `number` parameter.

##### Example

```js
var value = chain3.fromSha('21000000000000', 'Xiao');
console.log(value); // "21000"
```

***

<h4 id="chain3tosha">chain3.toSha </h4>

    chain3.toSha(number, unit)

Converts a LBR unit into sha. Possible units are:

- `ksha`/`femtomc`
- `msha`/`picomc`
- `gsha`/`nano`/`xiao`
- `micro`/`sand`
- `milli`
- `mc`
- `kmc`/`grand`
- `mmc`
- `gmc`
- `tmc`

##### Parameters

1. `Number|String|BigNumber` - A number or BigNumber instance.
2. `String` - One of the above LBR units.

##### Returns

`String|BigNumber` - Either a number string, or a BigNumber instance, depending on the given `number` parameter.

##### Example

```js
var value = chain3.toSha('1', 'mc');
console.log(value); // "1000000000000000000" = 1e18
```

***

<h4 id="chain3tobignumber">chain3.toBigNumber </h4>

    chain3.toBigNumber(numberOrHexString);

Converts a given number into a BigNumber instance.

See the [note on BigNumber](#a-note-on-big-numbers-in-javascript).

##### Parameters

1. `Number|String` - A number, number string or HEX string of a number.


##### Returns

`BigNumber` - A BigNumber instance representing the given value.


##### Example

```js
var value = chain3.toBigNumber('200000000000000000000001');
console.log(value); // instanceOf BigNumber
console.log(value.toNumber()); // 2.0000000000000002e+23
console.log(value.toString(10)); // '200000000000000000000001'
```

***

<h3 id="chain3net"> chain3.net </h3>

<h4 id="chain3netlistening">chain3.net.listening</h4>

    chain3.net.listening
    // or async
    chain3.net.getListening(callback(error, result){ ... })

This property is read only and says whether the node is actively listening for network connections or not.

##### Returns

`Boolean` - `true` if the client is actively listening for network connections, otherwise `false`.

##### Example

```js
var listening = chain3.net.listening;
console.log(listening); // true of false
```

***

<h4 id="chain3netpeercount">chain3.net.peerCount </h4>

    chain3.net.peerCount
    // or async
    chain3.net.getPeerCount(callback(error, result){ ... })

This property is read only and returns the number of connected peers.

##### Returns

`Number` - The number of peers currently connected to the client.

##### Example

```js
var peerCount = chain3.net.peerCount;
console.log(peerCount); // 4
```

***

<h3 id="chain3mc"> chain3.mc </h3>

Contains the LBR blockchain related methods.

##### Example

```js
var mc = chain3.mc;
```

***

<h4 id="chain3mcdefaultaccount">chain3.mc.defaultAccount </h4>

    chain3.mc.defaultAccount

This default address is used for the following methods (optionally you can overwrite it by specifying the `from` property):

- [chain3.mc.sendTransaction()](#chain3mcsendtransaction)
- [chain3.mc.call()](#chain3mccall)

##### Values

`String`, 20 Bytes - Any address you own, or where you have the private key for.

*Default is* `undefined`.

##### Returns

`String`, 20 Bytes - The currently set default address.

##### Example

```js
var defaultAccount = chain3.mc.defaultAccount;
console.log(defaultAccount); // ''

// set the default block
chain3.mc.defaultAccount = '0x8888f1f195afa192cfee860698584c030f4c9db1';
```

***

<h4 id="chain3mcdefaultblock">chain3.mc.defaultBlock </h4>

    chain3.mc.defaultBlock

This default block is used for the following methods (optionally you can overwrite the defaultBlock by passing it as the last parameter):

- [chain3.mc.getBalance()](#chain3mcgetbalance)
- [chain3.mc.getCode()](#chain3mcgetcode)
- [chain3.mc.getTransactionCount()](#chain3mcgettransactioncount)
- [chain3.mc.getStorageAt()](#chain3mcgetstorageat)
- [chain3.mc.call()](#chain3mccall)

##### Values

Default block parameters can be one of the following:

- `Number` - a block number
- `String` - `"earliest"`, the genisis block
- `String` - `"latest"`, the latest block (current head of the blockchain)
- `String` - `"pending"`, the currently mined block (including pending transactions)

*Default is* `latest`

##### Returns

`Number|String` - The default block number to use when querying a state.

##### Example

```js
var defaultBlock = chain3.mc.defaultBlock;
console.log(defaultBlock); // 'latest'

// set the default block
chain3.mc.defaultBlock = 231;
```

***

<h4 id="chain3mcsyncing">chain3.mc.syncing </h4>

    chain3.mc.syncing
    // or async
    chain3.mc.getSyncing(callback(error, result){ ... })

This property is read only and returns the either a sync object, when the node is syncing or `false`.

##### Returns

`Object|Boolean` - A sync object as follows, when the node is currently syncing or `false`:
   - `startingBlock`: `Number` - The block number where the sync started.
   - `currentBlock`: `Number` - The block number where at which block the node currently synced to already.
   - `highestBlock`: `Number` - The estimated block number to sync to.

##### Example

```js
var sync = chain3.mc.syncing;
console.log(sync);
/*
{
   startingBlock: 300,
   currentBlock: 312,
   highestBlock: 512
}
*/
```

***

<h4 id="chain3mcissyncing">chain3.mc.isSyncing </h4>

    chain3.mc.isSyncing(callback);

This convenience function calls the `callback` everytime a sync starts, updates and stops.

##### Returns

`Object` - a isSyncing object with the following methods:

  * `syncing.addCallback()`: Adds another callback, which will be called when the node starts or stops syncing.
  * `syncing.stopWatching()`: Stops the syncing callbacks.

##### Callback return value

- `Boolean` - The callback will be fired with `true` when the syncing starts and with `false` when it stopped.
- `Object` - While syncing it will return the syncing object:
   - `startingBlock`: `Number` - The block number where the sync started.
   - `currentBlock`: `Number` - The block number where at which block the node currently synced to already.
   - `highestBlock`: `Number` - The estimated block number to sync to.


##### Example

```js
chain3.mc.isSyncing(function(error, sync){
    if(!error) {
        // stop all app activity
        if(sync === true) {
           // we use `true`, so it stops all filters, but not the chain3.mc.syncing polling
           chain3.reset(true);
        
        // show sync info
        } else if(sync) {
           console.log(sync.currentBlock);
        
        // re-gain app operation
        } else {
            // run your app init function...
        }
    }
});
```

***

<h4 id="chain3mcissyncing">chain3.mc.isSyncing </h4>

    chain3.mc.coinbase
    // or async
    chain3.mc.getCoinbase(callback(error, result){ ... })

This property is read only and returns the coinbase address were the mining rewards go to.

##### Returns

`String` - The coinbase address of the client.

##### Example

```js
var coinbase = chain3.mc.coinbase;
console.log(coinbase); // "0x407d73d8a49eeb85d32cf465507dd71d507100c1"
```

***

<h4 id="chain3mcmining">chain3.mc.mining </h4>

    chain3.mc.mining
    // or async
    chain3.mc.getMining(callback(error, result){ ... })


This property is read only and says whether the node is mining or not.


##### Returns

`Boolean` - `true` if the client is mining, otherwise `false`.

##### Example

```js
var mining = chain3.mc.mining;
console.log(mining); // true or false
```

***

<h4 id="chain3mchashrate">chain3.mc.hashrate </h4>

    chain3.mc.hashrate
    // or async
    chain3.mc.getHashrate(callback(error, result){ ... })

This property is read only and returns the number of hashes per second that the node is mining with.


##### Returns

`Number` - number of hashes per second.

##### Example

```js
var hashrate = chain3.mc.hashrate;
console.log(hashrate); // 493736
```

***

<h4 id="chain3mcgasprice">chain3.mc.gasPrice </h4>

    chain3.mc.gasPrice
    // or async
    chain3.mc.getGasPrice(callback(error, result){ ... })


This property is read only and returns the current gas price.
The gas price is determined by the x latest blocks median gas price.

##### Returns

`BigNumber` - A BigNumber instance of the current gas price in sha.

See the [note on BigNumber](#a-note-on-big-numbers-in-javascript).

##### Example

```js
var gasPrice = chain3.mc.gasPrice;
console.log(gasPrice.toString(10)); // "10000000000000"
```

***

<h4 id="chain3mcaccounts">chain3.mc.accounts </h4>

    chain3.mc.accounts
    // or async
    chain3.mc.getAccounts(callback(error, result){ ... })

This property is read only and returns a list of accounts the node controls.

##### Returns

`Array` - An array of addresses controlled by client.

##### Example

```js
var accounts = chain3.mc.accounts;
console.log(accounts); // ["0x407d73d8a49eeb85d32cf465507dd71d507100c1"] 
```

***

<h4 id="chain3mcblocknumber">chain3.mc.blockNumber </h4>

    chain3.mc.blockNumber
    // or async
    chain3.mc.getBlockNumber(callback(error, result){ ... })

This property is read only and returns the current block number.

##### Returns

`Number` - The number of the most recent block.

##### Example

```js
var number = chain3.mc.blockNumber;
console.log(number); // 2744
```

***

<h4 id="chain3mcgetbalance">chain3.mc.getBalance </h4>

    chain3.mc.getBalance(addressHexString [, defaultBlock] [, callback])

Get the balance of an address at a given block.

##### Parameters

1. `String` - The address to get the balance of.
2. `Number|String` - (optional) If you pass this parameter it will not use the default block set with [chain3.mc.defaultBlock](#chain3mcdefaultblock).
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`String` - A BigNumber instance of the current balance for the given address in sha.

See the [note on BigNumber](#a-note-on-big-numbers-in-javascript).

##### Example

```js
var balance = chain3.mc.getBalance("0x407d73d8a49eeb85d32cf465507dd71d507100c1");
console.log(balance); // instanceof BigNumber
console.log(balance.toString(10)); // '1000000000000'
console.log(balance.toNumber()); // 1000000000000
```

***

<h4 id="chain3mcgetstorageat">chain3.mc.getStorageAt </h4>

    chain3.mc.getStorageAt(addressHexString, position [, defaultBlock] [, callback])

Get the storage at a specific position of an address.

##### Parameters

1. `String` - The address to get the storage from.
2. `Number` - The index position of the storage.
3. `Number|String` - (optional) If you pass this parameter it will not use the default block set with [chain3.mc.defaultBlock](#chain3mcdefaultblock).
4. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.


##### Returns

`String` - The value in storage at the given position.

##### Example

```js
var state = chain3.mc.getStorageAt("0x407d73d8a49eeb85d32cf465507dd71d507100c1", 0);
console.log(state); // "0x03"
```

***

<h4 id="chain3mcgetcode">chain3.mc.getCode </h4>

    chain3.mc.getCode(addressHexString [, defaultBlock] [, callback])

Get the code at a specific address.

##### Parameters

1. `String` - The address to get the code from.
2. `Number|String` - (optional) If you pass this parameter it will not use the default block set with [chain3.mc.defaultBlock](#chain3mcdefaultblock).
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`String` - The data at given address `addressHexString`.

##### Example

```js
var code = chain3.mc.getCode("0xd5677cf67b5aa051bb40496e68ad359eb97cfbf8");
console.log(code); // "0x600160008035811a818181146012578301005b601b6001356025565b8060005260206000f25b600060078202905091905056"
```

***

<h4 id="chain3mcgetblock">chain3.mc.getBlock </h4>

     chain3.mc.getBlock(blockHashOrBlockNumber [, returnTransactionObjects] [, callback])

Returns a block matching the block number or block hash.

##### Parameters

1. `String|Number` - The block number or hash. Or the string `"earliest"`, `"latest"` or `"pending"` as in the [default block parameter](#chain3mcdefaultblock).
2. `Boolean` - (optional, default `false`) If `true`, the returned block will contain all transactions as objects, if `false` it will only contains the transaction hashes.
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`Object` - The block object:

  - `number`: `Number` - the block number. `null` when its pending block.
  - `hash`: `String`, 32 Bytes - hash of the block. `null` when its pending block.
  - `parentHash`: `String`, 32 Bytes - hash of the parent block.
  - `nonce`: `String`, 8 Bytes - hash of the generated proof-of-work. `null` when its pending block.
  - `sha3Uncles`: `String`, 32 Bytes - SHA3 of the uncles data in the block.
  - `logsBloom`: `String`, 256 Bytes - the bloom filter for the logs of the block. `null` when its pending block.
  - `transactionsRoot`: `String`, 32 Bytes - the root of the transaction trie of the block
  - `stateRoot`: `String`, 32 Bytes - the root of the final state trie of the block.
  - `miner`: `String`, 20 Bytes - the address of the beneficiary to whom the mining rewards were given.
  - `difficulty`: `BigNumber` - integer of the difficulty for this block.
  - `totalDifficulty`: `BigNumber` - integer of the total difficulty of the chain until this block.
  - `extraData`: `String` - the "extra data" field of this block.
  - `size`: `Number` - integer the size of this block in bytes.
  - `gasLimit`: `Number` - the maximum gas allowed in this block.
  - `gasUsed`: `Number` - the total used gas by all transactions in this block.
  - `timestamp`: `Number` - the unix timestamp for when the block was collated.
  - `transactions`: `Array` - Array of transaction objects, or 32 Bytes transaction hashes depending on the last given parameter.
  - `uncles`: `Array` - Array of uncle hashes.

##### Example

```js
var info = chain3.mc.block(3150);
console.log(info);
/*
{
  difficulty: 142913,
  extraData: "0xd5820100846d6f616385676f312e398664617277696e",
  gasLimit: 1732862932,
  gasUsed: 20000,
  hash: "0x43ab9feea269585e11b0c4c9d91449cbeae44395dcc2eb1d726a64e5a00047f7",
  logsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
  miner: "0xa8863fc8ce3816411378685223c03daae9770ebb",
  mixHash: "0x9c879f1f4570f832091222f31495db378b43496e3115c96bc40f9f57915f1b59",
  nonce: "0x71b4365850bc8a10",
  number: 929,
  parentHash: "0x9e460922544a32fb9c634ee93793f19bcb033db53ac07c6008cbb2b354904d74",
  receiptsRoot: "0x1a8c28ca2760b9534a1d1edfb4b3748f62f875631eccd2656babeb9cefdc6fdb",
  sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
  size: 667,
  stateRoot: "0xb788c2ba4111b8ad4bd7fc39908dd63200fa93618f105bee8495723ac0f227c4",
  timestamp: 1518453915,
  totalDifficulty: 135134971,
  transactions: ["0x0c2a15d54db54f69d81dbb0046fc0e3acd8e790fda9e198d09827292ceb0ba13"],
  transactionsRoot: "0xdc160a6056769aa497940af2155ab2d4d106aab5fca352559a2918b17ff6087f",
  uncles: []
}
*/
```

***

<h4 id="chain3mcgetblocktransactioncount">chain3.mc.getBlockTransactionCount </h4>

    chain3.mc.getBlockTransactionCount(hashStringOrBlockNumber [, callback])

Returns the number of transaction in a given block.

##### Parameters

1. `String|Number` - The block number or hash. Or the string `"earliest"`, `"latest"` or `"pending"` as in the [default block parameter](#chain3mcdefaultblock).
2. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`Number` - The number of transactions in the given block.

##### Example

```js
var number = chain3.mc.getBlockTransactionCount("0x407d73d8a49eeb85d32cf465507dd71d507100c1");
console.log(number); // 1
```

***

<h4 id="chain3mcgetuncle">chain3.mc.getUncle </h4>

    chain3.mc.getUncle(blockHashStringOrNumber, uncleNumber [, returnTransactionObjects] [, callback])

Returns a blocks uncle by a given uncle index position.

##### Parameters

1. `String|Number` - The block number or hash. Or the string `"earliest"`, `"latest"` or `"pending"` as in the [default block parameter](#chain3mcdefaultblock).
2. `Number` - The index position of the uncle.
3. `Boolean` - (optional, default `false`) If `true`, the returned block will contain all transactions as objects, if `false` it will only contains the transaction hashes.
4. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.


##### Returns

`Object` - the returned uncle. For a return value see [chain3.mc.getBlock()](#chain3mcgetblock).

**Note**: An uncle doesn't contain individual transactions.

##### Example

```js
var uncle = chain3.mc.getUncle(500, 0);
console.log(uncle); // see chain3.mc.getBlock

```

***

<h4 id="chain3mcgettransaction">chain3.mc.getTransaction </h4>


    chain3.mc.getTransaction(transactionHash [, callback])

Returns a transaction matching the given transaction hash.

##### Parameters

1. `String` - The transaction hash.
2. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.


##### Returns

`Object` - A transaction object its hash `transactionHash`:

  - `hash`: `String`, 32 Bytes - hash of the transaction.
  - `nonce`: `Number` - the number of transactions made by the sender prior to this one.
  - `blockHash`: `String`, 32 Bytes - hash of the block where this transaction was in. `null` when its pending.
  - `blockNumber`: `Number` - block number where this transaction was in. `null` when its pending.
  - `transactionIndex`: `Number` - integer of the transactions index position in the block. `null` when its pending.
  - `from`: `String`, 20 Bytes - address of the sender.
  - `to`: `String`, 20 Bytes - address of the receiver. `null` when its a contract creation transaction.
  - `value`: `BigNumber` - value transferred in Sha.
  - `gasPrice`: `BigNumber` - gas price provided by the sender in Sha.
  - `gas`: `Number` - gas provided by the sender.
  - `input`: `String` - the data sent along with the transaction.


##### Example

```js
var txhash = "0x0c2a15d54db54f69d81dbb0046fc0e3acd8e790fda9e198d09827292ceb0ba13";

var transaction = chain3.mc.getTransaction(txhash);
console.log(transaction);
/*
{
  blockHash: "0x43ab9feea269585e11b0c4c9d91449cbeae44395dcc2eb1d726a64e5a00047f7",
  blockNumber: 929,
  from: "0x7312f4b8a4457a36827f185325fd6b66a3f8bb8b",
  gas: 2000000,
  gasPrice: 0,
  hash: "0x0c2a15d54db54f69d81dbb0046fc0e3acd8e790fda9e198d09827292ceb0ba13",
  input: "0x",
  nonce: 1,
  r: "0x5487c81b1ab58a7364b39263fd51afee9e7ad32be1c3b5ddeb656f06d75b1d68",
  s: "0x3f06414699c02f74c5075e24c5f266a25a64dbdf77ce82e57b1def6c5e198f9f",
  to: "0xd814f2ac2c4ca49b33066582e4e97ebae02f2ab9",
  transactionIndex: 0,
  v: "0x25",
  value: 500000000000000000
}
*/

```

***

<h4 id="chain3mcgettransactionfromBlock">chain3.mc.getTransactionFromBlock </h4>

    getTransactionFromBlock(hashStringOrNumber, indexNumber [, callback])

Returns a transaction based on a block hash or number and the transactions index position.

##### Parameters

1. `String` - A block number or hash. Or the string `"earliest"`, `"latest"` or `"pending"` as in the [default block parameter](#chain3mcdefaultblock).
2. `Number` - The transactions index position.
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`Object` - A transaction object, see [chain3.mc.getTransaction](#chain3mcgettransaction):


##### Example

```js
var transaction = chain3.mc.getTransactionFromBlock(921, 2);
console.log(transaction); // see chain3.mc.getTransaction

```

***

<h4 id="chain3mcgettransactionreceipt">chain3.mc.getTransactionReceipt </h4>


    chain3.mc.getTransactionReceipt(hashString [, callback])

Returns the receipt of a transaction by transaction hash.

**Note** That the receipt is not available for pending transactions.


##### Parameters

1. `String` - The transaction hash.
2. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`Object` - A transaction receipt object, or `null` when no receipt was found:

  - `blockHash`: `String`, 32 Bytes - hash of the block where this transaction was in.
  - `blockNumber`: `Number` - block number where this transaction was in.
  - `transactionHash`: `String`, 32 Bytes - hash of the transaction.
  - `transactionIndex`: `Number` - integer of the transactions index position in the block.
  - `from`: `String`, 20 Bytes - address of the sender.
  - `to`: `String`, 20 Bytes - address of the receiver. `null` when its a contract creation transaction.
  - `cumulativeGasUsed `: `Number ` - The total amount of gas used when this transaction was executed in the block.
  - `gasUsed `: `Number ` -  The amount of gas used by this specific transaction alone.
  - `contractAddress `: `String` - 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise `null`.
  - `logs `:  `Array` - Array of log objects, which this transaction generated.

##### Example
```js
var receipt = chain3.mc.getTransactionReceipt('0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b');
console.log(receipt);
{
  "transactionHash": "0x9fc76417374aa880d4449a1f7f31ec597f00b1f6f3dd2d66f4c9c6c445836d8b",
  "transactionIndex": 0,
  "blockHash": "0xef95f2f1ed3ca60b048b4bf67cde2195961e0bba6f70bcbea9a2c4e133e34b46",
  "blockNumber": 3,
  "contractAddress": "0xa94f5374fce5edbc8e2a8697c15331677e6ebf0b",
  "cumulativeGasUsed": 314159,
  "gasUsed": 30234,
  "logs": [{
         // logs as returned by getFilterLogs, etc.
     }, ...]
}
```

***

<h4 id="chain3mcgettransactioncount">chain3.mc.getTransactionCount </h4>

    chain3.mc.getTransactionCount(addressHexString [, defaultBlock] [, callback])

Get the numbers of transactions sent from this address.

##### Parameters

1. `String` - The address to get the numbers of transactions from.
2. `Number|String` - (optional) If you pass this parameter it will not use the default block set with [chain3.mc.defaultBlock](#chain3mcdefaultblock).
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`Number` - The number of transactions sent from the given address.

##### Example

```js
var number = chain3.mc.getTransactionCount("0x407d73d8a49eeb85d32cf465507dd71d507100c1");
console.log(number); // 1
```

***

<h4 id="chain3mcsendtransaction">chain3.mc.sendTransaction </h4>


    chain3.mc.sendTransaction(transactionObject [, callback])

Sends a transaction to the network.

##### Parameters

1. `Object` - The transaction object to send:
  - `from`: `String` - The address for the sending account. Uses the [chain3.mc.defaultAccount](#chain3mcdefaultaccount) property, if not specified.
  - `to`: `String` - (optional) The destination address of the message, left undefined for a contract-creation transaction.
  - `value`: `Number|String|BigNumber` - (optional) The value transferred for the transaction in Sha, also the endowment if it's a contract-creation transaction.
  - `gas`: `Number|String|BigNumber` - (optional, default: To-Be-Determined) The amount of gas to use for the transaction (unused gas is refunded).
  - `gasPrice`: `Number|String|BigNumber` - (optional, default: To-Be-Determined) The price of gas for this transaction in sha, defaults to the mean network gas price.
  - `data`: `String` - (optional) Either a byte string containing the associated data of the message, or in the case of a contract-creation transaction, the initialisation code.
  - `nonce`: `Number`  - (optional) Integer of a nonce. This allows to overwrite your own pending transactions that use the same nonce.
  - `scsConsensusAddr`: `String` - (optional) The address of the SCS consensus protocol, left undefined for global contract-creation transaction, required for direct contract-creation transaction
2. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`String` - The 32 Bytes transaction hash as HEX string.

If the transaction was a contract creation use [chain3.mc.getTransactionReceipt()](#chain3mcgettransactionreceipt) to get the contract address, after the transaction was mined.

##### Example

```js

// compiled solidity source code using https://chriseth.github.io/cpp-ethereum/
var code = "603d80600c6000396000f3007c01000000000000000000000000000000000000000000000000000000006000350463c6888fa18114602d57005b600760043502
8060005260206000f3";

chain3.mc.sendTransaction({data: code}, function(err, address) {
  if (!err)
    console.log(address); // "0x7f9fade1c0d57a7af66ab4ead7c2eb7b11a91385"
});
```

***

<h4 id="chain3mccall">chain3.mc.call </h4>


    chain3.mc.call(callObject [, defaultBlock] [, callback])

Executes a message call transaction, which is directly executed in the VM of the node, but never mined into the blockchain.

##### Parameters

1. `Object` - A transaction object see [chain3.mc.sendTransaction](#chain3mcsendtransaction), with the difference that for calls the `from` property is optional as well.
2. `Number|String` - (optional) If you pass this parameter it will not use the default block set with [chain3.mc.defaultBlock](#chain3mcdefaultblock).
3. `Function` - (optional) If you pass a callback the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`String` - The returned data of the call, e.g. a codes functions return value.

##### Example

```js
var result = chain3.mc.call({
    to: "0xc4abd0339eb8d57087278718986382264244252f", 
    data: "0xc6888fa10000000000000000000000000000000000000000000000000000000000000003"
});
console.log(result); // "0x0000000000000000000000000000000000000000000000000000000000000015"
```

***

<h4 id="chain3mcestimategas">chain3.mc.estimateGas </h4>


    chain3.mc.estimateGas(callObject [, defaultBlock] [, callback])

Executes a message call or transaction, which is directly executed in the VM of the node, but never mined into the blockchain and returns the amount of the gas used.

##### Parameters

See [chain3.mc.sendTransaction](#chain3mcsendtransaction), expect that all properties are optional.

##### Returns

`Number` - the used gas for the simulated call/transaction.

##### Example

```js
var result = chain3.mc.estimateGas({
    to: "0xc4abd0339eb8d57087278718986382264244252f", 
    data: "0xc6888fa10000000000000000000000000000000000000000000000000000000000000003"
});
console.log(result); // "20465"
```

***

<h4 id="chain3mcfilter">chain3.mc.filter </h4>


```js
// can be 'latest' or 'pending'
var filter = chain3.mc.filter(filterString);
// OR object are log filter options
var filter = chain3.mc.filter(options);

// watch for changes
filter.watch(function(error, result){
  if (!error)
    console.log(result);
});

// Additionally you can start watching right away, by passing a callback:
chain3.mc.filter(options, function(error, result){
  if (!error)
    console.log(result);
});
```

##### Parameters

1. `String|Object` - The string `"latest"` or `"pending"` to watch for changes in the latest block or pending transactions respectively. Or a filter options object as follows:
  * `fromBlock`: `Number|String` - The number of the earliest block (`latest` may be given to mean the most recent and `pending` currently mining, block). By default `latest`.
  * `toBlock`: `Number|String` - The number of the latest block (`latest` may be given to mean the most recent and `pending` currently mining, block). By default `latest`.
  * `address`: `String` - An address or a list of addresses to only get logs from particular account(s).
  * `topics`: `Array of Strings` - An array of values which must each appear in the log entries. The order is important, if you want to leave topics out use `null`, e.g. `[null, '0x00...']`. You can also pass another array for each topic with options for that topic e.g. `[null, ['option1', 'option2']]`

##### Returns

`Object` - A filter object with the following methods:

  * `filter.get(callback)`: Returns all of the log entries that fit the filter.
  * `filter.watch(callback)`: Watches for state changes that fit the filter and calls the callback. See [this note](#using-callbacks) for details.
  * `filter.stopWatching()`: Stops the watch and uninstalls the filter in the node. Should always be called once it is done.

##### Watch callback return value

- `String` - When using the `"latest"` parameter, it returns the block hash of the last incoming block.
- `String` - When using the `"pending"` parameter, it returns a transaction hash of the last add pending transaction.
- `Object` - When using manual filter options, it returns a log object as follows:
    - `logIndex`: `Number` - integer of the log index position in the block. `null` when its pending log.
    - `transactionIndex`: `Number` - integer of the transactions index position log was created from. `null` when its pending log.
    - `transactionHash`: `String`, 32 Bytes - hash of the transactions this log was created from. `null` when its pending log.
    - `blockHash`: `String`, 32 Bytes - hash of the block where this log was in. `null` when its pending. `null` when its pending log.
    - `blockNumber`: `Number` - the block number where this log was in. `null` when its pending. `null` when its pending log.
    - `address`: `String`, 32 Bytes - address from which this log originated.
    - `data`: `String` - contains one or more 32 Bytes non-indexed arguments of the log.
    - `topics`: `Array of Strings` - Array of 0 to 4 32 Bytes `DATA` of indexed log arguments. (In *solidity*: The first topic is the *hash* of the signature of the event (e.g. `Deposit(address,bytes32,uint256)`), except you declared the event with the `anonymous` specifier.)

**Note** For event filter return values see [Contract Events](#contract-events)

##### Example

```js
var filter = chain3.mc.filter('pending');

filter.watch(function (error, log) {
  console.log(log); //  {"address":"0x0000000000000000000000000000000000000000", "data":"0x0000000000000000000000000000000000000000000000000000000000000000", ...}
});

// get all past logs again.
var myResults = filter.get(function(error, logs){ ... });

...

// stops and uninstalls the filter
filter.stopWatching();

```

***

<h4 id="chain3mccontract">chain3.mc.contract </h4>


    chain3.mc.contract(abiArray)

Creates a contract object for a solidity contract, which can be used to initiate contracts on an address.
You can read more about events [here](https://github.com/LBRChain/wiki/wiki/LBR-Contract-ABI#example-javascript-usage).

##### Parameters

1. `Array` - ABI array with descriptions of functions and events of the contract.

##### Returns

`Object` - A contract object, which can be initiated as follows:

```js
var MyContract = chain3.mc.contract(abiArray);

// instantiate by address
var contractInstance = MyContract.at([address]);

// deploy new contract
var contractInstance = MyContract.new([contructorParam1] [, contructorParam2], {data: '0x12345...', from: myAccount, gas: 1000000});

// Get the data to deploy the contract manually
var contractData = MyContract.new.getData([contructorParam1] [, contructorParam2], {data: '0x12345...'});
// contractData = '0x12345643213456000000000023434234'
```

And then you can either initiate an existing contract on an address,
or deploy the contract using the compiled byte code:

```js
// Instantiate from an existing address:
var myContractInstance = MyContract.at(myContractAddress);


// Or deploy a new contract:

// Deploy the contract asyncronous:
var myContractReturned = MyContract.new(param1, param2, {
   data: myContractCode,
   gas: 300000,
   from: mySenderAddress}, function(err, myContract){
    if(!err) {
       // NOTE: The callback will fire twice!
       // Once the contract has the transactionHash property set and once its deployed on an address.

       // e.g. check tx hash on the first call (transaction send)
       if(!myContract.address) {
           console.log(myContract.transactionHash) // The hash of the transaction, which deploys the contract
       
       // check address on the second call (contract deployed)
       } else {
           console.log(myContract.address) // the contract address
       }

       // Note that the returned "myContractReturned" === "myContract",
       // so the returned "myContractReturned" object will also get the address set.
    }
  });

// Deploy contract syncronous: The address will be added as soon as the contract is mined.
// Additionally you can watch the transaction by using the "transactionHash" property
var myContractInstance = MyContract.new(param1, param2, {data: myContractCode, gas: 300000, from: mySenderAddress});
myContractInstance.transactionHash // The hash of the transaction, which created the contract
myContractInstance.address // undefined at start, but will be auto-filled later
```

**Note** When you deploy a new contract, you should check for the next 12 blocks or so if the contract code is still at the address (using [chain3.mc.getCode()](#chain3mcgetcode)), to make sure a fork didn't change that.

##### Example

```js
// contract abi
var abi = [{
     name: 'myConstantMethod',
     type: 'function',
     constant: true,
     inputs: [{ name: 'a', type: 'string' }],
     outputs: [{name: 'd', type: 'string' }]
}, {
     name: 'myStateChangingMethod',
     type: 'function',
     constant: false,
     inputs: [{ name: 'a', type: 'string' }, { name: 'b', type: 'int' }],
     outputs: []
}, {
     name: 'myEvent',
     type: 'event',
     inputs: [{name: 'a', type: 'int', indexed: true},{name: 'b', type: 'bool', indexed: false]
}];

// creation of contract object
var MyContract = chain3.mc.contract(abi);

// initiate contract for an address
var myContractInstance = MyContract.at('0xc4abd0339eb8d57087278718986382264244252f');

// call constant function
var result = myContractInstance.myConstantMethod('myParam');
console.log(result) // '0x25434534534'

// send a transaction to a function
myContractInstance.myStateChangingMethod('someParam1', 23, {value: 200, gas: 2000});

// short hand style
chain3.mc.contract(abi).at(address).myAwesomeMethod(...);

// create filter
var filter = myContractInstance.myEvent({a: 5}, function (error, result) {
  if (!error)
    console.log(result);
    /*
    {
        address: '0x8718986382264244252fc4abd0339eb8d5708727',
        topics: "0x12345678901234567890123456789012", "0x0000000000000000000000000000000000000000000000000000000000000005",
        data: "0x0000000000000000000000000000000000000000000000000000000000000001",
        ...
    }
    */
});
```

***

#### Contract Methods

```js
// Automatically determines the use of call or sendTransaction based on the method type
myContractInstance.myMethod(param1 [, param2, ...] [, transactionObject] [, callback]);

// Explicitly calling this method
myContractInstance.myMethod.call(param1 [, param2, ...] [, transactionObject] [, callback]);

// Explicitly sending a transaction to this method
myContractInstance.myMethod.sendTransaction(param1 [, param2, ...] [, transactionObject] [, callback]);

// Explicitly sending a transaction to this method
myContractInstance.myMethod.sendTransaction(param1 [, param2, ...] [, transactionObject] [, callback]);

// Get the call data, so you can call the contract through some other means
var myCallData = myContractInstance.myMethod.getData(param1 [, param2, ...]);
// myCallData = '0x45ff3ff6000000000004545345345345..'
```

The contract object exposes the contracts methods, which can be called using parameters and a transaction object.

##### Parameters

- `String|Number` - (optional) Zero or more parameters of the function.
- `Object` - (optional) The (previous) last parameter can be a transaction object, see [chain3.mc.sendTransaction](#chain3mcsendtransaction) parameter 1 for more.
- `Function` - (optional) If you pass a callback as the last parameter the HTTP request is made asynchronous. See [this note](#using-callbacks) for details.

##### Returns

`String` - If its a call the result data, if its a send transaction a created contract address, or the transaction hash, see [chain3.mc.sendTransaction](#chain3mcsendtransaction) for details.


##### Example

```js
// creation of contract object
var MyContract = chain3.mc.contract(abi);

// initiate contract for an address
var myContractInstance = MyContract.at('0x78e97bcc5b5dd9ed228fed7a4887c0d7287344a9');

var result = myContractInstance.myConstantMethod('myParam');
console.log(result) // '0x25434534534'

myContractInstance.myStateChangingMethod('someParam1', 23, {value: 200, gas: 2000}, function(err, result){ ... });
```

***


#### Contract Events

```js
var event = myContractInstance.MyEvent({valueA: 23} [, additionalFilterObject])

// watch for changes
event.watch(function(error, result){
  if (!error)
    console.log(result);
});

// Or pass a callback to start watching immediately
var event = myContractInstance.MyEvent([{valueA: 23}] [, additionalFilterObject] , function(error, result){
  if (!error)
    console.log(result);
});

```

You can use events like [filters](#chain3mcfilter) and they have the same methods, but you pass different objects to create the event filter.

##### Parameters

1. `Object` - Indexed return values you want to filter the logs by, e.g. `{'valueA': 1, 'valueB': [myFirstAddress, mySecondAddress]}`. By default all filter values are set to `null`. It means, that they will match any event of given type sent from this contract.
2. `Object` - Additional filter options, see [filters](#chain3mcfilter) parameter 1 for more. By default filterObject has field 'address' set to address of the contract. Also first topic is the signature of event.
3. `Function` - (optional) If you pass a callback as the last parameter it will immediately start watching and you don't need to call `myEvent.watch(function(){})`. See [this note](#using-callbacks) for details.

##### Callback return


`Object` - An event object as follows:

- `args`: `Object` - The arguments coming from the event.
- `event`: `String` - The event name.
- `logIndex`: `Number` - integer of the log index position in the block.
- `transactionIndex`: `Number` - integer of the transactions index position log was created from.
- `transactionHash`: `String`, 32 Bytes - hash of the transactions this log was created from.
- `address`: `String`, 32 Bytes - address from which this log originated.
- `blockHash`: `String`, 32 Bytes - hash of the block where this log was in. `null` when its pending.
- `blockNumber`: `Number` - the block number where this log was in. `null` when its pending.


##### Example

```js
var MyContract = chain3.mc.contract(abi);
var myContractInstance = MyContract.at('0x78e97bcc5b5dd9ed228fed7a4887c0d7287344a9');

// watch for an event with {some: 'args'}
var myEvent = myContractInstance.MyEvent({some: 'args'}, {fromBlock: 0, toBlock: 'latest'});
myEvent.watch(function(error, result){
   ...
});

// would get all past logs again.
var myResults = myEvent.get(function(error, logs){ ... });

...

// would stop and uninstall the filter
myEvent.stopWatching();
```

***

#### Contract allEvents

```js
var events = myContractInstance.allEvents([additionalFilterObject]);

// watch for changes
events.watch(function(error, event){
  if (!error)
    console.log(event);
});

// Or pass a callback to start watching immediately
var events = myContractInstance.allEvents([additionalFilterObject,] function(error, log){
  if (!error)
    console.log(log);
});

```

Will call the callback for all events which are created by this contract.

##### Parameters

1. `Object` - Additional filter options, see [filters](#chain3mcfilter) parameter 1 for more. By default filterObject has field 'address' set to address of the contract. Also first topic is the signature of event.
2. `Function` - (optional) If you pass a callback as the last parameter it will immediately start watching and you don't need to call `myEvent.watch(function(){})`. See [this note](#using-callbacks) for details.

##### Callback return


`Object` - See [Contract Events](#contract-events) for more.

##### Example

```js
var MyContract = chain3.mc.contract(abi);
var myContractInstance = MyContract.at('0x78e97bcc5b5dd9ed228fed7a4887c0d7287344a9');

// watch for an event with {some: 'args'}
var events = myContractInstance.allEvents({fromBlock: 0, toBlock: 'latest'});
events.watch(function(error, result){
   ...
});

// would get all past logs again.
events.get(function(error, logs){ ... });

...

// would stop and uninstall the filter
myEvent.stopWatching();
```

***

<h4 id="chain3encodeParams">chain3.encodeParams </h4>


    chain3.encodeParams

Encode a list of parameters array into HEX codes.
##### Parameters

- `types` - list of types of params
- `params ` - list of values of params

##### Example

```js
var Chain3 = require('../index.js');
var chain3 = new Chain3();


// Test with list of parameters
var types = ['int','string'];
var args = [100, '4000'];

var dataHex = '0x' + chain3.encodeParams(types, args);
console.log("encoded params:", dataHex);

// outputs
encoded params:0x0000000000000000000000000000000000000000000000000000000000000064000000000000000000000000000000000000000000000000000000000000004000000000000000000000000000000000000000000000000000000000000000043430303000000000000000000000000000000000000000000000000000000000
```


***

<h4 id="chain3mcnamereg">chain3.mc.namereg </h4>


    chain3.mc.namereg

Returns GlobalRegistrar object.

##### Usage

see [namereg](https://github.com/LBRChain/chain3.js/blob/master/example/namereg.html) example

***

<h4 id="chain3mcsendibantransaction">chain3.mc.sendIBANTransaction </h4>


```js
var txHash = chain3.mc.sendIBANTransaction('0x00c5496aee77c1ba1f0854206a26dda82a81d6d8', 'XE81ETHXREGGAVOFYORK', 0x100);
```

Sends IBAN transaction from user account to destination IBAN address.

##### Parameters

- `string` - address from which we want to send transaction
- `string` - IBAN address to which we want to send transaction
- `value` - value that we want to send in IBAN transaction

***

<h4 id="chain3mciban">chain3.mc.iban </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
```

***

<h4 id="chain3mcibanfromaddress">chain3.mc.iban.fromAddress </h4>


```js
var i = chain3.mc.iban.fromAddress('0x00c5496aee77c1ba1f0854206a26dda82a81d6d8');
console.log(i.toString()); // 'XE7338O073KYGTWWZN0F2WZ0R8PX5ZPPZS
```

***

<h4 id="chain3mcibanfrombban">chain3.mc.iban.fromBban </h4>


```js
var i = chain3.mc.iban.fromBban('ETHXREGGAVOFYORK');
console.log(i.toString()); // "XE81ETHXREGGAVOFYORK"
```

***

<h4 id="chain3mcibancreateindirect">chain3.mc.iban.createIndirect </h4>


```js
var i = chain3.mc.iban.createIndirect({
  institution: "XREG",
  identifier: "GAVOFYORK"
});
console.log(i.toString()); // "XE81ETHXREGGAVOFYORK"
```

***

<h4 id="chain3mcibanisvalid">chain3.mc.iban.isValid </h4>


```js
var valid = chain3.mc.iban.isValid("XE81ETHXREGGAVOFYORK");
console.log(valid); // true

var valid2 = chain3.mc.iban.isValid("XE82ETHXREGGAVOFYORK");
console.log(valid2); // false, cause checksum is incorrect

var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var valid3 = i.isValid();
console.log(valid3); // true

```

***

<h4 id="chain3mcibanisdirect">chain3.mc.iban.isDirect </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var direct = i.isDirect();
console.log(direct); // false
```

***

<h4 id="chain3mcibanisindirect">chain3.mc.iban.isIndirect </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var indirect = i.isIndirect();
console.log(indirect); // true
```

***

<h4 id="chain3mcibanchecksum">chain3.mc.iban.checksum </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var checksum = i.checksum();
console.log(checksum); // "81"
```

***

<h4 id="chain3mcibaninstitution">chain3.mc.iban.institution </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var institution = i.institution();
console.log(institution); // 'XREG'
```

***

<h4 id="chain3mcibanclient">chain3.mc.iban.client </h4>


```js
var i = new chain3.mc.iban("XE81ETHXREGGAVOFYORK");
var client = i.client();
console.log(client); // 'GAVOFYORK'
```

***

<h4 id="chain3mcibanaddress">chain3.mc.iban.address </h4>


```js
var i = new chain3.mc.iban('XE7338O073KYGTWWZN0F2WZ0R8PX5ZPPZS');
var address = i.address();
console.log(address); // '00c5496aee77c1ba1f0854206a26dda82a81d6d8'
```

***

<h4 id="chain3mcibantostring">chain3.mc.iban.toString </h4>


```js
var i = new chain3.mc.iban('XE7338O073KYGTWWZN0F2WZ0R8PX5ZPPZS');
console.log(i.toString()); // 'XE7338O073KYGTWWZN0F2WZ0R8PX5ZPPZS'
```

***

<h4 id="chain3admin_addPeer">chain3.admin.addPeer </h4>

Add a Peer into the current network. Require the VNODE open admin API.
The addPeer administrative method requests adding a new remote node to the list of tracked static nodes. The node will try to maintain connectivity to these nodes at all times, reconnecting every once in a while if the remote connection goes down.

The method accepts a single argument, the enode URL of the remote peer to start tracking and returns a BOOL indicating whether the peer was accepted for tracking or some error occurred.

```js
console.log("admin addpeer", chain3.admin.addPeer("enode://9f562d54e0ec6764514592615780838bfe051f1930696c86917013c6304ea92ba4f1371fdedf886da38238d79dc8fe62318b16d85f926351079a815d27a064b4@144.168.43.133:30333")); // true

```

***

<h4 id="chain3admin_datadir">chain3.admin.datadir </h4>

The datadir administrative property can be queried for the absolute path the running VNODE currently uses to store all its databases.

```js
console.log("admin datadir", chain3.admin.datadir);

//admin datadir /Users/admin/go/src/github.com/testnet

```

***

<h4 id="chain3admin_nodeInfo">chain3.admin.nodeInfo </h4>

The nodeInfo administrative property can be queried for all the information known about the running VNODE at the networking granularity. These include general information about the node itself as a participant of the P2P overlay protocol, as well as specialized information added by each of the running application protocols (e.g. mc, les, shh, bzz).

```js
console.log("admin nodeInfo", chain3.admin.nodeInfo);

//admin nodeInfo { id: 'a6f486af99679e00ec1a2bf77304e8c7f183987c8138a7c515a08ee42c5bebbda9f01474d43ba7176f891989dbdb78a6cbade67a941e6c2d5a83751039adba36',
  // name: 'LBR/v1.0.10-rc-89f6ab9c/darwin-amd64/go1.10',
  // enode: 'enode://a6f486af99679e00ec1a2bf77304e8c7f183987c8138a7c515a08ee42c5bebbda9f01474d43ba7176f891989dbdb78a6cbade67a941e6c2d5a83751039adba36:30336?servicecfgport=:50062&showtopublic=true&beneficialaddress=&ip=',
  // ip: '71.***.***.***',
  // ports: { discovery: 30336, listener: 30336 },
  // listenAddr: '[::]:30336',
  // protocols: 
  //  { mc: 
  //     { network: 106,
  //       difficulty: 1476985363965,
  //       genesis: '0x4e2972df43453f5b658656de1f2af40866b6d86b4e11b0c49eb1fc1a854d9796',
  //       head: '0x941bb3a1c9a8a26e0bc2f747c2a7ea805135e1e995464957aa7e09814a1575d7' } } }

```

***

<h4 id="chain3admin_peers">chain3.admin.peers </h4>

The peers administrative property can be queried for all the information known about the connected remote nodes at the networking granularity. These include general information about the nodes themselves as participants of the P2P overlay protocol, as well as specialized information added by each of the running application protocols (e.g. eth, les, shh, bzz).

```js
console.log("admin peers", chain3.admin.peers);

//admin peers [ { id: '089554d6929600b9c70bbd6e1c12594697d0aec43127b9b29c6eb96faf06884fd284f56c3de64155d65e540b59a43f4fd07d8802b4b5e95b2922531e6096c2d5',
  //   name: 'LBR/v1.0.9-rc-c5e47f69/linux-amd64/go1.11',
  //   caps: [ 'mc/62', 'mc/63' ],
  //   network: 
  //    { localAddress: '192.168.1.169:55681',
  //      remoteAddress: '52.15.143.41:30333' },
  //   protocols: { mc: [Object] } },
  // { id: '271c55ef39be9208e6ad75c935061412b39e51dd97a8e4dbba7d358e91132fd7c79ee687228edea3fd9c833b6ce9c365365aa526999956914d7ac81d00576e76',
  //   name: 'LBR/v1.0.9-rc-c5e47f69/linux-amd64/go1.11',
  //   caps: [ 'mc/62', 'mc/63' ],
  //   network: 
  //    { localAddress: '192.168.1.169:55680',
  //      remoteAddress: '18.217.180.94:30333' },
  //   protocols: { mc: [Object] } } ]

```

***

<h4 id="chain3admin_stopRPC">chain3.admin.stopRPC </h4>

The stopRPC administrative method closes the currently open HTTP RPC endpoint. As the node can only have a single HTTP endpoint running, this method takes no parameters, returning a boolean whether the endpoint was closed or not.

```js
console.log("admin stopRPC", chain3.admin.stopRPC());//true

```

***

<h4 id="chain3admin_startRPC">chain3.admin.startRPC </h4>

The startRPC administrative method starts an HTTP based JSON RPC API webserver to handle client requests. All the parameters are optional:

- host: network interface to open the listener socket on (defaults to "localhost")
- port: network port to open the listener socket on (defaults to 8545)
- cors: cross-origin resource sharing header to use (defaults to "")
- apis: API modules to offer over this interface (defaults to "mc,net,admin")

The method returns a boolean flag specifying whether the HTTP RPC listener was opened or not. Please note, only one HTTP endpoint is allowed to be active at any time.

Though this is a admin method in the console, it's not good to use it through HTTP RPC.
Since this method cannot be used when RPC port is not open and it cannot be used to override existing port.

```js
console.log("admin startRPC", chain3.admin.startRPC("127.0.0.1", 8545));

```

***

<h4 id="chain3txpool_content">chain3.txpool.content </h4>

The content inspection property can be queried to list the exact details of all the transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only.

The result is an object with two fields pending and queued. Each of these fields are associative arrays, in which each entry maps an origin-address to a batch of scheduled transactions. These batches themselves are maps associating nonces with actual transactions.

Please note, there may be multiple transactions associated with the same account and nonce. This can happen if the user broadcast mutliple ones with varying gas allowances (or even complerely different transactions).

```js
console.log("txpool content", chain3.txpool.content);

```

***

<h4 id="chain3txpool_status">chain3.txpool.status </h4>

The status inspection property can be queried for the number of transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only.

The result is an object with two fields pending and queued, each of which is a counter representing the number of transactions in that particular state.

```js
console.log("txpool status", chain3.txpool.status);

// {
//   pending: 10,
//   queued: 7
// }

```

***

<h4 id="chain3txpool_inspect">chain3.txpool.inspect </h4>

The inspect inspection property can be queried to list a textual summary of all the transactions currently pending for inclusion in the next block(s), as well as the ones that are being scheduled for future execution only. This is a method specifically tailored to developers to quickly see the transactions in the pool and find any potential issues.

The result is an object with two fields pending and queued. Each of these fields are associative arrays, in which each entry maps an origin-address to a batch of scheduled transactions. These batches themselves are maps associating nonces with transactions summary strings.

Please note, there may be multiple transactions associated with the same account and nonce. This can happen if the user broadcast mutliple ones with varying gas allowances (or even complerely different transactions).

```js
console.log("txpool inspect", chain3.txpool.inspect);

{
  pending: {
    0x26588a9301b0428d95e6fc3a5024fce8bec12d51: {
      31813: ["0x3375ee30428b2a71c428afa5e89e427905f95f7e: 0 sha + 500000 × 20000000000 gas"]
    },......
  },
  queued: {
    ......
  }
}


```

***

<h4 id="chain3vnode_address">chain3.vnode.address </h4>


```js
chain3.setProvider(new chain3.providers.HttpProvider('http://localhost:8545'));
console.log("VNODE:", chain3.vnode.address);
```

***

<h4 id="vnode_scsservice">chain3.vnode.scsService </h4>


```js
console.log("VNODE service:", chain3.vnode.scsService);
```

***

<h4 id="vnode_servicecfg">chain3.vnode.servicecfg </h4>


```js
console.log("VNODE servicecfg:", chain3.vnode.servicecfg);
```


***

<h4 id="vnode_showtopublic">chain3.vnode.showToPublic </h4>


```js
console.log("VNODE showToPublic:", chain3.vnode.showToPublic);
```


***

<h4 id="vnode_vnodeip">chain3.vnode.ip </h4>


```js
console.log("VNODE IP:", chain3.vnode.ip);
```

***

<h4 id="scs_directcall">chain3.scs.directCall </h4>


```js
console.log("Get MicroChain constant call:", chain3.scs.directCall(tx));
```

***

<h4 id="scs_getblock">chain3.scs.getBlock </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
mlist = chain3.scs.getMicroChainList();
console.log("SCS MicroChain List:", mlist);
mcAddress=mlist[0];
console.log("block 1:", chain3.scs.getBlock(mcAddress, '0x1'));
```

***

<h4 id="scs_getblocklist">chain3.scs.getBlockList </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
mlist = chain3.scs.getMicroChainList();
console.log("SCS MicroChain List:", mlist);
mcAddress=mlist[0];
console.log("block 1:", chain3.scs.getBlock(mcAddress, '0x1', '0x5'));
```

***

<h4 id="scs_getblocknumber">chain3.scs.getBlockNumber</h4>


```js
console.log("MicroChain block number:", chain3.scs.getBlockNumber('0xECd1e094Ee13d0B47b72F5c940C17bD0c7630326'));
```

***

<h4 id="scs_getdappstate">chain3.scs.getDappState</h4>


```js
console.log("MicroChain status:", chain3.scs.getDappState('0xECd1e094Ee13d0B47b72F5c940C17bD0c7630326'));
```

***

<h4 id="scs_getmicrochainlist">chain3.scs.getMicroChainList </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
mlist = chain3.scs.getMicroChainList();
console.log("SCS MicroChain List:", mlist);
```

***

<h4 id="scs_getmicrochaininfo">chain3.scs.getMicroChainInfo </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
minfo = chain3.scs.getMicroChainInfo('0xECd1e094Ee13d0B47b72F5c940C17bD0c7630326');
console.log("SCS MicroChain Info:", minfo);
```


***

<h4 id="scs_getnonce">chain3.scs.getNonce </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
console.log("Account Nonce on MicroChain:", chain3.scs.getNonce('0xECd1e094Ee13d0B47b72F5c940C17bD0c7630326', '0x7d0cba876cb9da5fa310a54d29f4687f5dd93fd7'));
```


***

<h4 id="scs_getscsid">chain3.scs.getScsid </h4>


```js
chain3.setScsProvider(new chain3.providers.HttpProvider('http://localhost:8548'));
console.log("SCS ID:", chain3.scs.getScsid());
```


***

<h4 id="scs_getReceiptByHash">chain3.scs.getReceiptByHash </h4>

Returns the receipt of a transaction by transaction hash. Note That the
receipt is not available for pending transactions.

##### Parameters

1. ``String`` - The MicroChain address. 
2. ``String`` - The transaction hash.

##### Callback return


``Object`` - A transaction receipt object, or ``null`` when no receipt
was found:

-  ``transactionHash``: ``DATA``, 32 Bytes - hash of the transaction.
-  ``transactionIndex``: ``QUANTITY`` - integer of the transactions
   index position in the block.
-  ``blockHash``: ``DATA``, 32 Bytes - hash of the block where this
   transaction was in.
-  ``blockNumber``: ``QUANTITY`` - block number where this transaction
   was in.

-  ``contractAddress``: ``DATA``, 20 Bytes - The contract address
   created, if the transaction was a contract creation, otherwise
   ``null``.
-  ``logs``: ``Array`` - Array of log objects, which this transaction
   generated.
-  ``logsBloom``: ``DATA``, 256 Bytes - Bloom filter for light clients
   to quickly retrieve related logs.
- ``failed``: ``Boolean`` - ``true`` if the filter was successfully uninstalled,
otherwise ``false``.
-  ``status``: ``QUANTITY`` either ``1`` (success) or ``0`` (failure)


##### Example


```js

    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    txhash1="0x688456221f7729f5c2c17006bbe4df163d09bea70c1a1ebb66b9b53ca10563df";
    console.log("TX Receipt:", chain3.scs.getReceiptByHash(mcAddress, txhash1));

    // Result
    {
      "id":101,
      "jsonrpc": "2.0",
      "result": {contractAddress: '0x0a674edac2ccd47ae2a3197ea3163aa81087fbd1',
  failed: false,"logs":[{"address":"0x2328537bc943ab1a89fe94a4b562ee7a7b013634","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a8863fc8ce3816411378685223c03daae9770ebb","0x0000000000000000000000007312f4b8a4457a36827f185325fd6b66a3f8bb8b"],"data":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGQ=","blockNumber":0,"transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69","transactionIndex":0,"blockHash":"0x78f092ca81a891ad6c467caa2881d00d8e19c8925ddfd71d793294fbfc5f15fe","logIndex":0,"removed":false}],"logsBloom":"0x00000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000008000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000800000000000080000000000000000000000000002000000000000000000000000000000000000080100002000000000000000000000000000000000000000000000000000000000000000000000000000","status":"0x1","transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69"}
    }
```

***

<h4 id="scs_getReceiptByNonce">chain3.scs.getReceiptByNonce </h4>

Returns the transaction result by address and nonce on the MicroChain. Note That the nonce is the nonce on the MicroChain. This nonce can be checked using scs_getNonce. 

##### Parameters


1. ``String`` - The MicroChain address. 
1. ``String`` - The transaction nonce.
1. ``QUANTITY`` - The nonce of the transaction.

##### Returns

``Object`` - A transaction receipt object, or null when no receipt was
found:.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    tAddress="0xf6a36118751c50f8932d31d6d092b11cc28f2258";
    console.log("SCS receipt:", chain3.scs.getReceiptByNonce(mcAddress, tAddress, 0));

    // Result
    SCS receipt: {contractAddress: '0x0a674edac2ccd47ae2a3197ea3163aa81087fbd1',
  failed: false,"logs":[{"address":"0x2328537bc943ab1a89fe94a4b562ee7a7b013634","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a8863fc8ce3816411378685223c03daae9770ebb","0x0000000000000000000000007312f4b8a4457a36827f185325fd6b66a3f8bb8b"],"data":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGQ=","blockNumber":0,"transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69","transactionIndex":0,"blockHash":"0x78f092ca81a891ad6c467caa2881d00d8e19c8925ddfd71d793294fbfc5f15fe","logIndex":0,"removed":false}],"logsBloom":"0x00000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000008000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000800000000000080000000000000000000000000002000000000000000000000000000000000000080100002000000000000000000000000000000000000000000000000000000000000000000000000000","status":"0x1","transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69"}
```

***

<h4 id="scs_getReceiptByHash">chain3.scs.getReceiptByHash </h4>

Returns the receipt of a transaction by transaction hash. Note That the
receipt is not available for pending transactions.

##### Parameters

1. ``String`` - The AppChain/MicroChain address. 
2. ``String`` - The transaction hash.

##### Returns


``Object`` - A transaction object, or null when no transaction was found.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    txhash1="0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69";
    console.log("TX by hash:", chain3.scs.getTransactionByHash(mcAddress, txhash1));

    // Result
    TX by hash: {
      "id":101,
      "jsonrpc": "2.0",
      "result": {contractAddress: '0x0a674edac2ccd47ae2a3197ea3163aa81087fbd1',
  failed: false,"logs":[{"address":"0x2328537bc943ab1a89fe94a4b562ee7a7b013634","topics":["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef","0x000000000000000000000000a8863fc8ce3816411378685223c03daae9770ebb","0x0000000000000000000000007312f4b8a4457a36827f185325fd6b66a3f8bb8b"],"data":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAGQ=","blockNumber":0,"transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69","transactionIndex":0,"blockHash":"0x78f092ca81a891ad6c467caa2881d00d8e19c8925ddfd71d793294fbfc5f15fe","logIndex":0,"removed":false}],"logsBloom":"0x00000000000000000000000000000000080000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008000000000008000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000000000000400000000000000000000000000000800000000000080000000000000000000000000002000000000000000000000000000000000000080100002000000000000000000000000000000000000000000000000000000000000000000000000000","status":"0x1","transactionHash":"0x67bfaa5a704e77a31d5e7eb866f8c662fa8313a7882d13d0d23e377cd66d2a69"}
    }
```

***

<h4 id="scs_getTransactionByNonce">chain3.scs.getTransactionByNonce </h4>

Returns the receipt of a transaction by transaction hash. Note That the
receipt is not available for pending transactions.

##### Parameters

1. ``String`` - The MicroChain address. 
2. ``String`` - The transaction nonce.
3. ``QUANTITY`` - The nonce of the transaction.

##### Returns


``Object`` - A transaction receipt object, or null when no receipt was
found:.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    tAddress="0xf6a36118751c50f8932d31d6d092b11cc28f2258";
    console.log("SCS TX:", chain3.scs.getTransactionByNonce(mcAddress, tAddress, 0));


    // Result
    SCS TX: { blockHash: '0x45ab47bde3a7caa62d80e8c38bef21ada499d52331e574f3a09d4d943aa133fa',
  blockNumber: 66,
  from: '0xf6a36118751c50f8932d31d6d092b11cc28f2258', input: '.....', nonce: 0,
  r: 1.1336589614028917e+77,
  s: 1.8585853533200337e+76,
  shardingFlag: 3,
  to: '0x25b0102b5826efa7ac469782f54f40ffa72154f5',
  transactionHash: '0x6eb3d33fab53317007927368238aef5bc00d1d1d9bf082930c372e3dabca507c',
  transactionIndex: 0,
  v: 248,
  value: BigNumber { s: 1, e: 21, c: [ 10000000 ] },
  gas: 0,
  gasPrice: BigNumber { s: 1, e: 0, c: [ 0 ] } }
```

***

<h4 id="scs_getExchangeByAddress">chain3.scs.getExchangeByAddress </h4>

Returns the Withdraw/Deposit exchange records between MicroChain and MotherChain for a certain address. This command returns both the ongoing exchanges and processed exchanges. To check all the ongoing exchanges, please use scs_getExchangeInfo. 

##### Parameters

1. `String` - The MicroChain address.
1. `String` - The address to be checked.
1. `Int` - Index of Deposit records >= 0.
1. `Int` - Number of Deposit records extracted.
1. `Int` - Index of Depositing records >= 0.
1. `Int` - Number of Depositing records extracted.
1. `Int` - Index of Withdraw records >= 0.
1. `Int` - Number of Withdraw records extracted.
1. `Int` - Index of Withdrawing records >= 0.
1. `Int` - Number of Withdrawing records extracted.

##### Returns


``Object`` - A JSON format object contains the token exchange info.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    tAddress="0xf6a36118751c50f8932d31d6d092b11cc28f2258";
    console.log("SCS token address exchange:", chain3.scs.getExchangeByAddress(mcAddress, tAddress));

    // Result
    SCS token address exchange: { DepositRecordCount: 0,
    DepositRecords: null,
    DepositingRecordCount: 0,
    DepositingRecords: null,
    WithdrawRecordCount: 0,
    WithdrawRecords: null,
    WithdrawingRecordCount: 0,
    WithdrawingRecords: null,
    microchain: '0x25b0102b5826efa7ac469782f54f40ffa72154f5',
    sender: '0xf6a36118751c50f8932d31d6d092b11cc28f2258' }
```

***

<h4 id="scs_getExchangeInfo">chain3.scs.getExchangeInfo </h4>

Returns the Withdraw/Deposit exchange records between MicroChain and MotherChain for a certain address. This command returns both the ongoing exchanges and processed exchanges. To check all the ongoing exchanges, please use scs_getExchangeInfo. 

##### Parameters

1. `String` - The MicroChain address.
1. `String` - The transaction hash.
1. `Int` - Index of Depositing records >= 0.
1. `Int` - Number of Depositing records extracted.
1. `Int` - Index of Withdrawing records >= 0.
1. `Int` - Number of Withdrawing records extracted.

##### Returns


``Object`` - A JSON object contains the token exchange info.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    console.log("SCS token exchanging info:", chain3.scs.getExchangeInfo(mcAddress));

    // Result
    SCS token exchanging info: { DepositingRecordCount: 0,
      DepositingRecords: null,
      WithdrawingRecordCount: 0,
      WithdrawingRecords: null,
      microchain: '0x25b0102b5826efa7ac469782f54f40ffa72154f5',
      scsid: '0xecd1e094ee13d0b47b72f5c940c17bd0c7630326' }
```

***

<h4 id="scs_getTxpool">chain3.scs.getTxpool </h4>

Returns the ongoing transactions in the MicroChain. 

##### Parameters

1. `String` - The MicroChain address.

##### Returns

``Object`` - A JSON format object contains two fields pending and queued. Each of these fields are associative arrays, in which each entry maps an origin-address to a batch of scheduled transactions. These batches themselves are maps associating nonces with actual transactions.

##### Example

```js
    // Request
    var mclist = chain3.scs.getMicroChainList(); //find the MicroChain on the SCS
    mcAddress = mclist[0]; //locate the 1st MicroChain
    console.log("SCS TXpool:", chain3.scs.getTxpool(mcAddress));

    // Result

    SCS TXpool: {"pending":{},"queued":{}}
```

***

<h4 id="scs_getReceiptByHash">chain3.scs.getReceiptByHash </h4>