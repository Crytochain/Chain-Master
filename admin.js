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

// New management APIs for VNODE
// need to enable the admin in the VNODE 
var Method = require('../method');
var Property = require('../property');
var formatters = require('../formatters');

function Admin(chain3) {
    this._requestManager = chain3._requestManager;

    var self = this;

    methods().forEach(function(method) {
        method.attachToObject(self);
        method.setRequestManager(self._requestManager);
    });

    // 
    properties().forEach(function(p) {
        p.attachToObject(self);
        p.setRequestManager(self._requestManager);
    });
}

var methods = function () {

    var addPeer = new Method({
        name: 'addPeer',
        call: 'admin_addPeer',
        params: 1
    });

    // Inputs are the paprameters
    // host: network interface to open the listener socket on (defaults to "localhost")
    // port: network port to open the listener socket on (defaults to 8546)
    // cors: cross-origin resource sharing header to use (defaults to "")
    // apis: API modules to offer over this interface (defaults to "mc,net,chain3")
    // e.g. admin.startRPC("127.0.0.1", 8545)
    var startRPC = new Method({
        name: 'startRPC',
        call: 'admin_startRPC',
        params: 4,
        inputFormatter: [null, null, null, null]
    });

    var stopRPC = new Method({
        name: 'stopRPC',
        call: 'admin_stopRPC'
    });

    // Input as signed Transaction
    var startWS = new Method({
        name: 'startWS',
		call: 'admin_startWS',
		params: 4,
        inputFormatter: [null, null, null, null]
    });

    // Input as signed Transaction
    var stopWS = new Method({
        name: 'stopWS',
        call: 'admin_stopWS'
    });

    return [
        addPeer,
        startRPC,
        stopRPC,
        startWS,
        stopWS
    ];
};

var properties = function () {
    return [
        new Property({
            name: 'datadir',
            getter: 'admin_datadir'
        }), 
        new Property({
            name: 'peers',
            getter: 'admin_peers'
        }),
        new Property({
            name: 'nodeInfo',
            getter: 'admin_nodeInfo'
        })
    ];
};


module.exports = Admin;
