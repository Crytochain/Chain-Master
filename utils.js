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
*/
/**
 * @file utils.js
 * @author Marek Kotewicz <marek@ethdev.com>
 * @date 2015
 * @modified to work with LBR network
 * renamed the unit map
 * Added bignumber.js
 * @LBR lab
 * @date 2018
 */

var _ = require('underscore');
var BigNumber = require('bignumber.js');
var sha3 = require('./sha3.js');
var utf8 = require('utf8');
var assert = require('assert');
var hashjs = require('hash.js');
var BN = require('bn.js');
var numberToBN = require('number-to-bn');

//used the utils function
//to avoid using the BUFFER lib, internal function convert to String
function bytesArrayToHEXString(in_bytes) {
    //Should check the array contents
    if (in_bytes.length === 0) throw new TypeError('Empty array!! Expected input bytes array length > 0')

    try {

        var out_str = ""; //output HEX string

        for (var i = 0; i < in_bytes.length; i++) {
            //convert the byte to HEX
            //and append to the string
            //Note the bigNumber only returns the value in HEX, not adding 0s
            //e.g. '10' -> 'a', need to add '0', '10' -> '0a'
            // 
            var tmp = new BigNumber(in_bytes[i]).toString(16);
            if (tmp.length < 2)
                out_str += '0' + tmp;
            else
                out_str += tmp;
            // out_str += (new BigNumber(in_bytes[i]).toString(16));

        }
        return out_str;

    } catch (e) {
        // Handle possible errors here
        throw new Error('Error from Bytes array to HEX string :' + e + '!');
    }

}

//used the utils function
//to avoid using the BN.js lib
//Note this function only to be used for the decode of the 
//input HEX string 

function hexStringToBytesArray(in_str) {
    //Should check the array contents

    var strlen = in_str.length;

    if (in_str.slice(0, 1) == '0x' && strlen > 2)
        in_str = in_str.slice(2, strlen);
    if (in_str.length === 0) throw new TypeError('Empty input string!! Expected input string length > 0')

    //the input HEX string should have even length, if not , add '0' to make it
    if (in_str.length % 2 != 0)
        in_str = '0' + in_str;


    try {

        var out_array = new Array(in_str.length / 2); //output byte array

        for (var i = 0; i < in_str.length; i = i + 2) {
            //convert the byte to HEX
            //and append to the string
            //Note the bigNumber only returns the value in HEX, not adding 0s
            //e.g. '10' -> 'a', need to add '0', '10' -> '0a'
            // 

            var tmp = in_str.slice(i, i + 2);

            out_array[i / 2] = new BigNumber(tmp, 16).toNumber(10);

        }
        return out_array;

    } catch (e) {
        // Handle possible errors here
        throw new Error('Error from HEX string decode error:' + e + '!');
    }

}

/*
 * Return the hash result of the input bytes
 */
function sha256(bytes) {
    return hashjs.sha256().update(bytes).digest();
}


/*
  Unit of LBR MC is sha (sand)
     To see a world in a grain of sand, 
     And a heaven in a wild flower,
     Hold infinty in the palm of your hand,
     And eternity in an hour.
    William Blake (1757–1827), Auguries Of Innocence 
    1 mc = 1,000,000 sand
    1 mc = 1,000 Xiao
*/

var unitMap = {
    'nomc': '0',
    'sha': '1',
    'ksha': '1000',
    'Ksha': '1000',
    'femtomc': '1000',
    'msha': '1000000',
    'Msha': '1000000',
    'picomc': '1000000',
    'gsha': '1000000000',
    'Gsha': '1000000000',
    'nanomc': '1000000000',
    'nano': '1000000000',
    'xiao': '1000000000',
    'micromc': '1000000000000',
    'micro': '1000000000000',
    'sand': '1000000000000',
    'millimc': '1000000000000000',
    'milli': '1000000000000000',
    'mc': '1000000000000000000',
    'ine': '1000000000000000000',
    'kmc': '1000000000000000000000',
    'grand': '1000000000000000000000',
    'mmc': '1000000000000000000000000',
    'gmc': '1000000000000000000000000000',
    'tmc': '1000000000000000000000000000000'
};

/**
 * Should be called to pad string to expected length
 *
 * @method padLeft
 * @param {String} string to be padded
 * @param {Number} characters that result string should have
 * @param {String} sign, by default 0
 * @returns {String} right aligned string
 */
