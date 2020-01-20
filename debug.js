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


"use strict";

var Method = require('../method');
// var Property = require('../property');
var formatters = require('../formatters');

function Debug(chain3) {
    this._requestManager = chain3._requestManager;

    var self = this;

    methods().forEach(function(method) {
        method.attachToObject(self);
        method.setRequestManager(self._requestManager);
    });

    // properties().forEach(function(p) {
    //     p.attachToObject(self);
    //     p.setRequestManager(self._requestManager);
    // });
}

var methods = function () {
    // Input is a TX HASH
    var traceTransaction = new Method({
        name: 'traceTransaction',
        call: 'debug_traceTransaction',
        params: 1,
        inputFormatter: [null]
    });

    // Input as signed Transaction
    var actualGas = new Method({
        name: 'actualGas',
		call: 'debug_actualGas',
		params: 1
    });


    return [
        traceTransaction,
        actualGas
    ];
};

// var properties = function () {
//     return [
//         new Property({
//             name: 'listAccounts',
//             getter: 'personal_listAccounts'
//         })
//     ];
// };


module.exports = Debug;
