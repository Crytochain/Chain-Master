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
/** @file mc.js
 * @authors:
 *   Marek Kotewicz <marek@ethdev.com>
 *  @date 2015
 * @modified for LBR project
 * @LBR lab
 * @date 2018
 */

var formatters = require('../formatters');
var Method = require('../method');
var utils = require('../../utils/utils');
var Property = require('../property');

var Net = function (chain3) {
    this._requestManager = chain3._requestManager;

    var self = this;

    properties().forEach(function(p) { 
        p.attachToObject(self);
        p.setRequestManager(chain3._requestManager);
    });

    methods().forEach(function(method) {
        method.attachToObject(self);
        method.setRequestManager(chain3._requestManager);
    });

};

/// @returns an array of objects describing chain3.net api properties
var properties = function () {
    return [
        new Property({
            name: 'listening',
            getter: 'net_listening'
        }),
        new Property({
            name: 'peerCount',
            getter: 'net_peerCount',
            outputFormatter: utils.toDecimal
        })
    ];
};

var methods = function() {
    var getVnodes = new Method({
        name: 'getVnodes',
        call: 'net_getVnodes',
        params: 0,
        outputFormatter: formatters.outputVnodesFormatter
    });

    return [getVnodes];
}

module.exports = Net;
