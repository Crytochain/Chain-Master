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
/**
 * @file chain3.js
 * @Modified from file web3.js
 * @authors:
 *   Jeffrey Wilcke <jeff@ethdev.com>
 *   Marek Kotewicz <marek@ethdev.com>
 *   Marian Oancea <marian@ethdev.com>
 *   Fabian Vogelsteller <fabian@ethdev.com>
 *   Gav Wood <g@ethdev.com>
 * @date 2014
 * @authors:
   @LBR tech
   @date 2018
 */

var RequestManager = require('./chain3/requestmanager');
var Iban = require('./chain3/iban');
var Mc = require('./chain3/methods/mc');
var Vnode = require('./chain3/methods/vnode');
var Scs = require('./chain3/methods/scs');
var Admin = require('./chain3/methods/admin');
var Txpool = require('./chain3/methods/txpool');
var Net = require('./chain3/methods/net');
var Personal = require('./chain3/methods/personal');
var Settings = require('./chain3/settings');
var version = require('./version.json');
var utils = require('./utils/utils');
var sha3 = require('./utils/sha3');
var extend = require('./chain3/extend');
var Batch = require('./chain3/batch');
var Property = require('./chain3/property');
var HttpProvider = require('./chain3/httpprovider');
var IpcProvider = require('./chain3/ipcprovider');
var BigNumber = require('bignumber.js');
var Coder = require('./solidity/coder');
var MicroChain = require('./microchain/microchain');
var sigutils = require('./accounts/sigUtils.js');
var McDapp = require('./chain3/dapp.js');
var Debug = require('./chain3/methods/debug');


function Chain3 (provider,scsProvider) {
    this._requestManager = new RequestManager(provider);
    this.currentProvider = provider;
    this._scsRequestManager = new RequestManager(scsProvider);
    this.scsCurrentProvider = scsProvider;
  
    this.mc = new Mc(this);
    this.admin = new Admin(this);
    this.txpool = new Txpool(this);
    this.vnode = new Vnode(this);
    this.scs = new Scs(this);
    this.net = new Net(this);
    this.personal = new Personal(this);
    this.debug = new Debug(this);
    this.settings = new Settings();

    this.version = {
        api: version.version
    };
    this.providers = {
        HttpProvider: HttpProvider,
        IpcProvider: IpcProvider
    };
    this._extend = extend(this);
    this._extend({
        properties: properties()
    });
}

// expose providers on the class
Chain3.providers = {
    HttpProvider: HttpProvider,
    IpcProvider: IpcProvider
};

Chain3.prototype.setProvider = function (provider) {
    this._requestManager.setProvider(provider);
    this.currentProvider = provider;
};

Chain3.prototype.setScsProvider = function (provider) {
    this._scsRequestManager.setProvider(provider);
    this.scsCurrentProvider = provider;
};

Chain3.prototype.reset = function (keepIsSyncing) {
    this._requestManager.reset(keepIsSyncing);
    this.settings = new Settings();
};

Chain3.prototype.BigNumber = BigNumber;
Chain3.prototype.toHex = utils.toHex;
Chain3.prototype.toAscii = utils.toAscii;
Chain3.prototype.toUtf8 = utils.toUtf8;
Chain3.prototype.fromAscii = utils.fromAscii;
Chain3.prototype.fromUtf8 = utils.fromUtf8;
Chain3.prototype.toDecimal = utils.toDecimal;
Chain3.prototype.fromDecimal = utils.fromDecimal;
Chain3.prototype.toBigNumber = utils.toBigNumber;
Chain3.prototype.toSha = utils.toSha;
Chain3.prototype.fromSha = utils.fromSha;
Chain3.prototype.isAddress = utils.isAddress;
Chain3.prototype.isChecksumAddress = utils.isChecksumAddress;
Chain3.prototype.toChecksumAddress = utils.toChecksumAddress;
Chain3.prototype.isIBAN            = utils.isIBAN;
Chain3.prototype.padLeft = utils.padLeft;
Chain3.prototype.padRight = utils.padRight;

//New functions to sign transaction
Chain3.prototype.intToHex = utils.BigIntToHex;

//Encode the input types and parameters
Chain3.prototype.encodeParams     = function(type, param) {
    return Coder.encodeParams(type, param);
};

Chain3.prototype.sha3 = function(string, options) {
    return '0x' + sha3(string, options);
};

/**
 * Transforms direct icap to address
 */
Chain3.prototype.fromICAP = function (icap) {
    var iban = new Iban(icap);
    return iban.address();
};

var properties = function () {
    return [
        new Property({
            name: 'version.node',
            getter: 'chain3_clientVersion'
        }),
        new Property({
            name: 'version.network',
            getter: 'net_version',
            inputFormatter: utils.toDecimal
        }),
        new Property({
            name: 'version.LBR',
            getter: 'mc_protocolVersion',
            inputFormatter: utils.toDecimal
        })
    ];
};

Chain3.prototype.isConnected = function(){
    return (this.currentProvider && this.currentProvider.isConnected());
};

Chain3.prototype.isScsConnected = function(){
    return (this.scsCurrentProvider && this.scsCurrentProvider.isConnected());
};

Chain3.prototype.createBatch = function () {
    return new Batch(this);
};

// //MicroChain constructor object for multi contract chain.
// TODO
Chain3.prototype.microchain = function (inabi, inAddress) {
    //Used the MicroChain with 
    var mcChain = new MicroChain(this.mc, this.scs, inabi);
    return mcChain;
};

//A collection of signing functions used for LBR chain
//Verify Signature function
//TX function
Chain3.prototype.signTransaction = sigutils.signTransaction;
Chain3.prototype.verifyMcSignature = sigutils.verifyMcSignature;
Chain3.prototype.signMcMessage = sigutils.signMcMessage;
Chain3.prototype.recoverPersonalSignature = sigutils.recoverPersonalSignature;

module.exports = Chain3;


