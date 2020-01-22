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
 * @file dapp.js
 * @LBR tech 
 * @date 2018
 * @provide access to LBR microChain dapps constant functions and variables.
 * @as a replacement for contract.js 
 */

var utils = require('../utils/utils');
var coder = require('../solidity/coder');
var SolidityEvent = require('./event');
var DappFunction = require('./dappfunction');
var AllEvents = require('./dappallevents');

/**
 * Should be called to encode constructor params
 *
 * @method encodeConstructorParams
 * @param {Array} abi
 * @param {Array} constructor params
 */
var encodeConstructorParams = function (abi, params) {
    return abi.filter(function (json) {
        return json.type === 'constructor' && json.inputs.length === params.length;
    }).map(function (json) {
        return json.inputs.map(function (input) {
            return input.type;
        });
    }).map(function (types) {
        return coder.encodeParams(types, params);
    })[0] || '';
};

/**
 * Should be called to add functions to Dapp object
 *
 * @method addDappFunctions
 * @param {Contract} contract
 * @param {Array} abi
 */
var addDappFunctions = function (mcdapp) {
    // contract.abi.filter(function (json) {
    mcdapp.abi.filter(function (json) {    
        return json.type === 'function';
    }).map(function (json) {
        return new DappFunction(mcdapp._mc, mcdapp._scs, json, mcdapp.address, mcdapp.via);
    }).forEach(function (f) {
        f.attachToDapp(mcdapp);
    });
};

/**
 * Should be called to add events to Dapp object
 *
 * @method addDappEvents
 * @param {Contract} contract
 * @param {Array} abi
 */
var addDappEvents = function (contract) {
    var events = contract.abi.filter(function (json) {
        return json.type === 'event';
    });

    var All = new AllEvents(contract._scs._requestManager, events, contract.address);
    All.attachToContract(contract);

    events.map(function (json) {
        return new SolidityEvent(contract._scs._requestManager, json, contract.address);
    }).forEach(function (e) {
        e.attachToContract(contract);
    });
};


/**
 * Should be called to check if the Dapp gets properly deployed.
 * This requires both VNODE and SCS monitor are connecting.
 * on the MicroChain.
 *
 * @method checkForMicroChainDapp
 * @param {Object} contract
 * @param {Function} callback
 * @returns {Undefined}
 */
var checkForMicroChainDapp = function(dapp, callback){

    var count = 0,
    callbackFired = false;

    // wait for receipt
    var filter = dapp._mc.filter('latest', function(e){
        if (!e && !callbackFired) {
            count++;

            // stop watching after 50 blocks (timeout)
            if (count > 50) {

                filter.stopWatching(function() {});
                callbackFired = true;

                if (callback)
                    callback(new Error('Contract transaction couldn\'t be found after 50 blocks'));
                else
                    throw new Error('Contract transaction couldn\'t be found after 50 blocks');


            } else {

                dapp._mc.getTransactionReceipt(dapp.transactionHash, function(e, receipt){
                    if(receipt && !callbackFired) {

                        dapp._scs.getCode(receipt.contractAddress, function(e, code){
                            /*jshint maxcomplexity: 6 */

                            if(callbackFired || !code)
                                return;

                            filter.stopWatching(function() {});
                            callbackFired = true;

                            if(code.length > 3) {

                                dapp.address = receipt.contractAddress;

                                // attach events and methods again after we have
                                addFunctionsToContract(dapp);
                                addEventsToContract(dapp);

                                // call callback for the second time
                                if(callback)
                                    callback(null, dapp);

                            } else {
                                if(callback)
                                    callback(new Error('The dapp code couldn\'t be stored, please check your gas amount.'));
                                else
                                    throw new Error('The dapp code couldn\'t be stored, please check your gas amount.');
                            }
                        });
                    }
                });
            }
        }
    });
};

/**
 * Should be called to deploy new Dapp instance
 * on a existing MicroChain
 * This is similar to the ContractFactory
 * object for mc, but only keeps the call, 
 * no sendTransaction.
 * 
 * @method MicroChainDapp
 * @param {Array} abi
 */
