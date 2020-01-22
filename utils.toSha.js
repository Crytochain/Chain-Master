/*
    'nomc':      '0',
    'sha':       '1',
    'ksha':      '1000',
    'Ksha':      '1000',
    'femtomc':   '1000',
    'msha':      '1000000',
    'Msha':      '1000000',
    'picomc':    '1000000',
    'gsha':      '1000000000',
    'Gsha':      '1000000000',
    'nanomc':    '1000000000',
    'nano':      '1000000000',
    'xiao':      '1000000000',
    'micromc':   '1000000000000',
    'micro':     '1000000000000',
    'sand':      '1000000000000',
    'millimc':   '1000000000000000',
    'milli':     '1000000000000000',
    'mc':        '1000000000000000000',
    'kmc':       '1000000000000000000000',
    'grand':     '1000000000000000000000',
    'mmc':       '1000000000000000000000000',
    'gmc':       '1000000000000000000000000000',
    'tmc':       '1000000000000000000000000000000'
*/
var chai = require('chai');
var utils = require('../lib/utils/utils');
var assert = chai.assert;

describe('lib/utils/utils', function () {
    describe('toSha', function () {
        it('should return the correct value', function () {
            assert.equal(utils.toSha(1, 'nomc'),    '0');
            assert.equal(utils.toSha(1, 'sha'),    '1');
            assert.equal(utils.toSha(1, 'ksha'),   '1000');
            assert.equal(utils.toSha(1, 'Ksha'),   '1000');
            assert.equal(utils.toSha(1, 'femtomc'),   '1000');
            assert.equal(utils.toSha(1, 'msha'),     '1000000');
            assert.equal(utils.toSha(1, 'Msha'),     '1000000');
            assert.equal(utils.toSha(1, 'picomc'),   '1000000');
            assert.equal(utils.toSha(1, 'gsha'),     '1000000000');
            assert.equal(utils.toSha(1, 'Gsha'),     '1000000000');
            assert.equal(utils.toSha(1, 'nanomc'),   '1000000000');
            assert.equal(utils.toSha(1, 'nano'),     '1000000000');
            assert.equal(utils.toSha(1, 'xiao'),     '1000000000');
            assert.equal(utils.toSha(1, 'micromc'),  '1000000000000');
            assert.equal(utils.toSha(1, 'sand'),     '1000000000000');
            assert.equal(utils.toSha(1, 'milli'),    '1000000000000000');
            assert.equal(utils.toSha(1, 'mc'),  '1000000000000000000');
            assert.equal(utils.toSha(1, 'kmc'), '1000000000000000000000');
            assert.equal(utils.toSha(1, 'grand'),  '1000000000000000000000');
            assert.equal(utils.toSha(1, 'mmc'), '1000000000000000000000000');
            assert.equal(utils.toSha(1, 'gmc'), '1000000000000000000000000000');
            assert.equal(utils.toSha(1, 'tmc'), '1000000000000000000000000000000');

            assert.equal(utils.toSha(1, 'ksha'),    utils.toSha(1, 'femtomc'));
            assert.equal(utils.toSha(1, 'sand'),   utils.toSha(1, 'micromc'));
            assert.equal(utils.toSha(1, 'xiao'),  utils.toSha(1, 'nanomc'));
            assert.equal(utils.toSha(1, 'milli'),    utils.toSha(1, 'millimc'));
            assert.equal(utils.toSha(1, 'milli'),    utils.toSha(1000, 'micro'));

            assert.throws(function () {utils.toSha(1, 'sha1');}, Error);
        });
    });
});
