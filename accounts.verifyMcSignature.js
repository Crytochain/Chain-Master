var chai = require('chai');
var Chain3 = require('../index');
var chain3 = new Chain3();
var BigNumber = require('bignumber.js');
var assert = chai.assert;

/*
 * Test the sign of the message using accounts.signMcMessage
 * Test recover the address from signature 
*/
var sigtests = [
    { msg: '0x7ab47b997cbb195fc158676188647d872de050ca7d4aee1c52756d9458522356', signature: '0x2684060a86a5f2cfeecc8cd1a7590357d888852d22bbc5f9b5aa58b19a3477945ff8ebb4983701b8efce66fd011b392bd8506239394d2392426443b58691eb121b'},
    { msg: '0x3ac225168df54212a25c1c01fd35bebfea408fdac2e31ddd6f80a4bbf9a5f1cb', signature: '0x98421f61ecb8af5ffb504a701960f79c7f8227d7498baaaa2fb452c8d9cf2a293304eb3e8999675ab818003a5e62d8acd16c4efb5dd9e7e81f23fea0490f246f1b'},
];

var addrtests = [
    { msg: '0xa4810c35f20a04bc9e3f642e33531ff0b42f146f8b757523ab5ef6ede54fae39', 
      addr: '0x976e36acb8f8185fb6bb8156d93bde7f9801c352',
      signature: '0xdeee42fa127dd89a10a67048dd7db57f23fb4f44db760f0a4688053a2d088581009e90b1c54c1a13fdf125fc9a1b1f6e060f208cfee6a6f143a8fbbee9676f941c'},
    { msg: '13&0xe7e52b94e9a82351302260ec39a300e9f00aee4c&222.zip', 
    addr:'0xe7e52b94e9a82351302260ec39a300e9f00aee4c',
    signature: '0x005a81058ac2f3651a6fb394bfc8bf81ab7bc493833ff39392c93a4c73fa68c27b5c787095e15030261fb4b43017ad75f869e68429323eda816ad1b5ba9562ac1b'},
];

var tacct = {
  "addr": "0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B"
};

describe('lib/accounts', function () {
    describe('verifyMcSignature', function () {
        sigtests.forEach(function (sigtests) {
            it('should verify ' + sigtests.msg + ' and ' + sigtests.signature, function () {
                assert.strictEqual(chain3.verifyMcSignature(sigtests.msg, sigtests.signature, tacct.addr), true);
            });
        });
    });

    describe('recover Address', function () {
        addrtests.forEach(function (addrtests) {
            it('should recover ' + addrtests.addr + ' from ' + addrtests.signature, function () {
                assert.strictEqual(chain3.recoverPersonalSignature(addrtests.msg, addrtests.signature), addrtests.addr.toLowerCase());
            });
        });
    });
});
