/*
 * Used for SCS test only
 * 
*/
var chai = require('chai');
var assert = chai.assert;
var Chain3 = require('../../index');

var FakeHttpProvider = require('./FakeHttpProvider');
var clone = function (object) { return JSON.parse(JSON.stringify(object)); };

var runTests = function (obj, method, tests) {

    var testName = obj ? 'chain3.' + obj : 'SCS ';

    describe(testName, function () {
        describe(method, function () {
            tests.forEach(function (test, index) {
                it('sync test: ' + index, function () {
                    
                    // given
                    var provider = new FakeHttpProvider();
                    var chain3 = new Chain3(provider);
                    provider.injectResult(test.result);
                    provider.injectValidation(function (payload) {
                        assert.equal(payload.jsonrpc, '2.0');
                        assert.equal(payload.method, test.call);
                        assert.deepEqual(payload.params, test.formattedArgs);
                    });

                    // Add to SCS provider
                    chain3.setScsProvider(provider);
                    var args = clone(test.args)

                    // when
                    if (obj) {
                        var result = chain3[obj][method].apply(chain3[obj], args);
                    } else {
                        var result = chain3[method].apply(chain3, args);
                    }
                    // when
                    //var result = (obj)
                        //? chain3[obj][method].apply(null, test.args.slice(0))
                        //: chain3[method].apply(null, test.args.slice(0));
                    
                    // then 
                    assert.deepEqual(test.formattedResult, result);
                });
                
                it('async test: ' + index, function (done) {
                    
                    // given
                    var provider = new FakeHttpProvider();
                    var chain3 = new Chain3(provider);
                    provider.injectResult(test.result);
                    provider.injectValidation(function (payload) {
                        assert.equal(payload.jsonrpc, '2.0');
                        assert.equal(payload.method, test.call);
                        assert.deepEqual(payload.params, test.formattedArgs);
                    });

                    // Add to SCS provider
                    chain3.setScsProvider(provider);
                    
                    var args = clone(test.args);
                   
                    // add callback
                    args.push(function (err, result) {
                        assert.deepEqual(test.formattedResult, result);
                        done();
                    });

                    // when
                    if (obj) {
                        chain3[obj][method].apply(chain3[obj], args);
                    } else {
                        chain3[method].apply(chain3, args);
                    }
                });
            });
        });
    });

};

module.exports = {
    runTests: runTests
}

