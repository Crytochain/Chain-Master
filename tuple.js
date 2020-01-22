var f = require('./formatters');
var SolidityType = require('./type');

/**
 * Maps the correct tuple type and name when the simplified format in encode/decodeParameter is used
 *
 * @method mapStructNameAndType
 * @param {string} structName
 * @return {{type: string, name: *}}
 */
// ABICoder.prototype.mapStructNameAndType = function (structName) {
//     var type = 'tuple';

//     if (structName.indexOf('[]') > -1) {
//         type = 'tuple[]';
//         structName = structName.slice(0, -2);
//     }

//     return {type: type, name: structName};
// };

/**
 * SolidityTypeTuple is a protype that represents structure type
 * in Solidity.
 * Solidity supports all the types as ABIs with the same names 
 * with the exception of tuples. 
 * Some Solidity types are not supported by the ABI. 
 * The following table shows on the left column Solidity types 
 * that are not part of the ABI, and on the right column the ABI types that represent them.
 * ========================================================================
 * Solidity  |  ABI
 * address   |  payable address
 * contract  |  address
 * enum      |  smallest uint type that is large enough to hold all values
 *              For example, an enum of 255 values or less is mapped to 
 *              uint8 and an enum of 256 values is mapped to uint16.
 * struct    |  tuple
 * ========================================================================
 * It is supported by solidity v0.4.21 and later version.
 * Tuple type is a list of objects of potentially different types 
 * whose size is a constant at compile-time. 
 * Those tuples can be used to return multiple values at the same 
 * time and also assign them to multiple variables 
 * (or LValues in general) at the same time:
 * https://solidity.readthedocs.io/en/v0.5.10/abi-spec.html#types
 * Handling tuple types
 * Despite that names are intentionally not part of the ABI encoding 
 * they do make a lot of sense to be included in the JSON to enable 
 * displaying it to the end user. The structure is nested in the following way:
 * An object with members name, type and potentially components describes 
 * a typed variable. The canonical type is determined until a tuple type 
 * is reached and the string description up to that point is stored in type 
 * prefix with the word tuple, i.e. it will be tuple followed by a sequence 
 * of [] and [k] with integers k. The components of the tuple are then stored 
 * in the member components, which is of array type and has the same structure 
 * as the top-level object except that indexed is not allowed there.
 * It matches:
 * 
 */
var SolidityTypeTuple = function () {
    
    this._inputFormatter = f.formatInputTuple;
    this._outputFormatter = f.formatOutputTuple;
};

SolidityTypeTuple.prototype = new SolidityType({});
SolidityTypeTuple.prototype.constructor = SolidityTypeTuple;

// Need to identify the type name
// Note its also possible to have struct
SolidityTypeTuple.prototype.isType = function (name) {
    // tuple type should be an object with 
    // components 
    return !!name.match(/tuple(\[([0-9]*)\])?/);
};

module.exports = SolidityTypeTuple;
