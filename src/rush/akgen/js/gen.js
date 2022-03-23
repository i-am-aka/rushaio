import Flatted from './flatted.js';

import document_ from './document.js';
document_.getElementsByTagName = () => [];
document_.documentElement = { getAttribute: () => null };

import navigator_ from './navigator.js';
navigator_.permissions.query = () => Promise.resolve();

import window_ from './window.js';
window_.setInterval = globalThis.setInterval;
window_.setTimeout = globalThis.setTimeout;
window_.document = document_;
window_.navigator = navigator_;

// import Sensor from './sensor-16.js';
// import Sensor from './1.63_deob.js';
import Sensor from './1636.js';

export async function gen(tOffset, abck) {
  var cookie = '_abck='+abck;
  var mDocument = new Proxy(document_, {
    get: function(obj, prop) {
      if (prop === 'cookie') {
        return cookie;
      } else {
        return Reflect.get(...arguments);
      }
    }
  })
  // var mWindow = {};
  // Object.assign(mWindow, window_);
  var dateWithOffset = new Proxy(Date, {
    get: function(obj, prop) {
      if (prop === 'now') {
        return function() {
          return Date.now() - tOffset;
        }
      } else {
        return Reflect.get(...arguments);
      }
    }
  });

  var timeouts = {};
  var intervals = {};
  var setTimeout_ = function() {
    if (arguments[0] === 0) {
      arguments[0] = 10;
    }
    var timeout = setTimeout(...arguments);
    timeouts[timeout] = 1;
    return timeout;
  }
  var setInterval_ = function() {
    if (arguments[0] === 0) {
      arguments[0] = 10;
    }
    var interval = setInterval(...arguments);
    intervals[interval] = 1;
    return interval;
  }
  var clearTimeout_ = function(x) {
     delete timeouts[x];
    if (timeouts[x]) {
      console.log('clearTimeout', {x});
      clearTimeout(x);
    }
  }
  var clearInterval_ = function(x) {
     delete intervals[x];
    if (intervals[x]) {
      console.log('clearInterval', {x});
      clearInterval(x);
    }
  }

  var sensorPromiseResolve;
  window_.screen.availWidth = 0;
  window_.screen.availHeight = 0;
  window_.screen.width = 0;
  window_.screen.height = 0;
  var sensorPromise = new Promise(function(resolve, reject) {
    var timeout = setTimeout(() => { reject(new Error('timed out'))}, 10000);
    sensorPromiseResolve = (dat) => { resolve(dat); clearTimeout(timeout); };
  });
  var bmak = Sensor(window_, mDocument, navigator_, window_.screen, dateWithOffset, sensorPromiseResolve, setTimeout_, setInterval_, clearTimeout_, clearInterval_);
  try {
    var tel = await sensorPromise;
    // var tel = bmak.get_telemetry();
    // console.log(bmak)

    return tel;
  } finally {
    await delay(10);
    Object.keys(timeouts).forEach(clearTimeout_);
    Object.keys(intervals).forEach(clearInterval_);
  }
}

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  })
}