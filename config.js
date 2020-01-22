/*
    This file is part of chain3.js.

    chain3.js is free software: you can redistribute it and/or modify
    it under the terms of the GNU Lesser General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    chain3.js is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Lesser General Public License for more details.

    You should have received a copy of the GNU Lesser General Public License
    along with chain3.js.  If not, see <http://www.gnu.org/licenses/>.
   @file config.js
   @LBR lab
   @date 2018
 */

/**
 * Utils
 * 
 * @module utils
 */

/**
 * Utility functions
 * 
 * @class [utils] config
 * @constructor
 */


/// required to define MC_BIGNUMBER_ROUNDING_MODE
var BigNumber = require('bignumber.js');

var MC_UNITS = [
    'sha',
    'ksha',
    'Msha',
    'Gsha',
    'femtomc',
    'picomc',
    'nanomc',
    'micromc',
    'millimc',
    'nano',
    'micro',
    'milli',
    'mc',
    'grand',
    'Mmc',
    'Gmc',
    'Tmc',
    'Pmc',
    'Emc',
    'Zmc',
    'Ymc',
    'Nmc',
    'Dmc',
    'Vmc',
    'Umc'
];

// MC_POLLING_TIMEOUT: 1000/2,
// Added defaultRecordIndex and defaultRecordSize
// for scs exchange info extract.
module.exports = {
    MC_PADDING: 32,
    MC_SIGNATURE_LENGTH: 4,
    MC_UNITS: MC_UNITS,
    MC_BIGNUMBER_ROUNDING_MODE: { ROUNDING_MODE: BigNumber.ROUND_DOWN },
    MC_POLLING_TIMEOUT: 500,
    defaultBlock: 'latest',
    defaultSCSId: undefined,
    defaultAccount: undefined,
    defaultRecordIndex: 0,
    defaultRecordSize: 20
};

