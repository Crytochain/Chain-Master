var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../index');
var chain3 = new Chain3();
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

var method = 'getBlockNumber';

// Test object
var tests = [{
    args: '0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B',//LBRchain address
    result: '0xb',
    formattedResult: 11,
    call: 'scs_'+ method
}];



describe('chain3.scs', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {
            it('getBlockNumber test: ' + index, function () {
                
                // given the result to the FakeProvider
                var provider = new FakeHttpProvider();
                chain3.setScsProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    assert.deepEqual(payload.params, [test.args]);
                });

                // when 
                var result = chain3.scs.getBlockNumber(test.args);//chain3.scs[method];
                // console.log("getNonde:", chain3.scs[method]);
               // then
                assert.strictEqual(test.formattedResult, result);
            });
            
            it('async get getBlockNumber test: ' + index, function (done) {
                
                // given
                var provider = new FakeHttpProvider();
                chain3.setScsProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    assert.deepEqual(payload.params, [test.args]);
                });

                // when 
                chain3.scs.getBlockNumber(test.args, function (err, result) {
                    assert.strictEqual(test.formattedResult, result);
                    done();
                });
                
            });
        });
    });
});

