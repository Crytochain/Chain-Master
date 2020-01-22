var chai = require('chai');
var Chain3 = require('../index');
var chain3 = new Chain3();
var assert = chai.assert;
var FakeHttpProvider = require('./helpers/FakeHttpProvider');
var errors = require('../lib/chain3/errors');

var method = 'filter';

var tests = [{
    args: [{
        fromBlock: 0,
        toBlock: 10,
        address: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855'
    }],
    formattedArgs: [{
        fromBlock: '0x0',
        toBlock: '0xa',
        address: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855',
        topics: []
    }],
    result: '0xf',
    formattedResult: '0xf',
    call: 'mc_newFilter'
}, {
    args: [{
        fromBlock: 'latest',
        toBlock: 'latest',
        address: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855'
    }],
    formattedArgs: [{
        fromBlock: 'latest',
        toBlock: 'latest',
        address: '0x47d33b27bb249a2dbab4c0612bf9caf4c1950855',
        topics: []
    }],
    result: '0xf',
    formattedResult: '0xf',
    call: 'mc_newFilter'
}, {
    args: ['latest'],
    formattedArgs: [],
    result: '0xf',
    formattedResult: '0xf',
    call: 'mc_newBlockFilter'
}, {
    args: ['pending'],
    formattedArgs: [],
    result: '0xf',
    formattedResult: '0xf',
    call: 'mc_newPendingTransactionFilter'
}];

describe('chain3.mc', function () {
    describe(method, function () {
        tests.forEach(function (test, index) {
            it('property test: ' + index, function (done) {

                // given
                var provider = new FakeHttpProvider();
                chain3.reset();
                chain3.setProvider(provider);
                provider.injectResult(test.result);
                provider.injectValidation(function (payload) {
                    assert.equal(payload.jsonrpc, '2.0');
                    assert.equal(payload.method, test.call);
                    assert.deepEqual(payload.params, test.formattedArgs);
                });

                // call
                var filter = chain3.mc[method].apply(chain3.mc, test.args);

                // test filter.get
                if (typeof test.args === 'object') {

                    var logs = [{
                        data: '0xb'
                    }, {
                        data: '0x11'
                    }];

                    provider.injectResult(logs);
                    provider.injectValidation(function (payload) {
                        assert.equal(payload.jsonrpc, '2.0');
                        assert.equal(payload.method, 'mc_getFilterLogs');
                        assert.deepEqual(payload.params, [test.formattedResult]);
                    });

                    // sync should throw an error
                    try {
                        assert.throws(filter.get());
                    } catch (e) {
                        assert.instanceOf(e, Error);
                    }

                    // async should get the fake logs
                    filter.get(function (e, res) {
                        assert.deepEqual(logs, res);
                        chain3.reset();
                        done();
                    });
                }
            });

            it('should call filterCreationErrorCallback on error while filter creation', function (done) {
                // given
                var provider = new FakeHttpProvider();
                chain3.reset();
                chain3.setProvider(provider);
                provider.injectError(errors.InvalidConnection());
                // call
                var args = test.args.slice();
                args.push(undefined);
                args.push(function (err) {
                    assert.include(errors, err);
                    done();
                });
                chain3.mc[method].apply(chain3.mc, args);
            })
        });
    });
});