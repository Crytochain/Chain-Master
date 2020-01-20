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
 * @file vnode.js
 * Display the vnodeconfig.json content
 * This is not a public service at this moment.
 * @LBR lab
 * @date 2018
 *
 */

"use strict";

var Method = require('../method');
var Property = require('../property');

var Vnode = function (chain3) {
    this._requestManager = chain3._requestManager;

    var self = this;
    
    //Save the properties
    properties().forEach(function(p) {
        p.attachToObject(self);
        p.setRequestManager(self._requestManager);
    });
};


var properties = function () {
    return [
        new Property({
            name: 'showToPublic',
            getter: 'vnode_showToPublic'
        }),
        new Property({
            name: 'ip',
            getter: 'vnode_vnodeIP'
        }),
        new Property({
            name: 'serviceCfg',
            getter: 'vnode_serviceCfg'
        }),
        new Property({
            name: 'address',
            getter: 'vnode_address'

        }),
        new Property({
            name: 'scsService',
            getter: 'vnode_scsService'

        })
    ];
};

// var methods = function() {

//     var getAddress = new Method({
//         name: 'getAddress',
//         call: 'vnode_getAddress',
//         params: 0
//     });
// }

module.exports = Vnode;
