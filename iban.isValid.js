var chai = require('chai');
var Iban = require('../lib/chain3/iban.js');
var assert = chai.assert;

var tests = [
    { obj: 'XE66MOACXREGGAVOFYORK', is: true},
    { obj: 'XE81MCXREGGAVOFYORK', is: false}
];

describe('lib/chain3/iban', function () {
    describe('isValid', function () {
        tests.forEach(function (test) {
            it('shoud test if value ' + test.obj + ' is iban: ' + test.is, function () {
                assert.equal(Iban.isValid(test.obj), test.is);
            });
        });   
    });
});

