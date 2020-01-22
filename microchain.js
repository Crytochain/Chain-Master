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
 * @file microchain.js
 * @LBR tech 
 * @date 2019
 * @provide access to LBR microChain dappBase constant to deploy the 
 *  dapp contractsfunctions and variables.
 * @ Used with LBR v1.0.9 with multiChain support 
 * constructor()
 * 
 * 
 * newContract(): 
 * 1. deploy the DAPPs on the MicroChain using sendtx；
 * 2. uses transactionReceipt to save the Dapps address；
 * 3. register the Dapp address to the DappBase；
 */

var utils = require('../utils/utils');
var coder = require('../solidity/coder');
var MicroEvent = require('./microevent');
var MicroFunction = require('./microfunction');
var AllEvents = require('./microallevents');
var DappBase = require('./microbase');
var ABIs = require('./microconstants');
var SolidityEvent = require('../chain3/event');
var SolidityFunction = require('../chain3/function');

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
 * For MIcroChain support multiple DAPPs,
 * need to input both MicroChain Address 
 * and Dapp address
 *
 * @method addDappFunctions
 * @param {Contract} mcdapp
 * @param {Array} abi
 */
var addDappFunctions = function (mcdapp) {
    // input object with methods
    // Processed the abi functions
    mcdapp.abi.filter(function (json) {    
        return json.type === 'function';
    }).map(function (json) {
        return new MicroFunction(mcdapp._mc, mcdapp._scs, json, mcdapp.address, mcdapp.dappAddress, mcdapp.via);
    }).forEach(function (f) {
        f.attachToMicroChain(mcdapp);
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
        return new MicroEvent(contract._scs._requestManager, json, contract.address);
    }).forEach(function (e) {
        e.attachToContract(contract);
    });
};

/**
 * Should be called to add functions to contract object
 * for Mother chain only.
 * @method addFunctionsToContract
 * @param {Contract} contract
 * @param {Array} abi
 */
var addFunctionsToContract = function (contract) {
    // console.log("addFunctionsToContract:", contract);
    // contract.abi.filter(function (json) {
    contract.abi.filter(function (json) {    
        return json.type === 'function';
    }).map(function (json) {
        return new SolidityFunction(contract._mc, json, contract.address);
    }).forEach(function (f) {
        f.attachToContract(contract);
    });
};

/**
 * Should be called to add events to contract object
 *
 * @method addEventsToContract
 * @param {Contract} contract
 * @param {Array} abi
 */
