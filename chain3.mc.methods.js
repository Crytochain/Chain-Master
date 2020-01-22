var chai = require('chai');
var assert = chai.assert; 
var Chain3 = require('../index.js');
var chain3 = new Chain3();
var u = require('./helpers/test.utils.js');


describe('chain3.mc', function() {
    describe('methods', function() {
        u.methodExists(chain3.mc, 'getBalance');
        // u.methodExists(chain3.mc, 'getStorageAt');
        // u.methodExists(chain3.mc, 'getTransactionCount');
        // u.methodExists(chain3.mc, 'getCode');
        // u.methodExists(chain3.mc, 'sendTransaction');
        // u.methodExists(chain3.mc, 'call');
        // u.methodExists(chain3.mc, 'getBlock');
        // u.methodExists(chain3.mc, 'getTransaction');
        // u.methodExists(chain3.mc, 'getUncle');
        // u.methodExists(chain3.mc, 'getBlockTransactionCount');
        // u.methodExists(chain3.mc, 'getBlockUncleCount');
        // u.methodExists(chain3.mc, 'filter');
        // u.methodExists(chain3.mc, 'contract');
        // u.methodExists(chain3, 'encodeParams');

        // u.propertyExists(chain3.mc, 'coinbase');
        // u.propertyExists(chain3.mc, 'mining');
        // u.propertyExists(chain3.mc, 'gasPrice');
        // u.propertyExists(chain3.mc, 'accounts');
        // u.propertyExists(chain3.mc, 'defaultBlock');
        u.propertyExists(chain3.mc, 'blockNumber');
        // u.propertyExists(chain3.mc, 'protocolVersion');

        
    });
});

// describe('chain3.vnode', function() {
//     describe('methods', function() {
//         u.methodExists(chain3.vnode, 'getBalance');
//         u.methodExists(chain3.vnode, 'getBlock');
//         u.methodExists(chain3.vnode, 'getBlockNumber');
//         u.methodExists(chain3.vnode, 'getNonce');
//     });
// });

// describe('chain3.scs', function() {
//     describe('methods', function() {
//         u.methodExists(chain3.scs, 'getBalance');
//         u.methodExists(chain3.scs, 'getBlock');
//         u.methodExists(chain3.scs, 'getBlockNumber');
//         u.methodExists(chain3.scs, 'getNonce');
//         u.methodExists(chain3.scs, 'getTransactionReceipt');
//         u.methodExists(chain3.scs, 'getSCSId');
//     });
// });
