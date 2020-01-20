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
/** @file watches.js
 * @authors:
 *   Marek Kotewicz <marek@ethdev.com>
 *  @date 2015
 * @modified for LBR project
 * @LBR lab
 * @date 2018
 */

var Method = require('../method');

/// @returns an array of objects describing chain3.mc.filter api methods
var mc = function () {
    var newFilterCall = function (args) {
        var type = args[0];

        switch(type) {
            case 'latest':
                args.shift();
                this.params = 0;
                return 'mc_newBlockFilter';
            case 'pending':
                args.shift();
                this.params = 0;
                return 'mc_newPendingTransactionFilter';
            default:
                return 'mc_newFilter';
        }
    };

    var newFilter = new Method({
        name: 'newFilter',
        call: newFilterCall,
        params: 1
    });

    var uninstallFilter = new Method({
        name: 'uninstallFilter',
        call: 'mc_uninstallFilter',
        params: 1
    });

    var getLogs = new Method({
        name: 'getLogs',
        call: 'mc_getFilterLogs',
        params: 1
    });

    var poll = new Method({
        name: 'poll',
        call: 'mc_getFilterChanges',
        params: 1
    });

    return [
        newFilter,
        uninstallFilter,
        getLogs,
        poll
    ];
};

// @returns an array of objects describing chain3.scs.filter api methods
// @
var scs = function () {
    var newFilterCall = function (args) {
        var type = args[0];

        switch(type) {
            case 'latest':
                args.shift();
                this.params = 1;
                return 'scs_newBlockFilter';
            case 'pending':
                args.shift();
                this.params = 1;
                return 'scs_newPendingTransactionFilter';
            default:
                return 'scs_newFilter';
        }
    };

    var newFilter = new Method({
        name: 'newFilter',
        call: newFilterCall,
        params: 1
    });

    var uninstallFilter = new Method({
        name: 'uninstallFilter',
        call: 'scs_uninstallFilter',
        params: 1
    });

    var getLogs = new Method({
        name: 'getLogs',
        call: 'scs_getFilterLogs',
        params: 1
    });

    var poll = new Method({
        name: 'poll',
        call: 'scs_getFilterChanges',
        params: 1
    });

    return [
        newFilter,
        uninstallFilter,
        getLogs,
        poll
    ];
};

module.exports = {
    mc: mc,
    scs: scs
};

