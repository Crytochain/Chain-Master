var chai = require('chai');
var Chain3 = require('../index');
var assert = chai.assert;
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

var method = 'isSyncing';

var tests = [{
    args: [],
    formattedArgs: [],
    result: [{
        startingBlock: '0xb',
        currentBlock: '0xb',
        highestBlock: '0xb'
    }],
    formattedResult: {
        startingBlock: 11,
        currentBlock: 11,
        highestBlock: 11
    },
    call: 'mc_syncing'
}, {
    args: [],
    formattedArgs: [],
    result: [{
        startingBlock: '0xb',
        currentBlock: '0xb',
        highestBlock: '0xb',
        knownStates: '0xb',
        pulledStates: '0xb'
    }],
    formattedResult: {
        startingBlock: 11,
        currentBlock: 11,
        highestBlock: 11,
        knownStates: 11,
        pulledStates: 11
    },
    call: 'mc_syncing'
}];

describe('mc', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {

            it('property test: ' + index, function (done) {

                // given
                var provider = new FakeHttpProvider();
                var chain3 = new Chain3(provider);
provider.injectBatchResults(test.result);
                // provider.injectResult(tests.result);

                provider.injectValidation(function(payload) {
                    //console.log("injectValidation payload:", payload[0])
                    
                    //console.log("Get payload:"+payload[0].jsonrpc, payload[0].method, payload[0].params, test.call);
                    assert.equal(payload[0].jsonrpc, '2.0', 'failed');
                    assert.equal(payload[0].method, test.call);
                    assert.deepEqual(payload[0].params, test.formattedArgs);
               });

                // TODO results seem to be overwritten

                count = 1;
                // call
                var syncing = chain3.mc[method](function(e, res){
                // console.log("Method", method, " res ", res);
                    if(count === 1) {
                        // console.log("Count:", count)
                        assert.isTrue(res);
                        count++;
                    } else {
                        assert.deepEqual(res, test.formattedResult);
                        syncing.stopWatching();
                        done();
                    }
                });

                 

            });
        });
    });
});

