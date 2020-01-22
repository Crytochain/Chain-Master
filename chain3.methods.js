var Chain3 = require('../index.js');
var chain3 = new Chain3();
var u = require('./helpers/test.utils.js');

describe('chain3', function() {
    describe('methods', function () {
        u.methodExists(chain3, 'sha3');
        u.methodExists(chain3, 'toAscii');
        u.methodExists(chain3, 'fromAscii');
        u.methodExists(chain3, 'toDecimal');
        u.methodExists(chain3, 'fromDecimal');
        u.methodExists(chain3, 'fromSha');
        u.methodExists(chain3, 'toSha');
        u.methodExists(chain3, 'toBigNumber');
        u.methodExists(chain3, 'isAddress');
        u.methodExists(chain3, 'setProvider');
        u.methodExists(chain3, 'reset');

        u.propertyExists(chain3, 'providers');
        u.propertyExists(chain3, 'mc');
        u.propertyExists(chain3, 'vnode');
        u.propertyExists(chain3, 'scs');

    });
});

