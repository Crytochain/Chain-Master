/* jshint ignore:start */


// Browser environment
if(typeof window !== 'undefined') {
    Chain3 = (typeof window.Chain3 !== 'undefined') ? window.Chain3 : require('chain3');
    BigNumber = (typeof window.BigNumber !== 'undefined') ? window.BigNumber : require('bignumber.js');
}


// Node environment
if(typeof global !== 'undefined') {
    Chain3 = (typeof global.Chain3 !== 'undefined') ? global.Chain3 : require('chain3');
    BigNumber = (typeof global.BigNumber !== 'undefined') ? global.BigNumber : require('bignumber.js');
}

/* jshint ignore:end */