var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../index');
var chain3 = new Chain3();
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

var method = 'blockNumber';

var tests = [{
    result: '0xb',
    formattedResult: 11,
    call: 'mc_'+ method
}];

describe('chain3.mc', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {
            it('property test: ' + index, function () {
                
                // given
                var provider = new FakeHttpProvider();
                chain3.setProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    assert.deepEqual(payload.params, []);
                });

                // when 
                var result = chain3.mc[method];
                
                // then
                assert.strictEqual(test.formattedResult, result);
            });
            
            it('async get property test: ' + index, function (done) {
                
                // given
                var provider = new FakeHttpProvider();
                chain3.setProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    assert.deepEqual(payload.params, []);
                });

                // when 
                chain3.mc.getBlockNumber(function (err, result) {
                    assert.strictEqual(test.formattedResult, result);
                    done();
                });
                
            });
        });
    });
});

