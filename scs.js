/*
    This file is part of chain3.js.

    chain3.js is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    chain3.js is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with chain3.js.  If not, see <http://www.gnu.org/licenses/>.
*/
/** @file scs.js
 * process RPC commands to SCS monitors
 * @authors:
 * @LBR lab
 * @date 2018
 * Methods supported by SCSserver:
 * getScsId
 * getMicroChainList
 * getNonce
 * getBlockNumber
 * getBlock
 * getBalance
 * 2019/07/01
 * Added new RPC methods
     getNonce,
    getSCSId,
    getDatadir,
    getDappAddrList,
    protocolVersion,
    getDappState,
    getDappAddrList,
    getMicroChainList,
    getMicroChainInfo,
    getAppChainList,
    getAppChainInfo,   
    getBlockNumber,
    getBlock,
    getBlockList,
    getBalance,
    getTransactionByHash,
    getTransactionByNonce,
    getReceiptByHash,
    getReceiptByNonce,
    getExchangeInfo,
    getExchangeByAddress,
    directCall
    getTxpool
 * 
 */

var formatters = require('../formatters');
var Method = require('../method');
var utils = require('../../utils/utils');
var Property = require('../property');
// var Dapp = require('../dapp');
var config = require('../../utils/config'); //for MicroChain address

// SCS object
var Scs = function (chain3) {
    this._requestManager = chain3._scsRequestManager;

    var self = this;

    properties().forEach(function (p) {
        p.attachToObject(self);
        p.setRequestManager(chain3._scsRequestManager);
    });

    methods().forEach(function (method) {
        method.attachToObject(self);
        method.setRequestManager(chain3._scsRequestManager);
    });

};

// SCSID
// 
/// @returns an objects describing chain3.scs properties
var properties = function () {
    return [
        new Property({
            name: 'listening',
            getter: 'scs_listening'
        })
    ];
};

Object.defineProperty(Scs.prototype, 'defaultAddress', {
    get: function () {
        return config.defaultSCSId;
    },
    set: function (val) {
        config.defaultSCSId = val;
        return val;
    }
});

