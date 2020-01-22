var chai = require('chai');
var Chain3 = require('../index');
var chain3 = new Chain3();
var BigNumber = require('bignumber.js');
var assert = chai.assert;

/*
 * Test the sign of the message using mc_sign method
 * just like the LBR node will do.
*/
// var tests = [
//     { value: '0x7ab47b997cbb195fc158676188647d872de050ca7d4aee1c52756d9458522356', expected: '0x2684060a86a5f2cfeecc8cd1a7590357d888852d22bbc5f9b5aa58b19a3477945ff8ebb4983701b8efce66fd011b392bd8506239394d2392426443b58691eb121b'},
//     { value: '0x3ac225168df54212a25c1c01fd35bebfea408fdac2e31ddd6f80a4bbf9a5f1cb', expected: '0x98421f61ecb8af5ffb504a701960f79c7f8227d7498baaaa2fb452c8d9cf2a293304eb3e8999675ab818003a5e62d8acd16c4efb5dd9e7e81f23fea0490f246f1b'},
// ];
var inTx =  { from: '0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B',
  nonce: '0x2',
  gasPrice: '0x5d21dba00',
  gasLimit: '0x5208',
  to: '0xD814F2ac2c4cA49b33066582E4e97EBae02F2aB9',
  value: '0xde0b6b3a7640000',
  shardingFlag: 0,
  data: '0x7a68656e6770656e676c69333031363035333932327061793130353130306d6f6163666f723135303030706173',
  chainId: '106' };

var signedTx = "0xf89d02808505d21dba0082520894d814f2ac2c4ca49b33066582e4e97ebae02f2ab9880de0b6b3a7640000ad7a68656e6770656e676c69333031363035333932327061793130353130306d6f6163666f723135303030706173808081f8a0dd38cdd4fb120cd964ec65e9fc9819b849a3b707ff59b60e852eab8b92990bcca04e3a2f0d808d0cbdb6480c4ce41eb0dae5446b9d0ad0a129bc35168251d79cb8";

var tacct = {
  "addr": "0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B", 
  "key": "0xc75a5f85ef779dcf95c651612efb3c3b9a6dfafb1bb5375905454d9fc8be8a6b"//put the private key here
};

describe('lib/utils/accounts', function () {
    describe('signMcMessage', function () {
        // tests.forEach(function (test) {
            it('should match signed TX from ' + tacct.addr + ' to ' + signedTx, function () {
                // assert.strictEqual(chain3.signMcMessage(test.value, tacct.key), test.expected);
                assert.strictEqual(chain3.signTransaction(inTx, tacct.key), signedTx);
            });
        // });
    });
});
