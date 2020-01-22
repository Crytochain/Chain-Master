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
 * @file microevents.js
 * @author
 * @events for LBR microchain event
 * @LBR lab
 * @date 2019
 */

var sha3 = require('../utils/sha3');
var MicroEvent = require('./microevent');
var formatters = require('../chain3/formatters');
var utils = require('../utils/utils');
var Filter = require('../chain3/filter');
var watches = require('../chain3/methods/watches');

var MicroSolidityEvents = function (requestManager, json, address) {
    this._requestManager = requestManager;
    this._json = json;
    this._address = address;
};

MicroSolidityEvents.prototype.encode = function (options) {
    options = options || {};
    var result = {};

    ['fromBlock', 'toBlock'].filter(function (f) {
        return options[f] !== undefined;
    }).forEach(function (f) {
        result[f] = formatters.inputBlockNumberFormatter(options[f]);
    });

    result.address = this._address;

    return result;
};

MicroSolidityEvents.prototype.decode = function (data) {
    data.data = data.data || '';

       var eventTopic = (utils.isArray(data.topics) && utils.isString(data.topics[0])) ? data.topics[0].slice(2) : '';
    var match = this._json.filter(function (j) {
        return eventTopic === sha3(utils.transformToFullName(j));
    })[0];

    if (!match) { // cannot find matching event?
        return formatters.outputLogFormatter(data);
    }

    var event = new MicroEvent(this._requestManager, match, this._address);
 
    return event.decode(data);
};

MicroSolidityEvents.prototype.execute = function (options, callback) {

    if (utils.isFunction(arguments[arguments.length - 1])) {
        callback = arguments[arguments.length - 1];
        if(arguments.length === 1)
            options = null;
    }

    var o = this.encode(options);
    var formatter = this.decode.bind(this);
    return new Filter(o, 'scs', this._requestManager, watches.scs(), formatter, callback);
};

MicroSolidityEvents.prototype.attachToContract = function (contract) {
    var execute = this.execute.bind(this);
    contract.allEvents = execute;
};

module.exports = MicroSolidityEvents;

