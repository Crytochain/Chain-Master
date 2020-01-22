var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../index');
var chain3 = new Chain3();
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

var method = 'getSCSId';

// Test object
// need to have input args, output results and formatted results
var tests = [{
    // args: ['0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B','0xD814F2ac2c4cA49b33066582E4e97EBae02F2aB9'],
    // formattedArgs: ['0x7312F4B8A4457a36827f185325Fd6B66a3f8BB8B','0xD814F2ac2c4cA49b33066582E4e97EBae02F2aB9'],
    // args: ['0x47d33b27bb249a2dbab4c0612bf9caf4c1950855','0x000000000000000000000000000000000000013d'],
    // formattedArgs: ['0x47d33b27bb249a2dbab4c0612bf9caf4c1950855','0x000000000000000000000000000000000000013d'],
    result: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855',
    formattedResult: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855',
    call: 'scs_'+ method
}];

describe('chain3.scs', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {
            it('getSCSId test: ' + index, function () {
                
                // given the result to the FakeProvider
                var provider = new FakeHttpProvider();
                chain3.setScsProvider(provider);
                provider.injectResult(test.result);

                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    // assert.deepEqual(payload.params, test.formattedArgs);
                });

                // when the input args more than 1 item, need to input separately
                var result = chain3.scs.getSCSId();
               // then
                assert.strictEqual(test.formattedResult, result);
            });
            
            it('async getSCSId test: ' + index, function (done) {
                
                // given
                var provider = new FakeHttpProvider();
                chain3.setScsProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    // assert.deepEqual(payload.params, test.formattedArgs);
                });

                // when 
                chain3.scs.getSCSId(function (err, result) {
                    assert.strictEqual(test.formattedResult, result);
                    done();
                });
                
            });
        });
    });
});

