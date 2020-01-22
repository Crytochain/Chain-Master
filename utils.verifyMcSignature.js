var chai = require('chai');
var Chain3 = require('../index');
var chain3 = new Chain3();
var BigNumber = require('bignumber.js');
var assert = chai.assert;

/*
 * Test the verify of the MOAC signature of the message
*/
var tests = [
    { msg: '0x7ab47b997cbb195fc158676188647d872de050ca7d4aee1c52756d9458522356', signature: '0x2684060a86a5f2cfeecc8cd1a7590357d888852d22bbc5f9b5aa58b19a3477945ff8ebb4983701b8efce66fd011b392bd8506239394d2392426443b58691eb121b'},
    { msg: '0x3ac225168df54212a25c1c01fd35bebfea408fdac2e31ddd6f80a4bbf9a5f1cb', signature: '0x98421f61ecb8af5ffb504a701960f79c7f8227d7498baaaa2fb452c8d9cf2a293304eb3e8999675ab818003a5e62d8acd16c4efb5dd9e7e81f23fea0490f246f1b'},
];

var tacct = {
  "addr": "0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B", 
  "key": ""
};

describe('lib/utils/accounts', function () {
    describe('verifyMcSignature', function () {
        tests.forEach(function (test) {
            it('should verify ' + test.msg + ' and ' + test.signature, function () {
                assert.strictEqual(chain3.verifyMcSignature(test.msg, test.signature, tacct.addr), true);
            });
        });
    });
});