var addEventsToContract = function (contract) {
    var events = contract.abi.filter(function (json) {
        return json.type === 'event';
    });

    var All = new AllEvents(contract._mc._requestManager, events, contract.address);
    All.attachToContract(contract);

    events.map(function (json) {
        return new SolidityEvent(contract._mc._requestManager, json, contract.address);
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
 * object for mc, but keep both the VNODE to send
 * the transaction and the SCS for constant call.
 * 
 * @method MicroChain
 * @param {Object} mc
 * @param {Object} scs
 * @param {Array} abi
 */
var MicroChain = function (mc, scs, inAbi) {
    this._mc = mc;
    this._scs = scs;
    // this.mcType = inName;
    this.via = null;
    // this.baseabi = JSON.parse(ABIs.dappBaseABI);

    // Set the MicroChain base ABI according to the input type
    // need to parse the abi string into object
    // if (inName == 'AST'){
    //     this.abi = JSON.parse(ABIs.astABI);
    // }else if( inName == 'ASM'){
    //     //Used the MicroChain with 
    //     this.abi = JSON.parse(ABIs.asmABI);
        
    // }else{
    //     // console.log("unsupported MicroChain type:", inName);
    //     throw new Error('unsupported MicroChain type');
    // }

    this.abi = inAbi;

    /**
     * Should be called to create new DappBase on the MicroChain 
     * by sending the transaction to VNODE with shardingFlag = 3,
     * This is different from the previous single contract version,
     * which requires shardingFlag = 1.
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
     ////Not workable 2019/06/07
    this.new = function () {
        
        // parse arguments
        var options = {}; // required!
        var callback;     // optional

        //Processing the arguments
        var args = Array.prototype.slice.call(arguments);

        if (args.length < 2){
            throw new Error('DAPP init requires MicroChain address and VNODE address as the first two input arguments!');
        }   


        // Check for the MicroChain address
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
        // if ( )
        // var mcDapp = new Dapp(this._mc, this._scs, this.abi, inAddress, viaAddress);
        var mcBase = new DappBase.MicrochainBase(this._mc, this.abi, address, this.via);
        
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

        // Dapp is not a contract, need to setup more fields
        // DAPP deploy need send to the microChain address
        options.to = inAddress;
        options.gas = 0;
        options.shardingFlag = 1;//shardingFlag = 3 for multiple contract
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
 * Should be called to operate the MicroChain
 * This is a Global contract, and use 
 * @method at
 * @param {Address} MicroChain address (required)
 * @param {Function} callback {optional)
 * @returns {mcDapp} returns mcirochain Dapp if no callback was passed,
 * otherwise calls callback function (err, mcDapp)
 * Possible errors:
 * 
 */
MicroChain.prototype.at = function (address, callback) {


    // Generate the MicroChain object
    var mcBase = new DappBase.MicrochainBase(this._mc, this.abi, address, this.via);

    // this functions are not part of prototype,
    // because we dont want to spoil the interface
    addFunctionsToContract(mcBase);
    addEventsToContract(mcBase);

    if (callback) {
        callback(null, mcBase);
    }
    return mcBase;
};

/**
 * Should be called to get access to existing DappBase
 * on the MicroChain. Need to check if the current
 * Microchain has DappBase deployed by using getDappState.
 * 
 * @method getDapp
 * @param {Address} MicroChain address (required)
 * @param {Function} callback {optional)
 * @returns {mcDapp} returns mcirochain Dapp if no callback was passed,
 * otherwise calls callback function (err, mcDapp)
 * Possible errors:
 * 
 */
MicroChain.prototype.getDapp = function (inMcAddress, inDAPPabi, inDappAddress, callback) {

    // Create a DAPP object with input MicroChain Address, and DAPP abi AND address
    var mcDapp = new DappBase.DappBase(this._mc, this._scs, inMcAddress, inDAPPabi, inDappAddress, this.via);

    // ADD THE functions and events to the Object
    addDappFunctions(mcDapp);
    addDappEvents(mcDapp);

    if (callback) {
        callback(null, mcDapp);
    }
    return mcDapp;
};

// MicroChain.prototype.getDapp = function (inAbi, inAddress, callback) {

// // console.log("MicroChain.prototype.at:", this.mcType);
//    var dappstate = this._scs.getDappState(inAddress)
//    if (  dappstate != 1 ){
//      throw new Error('MicroChain has no Dapp deployed! getDappState != 1');
//      return null;
//    }

//   dappAddrList =  this._scs.getDappList();
//   console.log("DappBase addr:",dappAddrList);
//   if (dappAddrList.length > 0){
//     var dappBaseAddress = dappAddrList[0];
//     // console.log("SCS TX receipt:", chain3.scs.getReceiptByNonce(mcAddress, coinbase, 1));
//     // Check if the name is set
//     var mcBase = new DappBase.DappBase(this._mc, this._scs, inAbi, inAddress, this.via);

//     // console.log("MicrochainBase", mcBase.abi);
//     console.log("=============getDappBase=================");

//     // this functions are not part of prototype,
//     // because we dont want to spoil the interface
//     addDappFunctions(mcBase);
//     addDappEvents(mcBase);
//   }


//     if (callback) {
//         callback(null, mcBase);
//     }
//     return mcBase;
// };

/**
 * Gets the data, which is data to deploy plus constructor params
 *
 * @method getData
 */
MicroChain.prototype.getData = function () {
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
 * Set the benefit address of the VNODE proxy server.
 *
 * @method setVnodeAddress
 */
MicroChain.prototype.setVnodeAddress = function (viaAddress) {
    
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
MicroChain.prototype.getVnodeAddress = function () {
    
    return this.via;
};

/**
 * Register a new contract on the MicroChain so 
 * user can call it through MicroChain methods
 *
 * @method registerContract
 */
MicroChain.prototype.registerContract = function (inAddress) {
   
   if (utils.isAddress(inAddress)){
        this.dappAddress = inAddress;
        return true;
    }
    throw errors.InvalidPublicAddress();
    return false;
};

/**
 * Register a new contract on the MicroChain so 
 * user can call it through MicroChain methods
 * The same with registerContract.
 * @method registerDapp
 */
MicroChain.prototype.registerDapp = function (inAddress) {
   
   if (utils.isAddress(inAddress)){
        this.dappAddress = inAddress;
        return true;
    }
    throw errors.InvalidPublicAddress();
    return false;
};

/**
 * Should be called to create new Dapp on the MicroChain
 * contract instance, 
 * moved to microbase.js
 *
 * @method Contract
 * @param {Array} abi
 * @param {Address} contract address
 */
// var Dapp = function (mc, scs, abi, address, viaAddress) {
//     this._mc = mc; //vnode server to sendTransaction
//     this._scs = scs;  //SCS server to sendCall
//     this.transactionHash = null;// This is needed to check if the Dapp is deploy or not, note this is different from MicroChain HASH
//     this.address = address; //MicroChain address
//     this.baseAddress = null;// dappbase address, this can to be extracted using deploy address and nonce 0 on the MicroChain. MicroChain address
//     this.dappAddress = null;// dapp address, this is 
//     this.abi = abi;
//     this.via = viaAddress;
// };

module.exports = MicroChain;
