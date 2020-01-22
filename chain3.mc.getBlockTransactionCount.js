var chai = require('chai');
var chain3 = require('../index');
var testMethod = require('./helpers/test.method.js');

var method = 'getBlockTransactionCount';


var tests = [{
    args: ['0x47d33b27bb249a2dbab4c0612bf9caf4c1950855'],
    formattedArgs: ['0x47d33b27bb249a2dbab4c0612bf9caf4c1950855'],
    result: '0xb',
    formattedResult: 11,
    call: 'mc_getBlockTransactionCountByHash'
},{
    args: [436],
    formattedArgs: ['0x1b4'],
    result: '0xb',
    formattedResult: 11,
    call: 'mc_getBlockTransactionCountByNumber'
},{
    args: ['pending'],
    formattedArgs: ['pending'],
    result: '0xb',
    formattedResult: 11,
    call: 'mc_getBlockTransactionCountByNumber'
}];

testMethod.runTests('mc', method, tests);

