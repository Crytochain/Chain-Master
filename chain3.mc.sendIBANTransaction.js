var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../index');
var chain3 = new Chain3();
var FakeHttpProvider2 = require('./helpers/FakeHttpProvider2');

describe('chain3.mc.sendIBANTransaction', function () {
    it('should send transaction', function () {

        var iban = 'XE66MOACXREGGAVOFYORK';
        var address =   '0x1234567890123456789012345678901234500000';
        var exAddress = '0x1234567890123456789012345678901234567890'

        var provider = new FakeHttpProvider2();
        chain3.setProvider(provider);
        chain3.reset();

        provider.injectResultList([{
            result: exAddress
        }, {
            result: ''
        }]);

        var step = 0;
        provider.injectValidation(function (payload) {
            if (step === 0) {
                step++;
                // console.log("Payload in injectValidation:", payload);
                assert.equal(payload.method, 'mc_call');
                assert.deepEqual(payload.params, [{
                   data: "0x3b3b57de5852454700000000000000000000000000000000000000000000000000000000",
                   to: chain3.mc.icapNamereg().address
                }, "latest"]);

                return;
            } 
            assert.equal(payload.method, 'mc_sendTransaction');
            
            assert.deepEqual(payload.params, [{
                data: '0xb214faa54741564f46594f524b0000000000000000000000000000000000000000000000',
                from: address,
                to: exAddress,
                value: payload.params[0].value // don't check this
            }]);
        });

        chain3.mc.sendIBANTransaction(address, iban, 10000);

    });
});

