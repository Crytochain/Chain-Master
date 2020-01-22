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
 * @file microfunction.js
 * @LBR tech 
 * @date 2018
 * @provide access to LBR microChain dapps functions and variables.
 * @as a replacement for function.js 
 */

var coder = require('../solidity/coder');
var utils = require('../utils/utils');
var errors = require('../chain3/errors');
var formatters = require('../chain3/formatters');
var sha3 = require('../utils/sha3');
var config = require('../utils/config');

/**
 * This prototype should be used to call/sendTransaction to solidity functions
 * for DAPPs.
 * 2019/03/01, nuwa 1.0.8, to support multiple Dapps, need to link this to dappbase.
 * uses mc to send non-constant to vnode
 * uses scs to send constant function to SCS and get returns.
 * Parameters:
 * mc - Object handles the communication with VNODE
 * scs - Object handles the comm
 */
var MicroFunction = function (mc, scs, json, mcAddress, address, via) {
    this._mc = mc;
    this._scs = scs;
    this._inputTypes = json.inputs.map(function (i) {
        return i.type;
    });
    this._outputTypes = json.outputs.map(function (i) {
        return i.type;
    });
    this._constant = json.constant;

    this._payable = json.payable;
    this._name = utils.transformToFullName(json);
    this._mcAddress = mcAddress;//MicroChain Address
    this._address = address;//Dapp Address
    this._from = mc.coinbase;//default set to vnode account, can be changed
    this._via = via;//need to be set before sendTransaction call

    // console.log("MicroFunction new: ",this._name, " constant: ", json.constant, this._constant);
};

MicroFunction.prototype.extractCallback = function (args) {
    if (utils.isFunction(args[args.length - 1])) {
        return args.pop(); // modify the args array!
    }
};

MicroFunction.prototype.extractDefaultBlock = function (args) {
    if (args.length > this._inputTypes.length && !utils.isObject(args[args.length -1])) {
        return formatters.inputDefaultBlockNumberFormatter(args.pop()); // modify the args array!
    }
};

/**
 * Should be used if connect with a remote VNODE server.
 * User need to setup a source account instead of default mc coinbase.
 * Unlike coinbase, the SCS mining rewards go to the scs beneficial
 * as set in the userconfig.json:
 * "Beneficiary": "0xd814f2ac2c4ca49b33066582e4e97ebae02f2ab9",
 * @method setSrcAccount
 * @param {ADDRESS} inAcct
 * @throws {Error} if input is not a valid address
 */

MicroFunction.prototype.setSrcAccount = function (inAcct){
    if (utils.isAddress(inAcct)){
        this._from = inAcct;
        return true;
    }
    throw errors.InvalidPublicAddress();
    return false;
} 

/**
 * Should be set with the benefit address from the VNODE server.
 *
 * @method setVnodeVia
 * @param {ADDRESS} inAcct
 * @throws {Error} if it is not
 */

MicroFunction.prototype.setVnodeAddress = function (inAcct){
    if (utils.isAddress(inAcct)){
        this._via = inAcct;
        return true;
    }
    throw errors.InvalidPublicAddress();
    return false;
} 

/**
 * Should be called to check if the number of arguments is correct
 *
 * @method validateArgs
 * @param {Array} arguments
 * @throws {Error} if it is not
 */
MicroFunction.prototype.validateArgs = function (args) {
    var inputArgs = args.filter(function (a) {
      // filter the options object but not arguments that are arrays
      return !( (utils.isObject(a) === true) &&
                (utils.isArray(a) === false) &&
                (utils.isBigNumber(a) === false)
              );
    });
    
    if (inputArgs.length !== this._inputTypes.length) {
        throw errors.InvalidNumberOfSolidityArgs();
    }
};

/**
 * Should be used to create payload from arguments
 * For microchain DAPPs, need to have both microchain address
 * 
 * @method toPayload
 * @param {Array} solidity function params
 * @param {Object} optional payload options
 */
MicroFunction.prototype.toPayload = function (args) {
    var options = {};
    if (args.length > this._inputTypes.length && utils.isObject(args[args.length -1])) {
        options = args[args.length - 1];
    }
    this.validateArgs(args);
    // Need to use MicroChain Address instead of Dapp address
    options.to = this._mcAddress;
    // options.data = this._address + this.signature() + coder.encodeParams(this._inputTypes, args);
    options.dappAddr = this._address;
    // Original
    options.data = '0x' + this.signature() + coder.encodeParams(this._inputTypes, args);
    return options;
};

/**
 * Should be used to get function signature
 *
 * @method signature
 * @return {String} function signature
 */
MicroFunction.prototype.signature = function () {
    return sha3(this._name).slice(0, 8);
};


MicroFunction.prototype.unpackOutput = function (output) {
    if (!output) {
        return;
    }

    output = output.length >= 2 ? output.slice(2) : output;
    var result = coder.decodeParams(this._outputTypes, output);
    return result.length === 1 ? result[0] : result;
};

/**
 * Calls a Solidity constant function by using arguments
 * and send to the SCS server, 
 * For multiple Contract DAPP support, nuwa1.0.8
 * Need to add the dapp address before the payload call
 * @method call
 * @param {...Object} Contract function arguments
 * @param {function} If the last argument is a function, the contract function
 *   call will be asynchronous, and the callback will be passed the
 *   error and result.
 * @return {String} output bytes
 */
