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

// New part of 

"use strict";

var Property = require('../property');
var formatters = require('../formatters');

// Only keep the properties
function Txpool(chain3) {
    this._requestManager = chain3._requestManager;

    var self = this;

    // methods().forEach(function(method) {
    //     method.attachToObject(self);
    //     method.setRequestManager(self._requestManager);
    // });

    properties().forEach(function(p) {
        p.attachToObject(self);
        p.setRequestManager(self._requestManager);
    });
}

var properties = function () {
    return [
        new Property({
            name: 'content',
            getter: 'txpool_content'
        }),
        new Property({
            name: 'status',
            getter: 'txpool_status'
        }),
        new Property({
            name: 'inspect',
            getter: 'txpool_inspect'
        })
    ];
};


module.exports = Txpool;