var padLeft = function (string, chars, sign) {
    return new Array(chars - string.length + 1).join(sign ? sign : "0") + string;
};

/**
 * Should be called to pad string to expected length
 *
 * @method padRight
 * @param {String} string to be padded
 * @param {Number} characters that result string should have
 * @param {String} sign, by default 0
 * @returns {String} right aligned string
 */
var padRight = function (string, chars, sign) {
    return string + (new Array(chars - string.length + 1).join(sign ? sign : "0"));
};

/**
 * Should be called to get utf8 from it's hex representation
 *
 * @method toUtf8
 * @param {String} string in hex
 * @returns {String} ascii string representation of hex value
 */
var toUtf8 = function (hex) {
    // Find termination
    var str = "";
    var i = 0,
        l = hex.length;
    if (hex.substring(0, 2) === '0x') {
        i = 2;
    }
    for (; i < l; i += 2) {
        var code = parseInt(hex.substr(i, 2), 16);
        if (code === 0)
            break;
        str += String.fromCharCode(code);
    }

    return utf8.decode(str);
};

/**
 * Should be called to get ascii from it's hex representation
 *
 * @method toAscii
 * @param {String} string in hex
 * @returns {String} ascii string representation of hex value
 */
var toAscii = function (hex) {
    // Find termination
    var str = "";
    var i = 0,
        l = hex.length;
    if (hex.substring(0, 2) === '0x') {
        i = 2;
    }
    for (; i < l; i += 2) {
        var code = parseInt(hex.substr(i, 2), 16);
        str += String.fromCharCode(code);
    }

    return str;
};

/**
 * Should be called to get hex representation (prefixed by 0x) of utf8 string
 *
 * @method fromUtf8
 * @param {String} string
 * @param {Number} optional padding
 * @returns {String} hex representation of input string
 */
var fromUtf8 = function (str, allowZero) {
    str = utf8.encode(str);
    var hex = "";
    for (var i = 0; i < str.length; i++) {
        var code = str.charCodeAt(i);
        if (code === 0) {
            if (allowZero) {
                hex += '00';
            } else {
                break;
            }
        } else {
            var n = code.toString(16);
            hex += n.length < 2 ? '0' + n : n;
        }
    }

    return "0x" + hex;
};

/**
 * Should be called to get hex representation (prefixed by 0x) of ascii string
 *
 * @method fromAscii
 * @param {String} string
 * @param {Number} optional padding
 * @returns {String} hex representation of input string
 */
var fromAscii = function (str) {
    var hex = "";
    for (var i = 0; i < str.length; i++) {
        var code = str.charCodeAt(i);
        var n = code.toString(16);
        hex += n.length < 2 ? '0' + n : n;
    }

    return "0x" + hex;
};

/**
 * Should be used to create full function/event name from json abi
 *
 * @method transformToFullName
 * @param {Object} json-abi
 * @return {String} full fnction/event name
 */
var transformToFullName = function (json) {
    if (json.name.indexOf('(') !== -1) {
        return json.name;
    }

    var typeName = json.inputs.map(function (i) {
        return i.type;
    }).join();
    return json.name + '(' + typeName + ')';
};

/**
 * Should be called to get display name of contract function
 *
 * @method extractDisplayName
 * @param {String} name of function/event
 * @returns {String} display name for function/event eg. multiply(uint256) -> multiply
 */
var extractDisplayName = function (name) {
    var length = name.indexOf('(');
    return length !== -1 ? name.substr(0, length) : name;
};

/// @returns overloaded part of function/event name
var extractTypeName = function (name) {
    /// TODO: make it invulnerable
    var length = name.indexOf('(');
    return length !== -1 ? name.substr(length + 1, name.length - 1 - (length + 1)).replace(' ', '') : "";
};

/**
 * Converts value to it's decimal representation in string
 *
 * @method toDecimal
 * @param {String|Number|BigNumber}
 * @return {String}
 */
var toDecimal = function (value) {
    return toBigNumber(value).toNumber();
};

/**
 * Converts value to it's hex representation
 *
 * @method fromDecimal
 * @param {String|Number|BigNumber}
 * @return {String}
 */
var fromDecimal = function (value) {
    var number = toBigNumber(value);
    var result = number.toString(16);

    return number.isLessThan(0) ? '-0x' + result.substr(1) : '0x' + result;
};