var methods = function () {

    // Return the nonce of the account in the MicroChain.
    var getNonce = new Method({
        name: 'getNonce',
        call: 'scs_getNonce',
        params: 2,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputAddressFormatter],
        outputFormatter: utils.toDecimal
    });

    // Return the SCS id, this is not the beneficial address, but the add to identify the SCS
    // in the LBR network.
    var getSCSId = new Method({
        name: 'getSCSId',
        call: 'scs_getSCSId',
        params: 0
    });

    //check the data dir of the SCS storage
    var getDatadir = new Method({
        name: 'getDatadir',
        call: 'scs_datadir',
        params: 0
    });

    // Display the Dapp list on the SCS
    // for use with v1.0.8 multi contract version
    // return multiple addresses

    var getDappAddrList = new Method({
        name: 'getDappAddrList',
        call: 'scs_getDappAddrList',
        params: 0
    });

    //display the protocol version info
    var protocolVersion = new Method({
        name: 'protocolVersion',
        call: 'scs_protocolVersion',
        params: 0
    });

    // Display the microChain list on the SCS
    var getMicroChainList = new Method({
        name: 'getMicroChainList',
        call: 'scs_getMicroChainList',
        params: 0
    });

    // Display the microChain list on the SCS
    var getMicroChainInfo = new Method({
        name: 'getMicroChainInfo',
        call: 'scs_getMicroChainInfo',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter]
    });

    // Display the appChain list on the SCS
    var getAppChainList = new Method({
        name: 'getAppChainList',
        call: 'scs_getMicroChainList',
        params: 0
    });

    // Display the microChain list on the SCS
    var getAppChainInfo = new Method({
        name: 'getAppChainInfo',
        call: 'scs_getMicroChainInfo',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter]
    });

    // call the DAPP function and return the data
    var getDappState = new Method({
        name: 'getDappState',
        call: 'scs_getDappState',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter]
    });

    // call the DAPP function and return the data
    var getDappAddrList = new Method({
        name: 'getDappAddrList',
        call: 'scs_getDappAddrList',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter]
    });

    // Get the block number of the MicroChain
    var getBlockNumber = new Method({
        name: 'getBlockNumber',
        call: 'scs_getBlockNumber',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter],
        outputFormatter: utils.toDecimal
    });

    // Get a single BLOCK info from the MicroChain.
    var getBlock = new Method({
        name: 'getBlock',
        call: 'scs_getBlock',
        params: 2,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputDefaultBlockNumberFormatter],
        outputFormatter: formatters.outputBlockFormatter
    });

    // Get multiple BLOCKs info from the MicroChain.
    var getBlockList = new Method({
        name: 'getBlockList',
        call: 'scs_getBlockList',
        params: 3,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputDefaultBlockNumberFormatter, formatters.inputDefaultBlockNumberFormatter]
    });

    // get the balance of the MicroChain token
    // for the account.
    // If the MicroChain is a non-token chain,
    // this always returns 0
    // Only return the balance in the lastest block
    var getBalance = new Method({
        name: 'getBalance',
        call: 'scs_getBalance',
        params: 2,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputAddressFormatter],
        outputFormatter: formatters.outputBigNumberFormatter
    });

    // call the DAPP function using input data 
    // This only returns the constant views
    // used with Dapp constructor
    var directCall = new Method({
        name: 'directCall',
        call: 'scs_directCall',
        params: 1,
        inputFormatter: [formatters.inputTransactionFormatter]
    });


    // Return the transaction info by 
    // MicroChainAddress
    // Account address
    // Account nonce (for this MicroChain)
    var getTransactionByNonce = new Method({
        name: 'getTransactionByNonce',
        call: 'scs_getTransactionByNonce',
        params: 3,
        outputFormatter: formatters.outputTransactionFormatter 
    });

    // Return the transaction info from the MicroChain
    // MicroChainAddress
    // Transaction Hash
    var getTransactionByHash = new Method({
        name: 'getTransactionByHash',
        call: 'scs_getTransactionByHash',
        params: 2,
        outputFormatter: formatters.outputTransactionFormatter
    });

    // Return the transaction Receipt 
    var getReceiptByHash = new Method({
        name: 'getReceiptByHash',
        call: 'scs_getReceiptByHash',
        params: 2,
        outputFormatter: formatters.outputTransactionFormatter
    });

    // Return the transaction info by 
    // MicroChainAddress
    // Account address
    // Account nonce (for this MicroChain)
    var getReceiptByNonce = new Method({
        name: 'getReceiptByNonce',
        call: 'scs_getReceiptByNonce',
        params: 3,
        outputFormatter: formatters.outputTransactionFormatter 
    });

    // Return the Exchange info between MotherChain and MicroChain
    // for a certain address
    var getExchangeByAddress = new Method({
        name: 'getExchangeByAddress',
        call: 'scs_getExchangeByAddress',
        params: 10,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputAddressFormatter,
        formatters.inputDefaultRecordIndexFormatter,formatters.inputDefaultRecordSizeFormatter, 
        formatters.inputDefaultRecordIndexFormatter,formatters.inputDefaultRecordIndexFormatter, 
        formatters.inputDefaultRecordIndexFormatter,formatters.inputDefaultRecordIndexFormatter, 
        formatters.inputDefaultRecordIndexFormatter,formatters.inputDefaultRecordSizeFormatter]
    });

    // Return the Exchange info between MotherChain and MicroChain
    var getExchangeInfo = new Method({
        name: 'getExchangeInfo',
        call: 'scs_getExchangeInfo',
        params: 5,
        inputFormatter: [formatters.inputAddressFormatter, formatters.inputDefaultRecordIndexFormatter,
        formatters.inputDefaultRecordSizeFormatter, formatters.inputDefaultRecordIndexFormatter, 
        formatters.inputDefaultRecordSizeFormatter]
    });

    // Return the TXPOOL info on the MicroChain
    var getTxpool = new Method({
        name: 'getTxpool',
        call: 'scs_getTxpool',
        params: 1,
        inputFormatter: [formatters.inputAddressFormatter]
    });

    return [
    getNonce,
    getSCSId,
    getDatadir,
    getDappAddrList,
    protocolVersion,
    getDappState,
    getDappAddrList,
    getMicroChainList,
    getMicroChainInfo,
    getAppChainList,
    getAppChainInfo,   
    getBlockNumber,
    getBlock,
    getBlockList,
    getBalance,
    getTransactionByHash,
    getTransactionByNonce,
    getReceiptByHash,
    getReceiptByNonce,
    getExchangeInfo,
    getExchangeByAddress,
    getTxpool,
    directCall
    ];
}

/*
 * Init the MicroChain Dapp 
 */
// Scs.prototype.dapp = function (abi) {
//     var factory = new Dapp(this, abi);
//     return factory;
// };

//TODO, not working, need to add watches.scs package
// Scs.prototype.filter = function (options, callback, filterCreationErrorCallback) {
//     return new Filter(options, 'scs', this._requestManager, watches.mc(), formatters.outputLogFormatter, callback, filterCreationErrorCallback);
// };

module.exports = Scs;