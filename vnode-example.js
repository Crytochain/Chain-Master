#!/usr/bin/env node

/*
 * Example program to use the Chain3 VNODE package
 * to get the  commands
 * 
*/

var Chain3 = require('../index.js');
var chain3 = new Chain3();

chain3.setProvider(new chain3.providers.HttpProvider('http://localhost:8545'));


if (!chain3.isConnected()){
    console.log("Chain3 RPC is not connected!");
    return;
}
console.log("VNODE:", chain3.vnode.ip);
console.log("VNODE:", chain3.vnode.address);
console.log("VNODE:", chain3.vnode.scsService);
console.log("VNODE:", chain3.vnode.showToPublic);
console.log("VNODE:", chain3.vnode.serviceCfg);

chain3.vnode.getAddress(function(err, res) {
        if (!err){
            
            console.log("Succeed!: ", res);
            return res;
        }else{
            console.log("Chain3 error:", err.message);
            // response.success = false;
            // response.error = err.message;
            return err.message;
        }
        console.log("Shouldn't come here")
});
return;