/**
 * Converts a `Number` into a hex `String`
 * This could have problem when the input number is
 * too large
 *  Use bn.js to convert the input to a HEX number
 * certain range.
 * @param {Number} i
 * @return {String}
 */
var BigIntToHex = function BigIntToHex(i) {

    var inNum = new BN(i, 10);
    var hex = makeEven(inNum.toString(16));


    // return `0x${makeEven(hex)}`;
    return '0x' + hex;
}

var makeEven = function (hex) {
    if (hex.length % 2 === 1) {
        hex = hex.replace('0x', '0x0');
    }
    return hex;
};

/**
 * Converts value to it's hex representation
 *
 * @method numberToHex
 * @param {String|Number|BN} value
 * @return {String}
 */
var numberToHex = function (value) {
    if (_.isNull(value) || _.isUndefined(value)) {
        return value;
    }

    if (!isFinite(value) && !isHexStrict(value)) {
        throw new Error('Given input "' + value + '" is not a number.');
    }

    var number = toBN(value);
    var result = number.toString(16);

    return number.lt(new BN(0)) ? '-0x' + result.substr(1) : '0x' + result;
};

/*
 * RLP usage, convert an integer to HEX
 * input could be a big number object
 */
function intToHex(i) {
    var hex = i.toString(16)
    if (hex.length % 2) {
        hex = '0' + hex
    }

    return hex
}

/**
 * Check if string is HEX, requires a 0x in front
 *
 * @method isHexStrict
 * @param {String} hex to be checked
 * @returns {Boolean}
 */
var isHexStrict = function (hex) {
    return ((_.isString(hex) || _.isNumber(hex)) && /^(-)?0x[0-9a-f]*$/i.test(hex));
};



/**
 * Auto converts any given value into it's hex representation.
 *
 * And even stringifys objects before.
 *
 * @method toHex
 * @param {String|Number|BigNumber|Object}
 * @return {String}
 */
var toHex = function (val) {
    /*jshint maxcomplexity: 8 */

    if (isBoolean(val))
        return fromDecimal(+val);

    if (isBigNumber(val))
        return fromDecimal(val);

    if (typeof val === 'object')
        return fromUtf8(JSON.stringify(val));

    // if its a negative number, pass it through fromDecimal
    if (isString(val)) {
        if (val.indexOf('-0x') === 0)
            return fromDecimal(val);
        else if (val.indexOf('0x') === 0)
            return val;
        else if (!isFinite(val)) {

            return fromUtf8(val, 1);
        }

    }

    return fromDecimal(val);
};

/**
 * Returns value of unit in Sha
 *
 * @method getValueOfUnit
 * @param {String} unit the unit to convert to, default mc
 * @returns {BigNumber} value of the unit (in Sha)
 * @throws error if the unit is not correct:w
 */
var getValueOfUnit = function (unit) {
    unit = unit ? unit.toLowerCase() : 'mc';
    var unitValue = unitMap[unit];
    if (unitValue === undefined) {
        throw new Error('This unit [' + unit + '] doesn\'t exists, please use the one of the following units' + JSON.stringify(unitMap, null, 2));
    }
    return new BigNumber(unitValue, 10);
};

/**
 * Takes a number of sha and converts it to any other LBR unit.
 * 1 sha = 1e-18 mc 
 * Possible units are:
 *   SI Short   SI Full     PowerOfTen  Other
 * - ksha       femtomc     -15
 * - msha       picomc      -12
 * - gsha       nanomc      -9          nano
 * - --         micromc     -6          micro
 * - --         millimc     -3          milli
 * - mc         -           0
 * - kmc                    3           grand
 * - mmc                    6 
 * - gmc                    9
 * - tmc                    12 
 *
 * @method fromSha
 * @param {Number|String} number can be a number, number string or a HEX of a decimal
 * @param {String} unit the unit to convert to, default mc
 * @return {String|Object} When given a BigNumber object it returns one as well, otherwise a number
 */
var fromSha = function (number, unit) {
    var returnValue = toBigNumber(number).dividedBy(getValueOfUnit(unit));

    return isBigNumber(number) ? returnValue : returnValue.toString(10);
};

