'use strict'

var chai = require('chai');
var Chain3 = require('../index');
var assert = chai.assert;
var FakeHttpProvider = require('./helpers/FakeHttpProvider');

describe('mc', function () {
    describe('getSyncing', function () {
        it('syncing object', function (done) {
            // given
            var provider = new FakeHttpProvider();
            var chain3 = new Chain3(provider);
            provider.injectResult({
                startingBlock: '0xb',
                currentBlock: '0xb',
                highestBlock: '0xb'
            });
            provider.injectValidation(function(payload) {
                assert.equal(payload.jsonrpc, '2.0', 'failed');
                assert.equal(payload.method, 'mc_syncing');
            });

            // call
            chain3.mc.getSyncing(function(err, res){
                assert.deepEqual(res, {
                    startingBlock: 11,
                    currentBlock: 11,
                    highestBlock: 11
                });
                done();
            });
        });

        it('false', function (done) {
            // given
            var provider = new FakeHttpProvider();
            var chain3 = new Chain3(provider);
            provider.injectResult(false);
            provider.injectValidation(function(payload) {
                assert.equal(payload.jsonrpc, '2.0', 'failed');
                assert.equal(payload.method, 'mc_syncing');
            });

            // call
            chain3.mc.getSyncing(function(err, res){
                assert.strictEqual(res, false);
                done();
            });
        });
    });
});