MicroFunction.prototype.call = function () {
    var args = Array.prototype.slice.call(arguments).filter(function (a) {return a !== undefined; });
    var callback = this.extractCallback(args);
    var defaultBlock = this.extractDefaultBlock(args);
    var payload = this.toPayload(args);

    //For direct call send to SCS, put in an empty from account
    //This won't change any state on the MicroChain.
    if ( payload.from == undefined){
        payload.from = this._from || config.defaultSCSId;
    }
    
    console.log("MicroFunction.prototype.call\n");
    console.log(payload);
    console.log("===================================");
 

    if (!callback) {
        var output = this._scs.directCall(payload, defaultBlock);
        return this.unpackOutput(output);
    }

    //Asyn callback
    var self = this;
    this._scs.directCall(payload, defaultBlock, function (error, output) {
        if (error) return callback(error, null);

        var unpacked = null;
        try {
            unpacked = self.unpackOutput(output);
        }
        catch (e) {
            error = e;
        }

        callback(error, unpacked);
    });
};

/**
 * Call the Solidity non-constant function
 * to VNODE. 
 * Note, this need to setup as Direct Call,
 * add sharding flag and via through the payload
 * @method sendTransaction
 */
MicroFunction.prototype.sendTransaction = function () {
    var args = Array.prototype.slice.call(arguments).filter(function (a) {return a !== undefined; });
    var callback = this.extractCallback(args);
    var payload = this.toPayload(args);

    if (payload.value > 0 && !this._payable) {
        throw new Error('Cannot send value to non-payable function');
    }


    srcnonce = this._scs.getNonce(payload.to,this._from);

    //For DAPP function, send as direct call
    if (payload.shardingFlag != 1)
    {
        payload.shardingFlag = 1   
    }

    if (utils.isAddress(this._from)){
        payload.from = this._from;
    }else{
        throw new Error('sendTransaction has no valid src account');
    }
    
    //Note, this via is from dapp, 
    if (utils.isAddress(this._via)){
        payload.via = this._via;
    }else{
        throw new Error('sendTransaction has no valid vnode via account');
    }
    
    //Gas need set to gas estimate
    payload.gas = 4000000;// use this as default

    payload.nonce = srcnonce;

    //Make sure the args has shardingFlag = 1
    //and via is set to the VNODE via address
    if (!callback) {
        return this._mc.sendTransaction(payload);
    }

    this._mc.sendTransaction(payload, callback);
};

/**
 * Should be used to estimateGas of Dapp function
 * Work with Nuwa 1.0.4 and later.
 *
 * @method estimateGas
 */
MicroFunction.prototype.estimateGas = function () {
    var args = Array.prototype.slice.call(arguments);
    var callback = this.extractCallback(args);
    var payload = this.toPayload(args);

    if (!callback) {
        return this._mc.estimateGas(payload);
    }

    this._mc.estimateGas(payload, callback);
};

/**
 * Return the encoded data of the call
 *
 * @method getData
 * @return {String} the encoded data
 */
MicroFunction.prototype.getData = function () {
    var args = Array.prototype.slice.call(arguments);
    var payload = this.toPayload(args);

    return payload.data;
};

/**
 * Should be used to get function display name
 *
 * @method displayName
 * @return {String} display name of the function
 */
MicroFunction.prototype.displayName = function () {
    return utils.extractDisplayName(this._name);
};

/**
 * Should be used to get function type name
 *
 * @method typeName
 * @return {String} type name of the function
 */
MicroFunction.prototype.typeName = function () {
    return utils.extractTypeName(this._name);
};

/**
 * Should be called to get rpc requests from solidity function
 *
 * @method request
 * @returns {Object}
 */
MicroFunction.prototype.request = function () {
    var args = Array.prototype.slice.call(arguments);
    var callback = this.extractCallback(args);
    var payload = this.toPayload(args);
    var format = this.unpackOutput.bind(this);

    //if costant function, use SCS call
    //otherwise use mc sendTransaction
    return {
        method: this._constant ? 'scs_call' : 'mc_sendTransaction',
        callback: callback,
        params: [payload],
        format: format
    };
};

/**
 * Should be called to execute function
 * Note only perform the Dappbase
 * @method execute
 */
MicroFunction.prototype.execute = function () {

    var transaction = !this._constant;

    // send transaction for non-constant function
    // to the VNODE
    if (transaction) {
        return this.sendTransaction.apply(this, Array.prototype.slice.call(arguments));
    }

    // call for constant function or variable
    // to SCS monitor
    return this.call.apply(this, Array.prototype.slice.call(arguments));
};

/**
 * Should be called to attach function to contract
 * from attachToContract
 * @method attachToMicroChain
 * @param {Contract}
 */
MicroFunction.prototype.attachToMicroChain = function (contract) {
    var execute = this.execute.bind(this);
    execute.request = this.request.bind(this);
    execute.call = this.call.bind(this);
    execute.sendTransaction = this.sendTransaction.bind(this);
    execute.estimateGas = this.estimateGas.bind(this);
    execute.getData = this.getData.bind(this);
    var displayName = this.displayName();
    if (!contract[displayName]) {
        contract[displayName] = execute;
    }
    contract[displayName][this.typeName()] = execute; // circular!!!!
};

module.exports = MicroFunction;