var MicroChainDapp = function (mc, scs, abi) {
    this._mc = mc;
    this._scs = scs;
    this.abi = abi;
    this.via = null;

    /**
     * Should be called to create new Dapp on the MicroChain blockchain
     * by sending the transaction to VNODE with shardingFlag = 0,
     * Note, the DAPP requires the microChain address as input
     * to deploy.
     * 
     * @method new
     * @param {inAddress} MicroChain Address to deploy the Dapp
     * @param {viaAddress} VNODE Address to connect for deploying the Dapp
     * @param {Any} Dapp contract constructor param1 (optional)
     * @param {Any} Dapp contract constructor param2 (optional)
     * @param {Object} contract transaction object (required)
     * @param {Function} callback
     * @returns {Dapp} returns Dapp instance
     */
    this.new = function () {
        
        // parse arguments
        var options = {}; // required!
        var callback;

        //Processing the arguments
        var args = Array.prototype.slice.call(arguments);

        if (args.length < 2){
            throw new Error('DAPP init requires MicroChain address and VNODE address as the first two input arguments!');
        }   

        inAddress = args[0];
        if (!utils.isAddress(inAddress)){
        throw new Error('No MicroChain address in the input arguments!');
        }
        // Remove the MicroChain address (1st element)
        args.shift();

        viaAddress = args[0];
        if (!utils.isAddress(viaAddress)){
        throw new Error('No VNODE address in the input arguments!');
        }
        // Start building the Dapp on the MicroChain
        var mcDapp = new Dapp(this._mc, this._scs, this.abi, inAddress, viaAddress);
        
        if ( this._scs.getDappState(inAddress) != 0 ){
            throw new Error('MicroChain is not ready or already has DAPP on it');
            return;
        }

        // Remove the VNODE address (2nd element)
        args.shift();
        
        // Remove the callback func (last element)
        if (utils.isFunction(args[args.length - 1])) {
            callback = args.pop();
        }

        var last = args[args.length - 1];
        if (utils.isObject(last) && !utils.isArray(last)) {
            options = args.pop();
        }

        if (options.value > 0) {
            var constructorAbi = abi.filter(function (json) {
                return json.type === 'constructor' && json.inputs.length === args.length;
            })[0] || {};

            if (!constructorAbi.payable) {
                throw new Error('Cannot send value to non-payable constructor');
            }
        }

        // Added the function input parameters from args

        var bytes = encodeConstructorParams(this.abi, args);
        options.data += bytes;

        //Dapp is not a contract, need to setup more fields
        // DAPP deploy need send to the microChain address
        options.to = inAddress;
        options.gas = 0;
        options.shardingFlag = 1;
        options.nonce = 0;//Should check, usually for a mciroChain without DAPP, all accounts nonce is 0
        options.via = viaAddress;


        //All transaction should send to the microChain through VNODE with shardingFlag = 1
        if (callback) {

            // wait for the mcDapp address adn check if the code was deployed
            this._mc.sendTransaction(options, function (err, hash) {
                if (err) {
                    callback(err);
                } else {
                    // add the transaction hash
                    mcDapp.transactionHash = hash;

                    // call callback for the first time
                    callback(null, mcDapp);

                    checkForMicroChainDapp(mcDapp, callback);
                }
            });
        } else {
            var hash = this._mc.sendTransaction(options);
            // add the transaction hash
            mcDapp.transactionHash = hash;
            checkForMicroChainDapp(mcDapp);
        }

        return mcDapp;
    };

    this.new.getData = this.getData.bind(this);
};


/**
 * Should be called to get access to existing Dapp on the MicroChain
 *
 * @method at
 * @param {Address} MicroChain address (required)
 * @param {Function} callback {optional)
 * @returns {mcDapp} returns mcirochain Dapp if no callback was passed,
 * otherwise calls callback function (err, mcDapp)
 * Possible errors:
 * 
 */
MicroChainDapp.prototype.at = function (address, callback) {
    var mcDapp = new Dapp(this._mc, this._scs, this.abi, address, this.via);

    // this functions are not part of prototype,
    // because we dont want to spoil the interface
    addDappFunctions(mcDapp);
    addDappEvents(mcDapp);

    if (callback) {
        callback(null, mcDapp);
    }
    return mcDapp;
};

/**
 * Gets the data, which is data to deploy plus constructor params
 *
 * @method getData
 */
MicroChainDapp.prototype.getData = function () {
    var options = {}; // required!
    var args = Array.prototype.slice.call(arguments);

    var last = args[args.length - 1];
    if (utils.isObject(last) && !utils.isArray(last)) {
        options = args.pop();
    }

    var bytes = encodeConstructorParams(this.abi, args);
    options.data += bytes;

    return options.data;
};

/**
 * Set the benefit address of the VNODE server.
 *
 * @method setVnodeAddress
 */
MicroChainDapp.prototype.setVnodeAddress = function (viaAddress) {
    
    if (utils.isAddress(viaAddress)){
        this.via = viaAddress;
        return true;
    }
    throw errors.InvalidPublicAddress();
    return false;
};

/**
 * Return the address of the VNODE server.
 *
 * @method getVnodeAddress
 */
MicroChainDapp.prototype.getVnodeAddress = function () {
    
    return this.via;
};

/**
 * Should be called to create new Dapp on the MicroChaib
 * contract instance
 *
 * @method Contract
 * @param {Array} abi
 * @param {Address} contract address
 */
var Dapp = function (mc, scs, abi, address, viaAddress) {
    this._mc = mc; //vnode server to sendTransaction
    this._scs = scs;  //SCS server to sendCall
    this.transactionHash = null;// This is needed to check if the Dapp is deploy or not, note this is different from MicroChain HASH
    this.address = address; //MicroChain address
    this.abi = abi;
    this.via = viaAddress;
};

module.exports = MicroChainDapp;
