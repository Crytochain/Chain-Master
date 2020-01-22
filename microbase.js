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
 * @file microbase.js
 * @LBR tech 
 * @date 2019
 * @provide access to LBR microChain dappBase constant to deploy the 
 *  dapp contractsfunctions and variables.
 * @ Used with LBR v1.0.8 with multiChain support 
*/

/**
 * For microChain only, 
 * @method Contract
 * @param {obj} mc, VNODE manager
 * @param {Array} abi, interface
 * @param {Address} microChain address
 */
var MicrochainBase = function (mc, inAbi, address) {
    this._mc = mc;
    this.transactionHash = null;
    this.address = address;
    this.abi = inAbi;
};

/**
 * Should be called to create a DappBase object on the 
 * MicroChain contract instance.
 *
 * @method DappBase
 * @param {obj} mc, VNODE manager
 * @param {obj} scs, SCS monitor
 * @param {Array} inAbi, ABI of the input contract
 * @param {Address} contract address
 */
var DappBase = function (mc, scs, inMcAddress, inAbi, inAddress, viaAddress) {
    this._mc = mc; //vnode server to sendTransaction
    this._scs = scs;  //SCS server to sendCall and get response from the Monitor
    this.transactionHash = null;// This is needed to check if the Dapp is deploy or not, note this is different from MicroChain HASH
    this.address = inMcAddress; //MicroChain address, need to use for every function calls
    this.baseAddress = null;// dappbase address, this need to be extracted using MicroChain address
    this.dappAddress = inAddress;// dapp address, need to be registered with dappbase, may not need 
    this.abi = inAbi;
    this.via = viaAddress;
};


/**
 * Should be called to create new Dapp 
 * contract instance, 
 * only support two internal abi structures
 * 
 * @method Contract
 * @param {Array} abi
 * @param {Address} contract address
 */
var MicroDapp = function (mc, scs, inAbi, inAddress, viaAddress) {

    this.abi = inAbi;

    this._mc = mc; //vnode server to sendTransaction
    this._scs = scs;  //SCS server to sendCall
    this.transactionHash = null;// This is needed to check if the Dapp is deploy or not, note this is different from MicroChain HASH
    this.address = inAddress; //MicroChain address
    this.baseAddress = null;// dappbase address, this need to be extracted using MicroChain address
    this.dappAddress = null;// dapp address, need to be registered with dappbase
    // this.baseAbi = ABIs.dappBaseABI;
    this.via = viaAddress;
};

// var data=subchainbase.buyMintToken.getData(value)
// This should call the subchainbase.
MicroDapp.prototype.buyMintToken = function (value, callback) {

    // console.log("MicrochainBase", mcBase.abi);
    console.log("=============buyMintToken=================");
    mcBase = null;
    // this functions are not part of prototype,
    // because we dont want to spoil the interface

    if (callback) {
        callback(null, mcBase);
    }
    return mcBase;
};

module.exports = {
    MicroDapp: MicroDapp,
    DappBase: DappBase,
    MicrochainBase:MicrochainBase
}