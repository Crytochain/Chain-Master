var chai = require('chai');
var Iban = require('../lib/chain3/iban.js');
var assert = chai.assert;

var tests = [
    { institution: 'XREG', identifier: 'GAVOFYORK', expected: 'XE66MOACXREGGAVOFYORK'}
];

describe('lib/chain3/iban', function () {
    describe('createIndirect', function () {
        tests.forEach(function (test) {
            it('shoud create indirect iban: ' +  test.expected, function () {
                assert.deepEqual(Iban.createIndirect({
                    institution: test.institution,
                    identifier:  test.identifier
                }), new Iban(test.expected));
            });
        });   
    });
});

