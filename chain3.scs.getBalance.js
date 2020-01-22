var BigNumber = require('bignumber.js');
var Chain3 = require('../index');
var chain3 = new Chain3();
var testMethod = require('./helpers/test.scsMethod.js');

/*
 * Return the Account balance from a LBRChain
 * args 
 * 1. 
 */
var method = 'getBalance';

var tests = [{
    args: ['0x000000000000000000000000000000000000012d', '0x000000000000000000000000000000000000013d'],
    formattedArgs: ['0x000000000000000000000000000000000000012d', '0x000000000000000000000000000000000000013d'],
    result: '0xb',
    formattedResult: new BigNumber('0xb', 16),
    call: 'scs_' + method
}];

testMethod.runTests('scs', method, tests);