/**
 * Takes a number of a unit and converts it to sha.
 *
 * Possible units are:
 *   SI Short   SI Full     PowerOfTen  Other
 * - ksha       femtomc     -15
 * - msha       picomc      -12
 * - gsha       nanomc      -9          nano
 * - --         micromc     -6          micro
 * - --         millimc     -3          milli
 * - mc         -           0
 * - kmc                    3           grand
 * - mmc                    6 
 * - gmc                    9
 * - tmc                    12 
 *
 * @method toSha
 * @param {Number|String|BigNumber} number can be a number, number string or a HEX of a decimal
 * @param {String} unit the unit to convert from, default mc
 * @return {String|Object} When given a BigNumber object it returns one as well, otherwise a number
 */
var toSha = function (number, unit) {
    var returnValue = toBigNumber(number).times(getValueOfUnit(unit));

    return isBigNumber(number) ? returnValue : returnValue.toString(10);
};

/**
 * Takes an input and transforms it into an bignumber
 *
 * @method toBigNumber
 * @param {Number|String|BigNumber} a number, string, HEX string or BigNumber
 * @return {BigNumber} BigNumber
 */
var toBigNumber = function (number) {
    /*jshint maxcomplexity:5 */
    number = number || 0;
    if (isBigNumber(number))
        return number;

    if (isString(number) && (number.indexOf('0x') === 0 || number.indexOf('-0x') === 0)) {
        return new BigNumber(number.toLowerCase().replace('0x', ''), 16);
    }

    return new BigNumber(number.toString(10), 10);
};

/**
 * Takes an input and transforms it into an BN
 *
 * @method toBN
 * @param {Number|String|BN} number, string, HEX string or BN
 * @return {BN} BN
 */
var toBN = function (number) {
    try {
        return numberToBN.apply(null, arguments);
    } catch (e) {
        throw new Error(e + ' Given value: "' + number + '"');
    }
};

/**
 * Takes and input transforms it into bignumber and if it is negative value, into two's complement
 *
 * @method toTwosComplement
 * @param {Number|String|BigNumber}
 * @return {BigNumber}
 */
var toTwosComplement = function (number) {
    var bigNumber = toBigNumber(number).integerValue();
    if (bigNumber.isLessThan(0)) {
        return new BigNumber("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff", 16).plus(bigNumber).plus(1);
    }
    return bigNumber;
};

/**
 * Checks if the given string is strictly an address
 *
 * @method isStrictAddress
 * @param {String} address the given HEX adress
 * @return {Boolean}
 */
var isStrictAddress = function (address) {
    return /^0x[0-9a-f]{40}$/i.test(address);
};

/**
 * Checks if the given string is an address
 *
 * @method isAddress
 * @param {String} address the given HEX adress
 * @return {Boolean}
 */
var isAddress = function (address) {
    if (!/^(0x)?[0-9a-f]{40}$/i.test(address)) {
        // check if it has the basic requirements of an address
        return false;
    } else if (/^(0x)?[0-9a-f]{40}$/.test(address) || /^(0x)?[0-9A-F]{40}$/.test(address)) {
        // If it's all small caps or all all caps, return true
        return true;
    } else {
        // Otherwise check each case
        return isChecksumAddress(address);
    }
};

/**
 * Checks if the given string is a checksummed address
 *
 * @method isChecksumAddress
 * @param {String} address the given HEX adress
 * @return {Boolean}
 */
var isChecksumAddress = function (address) {
    // Check each case
    address = address.replace('0x', '');
    var addressHash = sha3(address.toLowerCase());

    for (var i = 0; i < 40; i++) {
        // the nth letter should be uppercase if the nth digit of casemap is 1
        if ((parseInt(addressHash[i], 16) > 7 && address[i].toUpperCase() !== address[i]) || (parseInt(addressHash[i], 16) <= 7 && address[i].toLowerCase() !== address[i])) {
            return false;
        }
    }
    return true;
};



/**
 * Makes a checksum address
 *
 * @method toChecksumAddress
 * @param {String} address the given HEX adress
 * @return {String}
 */
var toChecksumAddress = function (address) {
    if (typeof address === 'undefined') return '';

    address = address.toLowerCase().replace('0x', '');
    var addressHash = sha3(address);
    var checksumAddress = '0x';

    for (var i = 0; i < address.length; i++) {
        // If ith character is 9 to f then make it uppercase
        if (parseInt(addressHash[i], 16) > 7) {
            checksumAddress += address[i].toUpperCase();
        } else {
            checksumAddress += address[i];
        }
    }
    return checksumAddress;
};

/**
 * Transforms given string to valid 20 bytes-length addres with 0x prefix
 *
 * @method toAddress
 * @param {String} address
 * @return {String} formatted address
 */
