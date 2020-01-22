var chai = require('chai');
var Iban = require('../lib/chain3/iban.js');
var assert = chai.assert;

var tests = [
    { direct: 'XE7338O073KYGTWWZN0F2WZ0R8PX5ZPPZS', address: '00c5496aee77c1ba1f0854206a26dda82a81d6d8'}
];

describe('lib/chain3/iban', function () {
    describe('toAddress', function () {
        tests.forEach(function (test) {
            it('shoud transform iban to address: ' +  test.address, function () {
                var iban = new Iban(test.direct);

                assert.deepEqual(iban.address(), test.address);
            });
        });   
    });
});

