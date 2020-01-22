var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../index');
var chain3 = new Chain3();
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

var method = 'protocolVersion';

var tests = [{
    result: ['1234'],
    call: 'mc_'+ method
}];

describe('mc.protocolVersion', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {
            it('property test: ' + index, function () {

                // given
                var provider = new FakeHttpProvider();
                chain3.setProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    // console.log("Get payload:", payload);

                    assert.equal(payload.jsonrpc, '2.0');
                    // console.log("pass 1");

                    assert.equal(payload.method, test.call);
                    // console.log("pass 2");
                    assert.deepEqual(payload.params, []);
                    // console.log("pass 3");
                });
                // when
                var result = chain3.mc[method];

                // then
                assert.deepEqual(test.result, result);
            });
        });
    });
});
