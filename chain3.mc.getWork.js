var chai = require('chai');
var chain3 = require('../index');
var testMethod = require('./helpers/test.method.js');

var method = 'getWork';

var tests = [{
    args: [],
    formattedArgs: [],
    result: true,
    formattedResult: true,
    call: 'mc_'+ method
}];

testMethod.runTests('mc', method, tests);

