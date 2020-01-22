var assert = require('assert');
var utils = require('../lib/utils/utils.js');

describe('lib/utils/utils', function () {
    describe('fromSha', function () {
        it('should return the correct value', function () {
            
            assert.equal(utils.fromSha(1000000000000000000, 'sha'),    '1000000000000000000');
            assert.equal(utils.fromSha(1000000000000000000, 'ksha'),   '1000000000000000');
            assert.equal(utils.fromSha(1000000000000000000, 'msha'),   '1000000000000');
            assert.equal(utils.fromSha(1000000000000000000, 'gsha'),   '1000000000');
            assert.equal(utils.fromSha(1000000000000000000, 'xiao'),  '1000000000');
            assert.equal(utils.fromSha(1000000000000000000, 'sand'), '1000000');
            assert.equal(utils.fromSha(1000000000000000000, 'mc'),  '1');
            assert.equal(utils.fromSha(1000000000000000000, 'kmc'), '0.001');
            assert.equal(utils.fromSha(1000000000000000000, 'grand'),  '0.001');
            assert.equal(utils.fromSha(1000000000000000000, 'mmc'), '0.000001');
            assert.equal(utils.fromSha(1000000000000000000, 'gmc'), '0.000000001');
            assert.equal(utils.fromSha(1000000000000000000, 'tmc'), '0.000000000001');
        });
    });
});