var toAddress = function (address) {
    if (isStrictAddress(address)) {
        return address;
    }

    if (/^[0-9a-f]{40}$/.test(address)) {
        return '0x' + address;
    }

    return '0x' + padLeft(toHex(address).substr(2), 40);
};

/**
 * Returns true if object is BigNumber, otherwise false
 *
 * @method isBigNumber
 * @param {Object}
 * @return {Boolean}
 */
var isBigNumber = function (object) {
    return object instanceof BigNumber ||
        (object && object.constructor && object.constructor.name === 'BigNumber');
};

/**
 * Returns true if object is string, otherwise false
 *
 * @method isString
 * @param {Object}
 * @return {Boolean}
 */
var isString = function (object) {
    return typeof object === 'string' ||
        (object && object.constructor && object.constructor.name === 'String');
};

/**
 * Check if string is HEX, requires a 0x in front
 *
 * @method isHexStrict
 * @param {String} hex to be checked
 * @returns {Boolean}
 */
var isHexStrict = function (hex) {
    return ((_.isString(hex) || _.isNumber(hex)) && /^(-)?0x[0-9a-f]*$/i.test(hex));
};

/**
 * Returns true if object is function, otherwise false
 *
 * @method isFunction
 * @param {Object}
 * @return {Boolean}
 */
var isFunction = function (object) {
    return typeof object === 'function';
};

/**
 * Returns true if object is Objet, otherwise false
 *
 * @method isObject
 * @param {Object}
 * @return {Boolean}
 */
var isObject = function (object) {
    return object !== null && !(Array.isArray(object)) && typeof object === 'object';
};

/**
 * Returns true if object is boolean, otherwise false
 *
 * @method isBoolean
 * @param {Object}
 * @return {Boolean}
 */
var isBoolean = function (object) {
    return typeof object === 'boolean';
};

/**
 * Returns true if object is array, otherwise false
 *
 * @method isArray
 * @param {Object}
 * @return {Boolean}
 */
var isArray = function (object) {
    return Array.isArray(object);
};

/**
 * Returns true if given string is valid json object
 *
 * @method isJson
 * @param {String}
 * @return {Boolean}
 */
var isJson = function (str) {
    try {
        return !!JSON.parse(str);
    } catch (e) {
        return false;
    }
};

/**
 * Returns true if given string is a valid LBR node block header bloom.
 *
 * @method isBloom
 * @param {String} hex encoded bloom filter
 * @return {Boolean}
 */
var isBloom = function (bloom) {
    if (!/^(0x)?[0-9a-f]{512}$/i.test(bloom)) {
        return false;
    } else if (/^(0x)?[0-9a-f]{512}$/.test(bloom) || /^(0x)?[0-9A-F]{512}$/.test(bloom)) {
        return true;
    }
    return false;
};

/**
 * Returns true if given string is a valid log topic.
 *
 * @method isTopic
 * @param {String} hex encoded topic
 * @return {Boolean}
 */
var isTopic = function (topic) {
    if (!/^(0x)?[0-9a-f]{64}$/i.test(topic)) {
        return false;
    } else if (/^(0x)?[0-9a-f]{64}$/.test(topic) || /^(0x)?[0-9A-F]{64}$/.test(topic)) {
        return true;
    }
    return false;
};

module.exports = {
    padLeft: padLeft,
    padRight: padRight,
    BigIntToHex: BigIntToHex,
    numberToHex: numberToHex,
    toHex: toHex,
    toDecimal: toDecimal,
    fromDecimal: fromDecimal,
    toUtf8: toUtf8,
    toAscii: toAscii,
    fromUtf8: fromUtf8,
    fromAscii: fromAscii,
    transformToFullName: transformToFullName,
    extractDisplayName: extractDisplayName,
    extractTypeName: extractTypeName,
    toSha: toSha,
    fromSha: fromSha,
    toBigNumber: toBigNumber,
    toTwosComplement: toTwosComplement,
    toAddress: toAddress,
    isBigNumber: isBigNumber,
    isStrictAddress: isStrictAddress,
    isAddress: isAddress,
    isChecksumAddress: isChecksumAddress,
    toChecksumAddress: toChecksumAddress,
    isFunction: isFunction,
    isString: isString,
    isObject: isObject,
    isBoolean: isBoolean,
    isArray: isArray,
    isJson: isJson,
    isBloom: isBloom,
    isTopic: isTopic,
    isHexString: isHexStrict
};

