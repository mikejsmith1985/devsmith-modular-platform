function _mergeNamespaces(n2, m2) {
  for (var i = 0; i < m2.length; i++) {
    const e = m2[i];
    if (typeof e !== "string" && !Array.isArray(e)) {
      for (const k2 in e) {
        if (k2 !== "default" && !(k2 in n2)) {
          const d = Object.getOwnPropertyDescriptor(e, k2);
          if (d) {
            Object.defineProperty(n2, k2, d.get ? d : {
              enumerable: true,
              get: () => e[k2]
            });
          }
        }
      }
    }
  }
  return Object.freeze(Object.defineProperty(n2, Symbol.toStringTag, { value: "Module" }));
}
(function polyfill() {
  const relList = document.createElement("link").relList;
  if (relList && relList.supports && relList.supports("modulepreload")) {
    return;
  }
  for (const link of document.querySelectorAll('link[rel="modulepreload"]')) {
    processPreload(link);
  }
  new MutationObserver((mutations) => {
    for (const mutation of mutations) {
      if (mutation.type !== "childList") {
        continue;
      }
      for (const node of mutation.addedNodes) {
        if (node.tagName === "LINK" && node.rel === "modulepreload")
          processPreload(node);
      }
    }
  }).observe(document, { childList: true, subtree: true });
  function getFetchOpts(link) {
    const fetchOpts = {};
    if (link.integrity) fetchOpts.integrity = link.integrity;
    if (link.referrerPolicy) fetchOpts.referrerPolicy = link.referrerPolicy;
    if (link.crossOrigin === "use-credentials")
      fetchOpts.credentials = "include";
    else if (link.crossOrigin === "anonymous") fetchOpts.credentials = "omit";
    else fetchOpts.credentials = "same-origin";
    return fetchOpts;
  }
  function processPreload(link) {
    if (link.ep)
      return;
    link.ep = true;
    const fetchOpts = getFetchOpts(link);
    fetch(link.href, fetchOpts);
  }
})();
function getDefaultExportFromCjs(x2) {
  return x2 && x2.__esModule && Object.prototype.hasOwnProperty.call(x2, "default") ? x2["default"] : x2;
}
var jsxRuntime = { exports: {} };
var reactJsxRuntime_production_min = {};
var react = { exports: {} };
var react_production_min = {};
/**
 * @license React
 * react.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
var l$2 = Symbol.for("react.element"), n$1 = Symbol.for("react.portal"), p$2 = Symbol.for("react.fragment"), q$1 = Symbol.for("react.strict_mode"), r = Symbol.for("react.profiler"), t = Symbol.for("react.provider"), u = Symbol.for("react.context"), v$2 = Symbol.for("react.forward_ref"), w = Symbol.for("react.suspense"), x = Symbol.for("react.memo"), y = Symbol.for("react.lazy"), z$1 = Symbol.iterator;
function A$1(a) {
  if (null === a || "object" !== typeof a) return null;
  a = z$1 && a[z$1] || a["@@iterator"];
  return "function" === typeof a ? a : null;
}
var B$1 = { isMounted: function() {
  return false;
}, enqueueForceUpdate: function() {
}, enqueueReplaceState: function() {
}, enqueueSetState: function() {
} }, C$1 = Object.assign, D$2 = {};
function E$1(a, b, e) {
  this.props = a;
  this.context = b;
  this.refs = D$2;
  this.updater = e || B$1;
}
E$1.prototype.isReactComponent = {};
E$1.prototype.setState = function(a, b) {
  if ("object" !== typeof a && "function" !== typeof a && null != a) throw Error("setState(...): takes an object of state variables to update or a function which returns an object of state variables.");
  this.updater.enqueueSetState(this, a, b, "setState");
};
E$1.prototype.forceUpdate = function(a) {
  this.updater.enqueueForceUpdate(this, a, "forceUpdate");
};
function F() {
}
F.prototype = E$1.prototype;
function G$1(a, b, e) {
  this.props = a;
  this.context = b;
  this.refs = D$2;
  this.updater = e || B$1;
}
var H$2 = G$1.prototype = new F();
H$2.constructor = G$1;
C$1(H$2, E$1.prototype);
H$2.isPureReactComponent = true;
var I$1 = Array.isArray, J = Object.prototype.hasOwnProperty, K$1 = { current: null }, L$1 = { key: true, ref: true, __self: true, __source: true };
function M$1(a, b, e) {
  var d, c = {}, k2 = null, h2 = null;
  if (null != b) for (d in void 0 !== b.ref && (h2 = b.ref), void 0 !== b.key && (k2 = "" + b.key), b) J.call(b, d) && !L$1.hasOwnProperty(d) && (c[d] = b[d]);
  var g = arguments.length - 2;
  if (1 === g) c.children = e;
  else if (1 < g) {
    for (var f2 = Array(g), m2 = 0; m2 < g; m2++) f2[m2] = arguments[m2 + 2];
    c.children = f2;
  }
  if (a && a.defaultProps) for (d in g = a.defaultProps, g) void 0 === c[d] && (c[d] = g[d]);
  return { $$typeof: l$2, type: a, key: k2, ref: h2, props: c, _owner: K$1.current };
}
function N$1(a, b) {
  return { $$typeof: l$2, type: a.type, key: b, ref: a.ref, props: a.props, _owner: a._owner };
}
function O$1(a) {
  return "object" === typeof a && null !== a && a.$$typeof === l$2;
}
function escape(a) {
  var b = { "=": "=0", ":": "=2" };
  return "$" + a.replace(/[=:]/g, function(a2) {
    return b[a2];
  });
}
var P$1 = /\/+/g;
function Q$1(a, b) {
  return "object" === typeof a && null !== a && null != a.key ? escape("" + a.key) : b.toString(36);
}
function R$1(a, b, e, d, c) {
  var k2 = typeof a;
  if ("undefined" === k2 || "boolean" === k2) a = null;
  var h2 = false;
  if (null === a) h2 = true;
  else switch (k2) {
    case "string":
    case "number":
      h2 = true;
      break;
    case "object":
      switch (a.$$typeof) {
        case l$2:
        case n$1:
          h2 = true;
      }
  }
  if (h2) return h2 = a, c = c(h2), a = "" === d ? "." + Q$1(h2, 0) : d, I$1(c) ? (e = "", null != a && (e = a.replace(P$1, "$&/") + "/"), R$1(c, b, e, "", function(a2) {
    return a2;
  })) : null != c && (O$1(c) && (c = N$1(c, e + (!c.key || h2 && h2.key === c.key ? "" : ("" + c.key).replace(P$1, "$&/") + "/") + a)), b.push(c)), 1;
  h2 = 0;
  d = "" === d ? "." : d + ":";
  if (I$1(a)) for (var g = 0; g < a.length; g++) {
    k2 = a[g];
    var f2 = d + Q$1(k2, g);
    h2 += R$1(k2, b, e, f2, c);
  }
  else if (f2 = A$1(a), "function" === typeof f2) for (a = f2.call(a), g = 0; !(k2 = a.next()).done; ) k2 = k2.value, f2 = d + Q$1(k2, g++), h2 += R$1(k2, b, e, f2, c);
  else if ("object" === k2) throw b = String(a), Error("Objects are not valid as a React child (found: " + ("[object Object]" === b ? "object with keys {" + Object.keys(a).join(", ") + "}" : b) + "). If you meant to render a collection of children, use an array instead.");
  return h2;
}
function S$1(a, b, e) {
  if (null == a) return a;
  var d = [], c = 0;
  R$1(a, d, "", "", function(a2) {
    return b.call(e, a2, c++);
  });
  return d;
}
function T$1(a) {
  if (-1 === a._status) {
    var b = a._result;
    b = b();
    b.then(function(b2) {
      if (0 === a._status || -1 === a._status) a._status = 1, a._result = b2;
    }, function(b2) {
      if (0 === a._status || -1 === a._status) a._status = 2, a._result = b2;
    });
    -1 === a._status && (a._status = 0, a._result = b);
  }
  if (1 === a._status) return a._result.default;
  throw a._result;
}
var U$1 = { current: null }, V$1 = { transition: null }, W$1 = { ReactCurrentDispatcher: U$1, ReactCurrentBatchConfig: V$1, ReactCurrentOwner: K$1 };
function X$1() {
  throw Error("act(...) is not supported in production builds of React.");
}
react_production_min.Children = { map: S$1, forEach: function(a, b, e) {
  S$1(a, function() {
    b.apply(this, arguments);
  }, e);
}, count: function(a) {
  var b = 0;
  S$1(a, function() {
    b++;
  });
  return b;
}, toArray: function(a) {
  return S$1(a, function(a2) {
    return a2;
  }) || [];
}, only: function(a) {
  if (!O$1(a)) throw Error("React.Children.only expected to receive a single React element child.");
  return a;
} };
react_production_min.Component = E$1;
react_production_min.Fragment = p$2;
react_production_min.Profiler = r;
react_production_min.PureComponent = G$1;
react_production_min.StrictMode = q$1;
react_production_min.Suspense = w;
react_production_min.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED = W$1;
react_production_min.act = X$1;
react_production_min.cloneElement = function(a, b, e) {
  if (null === a || void 0 === a) throw Error("React.cloneElement(...): The argument must be a React element, but you passed " + a + ".");
  var d = C$1({}, a.props), c = a.key, k2 = a.ref, h2 = a._owner;
  if (null != b) {
    void 0 !== b.ref && (k2 = b.ref, h2 = K$1.current);
    void 0 !== b.key && (c = "" + b.key);
    if (a.type && a.type.defaultProps) var g = a.type.defaultProps;
    for (f2 in b) J.call(b, f2) && !L$1.hasOwnProperty(f2) && (d[f2] = void 0 === b[f2] && void 0 !== g ? g[f2] : b[f2]);
  }
  var f2 = arguments.length - 2;
  if (1 === f2) d.children = e;
  else if (1 < f2) {
    g = Array(f2);
    for (var m2 = 0; m2 < f2; m2++) g[m2] = arguments[m2 + 2];
    d.children = g;
  }
  return { $$typeof: l$2, type: a.type, key: c, ref: k2, props: d, _owner: h2 };
};
react_production_min.createContext = function(a) {
  a = { $$typeof: u, _currentValue: a, _currentValue2: a, _threadCount: 0, Provider: null, Consumer: null, _defaultValue: null, _globalName: null };
  a.Provider = { $$typeof: t, _context: a };
  return a.Consumer = a;
};
react_production_min.createElement = M$1;
react_production_min.createFactory = function(a) {
  var b = M$1.bind(null, a);
  b.type = a;
  return b;
};
react_production_min.createRef = function() {
  return { current: null };
};
react_production_min.forwardRef = function(a) {
  return { $$typeof: v$2, render: a };
};
react_production_min.isValidElement = O$1;
react_production_min.lazy = function(a) {
  return { $$typeof: y, _payload: { _status: -1, _result: a }, _init: T$1 };
};
react_production_min.memo = function(a, b) {
  return { $$typeof: x, type: a, compare: void 0 === b ? null : b };
};
react_production_min.startTransition = function(a) {
  var b = V$1.transition;
  V$1.transition = {};
  try {
    a();
  } finally {
    V$1.transition = b;
  }
};
react_production_min.unstable_act = X$1;
react_production_min.useCallback = function(a, b) {
  return U$1.current.useCallback(a, b);
};
react_production_min.useContext = function(a) {
  return U$1.current.useContext(a);
};
react_production_min.useDebugValue = function() {
};
react_production_min.useDeferredValue = function(a) {
  return U$1.current.useDeferredValue(a);
};
react_production_min.useEffect = function(a, b) {
  return U$1.current.useEffect(a, b);
};
react_production_min.useId = function() {
  return U$1.current.useId();
};
react_production_min.useImperativeHandle = function(a, b, e) {
  return U$1.current.useImperativeHandle(a, b, e);
};
react_production_min.useInsertionEffect = function(a, b) {
  return U$1.current.useInsertionEffect(a, b);
};
react_production_min.useLayoutEffect = function(a, b) {
  return U$1.current.useLayoutEffect(a, b);
};
react_production_min.useMemo = function(a, b) {
  return U$1.current.useMemo(a, b);
};
react_production_min.useReducer = function(a, b, e) {
  return U$1.current.useReducer(a, b, e);
};
react_production_min.useRef = function(a) {
  return U$1.current.useRef(a);
};
react_production_min.useState = function(a) {
  return U$1.current.useState(a);
};
react_production_min.useSyncExternalStore = function(a, b, e) {
  return U$1.current.useSyncExternalStore(a, b, e);
};
react_production_min.useTransition = function() {
  return U$1.current.useTransition();
};
react_production_min.version = "18.3.1";
{
  react.exports = react_production_min;
}
var reactExports = react.exports;
const We$1 = /* @__PURE__ */ getDefaultExportFromCjs(reactExports);
const React = /* @__PURE__ */ _mergeNamespaces({
  __proto__: null,
  default: We$1
}, [reactExports]);
/**
 * @license React
 * react-jsx-runtime.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
var f = reactExports, k$1 = Symbol.for("react.element"), l$1 = Symbol.for("react.fragment"), m$1 = Object.prototype.hasOwnProperty, n = f.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED.ReactCurrentOwner, p$1 = { key: true, ref: true, __self: true, __source: true };
function q(c, a, g) {
  var b, d = {}, e = null, h2 = null;
  void 0 !== g && (e = "" + g);
  void 0 !== a.key && (e = "" + a.key);
  void 0 !== a.ref && (h2 = a.ref);
  for (b in a) m$1.call(a, b) && !p$1.hasOwnProperty(b) && (d[b] = a[b]);
  if (c && c.defaultProps) for (b in a = c.defaultProps, a) void 0 === d[b] && (d[b] = a[b]);
  return { $$typeof: k$1, type: c, key: e, ref: h2, props: d, _owner: n.current };
}
reactJsxRuntime_production_min.Fragment = l$1;
reactJsxRuntime_production_min.jsx = q;
reactJsxRuntime_production_min.jsxs = q;
{
  jsxRuntime.exports = reactJsxRuntime_production_min;
}
var jsxRuntimeExports = jsxRuntime.exports;
var reactDom = { exports: {} };
var reactDom_production_min = {};
var scheduler = { exports: {} };
var scheduler_production_min = {};
/**
 * @license React
 * scheduler.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
(function(exports) {
  function f2(a, b) {
    var c = a.length;
    a.push(b);
    a: for (; 0 < c; ) {
      var d = c - 1 >>> 1, e = a[d];
      if (0 < g(e, b)) a[d] = b, a[c] = e, c = d;
      else break a;
    }
  }
  function h2(a) {
    return 0 === a.length ? null : a[0];
  }
  function k2(a) {
    if (0 === a.length) return null;
    var b = a[0], c = a.pop();
    if (c !== b) {
      a[0] = c;
      a: for (var d = 0, e = a.length, w2 = e >>> 1; d < w2; ) {
        var m2 = 2 * (d + 1) - 1, C2 = a[m2], n2 = m2 + 1, x2 = a[n2];
        if (0 > g(C2, c)) n2 < e && 0 > g(x2, C2) ? (a[d] = x2, a[n2] = c, d = n2) : (a[d] = C2, a[m2] = c, d = m2);
        else if (n2 < e && 0 > g(x2, c)) a[d] = x2, a[n2] = c, d = n2;
        else break a;
      }
    }
    return b;
  }
  function g(a, b) {
    var c = a.sortIndex - b.sortIndex;
    return 0 !== c ? c : a.id - b.id;
  }
  if ("object" === typeof performance && "function" === typeof performance.now) {
    var l2 = performance;
    exports.unstable_now = function() {
      return l2.now();
    };
  } else {
    var p2 = Date, q2 = p2.now();
    exports.unstable_now = function() {
      return p2.now() - q2;
    };
  }
  var r2 = [], t2 = [], u2 = 1, v2 = null, y2 = 3, z2 = false, A2 = false, B2 = false, D2 = "function" === typeof setTimeout ? setTimeout : null, E2 = "function" === typeof clearTimeout ? clearTimeout : null, F2 = "undefined" !== typeof setImmediate ? setImmediate : null;
  "undefined" !== typeof navigator && void 0 !== navigator.scheduling && void 0 !== navigator.scheduling.isInputPending && navigator.scheduling.isInputPending.bind(navigator.scheduling);
  function G2(a) {
    for (var b = h2(t2); null !== b; ) {
      if (null === b.callback) k2(t2);
      else if (b.startTime <= a) k2(t2), b.sortIndex = b.expirationTime, f2(r2, b);
      else break;
      b = h2(t2);
    }
  }
  function H2(a) {
    B2 = false;
    G2(a);
    if (!A2) if (null !== h2(r2)) A2 = true, I2(J2);
    else {
      var b = h2(t2);
      null !== b && K2(H2, b.startTime - a);
    }
  }
  function J2(a, b) {
    A2 = false;
    B2 && (B2 = false, E2(L2), L2 = -1);
    z2 = true;
    var c = y2;
    try {
      G2(b);
      for (v2 = h2(r2); null !== v2 && (!(v2.expirationTime > b) || a && !M2()); ) {
        var d = v2.callback;
        if ("function" === typeof d) {
          v2.callback = null;
          y2 = v2.priorityLevel;
          var e = d(v2.expirationTime <= b);
          b = exports.unstable_now();
          "function" === typeof e ? v2.callback = e : v2 === h2(r2) && k2(r2);
          G2(b);
        } else k2(r2);
        v2 = h2(r2);
      }
      if (null !== v2) var w2 = true;
      else {
        var m2 = h2(t2);
        null !== m2 && K2(H2, m2.startTime - b);
        w2 = false;
      }
      return w2;
    } finally {
      v2 = null, y2 = c, z2 = false;
    }
  }
  var N2 = false, O2 = null, L2 = -1, P2 = 5, Q2 = -1;
  function M2() {
    return exports.unstable_now() - Q2 < P2 ? false : true;
  }
  function R2() {
    if (null !== O2) {
      var a = exports.unstable_now();
      Q2 = a;
      var b = true;
      try {
        b = O2(true, a);
      } finally {
        b ? S2() : (N2 = false, O2 = null);
      }
    } else N2 = false;
  }
  var S2;
  if ("function" === typeof F2) S2 = function() {
    F2(R2);
  };
  else if ("undefined" !== typeof MessageChannel) {
    var T2 = new MessageChannel(), U2 = T2.port2;
    T2.port1.onmessage = R2;
    S2 = function() {
      U2.postMessage(null);
    };
  } else S2 = function() {
    D2(R2, 0);
  };
  function I2(a) {
    O2 = a;
    N2 || (N2 = true, S2());
  }
  function K2(a, b) {
    L2 = D2(function() {
      a(exports.unstable_now());
    }, b);
  }
  exports.unstable_IdlePriority = 5;
  exports.unstable_ImmediatePriority = 1;
  exports.unstable_LowPriority = 4;
  exports.unstable_NormalPriority = 3;
  exports.unstable_Profiling = null;
  exports.unstable_UserBlockingPriority = 2;
  exports.unstable_cancelCallback = function(a) {
    a.callback = null;
  };
  exports.unstable_continueExecution = function() {
    A2 || z2 || (A2 = true, I2(J2));
  };
  exports.unstable_forceFrameRate = function(a) {
    0 > a || 125 < a ? console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported") : P2 = 0 < a ? Math.floor(1e3 / a) : 5;
  };
  exports.unstable_getCurrentPriorityLevel = function() {
    return y2;
  };
  exports.unstable_getFirstCallbackNode = function() {
    return h2(r2);
  };
  exports.unstable_next = function(a) {
    switch (y2) {
      case 1:
      case 2:
      case 3:
        var b = 3;
        break;
      default:
        b = y2;
    }
    var c = y2;
    y2 = b;
    try {
      return a();
    } finally {
      y2 = c;
    }
  };
  exports.unstable_pauseExecution = function() {
  };
  exports.unstable_requestPaint = function() {
  };
  exports.unstable_runWithPriority = function(a, b) {
    switch (a) {
      case 1:
      case 2:
      case 3:
      case 4:
      case 5:
        break;
      default:
        a = 3;
    }
    var c = y2;
    y2 = a;
    try {
      return b();
    } finally {
      y2 = c;
    }
  };
  exports.unstable_scheduleCallback = function(a, b, c) {
    var d = exports.unstable_now();
    "object" === typeof c && null !== c ? (c = c.delay, c = "number" === typeof c && 0 < c ? d + c : d) : c = d;
    switch (a) {
      case 1:
        var e = -1;
        break;
      case 2:
        e = 250;
        break;
      case 5:
        e = 1073741823;
        break;
      case 4:
        e = 1e4;
        break;
      default:
        e = 5e3;
    }
    e = c + e;
    a = { id: u2++, callback: b, priorityLevel: a, startTime: c, expirationTime: e, sortIndex: -1 };
    c > d ? (a.sortIndex = c, f2(t2, a), null === h2(r2) && a === h2(t2) && (B2 ? (E2(L2), L2 = -1) : B2 = true, K2(H2, c - d))) : (a.sortIndex = e, f2(r2, a), A2 || z2 || (A2 = true, I2(J2)));
    return a;
  };
  exports.unstable_shouldYield = M2;
  exports.unstable_wrapCallback = function(a) {
    var b = y2;
    return function() {
      var c = y2;
      y2 = b;
      try {
        return a.apply(this, arguments);
      } finally {
        y2 = c;
      }
    };
  };
})(scheduler_production_min);
{
  scheduler.exports = scheduler_production_min;
}
var schedulerExports = scheduler.exports;
/**
 * @license React
 * react-dom.production.min.js
 *
 * Copyright (c) Facebook, Inc. and its affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */
var aa = reactExports, ca = schedulerExports;
function p(a) {
  for (var b = "https://reactjs.org/docs/error-decoder.html?invariant=" + a, c = 1; c < arguments.length; c++) b += "&args[]=" + encodeURIComponent(arguments[c]);
  return "Minified React error #" + a + "; visit " + b + " for the full message or use the non-minified dev environment for full errors and additional helpful warnings.";
}
var da = /* @__PURE__ */ new Set(), ea = {};
function fa(a, b) {
  ha(a, b);
  ha(a + "Capture", b);
}
function ha(a, b) {
  ea[a] = b;
  for (a = 0; a < b.length; a++) da.add(b[a]);
}
var ia = !("undefined" === typeof window || "undefined" === typeof window.document || "undefined" === typeof window.document.createElement), ja = Object.prototype.hasOwnProperty, ka = /^[:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD][:A-Z_a-z\u00C0-\u00D6\u00D8-\u00F6\u00F8-\u02FF\u0370-\u037D\u037F-\u1FFF\u200C-\u200D\u2070-\u218F\u2C00-\u2FEF\u3001-\uD7FF\uF900-\uFDCF\uFDF0-\uFFFD\-.0-9\u00B7\u0300-\u036F\u203F-\u2040]*$/, la = {}, ma = {};
function oa(a) {
  if (ja.call(ma, a)) return true;
  if (ja.call(la, a)) return false;
  if (ka.test(a)) return ma[a] = true;
  la[a] = true;
  return false;
}
function pa(a, b, c, d) {
  if (null !== c && 0 === c.type) return false;
  switch (typeof b) {
    case "function":
    case "symbol":
      return true;
    case "boolean":
      if (d) return false;
      if (null !== c) return !c.acceptsBooleans;
      a = a.toLowerCase().slice(0, 5);
      return "data-" !== a && "aria-" !== a;
    default:
      return false;
  }
}
function qa(a, b, c, d) {
  if (null === b || "undefined" === typeof b || pa(a, b, c, d)) return true;
  if (d) return false;
  if (null !== c) switch (c.type) {
    case 3:
      return !b;
    case 4:
      return false === b;
    case 5:
      return isNaN(b);
    case 6:
      return isNaN(b) || 1 > b;
  }
  return false;
}
function v$1(a, b, c, d, e, f2, g) {
  this.acceptsBooleans = 2 === b || 3 === b || 4 === b;
  this.attributeName = d;
  this.attributeNamespace = e;
  this.mustUseProperty = c;
  this.propertyName = a;
  this.type = b;
  this.sanitizeURL = f2;
  this.removeEmptyString = g;
}
var z = {};
"children dangerouslySetInnerHTML defaultValue defaultChecked innerHTML suppressContentEditableWarning suppressHydrationWarning style".split(" ").forEach(function(a) {
  z[a] = new v$1(a, 0, false, a, null, false, false);
});
[["acceptCharset", "accept-charset"], ["className", "class"], ["htmlFor", "for"], ["httpEquiv", "http-equiv"]].forEach(function(a) {
  var b = a[0];
  z[b] = new v$1(b, 1, false, a[1], null, false, false);
});
["contentEditable", "draggable", "spellCheck", "value"].forEach(function(a) {
  z[a] = new v$1(a, 2, false, a.toLowerCase(), null, false, false);
});
["autoReverse", "externalResourcesRequired", "focusable", "preserveAlpha"].forEach(function(a) {
  z[a] = new v$1(a, 2, false, a, null, false, false);
});
"allowFullScreen async autoFocus autoPlay controls default defer disabled disablePictureInPicture disableRemotePlayback formNoValidate hidden loop noModule noValidate open playsInline readOnly required reversed scoped seamless itemScope".split(" ").forEach(function(a) {
  z[a] = new v$1(a, 3, false, a.toLowerCase(), null, false, false);
});
["checked", "multiple", "muted", "selected"].forEach(function(a) {
  z[a] = new v$1(a, 3, true, a, null, false, false);
});
["capture", "download"].forEach(function(a) {
  z[a] = new v$1(a, 4, false, a, null, false, false);
});
["cols", "rows", "size", "span"].forEach(function(a) {
  z[a] = new v$1(a, 6, false, a, null, false, false);
});
["rowSpan", "start"].forEach(function(a) {
  z[a] = new v$1(a, 5, false, a.toLowerCase(), null, false, false);
});
var ra = /[\-:]([a-z])/g;
function sa(a) {
  return a[1].toUpperCase();
}
"accent-height alignment-baseline arabic-form baseline-shift cap-height clip-path clip-rule color-interpolation color-interpolation-filters color-profile color-rendering dominant-baseline enable-background fill-opacity fill-rule flood-color flood-opacity font-family font-size font-size-adjust font-stretch font-style font-variant font-weight glyph-name glyph-orientation-horizontal glyph-orientation-vertical horiz-adv-x horiz-origin-x image-rendering letter-spacing lighting-color marker-end marker-mid marker-start overline-position overline-thickness paint-order panose-1 pointer-events rendering-intent shape-rendering stop-color stop-opacity strikethrough-position strikethrough-thickness stroke-dasharray stroke-dashoffset stroke-linecap stroke-linejoin stroke-miterlimit stroke-opacity stroke-width text-anchor text-decoration text-rendering underline-position underline-thickness unicode-bidi unicode-range units-per-em v-alphabetic v-hanging v-ideographic v-mathematical vector-effect vert-adv-y vert-origin-x vert-origin-y word-spacing writing-mode xmlns:xlink x-height".split(" ").forEach(function(a) {
  var b = a.replace(
    ra,
    sa
  );
  z[b] = new v$1(b, 1, false, a, null, false, false);
});
"xlink:actuate xlink:arcrole xlink:role xlink:show xlink:title xlink:type".split(" ").forEach(function(a) {
  var b = a.replace(ra, sa);
  z[b] = new v$1(b, 1, false, a, "http://www.w3.org/1999/xlink", false, false);
});
["xml:base", "xml:lang", "xml:space"].forEach(function(a) {
  var b = a.replace(ra, sa);
  z[b] = new v$1(b, 1, false, a, "http://www.w3.org/XML/1998/namespace", false, false);
});
["tabIndex", "crossOrigin"].forEach(function(a) {
  z[a] = new v$1(a, 1, false, a.toLowerCase(), null, false, false);
});
z.xlinkHref = new v$1("xlinkHref", 1, false, "xlink:href", "http://www.w3.org/1999/xlink", true, false);
["src", "href", "action", "formAction"].forEach(function(a) {
  z[a] = new v$1(a, 1, false, a.toLowerCase(), null, true, true);
});
function ta(a, b, c, d) {
  var e = z.hasOwnProperty(b) ? z[b] : null;
  if (null !== e ? 0 !== e.type : d || !(2 < b.length) || "o" !== b[0] && "O" !== b[0] || "n" !== b[1] && "N" !== b[1]) qa(b, c, e, d) && (c = null), d || null === e ? oa(b) && (null === c ? a.removeAttribute(b) : a.setAttribute(b, "" + c)) : e.mustUseProperty ? a[e.propertyName] = null === c ? 3 === e.type ? false : "" : c : (b = e.attributeName, d = e.attributeNamespace, null === c ? a.removeAttribute(b) : (e = e.type, c = 3 === e || 4 === e && true === c ? "" : "" + c, d ? a.setAttributeNS(d, b, c) : a.setAttribute(b, c)));
}
var ua = aa.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED, va = Symbol.for("react.element"), wa = Symbol.for("react.portal"), ya = Symbol.for("react.fragment"), za = Symbol.for("react.strict_mode"), Aa = Symbol.for("react.profiler"), Ba = Symbol.for("react.provider"), Ca = Symbol.for("react.context"), Da = Symbol.for("react.forward_ref"), Ea = Symbol.for("react.suspense"), Fa = Symbol.for("react.suspense_list"), Ga = Symbol.for("react.memo"), Ha = Symbol.for("react.lazy");
var Ia = Symbol.for("react.offscreen");
var Ja = Symbol.iterator;
function Ka(a) {
  if (null === a || "object" !== typeof a) return null;
  a = Ja && a[Ja] || a["@@iterator"];
  return "function" === typeof a ? a : null;
}
var A = Object.assign, La;
function Ma(a) {
  if (void 0 === La) try {
    throw Error();
  } catch (c) {
    var b = c.stack.trim().match(/\n( *(at )?)/);
    La = b && b[1] || "";
  }
  return "\n" + La + a;
}
var Na = false;
function Oa(a, b) {
  if (!a || Na) return "";
  Na = true;
  var c = Error.prepareStackTrace;
  Error.prepareStackTrace = void 0;
  try {
    if (b) if (b = function() {
      throw Error();
    }, Object.defineProperty(b.prototype, "props", { set: function() {
      throw Error();
    } }), "object" === typeof Reflect && Reflect.construct) {
      try {
        Reflect.construct(b, []);
      } catch (l2) {
        var d = l2;
      }
      Reflect.construct(a, [], b);
    } else {
      try {
        b.call();
      } catch (l2) {
        d = l2;
      }
      a.call(b.prototype);
    }
    else {
      try {
        throw Error();
      } catch (l2) {
        d = l2;
      }
      a();
    }
  } catch (l2) {
    if (l2 && d && "string" === typeof l2.stack) {
      for (var e = l2.stack.split("\n"), f2 = d.stack.split("\n"), g = e.length - 1, h2 = f2.length - 1; 1 <= g && 0 <= h2 && e[g] !== f2[h2]; ) h2--;
      for (; 1 <= g && 0 <= h2; g--, h2--) if (e[g] !== f2[h2]) {
        if (1 !== g || 1 !== h2) {
          do
            if (g--, h2--, 0 > h2 || e[g] !== f2[h2]) {
              var k2 = "\n" + e[g].replace(" at new ", " at ");
              a.displayName && k2.includes("<anonymous>") && (k2 = k2.replace("<anonymous>", a.displayName));
              return k2;
            }
          while (1 <= g && 0 <= h2);
        }
        break;
      }
    }
  } finally {
    Na = false, Error.prepareStackTrace = c;
  }
  return (a = a ? a.displayName || a.name : "") ? Ma(a) : "";
}
function Pa(a) {
  switch (a.tag) {
    case 5:
      return Ma(a.type);
    case 16:
      return Ma("Lazy");
    case 13:
      return Ma("Suspense");
    case 19:
      return Ma("SuspenseList");
    case 0:
    case 2:
    case 15:
      return a = Oa(a.type, false), a;
    case 11:
      return a = Oa(a.type.render, false), a;
    case 1:
      return a = Oa(a.type, true), a;
    default:
      return "";
  }
}
function Qa(a) {
  if (null == a) return null;
  if ("function" === typeof a) return a.displayName || a.name || null;
  if ("string" === typeof a) return a;
  switch (a) {
    case ya:
      return "Fragment";
    case wa:
      return "Portal";
    case Aa:
      return "Profiler";
    case za:
      return "StrictMode";
    case Ea:
      return "Suspense";
    case Fa:
      return "SuspenseList";
  }
  if ("object" === typeof a) switch (a.$$typeof) {
    case Ca:
      return (a.displayName || "Context") + ".Consumer";
    case Ba:
      return (a._context.displayName || "Context") + ".Provider";
    case Da:
      var b = a.render;
      a = a.displayName;
      a || (a = b.displayName || b.name || "", a = "" !== a ? "ForwardRef(" + a + ")" : "ForwardRef");
      return a;
    case Ga:
      return b = a.displayName || null, null !== b ? b : Qa(a.type) || "Memo";
    case Ha:
      b = a._payload;
      a = a._init;
      try {
        return Qa(a(b));
      } catch (c) {
      }
  }
  return null;
}
function Ra(a) {
  var b = a.type;
  switch (a.tag) {
    case 24:
      return "Cache";
    case 9:
      return (b.displayName || "Context") + ".Consumer";
    case 10:
      return (b._context.displayName || "Context") + ".Provider";
    case 18:
      return "DehydratedFragment";
    case 11:
      return a = b.render, a = a.displayName || a.name || "", b.displayName || ("" !== a ? "ForwardRef(" + a + ")" : "ForwardRef");
    case 7:
      return "Fragment";
    case 5:
      return b;
    case 4:
      return "Portal";
    case 3:
      return "Root";
    case 6:
      return "Text";
    case 16:
      return Qa(b);
    case 8:
      return b === za ? "StrictMode" : "Mode";
    case 22:
      return "Offscreen";
    case 12:
      return "Profiler";
    case 21:
      return "Scope";
    case 13:
      return "Suspense";
    case 19:
      return "SuspenseList";
    case 25:
      return "TracingMarker";
    case 1:
    case 0:
    case 17:
    case 2:
    case 14:
    case 15:
      if ("function" === typeof b) return b.displayName || b.name || null;
      if ("string" === typeof b) return b;
  }
  return null;
}
function Sa(a) {
  switch (typeof a) {
    case "boolean":
    case "number":
    case "string":
    case "undefined":
      return a;
    case "object":
      return a;
    default:
      return "";
  }
}
function Ta(a) {
  var b = a.type;
  return (a = a.nodeName) && "input" === a.toLowerCase() && ("checkbox" === b || "radio" === b);
}
function Ua(a) {
  var b = Ta(a) ? "checked" : "value", c = Object.getOwnPropertyDescriptor(a.constructor.prototype, b), d = "" + a[b];
  if (!a.hasOwnProperty(b) && "undefined" !== typeof c && "function" === typeof c.get && "function" === typeof c.set) {
    var e = c.get, f2 = c.set;
    Object.defineProperty(a, b, { configurable: true, get: function() {
      return e.call(this);
    }, set: function(a2) {
      d = "" + a2;
      f2.call(this, a2);
    } });
    Object.defineProperty(a, b, { enumerable: c.enumerable });
    return { getValue: function() {
      return d;
    }, setValue: function(a2) {
      d = "" + a2;
    }, stopTracking: function() {
      a._valueTracker = null;
      delete a[b];
    } };
  }
}
function Va(a) {
  a._valueTracker || (a._valueTracker = Ua(a));
}
function Wa(a) {
  if (!a) return false;
  var b = a._valueTracker;
  if (!b) return true;
  var c = b.getValue();
  var d = "";
  a && (d = Ta(a) ? a.checked ? "true" : "false" : a.value);
  a = d;
  return a !== c ? (b.setValue(a), true) : false;
}
function Xa(a) {
  a = a || ("undefined" !== typeof document ? document : void 0);
  if ("undefined" === typeof a) return null;
  try {
    return a.activeElement || a.body;
  } catch (b) {
    return a.body;
  }
}
function Ya(a, b) {
  var c = b.checked;
  return A({}, b, { defaultChecked: void 0, defaultValue: void 0, value: void 0, checked: null != c ? c : a._wrapperState.initialChecked });
}
function Za(a, b) {
  var c = null == b.defaultValue ? "" : b.defaultValue, d = null != b.checked ? b.checked : b.defaultChecked;
  c = Sa(null != b.value ? b.value : c);
  a._wrapperState = { initialChecked: d, initialValue: c, controlled: "checkbox" === b.type || "radio" === b.type ? null != b.checked : null != b.value };
}
function ab(a, b) {
  b = b.checked;
  null != b && ta(a, "checked", b, false);
}
function bb(a, b) {
  ab(a, b);
  var c = Sa(b.value), d = b.type;
  if (null != c) if ("number" === d) {
    if (0 === c && "" === a.value || a.value != c) a.value = "" + c;
  } else a.value !== "" + c && (a.value = "" + c);
  else if ("submit" === d || "reset" === d) {
    a.removeAttribute("value");
    return;
  }
  b.hasOwnProperty("value") ? cb(a, b.type, c) : b.hasOwnProperty("defaultValue") && cb(a, b.type, Sa(b.defaultValue));
  null == b.checked && null != b.defaultChecked && (a.defaultChecked = !!b.defaultChecked);
}
function db(a, b, c) {
  if (b.hasOwnProperty("value") || b.hasOwnProperty("defaultValue")) {
    var d = b.type;
    if (!("submit" !== d && "reset" !== d || void 0 !== b.value && null !== b.value)) return;
    b = "" + a._wrapperState.initialValue;
    c || b === a.value || (a.value = b);
    a.defaultValue = b;
  }
  c = a.name;
  "" !== c && (a.name = "");
  a.defaultChecked = !!a._wrapperState.initialChecked;
  "" !== c && (a.name = c);
}
function cb(a, b, c) {
  if ("number" !== b || Xa(a.ownerDocument) !== a) null == c ? a.defaultValue = "" + a._wrapperState.initialValue : a.defaultValue !== "" + c && (a.defaultValue = "" + c);
}
var eb = Array.isArray;
function fb(a, b, c, d) {
  a = a.options;
  if (b) {
    b = {};
    for (var e = 0; e < c.length; e++) b["$" + c[e]] = true;
    for (c = 0; c < a.length; c++) e = b.hasOwnProperty("$" + a[c].value), a[c].selected !== e && (a[c].selected = e), e && d && (a[c].defaultSelected = true);
  } else {
    c = "" + Sa(c);
    b = null;
    for (e = 0; e < a.length; e++) {
      if (a[e].value === c) {
        a[e].selected = true;
        d && (a[e].defaultSelected = true);
        return;
      }
      null !== b || a[e].disabled || (b = a[e]);
    }
    null !== b && (b.selected = true);
  }
}
function gb(a, b) {
  if (null != b.dangerouslySetInnerHTML) throw Error(p(91));
  return A({}, b, { value: void 0, defaultValue: void 0, children: "" + a._wrapperState.initialValue });
}
function hb(a, b) {
  var c = b.value;
  if (null == c) {
    c = b.children;
    b = b.defaultValue;
    if (null != c) {
      if (null != b) throw Error(p(92));
      if (eb(c)) {
        if (1 < c.length) throw Error(p(93));
        c = c[0];
      }
      b = c;
    }
    null == b && (b = "");
    c = b;
  }
  a._wrapperState = { initialValue: Sa(c) };
}
function ib(a, b) {
  var c = Sa(b.value), d = Sa(b.defaultValue);
  null != c && (c = "" + c, c !== a.value && (a.value = c), null == b.defaultValue && a.defaultValue !== c && (a.defaultValue = c));
  null != d && (a.defaultValue = "" + d);
}
function jb(a) {
  var b = a.textContent;
  b === a._wrapperState.initialValue && "" !== b && null !== b && (a.value = b);
}
function kb(a) {
  switch (a) {
    case "svg":
      return "http://www.w3.org/2000/svg";
    case "math":
      return "http://www.w3.org/1998/Math/MathML";
    default:
      return "http://www.w3.org/1999/xhtml";
  }
}
function lb(a, b) {
  return null == a || "http://www.w3.org/1999/xhtml" === a ? kb(b) : "http://www.w3.org/2000/svg" === a && "foreignObject" === b ? "http://www.w3.org/1999/xhtml" : a;
}
var mb, nb = function(a) {
  return "undefined" !== typeof MSApp && MSApp.execUnsafeLocalFunction ? function(b, c, d, e) {
    MSApp.execUnsafeLocalFunction(function() {
      return a(b, c, d, e);
    });
  } : a;
}(function(a, b) {
  if ("http://www.w3.org/2000/svg" !== a.namespaceURI || "innerHTML" in a) a.innerHTML = b;
  else {
    mb = mb || document.createElement("div");
    mb.innerHTML = "<svg>" + b.valueOf().toString() + "</svg>";
    for (b = mb.firstChild; a.firstChild; ) a.removeChild(a.firstChild);
    for (; b.firstChild; ) a.appendChild(b.firstChild);
  }
});
function ob(a, b) {
  if (b) {
    var c = a.firstChild;
    if (c && c === a.lastChild && 3 === c.nodeType) {
      c.nodeValue = b;
      return;
    }
  }
  a.textContent = b;
}
var pb = {
  animationIterationCount: true,
  aspectRatio: true,
  borderImageOutset: true,
  borderImageSlice: true,
  borderImageWidth: true,
  boxFlex: true,
  boxFlexGroup: true,
  boxOrdinalGroup: true,
  columnCount: true,
  columns: true,
  flex: true,
  flexGrow: true,
  flexPositive: true,
  flexShrink: true,
  flexNegative: true,
  flexOrder: true,
  gridArea: true,
  gridRow: true,
  gridRowEnd: true,
  gridRowSpan: true,
  gridRowStart: true,
  gridColumn: true,
  gridColumnEnd: true,
  gridColumnSpan: true,
  gridColumnStart: true,
  fontWeight: true,
  lineClamp: true,
  lineHeight: true,
  opacity: true,
  order: true,
  orphans: true,
  tabSize: true,
  widows: true,
  zIndex: true,
  zoom: true,
  fillOpacity: true,
  floodOpacity: true,
  stopOpacity: true,
  strokeDasharray: true,
  strokeDashoffset: true,
  strokeMiterlimit: true,
  strokeOpacity: true,
  strokeWidth: true
}, qb = ["Webkit", "ms", "Moz", "O"];
Object.keys(pb).forEach(function(a) {
  qb.forEach(function(b) {
    b = b + a.charAt(0).toUpperCase() + a.substring(1);
    pb[b] = pb[a];
  });
});
function rb(a, b, c) {
  return null == b || "boolean" === typeof b || "" === b ? "" : c || "number" !== typeof b || 0 === b || pb.hasOwnProperty(a) && pb[a] ? ("" + b).trim() : b + "px";
}
function sb(a, b) {
  a = a.style;
  for (var c in b) if (b.hasOwnProperty(c)) {
    var d = 0 === c.indexOf("--"), e = rb(c, b[c], d);
    "float" === c && (c = "cssFloat");
    d ? a.setProperty(c, e) : a[c] = e;
  }
}
var tb = A({ menuitem: true }, { area: true, base: true, br: true, col: true, embed: true, hr: true, img: true, input: true, keygen: true, link: true, meta: true, param: true, source: true, track: true, wbr: true });
function ub(a, b) {
  if (b) {
    if (tb[a] && (null != b.children || null != b.dangerouslySetInnerHTML)) throw Error(p(137, a));
    if (null != b.dangerouslySetInnerHTML) {
      if (null != b.children) throw Error(p(60));
      if ("object" !== typeof b.dangerouslySetInnerHTML || !("__html" in b.dangerouslySetInnerHTML)) throw Error(p(61));
    }
    if (null != b.style && "object" !== typeof b.style) throw Error(p(62));
  }
}
function vb(a, b) {
  if (-1 === a.indexOf("-")) return "string" === typeof b.is;
  switch (a) {
    case "annotation-xml":
    case "color-profile":
    case "font-face":
    case "font-face-src":
    case "font-face-uri":
    case "font-face-format":
    case "font-face-name":
    case "missing-glyph":
      return false;
    default:
      return true;
  }
}
var wb = null;
function xb(a) {
  a = a.target || a.srcElement || window;
  a.correspondingUseElement && (a = a.correspondingUseElement);
  return 3 === a.nodeType ? a.parentNode : a;
}
var yb = null, zb = null, Ab = null;
function Bb(a) {
  if (a = Cb(a)) {
    if ("function" !== typeof yb) throw Error(p(280));
    var b = a.stateNode;
    b && (b = Db(b), yb(a.stateNode, a.type, b));
  }
}
function Eb(a) {
  zb ? Ab ? Ab.push(a) : Ab = [a] : zb = a;
}
function Fb() {
  if (zb) {
    var a = zb, b = Ab;
    Ab = zb = null;
    Bb(a);
    if (b) for (a = 0; a < b.length; a++) Bb(b[a]);
  }
}
function Gb(a, b) {
  return a(b);
}
function Hb() {
}
var Ib = false;
function Jb(a, b, c) {
  if (Ib) return a(b, c);
  Ib = true;
  try {
    return Gb(a, b, c);
  } finally {
    if (Ib = false, null !== zb || null !== Ab) Hb(), Fb();
  }
}
function Kb(a, b) {
  var c = a.stateNode;
  if (null === c) return null;
  var d = Db(c);
  if (null === d) return null;
  c = d[b];
  a: switch (b) {
    case "onClick":
    case "onClickCapture":
    case "onDoubleClick":
    case "onDoubleClickCapture":
    case "onMouseDown":
    case "onMouseDownCapture":
    case "onMouseMove":
    case "onMouseMoveCapture":
    case "onMouseUp":
    case "onMouseUpCapture":
    case "onMouseEnter":
      (d = !d.disabled) || (a = a.type, d = !("button" === a || "input" === a || "select" === a || "textarea" === a));
      a = !d;
      break a;
    default:
      a = false;
  }
  if (a) return null;
  if (c && "function" !== typeof c) throw Error(p(231, b, typeof c));
  return c;
}
var Lb = false;
if (ia) try {
  var Mb = {};
  Object.defineProperty(Mb, "passive", { get: function() {
    Lb = true;
  } });
  window.addEventListener("test", Mb, Mb);
  window.removeEventListener("test", Mb, Mb);
} catch (a) {
  Lb = false;
}
function Nb(a, b, c, d, e, f2, g, h2, k2) {
  var l2 = Array.prototype.slice.call(arguments, 3);
  try {
    b.apply(c, l2);
  } catch (m2) {
    this.onError(m2);
  }
}
var Ob = false, Pb = null, Qb = false, Rb = null, Sb = { onError: function(a) {
  Ob = true;
  Pb = a;
} };
function Tb(a, b, c, d, e, f2, g, h2, k2) {
  Ob = false;
  Pb = null;
  Nb.apply(Sb, arguments);
}
function Ub(a, b, c, d, e, f2, g, h2, k2) {
  Tb.apply(this, arguments);
  if (Ob) {
    if (Ob) {
      var l2 = Pb;
      Ob = false;
      Pb = null;
    } else throw Error(p(198));
    Qb || (Qb = true, Rb = l2);
  }
}
function Vb(a) {
  var b = a, c = a;
  if (a.alternate) for (; b.return; ) b = b.return;
  else {
    a = b;
    do
      b = a, 0 !== (b.flags & 4098) && (c = b.return), a = b.return;
    while (a);
  }
  return 3 === b.tag ? c : null;
}
function Wb(a) {
  if (13 === a.tag) {
    var b = a.memoizedState;
    null === b && (a = a.alternate, null !== a && (b = a.memoizedState));
    if (null !== b) return b.dehydrated;
  }
  return null;
}
function Xb(a) {
  if (Vb(a) !== a) throw Error(p(188));
}
function Yb(a) {
  var b = a.alternate;
  if (!b) {
    b = Vb(a);
    if (null === b) throw Error(p(188));
    return b !== a ? null : a;
  }
  for (var c = a, d = b; ; ) {
    var e = c.return;
    if (null === e) break;
    var f2 = e.alternate;
    if (null === f2) {
      d = e.return;
      if (null !== d) {
        c = d;
        continue;
      }
      break;
    }
    if (e.child === f2.child) {
      for (f2 = e.child; f2; ) {
        if (f2 === c) return Xb(e), a;
        if (f2 === d) return Xb(e), b;
        f2 = f2.sibling;
      }
      throw Error(p(188));
    }
    if (c.return !== d.return) c = e, d = f2;
    else {
      for (var g = false, h2 = e.child; h2; ) {
        if (h2 === c) {
          g = true;
          c = e;
          d = f2;
          break;
        }
        if (h2 === d) {
          g = true;
          d = e;
          c = f2;
          break;
        }
        h2 = h2.sibling;
      }
      if (!g) {
        for (h2 = f2.child; h2; ) {
          if (h2 === c) {
            g = true;
            c = f2;
            d = e;
            break;
          }
          if (h2 === d) {
            g = true;
            d = f2;
            c = e;
            break;
          }
          h2 = h2.sibling;
        }
        if (!g) throw Error(p(189));
      }
    }
    if (c.alternate !== d) throw Error(p(190));
  }
  if (3 !== c.tag) throw Error(p(188));
  return c.stateNode.current === c ? a : b;
}
function Zb(a) {
  a = Yb(a);
  return null !== a ? $b(a) : null;
}
function $b(a) {
  if (5 === a.tag || 6 === a.tag) return a;
  for (a = a.child; null !== a; ) {
    var b = $b(a);
    if (null !== b) return b;
    a = a.sibling;
  }
  return null;
}
var ac = ca.unstable_scheduleCallback, bc = ca.unstable_cancelCallback, cc = ca.unstable_shouldYield, dc = ca.unstable_requestPaint, B = ca.unstable_now, ec = ca.unstable_getCurrentPriorityLevel, fc = ca.unstable_ImmediatePriority, gc = ca.unstable_UserBlockingPriority, hc = ca.unstable_NormalPriority, ic = ca.unstable_LowPriority, jc = ca.unstable_IdlePriority, kc = null, lc = null;
function mc(a) {
  if (lc && "function" === typeof lc.onCommitFiberRoot) try {
    lc.onCommitFiberRoot(kc, a, void 0, 128 === (a.current.flags & 128));
  } catch (b) {
  }
}
var oc = Math.clz32 ? Math.clz32 : nc, pc = Math.log, qc = Math.LN2;
function nc(a) {
  a >>>= 0;
  return 0 === a ? 32 : 31 - (pc(a) / qc | 0) | 0;
}
var rc = 64, sc = 4194304;
function tc(a) {
  switch (a & -a) {
    case 1:
      return 1;
    case 2:
      return 2;
    case 4:
      return 4;
    case 8:
      return 8;
    case 16:
      return 16;
    case 32:
      return 32;
    case 64:
    case 128:
    case 256:
    case 512:
    case 1024:
    case 2048:
    case 4096:
    case 8192:
    case 16384:
    case 32768:
    case 65536:
    case 131072:
    case 262144:
    case 524288:
    case 1048576:
    case 2097152:
      return a & 4194240;
    case 4194304:
    case 8388608:
    case 16777216:
    case 33554432:
    case 67108864:
      return a & 130023424;
    case 134217728:
      return 134217728;
    case 268435456:
      return 268435456;
    case 536870912:
      return 536870912;
    case 1073741824:
      return 1073741824;
    default:
      return a;
  }
}
function uc(a, b) {
  var c = a.pendingLanes;
  if (0 === c) return 0;
  var d = 0, e = a.suspendedLanes, f2 = a.pingedLanes, g = c & 268435455;
  if (0 !== g) {
    var h2 = g & ~e;
    0 !== h2 ? d = tc(h2) : (f2 &= g, 0 !== f2 && (d = tc(f2)));
  } else g = c & ~e, 0 !== g ? d = tc(g) : 0 !== f2 && (d = tc(f2));
  if (0 === d) return 0;
  if (0 !== b && b !== d && 0 === (b & e) && (e = d & -d, f2 = b & -b, e >= f2 || 16 === e && 0 !== (f2 & 4194240))) return b;
  0 !== (d & 4) && (d |= c & 16);
  b = a.entangledLanes;
  if (0 !== b) for (a = a.entanglements, b &= d; 0 < b; ) c = 31 - oc(b), e = 1 << c, d |= a[c], b &= ~e;
  return d;
}
function vc(a, b) {
  switch (a) {
    case 1:
    case 2:
    case 4:
      return b + 250;
    case 8:
    case 16:
    case 32:
    case 64:
    case 128:
    case 256:
    case 512:
    case 1024:
    case 2048:
    case 4096:
    case 8192:
    case 16384:
    case 32768:
    case 65536:
    case 131072:
    case 262144:
    case 524288:
    case 1048576:
    case 2097152:
      return b + 5e3;
    case 4194304:
    case 8388608:
    case 16777216:
    case 33554432:
    case 67108864:
      return -1;
    case 134217728:
    case 268435456:
    case 536870912:
    case 1073741824:
      return -1;
    default:
      return -1;
  }
}
function wc(a, b) {
  for (var c = a.suspendedLanes, d = a.pingedLanes, e = a.expirationTimes, f2 = a.pendingLanes; 0 < f2; ) {
    var g = 31 - oc(f2), h2 = 1 << g, k2 = e[g];
    if (-1 === k2) {
      if (0 === (h2 & c) || 0 !== (h2 & d)) e[g] = vc(h2, b);
    } else k2 <= b && (a.expiredLanes |= h2);
    f2 &= ~h2;
  }
}
function xc(a) {
  a = a.pendingLanes & -1073741825;
  return 0 !== a ? a : a & 1073741824 ? 1073741824 : 0;
}
function yc() {
  var a = rc;
  rc <<= 1;
  0 === (rc & 4194240) && (rc = 64);
  return a;
}
function zc(a) {
  for (var b = [], c = 0; 31 > c; c++) b.push(a);
  return b;
}
function Ac(a, b, c) {
  a.pendingLanes |= b;
  536870912 !== b && (a.suspendedLanes = 0, a.pingedLanes = 0);
  a = a.eventTimes;
  b = 31 - oc(b);
  a[b] = c;
}
function Bc(a, b) {
  var c = a.pendingLanes & ~b;
  a.pendingLanes = b;
  a.suspendedLanes = 0;
  a.pingedLanes = 0;
  a.expiredLanes &= b;
  a.mutableReadLanes &= b;
  a.entangledLanes &= b;
  b = a.entanglements;
  var d = a.eventTimes;
  for (a = a.expirationTimes; 0 < c; ) {
    var e = 31 - oc(c), f2 = 1 << e;
    b[e] = 0;
    d[e] = -1;
    a[e] = -1;
    c &= ~f2;
  }
}
function Cc(a, b) {
  var c = a.entangledLanes |= b;
  for (a = a.entanglements; c; ) {
    var d = 31 - oc(c), e = 1 << d;
    e & b | a[d] & b && (a[d] |= b);
    c &= ~e;
  }
}
var C = 0;
function Dc(a) {
  a &= -a;
  return 1 < a ? 4 < a ? 0 !== (a & 268435455) ? 16 : 536870912 : 4 : 1;
}
var Ec, Fc, Gc, Hc, Ic, Jc = false, Kc = [], Lc = null, Mc = null, Nc = null, Oc = /* @__PURE__ */ new Map(), Pc = /* @__PURE__ */ new Map(), Qc = [], Rc = "mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset submit".split(" ");
function Sc(a, b) {
  switch (a) {
    case "focusin":
    case "focusout":
      Lc = null;
      break;
    case "dragenter":
    case "dragleave":
      Mc = null;
      break;
    case "mouseover":
    case "mouseout":
      Nc = null;
      break;
    case "pointerover":
    case "pointerout":
      Oc.delete(b.pointerId);
      break;
    case "gotpointercapture":
    case "lostpointercapture":
      Pc.delete(b.pointerId);
  }
}
function Tc(a, b, c, d, e, f2) {
  if (null === a || a.nativeEvent !== f2) return a = { blockedOn: b, domEventName: c, eventSystemFlags: d, nativeEvent: f2, targetContainers: [e] }, null !== b && (b = Cb(b), null !== b && Fc(b)), a;
  a.eventSystemFlags |= d;
  b = a.targetContainers;
  null !== e && -1 === b.indexOf(e) && b.push(e);
  return a;
}
function Uc(a, b, c, d, e) {
  switch (b) {
    case "focusin":
      return Lc = Tc(Lc, a, b, c, d, e), true;
    case "dragenter":
      return Mc = Tc(Mc, a, b, c, d, e), true;
    case "mouseover":
      return Nc = Tc(Nc, a, b, c, d, e), true;
    case "pointerover":
      var f2 = e.pointerId;
      Oc.set(f2, Tc(Oc.get(f2) || null, a, b, c, d, e));
      return true;
    case "gotpointercapture":
      return f2 = e.pointerId, Pc.set(f2, Tc(Pc.get(f2) || null, a, b, c, d, e)), true;
  }
  return false;
}
function Vc(a) {
  var b = Wc(a.target);
  if (null !== b) {
    var c = Vb(b);
    if (null !== c) {
      if (b = c.tag, 13 === b) {
        if (b = Wb(c), null !== b) {
          a.blockedOn = b;
          Ic(a.priority, function() {
            Gc(c);
          });
          return;
        }
      } else if (3 === b && c.stateNode.current.memoizedState.isDehydrated) {
        a.blockedOn = 3 === c.tag ? c.stateNode.containerInfo : null;
        return;
      }
    }
  }
  a.blockedOn = null;
}
function Xc(a) {
  if (null !== a.blockedOn) return false;
  for (var b = a.targetContainers; 0 < b.length; ) {
    var c = Yc(a.domEventName, a.eventSystemFlags, b[0], a.nativeEvent);
    if (null === c) {
      c = a.nativeEvent;
      var d = new c.constructor(c.type, c);
      wb = d;
      c.target.dispatchEvent(d);
      wb = null;
    } else return b = Cb(c), null !== b && Fc(b), a.blockedOn = c, false;
    b.shift();
  }
  return true;
}
function Zc(a, b, c) {
  Xc(a) && c.delete(b);
}
function $c() {
  Jc = false;
  null !== Lc && Xc(Lc) && (Lc = null);
  null !== Mc && Xc(Mc) && (Mc = null);
  null !== Nc && Xc(Nc) && (Nc = null);
  Oc.forEach(Zc);
  Pc.forEach(Zc);
}
function ad(a, b) {
  a.blockedOn === b && (a.blockedOn = null, Jc || (Jc = true, ca.unstable_scheduleCallback(ca.unstable_NormalPriority, $c)));
}
function bd(a) {
  function b(b2) {
    return ad(b2, a);
  }
  if (0 < Kc.length) {
    ad(Kc[0], a);
    for (var c = 1; c < Kc.length; c++) {
      var d = Kc[c];
      d.blockedOn === a && (d.blockedOn = null);
    }
  }
  null !== Lc && ad(Lc, a);
  null !== Mc && ad(Mc, a);
  null !== Nc && ad(Nc, a);
  Oc.forEach(b);
  Pc.forEach(b);
  for (c = 0; c < Qc.length; c++) d = Qc[c], d.blockedOn === a && (d.blockedOn = null);
  for (; 0 < Qc.length && (c = Qc[0], null === c.blockedOn); ) Vc(c), null === c.blockedOn && Qc.shift();
}
var cd = ua.ReactCurrentBatchConfig, dd = true;
function ed(a, b, c, d) {
  var e = C, f2 = cd.transition;
  cd.transition = null;
  try {
    C = 1, fd(a, b, c, d);
  } finally {
    C = e, cd.transition = f2;
  }
}
function gd(a, b, c, d) {
  var e = C, f2 = cd.transition;
  cd.transition = null;
  try {
    C = 4, fd(a, b, c, d);
  } finally {
    C = e, cd.transition = f2;
  }
}
function fd(a, b, c, d) {
  if (dd) {
    var e = Yc(a, b, c, d);
    if (null === e) hd(a, b, d, id, c), Sc(a, d);
    else if (Uc(e, a, b, c, d)) d.stopPropagation();
    else if (Sc(a, d), b & 4 && -1 < Rc.indexOf(a)) {
      for (; null !== e; ) {
        var f2 = Cb(e);
        null !== f2 && Ec(f2);
        f2 = Yc(a, b, c, d);
        null === f2 && hd(a, b, d, id, c);
        if (f2 === e) break;
        e = f2;
      }
      null !== e && d.stopPropagation();
    } else hd(a, b, d, null, c);
  }
}
var id = null;
function Yc(a, b, c, d) {
  id = null;
  a = xb(d);
  a = Wc(a);
  if (null !== a) if (b = Vb(a), null === b) a = null;
  else if (c = b.tag, 13 === c) {
    a = Wb(b);
    if (null !== a) return a;
    a = null;
  } else if (3 === c) {
    if (b.stateNode.current.memoizedState.isDehydrated) return 3 === b.tag ? b.stateNode.containerInfo : null;
    a = null;
  } else b !== a && (a = null);
  id = a;
  return null;
}
function jd(a) {
  switch (a) {
    case "cancel":
    case "click":
    case "close":
    case "contextmenu":
    case "copy":
    case "cut":
    case "auxclick":
    case "dblclick":
    case "dragend":
    case "dragstart":
    case "drop":
    case "focusin":
    case "focusout":
    case "input":
    case "invalid":
    case "keydown":
    case "keypress":
    case "keyup":
    case "mousedown":
    case "mouseup":
    case "paste":
    case "pause":
    case "play":
    case "pointercancel":
    case "pointerdown":
    case "pointerup":
    case "ratechange":
    case "reset":
    case "resize":
    case "seeked":
    case "submit":
    case "touchcancel":
    case "touchend":
    case "touchstart":
    case "volumechange":
    case "change":
    case "selectionchange":
    case "textInput":
    case "compositionstart":
    case "compositionend":
    case "compositionupdate":
    case "beforeblur":
    case "afterblur":
    case "beforeinput":
    case "blur":
    case "fullscreenchange":
    case "focus":
    case "hashchange":
    case "popstate":
    case "select":
    case "selectstart":
      return 1;
    case "drag":
    case "dragenter":
    case "dragexit":
    case "dragleave":
    case "dragover":
    case "mousemove":
    case "mouseout":
    case "mouseover":
    case "pointermove":
    case "pointerout":
    case "pointerover":
    case "scroll":
    case "toggle":
    case "touchmove":
    case "wheel":
    case "mouseenter":
    case "mouseleave":
    case "pointerenter":
    case "pointerleave":
      return 4;
    case "message":
      switch (ec()) {
        case fc:
          return 1;
        case gc:
          return 4;
        case hc:
        case ic:
          return 16;
        case jc:
          return 536870912;
        default:
          return 16;
      }
    default:
      return 16;
  }
}
var kd = null, ld = null, md = null;
function nd() {
  if (md) return md;
  var a, b = ld, c = b.length, d, e = "value" in kd ? kd.value : kd.textContent, f2 = e.length;
  for (a = 0; a < c && b[a] === e[a]; a++) ;
  var g = c - a;
  for (d = 1; d <= g && b[c - d] === e[f2 - d]; d++) ;
  return md = e.slice(a, 1 < d ? 1 - d : void 0);
}
function od(a) {
  var b = a.keyCode;
  "charCode" in a ? (a = a.charCode, 0 === a && 13 === b && (a = 13)) : a = b;
  10 === a && (a = 13);
  return 32 <= a || 13 === a ? a : 0;
}
function pd() {
  return true;
}
function qd() {
  return false;
}
function rd(a) {
  function b(b2, d, e, f2, g) {
    this._reactName = b2;
    this._targetInst = e;
    this.type = d;
    this.nativeEvent = f2;
    this.target = g;
    this.currentTarget = null;
    for (var c in a) a.hasOwnProperty(c) && (b2 = a[c], this[c] = b2 ? b2(f2) : f2[c]);
    this.isDefaultPrevented = (null != f2.defaultPrevented ? f2.defaultPrevented : false === f2.returnValue) ? pd : qd;
    this.isPropagationStopped = qd;
    return this;
  }
  A(b.prototype, { preventDefault: function() {
    this.defaultPrevented = true;
    var a2 = this.nativeEvent;
    a2 && (a2.preventDefault ? a2.preventDefault() : "unknown" !== typeof a2.returnValue && (a2.returnValue = false), this.isDefaultPrevented = pd);
  }, stopPropagation: function() {
    var a2 = this.nativeEvent;
    a2 && (a2.stopPropagation ? a2.stopPropagation() : "unknown" !== typeof a2.cancelBubble && (a2.cancelBubble = true), this.isPropagationStopped = pd);
  }, persist: function() {
  }, isPersistent: pd });
  return b;
}
var sd = { eventPhase: 0, bubbles: 0, cancelable: 0, timeStamp: function(a) {
  return a.timeStamp || Date.now();
}, defaultPrevented: 0, isTrusted: 0 }, td = rd(sd), ud = A({}, sd, { view: 0, detail: 0 }), vd = rd(ud), wd, xd, yd, Ad = A({}, ud, { screenX: 0, screenY: 0, clientX: 0, clientY: 0, pageX: 0, pageY: 0, ctrlKey: 0, shiftKey: 0, altKey: 0, metaKey: 0, getModifierState: zd, button: 0, buttons: 0, relatedTarget: function(a) {
  return void 0 === a.relatedTarget ? a.fromElement === a.srcElement ? a.toElement : a.fromElement : a.relatedTarget;
}, movementX: function(a) {
  if ("movementX" in a) return a.movementX;
  a !== yd && (yd && "mousemove" === a.type ? (wd = a.screenX - yd.screenX, xd = a.screenY - yd.screenY) : xd = wd = 0, yd = a);
  return wd;
}, movementY: function(a) {
  return "movementY" in a ? a.movementY : xd;
} }), Bd = rd(Ad), Cd = A({}, Ad, { dataTransfer: 0 }), Dd = rd(Cd), Ed = A({}, ud, { relatedTarget: 0 }), Fd = rd(Ed), Gd = A({}, sd, { animationName: 0, elapsedTime: 0, pseudoElement: 0 }), Hd = rd(Gd), Id = A({}, sd, { clipboardData: function(a) {
  return "clipboardData" in a ? a.clipboardData : window.clipboardData;
} }), Jd = rd(Id), Kd = A({}, sd, { data: 0 }), Ld = rd(Kd), Md = {
  Esc: "Escape",
  Spacebar: " ",
  Left: "ArrowLeft",
  Up: "ArrowUp",
  Right: "ArrowRight",
  Down: "ArrowDown",
  Del: "Delete",
  Win: "OS",
  Menu: "ContextMenu",
  Apps: "ContextMenu",
  Scroll: "ScrollLock",
  MozPrintableKey: "Unidentified"
}, Nd = {
  8: "Backspace",
  9: "Tab",
  12: "Clear",
  13: "Enter",
  16: "Shift",
  17: "Control",
  18: "Alt",
  19: "Pause",
  20: "CapsLock",
  27: "Escape",
  32: " ",
  33: "PageUp",
  34: "PageDown",
  35: "End",
  36: "Home",
  37: "ArrowLeft",
  38: "ArrowUp",
  39: "ArrowRight",
  40: "ArrowDown",
  45: "Insert",
  46: "Delete",
  112: "F1",
  113: "F2",
  114: "F3",
  115: "F4",
  116: "F5",
  117: "F6",
  118: "F7",
  119: "F8",
  120: "F9",
  121: "F10",
  122: "F11",
  123: "F12",
  144: "NumLock",
  145: "ScrollLock",
  224: "Meta"
}, Od = { Alt: "altKey", Control: "ctrlKey", Meta: "metaKey", Shift: "shiftKey" };
function Pd(a) {
  var b = this.nativeEvent;
  return b.getModifierState ? b.getModifierState(a) : (a = Od[a]) ? !!b[a] : false;
}
function zd() {
  return Pd;
}
var Qd = A({}, ud, { key: function(a) {
  if (a.key) {
    var b = Md[a.key] || a.key;
    if ("Unidentified" !== b) return b;
  }
  return "keypress" === a.type ? (a = od(a), 13 === a ? "Enter" : String.fromCharCode(a)) : "keydown" === a.type || "keyup" === a.type ? Nd[a.keyCode] || "Unidentified" : "";
}, code: 0, location: 0, ctrlKey: 0, shiftKey: 0, altKey: 0, metaKey: 0, repeat: 0, locale: 0, getModifierState: zd, charCode: function(a) {
  return "keypress" === a.type ? od(a) : 0;
}, keyCode: function(a) {
  return "keydown" === a.type || "keyup" === a.type ? a.keyCode : 0;
}, which: function(a) {
  return "keypress" === a.type ? od(a) : "keydown" === a.type || "keyup" === a.type ? a.keyCode : 0;
} }), Rd = rd(Qd), Sd = A({}, Ad, { pointerId: 0, width: 0, height: 0, pressure: 0, tangentialPressure: 0, tiltX: 0, tiltY: 0, twist: 0, pointerType: 0, isPrimary: 0 }), Td = rd(Sd), Ud = A({}, ud, { touches: 0, targetTouches: 0, changedTouches: 0, altKey: 0, metaKey: 0, ctrlKey: 0, shiftKey: 0, getModifierState: zd }), Vd = rd(Ud), Wd = A({}, sd, { propertyName: 0, elapsedTime: 0, pseudoElement: 0 }), Xd = rd(Wd), Yd = A({}, Ad, {
  deltaX: function(a) {
    return "deltaX" in a ? a.deltaX : "wheelDeltaX" in a ? -a.wheelDeltaX : 0;
  },
  deltaY: function(a) {
    return "deltaY" in a ? a.deltaY : "wheelDeltaY" in a ? -a.wheelDeltaY : "wheelDelta" in a ? -a.wheelDelta : 0;
  },
  deltaZ: 0,
  deltaMode: 0
}), Zd = rd(Yd), $d = [9, 13, 27, 32], ae$1 = ia && "CompositionEvent" in window, be$1 = null;
ia && "documentMode" in document && (be$1 = document.documentMode);
var ce = ia && "TextEvent" in window && !be$1, de$1 = ia && (!ae$1 || be$1 && 8 < be$1 && 11 >= be$1), ee$1 = String.fromCharCode(32), fe$1 = false;
function ge(a, b) {
  switch (a) {
    case "keyup":
      return -1 !== $d.indexOf(b.keyCode);
    case "keydown":
      return 229 !== b.keyCode;
    case "keypress":
    case "mousedown":
    case "focusout":
      return true;
    default:
      return false;
  }
}
function he$1(a) {
  a = a.detail;
  return "object" === typeof a && "data" in a ? a.data : null;
}
var ie$1 = false;
function je(a, b) {
  switch (a) {
    case "compositionend":
      return he$1(b);
    case "keypress":
      if (32 !== b.which) return null;
      fe$1 = true;
      return ee$1;
    case "textInput":
      return a = b.data, a === ee$1 && fe$1 ? null : a;
    default:
      return null;
  }
}
function ke(a, b) {
  if (ie$1) return "compositionend" === a || !ae$1 && ge(a, b) ? (a = nd(), md = ld = kd = null, ie$1 = false, a) : null;
  switch (a) {
    case "paste":
      return null;
    case "keypress":
      if (!(b.ctrlKey || b.altKey || b.metaKey) || b.ctrlKey && b.altKey) {
        if (b.char && 1 < b.char.length) return b.char;
        if (b.which) return String.fromCharCode(b.which);
      }
      return null;
    case "compositionend":
      return de$1 && "ko" !== b.locale ? null : b.data;
    default:
      return null;
  }
}
var le$1 = { color: true, date: true, datetime: true, "datetime-local": true, email: true, month: true, number: true, password: true, range: true, search: true, tel: true, text: true, time: true, url: true, week: true };
function me(a) {
  var b = a && a.nodeName && a.nodeName.toLowerCase();
  return "input" === b ? !!le$1[a.type] : "textarea" === b ? true : false;
}
function ne(a, b, c, d) {
  Eb(d);
  b = oe(b, "onChange");
  0 < b.length && (c = new td("onChange", "change", null, c, d), a.push({ event: c, listeners: b }));
}
var pe = null, qe = null;
function re(a) {
  se$1(a, 0);
}
function te$1(a) {
  var b = ue(a);
  if (Wa(b)) return a;
}
function ve(a, b) {
  if ("change" === a) return b;
}
var we = false;
if (ia) {
  var xe;
  if (ia) {
    var ye = "oninput" in document;
    if (!ye) {
      var ze = document.createElement("div");
      ze.setAttribute("oninput", "return;");
      ye = "function" === typeof ze.oninput;
    }
    xe = ye;
  } else xe = false;
  we = xe && (!document.documentMode || 9 < document.documentMode);
}
function Ae() {
  pe && (pe.detachEvent("onpropertychange", Be), qe = pe = null);
}
function Be(a) {
  if ("value" === a.propertyName && te$1(qe)) {
    var b = [];
    ne(b, qe, a, xb(a));
    Jb(re, b);
  }
}
function Ce$1(a, b, c) {
  "focusin" === a ? (Ae(), pe = b, qe = c, pe.attachEvent("onpropertychange", Be)) : "focusout" === a && Ae();
}
function De$1(a) {
  if ("selectionchange" === a || "keyup" === a || "keydown" === a) return te$1(qe);
}
function Ee$1(a, b) {
  if ("click" === a) return te$1(b);
}
function Fe(a, b) {
  if ("input" === a || "change" === a) return te$1(b);
}
function Ge(a, b) {
  return a === b && (0 !== a || 1 / a === 1 / b) || a !== a && b !== b;
}
var He$1 = "function" === typeof Object.is ? Object.is : Ge;
function Ie(a, b) {
  if (He$1(a, b)) return true;
  if ("object" !== typeof a || null === a || "object" !== typeof b || null === b) return false;
  var c = Object.keys(a), d = Object.keys(b);
  if (c.length !== d.length) return false;
  for (d = 0; d < c.length; d++) {
    var e = c[d];
    if (!ja.call(b, e) || !He$1(a[e], b[e])) return false;
  }
  return true;
}
function Je(a) {
  for (; a && a.firstChild; ) a = a.firstChild;
  return a;
}
function Ke(a, b) {
  var c = Je(a);
  a = 0;
  for (var d; c; ) {
    if (3 === c.nodeType) {
      d = a + c.textContent.length;
      if (a <= b && d >= b) return { node: c, offset: b - a };
      a = d;
    }
    a: {
      for (; c; ) {
        if (c.nextSibling) {
          c = c.nextSibling;
          break a;
        }
        c = c.parentNode;
      }
      c = void 0;
    }
    c = Je(c);
  }
}
function Le(a, b) {
  return a && b ? a === b ? true : a && 3 === a.nodeType ? false : b && 3 === b.nodeType ? Le(a, b.parentNode) : "contains" in a ? a.contains(b) : a.compareDocumentPosition ? !!(a.compareDocumentPosition(b) & 16) : false : false;
}
function Me$1() {
  for (var a = window, b = Xa(); b instanceof a.HTMLIFrameElement; ) {
    try {
      var c = "string" === typeof b.contentWindow.location.href;
    } catch (d) {
      c = false;
    }
    if (c) a = b.contentWindow;
    else break;
    b = Xa(a.document);
  }
  return b;
}
function Ne(a) {
  var b = a && a.nodeName && a.nodeName.toLowerCase();
  return b && ("input" === b && ("text" === a.type || "search" === a.type || "tel" === a.type || "url" === a.type || "password" === a.type) || "textarea" === b || "true" === a.contentEditable);
}
function Oe$1(a) {
  var b = Me$1(), c = a.focusedElem, d = a.selectionRange;
  if (b !== c && c && c.ownerDocument && Le(c.ownerDocument.documentElement, c)) {
    if (null !== d && Ne(c)) {
      if (b = d.start, a = d.end, void 0 === a && (a = b), "selectionStart" in c) c.selectionStart = b, c.selectionEnd = Math.min(a, c.value.length);
      else if (a = (b = c.ownerDocument || document) && b.defaultView || window, a.getSelection) {
        a = a.getSelection();
        var e = c.textContent.length, f2 = Math.min(d.start, e);
        d = void 0 === d.end ? f2 : Math.min(d.end, e);
        !a.extend && f2 > d && (e = d, d = f2, f2 = e);
        e = Ke(c, f2);
        var g = Ke(
          c,
          d
        );
        e && g && (1 !== a.rangeCount || a.anchorNode !== e.node || a.anchorOffset !== e.offset || a.focusNode !== g.node || a.focusOffset !== g.offset) && (b = b.createRange(), b.setStart(e.node, e.offset), a.removeAllRanges(), f2 > d ? (a.addRange(b), a.extend(g.node, g.offset)) : (b.setEnd(g.node, g.offset), a.addRange(b)));
      }
    }
    b = [];
    for (a = c; a = a.parentNode; ) 1 === a.nodeType && b.push({ element: a, left: a.scrollLeft, top: a.scrollTop });
    "function" === typeof c.focus && c.focus();
    for (c = 0; c < b.length; c++) a = b[c], a.element.scrollLeft = a.left, a.element.scrollTop = a.top;
  }
}
var Pe = ia && "documentMode" in document && 11 >= document.documentMode, Qe = null, Re = null, Se = null, Te = false;
function Ue(a, b, c) {
  var d = c.window === c ? c.document : 9 === c.nodeType ? c : c.ownerDocument;
  Te || null == Qe || Qe !== Xa(d) || (d = Qe, "selectionStart" in d && Ne(d) ? d = { start: d.selectionStart, end: d.selectionEnd } : (d = (d.ownerDocument && d.ownerDocument.defaultView || window).getSelection(), d = { anchorNode: d.anchorNode, anchorOffset: d.anchorOffset, focusNode: d.focusNode, focusOffset: d.focusOffset }), Se && Ie(Se, d) || (Se = d, d = oe(Re, "onSelect"), 0 < d.length && (b = new td("onSelect", "select", null, b, c), a.push({ event: b, listeners: d }), b.target = Qe)));
}
function Ve$1(a, b) {
  var c = {};
  c[a.toLowerCase()] = b.toLowerCase();
  c["Webkit" + a] = "webkit" + b;
  c["Moz" + a] = "moz" + b;
  return c;
}
var We = { animationend: Ve$1("Animation", "AnimationEnd"), animationiteration: Ve$1("Animation", "AnimationIteration"), animationstart: Ve$1("Animation", "AnimationStart"), transitionend: Ve$1("Transition", "TransitionEnd") }, Xe = {}, Ye = {};
ia && (Ye = document.createElement("div").style, "AnimationEvent" in window || (delete We.animationend.animation, delete We.animationiteration.animation, delete We.animationstart.animation), "TransitionEvent" in window || delete We.transitionend.transition);
function Ze(a) {
  if (Xe[a]) return Xe[a];
  if (!We[a]) return a;
  var b = We[a], c;
  for (c in b) if (b.hasOwnProperty(c) && c in Ye) return Xe[a] = b[c];
  return a;
}
var $e = Ze("animationend"), af = Ze("animationiteration"), bf = Ze("animationstart"), cf = Ze("transitionend"), df = /* @__PURE__ */ new Map(), ef = "abort auxClick cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");
function ff(a, b) {
  df.set(a, b);
  fa(b, [a]);
}
for (var gf = 0; gf < ef.length; gf++) {
  var hf = ef[gf], jf = hf.toLowerCase(), kf = hf[0].toUpperCase() + hf.slice(1);
  ff(jf, "on" + kf);
}
ff($e, "onAnimationEnd");
ff(af, "onAnimationIteration");
ff(bf, "onAnimationStart");
ff("dblclick", "onDoubleClick");
ff("focusin", "onFocus");
ff("focusout", "onBlur");
ff(cf, "onTransitionEnd");
ha("onMouseEnter", ["mouseout", "mouseover"]);
ha("onMouseLeave", ["mouseout", "mouseover"]);
ha("onPointerEnter", ["pointerout", "pointerover"]);
ha("onPointerLeave", ["pointerout", "pointerover"]);
fa("onChange", "change click focusin focusout input keydown keyup selectionchange".split(" "));
fa("onSelect", "focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" "));
fa("onBeforeInput", ["compositionend", "keypress", "textInput", "paste"]);
fa("onCompositionEnd", "compositionend focusout keydown keypress keyup mousedown".split(" "));
fa("onCompositionStart", "compositionstart focusout keydown keypress keyup mousedown".split(" "));
fa("onCompositionUpdate", "compositionupdate focusout keydown keypress keyup mousedown".split(" "));
var lf = "abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "), mf = new Set("cancel close invalid load scroll toggle".split(" ").concat(lf));
function nf(a, b, c) {
  var d = a.type || "unknown-event";
  a.currentTarget = c;
  Ub(d, b, void 0, a);
  a.currentTarget = null;
}
function se$1(a, b) {
  b = 0 !== (b & 4);
  for (var c = 0; c < a.length; c++) {
    var d = a[c], e = d.event;
    d = d.listeners;
    a: {
      var f2 = void 0;
      if (b) for (var g = d.length - 1; 0 <= g; g--) {
        var h2 = d[g], k2 = h2.instance, l2 = h2.currentTarget;
        h2 = h2.listener;
        if (k2 !== f2 && e.isPropagationStopped()) break a;
        nf(e, h2, l2);
        f2 = k2;
      }
      else for (g = 0; g < d.length; g++) {
        h2 = d[g];
        k2 = h2.instance;
        l2 = h2.currentTarget;
        h2 = h2.listener;
        if (k2 !== f2 && e.isPropagationStopped()) break a;
        nf(e, h2, l2);
        f2 = k2;
      }
    }
  }
  if (Qb) throw a = Rb, Qb = false, Rb = null, a;
}
function D$1(a, b) {
  var c = b[of];
  void 0 === c && (c = b[of] = /* @__PURE__ */ new Set());
  var d = a + "__bubble";
  c.has(d) || (pf(b, a, 2, false), c.add(d));
}
function qf(a, b, c) {
  var d = 0;
  b && (d |= 4);
  pf(c, a, d, b);
}
var rf = "_reactListening" + Math.random().toString(36).slice(2);
function sf(a) {
  if (!a[rf]) {
    a[rf] = true;
    da.forEach(function(b2) {
      "selectionchange" !== b2 && (mf.has(b2) || qf(b2, false, a), qf(b2, true, a));
    });
    var b = 9 === a.nodeType ? a : a.ownerDocument;
    null === b || b[rf] || (b[rf] = true, qf("selectionchange", false, b));
  }
}
function pf(a, b, c, d) {
  switch (jd(b)) {
    case 1:
      var e = ed;
      break;
    case 4:
      e = gd;
      break;
    default:
      e = fd;
  }
  c = e.bind(null, b, c, a);
  e = void 0;
  !Lb || "touchstart" !== b && "touchmove" !== b && "wheel" !== b || (e = true);
  d ? void 0 !== e ? a.addEventListener(b, c, { capture: true, passive: e }) : a.addEventListener(b, c, true) : void 0 !== e ? a.addEventListener(b, c, { passive: e }) : a.addEventListener(b, c, false);
}
function hd(a, b, c, d, e) {
  var f2 = d;
  if (0 === (b & 1) && 0 === (b & 2) && null !== d) a: for (; ; ) {
    if (null === d) return;
    var g = d.tag;
    if (3 === g || 4 === g) {
      var h2 = d.stateNode.containerInfo;
      if (h2 === e || 8 === h2.nodeType && h2.parentNode === e) break;
      if (4 === g) for (g = d.return; null !== g; ) {
        var k2 = g.tag;
        if (3 === k2 || 4 === k2) {
          if (k2 = g.stateNode.containerInfo, k2 === e || 8 === k2.nodeType && k2.parentNode === e) return;
        }
        g = g.return;
      }
      for (; null !== h2; ) {
        g = Wc(h2);
        if (null === g) return;
        k2 = g.tag;
        if (5 === k2 || 6 === k2) {
          d = f2 = g;
          continue a;
        }
        h2 = h2.parentNode;
      }
    }
    d = d.return;
  }
  Jb(function() {
    var d2 = f2, e2 = xb(c), g2 = [];
    a: {
      var h3 = df.get(a);
      if (void 0 !== h3) {
        var k3 = td, n2 = a;
        switch (a) {
          case "keypress":
            if (0 === od(c)) break a;
          case "keydown":
          case "keyup":
            k3 = Rd;
            break;
          case "focusin":
            n2 = "focus";
            k3 = Fd;
            break;
          case "focusout":
            n2 = "blur";
            k3 = Fd;
            break;
          case "beforeblur":
          case "afterblur":
            k3 = Fd;
            break;
          case "click":
            if (2 === c.button) break a;
          case "auxclick":
          case "dblclick":
          case "mousedown":
          case "mousemove":
          case "mouseup":
          case "mouseout":
          case "mouseover":
          case "contextmenu":
            k3 = Bd;
            break;
          case "drag":
          case "dragend":
          case "dragenter":
          case "dragexit":
          case "dragleave":
          case "dragover":
          case "dragstart":
          case "drop":
            k3 = Dd;
            break;
          case "touchcancel":
          case "touchend":
          case "touchmove":
          case "touchstart":
            k3 = Vd;
            break;
          case $e:
          case af:
          case bf:
            k3 = Hd;
            break;
          case cf:
            k3 = Xd;
            break;
          case "scroll":
            k3 = vd;
            break;
          case "wheel":
            k3 = Zd;
            break;
          case "copy":
          case "cut":
          case "paste":
            k3 = Jd;
            break;
          case "gotpointercapture":
          case "lostpointercapture":
          case "pointercancel":
          case "pointerdown":
          case "pointermove":
          case "pointerout":
          case "pointerover":
          case "pointerup":
            k3 = Td;
        }
        var t2 = 0 !== (b & 4), J2 = !t2 && "scroll" === a, x2 = t2 ? null !== h3 ? h3 + "Capture" : null : h3;
        t2 = [];
        for (var w2 = d2, u2; null !== w2; ) {
          u2 = w2;
          var F2 = u2.stateNode;
          5 === u2.tag && null !== F2 && (u2 = F2, null !== x2 && (F2 = Kb(w2, x2), null != F2 && t2.push(tf(w2, F2, u2))));
          if (J2) break;
          w2 = w2.return;
        }
        0 < t2.length && (h3 = new k3(h3, n2, null, c, e2), g2.push({ event: h3, listeners: t2 }));
      }
    }
    if (0 === (b & 7)) {
      a: {
        h3 = "mouseover" === a || "pointerover" === a;
        k3 = "mouseout" === a || "pointerout" === a;
        if (h3 && c !== wb && (n2 = c.relatedTarget || c.fromElement) && (Wc(n2) || n2[uf])) break a;
        if (k3 || h3) {
          h3 = e2.window === e2 ? e2 : (h3 = e2.ownerDocument) ? h3.defaultView || h3.parentWindow : window;
          if (k3) {
            if (n2 = c.relatedTarget || c.toElement, k3 = d2, n2 = n2 ? Wc(n2) : null, null !== n2 && (J2 = Vb(n2), n2 !== J2 || 5 !== n2.tag && 6 !== n2.tag)) n2 = null;
          } else k3 = null, n2 = d2;
          if (k3 !== n2) {
            t2 = Bd;
            F2 = "onMouseLeave";
            x2 = "onMouseEnter";
            w2 = "mouse";
            if ("pointerout" === a || "pointerover" === a) t2 = Td, F2 = "onPointerLeave", x2 = "onPointerEnter", w2 = "pointer";
            J2 = null == k3 ? h3 : ue(k3);
            u2 = null == n2 ? h3 : ue(n2);
            h3 = new t2(F2, w2 + "leave", k3, c, e2);
            h3.target = J2;
            h3.relatedTarget = u2;
            F2 = null;
            Wc(e2) === d2 && (t2 = new t2(x2, w2 + "enter", n2, c, e2), t2.target = u2, t2.relatedTarget = J2, F2 = t2);
            J2 = F2;
            if (k3 && n2) b: {
              t2 = k3;
              x2 = n2;
              w2 = 0;
              for (u2 = t2; u2; u2 = vf(u2)) w2++;
              u2 = 0;
              for (F2 = x2; F2; F2 = vf(F2)) u2++;
              for (; 0 < w2 - u2; ) t2 = vf(t2), w2--;
              for (; 0 < u2 - w2; ) x2 = vf(x2), u2--;
              for (; w2--; ) {
                if (t2 === x2 || null !== x2 && t2 === x2.alternate) break b;
                t2 = vf(t2);
                x2 = vf(x2);
              }
              t2 = null;
            }
            else t2 = null;
            null !== k3 && wf(g2, h3, k3, t2, false);
            null !== n2 && null !== J2 && wf(g2, J2, n2, t2, true);
          }
        }
      }
      a: {
        h3 = d2 ? ue(d2) : window;
        k3 = h3.nodeName && h3.nodeName.toLowerCase();
        if ("select" === k3 || "input" === k3 && "file" === h3.type) var na = ve;
        else if (me(h3)) if (we) na = Fe;
        else {
          na = De$1;
          var xa = Ce$1;
        }
        else (k3 = h3.nodeName) && "input" === k3.toLowerCase() && ("checkbox" === h3.type || "radio" === h3.type) && (na = Ee$1);
        if (na && (na = na(a, d2))) {
          ne(g2, na, c, e2);
          break a;
        }
        xa && xa(a, h3, d2);
        "focusout" === a && (xa = h3._wrapperState) && xa.controlled && "number" === h3.type && cb(h3, "number", h3.value);
      }
      xa = d2 ? ue(d2) : window;
      switch (a) {
        case "focusin":
          if (me(xa) || "true" === xa.contentEditable) Qe = xa, Re = d2, Se = null;
          break;
        case "focusout":
          Se = Re = Qe = null;
          break;
        case "mousedown":
          Te = true;
          break;
        case "contextmenu":
        case "mouseup":
        case "dragend":
          Te = false;
          Ue(g2, c, e2);
          break;
        case "selectionchange":
          if (Pe) break;
        case "keydown":
        case "keyup":
          Ue(g2, c, e2);
      }
      var $a;
      if (ae$1) b: {
        switch (a) {
          case "compositionstart":
            var ba = "onCompositionStart";
            break b;
          case "compositionend":
            ba = "onCompositionEnd";
            break b;
          case "compositionupdate":
            ba = "onCompositionUpdate";
            break b;
        }
        ba = void 0;
      }
      else ie$1 ? ge(a, c) && (ba = "onCompositionEnd") : "keydown" === a && 229 === c.keyCode && (ba = "onCompositionStart");
      ba && (de$1 && "ko" !== c.locale && (ie$1 || "onCompositionStart" !== ba ? "onCompositionEnd" === ba && ie$1 && ($a = nd()) : (kd = e2, ld = "value" in kd ? kd.value : kd.textContent, ie$1 = true)), xa = oe(d2, ba), 0 < xa.length && (ba = new Ld(ba, a, null, c, e2), g2.push({ event: ba, listeners: xa }), $a ? ba.data = $a : ($a = he$1(c), null !== $a && (ba.data = $a))));
      if ($a = ce ? je(a, c) : ke(a, c)) d2 = oe(d2, "onBeforeInput"), 0 < d2.length && (e2 = new Ld("onBeforeInput", "beforeinput", null, c, e2), g2.push({ event: e2, listeners: d2 }), e2.data = $a);
    }
    se$1(g2, b);
  });
}
function tf(a, b, c) {
  return { instance: a, listener: b, currentTarget: c };
}
function oe(a, b) {
  for (var c = b + "Capture", d = []; null !== a; ) {
    var e = a, f2 = e.stateNode;
    5 === e.tag && null !== f2 && (e = f2, f2 = Kb(a, c), null != f2 && d.unshift(tf(a, f2, e)), f2 = Kb(a, b), null != f2 && d.push(tf(a, f2, e)));
    a = a.return;
  }
  return d;
}
function vf(a) {
  if (null === a) return null;
  do
    a = a.return;
  while (a && 5 !== a.tag);
  return a ? a : null;
}
function wf(a, b, c, d, e) {
  for (var f2 = b._reactName, g = []; null !== c && c !== d; ) {
    var h2 = c, k2 = h2.alternate, l2 = h2.stateNode;
    if (null !== k2 && k2 === d) break;
    5 === h2.tag && null !== l2 && (h2 = l2, e ? (k2 = Kb(c, f2), null != k2 && g.unshift(tf(c, k2, h2))) : e || (k2 = Kb(c, f2), null != k2 && g.push(tf(c, k2, h2))));
    c = c.return;
  }
  0 !== g.length && a.push({ event: b, listeners: g });
}
var xf = /\r\n?/g, yf = /\u0000|\uFFFD/g;
function zf(a) {
  return ("string" === typeof a ? a : "" + a).replace(xf, "\n").replace(yf, "");
}
function Af(a, b, c) {
  b = zf(b);
  if (zf(a) !== b && c) throw Error(p(425));
}
function Bf() {
}
var Cf = null, Df = null;
function Ef(a, b) {
  return "textarea" === a || "noscript" === a || "string" === typeof b.children || "number" === typeof b.children || "object" === typeof b.dangerouslySetInnerHTML && null !== b.dangerouslySetInnerHTML && null != b.dangerouslySetInnerHTML.__html;
}
var Ff = "function" === typeof setTimeout ? setTimeout : void 0, Gf = "function" === typeof clearTimeout ? clearTimeout : void 0, Hf = "function" === typeof Promise ? Promise : void 0, Jf = "function" === typeof queueMicrotask ? queueMicrotask : "undefined" !== typeof Hf ? function(a) {
  return Hf.resolve(null).then(a).catch(If);
} : Ff;
function If(a) {
  setTimeout(function() {
    throw a;
  });
}
function Kf(a, b) {
  var c = b, d = 0;
  do {
    var e = c.nextSibling;
    a.removeChild(c);
    if (e && 8 === e.nodeType) if (c = e.data, "/$" === c) {
      if (0 === d) {
        a.removeChild(e);
        bd(b);
        return;
      }
      d--;
    } else "$" !== c && "$?" !== c && "$!" !== c || d++;
    c = e;
  } while (c);
  bd(b);
}
function Lf(a) {
  for (; null != a; a = a.nextSibling) {
    var b = a.nodeType;
    if (1 === b || 3 === b) break;
    if (8 === b) {
      b = a.data;
      if ("$" === b || "$!" === b || "$?" === b) break;
      if ("/$" === b) return null;
    }
  }
  return a;
}
function Mf(a) {
  a = a.previousSibling;
  for (var b = 0; a; ) {
    if (8 === a.nodeType) {
      var c = a.data;
      if ("$" === c || "$!" === c || "$?" === c) {
        if (0 === b) return a;
        b--;
      } else "/$" === c && b++;
    }
    a = a.previousSibling;
  }
  return null;
}
var Nf = Math.random().toString(36).slice(2), Of = "__reactFiber$" + Nf, Pf = "__reactProps$" + Nf, uf = "__reactContainer$" + Nf, of = "__reactEvents$" + Nf, Qf = "__reactListeners$" + Nf, Rf = "__reactHandles$" + Nf;
function Wc(a) {
  var b = a[Of];
  if (b) return b;
  for (var c = a.parentNode; c; ) {
    if (b = c[uf] || c[Of]) {
      c = b.alternate;
      if (null !== b.child || null !== c && null !== c.child) for (a = Mf(a); null !== a; ) {
        if (c = a[Of]) return c;
        a = Mf(a);
      }
      return b;
    }
    a = c;
    c = a.parentNode;
  }
  return null;
}
function Cb(a) {
  a = a[Of] || a[uf];
  return !a || 5 !== a.tag && 6 !== a.tag && 13 !== a.tag && 3 !== a.tag ? null : a;
}
function ue(a) {
  if (5 === a.tag || 6 === a.tag) return a.stateNode;
  throw Error(p(33));
}
function Db(a) {
  return a[Pf] || null;
}
var Sf = [], Tf = -1;
function Uf(a) {
  return { current: a };
}
function E(a) {
  0 > Tf || (a.current = Sf[Tf], Sf[Tf] = null, Tf--);
}
function G(a, b) {
  Tf++;
  Sf[Tf] = a.current;
  a.current = b;
}
var Vf = {}, H$1 = Uf(Vf), Wf = Uf(false), Xf = Vf;
function Yf(a, b) {
  var c = a.type.contextTypes;
  if (!c) return Vf;
  var d = a.stateNode;
  if (d && d.__reactInternalMemoizedUnmaskedChildContext === b) return d.__reactInternalMemoizedMaskedChildContext;
  var e = {}, f2;
  for (f2 in c) e[f2] = b[f2];
  d && (a = a.stateNode, a.__reactInternalMemoizedUnmaskedChildContext = b, a.__reactInternalMemoizedMaskedChildContext = e);
  return e;
}
function Zf(a) {
  a = a.childContextTypes;
  return null !== a && void 0 !== a;
}
function $f() {
  E(Wf);
  E(H$1);
}
function ag(a, b, c) {
  if (H$1.current !== Vf) throw Error(p(168));
  G(H$1, b);
  G(Wf, c);
}
function bg(a, b, c) {
  var d = a.stateNode;
  b = b.childContextTypes;
  if ("function" !== typeof d.getChildContext) return c;
  d = d.getChildContext();
  for (var e in d) if (!(e in b)) throw Error(p(108, Ra(a) || "Unknown", e));
  return A({}, c, d);
}
function cg(a) {
  a = (a = a.stateNode) && a.__reactInternalMemoizedMergedChildContext || Vf;
  Xf = H$1.current;
  G(H$1, a);
  G(Wf, Wf.current);
  return true;
}
function dg(a, b, c) {
  var d = a.stateNode;
  if (!d) throw Error(p(169));
  c ? (a = bg(a, b, Xf), d.__reactInternalMemoizedMergedChildContext = a, E(Wf), E(H$1), G(H$1, a)) : E(Wf);
  G(Wf, c);
}
var eg = null, fg = false, gg = false;
function hg(a) {
  null === eg ? eg = [a] : eg.push(a);
}
function ig(a) {
  fg = true;
  hg(a);
}
function jg() {
  if (!gg && null !== eg) {
    gg = true;
    var a = 0, b = C;
    try {
      var c = eg;
      for (C = 1; a < c.length; a++) {
        var d = c[a];
        do
          d = d(true);
        while (null !== d);
      }
      eg = null;
      fg = false;
    } catch (e) {
      throw null !== eg && (eg = eg.slice(a + 1)), ac(fc, jg), e;
    } finally {
      C = b, gg = false;
    }
  }
  return null;
}
var kg = [], lg = 0, mg = null, ng = 0, og = [], pg = 0, qg = null, rg = 1, sg = "";
function tg(a, b) {
  kg[lg++] = ng;
  kg[lg++] = mg;
  mg = a;
  ng = b;
}
function ug(a, b, c) {
  og[pg++] = rg;
  og[pg++] = sg;
  og[pg++] = qg;
  qg = a;
  var d = rg;
  a = sg;
  var e = 32 - oc(d) - 1;
  d &= ~(1 << e);
  c += 1;
  var f2 = 32 - oc(b) + e;
  if (30 < f2) {
    var g = e - e % 5;
    f2 = (d & (1 << g) - 1).toString(32);
    d >>= g;
    e -= g;
    rg = 1 << 32 - oc(b) + e | c << e | d;
    sg = f2 + a;
  } else rg = 1 << f2 | c << e | d, sg = a;
}
function vg(a) {
  null !== a.return && (tg(a, 1), ug(a, 1, 0));
}
function wg(a) {
  for (; a === mg; ) mg = kg[--lg], kg[lg] = null, ng = kg[--lg], kg[lg] = null;
  for (; a === qg; ) qg = og[--pg], og[pg] = null, sg = og[--pg], og[pg] = null, rg = og[--pg], og[pg] = null;
}
var xg = null, yg = null, I = false, zg = null;
function Ag(a, b) {
  var c = Bg(5, null, null, 0);
  c.elementType = "DELETED";
  c.stateNode = b;
  c.return = a;
  b = a.deletions;
  null === b ? (a.deletions = [c], a.flags |= 16) : b.push(c);
}
function Cg(a, b) {
  switch (a.tag) {
    case 5:
      var c = a.type;
      b = 1 !== b.nodeType || c.toLowerCase() !== b.nodeName.toLowerCase() ? null : b;
      return null !== b ? (a.stateNode = b, xg = a, yg = Lf(b.firstChild), true) : false;
    case 6:
      return b = "" === a.pendingProps || 3 !== b.nodeType ? null : b, null !== b ? (a.stateNode = b, xg = a, yg = null, true) : false;
    case 13:
      return b = 8 !== b.nodeType ? null : b, null !== b ? (c = null !== qg ? { id: rg, overflow: sg } : null, a.memoizedState = { dehydrated: b, treeContext: c, retryLane: 1073741824 }, c = Bg(18, null, null, 0), c.stateNode = b, c.return = a, a.child = c, xg = a, yg = null, true) : false;
    default:
      return false;
  }
}
function Dg(a) {
  return 0 !== (a.mode & 1) && 0 === (a.flags & 128);
}
function Eg(a) {
  if (I) {
    var b = yg;
    if (b) {
      var c = b;
      if (!Cg(a, b)) {
        if (Dg(a)) throw Error(p(418));
        b = Lf(c.nextSibling);
        var d = xg;
        b && Cg(a, b) ? Ag(d, c) : (a.flags = a.flags & -4097 | 2, I = false, xg = a);
      }
    } else {
      if (Dg(a)) throw Error(p(418));
      a.flags = a.flags & -4097 | 2;
      I = false;
      xg = a;
    }
  }
}
function Fg(a) {
  for (a = a.return; null !== a && 5 !== a.tag && 3 !== a.tag && 13 !== a.tag; ) a = a.return;
  xg = a;
}
function Gg(a) {
  if (a !== xg) return false;
  if (!I) return Fg(a), I = true, false;
  var b;
  (b = 3 !== a.tag) && !(b = 5 !== a.tag) && (b = a.type, b = "head" !== b && "body" !== b && !Ef(a.type, a.memoizedProps));
  if (b && (b = yg)) {
    if (Dg(a)) throw Hg(), Error(p(418));
    for (; b; ) Ag(a, b), b = Lf(b.nextSibling);
  }
  Fg(a);
  if (13 === a.tag) {
    a = a.memoizedState;
    a = null !== a ? a.dehydrated : null;
    if (!a) throw Error(p(317));
    a: {
      a = a.nextSibling;
      for (b = 0; a; ) {
        if (8 === a.nodeType) {
          var c = a.data;
          if ("/$" === c) {
            if (0 === b) {
              yg = Lf(a.nextSibling);
              break a;
            }
            b--;
          } else "$" !== c && "$!" !== c && "$?" !== c || b++;
        }
        a = a.nextSibling;
      }
      yg = null;
    }
  } else yg = xg ? Lf(a.stateNode.nextSibling) : null;
  return true;
}
function Hg() {
  for (var a = yg; a; ) a = Lf(a.nextSibling);
}
function Ig() {
  yg = xg = null;
  I = false;
}
function Jg(a) {
  null === zg ? zg = [a] : zg.push(a);
}
var Kg = ua.ReactCurrentBatchConfig;
function Lg(a, b, c) {
  a = c.ref;
  if (null !== a && "function" !== typeof a && "object" !== typeof a) {
    if (c._owner) {
      c = c._owner;
      if (c) {
        if (1 !== c.tag) throw Error(p(309));
        var d = c.stateNode;
      }
      if (!d) throw Error(p(147, a));
      var e = d, f2 = "" + a;
      if (null !== b && null !== b.ref && "function" === typeof b.ref && b.ref._stringRef === f2) return b.ref;
      b = function(a2) {
        var b2 = e.refs;
        null === a2 ? delete b2[f2] : b2[f2] = a2;
      };
      b._stringRef = f2;
      return b;
    }
    if ("string" !== typeof a) throw Error(p(284));
    if (!c._owner) throw Error(p(290, a));
  }
  return a;
}
function Mg(a, b) {
  a = Object.prototype.toString.call(b);
  throw Error(p(31, "[object Object]" === a ? "object with keys {" + Object.keys(b).join(", ") + "}" : a));
}
function Ng(a) {
  var b = a._init;
  return b(a._payload);
}
function Og(a) {
  function b(b2, c2) {
    if (a) {
      var d2 = b2.deletions;
      null === d2 ? (b2.deletions = [c2], b2.flags |= 16) : d2.push(c2);
    }
  }
  function c(c2, d2) {
    if (!a) return null;
    for (; null !== d2; ) b(c2, d2), d2 = d2.sibling;
    return null;
  }
  function d(a2, b2) {
    for (a2 = /* @__PURE__ */ new Map(); null !== b2; ) null !== b2.key ? a2.set(b2.key, b2) : a2.set(b2.index, b2), b2 = b2.sibling;
    return a2;
  }
  function e(a2, b2) {
    a2 = Pg(a2, b2);
    a2.index = 0;
    a2.sibling = null;
    return a2;
  }
  function f2(b2, c2, d2) {
    b2.index = d2;
    if (!a) return b2.flags |= 1048576, c2;
    d2 = b2.alternate;
    if (null !== d2) return d2 = d2.index, d2 < c2 ? (b2.flags |= 2, c2) : d2;
    b2.flags |= 2;
    return c2;
  }
  function g(b2) {
    a && null === b2.alternate && (b2.flags |= 2);
    return b2;
  }
  function h2(a2, b2, c2, d2) {
    if (null === b2 || 6 !== b2.tag) return b2 = Qg(c2, a2.mode, d2), b2.return = a2, b2;
    b2 = e(b2, c2);
    b2.return = a2;
    return b2;
  }
  function k2(a2, b2, c2, d2) {
    var f3 = c2.type;
    if (f3 === ya) return m2(a2, b2, c2.props.children, d2, c2.key);
    if (null !== b2 && (b2.elementType === f3 || "object" === typeof f3 && null !== f3 && f3.$$typeof === Ha && Ng(f3) === b2.type)) return d2 = e(b2, c2.props), d2.ref = Lg(a2, b2, c2), d2.return = a2, d2;
    d2 = Rg(c2.type, c2.key, c2.props, null, a2.mode, d2);
    d2.ref = Lg(a2, b2, c2);
    d2.return = a2;
    return d2;
  }
  function l2(a2, b2, c2, d2) {
    if (null === b2 || 4 !== b2.tag || b2.stateNode.containerInfo !== c2.containerInfo || b2.stateNode.implementation !== c2.implementation) return b2 = Sg(c2, a2.mode, d2), b2.return = a2, b2;
    b2 = e(b2, c2.children || []);
    b2.return = a2;
    return b2;
  }
  function m2(a2, b2, c2, d2, f3) {
    if (null === b2 || 7 !== b2.tag) return b2 = Tg(c2, a2.mode, d2, f3), b2.return = a2, b2;
    b2 = e(b2, c2);
    b2.return = a2;
    return b2;
  }
  function q2(a2, b2, c2) {
    if ("string" === typeof b2 && "" !== b2 || "number" === typeof b2) return b2 = Qg("" + b2, a2.mode, c2), b2.return = a2, b2;
    if ("object" === typeof b2 && null !== b2) {
      switch (b2.$$typeof) {
        case va:
          return c2 = Rg(b2.type, b2.key, b2.props, null, a2.mode, c2), c2.ref = Lg(a2, null, b2), c2.return = a2, c2;
        case wa:
          return b2 = Sg(b2, a2.mode, c2), b2.return = a2, b2;
        case Ha:
          var d2 = b2._init;
          return q2(a2, d2(b2._payload), c2);
      }
      if (eb(b2) || Ka(b2)) return b2 = Tg(b2, a2.mode, c2, null), b2.return = a2, b2;
      Mg(a2, b2);
    }
    return null;
  }
  function r2(a2, b2, c2, d2) {
    var e2 = null !== b2 ? b2.key : null;
    if ("string" === typeof c2 && "" !== c2 || "number" === typeof c2) return null !== e2 ? null : h2(a2, b2, "" + c2, d2);
    if ("object" === typeof c2 && null !== c2) {
      switch (c2.$$typeof) {
        case va:
          return c2.key === e2 ? k2(a2, b2, c2, d2) : null;
        case wa:
          return c2.key === e2 ? l2(a2, b2, c2, d2) : null;
        case Ha:
          return e2 = c2._init, r2(
            a2,
            b2,
            e2(c2._payload),
            d2
          );
      }
      if (eb(c2) || Ka(c2)) return null !== e2 ? null : m2(a2, b2, c2, d2, null);
      Mg(a2, c2);
    }
    return null;
  }
  function y2(a2, b2, c2, d2, e2) {
    if ("string" === typeof d2 && "" !== d2 || "number" === typeof d2) return a2 = a2.get(c2) || null, h2(b2, a2, "" + d2, e2);
    if ("object" === typeof d2 && null !== d2) {
      switch (d2.$$typeof) {
        case va:
          return a2 = a2.get(null === d2.key ? c2 : d2.key) || null, k2(b2, a2, d2, e2);
        case wa:
          return a2 = a2.get(null === d2.key ? c2 : d2.key) || null, l2(b2, a2, d2, e2);
        case Ha:
          var f3 = d2._init;
          return y2(a2, b2, c2, f3(d2._payload), e2);
      }
      if (eb(d2) || Ka(d2)) return a2 = a2.get(c2) || null, m2(b2, a2, d2, e2, null);
      Mg(b2, d2);
    }
    return null;
  }
  function n2(e2, g2, h3, k3) {
    for (var l3 = null, m3 = null, u2 = g2, w2 = g2 = 0, x2 = null; null !== u2 && w2 < h3.length; w2++) {
      u2.index > w2 ? (x2 = u2, u2 = null) : x2 = u2.sibling;
      var n3 = r2(e2, u2, h3[w2], k3);
      if (null === n3) {
        null === u2 && (u2 = x2);
        break;
      }
      a && u2 && null === n3.alternate && b(e2, u2);
      g2 = f2(n3, g2, w2);
      null === m3 ? l3 = n3 : m3.sibling = n3;
      m3 = n3;
      u2 = x2;
    }
    if (w2 === h3.length) return c(e2, u2), I && tg(e2, w2), l3;
    if (null === u2) {
      for (; w2 < h3.length; w2++) u2 = q2(e2, h3[w2], k3), null !== u2 && (g2 = f2(u2, g2, w2), null === m3 ? l3 = u2 : m3.sibling = u2, m3 = u2);
      I && tg(e2, w2);
      return l3;
    }
    for (u2 = d(e2, u2); w2 < h3.length; w2++) x2 = y2(u2, e2, w2, h3[w2], k3), null !== x2 && (a && null !== x2.alternate && u2.delete(null === x2.key ? w2 : x2.key), g2 = f2(x2, g2, w2), null === m3 ? l3 = x2 : m3.sibling = x2, m3 = x2);
    a && u2.forEach(function(a2) {
      return b(e2, a2);
    });
    I && tg(e2, w2);
    return l3;
  }
  function t2(e2, g2, h3, k3) {
    var l3 = Ka(h3);
    if ("function" !== typeof l3) throw Error(p(150));
    h3 = l3.call(h3);
    if (null == h3) throw Error(p(151));
    for (var u2 = l3 = null, m3 = g2, w2 = g2 = 0, x2 = null, n3 = h3.next(); null !== m3 && !n3.done; w2++, n3 = h3.next()) {
      m3.index > w2 ? (x2 = m3, m3 = null) : x2 = m3.sibling;
      var t3 = r2(e2, m3, n3.value, k3);
      if (null === t3) {
        null === m3 && (m3 = x2);
        break;
      }
      a && m3 && null === t3.alternate && b(e2, m3);
      g2 = f2(t3, g2, w2);
      null === u2 ? l3 = t3 : u2.sibling = t3;
      u2 = t3;
      m3 = x2;
    }
    if (n3.done) return c(
      e2,
      m3
    ), I && tg(e2, w2), l3;
    if (null === m3) {
      for (; !n3.done; w2++, n3 = h3.next()) n3 = q2(e2, n3.value, k3), null !== n3 && (g2 = f2(n3, g2, w2), null === u2 ? l3 = n3 : u2.sibling = n3, u2 = n3);
      I && tg(e2, w2);
      return l3;
    }
    for (m3 = d(e2, m3); !n3.done; w2++, n3 = h3.next()) n3 = y2(m3, e2, w2, n3.value, k3), null !== n3 && (a && null !== n3.alternate && m3.delete(null === n3.key ? w2 : n3.key), g2 = f2(n3, g2, w2), null === u2 ? l3 = n3 : u2.sibling = n3, u2 = n3);
    a && m3.forEach(function(a2) {
      return b(e2, a2);
    });
    I && tg(e2, w2);
    return l3;
  }
  function J2(a2, d2, f3, h3) {
    "object" === typeof f3 && null !== f3 && f3.type === ya && null === f3.key && (f3 = f3.props.children);
    if ("object" === typeof f3 && null !== f3) {
      switch (f3.$$typeof) {
        case va:
          a: {
            for (var k3 = f3.key, l3 = d2; null !== l3; ) {
              if (l3.key === k3) {
                k3 = f3.type;
                if (k3 === ya) {
                  if (7 === l3.tag) {
                    c(a2, l3.sibling);
                    d2 = e(l3, f3.props.children);
                    d2.return = a2;
                    a2 = d2;
                    break a;
                  }
                } else if (l3.elementType === k3 || "object" === typeof k3 && null !== k3 && k3.$$typeof === Ha && Ng(k3) === l3.type) {
                  c(a2, l3.sibling);
                  d2 = e(l3, f3.props);
                  d2.ref = Lg(a2, l3, f3);
                  d2.return = a2;
                  a2 = d2;
                  break a;
                }
                c(a2, l3);
                break;
              } else b(a2, l3);
              l3 = l3.sibling;
            }
            f3.type === ya ? (d2 = Tg(f3.props.children, a2.mode, h3, f3.key), d2.return = a2, a2 = d2) : (h3 = Rg(f3.type, f3.key, f3.props, null, a2.mode, h3), h3.ref = Lg(a2, d2, f3), h3.return = a2, a2 = h3);
          }
          return g(a2);
        case wa:
          a: {
            for (l3 = f3.key; null !== d2; ) {
              if (d2.key === l3) if (4 === d2.tag && d2.stateNode.containerInfo === f3.containerInfo && d2.stateNode.implementation === f3.implementation) {
                c(a2, d2.sibling);
                d2 = e(d2, f3.children || []);
                d2.return = a2;
                a2 = d2;
                break a;
              } else {
                c(a2, d2);
                break;
              }
              else b(a2, d2);
              d2 = d2.sibling;
            }
            d2 = Sg(f3, a2.mode, h3);
            d2.return = a2;
            a2 = d2;
          }
          return g(a2);
        case Ha:
          return l3 = f3._init, J2(a2, d2, l3(f3._payload), h3);
      }
      if (eb(f3)) return n2(a2, d2, f3, h3);
      if (Ka(f3)) return t2(a2, d2, f3, h3);
      Mg(a2, f3);
    }
    return "string" === typeof f3 && "" !== f3 || "number" === typeof f3 ? (f3 = "" + f3, null !== d2 && 6 === d2.tag ? (c(a2, d2.sibling), d2 = e(d2, f3), d2.return = a2, a2 = d2) : (c(a2, d2), d2 = Qg(f3, a2.mode, h3), d2.return = a2, a2 = d2), g(a2)) : c(a2, d2);
  }
  return J2;
}
var Ug = Og(true), Vg = Og(false), Wg = Uf(null), Xg = null, Yg = null, Zg = null;
function $g() {
  Zg = Yg = Xg = null;
}
function ah(a) {
  var b = Wg.current;
  E(Wg);
  a._currentValue = b;
}
function bh(a, b, c) {
  for (; null !== a; ) {
    var d = a.alternate;
    (a.childLanes & b) !== b ? (a.childLanes |= b, null !== d && (d.childLanes |= b)) : null !== d && (d.childLanes & b) !== b && (d.childLanes |= b);
    if (a === c) break;
    a = a.return;
  }
}
function ch(a, b) {
  Xg = a;
  Zg = Yg = null;
  a = a.dependencies;
  null !== a && null !== a.firstContext && (0 !== (a.lanes & b) && (dh = true), a.firstContext = null);
}
function eh(a) {
  var b = a._currentValue;
  if (Zg !== a) if (a = { context: a, memoizedValue: b, next: null }, null === Yg) {
    if (null === Xg) throw Error(p(308));
    Yg = a;
    Xg.dependencies = { lanes: 0, firstContext: a };
  } else Yg = Yg.next = a;
  return b;
}
var fh = null;
function gh(a) {
  null === fh ? fh = [a] : fh.push(a);
}
function hh(a, b, c, d) {
  var e = b.interleaved;
  null === e ? (c.next = c, gh(b)) : (c.next = e.next, e.next = c);
  b.interleaved = c;
  return ih(a, d);
}
function ih(a, b) {
  a.lanes |= b;
  var c = a.alternate;
  null !== c && (c.lanes |= b);
  c = a;
  for (a = a.return; null !== a; ) a.childLanes |= b, c = a.alternate, null !== c && (c.childLanes |= b), c = a, a = a.return;
  return 3 === c.tag ? c.stateNode : null;
}
var jh = false;
function kh(a) {
  a.updateQueue = { baseState: a.memoizedState, firstBaseUpdate: null, lastBaseUpdate: null, shared: { pending: null, interleaved: null, lanes: 0 }, effects: null };
}
function lh(a, b) {
  a = a.updateQueue;
  b.updateQueue === a && (b.updateQueue = { baseState: a.baseState, firstBaseUpdate: a.firstBaseUpdate, lastBaseUpdate: a.lastBaseUpdate, shared: a.shared, effects: a.effects });
}
function mh(a, b) {
  return { eventTime: a, lane: b, tag: 0, payload: null, callback: null, next: null };
}
function nh(a, b, c) {
  var d = a.updateQueue;
  if (null === d) return null;
  d = d.shared;
  if (0 !== (K & 2)) {
    var e = d.pending;
    null === e ? b.next = b : (b.next = e.next, e.next = b);
    d.pending = b;
    return ih(a, c);
  }
  e = d.interleaved;
  null === e ? (b.next = b, gh(d)) : (b.next = e.next, e.next = b);
  d.interleaved = b;
  return ih(a, c);
}
function oh(a, b, c) {
  b = b.updateQueue;
  if (null !== b && (b = b.shared, 0 !== (c & 4194240))) {
    var d = b.lanes;
    d &= a.pendingLanes;
    c |= d;
    b.lanes = c;
    Cc(a, c);
  }
}
function ph(a, b) {
  var c = a.updateQueue, d = a.alternate;
  if (null !== d && (d = d.updateQueue, c === d)) {
    var e = null, f2 = null;
    c = c.firstBaseUpdate;
    if (null !== c) {
      do {
        var g = { eventTime: c.eventTime, lane: c.lane, tag: c.tag, payload: c.payload, callback: c.callback, next: null };
        null === f2 ? e = f2 = g : f2 = f2.next = g;
        c = c.next;
      } while (null !== c);
      null === f2 ? e = f2 = b : f2 = f2.next = b;
    } else e = f2 = b;
    c = { baseState: d.baseState, firstBaseUpdate: e, lastBaseUpdate: f2, shared: d.shared, effects: d.effects };
    a.updateQueue = c;
    return;
  }
  a = c.lastBaseUpdate;
  null === a ? c.firstBaseUpdate = b : a.next = b;
  c.lastBaseUpdate = b;
}
function qh(a, b, c, d) {
  var e = a.updateQueue;
  jh = false;
  var f2 = e.firstBaseUpdate, g = e.lastBaseUpdate, h2 = e.shared.pending;
  if (null !== h2) {
    e.shared.pending = null;
    var k2 = h2, l2 = k2.next;
    k2.next = null;
    null === g ? f2 = l2 : g.next = l2;
    g = k2;
    var m2 = a.alternate;
    null !== m2 && (m2 = m2.updateQueue, h2 = m2.lastBaseUpdate, h2 !== g && (null === h2 ? m2.firstBaseUpdate = l2 : h2.next = l2, m2.lastBaseUpdate = k2));
  }
  if (null !== f2) {
    var q2 = e.baseState;
    g = 0;
    m2 = l2 = k2 = null;
    h2 = f2;
    do {
      var r2 = h2.lane, y2 = h2.eventTime;
      if ((d & r2) === r2) {
        null !== m2 && (m2 = m2.next = {
          eventTime: y2,
          lane: 0,
          tag: h2.tag,
          payload: h2.payload,
          callback: h2.callback,
          next: null
        });
        a: {
          var n2 = a, t2 = h2;
          r2 = b;
          y2 = c;
          switch (t2.tag) {
            case 1:
              n2 = t2.payload;
              if ("function" === typeof n2) {
                q2 = n2.call(y2, q2, r2);
                break a;
              }
              q2 = n2;
              break a;
            case 3:
              n2.flags = n2.flags & -65537 | 128;
            case 0:
              n2 = t2.payload;
              r2 = "function" === typeof n2 ? n2.call(y2, q2, r2) : n2;
              if (null === r2 || void 0 === r2) break a;
              q2 = A({}, q2, r2);
              break a;
            case 2:
              jh = true;
          }
        }
        null !== h2.callback && 0 !== h2.lane && (a.flags |= 64, r2 = e.effects, null === r2 ? e.effects = [h2] : r2.push(h2));
      } else y2 = { eventTime: y2, lane: r2, tag: h2.tag, payload: h2.payload, callback: h2.callback, next: null }, null === m2 ? (l2 = m2 = y2, k2 = q2) : m2 = m2.next = y2, g |= r2;
      h2 = h2.next;
      if (null === h2) if (h2 = e.shared.pending, null === h2) break;
      else r2 = h2, h2 = r2.next, r2.next = null, e.lastBaseUpdate = r2, e.shared.pending = null;
    } while (1);
    null === m2 && (k2 = q2);
    e.baseState = k2;
    e.firstBaseUpdate = l2;
    e.lastBaseUpdate = m2;
    b = e.shared.interleaved;
    if (null !== b) {
      e = b;
      do
        g |= e.lane, e = e.next;
      while (e !== b);
    } else null === f2 && (e.shared.lanes = 0);
    rh |= g;
    a.lanes = g;
    a.memoizedState = q2;
  }
}
function sh(a, b, c) {
  a = b.effects;
  b.effects = null;
  if (null !== a) for (b = 0; b < a.length; b++) {
    var d = a[b], e = d.callback;
    if (null !== e) {
      d.callback = null;
      d = c;
      if ("function" !== typeof e) throw Error(p(191, e));
      e.call(d);
    }
  }
}
var th = {}, uh = Uf(th), vh = Uf(th), wh = Uf(th);
function xh(a) {
  if (a === th) throw Error(p(174));
  return a;
}
function yh(a, b) {
  G(wh, b);
  G(vh, a);
  G(uh, th);
  a = b.nodeType;
  switch (a) {
    case 9:
    case 11:
      b = (b = b.documentElement) ? b.namespaceURI : lb(null, "");
      break;
    default:
      a = 8 === a ? b.parentNode : b, b = a.namespaceURI || null, a = a.tagName, b = lb(b, a);
  }
  E(uh);
  G(uh, b);
}
function zh() {
  E(uh);
  E(vh);
  E(wh);
}
function Ah(a) {
  xh(wh.current);
  var b = xh(uh.current);
  var c = lb(b, a.type);
  b !== c && (G(vh, a), G(uh, c));
}
function Bh(a) {
  vh.current === a && (E(uh), E(vh));
}
var L = Uf(0);
function Ch(a) {
  for (var b = a; null !== b; ) {
    if (13 === b.tag) {
      var c = b.memoizedState;
      if (null !== c && (c = c.dehydrated, null === c || "$?" === c.data || "$!" === c.data)) return b;
    } else if (19 === b.tag && void 0 !== b.memoizedProps.revealOrder) {
      if (0 !== (b.flags & 128)) return b;
    } else if (null !== b.child) {
      b.child.return = b;
      b = b.child;
      continue;
    }
    if (b === a) break;
    for (; null === b.sibling; ) {
      if (null === b.return || b.return === a) return null;
      b = b.return;
    }
    b.sibling.return = b.return;
    b = b.sibling;
  }
  return null;
}
var Dh = [];
function Eh() {
  for (var a = 0; a < Dh.length; a++) Dh[a]._workInProgressVersionPrimary = null;
  Dh.length = 0;
}
var Fh = ua.ReactCurrentDispatcher, Gh = ua.ReactCurrentBatchConfig, Hh = 0, M = null, N = null, O = null, Ih = false, Jh = false, Kh = 0, Lh = 0;
function P() {
  throw Error(p(321));
}
function Mh(a, b) {
  if (null === b) return false;
  for (var c = 0; c < b.length && c < a.length; c++) if (!He$1(a[c], b[c])) return false;
  return true;
}
function Nh(a, b, c, d, e, f2) {
  Hh = f2;
  M = b;
  b.memoizedState = null;
  b.updateQueue = null;
  b.lanes = 0;
  Fh.current = null === a || null === a.memoizedState ? Oh : Ph;
  a = c(d, e);
  if (Jh) {
    f2 = 0;
    do {
      Jh = false;
      Kh = 0;
      if (25 <= f2) throw Error(p(301));
      f2 += 1;
      O = N = null;
      b.updateQueue = null;
      Fh.current = Qh;
      a = c(d, e);
    } while (Jh);
  }
  Fh.current = Rh;
  b = null !== N && null !== N.next;
  Hh = 0;
  O = N = M = null;
  Ih = false;
  if (b) throw Error(p(300));
  return a;
}
function Sh() {
  var a = 0 !== Kh;
  Kh = 0;
  return a;
}
function Th() {
  var a = { memoizedState: null, baseState: null, baseQueue: null, queue: null, next: null };
  null === O ? M.memoizedState = O = a : O = O.next = a;
  return O;
}
function Uh() {
  if (null === N) {
    var a = M.alternate;
    a = null !== a ? a.memoizedState : null;
  } else a = N.next;
  var b = null === O ? M.memoizedState : O.next;
  if (null !== b) O = b, N = a;
  else {
    if (null === a) throw Error(p(310));
    N = a;
    a = { memoizedState: N.memoizedState, baseState: N.baseState, baseQueue: N.baseQueue, queue: N.queue, next: null };
    null === O ? M.memoizedState = O = a : O = O.next = a;
  }
  return O;
}
function Vh(a, b) {
  return "function" === typeof b ? b(a) : b;
}
function Wh(a) {
  var b = Uh(), c = b.queue;
  if (null === c) throw Error(p(311));
  c.lastRenderedReducer = a;
  var d = N, e = d.baseQueue, f2 = c.pending;
  if (null !== f2) {
    if (null !== e) {
      var g = e.next;
      e.next = f2.next;
      f2.next = g;
    }
    d.baseQueue = e = f2;
    c.pending = null;
  }
  if (null !== e) {
    f2 = e.next;
    d = d.baseState;
    var h2 = g = null, k2 = null, l2 = f2;
    do {
      var m2 = l2.lane;
      if ((Hh & m2) === m2) null !== k2 && (k2 = k2.next = { lane: 0, action: l2.action, hasEagerState: l2.hasEagerState, eagerState: l2.eagerState, next: null }), d = l2.hasEagerState ? l2.eagerState : a(d, l2.action);
      else {
        var q2 = {
          lane: m2,
          action: l2.action,
          hasEagerState: l2.hasEagerState,
          eagerState: l2.eagerState,
          next: null
        };
        null === k2 ? (h2 = k2 = q2, g = d) : k2 = k2.next = q2;
        M.lanes |= m2;
        rh |= m2;
      }
      l2 = l2.next;
    } while (null !== l2 && l2 !== f2);
    null === k2 ? g = d : k2.next = h2;
    He$1(d, b.memoizedState) || (dh = true);
    b.memoizedState = d;
    b.baseState = g;
    b.baseQueue = k2;
    c.lastRenderedState = d;
  }
  a = c.interleaved;
  if (null !== a) {
    e = a;
    do
      f2 = e.lane, M.lanes |= f2, rh |= f2, e = e.next;
    while (e !== a);
  } else null === e && (c.lanes = 0);
  return [b.memoizedState, c.dispatch];
}
function Xh(a) {
  var b = Uh(), c = b.queue;
  if (null === c) throw Error(p(311));
  c.lastRenderedReducer = a;
  var d = c.dispatch, e = c.pending, f2 = b.memoizedState;
  if (null !== e) {
    c.pending = null;
    var g = e = e.next;
    do
      f2 = a(f2, g.action), g = g.next;
    while (g !== e);
    He$1(f2, b.memoizedState) || (dh = true);
    b.memoizedState = f2;
    null === b.baseQueue && (b.baseState = f2);
    c.lastRenderedState = f2;
  }
  return [f2, d];
}
function Yh() {
}
function Zh(a, b) {
  var c = M, d = Uh(), e = b(), f2 = !He$1(d.memoizedState, e);
  f2 && (d.memoizedState = e, dh = true);
  d = d.queue;
  $h(ai.bind(null, c, d, a), [a]);
  if (d.getSnapshot !== b || f2 || null !== O && O.memoizedState.tag & 1) {
    c.flags |= 2048;
    bi(9, ci.bind(null, c, d, e, b), void 0, null);
    if (null === Q) throw Error(p(349));
    0 !== (Hh & 30) || di(c, b, e);
  }
  return e;
}
function di(a, b, c) {
  a.flags |= 16384;
  a = { getSnapshot: b, value: c };
  b = M.updateQueue;
  null === b ? (b = { lastEffect: null, stores: null }, M.updateQueue = b, b.stores = [a]) : (c = b.stores, null === c ? b.stores = [a] : c.push(a));
}
function ci(a, b, c, d) {
  b.value = c;
  b.getSnapshot = d;
  ei(b) && fi(a);
}
function ai(a, b, c) {
  return c(function() {
    ei(b) && fi(a);
  });
}
function ei(a) {
  var b = a.getSnapshot;
  a = a.value;
  try {
    var c = b();
    return !He$1(a, c);
  } catch (d) {
    return true;
  }
}
function fi(a) {
  var b = ih(a, 1);
  null !== b && gi(b, a, 1, -1);
}
function hi(a) {
  var b = Th();
  "function" === typeof a && (a = a());
  b.memoizedState = b.baseState = a;
  a = { pending: null, interleaved: null, lanes: 0, dispatch: null, lastRenderedReducer: Vh, lastRenderedState: a };
  b.queue = a;
  a = a.dispatch = ii.bind(null, M, a);
  return [b.memoizedState, a];
}
function bi(a, b, c, d) {
  a = { tag: a, create: b, destroy: c, deps: d, next: null };
  b = M.updateQueue;
  null === b ? (b = { lastEffect: null, stores: null }, M.updateQueue = b, b.lastEffect = a.next = a) : (c = b.lastEffect, null === c ? b.lastEffect = a.next = a : (d = c.next, c.next = a, a.next = d, b.lastEffect = a));
  return a;
}
function ji() {
  return Uh().memoizedState;
}
function ki(a, b, c, d) {
  var e = Th();
  M.flags |= a;
  e.memoizedState = bi(1 | b, c, void 0, void 0 === d ? null : d);
}
function li(a, b, c, d) {
  var e = Uh();
  d = void 0 === d ? null : d;
  var f2 = void 0;
  if (null !== N) {
    var g = N.memoizedState;
    f2 = g.destroy;
    if (null !== d && Mh(d, g.deps)) {
      e.memoizedState = bi(b, c, f2, d);
      return;
    }
  }
  M.flags |= a;
  e.memoizedState = bi(1 | b, c, f2, d);
}
function mi(a, b) {
  return ki(8390656, 8, a, b);
}
function $h(a, b) {
  return li(2048, 8, a, b);
}
function ni(a, b) {
  return li(4, 2, a, b);
}
function oi(a, b) {
  return li(4, 4, a, b);
}
function pi(a, b) {
  if ("function" === typeof b) return a = a(), b(a), function() {
    b(null);
  };
  if (null !== b && void 0 !== b) return a = a(), b.current = a, function() {
    b.current = null;
  };
}
function qi(a, b, c) {
  c = null !== c && void 0 !== c ? c.concat([a]) : null;
  return li(4, 4, pi.bind(null, b, a), c);
}
function ri() {
}
function si(a, b) {
  var c = Uh();
  b = void 0 === b ? null : b;
  var d = c.memoizedState;
  if (null !== d && null !== b && Mh(b, d[1])) return d[0];
  c.memoizedState = [a, b];
  return a;
}
function ti(a, b) {
  var c = Uh();
  b = void 0 === b ? null : b;
  var d = c.memoizedState;
  if (null !== d && null !== b && Mh(b, d[1])) return d[0];
  a = a();
  c.memoizedState = [a, b];
  return a;
}
function ui(a, b, c) {
  if (0 === (Hh & 21)) return a.baseState && (a.baseState = false, dh = true), a.memoizedState = c;
  He$1(c, b) || (c = yc(), M.lanes |= c, rh |= c, a.baseState = true);
  return b;
}
function vi(a, b) {
  var c = C;
  C = 0 !== c && 4 > c ? c : 4;
  a(true);
  var d = Gh.transition;
  Gh.transition = {};
  try {
    a(false), b();
  } finally {
    C = c, Gh.transition = d;
  }
}
function wi() {
  return Uh().memoizedState;
}
function xi(a, b, c) {
  var d = yi(a);
  c = { lane: d, action: c, hasEagerState: false, eagerState: null, next: null };
  if (zi(a)) Ai(b, c);
  else if (c = hh(a, b, c, d), null !== c) {
    var e = R();
    gi(c, a, d, e);
    Bi(c, b, d);
  }
}
function ii(a, b, c) {
  var d = yi(a), e = { lane: d, action: c, hasEagerState: false, eagerState: null, next: null };
  if (zi(a)) Ai(b, e);
  else {
    var f2 = a.alternate;
    if (0 === a.lanes && (null === f2 || 0 === f2.lanes) && (f2 = b.lastRenderedReducer, null !== f2)) try {
      var g = b.lastRenderedState, h2 = f2(g, c);
      e.hasEagerState = true;
      e.eagerState = h2;
      if (He$1(h2, g)) {
        var k2 = b.interleaved;
        null === k2 ? (e.next = e, gh(b)) : (e.next = k2.next, k2.next = e);
        b.interleaved = e;
        return;
      }
    } catch (l2) {
    } finally {
    }
    c = hh(a, b, e, d);
    null !== c && (e = R(), gi(c, a, d, e), Bi(c, b, d));
  }
}
function zi(a) {
  var b = a.alternate;
  return a === M || null !== b && b === M;
}
function Ai(a, b) {
  Jh = Ih = true;
  var c = a.pending;
  null === c ? b.next = b : (b.next = c.next, c.next = b);
  a.pending = b;
}
function Bi(a, b, c) {
  if (0 !== (c & 4194240)) {
    var d = b.lanes;
    d &= a.pendingLanes;
    c |= d;
    b.lanes = c;
    Cc(a, c);
  }
}
var Rh = { readContext: eh, useCallback: P, useContext: P, useEffect: P, useImperativeHandle: P, useInsertionEffect: P, useLayoutEffect: P, useMemo: P, useReducer: P, useRef: P, useState: P, useDebugValue: P, useDeferredValue: P, useTransition: P, useMutableSource: P, useSyncExternalStore: P, useId: P, unstable_isNewReconciler: false }, Oh = { readContext: eh, useCallback: function(a, b) {
  Th().memoizedState = [a, void 0 === b ? null : b];
  return a;
}, useContext: eh, useEffect: mi, useImperativeHandle: function(a, b, c) {
  c = null !== c && void 0 !== c ? c.concat([a]) : null;
  return ki(
    4194308,
    4,
    pi.bind(null, b, a),
    c
  );
}, useLayoutEffect: function(a, b) {
  return ki(4194308, 4, a, b);
}, useInsertionEffect: function(a, b) {
  return ki(4, 2, a, b);
}, useMemo: function(a, b) {
  var c = Th();
  b = void 0 === b ? null : b;
  a = a();
  c.memoizedState = [a, b];
  return a;
}, useReducer: function(a, b, c) {
  var d = Th();
  b = void 0 !== c ? c(b) : b;
  d.memoizedState = d.baseState = b;
  a = { pending: null, interleaved: null, lanes: 0, dispatch: null, lastRenderedReducer: a, lastRenderedState: b };
  d.queue = a;
  a = a.dispatch = xi.bind(null, M, a);
  return [d.memoizedState, a];
}, useRef: function(a) {
  var b = Th();
  a = { current: a };
  return b.memoizedState = a;
}, useState: hi, useDebugValue: ri, useDeferredValue: function(a) {
  return Th().memoizedState = a;
}, useTransition: function() {
  var a = hi(false), b = a[0];
  a = vi.bind(null, a[1]);
  Th().memoizedState = a;
  return [b, a];
}, useMutableSource: function() {
}, useSyncExternalStore: function(a, b, c) {
  var d = M, e = Th();
  if (I) {
    if (void 0 === c) throw Error(p(407));
    c = c();
  } else {
    c = b();
    if (null === Q) throw Error(p(349));
    0 !== (Hh & 30) || di(d, b, c);
  }
  e.memoizedState = c;
  var f2 = { value: c, getSnapshot: b };
  e.queue = f2;
  mi(ai.bind(
    null,
    d,
    f2,
    a
  ), [a]);
  d.flags |= 2048;
  bi(9, ci.bind(null, d, f2, c, b), void 0, null);
  return c;
}, useId: function() {
  var a = Th(), b = Q.identifierPrefix;
  if (I) {
    var c = sg;
    var d = rg;
    c = (d & ~(1 << 32 - oc(d) - 1)).toString(32) + c;
    b = ":" + b + "R" + c;
    c = Kh++;
    0 < c && (b += "H" + c.toString(32));
    b += ":";
  } else c = Lh++, b = ":" + b + "r" + c.toString(32) + ":";
  return a.memoizedState = b;
}, unstable_isNewReconciler: false }, Ph = {
  readContext: eh,
  useCallback: si,
  useContext: eh,
  useEffect: $h,
  useImperativeHandle: qi,
  useInsertionEffect: ni,
  useLayoutEffect: oi,
  useMemo: ti,
  useReducer: Wh,
  useRef: ji,
  useState: function() {
    return Wh(Vh);
  },
  useDebugValue: ri,
  useDeferredValue: function(a) {
    var b = Uh();
    return ui(b, N.memoizedState, a);
  },
  useTransition: function() {
    var a = Wh(Vh)[0], b = Uh().memoizedState;
    return [a, b];
  },
  useMutableSource: Yh,
  useSyncExternalStore: Zh,
  useId: wi,
  unstable_isNewReconciler: false
}, Qh = { readContext: eh, useCallback: si, useContext: eh, useEffect: $h, useImperativeHandle: qi, useInsertionEffect: ni, useLayoutEffect: oi, useMemo: ti, useReducer: Xh, useRef: ji, useState: function() {
  return Xh(Vh);
}, useDebugValue: ri, useDeferredValue: function(a) {
  var b = Uh();
  return null === N ? b.memoizedState = a : ui(b, N.memoizedState, a);
}, useTransition: function() {
  var a = Xh(Vh)[0], b = Uh().memoizedState;
  return [a, b];
}, useMutableSource: Yh, useSyncExternalStore: Zh, useId: wi, unstable_isNewReconciler: false };
function Ci(a, b) {
  if (a && a.defaultProps) {
    b = A({}, b);
    a = a.defaultProps;
    for (var c in a) void 0 === b[c] && (b[c] = a[c]);
    return b;
  }
  return b;
}
function Di(a, b, c, d) {
  b = a.memoizedState;
  c = c(d, b);
  c = null === c || void 0 === c ? b : A({}, b, c);
  a.memoizedState = c;
  0 === a.lanes && (a.updateQueue.baseState = c);
}
var Ei = { isMounted: function(a) {
  return (a = a._reactInternals) ? Vb(a) === a : false;
}, enqueueSetState: function(a, b, c) {
  a = a._reactInternals;
  var d = R(), e = yi(a), f2 = mh(d, e);
  f2.payload = b;
  void 0 !== c && null !== c && (f2.callback = c);
  b = nh(a, f2, e);
  null !== b && (gi(b, a, e, d), oh(b, a, e));
}, enqueueReplaceState: function(a, b, c) {
  a = a._reactInternals;
  var d = R(), e = yi(a), f2 = mh(d, e);
  f2.tag = 1;
  f2.payload = b;
  void 0 !== c && null !== c && (f2.callback = c);
  b = nh(a, f2, e);
  null !== b && (gi(b, a, e, d), oh(b, a, e));
}, enqueueForceUpdate: function(a, b) {
  a = a._reactInternals;
  var c = R(), d = yi(a), e = mh(c, d);
  e.tag = 2;
  void 0 !== b && null !== b && (e.callback = b);
  b = nh(a, e, d);
  null !== b && (gi(b, a, d, c), oh(b, a, d));
} };
function Fi(a, b, c, d, e, f2, g) {
  a = a.stateNode;
  return "function" === typeof a.shouldComponentUpdate ? a.shouldComponentUpdate(d, f2, g) : b.prototype && b.prototype.isPureReactComponent ? !Ie(c, d) || !Ie(e, f2) : true;
}
function Gi(a, b, c) {
  var d = false, e = Vf;
  var f2 = b.contextType;
  "object" === typeof f2 && null !== f2 ? f2 = eh(f2) : (e = Zf(b) ? Xf : H$1.current, d = b.contextTypes, f2 = (d = null !== d && void 0 !== d) ? Yf(a, e) : Vf);
  b = new b(c, f2);
  a.memoizedState = null !== b.state && void 0 !== b.state ? b.state : null;
  b.updater = Ei;
  a.stateNode = b;
  b._reactInternals = a;
  d && (a = a.stateNode, a.__reactInternalMemoizedUnmaskedChildContext = e, a.__reactInternalMemoizedMaskedChildContext = f2);
  return b;
}
function Hi(a, b, c, d) {
  a = b.state;
  "function" === typeof b.componentWillReceiveProps && b.componentWillReceiveProps(c, d);
  "function" === typeof b.UNSAFE_componentWillReceiveProps && b.UNSAFE_componentWillReceiveProps(c, d);
  b.state !== a && Ei.enqueueReplaceState(b, b.state, null);
}
function Ii(a, b, c, d) {
  var e = a.stateNode;
  e.props = c;
  e.state = a.memoizedState;
  e.refs = {};
  kh(a);
  var f2 = b.contextType;
  "object" === typeof f2 && null !== f2 ? e.context = eh(f2) : (f2 = Zf(b) ? Xf : H$1.current, e.context = Yf(a, f2));
  e.state = a.memoizedState;
  f2 = b.getDerivedStateFromProps;
  "function" === typeof f2 && (Di(a, b, f2, c), e.state = a.memoizedState);
  "function" === typeof b.getDerivedStateFromProps || "function" === typeof e.getSnapshotBeforeUpdate || "function" !== typeof e.UNSAFE_componentWillMount && "function" !== typeof e.componentWillMount || (b = e.state, "function" === typeof e.componentWillMount && e.componentWillMount(), "function" === typeof e.UNSAFE_componentWillMount && e.UNSAFE_componentWillMount(), b !== e.state && Ei.enqueueReplaceState(e, e.state, null), qh(a, c, e, d), e.state = a.memoizedState);
  "function" === typeof e.componentDidMount && (a.flags |= 4194308);
}
function Ji(a, b) {
  try {
    var c = "", d = b;
    do
      c += Pa(d), d = d.return;
    while (d);
    var e = c;
  } catch (f2) {
    e = "\nError generating stack: " + f2.message + "\n" + f2.stack;
  }
  return { value: a, source: b, stack: e, digest: null };
}
function Ki(a, b, c) {
  return { value: a, source: null, stack: null != c ? c : null, digest: null != b ? b : null };
}
function Li(a, b) {
  try {
    console.error(b.value);
  } catch (c) {
    setTimeout(function() {
      throw c;
    });
  }
}
var Mi = "function" === typeof WeakMap ? WeakMap : Map;
function Ni(a, b, c) {
  c = mh(-1, c);
  c.tag = 3;
  c.payload = { element: null };
  var d = b.value;
  c.callback = function() {
    Oi || (Oi = true, Pi = d);
    Li(a, b);
  };
  return c;
}
function Qi(a, b, c) {
  c = mh(-1, c);
  c.tag = 3;
  var d = a.type.getDerivedStateFromError;
  if ("function" === typeof d) {
    var e = b.value;
    c.payload = function() {
      return d(e);
    };
    c.callback = function() {
      Li(a, b);
    };
  }
  var f2 = a.stateNode;
  null !== f2 && "function" === typeof f2.componentDidCatch && (c.callback = function() {
    Li(a, b);
    "function" !== typeof d && (null === Ri ? Ri = /* @__PURE__ */ new Set([this]) : Ri.add(this));
    var c2 = b.stack;
    this.componentDidCatch(b.value, { componentStack: null !== c2 ? c2 : "" });
  });
  return c;
}
function Si(a, b, c) {
  var d = a.pingCache;
  if (null === d) {
    d = a.pingCache = new Mi();
    var e = /* @__PURE__ */ new Set();
    d.set(b, e);
  } else e = d.get(b), void 0 === e && (e = /* @__PURE__ */ new Set(), d.set(b, e));
  e.has(c) || (e.add(c), a = Ti.bind(null, a, b, c), b.then(a, a));
}
function Ui(a) {
  do {
    var b;
    if (b = 13 === a.tag) b = a.memoizedState, b = null !== b ? null !== b.dehydrated ? true : false : true;
    if (b) return a;
    a = a.return;
  } while (null !== a);
  return null;
}
function Vi(a, b, c, d, e) {
  if (0 === (a.mode & 1)) return a === b ? a.flags |= 65536 : (a.flags |= 128, c.flags |= 131072, c.flags &= -52805, 1 === c.tag && (null === c.alternate ? c.tag = 17 : (b = mh(-1, 1), b.tag = 2, nh(c, b, 1))), c.lanes |= 1), a;
  a.flags |= 65536;
  a.lanes = e;
  return a;
}
var Wi = ua.ReactCurrentOwner, dh = false;
function Xi(a, b, c, d) {
  b.child = null === a ? Vg(b, null, c, d) : Ug(b, a.child, c, d);
}
function Yi(a, b, c, d, e) {
  c = c.render;
  var f2 = b.ref;
  ch(b, e);
  d = Nh(a, b, c, d, f2, e);
  c = Sh();
  if (null !== a && !dh) return b.updateQueue = a.updateQueue, b.flags &= -2053, a.lanes &= ~e, Zi(a, b, e);
  I && c && vg(b);
  b.flags |= 1;
  Xi(a, b, d, e);
  return b.child;
}
function $i(a, b, c, d, e) {
  if (null === a) {
    var f2 = c.type;
    if ("function" === typeof f2 && !aj(f2) && void 0 === f2.defaultProps && null === c.compare && void 0 === c.defaultProps) return b.tag = 15, b.type = f2, bj(a, b, f2, d, e);
    a = Rg(c.type, null, d, b, b.mode, e);
    a.ref = b.ref;
    a.return = b;
    return b.child = a;
  }
  f2 = a.child;
  if (0 === (a.lanes & e)) {
    var g = f2.memoizedProps;
    c = c.compare;
    c = null !== c ? c : Ie;
    if (c(g, d) && a.ref === b.ref) return Zi(a, b, e);
  }
  b.flags |= 1;
  a = Pg(f2, d);
  a.ref = b.ref;
  a.return = b;
  return b.child = a;
}
function bj(a, b, c, d, e) {
  if (null !== a) {
    var f2 = a.memoizedProps;
    if (Ie(f2, d) && a.ref === b.ref) if (dh = false, b.pendingProps = d = f2, 0 !== (a.lanes & e)) 0 !== (a.flags & 131072) && (dh = true);
    else return b.lanes = a.lanes, Zi(a, b, e);
  }
  return cj(a, b, c, d, e);
}
function dj(a, b, c) {
  var d = b.pendingProps, e = d.children, f2 = null !== a ? a.memoizedState : null;
  if ("hidden" === d.mode) if (0 === (b.mode & 1)) b.memoizedState = { baseLanes: 0, cachePool: null, transitions: null }, G(ej, fj), fj |= c;
  else {
    if (0 === (c & 1073741824)) return a = null !== f2 ? f2.baseLanes | c : c, b.lanes = b.childLanes = 1073741824, b.memoizedState = { baseLanes: a, cachePool: null, transitions: null }, b.updateQueue = null, G(ej, fj), fj |= a, null;
    b.memoizedState = { baseLanes: 0, cachePool: null, transitions: null };
    d = null !== f2 ? f2.baseLanes : c;
    G(ej, fj);
    fj |= d;
  }
  else null !== f2 ? (d = f2.baseLanes | c, b.memoizedState = null) : d = c, G(ej, fj), fj |= d;
  Xi(a, b, e, c);
  return b.child;
}
function gj(a, b) {
  var c = b.ref;
  if (null === a && null !== c || null !== a && a.ref !== c) b.flags |= 512, b.flags |= 2097152;
}
function cj(a, b, c, d, e) {
  var f2 = Zf(c) ? Xf : H$1.current;
  f2 = Yf(b, f2);
  ch(b, e);
  c = Nh(a, b, c, d, f2, e);
  d = Sh();
  if (null !== a && !dh) return b.updateQueue = a.updateQueue, b.flags &= -2053, a.lanes &= ~e, Zi(a, b, e);
  I && d && vg(b);
  b.flags |= 1;
  Xi(a, b, c, e);
  return b.child;
}
function hj(a, b, c, d, e) {
  if (Zf(c)) {
    var f2 = true;
    cg(b);
  } else f2 = false;
  ch(b, e);
  if (null === b.stateNode) ij(a, b), Gi(b, c, d), Ii(b, c, d, e), d = true;
  else if (null === a) {
    var g = b.stateNode, h2 = b.memoizedProps;
    g.props = h2;
    var k2 = g.context, l2 = c.contextType;
    "object" === typeof l2 && null !== l2 ? l2 = eh(l2) : (l2 = Zf(c) ? Xf : H$1.current, l2 = Yf(b, l2));
    var m2 = c.getDerivedStateFromProps, q2 = "function" === typeof m2 || "function" === typeof g.getSnapshotBeforeUpdate;
    q2 || "function" !== typeof g.UNSAFE_componentWillReceiveProps && "function" !== typeof g.componentWillReceiveProps || (h2 !== d || k2 !== l2) && Hi(b, g, d, l2);
    jh = false;
    var r2 = b.memoizedState;
    g.state = r2;
    qh(b, d, g, e);
    k2 = b.memoizedState;
    h2 !== d || r2 !== k2 || Wf.current || jh ? ("function" === typeof m2 && (Di(b, c, m2, d), k2 = b.memoizedState), (h2 = jh || Fi(b, c, h2, d, r2, k2, l2)) ? (q2 || "function" !== typeof g.UNSAFE_componentWillMount && "function" !== typeof g.componentWillMount || ("function" === typeof g.componentWillMount && g.componentWillMount(), "function" === typeof g.UNSAFE_componentWillMount && g.UNSAFE_componentWillMount()), "function" === typeof g.componentDidMount && (b.flags |= 4194308)) : ("function" === typeof g.componentDidMount && (b.flags |= 4194308), b.memoizedProps = d, b.memoizedState = k2), g.props = d, g.state = k2, g.context = l2, d = h2) : ("function" === typeof g.componentDidMount && (b.flags |= 4194308), d = false);
  } else {
    g = b.stateNode;
    lh(a, b);
    h2 = b.memoizedProps;
    l2 = b.type === b.elementType ? h2 : Ci(b.type, h2);
    g.props = l2;
    q2 = b.pendingProps;
    r2 = g.context;
    k2 = c.contextType;
    "object" === typeof k2 && null !== k2 ? k2 = eh(k2) : (k2 = Zf(c) ? Xf : H$1.current, k2 = Yf(b, k2));
    var y2 = c.getDerivedStateFromProps;
    (m2 = "function" === typeof y2 || "function" === typeof g.getSnapshotBeforeUpdate) || "function" !== typeof g.UNSAFE_componentWillReceiveProps && "function" !== typeof g.componentWillReceiveProps || (h2 !== q2 || r2 !== k2) && Hi(b, g, d, k2);
    jh = false;
    r2 = b.memoizedState;
    g.state = r2;
    qh(b, d, g, e);
    var n2 = b.memoizedState;
    h2 !== q2 || r2 !== n2 || Wf.current || jh ? ("function" === typeof y2 && (Di(b, c, y2, d), n2 = b.memoizedState), (l2 = jh || Fi(b, c, l2, d, r2, n2, k2) || false) ? (m2 || "function" !== typeof g.UNSAFE_componentWillUpdate && "function" !== typeof g.componentWillUpdate || ("function" === typeof g.componentWillUpdate && g.componentWillUpdate(d, n2, k2), "function" === typeof g.UNSAFE_componentWillUpdate && g.UNSAFE_componentWillUpdate(d, n2, k2)), "function" === typeof g.componentDidUpdate && (b.flags |= 4), "function" === typeof g.getSnapshotBeforeUpdate && (b.flags |= 1024)) : ("function" !== typeof g.componentDidUpdate || h2 === a.memoizedProps && r2 === a.memoizedState || (b.flags |= 4), "function" !== typeof g.getSnapshotBeforeUpdate || h2 === a.memoizedProps && r2 === a.memoizedState || (b.flags |= 1024), b.memoizedProps = d, b.memoizedState = n2), g.props = d, g.state = n2, g.context = k2, d = l2) : ("function" !== typeof g.componentDidUpdate || h2 === a.memoizedProps && r2 === a.memoizedState || (b.flags |= 4), "function" !== typeof g.getSnapshotBeforeUpdate || h2 === a.memoizedProps && r2 === a.memoizedState || (b.flags |= 1024), d = false);
  }
  return jj(a, b, c, d, f2, e);
}
function jj(a, b, c, d, e, f2) {
  gj(a, b);
  var g = 0 !== (b.flags & 128);
  if (!d && !g) return e && dg(b, c, false), Zi(a, b, f2);
  d = b.stateNode;
  Wi.current = b;
  var h2 = g && "function" !== typeof c.getDerivedStateFromError ? null : d.render();
  b.flags |= 1;
  null !== a && g ? (b.child = Ug(b, a.child, null, f2), b.child = Ug(b, null, h2, f2)) : Xi(a, b, h2, f2);
  b.memoizedState = d.state;
  e && dg(b, c, true);
  return b.child;
}
function kj(a) {
  var b = a.stateNode;
  b.pendingContext ? ag(a, b.pendingContext, b.pendingContext !== b.context) : b.context && ag(a, b.context, false);
  yh(a, b.containerInfo);
}
function lj(a, b, c, d, e) {
  Ig();
  Jg(e);
  b.flags |= 256;
  Xi(a, b, c, d);
  return b.child;
}
var mj = { dehydrated: null, treeContext: null, retryLane: 0 };
function nj(a) {
  return { baseLanes: a, cachePool: null, transitions: null };
}
function oj(a, b, c) {
  var d = b.pendingProps, e = L.current, f2 = false, g = 0 !== (b.flags & 128), h2;
  (h2 = g) || (h2 = null !== a && null === a.memoizedState ? false : 0 !== (e & 2));
  if (h2) f2 = true, b.flags &= -129;
  else if (null === a || null !== a.memoizedState) e |= 1;
  G(L, e & 1);
  if (null === a) {
    Eg(b);
    a = b.memoizedState;
    if (null !== a && (a = a.dehydrated, null !== a)) return 0 === (b.mode & 1) ? b.lanes = 1 : "$!" === a.data ? b.lanes = 8 : b.lanes = 1073741824, null;
    g = d.children;
    a = d.fallback;
    return f2 ? (d = b.mode, f2 = b.child, g = { mode: "hidden", children: g }, 0 === (d & 1) && null !== f2 ? (f2.childLanes = 0, f2.pendingProps = g) : f2 = pj(g, d, 0, null), a = Tg(a, d, c, null), f2.return = b, a.return = b, f2.sibling = a, b.child = f2, b.child.memoizedState = nj(c), b.memoizedState = mj, a) : qj(b, g);
  }
  e = a.memoizedState;
  if (null !== e && (h2 = e.dehydrated, null !== h2)) return rj(a, b, g, d, h2, e, c);
  if (f2) {
    f2 = d.fallback;
    g = b.mode;
    e = a.child;
    h2 = e.sibling;
    var k2 = { mode: "hidden", children: d.children };
    0 === (g & 1) && b.child !== e ? (d = b.child, d.childLanes = 0, d.pendingProps = k2, b.deletions = null) : (d = Pg(e, k2), d.subtreeFlags = e.subtreeFlags & 14680064);
    null !== h2 ? f2 = Pg(h2, f2) : (f2 = Tg(f2, g, c, null), f2.flags |= 2);
    f2.return = b;
    d.return = b;
    d.sibling = f2;
    b.child = d;
    d = f2;
    f2 = b.child;
    g = a.child.memoizedState;
    g = null === g ? nj(c) : { baseLanes: g.baseLanes | c, cachePool: null, transitions: g.transitions };
    f2.memoizedState = g;
    f2.childLanes = a.childLanes & ~c;
    b.memoizedState = mj;
    return d;
  }
  f2 = a.child;
  a = f2.sibling;
  d = Pg(f2, { mode: "visible", children: d.children });
  0 === (b.mode & 1) && (d.lanes = c);
  d.return = b;
  d.sibling = null;
  null !== a && (c = b.deletions, null === c ? (b.deletions = [a], b.flags |= 16) : c.push(a));
  b.child = d;
  b.memoizedState = null;
  return d;
}
function qj(a, b) {
  b = pj({ mode: "visible", children: b }, a.mode, 0, null);
  b.return = a;
  return a.child = b;
}
function sj(a, b, c, d) {
  null !== d && Jg(d);
  Ug(b, a.child, null, c);
  a = qj(b, b.pendingProps.children);
  a.flags |= 2;
  b.memoizedState = null;
  return a;
}
function rj(a, b, c, d, e, f2, g) {
  if (c) {
    if (b.flags & 256) return b.flags &= -257, d = Ki(Error(p(422))), sj(a, b, g, d);
    if (null !== b.memoizedState) return b.child = a.child, b.flags |= 128, null;
    f2 = d.fallback;
    e = b.mode;
    d = pj({ mode: "visible", children: d.children }, e, 0, null);
    f2 = Tg(f2, e, g, null);
    f2.flags |= 2;
    d.return = b;
    f2.return = b;
    d.sibling = f2;
    b.child = d;
    0 !== (b.mode & 1) && Ug(b, a.child, null, g);
    b.child.memoizedState = nj(g);
    b.memoizedState = mj;
    return f2;
  }
  if (0 === (b.mode & 1)) return sj(a, b, g, null);
  if ("$!" === e.data) {
    d = e.nextSibling && e.nextSibling.dataset;
    if (d) var h2 = d.dgst;
    d = h2;
    f2 = Error(p(419));
    d = Ki(f2, d, void 0);
    return sj(a, b, g, d);
  }
  h2 = 0 !== (g & a.childLanes);
  if (dh || h2) {
    d = Q;
    if (null !== d) {
      switch (g & -g) {
        case 4:
          e = 2;
          break;
        case 16:
          e = 8;
          break;
        case 64:
        case 128:
        case 256:
        case 512:
        case 1024:
        case 2048:
        case 4096:
        case 8192:
        case 16384:
        case 32768:
        case 65536:
        case 131072:
        case 262144:
        case 524288:
        case 1048576:
        case 2097152:
        case 4194304:
        case 8388608:
        case 16777216:
        case 33554432:
        case 67108864:
          e = 32;
          break;
        case 536870912:
          e = 268435456;
          break;
        default:
          e = 0;
      }
      e = 0 !== (e & (d.suspendedLanes | g)) ? 0 : e;
      0 !== e && e !== f2.retryLane && (f2.retryLane = e, ih(a, e), gi(d, a, e, -1));
    }
    tj();
    d = Ki(Error(p(421)));
    return sj(a, b, g, d);
  }
  if ("$?" === e.data) return b.flags |= 128, b.child = a.child, b = uj.bind(null, a), e._reactRetry = b, null;
  a = f2.treeContext;
  yg = Lf(e.nextSibling);
  xg = b;
  I = true;
  zg = null;
  null !== a && (og[pg++] = rg, og[pg++] = sg, og[pg++] = qg, rg = a.id, sg = a.overflow, qg = b);
  b = qj(b, d.children);
  b.flags |= 4096;
  return b;
}
function vj(a, b, c) {
  a.lanes |= b;
  var d = a.alternate;
  null !== d && (d.lanes |= b);
  bh(a.return, b, c);
}
function wj(a, b, c, d, e) {
  var f2 = a.memoizedState;
  null === f2 ? a.memoizedState = { isBackwards: b, rendering: null, renderingStartTime: 0, last: d, tail: c, tailMode: e } : (f2.isBackwards = b, f2.rendering = null, f2.renderingStartTime = 0, f2.last = d, f2.tail = c, f2.tailMode = e);
}
function xj(a, b, c) {
  var d = b.pendingProps, e = d.revealOrder, f2 = d.tail;
  Xi(a, b, d.children, c);
  d = L.current;
  if (0 !== (d & 2)) d = d & 1 | 2, b.flags |= 128;
  else {
    if (null !== a && 0 !== (a.flags & 128)) a: for (a = b.child; null !== a; ) {
      if (13 === a.tag) null !== a.memoizedState && vj(a, c, b);
      else if (19 === a.tag) vj(a, c, b);
      else if (null !== a.child) {
        a.child.return = a;
        a = a.child;
        continue;
      }
      if (a === b) break a;
      for (; null === a.sibling; ) {
        if (null === a.return || a.return === b) break a;
        a = a.return;
      }
      a.sibling.return = a.return;
      a = a.sibling;
    }
    d &= 1;
  }
  G(L, d);
  if (0 === (b.mode & 1)) b.memoizedState = null;
  else switch (e) {
    case "forwards":
      c = b.child;
      for (e = null; null !== c; ) a = c.alternate, null !== a && null === Ch(a) && (e = c), c = c.sibling;
      c = e;
      null === c ? (e = b.child, b.child = null) : (e = c.sibling, c.sibling = null);
      wj(b, false, e, c, f2);
      break;
    case "backwards":
      c = null;
      e = b.child;
      for (b.child = null; null !== e; ) {
        a = e.alternate;
        if (null !== a && null === Ch(a)) {
          b.child = e;
          break;
        }
        a = e.sibling;
        e.sibling = c;
        c = e;
        e = a;
      }
      wj(b, true, c, null, f2);
      break;
    case "together":
      wj(b, false, null, null, void 0);
      break;
    default:
      b.memoizedState = null;
  }
  return b.child;
}
function ij(a, b) {
  0 === (b.mode & 1) && null !== a && (a.alternate = null, b.alternate = null, b.flags |= 2);
}
function Zi(a, b, c) {
  null !== a && (b.dependencies = a.dependencies);
  rh |= b.lanes;
  if (0 === (c & b.childLanes)) return null;
  if (null !== a && b.child !== a.child) throw Error(p(153));
  if (null !== b.child) {
    a = b.child;
    c = Pg(a, a.pendingProps);
    b.child = c;
    for (c.return = b; null !== a.sibling; ) a = a.sibling, c = c.sibling = Pg(a, a.pendingProps), c.return = b;
    c.sibling = null;
  }
  return b.child;
}
function yj(a, b, c) {
  switch (b.tag) {
    case 3:
      kj(b);
      Ig();
      break;
    case 5:
      Ah(b);
      break;
    case 1:
      Zf(b.type) && cg(b);
      break;
    case 4:
      yh(b, b.stateNode.containerInfo);
      break;
    case 10:
      var d = b.type._context, e = b.memoizedProps.value;
      G(Wg, d._currentValue);
      d._currentValue = e;
      break;
    case 13:
      d = b.memoizedState;
      if (null !== d) {
        if (null !== d.dehydrated) return G(L, L.current & 1), b.flags |= 128, null;
        if (0 !== (c & b.child.childLanes)) return oj(a, b, c);
        G(L, L.current & 1);
        a = Zi(a, b, c);
        return null !== a ? a.sibling : null;
      }
      G(L, L.current & 1);
      break;
    case 19:
      d = 0 !== (c & b.childLanes);
      if (0 !== (a.flags & 128)) {
        if (d) return xj(a, b, c);
        b.flags |= 128;
      }
      e = b.memoizedState;
      null !== e && (e.rendering = null, e.tail = null, e.lastEffect = null);
      G(L, L.current);
      if (d) break;
      else return null;
    case 22:
    case 23:
      return b.lanes = 0, dj(a, b, c);
  }
  return Zi(a, b, c);
}
var zj, Aj, Bj, Cj;
zj = function(a, b) {
  for (var c = b.child; null !== c; ) {
    if (5 === c.tag || 6 === c.tag) a.appendChild(c.stateNode);
    else if (4 !== c.tag && null !== c.child) {
      c.child.return = c;
      c = c.child;
      continue;
    }
    if (c === b) break;
    for (; null === c.sibling; ) {
      if (null === c.return || c.return === b) return;
      c = c.return;
    }
    c.sibling.return = c.return;
    c = c.sibling;
  }
};
Aj = function() {
};
Bj = function(a, b, c, d) {
  var e = a.memoizedProps;
  if (e !== d) {
    a = b.stateNode;
    xh(uh.current);
    var f2 = null;
    switch (c) {
      case "input":
        e = Ya(a, e);
        d = Ya(a, d);
        f2 = [];
        break;
      case "select":
        e = A({}, e, { value: void 0 });
        d = A({}, d, { value: void 0 });
        f2 = [];
        break;
      case "textarea":
        e = gb(a, e);
        d = gb(a, d);
        f2 = [];
        break;
      default:
        "function" !== typeof e.onClick && "function" === typeof d.onClick && (a.onclick = Bf);
    }
    ub(c, d);
    var g;
    c = null;
    for (l2 in e) if (!d.hasOwnProperty(l2) && e.hasOwnProperty(l2) && null != e[l2]) if ("style" === l2) {
      var h2 = e[l2];
      for (g in h2) h2.hasOwnProperty(g) && (c || (c = {}), c[g] = "");
    } else "dangerouslySetInnerHTML" !== l2 && "children" !== l2 && "suppressContentEditableWarning" !== l2 && "suppressHydrationWarning" !== l2 && "autoFocus" !== l2 && (ea.hasOwnProperty(l2) ? f2 || (f2 = []) : (f2 = f2 || []).push(l2, null));
    for (l2 in d) {
      var k2 = d[l2];
      h2 = null != e ? e[l2] : void 0;
      if (d.hasOwnProperty(l2) && k2 !== h2 && (null != k2 || null != h2)) if ("style" === l2) if (h2) {
        for (g in h2) !h2.hasOwnProperty(g) || k2 && k2.hasOwnProperty(g) || (c || (c = {}), c[g] = "");
        for (g in k2) k2.hasOwnProperty(g) && h2[g] !== k2[g] && (c || (c = {}), c[g] = k2[g]);
      } else c || (f2 || (f2 = []), f2.push(
        l2,
        c
      )), c = k2;
      else "dangerouslySetInnerHTML" === l2 ? (k2 = k2 ? k2.__html : void 0, h2 = h2 ? h2.__html : void 0, null != k2 && h2 !== k2 && (f2 = f2 || []).push(l2, k2)) : "children" === l2 ? "string" !== typeof k2 && "number" !== typeof k2 || (f2 = f2 || []).push(l2, "" + k2) : "suppressContentEditableWarning" !== l2 && "suppressHydrationWarning" !== l2 && (ea.hasOwnProperty(l2) ? (null != k2 && "onScroll" === l2 && D$1("scroll", a), f2 || h2 === k2 || (f2 = [])) : (f2 = f2 || []).push(l2, k2));
    }
    c && (f2 = f2 || []).push("style", c);
    var l2 = f2;
    if (b.updateQueue = l2) b.flags |= 4;
  }
};
Cj = function(a, b, c, d) {
  c !== d && (b.flags |= 4);
};
function Dj(a, b) {
  if (!I) switch (a.tailMode) {
    case "hidden":
      b = a.tail;
      for (var c = null; null !== b; ) null !== b.alternate && (c = b), b = b.sibling;
      null === c ? a.tail = null : c.sibling = null;
      break;
    case "collapsed":
      c = a.tail;
      for (var d = null; null !== c; ) null !== c.alternate && (d = c), c = c.sibling;
      null === d ? b || null === a.tail ? a.tail = null : a.tail.sibling = null : d.sibling = null;
  }
}
function S(a) {
  var b = null !== a.alternate && a.alternate.child === a.child, c = 0, d = 0;
  if (b) for (var e = a.child; null !== e; ) c |= e.lanes | e.childLanes, d |= e.subtreeFlags & 14680064, d |= e.flags & 14680064, e.return = a, e = e.sibling;
  else for (e = a.child; null !== e; ) c |= e.lanes | e.childLanes, d |= e.subtreeFlags, d |= e.flags, e.return = a, e = e.sibling;
  a.subtreeFlags |= d;
  a.childLanes = c;
  return b;
}
function Ej(a, b, c) {
  var d = b.pendingProps;
  wg(b);
  switch (b.tag) {
    case 2:
    case 16:
    case 15:
    case 0:
    case 11:
    case 7:
    case 8:
    case 12:
    case 9:
    case 14:
      return S(b), null;
    case 1:
      return Zf(b.type) && $f(), S(b), null;
    case 3:
      d = b.stateNode;
      zh();
      E(Wf);
      E(H$1);
      Eh();
      d.pendingContext && (d.context = d.pendingContext, d.pendingContext = null);
      if (null === a || null === a.child) Gg(b) ? b.flags |= 4 : null === a || a.memoizedState.isDehydrated && 0 === (b.flags & 256) || (b.flags |= 1024, null !== zg && (Fj(zg), zg = null));
      Aj(a, b);
      S(b);
      return null;
    case 5:
      Bh(b);
      var e = xh(wh.current);
      c = b.type;
      if (null !== a && null != b.stateNode) Bj(a, b, c, d, e), a.ref !== b.ref && (b.flags |= 512, b.flags |= 2097152);
      else {
        if (!d) {
          if (null === b.stateNode) throw Error(p(166));
          S(b);
          return null;
        }
        a = xh(uh.current);
        if (Gg(b)) {
          d = b.stateNode;
          c = b.type;
          var f2 = b.memoizedProps;
          d[Of] = b;
          d[Pf] = f2;
          a = 0 !== (b.mode & 1);
          switch (c) {
            case "dialog":
              D$1("cancel", d);
              D$1("close", d);
              break;
            case "iframe":
            case "object":
            case "embed":
              D$1("load", d);
              break;
            case "video":
            case "audio":
              for (e = 0; e < lf.length; e++) D$1(lf[e], d);
              break;
            case "source":
              D$1("error", d);
              break;
            case "img":
            case "image":
            case "link":
              D$1(
                "error",
                d
              );
              D$1("load", d);
              break;
            case "details":
              D$1("toggle", d);
              break;
            case "input":
              Za(d, f2);
              D$1("invalid", d);
              break;
            case "select":
              d._wrapperState = { wasMultiple: !!f2.multiple };
              D$1("invalid", d);
              break;
            case "textarea":
              hb(d, f2), D$1("invalid", d);
          }
          ub(c, f2);
          e = null;
          for (var g in f2) if (f2.hasOwnProperty(g)) {
            var h2 = f2[g];
            "children" === g ? "string" === typeof h2 ? d.textContent !== h2 && (true !== f2.suppressHydrationWarning && Af(d.textContent, h2, a), e = ["children", h2]) : "number" === typeof h2 && d.textContent !== "" + h2 && (true !== f2.suppressHydrationWarning && Af(
              d.textContent,
              h2,
              a
            ), e = ["children", "" + h2]) : ea.hasOwnProperty(g) && null != h2 && "onScroll" === g && D$1("scroll", d);
          }
          switch (c) {
            case "input":
              Va(d);
              db(d, f2, true);
              break;
            case "textarea":
              Va(d);
              jb(d);
              break;
            case "select":
            case "option":
              break;
            default:
              "function" === typeof f2.onClick && (d.onclick = Bf);
          }
          d = e;
          b.updateQueue = d;
          null !== d && (b.flags |= 4);
        } else {
          g = 9 === e.nodeType ? e : e.ownerDocument;
          "http://www.w3.org/1999/xhtml" === a && (a = kb(c));
          "http://www.w3.org/1999/xhtml" === a ? "script" === c ? (a = g.createElement("div"), a.innerHTML = "<script><\/script>", a = a.removeChild(a.firstChild)) : "string" === typeof d.is ? a = g.createElement(c, { is: d.is }) : (a = g.createElement(c), "select" === c && (g = a, d.multiple ? g.multiple = true : d.size && (g.size = d.size))) : a = g.createElementNS(a, c);
          a[Of] = b;
          a[Pf] = d;
          zj(a, b, false, false);
          b.stateNode = a;
          a: {
            g = vb(c, d);
            switch (c) {
              case "dialog":
                D$1("cancel", a);
                D$1("close", a);
                e = d;
                break;
              case "iframe":
              case "object":
              case "embed":
                D$1("load", a);
                e = d;
                break;
              case "video":
              case "audio":
                for (e = 0; e < lf.length; e++) D$1(lf[e], a);
                e = d;
                break;
              case "source":
                D$1("error", a);
                e = d;
                break;
              case "img":
              case "image":
              case "link":
                D$1(
                  "error",
                  a
                );
                D$1("load", a);
                e = d;
                break;
              case "details":
                D$1("toggle", a);
                e = d;
                break;
              case "input":
                Za(a, d);
                e = Ya(a, d);
                D$1("invalid", a);
                break;
              case "option":
                e = d;
                break;
              case "select":
                a._wrapperState = { wasMultiple: !!d.multiple };
                e = A({}, d, { value: void 0 });
                D$1("invalid", a);
                break;
              case "textarea":
                hb(a, d);
                e = gb(a, d);
                D$1("invalid", a);
                break;
              default:
                e = d;
            }
            ub(c, e);
            h2 = e;
            for (f2 in h2) if (h2.hasOwnProperty(f2)) {
              var k2 = h2[f2];
              "style" === f2 ? sb(a, k2) : "dangerouslySetInnerHTML" === f2 ? (k2 = k2 ? k2.__html : void 0, null != k2 && nb(a, k2)) : "children" === f2 ? "string" === typeof k2 ? ("textarea" !== c || "" !== k2) && ob(a, k2) : "number" === typeof k2 && ob(a, "" + k2) : "suppressContentEditableWarning" !== f2 && "suppressHydrationWarning" !== f2 && "autoFocus" !== f2 && (ea.hasOwnProperty(f2) ? null != k2 && "onScroll" === f2 && D$1("scroll", a) : null != k2 && ta(a, f2, k2, g));
            }
            switch (c) {
              case "input":
                Va(a);
                db(a, d, false);
                break;
              case "textarea":
                Va(a);
                jb(a);
                break;
              case "option":
                null != d.value && a.setAttribute("value", "" + Sa(d.value));
                break;
              case "select":
                a.multiple = !!d.multiple;
                f2 = d.value;
                null != f2 ? fb(a, !!d.multiple, f2, false) : null != d.defaultValue && fb(
                  a,
                  !!d.multiple,
                  d.defaultValue,
                  true
                );
                break;
              default:
                "function" === typeof e.onClick && (a.onclick = Bf);
            }
            switch (c) {
              case "button":
              case "input":
              case "select":
              case "textarea":
                d = !!d.autoFocus;
                break a;
              case "img":
                d = true;
                break a;
              default:
                d = false;
            }
          }
          d && (b.flags |= 4);
        }
        null !== b.ref && (b.flags |= 512, b.flags |= 2097152);
      }
      S(b);
      return null;
    case 6:
      if (a && null != b.stateNode) Cj(a, b, a.memoizedProps, d);
      else {
        if ("string" !== typeof d && null === b.stateNode) throw Error(p(166));
        c = xh(wh.current);
        xh(uh.current);
        if (Gg(b)) {
          d = b.stateNode;
          c = b.memoizedProps;
          d[Of] = b;
          if (f2 = d.nodeValue !== c) {
            if (a = xg, null !== a) switch (a.tag) {
              case 3:
                Af(d.nodeValue, c, 0 !== (a.mode & 1));
                break;
              case 5:
                true !== a.memoizedProps.suppressHydrationWarning && Af(d.nodeValue, c, 0 !== (a.mode & 1));
            }
          }
          f2 && (b.flags |= 4);
        } else d = (9 === c.nodeType ? c : c.ownerDocument).createTextNode(d), d[Of] = b, b.stateNode = d;
      }
      S(b);
      return null;
    case 13:
      E(L);
      d = b.memoizedState;
      if (null === a || null !== a.memoizedState && null !== a.memoizedState.dehydrated) {
        if (I && null !== yg && 0 !== (b.mode & 1) && 0 === (b.flags & 128)) Hg(), Ig(), b.flags |= 98560, f2 = false;
        else if (f2 = Gg(b), null !== d && null !== d.dehydrated) {
          if (null === a) {
            if (!f2) throw Error(p(318));
            f2 = b.memoizedState;
            f2 = null !== f2 ? f2.dehydrated : null;
            if (!f2) throw Error(p(317));
            f2[Of] = b;
          } else Ig(), 0 === (b.flags & 128) && (b.memoizedState = null), b.flags |= 4;
          S(b);
          f2 = false;
        } else null !== zg && (Fj(zg), zg = null), f2 = true;
        if (!f2) return b.flags & 65536 ? b : null;
      }
      if (0 !== (b.flags & 128)) return b.lanes = c, b;
      d = null !== d;
      d !== (null !== a && null !== a.memoizedState) && d && (b.child.flags |= 8192, 0 !== (b.mode & 1) && (null === a || 0 !== (L.current & 1) ? 0 === T && (T = 3) : tj()));
      null !== b.updateQueue && (b.flags |= 4);
      S(b);
      return null;
    case 4:
      return zh(), Aj(a, b), null === a && sf(b.stateNode.containerInfo), S(b), null;
    case 10:
      return ah(b.type._context), S(b), null;
    case 17:
      return Zf(b.type) && $f(), S(b), null;
    case 19:
      E(L);
      f2 = b.memoizedState;
      if (null === f2) return S(b), null;
      d = 0 !== (b.flags & 128);
      g = f2.rendering;
      if (null === g) if (d) Dj(f2, false);
      else {
        if (0 !== T || null !== a && 0 !== (a.flags & 128)) for (a = b.child; null !== a; ) {
          g = Ch(a);
          if (null !== g) {
            b.flags |= 128;
            Dj(f2, false);
            d = g.updateQueue;
            null !== d && (b.updateQueue = d, b.flags |= 4);
            b.subtreeFlags = 0;
            d = c;
            for (c = b.child; null !== c; ) f2 = c, a = d, f2.flags &= 14680066, g = f2.alternate, null === g ? (f2.childLanes = 0, f2.lanes = a, f2.child = null, f2.subtreeFlags = 0, f2.memoizedProps = null, f2.memoizedState = null, f2.updateQueue = null, f2.dependencies = null, f2.stateNode = null) : (f2.childLanes = g.childLanes, f2.lanes = g.lanes, f2.child = g.child, f2.subtreeFlags = 0, f2.deletions = null, f2.memoizedProps = g.memoizedProps, f2.memoizedState = g.memoizedState, f2.updateQueue = g.updateQueue, f2.type = g.type, a = g.dependencies, f2.dependencies = null === a ? null : { lanes: a.lanes, firstContext: a.firstContext }), c = c.sibling;
            G(L, L.current & 1 | 2);
            return b.child;
          }
          a = a.sibling;
        }
        null !== f2.tail && B() > Gj && (b.flags |= 128, d = true, Dj(f2, false), b.lanes = 4194304);
      }
      else {
        if (!d) if (a = Ch(g), null !== a) {
          if (b.flags |= 128, d = true, c = a.updateQueue, null !== c && (b.updateQueue = c, b.flags |= 4), Dj(f2, true), null === f2.tail && "hidden" === f2.tailMode && !g.alternate && !I) return S(b), null;
        } else 2 * B() - f2.renderingStartTime > Gj && 1073741824 !== c && (b.flags |= 128, d = true, Dj(f2, false), b.lanes = 4194304);
        f2.isBackwards ? (g.sibling = b.child, b.child = g) : (c = f2.last, null !== c ? c.sibling = g : b.child = g, f2.last = g);
      }
      if (null !== f2.tail) return b = f2.tail, f2.rendering = b, f2.tail = b.sibling, f2.renderingStartTime = B(), b.sibling = null, c = L.current, G(L, d ? c & 1 | 2 : c & 1), b;
      S(b);
      return null;
    case 22:
    case 23:
      return Hj(), d = null !== b.memoizedState, null !== a && null !== a.memoizedState !== d && (b.flags |= 8192), d && 0 !== (b.mode & 1) ? 0 !== (fj & 1073741824) && (S(b), b.subtreeFlags & 6 && (b.flags |= 8192)) : S(b), null;
    case 24:
      return null;
    case 25:
      return null;
  }
  throw Error(p(156, b.tag));
}
function Ij(a, b) {
  wg(b);
  switch (b.tag) {
    case 1:
      return Zf(b.type) && $f(), a = b.flags, a & 65536 ? (b.flags = a & -65537 | 128, b) : null;
    case 3:
      return zh(), E(Wf), E(H$1), Eh(), a = b.flags, 0 !== (a & 65536) && 0 === (a & 128) ? (b.flags = a & -65537 | 128, b) : null;
    case 5:
      return Bh(b), null;
    case 13:
      E(L);
      a = b.memoizedState;
      if (null !== a && null !== a.dehydrated) {
        if (null === b.alternate) throw Error(p(340));
        Ig();
      }
      a = b.flags;
      return a & 65536 ? (b.flags = a & -65537 | 128, b) : null;
    case 19:
      return E(L), null;
    case 4:
      return zh(), null;
    case 10:
      return ah(b.type._context), null;
    case 22:
    case 23:
      return Hj(), null;
    case 24:
      return null;
    default:
      return null;
  }
}
var Jj = false, U = false, Kj = "function" === typeof WeakSet ? WeakSet : Set, V = null;
function Lj(a, b) {
  var c = a.ref;
  if (null !== c) if ("function" === typeof c) try {
    c(null);
  } catch (d) {
    W(a, b, d);
  }
  else c.current = null;
}
function Mj(a, b, c) {
  try {
    c();
  } catch (d) {
    W(a, b, d);
  }
}
var Nj = false;
function Oj(a, b) {
  Cf = dd;
  a = Me$1();
  if (Ne(a)) {
    if ("selectionStart" in a) var c = { start: a.selectionStart, end: a.selectionEnd };
    else a: {
      c = (c = a.ownerDocument) && c.defaultView || window;
      var d = c.getSelection && c.getSelection();
      if (d && 0 !== d.rangeCount) {
        c = d.anchorNode;
        var e = d.anchorOffset, f2 = d.focusNode;
        d = d.focusOffset;
        try {
          c.nodeType, f2.nodeType;
        } catch (F2) {
          c = null;
          break a;
        }
        var g = 0, h2 = -1, k2 = -1, l2 = 0, m2 = 0, q2 = a, r2 = null;
        b: for (; ; ) {
          for (var y2; ; ) {
            q2 !== c || 0 !== e && 3 !== q2.nodeType || (h2 = g + e);
            q2 !== f2 || 0 !== d && 3 !== q2.nodeType || (k2 = g + d);
            3 === q2.nodeType && (g += q2.nodeValue.length);
            if (null === (y2 = q2.firstChild)) break;
            r2 = q2;
            q2 = y2;
          }
          for (; ; ) {
            if (q2 === a) break b;
            r2 === c && ++l2 === e && (h2 = g);
            r2 === f2 && ++m2 === d && (k2 = g);
            if (null !== (y2 = q2.nextSibling)) break;
            q2 = r2;
            r2 = q2.parentNode;
          }
          q2 = y2;
        }
        c = -1 === h2 || -1 === k2 ? null : { start: h2, end: k2 };
      } else c = null;
    }
    c = c || { start: 0, end: 0 };
  } else c = null;
  Df = { focusedElem: a, selectionRange: c };
  dd = false;
  for (V = b; null !== V; ) if (b = V, a = b.child, 0 !== (b.subtreeFlags & 1028) && null !== a) a.return = b, V = a;
  else for (; null !== V; ) {
    b = V;
    try {
      var n2 = b.alternate;
      if (0 !== (b.flags & 1024)) switch (b.tag) {
        case 0:
        case 11:
        case 15:
          break;
        case 1:
          if (null !== n2) {
            var t2 = n2.memoizedProps, J2 = n2.memoizedState, x2 = b.stateNode, w2 = x2.getSnapshotBeforeUpdate(b.elementType === b.type ? t2 : Ci(b.type, t2), J2);
            x2.__reactInternalSnapshotBeforeUpdate = w2;
          }
          break;
        case 3:
          var u2 = b.stateNode.containerInfo;
          1 === u2.nodeType ? u2.textContent = "" : 9 === u2.nodeType && u2.documentElement && u2.removeChild(u2.documentElement);
          break;
        case 5:
        case 6:
        case 4:
        case 17:
          break;
        default:
          throw Error(p(163));
      }
    } catch (F2) {
      W(b, b.return, F2);
    }
    a = b.sibling;
    if (null !== a) {
      a.return = b.return;
      V = a;
      break;
    }
    V = b.return;
  }
  n2 = Nj;
  Nj = false;
  return n2;
}
function Pj(a, b, c) {
  var d = b.updateQueue;
  d = null !== d ? d.lastEffect : null;
  if (null !== d) {
    var e = d = d.next;
    do {
      if ((e.tag & a) === a) {
        var f2 = e.destroy;
        e.destroy = void 0;
        void 0 !== f2 && Mj(b, c, f2);
      }
      e = e.next;
    } while (e !== d);
  }
}
function Qj(a, b) {
  b = b.updateQueue;
  b = null !== b ? b.lastEffect : null;
  if (null !== b) {
    var c = b = b.next;
    do {
      if ((c.tag & a) === a) {
        var d = c.create;
        c.destroy = d();
      }
      c = c.next;
    } while (c !== b);
  }
}
function Rj(a) {
  var b = a.ref;
  if (null !== b) {
    var c = a.stateNode;
    switch (a.tag) {
      case 5:
        a = c;
        break;
      default:
        a = c;
    }
    "function" === typeof b ? b(a) : b.current = a;
  }
}
function Sj(a) {
  var b = a.alternate;
  null !== b && (a.alternate = null, Sj(b));
  a.child = null;
  a.deletions = null;
  a.sibling = null;
  5 === a.tag && (b = a.stateNode, null !== b && (delete b[Of], delete b[Pf], delete b[of], delete b[Qf], delete b[Rf]));
  a.stateNode = null;
  a.return = null;
  a.dependencies = null;
  a.memoizedProps = null;
  a.memoizedState = null;
  a.pendingProps = null;
  a.stateNode = null;
  a.updateQueue = null;
}
function Tj(a) {
  return 5 === a.tag || 3 === a.tag || 4 === a.tag;
}
function Uj(a) {
  a: for (; ; ) {
    for (; null === a.sibling; ) {
      if (null === a.return || Tj(a.return)) return null;
      a = a.return;
    }
    a.sibling.return = a.return;
    for (a = a.sibling; 5 !== a.tag && 6 !== a.tag && 18 !== a.tag; ) {
      if (a.flags & 2) continue a;
      if (null === a.child || 4 === a.tag) continue a;
      else a.child.return = a, a = a.child;
    }
    if (!(a.flags & 2)) return a.stateNode;
  }
}
function Vj(a, b, c) {
  var d = a.tag;
  if (5 === d || 6 === d) a = a.stateNode, b ? 8 === c.nodeType ? c.parentNode.insertBefore(a, b) : c.insertBefore(a, b) : (8 === c.nodeType ? (b = c.parentNode, b.insertBefore(a, c)) : (b = c, b.appendChild(a)), c = c._reactRootContainer, null !== c && void 0 !== c || null !== b.onclick || (b.onclick = Bf));
  else if (4 !== d && (a = a.child, null !== a)) for (Vj(a, b, c), a = a.sibling; null !== a; ) Vj(a, b, c), a = a.sibling;
}
function Wj(a, b, c) {
  var d = a.tag;
  if (5 === d || 6 === d) a = a.stateNode, b ? c.insertBefore(a, b) : c.appendChild(a);
  else if (4 !== d && (a = a.child, null !== a)) for (Wj(a, b, c), a = a.sibling; null !== a; ) Wj(a, b, c), a = a.sibling;
}
var X = null, Xj = false;
function Yj(a, b, c) {
  for (c = c.child; null !== c; ) Zj(a, b, c), c = c.sibling;
}
function Zj(a, b, c) {
  if (lc && "function" === typeof lc.onCommitFiberUnmount) try {
    lc.onCommitFiberUnmount(kc, c);
  } catch (h2) {
  }
  switch (c.tag) {
    case 5:
      U || Lj(c, b);
    case 6:
      var d = X, e = Xj;
      X = null;
      Yj(a, b, c);
      X = d;
      Xj = e;
      null !== X && (Xj ? (a = X, c = c.stateNode, 8 === a.nodeType ? a.parentNode.removeChild(c) : a.removeChild(c)) : X.removeChild(c.stateNode));
      break;
    case 18:
      null !== X && (Xj ? (a = X, c = c.stateNode, 8 === a.nodeType ? Kf(a.parentNode, c) : 1 === a.nodeType && Kf(a, c), bd(a)) : Kf(X, c.stateNode));
      break;
    case 4:
      d = X;
      e = Xj;
      X = c.stateNode.containerInfo;
      Xj = true;
      Yj(a, b, c);
      X = d;
      Xj = e;
      break;
    case 0:
    case 11:
    case 14:
    case 15:
      if (!U && (d = c.updateQueue, null !== d && (d = d.lastEffect, null !== d))) {
        e = d = d.next;
        do {
          var f2 = e, g = f2.destroy;
          f2 = f2.tag;
          void 0 !== g && (0 !== (f2 & 2) ? Mj(c, b, g) : 0 !== (f2 & 4) && Mj(c, b, g));
          e = e.next;
        } while (e !== d);
      }
      Yj(a, b, c);
      break;
    case 1:
      if (!U && (Lj(c, b), d = c.stateNode, "function" === typeof d.componentWillUnmount)) try {
        d.props = c.memoizedProps, d.state = c.memoizedState, d.componentWillUnmount();
      } catch (h2) {
        W(c, b, h2);
      }
      Yj(a, b, c);
      break;
    case 21:
      Yj(a, b, c);
      break;
    case 22:
      c.mode & 1 ? (U = (d = U) || null !== c.memoizedState, Yj(a, b, c), U = d) : Yj(a, b, c);
      break;
    default:
      Yj(a, b, c);
  }
}
function ak(a) {
  var b = a.updateQueue;
  if (null !== b) {
    a.updateQueue = null;
    var c = a.stateNode;
    null === c && (c = a.stateNode = new Kj());
    b.forEach(function(b2) {
      var d = bk.bind(null, a, b2);
      c.has(b2) || (c.add(b2), b2.then(d, d));
    });
  }
}
function ck(a, b) {
  var c = b.deletions;
  if (null !== c) for (var d = 0; d < c.length; d++) {
    var e = c[d];
    try {
      var f2 = a, g = b, h2 = g;
      a: for (; null !== h2; ) {
        switch (h2.tag) {
          case 5:
            X = h2.stateNode;
            Xj = false;
            break a;
          case 3:
            X = h2.stateNode.containerInfo;
            Xj = true;
            break a;
          case 4:
            X = h2.stateNode.containerInfo;
            Xj = true;
            break a;
        }
        h2 = h2.return;
      }
      if (null === X) throw Error(p(160));
      Zj(f2, g, e);
      X = null;
      Xj = false;
      var k2 = e.alternate;
      null !== k2 && (k2.return = null);
      e.return = null;
    } catch (l2) {
      W(e, b, l2);
    }
  }
  if (b.subtreeFlags & 12854) for (b = b.child; null !== b; ) dk(b, a), b = b.sibling;
}
function dk(a, b) {
  var c = a.alternate, d = a.flags;
  switch (a.tag) {
    case 0:
    case 11:
    case 14:
    case 15:
      ck(b, a);
      ek(a);
      if (d & 4) {
        try {
          Pj(3, a, a.return), Qj(3, a);
        } catch (t2) {
          W(a, a.return, t2);
        }
        try {
          Pj(5, a, a.return);
        } catch (t2) {
          W(a, a.return, t2);
        }
      }
      break;
    case 1:
      ck(b, a);
      ek(a);
      d & 512 && null !== c && Lj(c, c.return);
      break;
    case 5:
      ck(b, a);
      ek(a);
      d & 512 && null !== c && Lj(c, c.return);
      if (a.flags & 32) {
        var e = a.stateNode;
        try {
          ob(e, "");
        } catch (t2) {
          W(a, a.return, t2);
        }
      }
      if (d & 4 && (e = a.stateNode, null != e)) {
        var f2 = a.memoizedProps, g = null !== c ? c.memoizedProps : f2, h2 = a.type, k2 = a.updateQueue;
        a.updateQueue = null;
        if (null !== k2) try {
          "input" === h2 && "radio" === f2.type && null != f2.name && ab(e, f2);
          vb(h2, g);
          var l2 = vb(h2, f2);
          for (g = 0; g < k2.length; g += 2) {
            var m2 = k2[g], q2 = k2[g + 1];
            "style" === m2 ? sb(e, q2) : "dangerouslySetInnerHTML" === m2 ? nb(e, q2) : "children" === m2 ? ob(e, q2) : ta(e, m2, q2, l2);
          }
          switch (h2) {
            case "input":
              bb(e, f2);
              break;
            case "textarea":
              ib(e, f2);
              break;
            case "select":
              var r2 = e._wrapperState.wasMultiple;
              e._wrapperState.wasMultiple = !!f2.multiple;
              var y2 = f2.value;
              null != y2 ? fb(e, !!f2.multiple, y2, false) : r2 !== !!f2.multiple && (null != f2.defaultValue ? fb(
                e,
                !!f2.multiple,
                f2.defaultValue,
                true
              ) : fb(e, !!f2.multiple, f2.multiple ? [] : "", false));
          }
          e[Pf] = f2;
        } catch (t2) {
          W(a, a.return, t2);
        }
      }
      break;
    case 6:
      ck(b, a);
      ek(a);
      if (d & 4) {
        if (null === a.stateNode) throw Error(p(162));
        e = a.stateNode;
        f2 = a.memoizedProps;
        try {
          e.nodeValue = f2;
        } catch (t2) {
          W(a, a.return, t2);
        }
      }
      break;
    case 3:
      ck(b, a);
      ek(a);
      if (d & 4 && null !== c && c.memoizedState.isDehydrated) try {
        bd(b.containerInfo);
      } catch (t2) {
        W(a, a.return, t2);
      }
      break;
    case 4:
      ck(b, a);
      ek(a);
      break;
    case 13:
      ck(b, a);
      ek(a);
      e = a.child;
      e.flags & 8192 && (f2 = null !== e.memoizedState, e.stateNode.isHidden = f2, !f2 || null !== e.alternate && null !== e.alternate.memoizedState || (fk = B()));
      d & 4 && ak(a);
      break;
    case 22:
      m2 = null !== c && null !== c.memoizedState;
      a.mode & 1 ? (U = (l2 = U) || m2, ck(b, a), U = l2) : ck(b, a);
      ek(a);
      if (d & 8192) {
        l2 = null !== a.memoizedState;
        if ((a.stateNode.isHidden = l2) && !m2 && 0 !== (a.mode & 1)) for (V = a, m2 = a.child; null !== m2; ) {
          for (q2 = V = m2; null !== V; ) {
            r2 = V;
            y2 = r2.child;
            switch (r2.tag) {
              case 0:
              case 11:
              case 14:
              case 15:
                Pj(4, r2, r2.return);
                break;
              case 1:
                Lj(r2, r2.return);
                var n2 = r2.stateNode;
                if ("function" === typeof n2.componentWillUnmount) {
                  d = r2;
                  c = r2.return;
                  try {
                    b = d, n2.props = b.memoizedProps, n2.state = b.memoizedState, n2.componentWillUnmount();
                  } catch (t2) {
                    W(d, c, t2);
                  }
                }
                break;
              case 5:
                Lj(r2, r2.return);
                break;
              case 22:
                if (null !== r2.memoizedState) {
                  gk(q2);
                  continue;
                }
            }
            null !== y2 ? (y2.return = r2, V = y2) : gk(q2);
          }
          m2 = m2.sibling;
        }
        a: for (m2 = null, q2 = a; ; ) {
          if (5 === q2.tag) {
            if (null === m2) {
              m2 = q2;
              try {
                e = q2.stateNode, l2 ? (f2 = e.style, "function" === typeof f2.setProperty ? f2.setProperty("display", "none", "important") : f2.display = "none") : (h2 = q2.stateNode, k2 = q2.memoizedProps.style, g = void 0 !== k2 && null !== k2 && k2.hasOwnProperty("display") ? k2.display : null, h2.style.display = rb("display", g));
              } catch (t2) {
                W(a, a.return, t2);
              }
            }
          } else if (6 === q2.tag) {
            if (null === m2) try {
              q2.stateNode.nodeValue = l2 ? "" : q2.memoizedProps;
            } catch (t2) {
              W(a, a.return, t2);
            }
          } else if ((22 !== q2.tag && 23 !== q2.tag || null === q2.memoizedState || q2 === a) && null !== q2.child) {
            q2.child.return = q2;
            q2 = q2.child;
            continue;
          }
          if (q2 === a) break a;
          for (; null === q2.sibling; ) {
            if (null === q2.return || q2.return === a) break a;
            m2 === q2 && (m2 = null);
            q2 = q2.return;
          }
          m2 === q2 && (m2 = null);
          q2.sibling.return = q2.return;
          q2 = q2.sibling;
        }
      }
      break;
    case 19:
      ck(b, a);
      ek(a);
      d & 4 && ak(a);
      break;
    case 21:
      break;
    default:
      ck(
        b,
        a
      ), ek(a);
  }
}
function ek(a) {
  var b = a.flags;
  if (b & 2) {
    try {
      a: {
        for (var c = a.return; null !== c; ) {
          if (Tj(c)) {
            var d = c;
            break a;
          }
          c = c.return;
        }
        throw Error(p(160));
      }
      switch (d.tag) {
        case 5:
          var e = d.stateNode;
          d.flags & 32 && (ob(e, ""), d.flags &= -33);
          var f2 = Uj(a);
          Wj(a, f2, e);
          break;
        case 3:
        case 4:
          var g = d.stateNode.containerInfo, h2 = Uj(a);
          Vj(a, h2, g);
          break;
        default:
          throw Error(p(161));
      }
    } catch (k2) {
      W(a, a.return, k2);
    }
    a.flags &= -3;
  }
  b & 4096 && (a.flags &= -4097);
}
function hk(a, b, c) {
  V = a;
  ik(a);
}
function ik(a, b, c) {
  for (var d = 0 !== (a.mode & 1); null !== V; ) {
    var e = V, f2 = e.child;
    if (22 === e.tag && d) {
      var g = null !== e.memoizedState || Jj;
      if (!g) {
        var h2 = e.alternate, k2 = null !== h2 && null !== h2.memoizedState || U;
        h2 = Jj;
        var l2 = U;
        Jj = g;
        if ((U = k2) && !l2) for (V = e; null !== V; ) g = V, k2 = g.child, 22 === g.tag && null !== g.memoizedState ? jk(e) : null !== k2 ? (k2.return = g, V = k2) : jk(e);
        for (; null !== f2; ) V = f2, ik(f2), f2 = f2.sibling;
        V = e;
        Jj = h2;
        U = l2;
      }
      kk(a);
    } else 0 !== (e.subtreeFlags & 8772) && null !== f2 ? (f2.return = e, V = f2) : kk(a);
  }
}
function kk(a) {
  for (; null !== V; ) {
    var b = V;
    if (0 !== (b.flags & 8772)) {
      var c = b.alternate;
      try {
        if (0 !== (b.flags & 8772)) switch (b.tag) {
          case 0:
          case 11:
          case 15:
            U || Qj(5, b);
            break;
          case 1:
            var d = b.stateNode;
            if (b.flags & 4 && !U) if (null === c) d.componentDidMount();
            else {
              var e = b.elementType === b.type ? c.memoizedProps : Ci(b.type, c.memoizedProps);
              d.componentDidUpdate(e, c.memoizedState, d.__reactInternalSnapshotBeforeUpdate);
            }
            var f2 = b.updateQueue;
            null !== f2 && sh(b, f2, d);
            break;
          case 3:
            var g = b.updateQueue;
            if (null !== g) {
              c = null;
              if (null !== b.child) switch (b.child.tag) {
                case 5:
                  c = b.child.stateNode;
                  break;
                case 1:
                  c = b.child.stateNode;
              }
              sh(b, g, c);
            }
            break;
          case 5:
            var h2 = b.stateNode;
            if (null === c && b.flags & 4) {
              c = h2;
              var k2 = b.memoizedProps;
              switch (b.type) {
                case "button":
                case "input":
                case "select":
                case "textarea":
                  k2.autoFocus && c.focus();
                  break;
                case "img":
                  k2.src && (c.src = k2.src);
              }
            }
            break;
          case 6:
            break;
          case 4:
            break;
          case 12:
            break;
          case 13:
            if (null === b.memoizedState) {
              var l2 = b.alternate;
              if (null !== l2) {
                var m2 = l2.memoizedState;
                if (null !== m2) {
                  var q2 = m2.dehydrated;
                  null !== q2 && bd(q2);
                }
              }
            }
            break;
          case 19:
          case 17:
          case 21:
          case 22:
          case 23:
          case 25:
            break;
          default:
            throw Error(p(163));
        }
        U || b.flags & 512 && Rj(b);
      } catch (r2) {
        W(b, b.return, r2);
      }
    }
    if (b === a) {
      V = null;
      break;
    }
    c = b.sibling;
    if (null !== c) {
      c.return = b.return;
      V = c;
      break;
    }
    V = b.return;
  }
}
function gk(a) {
  for (; null !== V; ) {
    var b = V;
    if (b === a) {
      V = null;
      break;
    }
    var c = b.sibling;
    if (null !== c) {
      c.return = b.return;
      V = c;
      break;
    }
    V = b.return;
  }
}
function jk(a) {
  for (; null !== V; ) {
    var b = V;
    try {
      switch (b.tag) {
        case 0:
        case 11:
        case 15:
          var c = b.return;
          try {
            Qj(4, b);
          } catch (k2) {
            W(b, c, k2);
          }
          break;
        case 1:
          var d = b.stateNode;
          if ("function" === typeof d.componentDidMount) {
            var e = b.return;
            try {
              d.componentDidMount();
            } catch (k2) {
              W(b, e, k2);
            }
          }
          var f2 = b.return;
          try {
            Rj(b);
          } catch (k2) {
            W(b, f2, k2);
          }
          break;
        case 5:
          var g = b.return;
          try {
            Rj(b);
          } catch (k2) {
            W(b, g, k2);
          }
      }
    } catch (k2) {
      W(b, b.return, k2);
    }
    if (b === a) {
      V = null;
      break;
    }
    var h2 = b.sibling;
    if (null !== h2) {
      h2.return = b.return;
      V = h2;
      break;
    }
    V = b.return;
  }
}
var lk = Math.ceil, mk = ua.ReactCurrentDispatcher, nk = ua.ReactCurrentOwner, ok = ua.ReactCurrentBatchConfig, K = 0, Q = null, Y$1 = null, Z$1 = 0, fj = 0, ej = Uf(0), T = 0, pk = null, rh = 0, qk = 0, rk = 0, sk = null, tk = null, fk = 0, Gj = Infinity, uk = null, Oi = false, Pi = null, Ri = null, vk = false, wk = null, xk = 0, yk = 0, zk = null, Ak = -1, Bk = 0;
function R() {
  return 0 !== (K & 6) ? B() : -1 !== Ak ? Ak : Ak = B();
}
function yi(a) {
  if (0 === (a.mode & 1)) return 1;
  if (0 !== (K & 2) && 0 !== Z$1) return Z$1 & -Z$1;
  if (null !== Kg.transition) return 0 === Bk && (Bk = yc()), Bk;
  a = C;
  if (0 !== a) return a;
  a = window.event;
  a = void 0 === a ? 16 : jd(a.type);
  return a;
}
function gi(a, b, c, d) {
  if (50 < yk) throw yk = 0, zk = null, Error(p(185));
  Ac(a, c, d);
  if (0 === (K & 2) || a !== Q) a === Q && (0 === (K & 2) && (qk |= c), 4 === T && Ck(a, Z$1)), Dk(a, d), 1 === c && 0 === K && 0 === (b.mode & 1) && (Gj = B() + 500, fg && jg());
}
function Dk(a, b) {
  var c = a.callbackNode;
  wc(a, b);
  var d = uc(a, a === Q ? Z$1 : 0);
  if (0 === d) null !== c && bc(c), a.callbackNode = null, a.callbackPriority = 0;
  else if (b = d & -d, a.callbackPriority !== b) {
    null != c && bc(c);
    if (1 === b) 0 === a.tag ? ig(Ek.bind(null, a)) : hg(Ek.bind(null, a)), Jf(function() {
      0 === (K & 6) && jg();
    }), c = null;
    else {
      switch (Dc(d)) {
        case 1:
          c = fc;
          break;
        case 4:
          c = gc;
          break;
        case 16:
          c = hc;
          break;
        case 536870912:
          c = jc;
          break;
        default:
          c = hc;
      }
      c = Fk(c, Gk.bind(null, a));
    }
    a.callbackPriority = b;
    a.callbackNode = c;
  }
}
function Gk(a, b) {
  Ak = -1;
  Bk = 0;
  if (0 !== (K & 6)) throw Error(p(327));
  var c = a.callbackNode;
  if (Hk() && a.callbackNode !== c) return null;
  var d = uc(a, a === Q ? Z$1 : 0);
  if (0 === d) return null;
  if (0 !== (d & 30) || 0 !== (d & a.expiredLanes) || b) b = Ik(a, d);
  else {
    b = d;
    var e = K;
    K |= 2;
    var f2 = Jk();
    if (Q !== a || Z$1 !== b) uk = null, Gj = B() + 500, Kk(a, b);
    do
      try {
        Lk();
        break;
      } catch (h2) {
        Mk(a, h2);
      }
    while (1);
    $g();
    mk.current = f2;
    K = e;
    null !== Y$1 ? b = 0 : (Q = null, Z$1 = 0, b = T);
  }
  if (0 !== b) {
    2 === b && (e = xc(a), 0 !== e && (d = e, b = Nk(a, e)));
    if (1 === b) throw c = pk, Kk(a, 0), Ck(a, d), Dk(a, B()), c;
    if (6 === b) Ck(a, d);
    else {
      e = a.current.alternate;
      if (0 === (d & 30) && !Ok(e) && (b = Ik(a, d), 2 === b && (f2 = xc(a), 0 !== f2 && (d = f2, b = Nk(a, f2))), 1 === b)) throw c = pk, Kk(a, 0), Ck(a, d), Dk(a, B()), c;
      a.finishedWork = e;
      a.finishedLanes = d;
      switch (b) {
        case 0:
        case 1:
          throw Error(p(345));
        case 2:
          Pk(a, tk, uk);
          break;
        case 3:
          Ck(a, d);
          if ((d & 130023424) === d && (b = fk + 500 - B(), 10 < b)) {
            if (0 !== uc(a, 0)) break;
            e = a.suspendedLanes;
            if ((e & d) !== d) {
              R();
              a.pingedLanes |= a.suspendedLanes & e;
              break;
            }
            a.timeoutHandle = Ff(Pk.bind(null, a, tk, uk), b);
            break;
          }
          Pk(a, tk, uk);
          break;
        case 4:
          Ck(a, d);
          if ((d & 4194240) === d) break;
          b = a.eventTimes;
          for (e = -1; 0 < d; ) {
            var g = 31 - oc(d);
            f2 = 1 << g;
            g = b[g];
            g > e && (e = g);
            d &= ~f2;
          }
          d = e;
          d = B() - d;
          d = (120 > d ? 120 : 480 > d ? 480 : 1080 > d ? 1080 : 1920 > d ? 1920 : 3e3 > d ? 3e3 : 4320 > d ? 4320 : 1960 * lk(d / 1960)) - d;
          if (10 < d) {
            a.timeoutHandle = Ff(Pk.bind(null, a, tk, uk), d);
            break;
          }
          Pk(a, tk, uk);
          break;
        case 5:
          Pk(a, tk, uk);
          break;
        default:
          throw Error(p(329));
      }
    }
  }
  Dk(a, B());
  return a.callbackNode === c ? Gk.bind(null, a) : null;
}
function Nk(a, b) {
  var c = sk;
  a.current.memoizedState.isDehydrated && (Kk(a, b).flags |= 256);
  a = Ik(a, b);
  2 !== a && (b = tk, tk = c, null !== b && Fj(b));
  return a;
}
function Fj(a) {
  null === tk ? tk = a : tk.push.apply(tk, a);
}
function Ok(a) {
  for (var b = a; ; ) {
    if (b.flags & 16384) {
      var c = b.updateQueue;
      if (null !== c && (c = c.stores, null !== c)) for (var d = 0; d < c.length; d++) {
        var e = c[d], f2 = e.getSnapshot;
        e = e.value;
        try {
          if (!He$1(f2(), e)) return false;
        } catch (g) {
          return false;
        }
      }
    }
    c = b.child;
    if (b.subtreeFlags & 16384 && null !== c) c.return = b, b = c;
    else {
      if (b === a) break;
      for (; null === b.sibling; ) {
        if (null === b.return || b.return === a) return true;
        b = b.return;
      }
      b.sibling.return = b.return;
      b = b.sibling;
    }
  }
  return true;
}
function Ck(a, b) {
  b &= ~rk;
  b &= ~qk;
  a.suspendedLanes |= b;
  a.pingedLanes &= ~b;
  for (a = a.expirationTimes; 0 < b; ) {
    var c = 31 - oc(b), d = 1 << c;
    a[c] = -1;
    b &= ~d;
  }
}
function Ek(a) {
  if (0 !== (K & 6)) throw Error(p(327));
  Hk();
  var b = uc(a, 0);
  if (0 === (b & 1)) return Dk(a, B()), null;
  var c = Ik(a, b);
  if (0 !== a.tag && 2 === c) {
    var d = xc(a);
    0 !== d && (b = d, c = Nk(a, d));
  }
  if (1 === c) throw c = pk, Kk(a, 0), Ck(a, b), Dk(a, B()), c;
  if (6 === c) throw Error(p(345));
  a.finishedWork = a.current.alternate;
  a.finishedLanes = b;
  Pk(a, tk, uk);
  Dk(a, B());
  return null;
}
function Qk(a, b) {
  var c = K;
  K |= 1;
  try {
    return a(b);
  } finally {
    K = c, 0 === K && (Gj = B() + 500, fg && jg());
  }
}
function Rk(a) {
  null !== wk && 0 === wk.tag && 0 === (K & 6) && Hk();
  var b = K;
  K |= 1;
  var c = ok.transition, d = C;
  try {
    if (ok.transition = null, C = 1, a) return a();
  } finally {
    C = d, ok.transition = c, K = b, 0 === (K & 6) && jg();
  }
}
function Hj() {
  fj = ej.current;
  E(ej);
}
function Kk(a, b) {
  a.finishedWork = null;
  a.finishedLanes = 0;
  var c = a.timeoutHandle;
  -1 !== c && (a.timeoutHandle = -1, Gf(c));
  if (null !== Y$1) for (c = Y$1.return; null !== c; ) {
    var d = c;
    wg(d);
    switch (d.tag) {
      case 1:
        d = d.type.childContextTypes;
        null !== d && void 0 !== d && $f();
        break;
      case 3:
        zh();
        E(Wf);
        E(H$1);
        Eh();
        break;
      case 5:
        Bh(d);
        break;
      case 4:
        zh();
        break;
      case 13:
        E(L);
        break;
      case 19:
        E(L);
        break;
      case 10:
        ah(d.type._context);
        break;
      case 22:
      case 23:
        Hj();
    }
    c = c.return;
  }
  Q = a;
  Y$1 = a = Pg(a.current, null);
  Z$1 = fj = b;
  T = 0;
  pk = null;
  rk = qk = rh = 0;
  tk = sk = null;
  if (null !== fh) {
    for (b = 0; b < fh.length; b++) if (c = fh[b], d = c.interleaved, null !== d) {
      c.interleaved = null;
      var e = d.next, f2 = c.pending;
      if (null !== f2) {
        var g = f2.next;
        f2.next = e;
        d.next = g;
      }
      c.pending = d;
    }
    fh = null;
  }
  return a;
}
function Mk(a, b) {
  do {
    var c = Y$1;
    try {
      $g();
      Fh.current = Rh;
      if (Ih) {
        for (var d = M.memoizedState; null !== d; ) {
          var e = d.queue;
          null !== e && (e.pending = null);
          d = d.next;
        }
        Ih = false;
      }
      Hh = 0;
      O = N = M = null;
      Jh = false;
      Kh = 0;
      nk.current = null;
      if (null === c || null === c.return) {
        T = 1;
        pk = b;
        Y$1 = null;
        break;
      }
      a: {
        var f2 = a, g = c.return, h2 = c, k2 = b;
        b = Z$1;
        h2.flags |= 32768;
        if (null !== k2 && "object" === typeof k2 && "function" === typeof k2.then) {
          var l2 = k2, m2 = h2, q2 = m2.tag;
          if (0 === (m2.mode & 1) && (0 === q2 || 11 === q2 || 15 === q2)) {
            var r2 = m2.alternate;
            r2 ? (m2.updateQueue = r2.updateQueue, m2.memoizedState = r2.memoizedState, m2.lanes = r2.lanes) : (m2.updateQueue = null, m2.memoizedState = null);
          }
          var y2 = Ui(g);
          if (null !== y2) {
            y2.flags &= -257;
            Vi(y2, g, h2, f2, b);
            y2.mode & 1 && Si(f2, l2, b);
            b = y2;
            k2 = l2;
            var n2 = b.updateQueue;
            if (null === n2) {
              var t2 = /* @__PURE__ */ new Set();
              t2.add(k2);
              b.updateQueue = t2;
            } else n2.add(k2);
            break a;
          } else {
            if (0 === (b & 1)) {
              Si(f2, l2, b);
              tj();
              break a;
            }
            k2 = Error(p(426));
          }
        } else if (I && h2.mode & 1) {
          var J2 = Ui(g);
          if (null !== J2) {
            0 === (J2.flags & 65536) && (J2.flags |= 256);
            Vi(J2, g, h2, f2, b);
            Jg(Ji(k2, h2));
            break a;
          }
        }
        f2 = k2 = Ji(k2, h2);
        4 !== T && (T = 2);
        null === sk ? sk = [f2] : sk.push(f2);
        f2 = g;
        do {
          switch (f2.tag) {
            case 3:
              f2.flags |= 65536;
              b &= -b;
              f2.lanes |= b;
              var x2 = Ni(f2, k2, b);
              ph(f2, x2);
              break a;
            case 1:
              h2 = k2;
              var w2 = f2.type, u2 = f2.stateNode;
              if (0 === (f2.flags & 128) && ("function" === typeof w2.getDerivedStateFromError || null !== u2 && "function" === typeof u2.componentDidCatch && (null === Ri || !Ri.has(u2)))) {
                f2.flags |= 65536;
                b &= -b;
                f2.lanes |= b;
                var F2 = Qi(f2, h2, b);
                ph(f2, F2);
                break a;
              }
          }
          f2 = f2.return;
        } while (null !== f2);
      }
      Sk(c);
    } catch (na) {
      b = na;
      Y$1 === c && null !== c && (Y$1 = c = c.return);
      continue;
    }
    break;
  } while (1);
}
function Jk() {
  var a = mk.current;
  mk.current = Rh;
  return null === a ? Rh : a;
}
function tj() {
  if (0 === T || 3 === T || 2 === T) T = 4;
  null === Q || 0 === (rh & 268435455) && 0 === (qk & 268435455) || Ck(Q, Z$1);
}
function Ik(a, b) {
  var c = K;
  K |= 2;
  var d = Jk();
  if (Q !== a || Z$1 !== b) uk = null, Kk(a, b);
  do
    try {
      Tk();
      break;
    } catch (e) {
      Mk(a, e);
    }
  while (1);
  $g();
  K = c;
  mk.current = d;
  if (null !== Y$1) throw Error(p(261));
  Q = null;
  Z$1 = 0;
  return T;
}
function Tk() {
  for (; null !== Y$1; ) Uk(Y$1);
}
function Lk() {
  for (; null !== Y$1 && !cc(); ) Uk(Y$1);
}
function Uk(a) {
  var b = Vk(a.alternate, a, fj);
  a.memoizedProps = a.pendingProps;
  null === b ? Sk(a) : Y$1 = b;
  nk.current = null;
}
function Sk(a) {
  var b = a;
  do {
    var c = b.alternate;
    a = b.return;
    if (0 === (b.flags & 32768)) {
      if (c = Ej(c, b, fj), null !== c) {
        Y$1 = c;
        return;
      }
    } else {
      c = Ij(c, b);
      if (null !== c) {
        c.flags &= 32767;
        Y$1 = c;
        return;
      }
      if (null !== a) a.flags |= 32768, a.subtreeFlags = 0, a.deletions = null;
      else {
        T = 6;
        Y$1 = null;
        return;
      }
    }
    b = b.sibling;
    if (null !== b) {
      Y$1 = b;
      return;
    }
    Y$1 = b = a;
  } while (null !== b);
  0 === T && (T = 5);
}
function Pk(a, b, c) {
  var d = C, e = ok.transition;
  try {
    ok.transition = null, C = 1, Wk(a, b, c, d);
  } finally {
    ok.transition = e, C = d;
  }
  return null;
}
function Wk(a, b, c, d) {
  do
    Hk();
  while (null !== wk);
  if (0 !== (K & 6)) throw Error(p(327));
  c = a.finishedWork;
  var e = a.finishedLanes;
  if (null === c) return null;
  a.finishedWork = null;
  a.finishedLanes = 0;
  if (c === a.current) throw Error(p(177));
  a.callbackNode = null;
  a.callbackPriority = 0;
  var f2 = c.lanes | c.childLanes;
  Bc(a, f2);
  a === Q && (Y$1 = Q = null, Z$1 = 0);
  0 === (c.subtreeFlags & 2064) && 0 === (c.flags & 2064) || vk || (vk = true, Fk(hc, function() {
    Hk();
    return null;
  }));
  f2 = 0 !== (c.flags & 15990);
  if (0 !== (c.subtreeFlags & 15990) || f2) {
    f2 = ok.transition;
    ok.transition = null;
    var g = C;
    C = 1;
    var h2 = K;
    K |= 4;
    nk.current = null;
    Oj(a, c);
    dk(c, a);
    Oe$1(Df);
    dd = !!Cf;
    Df = Cf = null;
    a.current = c;
    hk(c);
    dc();
    K = h2;
    C = g;
    ok.transition = f2;
  } else a.current = c;
  vk && (vk = false, wk = a, xk = e);
  f2 = a.pendingLanes;
  0 === f2 && (Ri = null);
  mc(c.stateNode);
  Dk(a, B());
  if (null !== b) for (d = a.onRecoverableError, c = 0; c < b.length; c++) e = b[c], d(e.value, { componentStack: e.stack, digest: e.digest });
  if (Oi) throw Oi = false, a = Pi, Pi = null, a;
  0 !== (xk & 1) && 0 !== a.tag && Hk();
  f2 = a.pendingLanes;
  0 !== (f2 & 1) ? a === zk ? yk++ : (yk = 0, zk = a) : yk = 0;
  jg();
  return null;
}
function Hk() {
  if (null !== wk) {
    var a = Dc(xk), b = ok.transition, c = C;
    try {
      ok.transition = null;
      C = 16 > a ? 16 : a;
      if (null === wk) var d = false;
      else {
        a = wk;
        wk = null;
        xk = 0;
        if (0 !== (K & 6)) throw Error(p(331));
        var e = K;
        K |= 4;
        for (V = a.current; null !== V; ) {
          var f2 = V, g = f2.child;
          if (0 !== (V.flags & 16)) {
            var h2 = f2.deletions;
            if (null !== h2) {
              for (var k2 = 0; k2 < h2.length; k2++) {
                var l2 = h2[k2];
                for (V = l2; null !== V; ) {
                  var m2 = V;
                  switch (m2.tag) {
                    case 0:
                    case 11:
                    case 15:
                      Pj(8, m2, f2);
                  }
                  var q2 = m2.child;
                  if (null !== q2) q2.return = m2, V = q2;
                  else for (; null !== V; ) {
                    m2 = V;
                    var r2 = m2.sibling, y2 = m2.return;
                    Sj(m2);
                    if (m2 === l2) {
                      V = null;
                      break;
                    }
                    if (null !== r2) {
                      r2.return = y2;
                      V = r2;
                      break;
                    }
                    V = y2;
                  }
                }
              }
              var n2 = f2.alternate;
              if (null !== n2) {
                var t2 = n2.child;
                if (null !== t2) {
                  n2.child = null;
                  do {
                    var J2 = t2.sibling;
                    t2.sibling = null;
                    t2 = J2;
                  } while (null !== t2);
                }
              }
              V = f2;
            }
          }
          if (0 !== (f2.subtreeFlags & 2064) && null !== g) g.return = f2, V = g;
          else b: for (; null !== V; ) {
            f2 = V;
            if (0 !== (f2.flags & 2048)) switch (f2.tag) {
              case 0:
              case 11:
              case 15:
                Pj(9, f2, f2.return);
            }
            var x2 = f2.sibling;
            if (null !== x2) {
              x2.return = f2.return;
              V = x2;
              break b;
            }
            V = f2.return;
          }
        }
        var w2 = a.current;
        for (V = w2; null !== V; ) {
          g = V;
          var u2 = g.child;
          if (0 !== (g.subtreeFlags & 2064) && null !== u2) u2.return = g, V = u2;
          else b: for (g = w2; null !== V; ) {
            h2 = V;
            if (0 !== (h2.flags & 2048)) try {
              switch (h2.tag) {
                case 0:
                case 11:
                case 15:
                  Qj(9, h2);
              }
            } catch (na) {
              W(h2, h2.return, na);
            }
            if (h2 === g) {
              V = null;
              break b;
            }
            var F2 = h2.sibling;
            if (null !== F2) {
              F2.return = h2.return;
              V = F2;
              break b;
            }
            V = h2.return;
          }
        }
        K = e;
        jg();
        if (lc && "function" === typeof lc.onPostCommitFiberRoot) try {
          lc.onPostCommitFiberRoot(kc, a);
        } catch (na) {
        }
        d = true;
      }
      return d;
    } finally {
      C = c, ok.transition = b;
    }
  }
  return false;
}
function Xk(a, b, c) {
  b = Ji(c, b);
  b = Ni(a, b, 1);
  a = nh(a, b, 1);
  b = R();
  null !== a && (Ac(a, 1, b), Dk(a, b));
}
function W(a, b, c) {
  if (3 === a.tag) Xk(a, a, c);
  else for (; null !== b; ) {
    if (3 === b.tag) {
      Xk(b, a, c);
      break;
    } else if (1 === b.tag) {
      var d = b.stateNode;
      if ("function" === typeof b.type.getDerivedStateFromError || "function" === typeof d.componentDidCatch && (null === Ri || !Ri.has(d))) {
        a = Ji(c, a);
        a = Qi(b, a, 1);
        b = nh(b, a, 1);
        a = R();
        null !== b && (Ac(b, 1, a), Dk(b, a));
        break;
      }
    }
    b = b.return;
  }
}
function Ti(a, b, c) {
  var d = a.pingCache;
  null !== d && d.delete(b);
  b = R();
  a.pingedLanes |= a.suspendedLanes & c;
  Q === a && (Z$1 & c) === c && (4 === T || 3 === T && (Z$1 & 130023424) === Z$1 && 500 > B() - fk ? Kk(a, 0) : rk |= c);
  Dk(a, b);
}
function Yk(a, b) {
  0 === b && (0 === (a.mode & 1) ? b = 1 : (b = sc, sc <<= 1, 0 === (sc & 130023424) && (sc = 4194304)));
  var c = R();
  a = ih(a, b);
  null !== a && (Ac(a, b, c), Dk(a, c));
}
function uj(a) {
  var b = a.memoizedState, c = 0;
  null !== b && (c = b.retryLane);
  Yk(a, c);
}
function bk(a, b) {
  var c = 0;
  switch (a.tag) {
    case 13:
      var d = a.stateNode;
      var e = a.memoizedState;
      null !== e && (c = e.retryLane);
      break;
    case 19:
      d = a.stateNode;
      break;
    default:
      throw Error(p(314));
  }
  null !== d && d.delete(b);
  Yk(a, c);
}
var Vk;
Vk = function(a, b, c) {
  if (null !== a) if (a.memoizedProps !== b.pendingProps || Wf.current) dh = true;
  else {
    if (0 === (a.lanes & c) && 0 === (b.flags & 128)) return dh = false, yj(a, b, c);
    dh = 0 !== (a.flags & 131072) ? true : false;
  }
  else dh = false, I && 0 !== (b.flags & 1048576) && ug(b, ng, b.index);
  b.lanes = 0;
  switch (b.tag) {
    case 2:
      var d = b.type;
      ij(a, b);
      a = b.pendingProps;
      var e = Yf(b, H$1.current);
      ch(b, c);
      e = Nh(null, b, d, a, e, c);
      var f2 = Sh();
      b.flags |= 1;
      "object" === typeof e && null !== e && "function" === typeof e.render && void 0 === e.$$typeof ? (b.tag = 1, b.memoizedState = null, b.updateQueue = null, Zf(d) ? (f2 = true, cg(b)) : f2 = false, b.memoizedState = null !== e.state && void 0 !== e.state ? e.state : null, kh(b), e.updater = Ei, b.stateNode = e, e._reactInternals = b, Ii(b, d, a, c), b = jj(null, b, d, true, f2, c)) : (b.tag = 0, I && f2 && vg(b), Xi(null, b, e, c), b = b.child);
      return b;
    case 16:
      d = b.elementType;
      a: {
        ij(a, b);
        a = b.pendingProps;
        e = d._init;
        d = e(d._payload);
        b.type = d;
        e = b.tag = Zk(d);
        a = Ci(d, a);
        switch (e) {
          case 0:
            b = cj(null, b, d, a, c);
            break a;
          case 1:
            b = hj(null, b, d, a, c);
            break a;
          case 11:
            b = Yi(null, b, d, a, c);
            break a;
          case 14:
            b = $i(null, b, d, Ci(d.type, a), c);
            break a;
        }
        throw Error(p(
          306,
          d,
          ""
        ));
      }
      return b;
    case 0:
      return d = b.type, e = b.pendingProps, e = b.elementType === d ? e : Ci(d, e), cj(a, b, d, e, c);
    case 1:
      return d = b.type, e = b.pendingProps, e = b.elementType === d ? e : Ci(d, e), hj(a, b, d, e, c);
    case 3:
      a: {
        kj(b);
        if (null === a) throw Error(p(387));
        d = b.pendingProps;
        f2 = b.memoizedState;
        e = f2.element;
        lh(a, b);
        qh(b, d, null, c);
        var g = b.memoizedState;
        d = g.element;
        if (f2.isDehydrated) if (f2 = { element: d, isDehydrated: false, cache: g.cache, pendingSuspenseBoundaries: g.pendingSuspenseBoundaries, transitions: g.transitions }, b.updateQueue.baseState = f2, b.memoizedState = f2, b.flags & 256) {
          e = Ji(Error(p(423)), b);
          b = lj(a, b, d, c, e);
          break a;
        } else if (d !== e) {
          e = Ji(Error(p(424)), b);
          b = lj(a, b, d, c, e);
          break a;
        } else for (yg = Lf(b.stateNode.containerInfo.firstChild), xg = b, I = true, zg = null, c = Vg(b, null, d, c), b.child = c; c; ) c.flags = c.flags & -3 | 4096, c = c.sibling;
        else {
          Ig();
          if (d === e) {
            b = Zi(a, b, c);
            break a;
          }
          Xi(a, b, d, c);
        }
        b = b.child;
      }
      return b;
    case 5:
      return Ah(b), null === a && Eg(b), d = b.type, e = b.pendingProps, f2 = null !== a ? a.memoizedProps : null, g = e.children, Ef(d, e) ? g = null : null !== f2 && Ef(d, f2) && (b.flags |= 32), gj(a, b), Xi(a, b, g, c), b.child;
    case 6:
      return null === a && Eg(b), null;
    case 13:
      return oj(a, b, c);
    case 4:
      return yh(b, b.stateNode.containerInfo), d = b.pendingProps, null === a ? b.child = Ug(b, null, d, c) : Xi(a, b, d, c), b.child;
    case 11:
      return d = b.type, e = b.pendingProps, e = b.elementType === d ? e : Ci(d, e), Yi(a, b, d, e, c);
    case 7:
      return Xi(a, b, b.pendingProps, c), b.child;
    case 8:
      return Xi(a, b, b.pendingProps.children, c), b.child;
    case 12:
      return Xi(a, b, b.pendingProps.children, c), b.child;
    case 10:
      a: {
        d = b.type._context;
        e = b.pendingProps;
        f2 = b.memoizedProps;
        g = e.value;
        G(Wg, d._currentValue);
        d._currentValue = g;
        if (null !== f2) if (He$1(f2.value, g)) {
          if (f2.children === e.children && !Wf.current) {
            b = Zi(a, b, c);
            break a;
          }
        } else for (f2 = b.child, null !== f2 && (f2.return = b); null !== f2; ) {
          var h2 = f2.dependencies;
          if (null !== h2) {
            g = f2.child;
            for (var k2 = h2.firstContext; null !== k2; ) {
              if (k2.context === d) {
                if (1 === f2.tag) {
                  k2 = mh(-1, c & -c);
                  k2.tag = 2;
                  var l2 = f2.updateQueue;
                  if (null !== l2) {
                    l2 = l2.shared;
                    var m2 = l2.pending;
                    null === m2 ? k2.next = k2 : (k2.next = m2.next, m2.next = k2);
                    l2.pending = k2;
                  }
                }
                f2.lanes |= c;
                k2 = f2.alternate;
                null !== k2 && (k2.lanes |= c);
                bh(
                  f2.return,
                  c,
                  b
                );
                h2.lanes |= c;
                break;
              }
              k2 = k2.next;
            }
          } else if (10 === f2.tag) g = f2.type === b.type ? null : f2.child;
          else if (18 === f2.tag) {
            g = f2.return;
            if (null === g) throw Error(p(341));
            g.lanes |= c;
            h2 = g.alternate;
            null !== h2 && (h2.lanes |= c);
            bh(g, c, b);
            g = f2.sibling;
          } else g = f2.child;
          if (null !== g) g.return = f2;
          else for (g = f2; null !== g; ) {
            if (g === b) {
              g = null;
              break;
            }
            f2 = g.sibling;
            if (null !== f2) {
              f2.return = g.return;
              g = f2;
              break;
            }
            g = g.return;
          }
          f2 = g;
        }
        Xi(a, b, e.children, c);
        b = b.child;
      }
      return b;
    case 9:
      return e = b.type, d = b.pendingProps.children, ch(b, c), e = eh(e), d = d(e), b.flags |= 1, Xi(a, b, d, c), b.child;
    case 14:
      return d = b.type, e = Ci(d, b.pendingProps), e = Ci(d.type, e), $i(a, b, d, e, c);
    case 15:
      return bj(a, b, b.type, b.pendingProps, c);
    case 17:
      return d = b.type, e = b.pendingProps, e = b.elementType === d ? e : Ci(d, e), ij(a, b), b.tag = 1, Zf(d) ? (a = true, cg(b)) : a = false, ch(b, c), Gi(b, d, e), Ii(b, d, e, c), jj(null, b, d, true, a, c);
    case 19:
      return xj(a, b, c);
    case 22:
      return dj(a, b, c);
  }
  throw Error(p(156, b.tag));
};
function Fk(a, b) {
  return ac(a, b);
}
function $k(a, b, c, d) {
  this.tag = a;
  this.key = c;
  this.sibling = this.child = this.return = this.stateNode = this.type = this.elementType = null;
  this.index = 0;
  this.ref = null;
  this.pendingProps = b;
  this.dependencies = this.memoizedState = this.updateQueue = this.memoizedProps = null;
  this.mode = d;
  this.subtreeFlags = this.flags = 0;
  this.deletions = null;
  this.childLanes = this.lanes = 0;
  this.alternate = null;
}
function Bg(a, b, c, d) {
  return new $k(a, b, c, d);
}
function aj(a) {
  a = a.prototype;
  return !(!a || !a.isReactComponent);
}
function Zk(a) {
  if ("function" === typeof a) return aj(a) ? 1 : 0;
  if (void 0 !== a && null !== a) {
    a = a.$$typeof;
    if (a === Da) return 11;
    if (a === Ga) return 14;
  }
  return 2;
}
function Pg(a, b) {
  var c = a.alternate;
  null === c ? (c = Bg(a.tag, b, a.key, a.mode), c.elementType = a.elementType, c.type = a.type, c.stateNode = a.stateNode, c.alternate = a, a.alternate = c) : (c.pendingProps = b, c.type = a.type, c.flags = 0, c.subtreeFlags = 0, c.deletions = null);
  c.flags = a.flags & 14680064;
  c.childLanes = a.childLanes;
  c.lanes = a.lanes;
  c.child = a.child;
  c.memoizedProps = a.memoizedProps;
  c.memoizedState = a.memoizedState;
  c.updateQueue = a.updateQueue;
  b = a.dependencies;
  c.dependencies = null === b ? null : { lanes: b.lanes, firstContext: b.firstContext };
  c.sibling = a.sibling;
  c.index = a.index;
  c.ref = a.ref;
  return c;
}
function Rg(a, b, c, d, e, f2) {
  var g = 2;
  d = a;
  if ("function" === typeof a) aj(a) && (g = 1);
  else if ("string" === typeof a) g = 5;
  else a: switch (a) {
    case ya:
      return Tg(c.children, e, f2, b);
    case za:
      g = 8;
      e |= 8;
      break;
    case Aa:
      return a = Bg(12, c, b, e | 2), a.elementType = Aa, a.lanes = f2, a;
    case Ea:
      return a = Bg(13, c, b, e), a.elementType = Ea, a.lanes = f2, a;
    case Fa:
      return a = Bg(19, c, b, e), a.elementType = Fa, a.lanes = f2, a;
    case Ia:
      return pj(c, e, f2, b);
    default:
      if ("object" === typeof a && null !== a) switch (a.$$typeof) {
        case Ba:
          g = 10;
          break a;
        case Ca:
          g = 9;
          break a;
        case Da:
          g = 11;
          break a;
        case Ga:
          g = 14;
          break a;
        case Ha:
          g = 16;
          d = null;
          break a;
      }
      throw Error(p(130, null == a ? a : typeof a, ""));
  }
  b = Bg(g, c, b, e);
  b.elementType = a;
  b.type = d;
  b.lanes = f2;
  return b;
}
function Tg(a, b, c, d) {
  a = Bg(7, a, d, b);
  a.lanes = c;
  return a;
}
function pj(a, b, c, d) {
  a = Bg(22, a, d, b);
  a.elementType = Ia;
  a.lanes = c;
  a.stateNode = { isHidden: false };
  return a;
}
function Qg(a, b, c) {
  a = Bg(6, a, null, b);
  a.lanes = c;
  return a;
}
function Sg(a, b, c) {
  b = Bg(4, null !== a.children ? a.children : [], a.key, b);
  b.lanes = c;
  b.stateNode = { containerInfo: a.containerInfo, pendingChildren: null, implementation: a.implementation };
  return b;
}
function al(a, b, c, d, e) {
  this.tag = b;
  this.containerInfo = a;
  this.finishedWork = this.pingCache = this.current = this.pendingChildren = null;
  this.timeoutHandle = -1;
  this.callbackNode = this.pendingContext = this.context = null;
  this.callbackPriority = 0;
  this.eventTimes = zc(0);
  this.expirationTimes = zc(-1);
  this.entangledLanes = this.finishedLanes = this.mutableReadLanes = this.expiredLanes = this.pingedLanes = this.suspendedLanes = this.pendingLanes = 0;
  this.entanglements = zc(0);
  this.identifierPrefix = d;
  this.onRecoverableError = e;
  this.mutableSourceEagerHydrationData = null;
}
function bl(a, b, c, d, e, f2, g, h2, k2) {
  a = new al(a, b, c, h2, k2);
  1 === b ? (b = 1, true === f2 && (b |= 8)) : b = 0;
  f2 = Bg(3, null, null, b);
  a.current = f2;
  f2.stateNode = a;
  f2.memoizedState = { element: d, isDehydrated: c, cache: null, transitions: null, pendingSuspenseBoundaries: null };
  kh(f2);
  return a;
}
function cl(a, b, c) {
  var d = 3 < arguments.length && void 0 !== arguments[3] ? arguments[3] : null;
  return { $$typeof: wa, key: null == d ? null : "" + d, children: a, containerInfo: b, implementation: c };
}
function dl(a) {
  if (!a) return Vf;
  a = a._reactInternals;
  a: {
    if (Vb(a) !== a || 1 !== a.tag) throw Error(p(170));
    var b = a;
    do {
      switch (b.tag) {
        case 3:
          b = b.stateNode.context;
          break a;
        case 1:
          if (Zf(b.type)) {
            b = b.stateNode.__reactInternalMemoizedMergedChildContext;
            break a;
          }
      }
      b = b.return;
    } while (null !== b);
    throw Error(p(171));
  }
  if (1 === a.tag) {
    var c = a.type;
    if (Zf(c)) return bg(a, c, b);
  }
  return b;
}
function el(a, b, c, d, e, f2, g, h2, k2) {
  a = bl(c, d, true, a, e, f2, g, h2, k2);
  a.context = dl(null);
  c = a.current;
  d = R();
  e = yi(c);
  f2 = mh(d, e);
  f2.callback = void 0 !== b && null !== b ? b : null;
  nh(c, f2, e);
  a.current.lanes = e;
  Ac(a, e, d);
  Dk(a, d);
  return a;
}
function fl(a, b, c, d) {
  var e = b.current, f2 = R(), g = yi(e);
  c = dl(c);
  null === b.context ? b.context = c : b.pendingContext = c;
  b = mh(f2, g);
  b.payload = { element: a };
  d = void 0 === d ? null : d;
  null !== d && (b.callback = d);
  a = nh(e, b, g);
  null !== a && (gi(a, e, g, f2), oh(a, e, g));
  return g;
}
function gl(a) {
  a = a.current;
  if (!a.child) return null;
  switch (a.child.tag) {
    case 5:
      return a.child.stateNode;
    default:
      return a.child.stateNode;
  }
}
function hl(a, b) {
  a = a.memoizedState;
  if (null !== a && null !== a.dehydrated) {
    var c = a.retryLane;
    a.retryLane = 0 !== c && c < b ? c : b;
  }
}
function il(a, b) {
  hl(a, b);
  (a = a.alternate) && hl(a, b);
}
function jl() {
  return null;
}
var kl = "function" === typeof reportError ? reportError : function(a) {
  console.error(a);
};
function ll(a) {
  this._internalRoot = a;
}
ml.prototype.render = ll.prototype.render = function(a) {
  var b = this._internalRoot;
  if (null === b) throw Error(p(409));
  fl(a, b, null, null);
};
ml.prototype.unmount = ll.prototype.unmount = function() {
  var a = this._internalRoot;
  if (null !== a) {
    this._internalRoot = null;
    var b = a.containerInfo;
    Rk(function() {
      fl(null, a, null, null);
    });
    b[uf] = null;
  }
};
function ml(a) {
  this._internalRoot = a;
}
ml.prototype.unstable_scheduleHydration = function(a) {
  if (a) {
    var b = Hc();
    a = { blockedOn: null, target: a, priority: b };
    for (var c = 0; c < Qc.length && 0 !== b && b < Qc[c].priority; c++) ;
    Qc.splice(c, 0, a);
    0 === c && Vc(a);
  }
};
function nl(a) {
  return !(!a || 1 !== a.nodeType && 9 !== a.nodeType && 11 !== a.nodeType);
}
function ol(a) {
  return !(!a || 1 !== a.nodeType && 9 !== a.nodeType && 11 !== a.nodeType && (8 !== a.nodeType || " react-mount-point-unstable " !== a.nodeValue));
}
function pl() {
}
function ql(a, b, c, d, e) {
  if (e) {
    if ("function" === typeof d) {
      var f2 = d;
      d = function() {
        var a2 = gl(g);
        f2.call(a2);
      };
    }
    var g = el(b, d, a, 0, null, false, false, "", pl);
    a._reactRootContainer = g;
    a[uf] = g.current;
    sf(8 === a.nodeType ? a.parentNode : a);
    Rk();
    return g;
  }
  for (; e = a.lastChild; ) a.removeChild(e);
  if ("function" === typeof d) {
    var h2 = d;
    d = function() {
      var a2 = gl(k2);
      h2.call(a2);
    };
  }
  var k2 = bl(a, 0, false, null, null, false, false, "", pl);
  a._reactRootContainer = k2;
  a[uf] = k2.current;
  sf(8 === a.nodeType ? a.parentNode : a);
  Rk(function() {
    fl(b, k2, c, d);
  });
  return k2;
}
function rl(a, b, c, d, e) {
  var f2 = c._reactRootContainer;
  if (f2) {
    var g = f2;
    if ("function" === typeof e) {
      var h2 = e;
      e = function() {
        var a2 = gl(g);
        h2.call(a2);
      };
    }
    fl(b, g, a, e);
  } else g = ql(c, b, a, e, d);
  return gl(g);
}
Ec = function(a) {
  switch (a.tag) {
    case 3:
      var b = a.stateNode;
      if (b.current.memoizedState.isDehydrated) {
        var c = tc(b.pendingLanes);
        0 !== c && (Cc(b, c | 1), Dk(b, B()), 0 === (K & 6) && (Gj = B() + 500, jg()));
      }
      break;
    case 13:
      Rk(function() {
        var b2 = ih(a, 1);
        if (null !== b2) {
          var c2 = R();
          gi(b2, a, 1, c2);
        }
      }), il(a, 1);
  }
};
Fc = function(a) {
  if (13 === a.tag) {
    var b = ih(a, 134217728);
    if (null !== b) {
      var c = R();
      gi(b, a, 134217728, c);
    }
    il(a, 134217728);
  }
};
Gc = function(a) {
  if (13 === a.tag) {
    var b = yi(a), c = ih(a, b);
    if (null !== c) {
      var d = R();
      gi(c, a, b, d);
    }
    il(a, b);
  }
};
Hc = function() {
  return C;
};
Ic = function(a, b) {
  var c = C;
  try {
    return C = a, b();
  } finally {
    C = c;
  }
};
yb = function(a, b, c) {
  switch (b) {
    case "input":
      bb(a, c);
      b = c.name;
      if ("radio" === c.type && null != b) {
        for (c = a; c.parentNode; ) c = c.parentNode;
        c = c.querySelectorAll("input[name=" + JSON.stringify("" + b) + '][type="radio"]');
        for (b = 0; b < c.length; b++) {
          var d = c[b];
          if (d !== a && d.form === a.form) {
            var e = Db(d);
            if (!e) throw Error(p(90));
            Wa(d);
            bb(d, e);
          }
        }
      }
      break;
    case "textarea":
      ib(a, c);
      break;
    case "select":
      b = c.value, null != b && fb(a, !!c.multiple, b, false);
  }
};
Gb = Qk;
Hb = Rk;
var sl = { usingClientEntryPoint: false, Events: [Cb, ue, Db, Eb, Fb, Qk] }, tl = { findFiberByHostInstance: Wc, bundleType: 0, version: "18.3.1", rendererPackageName: "react-dom" };
var ul = { bundleType: tl.bundleType, version: tl.version, rendererPackageName: tl.rendererPackageName, rendererConfig: tl.rendererConfig, overrideHookState: null, overrideHookStateDeletePath: null, overrideHookStateRenamePath: null, overrideProps: null, overridePropsDeletePath: null, overridePropsRenamePath: null, setErrorHandler: null, setSuspenseHandler: null, scheduleUpdate: null, currentDispatcherRef: ua.ReactCurrentDispatcher, findHostInstanceByFiber: function(a) {
  a = Zb(a);
  return null === a ? null : a.stateNode;
}, findFiberByHostInstance: tl.findFiberByHostInstance || jl, findHostInstancesForRefresh: null, scheduleRefresh: null, scheduleRoot: null, setRefreshHandler: null, getCurrentFiber: null, reconcilerVersion: "18.3.1-next-f1338f8080-20240426" };
if ("undefined" !== typeof __REACT_DEVTOOLS_GLOBAL_HOOK__) {
  var vl = __REACT_DEVTOOLS_GLOBAL_HOOK__;
  if (!vl.isDisabled && vl.supportsFiber) try {
    kc = vl.inject(ul), lc = vl;
  } catch (a) {
  }
}
reactDom_production_min.__SECRET_INTERNALS_DO_NOT_USE_OR_YOU_WILL_BE_FIRED = sl;
reactDom_production_min.createPortal = function(a, b) {
  var c = 2 < arguments.length && void 0 !== arguments[2] ? arguments[2] : null;
  if (!nl(b)) throw Error(p(200));
  return cl(a, b, null, c);
};
reactDom_production_min.createRoot = function(a, b) {
  if (!nl(a)) throw Error(p(299));
  var c = false, d = "", e = kl;
  null !== b && void 0 !== b && (true === b.unstable_strictMode && (c = true), void 0 !== b.identifierPrefix && (d = b.identifierPrefix), void 0 !== b.onRecoverableError && (e = b.onRecoverableError));
  b = bl(a, 1, false, null, null, c, false, d, e);
  a[uf] = b.current;
  sf(8 === a.nodeType ? a.parentNode : a);
  return new ll(b);
};
reactDom_production_min.findDOMNode = function(a) {
  if (null == a) return null;
  if (1 === a.nodeType) return a;
  var b = a._reactInternals;
  if (void 0 === b) {
    if ("function" === typeof a.render) throw Error(p(188));
    a = Object.keys(a).join(",");
    throw Error(p(268, a));
  }
  a = Zb(b);
  a = null === a ? null : a.stateNode;
  return a;
};
reactDom_production_min.flushSync = function(a) {
  return Rk(a);
};
reactDom_production_min.hydrate = function(a, b, c) {
  if (!ol(b)) throw Error(p(200));
  return rl(null, a, b, true, c);
};
reactDom_production_min.hydrateRoot = function(a, b, c) {
  if (!nl(a)) throw Error(p(405));
  var d = null != c && c.hydratedSources || null, e = false, f2 = "", g = kl;
  null !== c && void 0 !== c && (true === c.unstable_strictMode && (e = true), void 0 !== c.identifierPrefix && (f2 = c.identifierPrefix), void 0 !== c.onRecoverableError && (g = c.onRecoverableError));
  b = el(b, null, a, 1, null != c ? c : null, e, false, f2, g);
  a[uf] = b.current;
  sf(a);
  if (d) for (a = 0; a < d.length; a++) c = d[a], e = c._getVersion, e = e(c._source), null == b.mutableSourceEagerHydrationData ? b.mutableSourceEagerHydrationData = [c, e] : b.mutableSourceEagerHydrationData.push(
    c,
    e
  );
  return new ml(b);
};
reactDom_production_min.render = function(a, b, c) {
  if (!ol(b)) throw Error(p(200));
  return rl(null, a, b, false, c);
};
reactDom_production_min.unmountComponentAtNode = function(a) {
  if (!ol(a)) throw Error(p(40));
  return a._reactRootContainer ? (Rk(function() {
    rl(null, null, a, false, function() {
      a._reactRootContainer = null;
      a[uf] = null;
    });
  }), true) : false;
};
reactDom_production_min.unstable_batchedUpdates = Qk;
reactDom_production_min.unstable_renderSubtreeIntoContainer = function(a, b, c, d) {
  if (!ol(c)) throw Error(p(200));
  if (null == a || void 0 === a._reactInternals) throw Error(p(38));
  return rl(a, b, c, false, d);
};
reactDom_production_min.version = "18.3.1-next-f1338f8080-20240426";
function checkDCE() {
  if (typeof __REACT_DEVTOOLS_GLOBAL_HOOK__ === "undefined" || typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE !== "function") {
    return;
  }
  try {
    __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(checkDCE);
  } catch (err) {
    console.error(err);
  }
}
{
  checkDCE();
  reactDom.exports = reactDom_production_min;
}
var reactDomExports = reactDom.exports;
const ReactDOM = /* @__PURE__ */ getDefaultExportFromCjs(reactDomExports);
var createRoot;
var m = reactDomExports;
{
  createRoot = m.createRoot;
  m.hydrateRoot;
}
/**
 * @remix-run/router v1.23.0
 *
 * Copyright (c) Remix Software Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE.md file in the root directory of this source tree.
 *
 * @license MIT
 */
function _extends$2() {
  _extends$2 = Object.assign ? Object.assign.bind() : function(target) {
    for (var i = 1; i < arguments.length; i++) {
      var source = arguments[i];
      for (var key in source) {
        if (Object.prototype.hasOwnProperty.call(source, key)) {
          target[key] = source[key];
        }
      }
    }
    return target;
  };
  return _extends$2.apply(this, arguments);
}
var Action;
(function(Action2) {
  Action2["Pop"] = "POP";
  Action2["Push"] = "PUSH";
  Action2["Replace"] = "REPLACE";
})(Action || (Action = {}));
const PopStateEventType = "popstate";
function createBrowserHistory(options) {
  if (options === void 0) {
    options = {};
  }
  function createBrowserLocation(window2, globalHistory) {
    let {
      pathname,
      search,
      hash
    } = window2.location;
    return createLocation(
      "",
      {
        pathname,
        search,
        hash
      },
      // state defaults to `null` because `window.history.state` does
      globalHistory.state && globalHistory.state.usr || null,
      globalHistory.state && globalHistory.state.key || "default"
    );
  }
  function createBrowserHref(window2, to) {
    return typeof to === "string" ? to : createPath(to);
  }
  return getUrlBasedHistory(createBrowserLocation, createBrowserHref, null, options);
}
function invariant(value, message) {
  if (value === false || value === null || typeof value === "undefined") {
    throw new Error(message);
  }
}
function warning(cond, message) {
  if (!cond) {
    if (typeof console !== "undefined") console.warn(message);
    try {
      throw new Error(message);
    } catch (e) {
    }
  }
}
function createKey() {
  return Math.random().toString(36).substr(2, 8);
}
function getHistoryState(location, index2) {
  return {
    usr: location.state,
    key: location.key,
    idx: index2
  };
}
function createLocation(current, to, state, key) {
  if (state === void 0) {
    state = null;
  }
  let location = _extends$2({
    pathname: typeof current === "string" ? current : current.pathname,
    search: "",
    hash: ""
  }, typeof to === "string" ? parsePath(to) : to, {
    state,
    // TODO: This could be cleaned up.  push/replace should probably just take
    // full Locations now and avoid the need to run through this flow at all
    // But that's a pretty big refactor to the current test suite so going to
    // keep as is for the time being and just let any incoming keys take precedence
    key: to && to.key || key || createKey()
  });
  return location;
}
function createPath(_ref) {
  let {
    pathname = "/",
    search = "",
    hash = ""
  } = _ref;
  if (search && search !== "?") pathname += search.charAt(0) === "?" ? search : "?" + search;
  if (hash && hash !== "#") pathname += hash.charAt(0) === "#" ? hash : "#" + hash;
  return pathname;
}
function parsePath(path) {
  let parsedPath = {};
  if (path) {
    let hashIndex = path.indexOf("#");
    if (hashIndex >= 0) {
      parsedPath.hash = path.substr(hashIndex);
      path = path.substr(0, hashIndex);
    }
    let searchIndex = path.indexOf("?");
    if (searchIndex >= 0) {
      parsedPath.search = path.substr(searchIndex);
      path = path.substr(0, searchIndex);
    }
    if (path) {
      parsedPath.pathname = path;
    }
  }
  return parsedPath;
}
function getUrlBasedHistory(getLocation, createHref, validateLocation, options) {
  if (options === void 0) {
    options = {};
  }
  let {
    window: window2 = document.defaultView,
    v5Compat = false
  } = options;
  let globalHistory = window2.history;
  let action = Action.Pop;
  let listener = null;
  let index2 = getIndex();
  if (index2 == null) {
    index2 = 0;
    globalHistory.replaceState(_extends$2({}, globalHistory.state, {
      idx: index2
    }), "");
  }
  function getIndex() {
    let state = globalHistory.state || {
      idx: null
    };
    return state.idx;
  }
  function handlePop() {
    action = Action.Pop;
    let nextIndex = getIndex();
    let delta = nextIndex == null ? null : nextIndex - index2;
    index2 = nextIndex;
    if (listener) {
      listener({
        action,
        location: history.location,
        delta
      });
    }
  }
  function push(to, state) {
    action = Action.Push;
    let location = createLocation(history.location, to, state);
    index2 = getIndex() + 1;
    let historyState = getHistoryState(location, index2);
    let url = history.createHref(location);
    try {
      globalHistory.pushState(historyState, "", url);
    } catch (error) {
      if (error instanceof DOMException && error.name === "DataCloneError") {
        throw error;
      }
      window2.location.assign(url);
    }
    if (v5Compat && listener) {
      listener({
        action,
        location: history.location,
        delta: 1
      });
    }
  }
  function replace(to, state) {
    action = Action.Replace;
    let location = createLocation(history.location, to, state);
    index2 = getIndex();
    let historyState = getHistoryState(location, index2);
    let url = history.createHref(location);
    globalHistory.replaceState(historyState, "", url);
    if (v5Compat && listener) {
      listener({
        action,
        location: history.location,
        delta: 0
      });
    }
  }
  function createURL(to) {
    let base = window2.location.origin !== "null" ? window2.location.origin : window2.location.href;
    let href = typeof to === "string" ? to : createPath(to);
    href = href.replace(/ $/, "%20");
    invariant(base, "No window.location.(origin|href) available to create URL for href: " + href);
    return new URL(href, base);
  }
  let history = {
    get action() {
      return action;
    },
    get location() {
      return getLocation(window2, globalHistory);
    },
    listen(fn) {
      if (listener) {
        throw new Error("A history only accepts one active listener");
      }
      window2.addEventListener(PopStateEventType, handlePop);
      listener = fn;
      return () => {
        window2.removeEventListener(PopStateEventType, handlePop);
        listener = null;
      };
    },
    createHref(to) {
      return createHref(window2, to);
    },
    createURL,
    encodeLocation(to) {
      let url = createURL(to);
      return {
        pathname: url.pathname,
        search: url.search,
        hash: url.hash
      };
    },
    push,
    replace,
    go(n2) {
      return globalHistory.go(n2);
    }
  };
  return history;
}
var ResultType;
(function(ResultType2) {
  ResultType2["data"] = "data";
  ResultType2["deferred"] = "deferred";
  ResultType2["redirect"] = "redirect";
  ResultType2["error"] = "error";
})(ResultType || (ResultType = {}));
function matchRoutes(routes, locationArg, basename) {
  if (basename === void 0) {
    basename = "/";
  }
  return matchRoutesImpl(routes, locationArg, basename);
}
function matchRoutesImpl(routes, locationArg, basename, allowPartial) {
  let location = typeof locationArg === "string" ? parsePath(locationArg) : locationArg;
  let pathname = stripBasename(location.pathname || "/", basename);
  if (pathname == null) {
    return null;
  }
  let branches = flattenRoutes(routes);
  rankRouteBranches(branches);
  let matches = null;
  for (let i = 0; matches == null && i < branches.length; ++i) {
    let decoded = decodePath(pathname);
    matches = matchRouteBranch(branches[i], decoded);
  }
  return matches;
}
function flattenRoutes(routes, branches, parentsMeta, parentPath) {
  if (branches === void 0) {
    branches = [];
  }
  if (parentsMeta === void 0) {
    parentsMeta = [];
  }
  if (parentPath === void 0) {
    parentPath = "";
  }
  let flattenRoute = (route, index2, relativePath) => {
    let meta = {
      relativePath: relativePath === void 0 ? route.path || "" : relativePath,
      caseSensitive: route.caseSensitive === true,
      childrenIndex: index2,
      route
    };
    if (meta.relativePath.startsWith("/")) {
      invariant(meta.relativePath.startsWith(parentPath), 'Absolute route path "' + meta.relativePath + '" nested under path ' + ('"' + parentPath + '" is not valid. An absolute child route path ') + "must start with the combined path of all its parent routes.");
      meta.relativePath = meta.relativePath.slice(parentPath.length);
    }
    let path = joinPaths([parentPath, meta.relativePath]);
    let routesMeta = parentsMeta.concat(meta);
    if (route.children && route.children.length > 0) {
      invariant(
        // Our types know better, but runtime JS may not!
        // @ts-expect-error
        route.index !== true,
        "Index routes must not have child routes. Please remove " + ('all child routes from route path "' + path + '".')
      );
      flattenRoutes(route.children, branches, routesMeta, path);
    }
    if (route.path == null && !route.index) {
      return;
    }
    branches.push({
      path,
      score: computeScore(path, route.index),
      routesMeta
    });
  };
  routes.forEach((route, index2) => {
    var _route$path;
    if (route.path === "" || !((_route$path = route.path) != null && _route$path.includes("?"))) {
      flattenRoute(route, index2);
    } else {
      for (let exploded of explodeOptionalSegments(route.path)) {
        flattenRoute(route, index2, exploded);
      }
    }
  });
  return branches;
}
function explodeOptionalSegments(path) {
  let segments = path.split("/");
  if (segments.length === 0) return [];
  let [first, ...rest] = segments;
  let isOptional = first.endsWith("?");
  let required = first.replace(/\?$/, "");
  if (rest.length === 0) {
    return isOptional ? [required, ""] : [required];
  }
  let restExploded = explodeOptionalSegments(rest.join("/"));
  let result = [];
  result.push(...restExploded.map((subpath) => subpath === "" ? required : [required, subpath].join("/")));
  if (isOptional) {
    result.push(...restExploded);
  }
  return result.map((exploded) => path.startsWith("/") && exploded === "" ? "/" : exploded);
}
function rankRouteBranches(branches) {
  branches.sort((a, b) => a.score !== b.score ? b.score - a.score : compareIndexes(a.routesMeta.map((meta) => meta.childrenIndex), b.routesMeta.map((meta) => meta.childrenIndex)));
}
const paramRe = /^:[\w-]+$/;
const dynamicSegmentValue = 3;
const indexRouteValue = 2;
const emptySegmentValue = 1;
const staticSegmentValue = 10;
const splatPenalty = -2;
const isSplat = (s) => s === "*";
function computeScore(path, index2) {
  let segments = path.split("/");
  let initialScore = segments.length;
  if (segments.some(isSplat)) {
    initialScore += splatPenalty;
  }
  if (index2) {
    initialScore += indexRouteValue;
  }
  return segments.filter((s) => !isSplat(s)).reduce((score, segment) => score + (paramRe.test(segment) ? dynamicSegmentValue : segment === "" ? emptySegmentValue : staticSegmentValue), initialScore);
}
function compareIndexes(a, b) {
  let siblings = a.length === b.length && a.slice(0, -1).every((n2, i) => n2 === b[i]);
  return siblings ? (
    // If two routes are siblings, we should try to match the earlier sibling
    // first. This allows people to have fine-grained control over the matching
    // behavior by simply putting routes with identical paths in the order they
    // want them tried.
    a[a.length - 1] - b[b.length - 1]
  ) : (
    // Otherwise, it doesn't really make sense to rank non-siblings by index,
    // so they sort equally.
    0
  );
}
function matchRouteBranch(branch, pathname, allowPartial) {
  let {
    routesMeta
  } = branch;
  let matchedParams = {};
  let matchedPathname = "/";
  let matches = [];
  for (let i = 0; i < routesMeta.length; ++i) {
    let meta = routesMeta[i];
    let end = i === routesMeta.length - 1;
    let remainingPathname = matchedPathname === "/" ? pathname : pathname.slice(matchedPathname.length) || "/";
    let match = matchPath({
      path: meta.relativePath,
      caseSensitive: meta.caseSensitive,
      end
    }, remainingPathname);
    let route = meta.route;
    if (!match) {
      return null;
    }
    Object.assign(matchedParams, match.params);
    matches.push({
      // TODO: Can this as be avoided?
      params: matchedParams,
      pathname: joinPaths([matchedPathname, match.pathname]),
      pathnameBase: normalizePathname(joinPaths([matchedPathname, match.pathnameBase])),
      route
    });
    if (match.pathnameBase !== "/") {
      matchedPathname = joinPaths([matchedPathname, match.pathnameBase]);
    }
  }
  return matches;
}
function matchPath(pattern, pathname) {
  if (typeof pattern === "string") {
    pattern = {
      path: pattern,
      caseSensitive: false,
      end: true
    };
  }
  let [matcher, compiledParams] = compilePath(pattern.path, pattern.caseSensitive, pattern.end);
  let match = pathname.match(matcher);
  if (!match) return null;
  let matchedPathname = match[0];
  let pathnameBase = matchedPathname.replace(/(.)\/+$/, "$1");
  let captureGroups = match.slice(1);
  let params = compiledParams.reduce((memo, _ref, index2) => {
    let {
      paramName,
      isOptional
    } = _ref;
    if (paramName === "*") {
      let splatValue = captureGroups[index2] || "";
      pathnameBase = matchedPathname.slice(0, matchedPathname.length - splatValue.length).replace(/(.)\/+$/, "$1");
    }
    const value = captureGroups[index2];
    if (isOptional && !value) {
      memo[paramName] = void 0;
    } else {
      memo[paramName] = (value || "").replace(/%2F/g, "/");
    }
    return memo;
  }, {});
  return {
    params,
    pathname: matchedPathname,
    pathnameBase,
    pattern
  };
}
function compilePath(path, caseSensitive, end) {
  if (caseSensitive === void 0) {
    caseSensitive = false;
  }
  if (end === void 0) {
    end = true;
  }
  warning(path === "*" || !path.endsWith("*") || path.endsWith("/*"), 'Route path "' + path + '" will be treated as if it were ' + ('"' + path.replace(/\*$/, "/*") + '" because the `*` character must ') + "always follow a `/` in the pattern. To get rid of this warning, " + ('please change the route path to "' + path.replace(/\*$/, "/*") + '".'));
  let params = [];
  let regexpSource = "^" + path.replace(/\/*\*?$/, "").replace(/^\/*/, "/").replace(/[\\.*+^${}|()[\]]/g, "\\$&").replace(/\/:([\w-]+)(\?)?/g, (_2, paramName, isOptional) => {
    params.push({
      paramName,
      isOptional: isOptional != null
    });
    return isOptional ? "/?([^\\/]+)?" : "/([^\\/]+)";
  });
  if (path.endsWith("*")) {
    params.push({
      paramName: "*"
    });
    regexpSource += path === "*" || path === "/*" ? "(.*)$" : "(?:\\/(.+)|\\/*)$";
  } else if (end) {
    regexpSource += "\\/*$";
  } else if (path !== "" && path !== "/") {
    regexpSource += "(?:(?=\\/|$))";
  } else ;
  let matcher = new RegExp(regexpSource, caseSensitive ? void 0 : "i");
  return [matcher, params];
}
function decodePath(value) {
  try {
    return value.split("/").map((v2) => decodeURIComponent(v2).replace(/\//g, "%2F")).join("/");
  } catch (error) {
    warning(false, 'The URL path "' + value + '" could not be decoded because it is is a malformed URL segment. This is probably due to a bad percent ' + ("encoding (" + error + ")."));
    return value;
  }
}
function stripBasename(pathname, basename) {
  if (basename === "/") return pathname;
  if (!pathname.toLowerCase().startsWith(basename.toLowerCase())) {
    return null;
  }
  let startIndex = basename.endsWith("/") ? basename.length - 1 : basename.length;
  let nextChar = pathname.charAt(startIndex);
  if (nextChar && nextChar !== "/") {
    return null;
  }
  return pathname.slice(startIndex) || "/";
}
function resolvePath(to, fromPathname) {
  if (fromPathname === void 0) {
    fromPathname = "/";
  }
  let {
    pathname: toPathname,
    search = "",
    hash = ""
  } = typeof to === "string" ? parsePath(to) : to;
  let pathname = toPathname ? toPathname.startsWith("/") ? toPathname : resolvePathname(toPathname, fromPathname) : fromPathname;
  return {
    pathname,
    search: normalizeSearch(search),
    hash: normalizeHash(hash)
  };
}
function resolvePathname(relativePath, fromPathname) {
  let segments = fromPathname.replace(/\/+$/, "").split("/");
  let relativeSegments = relativePath.split("/");
  relativeSegments.forEach((segment) => {
    if (segment === "..") {
      if (segments.length > 1) segments.pop();
    } else if (segment !== ".") {
      segments.push(segment);
    }
  });
  return segments.length > 1 ? segments.join("/") : "/";
}
function getInvalidPathError(char, field, dest, path) {
  return "Cannot include a '" + char + "' character in a manually specified " + ("`to." + field + "` field [" + JSON.stringify(path) + "].  Please separate it out to the ") + ("`to." + dest + "` field. Alternatively you may provide the full path as ") + 'a string in <Link to="..."> and the router will parse it for you.';
}
function getPathContributingMatches(matches) {
  return matches.filter((match, index2) => index2 === 0 || match.route.path && match.route.path.length > 0);
}
function getResolveToMatches(matches, v7_relativeSplatPath) {
  let pathMatches = getPathContributingMatches(matches);
  if (v7_relativeSplatPath) {
    return pathMatches.map((match, idx) => idx === pathMatches.length - 1 ? match.pathname : match.pathnameBase);
  }
  return pathMatches.map((match) => match.pathnameBase);
}
function resolveTo(toArg, routePathnames, locationPathname, isPathRelative) {
  if (isPathRelative === void 0) {
    isPathRelative = false;
  }
  let to;
  if (typeof toArg === "string") {
    to = parsePath(toArg);
  } else {
    to = _extends$2({}, toArg);
    invariant(!to.pathname || !to.pathname.includes("?"), getInvalidPathError("?", "pathname", "search", to));
    invariant(!to.pathname || !to.pathname.includes("#"), getInvalidPathError("#", "pathname", "hash", to));
    invariant(!to.search || !to.search.includes("#"), getInvalidPathError("#", "search", "hash", to));
  }
  let isEmptyPath = toArg === "" || to.pathname === "";
  let toPathname = isEmptyPath ? "/" : to.pathname;
  let from;
  if (toPathname == null) {
    from = locationPathname;
  } else {
    let routePathnameIndex = routePathnames.length - 1;
    if (!isPathRelative && toPathname.startsWith("..")) {
      let toSegments = toPathname.split("/");
      while (toSegments[0] === "..") {
        toSegments.shift();
        routePathnameIndex -= 1;
      }
      to.pathname = toSegments.join("/");
    }
    from = routePathnameIndex >= 0 ? routePathnames[routePathnameIndex] : "/";
  }
  let path = resolvePath(to, from);
  let hasExplicitTrailingSlash = toPathname && toPathname !== "/" && toPathname.endsWith("/");
  let hasCurrentTrailingSlash = (isEmptyPath || toPathname === ".") && locationPathname.endsWith("/");
  if (!path.pathname.endsWith("/") && (hasExplicitTrailingSlash || hasCurrentTrailingSlash)) {
    path.pathname += "/";
  }
  return path;
}
const joinPaths = (paths) => paths.join("/").replace(/\/\/+/g, "/");
const normalizePathname = (pathname) => pathname.replace(/\/+$/, "").replace(/^\/*/, "/");
const normalizeSearch = (search) => !search || search === "?" ? "" : search.startsWith("?") ? search : "?" + search;
const normalizeHash = (hash) => !hash || hash === "#" ? "" : hash.startsWith("#") ? hash : "#" + hash;
function isRouteErrorResponse(error) {
  return error != null && typeof error.status === "number" && typeof error.statusText === "string" && typeof error.internal === "boolean" && "data" in error;
}
const validMutationMethodsArr = ["post", "put", "patch", "delete"];
new Set(validMutationMethodsArr);
const validRequestMethodsArr = ["get", ...validMutationMethodsArr];
new Set(validRequestMethodsArr);
/**
 * React Router v6.30.1
 *
 * Copyright (c) Remix Software Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE.md file in the root directory of this source tree.
 *
 * @license MIT
 */
function _extends$1() {
  _extends$1 = Object.assign ? Object.assign.bind() : function(target) {
    for (var i = 1; i < arguments.length; i++) {
      var source = arguments[i];
      for (var key in source) {
        if (Object.prototype.hasOwnProperty.call(source, key)) {
          target[key] = source[key];
        }
      }
    }
    return target;
  };
  return _extends$1.apply(this, arguments);
}
const DataRouterContext = /* @__PURE__ */ reactExports.createContext(null);
const DataRouterStateContext = /* @__PURE__ */ reactExports.createContext(null);
const NavigationContext = /* @__PURE__ */ reactExports.createContext(null);
const LocationContext = /* @__PURE__ */ reactExports.createContext(null);
const RouteContext = /* @__PURE__ */ reactExports.createContext({
  outlet: null,
  matches: [],
  isDataRoute: false
});
const RouteErrorContext = /* @__PURE__ */ reactExports.createContext(null);
function useHref(to, _temp) {
  let {
    relative
  } = _temp === void 0 ? {} : _temp;
  !useInRouterContext() ? invariant(false) : void 0;
  let {
    basename,
    navigator: navigator2
  } = reactExports.useContext(NavigationContext);
  let {
    hash,
    pathname,
    search
  } = useResolvedPath(to, {
    relative
  });
  let joinedPathname = pathname;
  if (basename !== "/") {
    joinedPathname = pathname === "/" ? basename : joinPaths([basename, pathname]);
  }
  return navigator2.createHref({
    pathname: joinedPathname,
    search,
    hash
  });
}
function useInRouterContext() {
  return reactExports.useContext(LocationContext) != null;
}
function useLocation() {
  !useInRouterContext() ? invariant(false) : void 0;
  return reactExports.useContext(LocationContext).location;
}
function useIsomorphicLayoutEffect(cb2) {
  let isStatic = reactExports.useContext(NavigationContext).static;
  if (!isStatic) {
    reactExports.useLayoutEffect(cb2);
  }
}
function useNavigate() {
  let {
    isDataRoute
  } = reactExports.useContext(RouteContext);
  return isDataRoute ? useNavigateStable() : useNavigateUnstable();
}
function useNavigateUnstable() {
  !useInRouterContext() ? invariant(false) : void 0;
  let dataRouterContext = reactExports.useContext(DataRouterContext);
  let {
    basename,
    future,
    navigator: navigator2
  } = reactExports.useContext(NavigationContext);
  let {
    matches
  } = reactExports.useContext(RouteContext);
  let {
    pathname: locationPathname
  } = useLocation();
  let routePathnamesJson = JSON.stringify(getResolveToMatches(matches, future.v7_relativeSplatPath));
  let activeRef = reactExports.useRef(false);
  useIsomorphicLayoutEffect(() => {
    activeRef.current = true;
  });
  let navigate = reactExports.useCallback(function(to, options) {
    if (options === void 0) {
      options = {};
    }
    if (!activeRef.current) return;
    if (typeof to === "number") {
      navigator2.go(to);
      return;
    }
    let path = resolveTo(to, JSON.parse(routePathnamesJson), locationPathname, options.relative === "path");
    if (dataRouterContext == null && basename !== "/") {
      path.pathname = path.pathname === "/" ? basename : joinPaths([basename, path.pathname]);
    }
    (!!options.replace ? navigator2.replace : navigator2.push)(path, options.state, options);
  }, [basename, navigator2, routePathnamesJson, locationPathname, dataRouterContext]);
  return navigate;
}
function useResolvedPath(to, _temp2) {
  let {
    relative
  } = _temp2 === void 0 ? {} : _temp2;
  let {
    future
  } = reactExports.useContext(NavigationContext);
  let {
    matches
  } = reactExports.useContext(RouteContext);
  let {
    pathname: locationPathname
  } = useLocation();
  let routePathnamesJson = JSON.stringify(getResolveToMatches(matches, future.v7_relativeSplatPath));
  return reactExports.useMemo(() => resolveTo(to, JSON.parse(routePathnamesJson), locationPathname, relative === "path"), [to, routePathnamesJson, locationPathname, relative]);
}
function useRoutes(routes, locationArg) {
  return useRoutesImpl(routes, locationArg);
}
function useRoutesImpl(routes, locationArg, dataRouterState, future) {
  !useInRouterContext() ? invariant(false) : void 0;
  let {
    navigator: navigator2
  } = reactExports.useContext(NavigationContext);
  let {
    matches: parentMatches
  } = reactExports.useContext(RouteContext);
  let routeMatch = parentMatches[parentMatches.length - 1];
  let parentParams = routeMatch ? routeMatch.params : {};
  routeMatch ? routeMatch.pathname : "/";
  let parentPathnameBase = routeMatch ? routeMatch.pathnameBase : "/";
  routeMatch && routeMatch.route;
  let locationFromContext = useLocation();
  let location;
  if (locationArg) {
    var _parsedLocationArg$pa;
    let parsedLocationArg = typeof locationArg === "string" ? parsePath(locationArg) : locationArg;
    !(parentPathnameBase === "/" || ((_parsedLocationArg$pa = parsedLocationArg.pathname) == null ? void 0 : _parsedLocationArg$pa.startsWith(parentPathnameBase))) ? invariant(false) : void 0;
    location = parsedLocationArg;
  } else {
    location = locationFromContext;
  }
  let pathname = location.pathname || "/";
  let remainingPathname = pathname;
  if (parentPathnameBase !== "/") {
    let parentSegments = parentPathnameBase.replace(/^\//, "").split("/");
    let segments = pathname.replace(/^\//, "").split("/");
    remainingPathname = "/" + segments.slice(parentSegments.length).join("/");
  }
  let matches = matchRoutes(routes, {
    pathname: remainingPathname
  });
  let renderedMatches = _renderMatches(matches && matches.map((match) => Object.assign({}, match, {
    params: Object.assign({}, parentParams, match.params),
    pathname: joinPaths([
      parentPathnameBase,
      // Re-encode pathnames that were decoded inside matchRoutes
      navigator2.encodeLocation ? navigator2.encodeLocation(match.pathname).pathname : match.pathname
    ]),
    pathnameBase: match.pathnameBase === "/" ? parentPathnameBase : joinPaths([
      parentPathnameBase,
      // Re-encode pathnames that were decoded inside matchRoutes
      navigator2.encodeLocation ? navigator2.encodeLocation(match.pathnameBase).pathname : match.pathnameBase
    ])
  })), parentMatches, dataRouterState, future);
  if (locationArg && renderedMatches) {
    return /* @__PURE__ */ reactExports.createElement(LocationContext.Provider, {
      value: {
        location: _extends$1({
          pathname: "/",
          search: "",
          hash: "",
          state: null,
          key: "default"
        }, location),
        navigationType: Action.Pop
      }
    }, renderedMatches);
  }
  return renderedMatches;
}
function DefaultErrorComponent() {
  let error = useRouteError();
  let message = isRouteErrorResponse(error) ? error.status + " " + error.statusText : error instanceof Error ? error.message : JSON.stringify(error);
  let stack = error instanceof Error ? error.stack : null;
  let lightgrey = "rgba(200,200,200, 0.5)";
  let preStyles = {
    padding: "0.5rem",
    backgroundColor: lightgrey
  };
  let devInfo = null;
  return /* @__PURE__ */ reactExports.createElement(reactExports.Fragment, null, /* @__PURE__ */ reactExports.createElement("h2", null, "Unexpected Application Error!"), /* @__PURE__ */ reactExports.createElement("h3", {
    style: {
      fontStyle: "italic"
    }
  }, message), stack ? /* @__PURE__ */ reactExports.createElement("pre", {
    style: preStyles
  }, stack) : null, devInfo);
}
const defaultErrorElement = /* @__PURE__ */ reactExports.createElement(DefaultErrorComponent, null);
class RenderErrorBoundary extends reactExports.Component {
  constructor(props) {
    super(props);
    this.state = {
      location: props.location,
      revalidation: props.revalidation,
      error: props.error
    };
  }
  static getDerivedStateFromError(error) {
    return {
      error
    };
  }
  static getDerivedStateFromProps(props, state) {
    if (state.location !== props.location || state.revalidation !== "idle" && props.revalidation === "idle") {
      return {
        error: props.error,
        location: props.location,
        revalidation: props.revalidation
      };
    }
    return {
      error: props.error !== void 0 ? props.error : state.error,
      location: state.location,
      revalidation: props.revalidation || state.revalidation
    };
  }
  componentDidCatch(error, errorInfo) {
    console.error("React Router caught the following error during render", error, errorInfo);
  }
  render() {
    return this.state.error !== void 0 ? /* @__PURE__ */ reactExports.createElement(RouteContext.Provider, {
      value: this.props.routeContext
    }, /* @__PURE__ */ reactExports.createElement(RouteErrorContext.Provider, {
      value: this.state.error,
      children: this.props.component
    })) : this.props.children;
  }
}
function RenderedRoute(_ref) {
  let {
    routeContext,
    match,
    children
  } = _ref;
  let dataRouterContext = reactExports.useContext(DataRouterContext);
  if (dataRouterContext && dataRouterContext.static && dataRouterContext.staticContext && (match.route.errorElement || match.route.ErrorBoundary)) {
    dataRouterContext.staticContext._deepestRenderedBoundaryId = match.route.id;
  }
  return /* @__PURE__ */ reactExports.createElement(RouteContext.Provider, {
    value: routeContext
  }, children);
}
function _renderMatches(matches, parentMatches, dataRouterState, future) {
  var _dataRouterState;
  if (parentMatches === void 0) {
    parentMatches = [];
  }
  if (dataRouterState === void 0) {
    dataRouterState = null;
  }
  if (future === void 0) {
    future = null;
  }
  if (matches == null) {
    var _future;
    if (!dataRouterState) {
      return null;
    }
    if (dataRouterState.errors) {
      matches = dataRouterState.matches;
    } else if ((_future = future) != null && _future.v7_partialHydration && parentMatches.length === 0 && !dataRouterState.initialized && dataRouterState.matches.length > 0) {
      matches = dataRouterState.matches;
    } else {
      return null;
    }
  }
  let renderedMatches = matches;
  let errors = (_dataRouterState = dataRouterState) == null ? void 0 : _dataRouterState.errors;
  if (errors != null) {
    let errorIndex = renderedMatches.findIndex((m2) => m2.route.id && (errors == null ? void 0 : errors[m2.route.id]) !== void 0);
    !(errorIndex >= 0) ? invariant(false) : void 0;
    renderedMatches = renderedMatches.slice(0, Math.min(renderedMatches.length, errorIndex + 1));
  }
  let renderFallback = false;
  let fallbackIndex = -1;
  if (dataRouterState && future && future.v7_partialHydration) {
    for (let i = 0; i < renderedMatches.length; i++) {
      let match = renderedMatches[i];
      if (match.route.HydrateFallback || match.route.hydrateFallbackElement) {
        fallbackIndex = i;
      }
      if (match.route.id) {
        let {
          loaderData,
          errors: errors2
        } = dataRouterState;
        let needsToRunLoader = match.route.loader && loaderData[match.route.id] === void 0 && (!errors2 || errors2[match.route.id] === void 0);
        if (match.route.lazy || needsToRunLoader) {
          renderFallback = true;
          if (fallbackIndex >= 0) {
            renderedMatches = renderedMatches.slice(0, fallbackIndex + 1);
          } else {
            renderedMatches = [renderedMatches[0]];
          }
          break;
        }
      }
    }
  }
  return renderedMatches.reduceRight((outlet, match, index2) => {
    let error;
    let shouldRenderHydrateFallback = false;
    let errorElement = null;
    let hydrateFallbackElement = null;
    if (dataRouterState) {
      error = errors && match.route.id ? errors[match.route.id] : void 0;
      errorElement = match.route.errorElement || defaultErrorElement;
      if (renderFallback) {
        if (fallbackIndex < 0 && index2 === 0) {
          warningOnce("route-fallback");
          shouldRenderHydrateFallback = true;
          hydrateFallbackElement = null;
        } else if (fallbackIndex === index2) {
          shouldRenderHydrateFallback = true;
          hydrateFallbackElement = match.route.hydrateFallbackElement || null;
        }
      }
    }
    let matches2 = parentMatches.concat(renderedMatches.slice(0, index2 + 1));
    let getChildren = () => {
      let children;
      if (error) {
        children = errorElement;
      } else if (shouldRenderHydrateFallback) {
        children = hydrateFallbackElement;
      } else if (match.route.Component) {
        children = /* @__PURE__ */ reactExports.createElement(match.route.Component, null);
      } else if (match.route.element) {
        children = match.route.element;
      } else {
        children = outlet;
      }
      return /* @__PURE__ */ reactExports.createElement(RenderedRoute, {
        match,
        routeContext: {
          outlet,
          matches: matches2,
          isDataRoute: dataRouterState != null
        },
        children
      });
    };
    return dataRouterState && (match.route.ErrorBoundary || match.route.errorElement || index2 === 0) ? /* @__PURE__ */ reactExports.createElement(RenderErrorBoundary, {
      location: dataRouterState.location,
      revalidation: dataRouterState.revalidation,
      component: errorElement,
      error,
      children: getChildren(),
      routeContext: {
        outlet: null,
        matches: matches2,
        isDataRoute: true
      }
    }) : getChildren();
  }, null);
}
var DataRouterHook$1 = /* @__PURE__ */ function(DataRouterHook2) {
  DataRouterHook2["UseBlocker"] = "useBlocker";
  DataRouterHook2["UseRevalidator"] = "useRevalidator";
  DataRouterHook2["UseNavigateStable"] = "useNavigate";
  return DataRouterHook2;
}(DataRouterHook$1 || {});
var DataRouterStateHook$1 = /* @__PURE__ */ function(DataRouterStateHook2) {
  DataRouterStateHook2["UseBlocker"] = "useBlocker";
  DataRouterStateHook2["UseLoaderData"] = "useLoaderData";
  DataRouterStateHook2["UseActionData"] = "useActionData";
  DataRouterStateHook2["UseRouteError"] = "useRouteError";
  DataRouterStateHook2["UseNavigation"] = "useNavigation";
  DataRouterStateHook2["UseRouteLoaderData"] = "useRouteLoaderData";
  DataRouterStateHook2["UseMatches"] = "useMatches";
  DataRouterStateHook2["UseRevalidator"] = "useRevalidator";
  DataRouterStateHook2["UseNavigateStable"] = "useNavigate";
  DataRouterStateHook2["UseRouteId"] = "useRouteId";
  return DataRouterStateHook2;
}(DataRouterStateHook$1 || {});
function useDataRouterContext(hookName) {
  let ctx = reactExports.useContext(DataRouterContext);
  !ctx ? invariant(false) : void 0;
  return ctx;
}
function useDataRouterState(hookName) {
  let state = reactExports.useContext(DataRouterStateContext);
  !state ? invariant(false) : void 0;
  return state;
}
function useRouteContext(hookName) {
  let route = reactExports.useContext(RouteContext);
  !route ? invariant(false) : void 0;
  return route;
}
function useCurrentRouteId(hookName) {
  let route = useRouteContext();
  let thisRoute = route.matches[route.matches.length - 1];
  !thisRoute.route.id ? invariant(false) : void 0;
  return thisRoute.route.id;
}
function useRouteError() {
  var _state$errors;
  let error = reactExports.useContext(RouteErrorContext);
  let state = useDataRouterState();
  let routeId = useCurrentRouteId();
  if (error !== void 0) {
    return error;
  }
  return (_state$errors = state.errors) == null ? void 0 : _state$errors[routeId];
}
function useNavigateStable() {
  let {
    router
  } = useDataRouterContext(DataRouterHook$1.UseNavigateStable);
  let id2 = useCurrentRouteId(DataRouterStateHook$1.UseNavigateStable);
  let activeRef = reactExports.useRef(false);
  useIsomorphicLayoutEffect(() => {
    activeRef.current = true;
  });
  let navigate = reactExports.useCallback(function(to, options) {
    if (options === void 0) {
      options = {};
    }
    if (!activeRef.current) return;
    if (typeof to === "number") {
      router.navigate(to);
    } else {
      router.navigate(to, _extends$1({
        fromRouteId: id2
      }, options));
    }
  }, [router, id2]);
  return navigate;
}
const alreadyWarned$1 = {};
function warningOnce(key, cond, message) {
  if (!alreadyWarned$1[key]) {
    alreadyWarned$1[key] = true;
  }
}
function logV6DeprecationWarnings(renderFuture, routerFuture) {
  if ((renderFuture == null ? void 0 : renderFuture.v7_startTransition) === void 0) ;
  if ((renderFuture == null ? void 0 : renderFuture.v7_relativeSplatPath) === void 0 && true) ;
}
function Navigate(_ref4) {
  let {
    to,
    replace: replace2,
    state,
    relative
  } = _ref4;
  !useInRouterContext() ? invariant(false) : void 0;
  let {
    future,
    static: isStatic
  } = reactExports.useContext(NavigationContext);
  let {
    matches
  } = reactExports.useContext(RouteContext);
  let {
    pathname: locationPathname
  } = useLocation();
  let navigate = useNavigate();
  let path = resolveTo(to, getResolveToMatches(matches, future.v7_relativeSplatPath), locationPathname, relative === "path");
  let jsonPath = JSON.stringify(path);
  reactExports.useEffect(() => navigate(JSON.parse(jsonPath), {
    replace: replace2,
    state,
    relative
  }), [navigate, jsonPath, relative, replace2, state]);
  return null;
}
function Route(_props) {
  invariant(false);
}
function Router(_ref5) {
  let {
    basename: basenameProp = "/",
    children = null,
    location: locationProp,
    navigationType = Action.Pop,
    navigator: navigator2,
    static: staticProp = false,
    future
  } = _ref5;
  !!useInRouterContext() ? invariant(false) : void 0;
  let basename = basenameProp.replace(/^\/*/, "/");
  let navigationContext = reactExports.useMemo(() => ({
    basename,
    navigator: navigator2,
    static: staticProp,
    future: _extends$1({
      v7_relativeSplatPath: false
    }, future)
  }), [basename, future, navigator2, staticProp]);
  if (typeof locationProp === "string") {
    locationProp = parsePath(locationProp);
  }
  let {
    pathname = "/",
    search = "",
    hash = "",
    state = null,
    key = "default"
  } = locationProp;
  let locationContext = reactExports.useMemo(() => {
    let trailingPathname = stripBasename(pathname, basename);
    if (trailingPathname == null) {
      return null;
    }
    return {
      location: {
        pathname: trailingPathname,
        search,
        hash,
        state,
        key
      },
      navigationType
    };
  }, [basename, pathname, search, hash, state, key, navigationType]);
  if (locationContext == null) {
    return null;
  }
  return /* @__PURE__ */ reactExports.createElement(NavigationContext.Provider, {
    value: navigationContext
  }, /* @__PURE__ */ reactExports.createElement(LocationContext.Provider, {
    children,
    value: locationContext
  }));
}
function Routes(_ref6) {
  let {
    children,
    location
  } = _ref6;
  return useRoutes(createRoutesFromChildren(children), location);
}
new Promise(() => {
});
function createRoutesFromChildren(children, parentPath) {
  if (parentPath === void 0) {
    parentPath = [];
  }
  let routes = [];
  reactExports.Children.forEach(children, (element, index2) => {
    if (!/* @__PURE__ */ reactExports.isValidElement(element)) {
      return;
    }
    let treePath = [...parentPath, index2];
    if (element.type === reactExports.Fragment) {
      routes.push.apply(routes, createRoutesFromChildren(element.props.children, treePath));
      return;
    }
    !(element.type === Route) ? invariant(false) : void 0;
    !(!element.props.index || !element.props.children) ? invariant(false) : void 0;
    let route = {
      id: element.props.id || treePath.join("-"),
      caseSensitive: element.props.caseSensitive,
      element: element.props.element,
      Component: element.props.Component,
      index: element.props.index,
      path: element.props.path,
      loader: element.props.loader,
      action: element.props.action,
      errorElement: element.props.errorElement,
      ErrorBoundary: element.props.ErrorBoundary,
      hasErrorBoundary: element.props.ErrorBoundary != null || element.props.errorElement != null,
      shouldRevalidate: element.props.shouldRevalidate,
      handle: element.props.handle,
      lazy: element.props.lazy
    };
    if (element.props.children) {
      route.children = createRoutesFromChildren(element.props.children, treePath);
    }
    routes.push(route);
  });
  return routes;
}
/**
 * React Router DOM v6.30.1
 *
 * Copyright (c) Remix Software Inc.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE.md file in the root directory of this source tree.
 *
 * @license MIT
 */
function _extends() {
  _extends = Object.assign ? Object.assign.bind() : function(target) {
    for (var i = 1; i < arguments.length; i++) {
      var source = arguments[i];
      for (var key in source) {
        if (Object.prototype.hasOwnProperty.call(source, key)) {
          target[key] = source[key];
        }
      }
    }
    return target;
  };
  return _extends.apply(this, arguments);
}
function _objectWithoutPropertiesLoose$5(source, excluded) {
  if (source == null) return {};
  var target = {};
  var sourceKeys = Object.keys(source);
  var key, i;
  for (i = 0; i < sourceKeys.length; i++) {
    key = sourceKeys[i];
    if (excluded.indexOf(key) >= 0) continue;
    target[key] = source[key];
  }
  return target;
}
function isModifiedEvent(event) {
  return !!(event.metaKey || event.altKey || event.ctrlKey || event.shiftKey);
}
function shouldProcessLinkClick(event, target) {
  return event.button === 0 && // Ignore everything but left clicks
  (!target || target === "_self") && // Let browser handle "target=_blank" etc.
  !isModifiedEvent(event);
}
function createSearchParams(init2) {
  if (init2 === void 0) {
    init2 = "";
  }
  return new URLSearchParams(typeof init2 === "string" || Array.isArray(init2) || init2 instanceof URLSearchParams ? init2 : Object.keys(init2).reduce((memo, key) => {
    let value = init2[key];
    return memo.concat(Array.isArray(value) ? value.map((v2) => [key, v2]) : [[key, value]]);
  }, []));
}
function getSearchParamsForLocation(locationSearch, defaultSearchParams) {
  let searchParams = createSearchParams(locationSearch);
  if (defaultSearchParams) {
    defaultSearchParams.forEach((_2, key) => {
      if (!searchParams.has(key)) {
        defaultSearchParams.getAll(key).forEach((value) => {
          searchParams.append(key, value);
        });
      }
    });
  }
  return searchParams;
}
const _excluded$4 = ["onClick", "relative", "reloadDocument", "replace", "state", "target", "to", "preventScrollReset", "viewTransition"];
const REACT_ROUTER_VERSION = "6";
try {
  window.__reactRouterVersion = REACT_ROUTER_VERSION;
} catch (e) {
}
const START_TRANSITION = "startTransition";
const startTransitionImpl = React[START_TRANSITION];
function BrowserRouter(_ref4) {
  let {
    basename,
    children,
    future,
    window: window2
  } = _ref4;
  let historyRef = reactExports.useRef();
  if (historyRef.current == null) {
    historyRef.current = createBrowserHistory({
      window: window2,
      v5Compat: true
    });
  }
  let history = historyRef.current;
  let [state, setStateImpl] = reactExports.useState({
    action: history.action,
    location: history.location
  });
  let {
    v7_startTransition
  } = future || {};
  let setState2 = reactExports.useCallback((newState) => {
    v7_startTransition && startTransitionImpl ? startTransitionImpl(() => setStateImpl(newState)) : setStateImpl(newState);
  }, [setStateImpl, v7_startTransition]);
  reactExports.useLayoutEffect(() => history.listen(setState2), [history, setState2]);
  reactExports.useEffect(() => logV6DeprecationWarnings(future), [future]);
  return /* @__PURE__ */ reactExports.createElement(Router, {
    basename,
    children,
    location: state.location,
    navigationType: state.action,
    navigator: history,
    future
  });
}
const isBrowser = typeof window !== "undefined" && typeof window.document !== "undefined" && typeof window.document.createElement !== "undefined";
const ABSOLUTE_URL_REGEX = /^(?:[a-z][a-z0-9+.-]*:|\/\/)/i;
const Link = /* @__PURE__ */ reactExports.forwardRef(function LinkWithRef(_ref7, ref) {
  let {
    onClick,
    relative,
    reloadDocument,
    replace: replace2,
    state,
    target,
    to,
    preventScrollReset,
    viewTransition
  } = _ref7, rest = _objectWithoutPropertiesLoose$5(_ref7, _excluded$4);
  let {
    basename
  } = reactExports.useContext(NavigationContext);
  let absoluteHref;
  let isExternal = false;
  if (typeof to === "string" && ABSOLUTE_URL_REGEX.test(to)) {
    absoluteHref = to;
    if (isBrowser) {
      try {
        let currentUrl = new URL(window.location.href);
        let targetUrl = to.startsWith("//") ? new URL(currentUrl.protocol + to) : new URL(to);
        let path = stripBasename(targetUrl.pathname, basename);
        if (targetUrl.origin === currentUrl.origin && path != null) {
          to = path + targetUrl.search + targetUrl.hash;
        } else {
          isExternal = true;
        }
      } catch (e) {
      }
    }
  }
  let href = useHref(to, {
    relative
  });
  let internalOnClick = useLinkClickHandler(to, {
    replace: replace2,
    state,
    target,
    preventScrollReset,
    relative,
    viewTransition
  });
  function handleClick(event) {
    if (onClick) onClick(event);
    if (!event.defaultPrevented) {
      internalOnClick(event);
    }
  }
  return (
    // eslint-disable-next-line jsx-a11y/anchor-has-content
    /* @__PURE__ */ reactExports.createElement("a", _extends({}, rest, {
      href: absoluteHref || href,
      onClick: isExternal || reloadDocument ? onClick : handleClick,
      ref,
      target
    }))
  );
});
var DataRouterHook;
(function(DataRouterHook2) {
  DataRouterHook2["UseScrollRestoration"] = "useScrollRestoration";
  DataRouterHook2["UseSubmit"] = "useSubmit";
  DataRouterHook2["UseSubmitFetcher"] = "useSubmitFetcher";
  DataRouterHook2["UseFetcher"] = "useFetcher";
  DataRouterHook2["useViewTransitionState"] = "useViewTransitionState";
})(DataRouterHook || (DataRouterHook = {}));
var DataRouterStateHook;
(function(DataRouterStateHook2) {
  DataRouterStateHook2["UseFetcher"] = "useFetcher";
  DataRouterStateHook2["UseFetchers"] = "useFetchers";
  DataRouterStateHook2["UseScrollRestoration"] = "useScrollRestoration";
})(DataRouterStateHook || (DataRouterStateHook = {}));
function useLinkClickHandler(to, _temp) {
  let {
    target,
    replace: replaceProp,
    state,
    preventScrollReset,
    relative,
    viewTransition
  } = _temp === void 0 ? {} : _temp;
  let navigate = useNavigate();
  let location = useLocation();
  let path = useResolvedPath(to, {
    relative
  });
  return reactExports.useCallback((event) => {
    if (shouldProcessLinkClick(event, target)) {
      event.preventDefault();
      let replace2 = replaceProp !== void 0 ? replaceProp : createPath(location) === createPath(path);
      navigate(to, {
        replace: replace2,
        state,
        preventScrollReset,
        relative,
        viewTransition
      });
    }
  }, [location, navigate, path, replaceProp, state, target, to, preventScrollReset, relative, viewTransition]);
}
function useSearchParams(defaultInit) {
  let defaultSearchParamsRef = reactExports.useRef(createSearchParams(defaultInit));
  let hasSetSearchParamsRef = reactExports.useRef(false);
  let location = useLocation();
  let searchParams = reactExports.useMemo(() => (
    // Only merge in the defaults if we haven't yet called setSearchParams.
    // Once we call that we want those to take precedence, otherwise you can't
    // remove a param with setSearchParams({}) if it has an initial value
    getSearchParamsForLocation(location.search, hasSetSearchParamsRef.current ? null : defaultSearchParamsRef.current)
  ), [location.search]);
  let navigate = useNavigate();
  let setSearchParams = reactExports.useCallback((nextInit, navigateOptions) => {
    const newSearchParams = createSearchParams(typeof nextInit === "function" ? nextInit(searchParams) : nextInit);
    hasSetSearchParamsRef.current = true;
    navigate("?" + newSearchParams, navigateOptions);
  }, [navigate, searchParams]);
  return [searchParams, setSearchParams];
}
const AuthContext = reactExports.createContext(null);
function AuthProvider({ children }) {
  const [user, setUser] = reactExports.useState(null);
  const [token, setToken] = reactExports.useState(null);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  const API_URL = "http://localhost:3000";
  reactExports.useEffect(() => {
    const storedToken = localStorage.getItem("devsmith_token");
    if (storedToken) {
      setToken(storedToken);
      fetchCurrentUser(storedToken);
    } else {
      setLoading(false);
    }
  }, []);
  const fetchCurrentUser = async (authToken) => {
    try {
      const response = await fetch(`${API_URL}/api/portal/auth/me`, {
        headers: {
          "Authorization": `Bearer ${authToken}`,
          "Content-Type": "application/json"
        }
      });
      if (!response.ok) {
        console.log("User not authenticated (401) - clearing session");
        setUser(null);
        setToken(null);
        localStorage.removeItem("devsmith_token");
        setLoading(false);
        return;
      }
      const userData = await response.json();
      setUser(userData);
      setError(null);
    } catch (err) {
      console.error("Error fetching user:", err);
      setUser(null);
      setToken(null);
      localStorage.removeItem("devsmith_token");
    } finally {
      setLoading(false);
    }
  };
  const login = async (email, password) => {
    try {
      const response = await fetch(`${API_URL}/api/portal/auth/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({ email, password })
      });
      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || "Login failed");
      }
      const data = await response.json();
      const authToken = data.token;
      setToken(authToken);
      localStorage.setItem("devsmith_token", authToken);
      await fetchCurrentUser(authToken);
      return { success: true };
    } catch (err) {
      console.error("Login error:", err);
      setError(err.message);
      return { success: false, error: err.message };
    }
  };
  const logout = () => {
    setToken(null);
    setUser(null);
    localStorage.removeItem("devsmith_token");
    setError(null);
  };
  const value = {
    user,
    token,
    loading,
    error,
    login,
    logout,
    isAuthenticated: !!user
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsx(AuthContext.Provider, { value, children });
}
function useAuth() {
  const context = reactExports.useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
}
const ThemeContext$1 = reactExports.createContext();
function ThemeProvider({ children }) {
  const [isDarkMode, setIsDarkMode] = reactExports.useState(() => {
    const saved = localStorage.getItem("darkMode");
    return saved ? JSON.parse(saved) : false;
  });
  reactExports.useEffect(() => {
    localStorage.setItem("darkMode", JSON.stringify(isDarkMode));
    if (isDarkMode) {
      document.body.classList.add("dark-mode");
    } else {
      document.body.classList.remove("dark-mode");
    }
  }, [isDarkMode]);
  const toggleTheme = () => setIsDarkMode(!isDarkMode);
  return /* @__PURE__ */ jsxRuntimeExports.jsx(ThemeContext$1.Provider, { value: { isDarkMode, toggleTheme }, children });
}
function useTheme() {
  const context = reactExports.useContext(ThemeContext$1);
  if (!context) {
    throw new Error("useTheme must be used within ThemeProvider");
  }
  return context;
}
function Dashboard() {
  const { user, logout } = useAuth();
  const { isDarkMode, toggleTheme } = useTheme();
  const navigate = useNavigate();
  const handleLogout = () => {
    logout();
    navigate("/login");
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container mt-4", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("nav", { className: "navbar navbar-expand-lg navbar-light frosted-card mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "navbar-brand fw-bold", style: { fontSize: "1.5rem", color: isDarkMode ? "#e0e7ff" : "#1e293b" }, children: "DevSmith Platform" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center gap-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            onClick: toggleTheme,
            className: "theme-toggle",
            title: "Toggle dark/light mode",
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi ${isDarkMode ? "bi-sun-fill" : "bi-moon-fill"}` })
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "me-3", children: [
          "Welcome, ",
          (user == null ? void 0 : user.username) || (user == null ? void 0 : user.name),
          "!"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-outline-danger btn-sm",
            onClick: handleLogout,
            children: "Logout"
          }
        )
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h2", { className: "mb-3", children: "Dashboard" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", children: "Welcome to DevSmith Platform! Choose an application below to get started." })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-3 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx(Link, { to: "/health", className: "text-decoration-none", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-heart-pulse mb-3", style: { fontSize: "3.3rem", color: "#06b6d4" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-3", children: "Health" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { fontSize: "0.9rem" }, children: "System Health monitoring, service logs, and diagnostics" })
      ] }) }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-3 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx(Link, { to: "/review", className: "text-decoration-none", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-code-square mb-3", style: { fontSize: "3.3rem", color: "#8b5cf6" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-3", children: "Code Review" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { fontSize: "0.9rem" }, children: "AI-powered code review with five reading modes" })
      ] }) }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-3 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx(Link, { to: "/llm-config", className: "text-decoration-none", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-robot mb-3", style: { fontSize: "3.3rem", color: "#10b981" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-3", children: "AI Factory" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { fontSize: "0.9rem" }, children: "Configure AI models and API keys for each app" })
      ] }) }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-3 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx(Link, { to: "/projects", className: "text-decoration-none", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder2-open mb-3", style: { fontSize: "3.3rem", color: "#f59e0b" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-3", children: "Projects" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { fontSize: "0.9rem" }, children: "Manage cross-repo logging projects and API keys" })
      ] }) }) })
    ] })
  ] });
}
var classnames = { exports: {} };
/*!
	Copyright (c) 2018 Jed Watson.
	Licensed under the MIT License (MIT), see
	http://jedwatson.github.io/classnames
*/
(function(module) {
  (function() {
    var hasOwn = {}.hasOwnProperty;
    function classNames2() {
      var classes = "";
      for (var i = 0; i < arguments.length; i++) {
        var arg = arguments[i];
        if (arg) {
          classes = appendClass(classes, parseValue(arg));
        }
      }
      return classes;
    }
    function parseValue(arg) {
      if (typeof arg === "string" || typeof arg === "number") {
        return arg;
      }
      if (typeof arg !== "object") {
        return "";
      }
      if (Array.isArray(arg)) {
        return classNames2.apply(null, arg);
      }
      if (arg.toString !== Object.prototype.toString && !arg.toString.toString().includes("[native code]")) {
        return arg.toString();
      }
      var classes = "";
      for (var key in arg) {
        if (hasOwn.call(arg, key) && arg[key]) {
          classes = appendClass(classes, key);
        }
      }
      return classes;
    }
    function appendClass(value, newClass) {
      if (!newClass) {
        return value;
      }
      if (value) {
        return value + " " + newClass;
      }
      return value + newClass;
    }
    if (module.exports) {
      classNames2.default = classNames2;
      module.exports = classNames2;
    } else {
      window.classNames = classNames2;
    }
  })();
})(classnames);
var classnamesExports = classnames.exports;
const classNames = /* @__PURE__ */ getDefaultExportFromCjs(classnamesExports);
function _objectWithoutPropertiesLoose$4(r2, e) {
  if (null == r2) return {};
  var t2 = {};
  for (var n2 in r2) if ({}.hasOwnProperty.call(r2, n2)) {
    if (-1 !== e.indexOf(n2)) continue;
    t2[n2] = r2[n2];
  }
  return t2;
}
function _setPrototypeOf(t2, e) {
  return _setPrototypeOf = Object.setPrototypeOf ? Object.setPrototypeOf.bind() : function(t3, e2) {
    return t3.__proto__ = e2, t3;
  }, _setPrototypeOf(t2, e);
}
function _inheritsLoose(t2, o) {
  t2.prototype = Object.create(o.prototype), t2.prototype.constructor = t2, _setPrototypeOf(t2, o);
}
const DEFAULT_BREAKPOINTS = ["xxl", "xl", "lg", "md", "sm", "xs"];
const DEFAULT_MIN_BREAKPOINT = "xs";
const ThemeContext = /* @__PURE__ */ reactExports.createContext({
  prefixes: {},
  breakpoints: DEFAULT_BREAKPOINTS,
  minBreakpoint: DEFAULT_MIN_BREAKPOINT
});
const {
  Consumer,
  Provider
} = ThemeContext;
function useBootstrapPrefix(prefix, defaultPrefix) {
  const {
    prefixes
  } = reactExports.useContext(ThemeContext);
  return prefix || prefixes[defaultPrefix] || defaultPrefix;
}
function useIsRTL() {
  const {
    dir
  } = reactExports.useContext(ThemeContext);
  return dir === "rtl";
}
function ownerDocument(node) {
  return node && node.ownerDocument || document;
}
function ownerWindow(node) {
  var doc = ownerDocument(node);
  return doc && doc.defaultView || window;
}
function getComputedStyle(node, psuedoElement) {
  return ownerWindow(node).getComputedStyle(node, psuedoElement);
}
var rUpper = /([A-Z])/g;
function hyphenate(string) {
  return string.replace(rUpper, "-$1").toLowerCase();
}
var msPattern = /^ms-/;
function hyphenateStyleName(string) {
  return hyphenate(string).replace(msPattern, "-ms-");
}
var supportedTransforms = /^((translate|rotate|scale)(X|Y|Z|3d)?|matrix(3d)?|perspective|skew(X|Y)?)$/i;
function isTransform(value) {
  return !!(value && supportedTransforms.test(value));
}
function style(node, property) {
  var css = "";
  var transforms = "";
  if (typeof property === "string") {
    return node.style.getPropertyValue(hyphenateStyleName(property)) || getComputedStyle(node).getPropertyValue(hyphenateStyleName(property));
  }
  Object.keys(property).forEach(function(key) {
    var value = property[key];
    if (!value && value !== 0) {
      node.style.removeProperty(hyphenateStyleName(key));
    } else if (isTransform(key)) {
      transforms += key + "(" + value + ") ";
    } else {
      css += hyphenateStyleName(key) + ": " + value + ";";
    }
  });
  if (transforms) {
    css += "transform: " + transforms + ";";
  }
  node.style.cssText += ";" + css;
}
var propTypes$1 = { exports: {} };
var ReactPropTypesSecret$1 = "SECRET_DO_NOT_PASS_THIS_OR_YOU_WILL_BE_FIRED";
var ReactPropTypesSecret_1 = ReactPropTypesSecret$1;
var ReactPropTypesSecret = ReactPropTypesSecret_1;
function emptyFunction() {
}
function emptyFunctionWithReset() {
}
emptyFunctionWithReset.resetWarningCache = emptyFunction;
var factoryWithThrowingShims = function() {
  function shim(props, propName, componentName, location, propFullName, secret) {
    if (secret === ReactPropTypesSecret) {
      return;
    }
    var err = new Error(
      "Calling PropTypes validators directly is not supported by the `prop-types` package. Use PropTypes.checkPropTypes() to call them. Read more at http://fb.me/use-check-prop-types"
    );
    err.name = "Invariant Violation";
    throw err;
  }
  shim.isRequired = shim;
  function getShim() {
    return shim;
  }
  var ReactPropTypes = {
    array: shim,
    bigint: shim,
    bool: shim,
    func: shim,
    number: shim,
    object: shim,
    string: shim,
    symbol: shim,
    any: shim,
    arrayOf: getShim,
    element: shim,
    elementType: shim,
    instanceOf: getShim,
    node: shim,
    objectOf: getShim,
    oneOf: getShim,
    oneOfType: getShim,
    shape: getShim,
    exact: getShim,
    checkPropTypes: emptyFunctionWithReset,
    resetWarningCache: emptyFunction
  };
  ReactPropTypes.PropTypes = ReactPropTypes;
  return ReactPropTypes;
};
{
  propTypes$1.exports = factoryWithThrowingShims();
}
var propTypesExports = propTypes$1.exports;
const PropTypes = /* @__PURE__ */ getDefaultExportFromCjs(propTypesExports);
const config$2 = {
  disabled: false
};
const TransitionGroupContext = We$1.createContext(null);
var forceReflow = function forceReflow2(node) {
  return node.scrollTop;
};
var UNMOUNTED = "unmounted";
var EXITED = "exited";
var ENTERING = "entering";
var ENTERED = "entered";
var EXITING = "exiting";
var Transition = /* @__PURE__ */ function(_React$Component) {
  _inheritsLoose(Transition2, _React$Component);
  function Transition2(props, context) {
    var _this;
    _this = _React$Component.call(this, props, context) || this;
    var parentGroup = context;
    var appear = parentGroup && !parentGroup.isMounting ? props.enter : props.appear;
    var initialStatus;
    _this.appearStatus = null;
    if (props.in) {
      if (appear) {
        initialStatus = EXITED;
        _this.appearStatus = ENTERING;
      } else {
        initialStatus = ENTERED;
      }
    } else {
      if (props.unmountOnExit || props.mountOnEnter) {
        initialStatus = UNMOUNTED;
      } else {
        initialStatus = EXITED;
      }
    }
    _this.state = {
      status: initialStatus
    };
    _this.nextCallback = null;
    return _this;
  }
  Transition2.getDerivedStateFromProps = function getDerivedStateFromProps(_ref, prevState) {
    var nextIn = _ref.in;
    if (nextIn && prevState.status === UNMOUNTED) {
      return {
        status: EXITED
      };
    }
    return null;
  };
  var _proto = Transition2.prototype;
  _proto.componentDidMount = function componentDidMount() {
    this.updateStatus(true, this.appearStatus);
  };
  _proto.componentDidUpdate = function componentDidUpdate(prevProps) {
    var nextStatus = null;
    if (prevProps !== this.props) {
      var status = this.state.status;
      if (this.props.in) {
        if (status !== ENTERING && status !== ENTERED) {
          nextStatus = ENTERING;
        }
      } else {
        if (status === ENTERING || status === ENTERED) {
          nextStatus = EXITING;
        }
      }
    }
    this.updateStatus(false, nextStatus);
  };
  _proto.componentWillUnmount = function componentWillUnmount() {
    this.cancelNextCallback();
  };
  _proto.getTimeouts = function getTimeouts() {
    var timeout2 = this.props.timeout;
    var exit, enter, appear;
    exit = enter = appear = timeout2;
    if (timeout2 != null && typeof timeout2 !== "number") {
      exit = timeout2.exit;
      enter = timeout2.enter;
      appear = timeout2.appear !== void 0 ? timeout2.appear : enter;
    }
    return {
      exit,
      enter,
      appear
    };
  };
  _proto.updateStatus = function updateStatus(mounting, nextStatus) {
    if (mounting === void 0) {
      mounting = false;
    }
    if (nextStatus !== null) {
      this.cancelNextCallback();
      if (nextStatus === ENTERING) {
        if (this.props.unmountOnExit || this.props.mountOnEnter) {
          var node = this.props.nodeRef ? this.props.nodeRef.current : ReactDOM.findDOMNode(this);
          if (node) forceReflow(node);
        }
        this.performEnter(mounting);
      } else {
        this.performExit();
      }
    } else if (this.props.unmountOnExit && this.state.status === EXITED) {
      this.setState({
        status: UNMOUNTED
      });
    }
  };
  _proto.performEnter = function performEnter(mounting) {
    var _this2 = this;
    var enter = this.props.enter;
    var appearing = this.context ? this.context.isMounting : mounting;
    var _ref2 = this.props.nodeRef ? [appearing] : [ReactDOM.findDOMNode(this), appearing], maybeNode = _ref2[0], maybeAppearing = _ref2[1];
    var timeouts = this.getTimeouts();
    var enterTimeout = appearing ? timeouts.appear : timeouts.enter;
    if (!mounting && !enter || config$2.disabled) {
      this.safeSetState({
        status: ENTERED
      }, function() {
        _this2.props.onEntered(maybeNode);
      });
      return;
    }
    this.props.onEnter(maybeNode, maybeAppearing);
    this.safeSetState({
      status: ENTERING
    }, function() {
      _this2.props.onEntering(maybeNode, maybeAppearing);
      _this2.onTransitionEnd(enterTimeout, function() {
        _this2.safeSetState({
          status: ENTERED
        }, function() {
          _this2.props.onEntered(maybeNode, maybeAppearing);
        });
      });
    });
  };
  _proto.performExit = function performExit() {
    var _this3 = this;
    var exit = this.props.exit;
    var timeouts = this.getTimeouts();
    var maybeNode = this.props.nodeRef ? void 0 : ReactDOM.findDOMNode(this);
    if (!exit || config$2.disabled) {
      this.safeSetState({
        status: EXITED
      }, function() {
        _this3.props.onExited(maybeNode);
      });
      return;
    }
    this.props.onExit(maybeNode);
    this.safeSetState({
      status: EXITING
    }, function() {
      _this3.props.onExiting(maybeNode);
      _this3.onTransitionEnd(timeouts.exit, function() {
        _this3.safeSetState({
          status: EXITED
        }, function() {
          _this3.props.onExited(maybeNode);
        });
      });
    });
  };
  _proto.cancelNextCallback = function cancelNextCallback() {
    if (this.nextCallback !== null) {
      this.nextCallback.cancel();
      this.nextCallback = null;
    }
  };
  _proto.safeSetState = function safeSetState(nextState, callback) {
    callback = this.setNextCallback(callback);
    this.setState(nextState, callback);
  };
  _proto.setNextCallback = function setNextCallback(callback) {
    var _this4 = this;
    var active = true;
    this.nextCallback = function(event) {
      if (active) {
        active = false;
        _this4.nextCallback = null;
        callback(event);
      }
    };
    this.nextCallback.cancel = function() {
      active = false;
    };
    return this.nextCallback;
  };
  _proto.onTransitionEnd = function onTransitionEnd(timeout2, handler) {
    this.setNextCallback(handler);
    var node = this.props.nodeRef ? this.props.nodeRef.current : ReactDOM.findDOMNode(this);
    var doesNotHaveTimeoutOrListener = timeout2 == null && !this.props.addEndListener;
    if (!node || doesNotHaveTimeoutOrListener) {
      setTimeout(this.nextCallback, 0);
      return;
    }
    if (this.props.addEndListener) {
      var _ref3 = this.props.nodeRef ? [this.nextCallback] : [node, this.nextCallback], maybeNode = _ref3[0], maybeNextCallback = _ref3[1];
      this.props.addEndListener(maybeNode, maybeNextCallback);
    }
    if (timeout2 != null) {
      setTimeout(this.nextCallback, timeout2);
    }
  };
  _proto.render = function render() {
    var status = this.state.status;
    if (status === UNMOUNTED) {
      return null;
    }
    var _this$props = this.props, children = _this$props.children;
    _this$props.in;
    _this$props.mountOnEnter;
    _this$props.unmountOnExit;
    _this$props.appear;
    _this$props.enter;
    _this$props.exit;
    _this$props.timeout;
    _this$props.addEndListener;
    _this$props.onEnter;
    _this$props.onEntering;
    _this$props.onEntered;
    _this$props.onExit;
    _this$props.onExiting;
    _this$props.onExited;
    _this$props.nodeRef;
    var childProps = _objectWithoutPropertiesLoose$4(_this$props, ["children", "in", "mountOnEnter", "unmountOnExit", "appear", "enter", "exit", "timeout", "addEndListener", "onEnter", "onEntering", "onEntered", "onExit", "onExiting", "onExited", "nodeRef"]);
    return (
      // allows for nested Transitions
      /* @__PURE__ */ We$1.createElement(TransitionGroupContext.Provider, {
        value: null
      }, typeof children === "function" ? children(status, childProps) : We$1.cloneElement(We$1.Children.only(children), childProps))
    );
  };
  return Transition2;
}(We$1.Component);
Transition.contextType = TransitionGroupContext;
Transition.propTypes = {};
function noop() {
}
Transition.defaultProps = {
  in: false,
  mountOnEnter: false,
  unmountOnExit: false,
  appear: false,
  enter: true,
  exit: true,
  onEnter: noop,
  onEntering: noop,
  onEntered: noop,
  onExit: noop,
  onExiting: noop,
  onExited: noop
};
Transition.UNMOUNTED = UNMOUNTED;
Transition.EXITED = EXITED;
Transition.ENTERING = ENTERING;
Transition.ENTERED = ENTERED;
Transition.EXITING = EXITING;
function isEscKey(e) {
  return e.code === "Escape" || e.keyCode === 27;
}
function getReactVersion() {
  const parts = reactExports.version.split(".");
  return {
    major: +parts[0],
    minor: +parts[1],
    patch: +parts[2]
  };
}
function getChildRef(element) {
  if (!element || typeof element === "function") {
    return null;
  }
  const {
    major
  } = getReactVersion();
  const childRef = major >= 19 ? element.props.ref : element.ref;
  return childRef;
}
const canUseDOM = !!(typeof window !== "undefined" && window.document && window.document.createElement);
var optionsSupported = false;
var onceSupported = false;
try {
  var options = {
    get passive() {
      return optionsSupported = true;
    },
    get once() {
      return onceSupported = optionsSupported = true;
    }
  };
  if (canUseDOM) {
    window.addEventListener("test", options, options);
    window.removeEventListener("test", options, true);
  }
} catch (e) {
}
function addEventListener(node, eventName, handler, options) {
  if (options && typeof options !== "boolean" && !onceSupported) {
    var once = options.once, capture = options.capture;
    var wrappedHandler = handler;
    if (!onceSupported && once) {
      wrappedHandler = handler.__once || function onceHandler(event) {
        this.removeEventListener(eventName, onceHandler, capture);
        handler.call(this, event);
      };
      handler.__once = wrappedHandler;
    }
    node.addEventListener(eventName, wrappedHandler, optionsSupported ? options : capture);
  }
  node.addEventListener(eventName, handler, options);
}
function removeEventListener(node, eventName, handler, options) {
  var capture = options && typeof options !== "boolean" ? options.capture : options;
  node.removeEventListener(eventName, handler, capture);
  if (handler.__once) {
    node.removeEventListener(eventName, handler.__once, capture);
  }
}
function listen(node, eventName, handler, options) {
  addEventListener(node, eventName, handler, options);
  return function() {
    removeEventListener(node, eventName, handler, options);
  };
}
function triggerEvent(node, eventName, bubbles, cancelable) {
  if (cancelable === void 0) {
    cancelable = true;
  }
  if (node) {
    var event = document.createEvent("HTMLEvents");
    event.initEvent(eventName, bubbles, cancelable);
    node.dispatchEvent(event);
  }
}
function parseDuration$1(node) {
  var str = style(node, "transitionDuration") || "";
  var mult = str.indexOf("ms") === -1 ? 1e3 : 1;
  return parseFloat(str) * mult;
}
function emulateTransitionEnd(element, duration, padding) {
  if (padding === void 0) {
    padding = 5;
  }
  var called = false;
  var handle = setTimeout(function() {
    if (!called) triggerEvent(element, "transitionend", true);
  }, duration + padding);
  var remove = listen(element, "transitionend", function() {
    called = true;
  }, {
    once: true
  });
  return function() {
    clearTimeout(handle);
    remove();
  };
}
function transitionEnd(element, handler, duration, padding) {
  if (duration == null) duration = parseDuration$1(element) || 0;
  var removeEmulate = emulateTransitionEnd(element, duration, padding);
  var remove = listen(element, "transitionend", handler);
  return function() {
    removeEmulate();
    remove();
  };
}
function parseDuration(node, property) {
  const str = style(node, property) || "";
  const mult = str.indexOf("ms") === -1 ? 1e3 : 1;
  return parseFloat(str) * mult;
}
function transitionEndListener(element, handler) {
  const duration = parseDuration(element, "transitionDuration");
  const delay = parseDuration(element, "transitionDelay");
  const remove = transitionEnd(element, (e) => {
    if (e.target === element) {
      remove();
      handler(e);
    }
  }, duration + delay);
}
function triggerBrowserReflow(node) {
  node.offsetHeight;
}
const toFnRef$1 = (ref) => !ref || typeof ref === "function" ? ref : (value) => {
  ref.current = value;
};
function mergeRefs$1(refA, refB) {
  const a = toFnRef$1(refA);
  const b = toFnRef$1(refB);
  return (value) => {
    if (a) a(value);
    if (b) b(value);
  };
}
function useMergedRefs$1(refA, refB) {
  return reactExports.useMemo(() => mergeRefs$1(refA, refB), [refA, refB]);
}
function safeFindDOMNode(componentOrElement) {
  if (componentOrElement && "setState" in componentOrElement) {
    return ReactDOM.findDOMNode(componentOrElement);
  }
  return componentOrElement != null ? componentOrElement : null;
}
const TransitionWrapper = /* @__PURE__ */ We$1.forwardRef(({
  onEnter,
  onEntering,
  onEntered,
  onExit,
  onExiting,
  onExited,
  addEndListener,
  children,
  childRef,
  ...props
}, ref) => {
  const nodeRef = reactExports.useRef(null);
  const mergedRef = useMergedRefs$1(nodeRef, childRef);
  const attachRef = (r2) => {
    mergedRef(safeFindDOMNode(r2));
  };
  const normalize = (callback) => (param) => {
    if (callback && nodeRef.current) {
      callback(nodeRef.current, param);
    }
  };
  const handleEnter = reactExports.useCallback(normalize(onEnter), [onEnter]);
  const handleEntering = reactExports.useCallback(normalize(onEntering), [onEntering]);
  const handleEntered = reactExports.useCallback(normalize(onEntered), [onEntered]);
  const handleExit = reactExports.useCallback(normalize(onExit), [onExit]);
  const handleExiting = reactExports.useCallback(normalize(onExiting), [onExiting]);
  const handleExited = reactExports.useCallback(normalize(onExited), [onExited]);
  const handleAddEndListener = reactExports.useCallback(normalize(addEndListener), [addEndListener]);
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Transition, {
    ref,
    ...props,
    onEnter: handleEnter,
    onEntered: handleEntered,
    onEntering: handleEntering,
    onExit: handleExit,
    onExited: handleExited,
    onExiting: handleExiting,
    addEndListener: handleAddEndListener,
    nodeRef,
    children: typeof children === "function" ? (status, innerProps) => (
      // TODO: Types for RTG missing innerProps, so need to cast.
      children(status, {
        ...innerProps,
        ref: attachRef
      })
    ) : /* @__PURE__ */ We$1.cloneElement(children, {
      ref: attachRef
    })
  });
});
TransitionWrapper.displayName = "TransitionWrapper";
function useCommittedRef$1(value) {
  const ref = reactExports.useRef(value);
  reactExports.useEffect(() => {
    ref.current = value;
  }, [value]);
  return ref;
}
function useEventCallback$1(fn) {
  const ref = useCommittedRef$1(fn);
  return reactExports.useCallback(function(...args) {
    return ref.current && ref.current(...args);
  }, [ref]);
}
const divWithClassName = (className) => (
  // eslint-disable-next-line react/display-name
  /* @__PURE__ */ reactExports.forwardRef((p2, ref) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", {
    ...p2,
    ref,
    className: classNames(p2.className, className)
  }))
);
function useCommittedRef(value) {
  const ref = reactExports.useRef(value);
  reactExports.useEffect(() => {
    ref.current = value;
  }, [value]);
  return ref;
}
function useEventCallback(fn) {
  const ref = useCommittedRef(fn);
  return reactExports.useCallback(function(...args) {
    return ref.current && ref.current(...args);
  }, [ref]);
}
function useMounted() {
  const mounted = reactExports.useRef(true);
  const isMounted = reactExports.useRef(() => mounted.current);
  reactExports.useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  return isMounted.current;
}
function usePrevious(value) {
  const ref = reactExports.useRef(null);
  reactExports.useEffect(() => {
    ref.current = value;
  });
  return ref.current;
}
const isReactNative = typeof global !== "undefined" && // @ts-ignore
global.navigator && // @ts-ignore
global.navigator.product === "ReactNative";
const isDOM = typeof document !== "undefined";
const useIsomorphicEffect = isDOM || isReactNative ? reactExports.useLayoutEffect : reactExports.useEffect;
const fadeStyles = {
  [ENTERING]: "show",
  [ENTERED]: "show"
};
const Fade = /* @__PURE__ */ reactExports.forwardRef(({
  className,
  children,
  transitionClasses = {},
  onEnter,
  ...rest
}, ref) => {
  const props = {
    in: false,
    timeout: 300,
    mountOnEnter: false,
    unmountOnExit: false,
    appear: false,
    ...rest
  };
  const handleEnter = reactExports.useCallback((node, isAppearing) => {
    triggerBrowserReflow(node);
    onEnter == null || onEnter(node, isAppearing);
  }, [onEnter]);
  return /* @__PURE__ */ jsxRuntimeExports.jsx(TransitionWrapper, {
    ref,
    addEndListener: transitionEndListener,
    ...props,
    onEnter: handleEnter,
    childRef: getChildRef(children),
    children: (status, innerProps) => /* @__PURE__ */ reactExports.cloneElement(children, {
      ...innerProps,
      className: classNames("fade", className, children.props.className, fadeStyles[status], transitionClasses[status])
    })
  });
});
Fade.displayName = "Fade";
const propTypes = {
  /** An accessible label indicating the relevant information about the Close Button. */
  "aria-label": PropTypes.string,
  /** A callback fired after the Close Button is clicked. */
  onClick: PropTypes.func,
  /**
   * Render different color variant for the button.
   *
   * Omitting this will render the default dark color.
   */
  variant: PropTypes.oneOf(["white"])
};
const CloseButton = /* @__PURE__ */ reactExports.forwardRef(({
  className,
  variant,
  "aria-label": ariaLabel = "Close",
  ...props
}, ref) => /* @__PURE__ */ jsxRuntimeExports.jsx("button", {
  ref,
  type: "button",
  className: classNames("btn-close", variant && `btn-close-${variant}`, className),
  "aria-label": ariaLabel,
  ...props
}));
CloseButton.displayName = "CloseButton";
CloseButton.propTypes = propTypes;
function useUpdatedRef$1(value) {
  const valueRef = reactExports.useRef(value);
  valueRef.current = value;
  return valueRef;
}
function useWillUnmount$1(fn) {
  const onUnmount = useUpdatedRef$1(fn);
  reactExports.useEffect(() => () => onUnmount.current(), []);
}
var toArray = Function.prototype.bind.call(Function.prototype.call, [].slice);
function qsa(element, selector) {
  return toArray(element.querySelectorAll(selector));
}
function contains(context, node) {
  if (context.contains) return context.contains(node);
  if (context.compareDocumentPosition) return context === node || !!(context.compareDocumentPosition(node) & 16);
}
const ATTRIBUTE_PREFIX = `data-rr-ui-`;
function dataAttr(property) {
  return `${ATTRIBUTE_PREFIX}${property}`;
}
const Context = /* @__PURE__ */ reactExports.createContext(canUseDOM ? window : void 0);
Context.Provider;
function useWindow() {
  return reactExports.useContext(Context);
}
const toFnRef = (ref) => !ref || typeof ref === "function" ? ref : (value) => {
  ref.current = value;
};
function mergeRefs(refA, refB) {
  const a = toFnRef(refA);
  const b = toFnRef(refB);
  return (value) => {
    if (a) a(value);
    if (b) b(value);
  };
}
function useMergedRefs(refA, refB) {
  return reactExports.useMemo(() => mergeRefs(refA, refB), [refA, refB]);
}
var size;
function scrollbarSize(recalc) {
  if (!size && size !== 0 || recalc) {
    if (canUseDOM) {
      var scrollDiv = document.createElement("div");
      scrollDiv.style.position = "absolute";
      scrollDiv.style.top = "-9999px";
      scrollDiv.style.width = "50px";
      scrollDiv.style.height = "50px";
      scrollDiv.style.overflow = "scroll";
      document.body.appendChild(scrollDiv);
      size = scrollDiv.offsetWidth - scrollDiv.clientWidth;
      document.body.removeChild(scrollDiv);
    }
  }
  return size;
}
function useCallbackRef() {
  return reactExports.useState(null);
}
function activeElement(doc) {
  if (doc === void 0) {
    doc = ownerDocument();
  }
  try {
    var active = doc.activeElement;
    if (!active || !active.nodeName) return null;
    return active;
  } catch (e) {
    return doc.body;
  }
}
function useUpdatedRef(value) {
  const valueRef = reactExports.useRef(value);
  valueRef.current = value;
  return valueRef;
}
function useWillUnmount(fn) {
  const onUnmount = useUpdatedRef(fn);
  reactExports.useEffect(() => () => onUnmount.current(), []);
}
function getBodyScrollbarWidth(ownerDocument2 = document) {
  const window2 = ownerDocument2.defaultView;
  return Math.abs(window2.innerWidth - ownerDocument2.documentElement.clientWidth);
}
const OPEN_DATA_ATTRIBUTE = dataAttr("modal-open");
class ModalManager {
  constructor({
    ownerDocument: ownerDocument2,
    handleContainerOverflow = true,
    isRTL = false
  } = {}) {
    this.handleContainerOverflow = handleContainerOverflow;
    this.isRTL = isRTL;
    this.modals = [];
    this.ownerDocument = ownerDocument2;
  }
  getScrollbarWidth() {
    return getBodyScrollbarWidth(this.ownerDocument);
  }
  getElement() {
    return (this.ownerDocument || document).body;
  }
  setModalAttributes(_modal) {
  }
  removeModalAttributes(_modal) {
  }
  setContainerStyle(containerState) {
    const style$1 = {
      overflow: "hidden"
    };
    const paddingProp = this.isRTL ? "paddingLeft" : "paddingRight";
    const container = this.getElement();
    containerState.style = {
      overflow: container.style.overflow,
      [paddingProp]: container.style[paddingProp]
    };
    if (containerState.scrollBarWidth) {
      style$1[paddingProp] = `${parseInt(style(container, paddingProp) || "0", 10) + containerState.scrollBarWidth}px`;
    }
    container.setAttribute(OPEN_DATA_ATTRIBUTE, "");
    style(container, style$1);
  }
  reset() {
    [...this.modals].forEach((m2) => this.remove(m2));
  }
  removeContainerStyle(containerState) {
    const container = this.getElement();
    container.removeAttribute(OPEN_DATA_ATTRIBUTE);
    Object.assign(container.style, containerState.style);
  }
  add(modal) {
    let modalIdx = this.modals.indexOf(modal);
    if (modalIdx !== -1) {
      return modalIdx;
    }
    modalIdx = this.modals.length;
    this.modals.push(modal);
    this.setModalAttributes(modal);
    if (modalIdx !== 0) {
      return modalIdx;
    }
    this.state = {
      scrollBarWidth: this.getScrollbarWidth(),
      style: {}
    };
    if (this.handleContainerOverflow) {
      this.setContainerStyle(this.state);
    }
    return modalIdx;
  }
  remove(modal) {
    const modalIdx = this.modals.indexOf(modal);
    if (modalIdx === -1) {
      return;
    }
    this.modals.splice(modalIdx, 1);
    if (!this.modals.length && this.handleContainerOverflow) {
      this.removeContainerStyle(this.state);
    }
    this.removeModalAttributes(modal);
  }
  isTopModal(modal) {
    return !!this.modals.length && this.modals[this.modals.length - 1] === modal;
  }
}
const resolveContainerRef = (ref, document2) => {
  if (!canUseDOM) return null;
  if (ref == null) return (document2 || ownerDocument()).body;
  if (typeof ref === "function") ref = ref();
  if (ref && "current" in ref) ref = ref.current;
  if (ref && ("nodeType" in ref || ref.getBoundingClientRect)) return ref;
  return null;
};
function useWaitForDOMRef(ref, onResolved) {
  const window2 = useWindow();
  const [resolvedRef, setRef] = reactExports.useState(() => resolveContainerRef(ref, window2 == null ? void 0 : window2.document));
  if (!resolvedRef) {
    const earlyRef = resolveContainerRef(ref);
    if (earlyRef) setRef(earlyRef);
  }
  reactExports.useEffect(() => {
  }, [onResolved, resolvedRef]);
  reactExports.useEffect(() => {
    const nextRef = resolveContainerRef(ref);
    if (nextRef !== resolvedRef) {
      setRef(nextRef);
    }
  }, [ref, resolvedRef]);
  return resolvedRef;
}
function NoopTransition({
  children,
  in: inProp,
  onExited,
  mountOnEnter,
  unmountOnExit
}) {
  const ref = reactExports.useRef(null);
  const hasEnteredRef = reactExports.useRef(inProp);
  const handleExited = useEventCallback(onExited);
  reactExports.useEffect(() => {
    if (inProp) hasEnteredRef.current = true;
    else {
      handleExited(ref.current);
    }
  }, [inProp, handleExited]);
  const combinedRef = useMergedRefs(ref, getChildRef(children));
  const child = /* @__PURE__ */ reactExports.cloneElement(children, {
    ref: combinedRef
  });
  if (inProp) return child;
  if (unmountOnExit) {
    return null;
  }
  if (!hasEnteredRef.current && mountOnEnter) {
    return null;
  }
  return child;
}
const _excluded$3 = ["onEnter", "onEntering", "onEntered", "onExit", "onExiting", "onExited", "addEndListener", "children"];
function _objectWithoutPropertiesLoose$3(r2, e) {
  if (null == r2) return {};
  var t2 = {};
  for (var n2 in r2) if ({}.hasOwnProperty.call(r2, n2)) {
    if (e.indexOf(n2) >= 0) continue;
    t2[n2] = r2[n2];
  }
  return t2;
}
function useRTGTransitionProps(_ref) {
  let {
    onEnter,
    onEntering,
    onEntered,
    onExit,
    onExiting,
    onExited,
    addEndListener,
    children
  } = _ref, props = _objectWithoutPropertiesLoose$3(_ref, _excluded$3);
  const nodeRef = reactExports.useRef(null);
  const mergedRef = useMergedRefs(nodeRef, getChildRef(children));
  const normalize = (callback) => (param) => {
    if (callback && nodeRef.current) {
      callback(nodeRef.current, param);
    }
  };
  const handleEnter = reactExports.useCallback(normalize(onEnter), [onEnter]);
  const handleEntering = reactExports.useCallback(normalize(onEntering), [onEntering]);
  const handleEntered = reactExports.useCallback(normalize(onEntered), [onEntered]);
  const handleExit = reactExports.useCallback(normalize(onExit), [onExit]);
  const handleExiting = reactExports.useCallback(normalize(onExiting), [onExiting]);
  const handleExited = reactExports.useCallback(normalize(onExited), [onExited]);
  const handleAddEndListener = reactExports.useCallback(normalize(addEndListener), [addEndListener]);
  return Object.assign({}, props, {
    nodeRef
  }, onEnter && {
    onEnter: handleEnter
  }, onEntering && {
    onEntering: handleEntering
  }, onEntered && {
    onEntered: handleEntered
  }, onExit && {
    onExit: handleExit
  }, onExiting && {
    onExiting: handleExiting
  }, onExited && {
    onExited: handleExited
  }, addEndListener && {
    addEndListener: handleAddEndListener
  }, {
    children: typeof children === "function" ? (status, innerProps) => (
      // TODO: Types for RTG missing innerProps, so need to cast.
      children(status, Object.assign({}, innerProps, {
        ref: mergedRef
      }))
    ) : /* @__PURE__ */ reactExports.cloneElement(children, {
      ref: mergedRef
    })
  });
}
const _excluded$2 = ["component"];
function _objectWithoutPropertiesLoose$2(r2, e) {
  if (null == r2) return {};
  var t2 = {};
  for (var n2 in r2) if ({}.hasOwnProperty.call(r2, n2)) {
    if (e.indexOf(n2) >= 0) continue;
    t2[n2] = r2[n2];
  }
  return t2;
}
const RTGTransition = /* @__PURE__ */ reactExports.forwardRef((_ref, ref) => {
  let {
    component: Component
  } = _ref, props = _objectWithoutPropertiesLoose$2(_ref, _excluded$2);
  const transitionProps = useRTGTransitionProps(props);
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Component, Object.assign({
    ref
  }, transitionProps));
});
function useTransition({
  in: inProp,
  onTransition
}) {
  const ref = reactExports.useRef(null);
  const isInitialRef = reactExports.useRef(true);
  const handleTransition = useEventCallback(onTransition);
  useIsomorphicEffect(() => {
    if (!ref.current) {
      return void 0;
    }
    let stale = false;
    handleTransition({
      in: inProp,
      element: ref.current,
      initial: isInitialRef.current,
      isStale: () => stale
    });
    return () => {
      stale = true;
    };
  }, [inProp, handleTransition]);
  useIsomorphicEffect(() => {
    isInitialRef.current = false;
    return () => {
      isInitialRef.current = true;
    };
  }, []);
  return ref;
}
function ImperativeTransition({
  children,
  in: inProp,
  onExited,
  onEntered,
  transition
}) {
  const [exited, setExited] = reactExports.useState(!inProp);
  if (inProp && exited) {
    setExited(false);
  }
  const ref = useTransition({
    in: !!inProp,
    onTransition: (options) => {
      const onFinish = () => {
        if (options.isStale()) return;
        if (options.in) {
          onEntered == null ? void 0 : onEntered(options.element, options.initial);
        } else {
          setExited(true);
          onExited == null ? void 0 : onExited(options.element);
        }
      };
      Promise.resolve(transition(options)).then(onFinish, (error) => {
        if (!options.in) setExited(true);
        throw error;
      });
    }
  });
  const combinedRef = useMergedRefs(ref, getChildRef(children));
  return exited && !inProp ? null : /* @__PURE__ */ reactExports.cloneElement(children, {
    ref: combinedRef
  });
}
function renderTransition(component, runTransition, props) {
  if (component) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx(RTGTransition, Object.assign({}, props, {
      component
    }));
  }
  if (runTransition) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx(ImperativeTransition, Object.assign({}, props, {
      transition: runTransition
    }));
  }
  return /* @__PURE__ */ jsxRuntimeExports.jsx(NoopTransition, Object.assign({}, props));
}
const _excluded$1 = ["show", "role", "className", "style", "children", "backdrop", "keyboard", "onBackdropClick", "onEscapeKeyDown", "transition", "runTransition", "backdropTransition", "runBackdropTransition", "autoFocus", "enforceFocus", "restoreFocus", "restoreFocusOptions", "renderDialog", "renderBackdrop", "manager", "container", "onShow", "onHide", "onExit", "onExited", "onExiting", "onEnter", "onEntering", "onEntered"];
function _objectWithoutPropertiesLoose$1(r2, e) {
  if (null == r2) return {};
  var t2 = {};
  for (var n2 in r2) if ({}.hasOwnProperty.call(r2, n2)) {
    if (e.indexOf(n2) >= 0) continue;
    t2[n2] = r2[n2];
  }
  return t2;
}
let manager;
function getManager(window2) {
  if (!manager) manager = new ModalManager({
    ownerDocument: window2 == null ? void 0 : window2.document
  });
  return manager;
}
function useModalManager(provided) {
  const window2 = useWindow();
  const modalManager = provided || getManager(window2);
  const modal = reactExports.useRef({
    dialog: null,
    backdrop: null
  });
  return Object.assign(modal.current, {
    add: () => modalManager.add(modal.current),
    remove: () => modalManager.remove(modal.current),
    isTopModal: () => modalManager.isTopModal(modal.current),
    setDialogRef: reactExports.useCallback((ref) => {
      modal.current.dialog = ref;
    }, []),
    setBackdropRef: reactExports.useCallback((ref) => {
      modal.current.backdrop = ref;
    }, [])
  });
}
const Modal$2 = /* @__PURE__ */ reactExports.forwardRef((_ref, ref) => {
  let {
    show = false,
    role = "dialog",
    className,
    style: style2,
    children,
    backdrop = true,
    keyboard = true,
    onBackdropClick,
    onEscapeKeyDown,
    transition,
    runTransition,
    backdropTransition,
    runBackdropTransition,
    autoFocus = true,
    enforceFocus = true,
    restoreFocus = true,
    restoreFocusOptions,
    renderDialog,
    renderBackdrop = (props) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", Object.assign({}, props)),
    manager: providedManager,
    container: containerRef,
    onShow,
    onHide = () => {
    },
    onExit,
    onExited,
    onExiting,
    onEnter,
    onEntering,
    onEntered
  } = _ref, rest = _objectWithoutPropertiesLoose$1(_ref, _excluded$1);
  const ownerWindow2 = useWindow();
  const container = useWaitForDOMRef(containerRef);
  const modal = useModalManager(providedManager);
  const isMounted = useMounted();
  const prevShow = usePrevious(show);
  const [exited, setExited] = reactExports.useState(!show);
  const lastFocusRef = reactExports.useRef(null);
  reactExports.useImperativeHandle(ref, () => modal, [modal]);
  if (canUseDOM && !prevShow && show) {
    lastFocusRef.current = activeElement(ownerWindow2 == null ? void 0 : ownerWindow2.document);
  }
  if (show && exited) {
    setExited(false);
  }
  const handleShow = useEventCallback(() => {
    modal.add();
    removeKeydownListenerRef.current = listen(document, "keydown", handleDocumentKeyDown);
    removeFocusListenerRef.current = listen(
      document,
      "focus",
      // the timeout is necessary b/c this will run before the new modal is mounted
      // and so steals focus from it
      () => setTimeout(handleEnforceFocus),
      true
    );
    if (onShow) {
      onShow();
    }
    if (autoFocus) {
      var _modal$dialog$ownerDo, _modal$dialog;
      const currentActiveElement = activeElement((_modal$dialog$ownerDo = (_modal$dialog = modal.dialog) == null ? void 0 : _modal$dialog.ownerDocument) != null ? _modal$dialog$ownerDo : ownerWindow2 == null ? void 0 : ownerWindow2.document);
      if (modal.dialog && currentActiveElement && !contains(modal.dialog, currentActiveElement)) {
        lastFocusRef.current = currentActiveElement;
        modal.dialog.focus();
      }
    }
  });
  const handleHide = useEventCallback(() => {
    modal.remove();
    removeKeydownListenerRef.current == null ? void 0 : removeKeydownListenerRef.current();
    removeFocusListenerRef.current == null ? void 0 : removeFocusListenerRef.current();
    if (restoreFocus) {
      var _lastFocusRef$current;
      (_lastFocusRef$current = lastFocusRef.current) == null ? void 0 : _lastFocusRef$current.focus == null ? void 0 : _lastFocusRef$current.focus(restoreFocusOptions);
      lastFocusRef.current = null;
    }
  });
  reactExports.useEffect(() => {
    if (!show || !container) return;
    handleShow();
  }, [
    show,
    container,
    /* should never change: */
    handleShow
  ]);
  reactExports.useEffect(() => {
    if (!exited) return;
    handleHide();
  }, [exited, handleHide]);
  useWillUnmount(() => {
    handleHide();
  });
  const handleEnforceFocus = useEventCallback(() => {
    if (!enforceFocus || !isMounted() || !modal.isTopModal()) {
      return;
    }
    const currentActiveElement = activeElement(ownerWindow2 == null ? void 0 : ownerWindow2.document);
    if (modal.dialog && currentActiveElement && !contains(modal.dialog, currentActiveElement)) {
      modal.dialog.focus();
    }
  });
  const handleBackdropClick = useEventCallback((e) => {
    if (e.target !== e.currentTarget) {
      return;
    }
    onBackdropClick == null ? void 0 : onBackdropClick(e);
    if (backdrop === true) {
      onHide();
    }
  });
  const handleDocumentKeyDown = useEventCallback((e) => {
    if (keyboard && isEscKey(e) && modal.isTopModal()) {
      onEscapeKeyDown == null ? void 0 : onEscapeKeyDown(e);
      if (!e.defaultPrevented) {
        onHide();
      }
    }
  });
  const removeFocusListenerRef = reactExports.useRef();
  const removeKeydownListenerRef = reactExports.useRef();
  const handleHidden = (...args) => {
    setExited(true);
    onExited == null ? void 0 : onExited(...args);
  };
  if (!container) {
    return null;
  }
  const dialogProps = Object.assign({
    role,
    ref: modal.setDialogRef,
    // apparently only works on the dialog role element
    "aria-modal": role === "dialog" ? true : void 0
  }, rest, {
    style: style2,
    className,
    tabIndex: -1
  });
  let dialog = renderDialog ? renderDialog(dialogProps) : /* @__PURE__ */ jsxRuntimeExports.jsx("div", Object.assign({}, dialogProps, {
    children: /* @__PURE__ */ reactExports.cloneElement(children, {
      role: "document"
    })
  }));
  dialog = renderTransition(transition, runTransition, {
    unmountOnExit: true,
    mountOnEnter: true,
    appear: true,
    in: !!show,
    onExit,
    onExiting,
    onExited: handleHidden,
    onEnter,
    onEntering,
    onEntered,
    children: dialog
  });
  let backdropElement = null;
  if (backdrop) {
    backdropElement = renderBackdrop({
      ref: modal.setBackdropRef,
      onClick: handleBackdropClick
    });
    backdropElement = renderTransition(backdropTransition, runBackdropTransition, {
      in: !!show,
      appear: true,
      mountOnEnter: true,
      unmountOnExit: true,
      children: backdropElement
    });
  }
  return /* @__PURE__ */ jsxRuntimeExports.jsx(jsxRuntimeExports.Fragment, {
    children: /* @__PURE__ */ ReactDOM.createPortal(/* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, {
      children: [backdropElement, dialog]
    }), container)
  });
});
Modal$2.displayName = "Modal";
const BaseModal = Object.assign(Modal$2, {
  Manager: ModalManager
});
function hasClass(element, className) {
  if (element.classList) return element.classList.contains(className);
  return (" " + (element.className.baseVal || element.className) + " ").indexOf(" " + className + " ") !== -1;
}
function addClass(element, className) {
  if (element.classList) element.classList.add(className);
  else if (!hasClass(element, className)) if (typeof element.className === "string") element.className = element.className + " " + className;
  else element.setAttribute("class", (element.className && element.className.baseVal || "") + " " + className);
}
function replaceClassName(origClass, classToRemove) {
  return origClass.replace(new RegExp("(^|\\s)" + classToRemove + "(?:\\s|$)", "g"), "$1").replace(/\s+/g, " ").replace(/^\s*|\s*$/g, "");
}
function removeClass(element, className) {
  if (element.classList) {
    element.classList.remove(className);
  } else if (typeof element.className === "string") {
    element.className = replaceClassName(element.className, className);
  } else {
    element.setAttribute("class", replaceClassName(element.className && element.className.baseVal || "", className));
  }
}
const Selector = {
  FIXED_CONTENT: ".fixed-top, .fixed-bottom, .is-fixed, .sticky-top",
  STICKY_CONTENT: ".sticky-top",
  NAVBAR_TOGGLER: ".navbar-toggler"
};
class BootstrapModalManager extends ModalManager {
  adjustAndStore(prop, element, adjust) {
    const actual = element.style[prop];
    element.dataset[prop] = actual;
    style(element, {
      [prop]: `${parseFloat(style(element, prop)) + adjust}px`
    });
  }
  restore(prop, element) {
    const value = element.dataset[prop];
    if (value !== void 0) {
      delete element.dataset[prop];
      style(element, {
        [prop]: value
      });
    }
  }
  setContainerStyle(containerState) {
    super.setContainerStyle(containerState);
    const container = this.getElement();
    addClass(container, "modal-open");
    if (!containerState.scrollBarWidth) return;
    const paddingProp = this.isRTL ? "paddingLeft" : "paddingRight";
    const marginProp = this.isRTL ? "marginLeft" : "marginRight";
    qsa(container, Selector.FIXED_CONTENT).forEach((el2) => this.adjustAndStore(paddingProp, el2, containerState.scrollBarWidth));
    qsa(container, Selector.STICKY_CONTENT).forEach((el2) => this.adjustAndStore(marginProp, el2, -containerState.scrollBarWidth));
    qsa(container, Selector.NAVBAR_TOGGLER).forEach((el2) => this.adjustAndStore(marginProp, el2, containerState.scrollBarWidth));
  }
  removeContainerStyle(containerState) {
    super.removeContainerStyle(containerState);
    const container = this.getElement();
    removeClass(container, "modal-open");
    const paddingProp = this.isRTL ? "paddingLeft" : "paddingRight";
    const marginProp = this.isRTL ? "marginLeft" : "marginRight";
    qsa(container, Selector.FIXED_CONTENT).forEach((el2) => this.restore(paddingProp, el2));
    qsa(container, Selector.STICKY_CONTENT).forEach((el2) => this.restore(marginProp, el2));
    qsa(container, Selector.NAVBAR_TOGGLER).forEach((el2) => this.restore(marginProp, el2));
  }
}
let sharedManager;
function getSharedManager(options) {
  if (!sharedManager) sharedManager = new BootstrapModalManager(options);
  return sharedManager;
}
const ModalBody = /* @__PURE__ */ reactExports.forwardRef(({
  className,
  bsPrefix,
  as: Component = "div",
  ...props
}, ref) => {
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal-body");
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Component, {
    ref,
    className: classNames(className, bsPrefix),
    ...props
  });
});
ModalBody.displayName = "ModalBody";
const ModalContext = /* @__PURE__ */ reactExports.createContext({
  onHide() {
  }
});
const ModalDialog = /* @__PURE__ */ reactExports.forwardRef(({
  bsPrefix,
  className,
  contentClassName,
  centered,
  size: size2,
  fullscreen,
  children,
  scrollable,
  ...props
}, ref) => {
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal");
  const dialogClass = `${bsPrefix}-dialog`;
  const fullScreenClass = typeof fullscreen === "string" ? `${bsPrefix}-fullscreen-${fullscreen}` : `${bsPrefix}-fullscreen`;
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", {
    ...props,
    ref,
    className: classNames(dialogClass, className, size2 && `${bsPrefix}-${size2}`, centered && `${dialogClass}-centered`, scrollable && `${dialogClass}-scrollable`, fullscreen && fullScreenClass),
    children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", {
      className: classNames(`${bsPrefix}-content`, contentClassName),
      children
    })
  });
});
ModalDialog.displayName = "ModalDialog";
const ModalFooter = /* @__PURE__ */ reactExports.forwardRef(({
  className,
  bsPrefix,
  as: Component = "div",
  ...props
}, ref) => {
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal-footer");
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Component, {
    ref,
    className: classNames(className, bsPrefix),
    ...props
  });
});
ModalFooter.displayName = "ModalFooter";
const AbstractModalHeader = /* @__PURE__ */ reactExports.forwardRef(({
  closeLabel = "Close",
  closeVariant,
  closeButton = false,
  onHide,
  children,
  ...props
}, ref) => {
  const context = reactExports.useContext(ModalContext);
  const handleClick = useEventCallback$1(() => {
    context == null || context.onHide();
    onHide == null || onHide();
  });
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", {
    ref,
    ...props,
    children: [children, closeButton && /* @__PURE__ */ jsxRuntimeExports.jsx(CloseButton, {
      "aria-label": closeLabel,
      variant: closeVariant,
      onClick: handleClick
    })]
  });
});
AbstractModalHeader.displayName = "AbstractModalHeader";
const ModalHeader = /* @__PURE__ */ reactExports.forwardRef(({
  bsPrefix,
  className,
  closeLabel = "Close",
  closeButton = false,
  ...props
}, ref) => {
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal-header");
  return /* @__PURE__ */ jsxRuntimeExports.jsx(AbstractModalHeader, {
    ref,
    ...props,
    className: classNames(className, bsPrefix),
    closeLabel,
    closeButton
  });
});
ModalHeader.displayName = "ModalHeader";
const DivStyledAsH4 = divWithClassName("h4");
const ModalTitle = /* @__PURE__ */ reactExports.forwardRef(({
  className,
  bsPrefix,
  as: Component = DivStyledAsH4,
  ...props
}, ref) => {
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal-title");
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Component, {
    ref,
    className: classNames(className, bsPrefix),
    ...props
  });
});
ModalTitle.displayName = "ModalTitle";
function DialogTransition(props) {
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Fade, {
    ...props,
    timeout: null
  });
}
function BackdropTransition(props) {
  return /* @__PURE__ */ jsxRuntimeExports.jsx(Fade, {
    ...props,
    timeout: null
  });
}
const Modal = /* @__PURE__ */ reactExports.forwardRef(({
  bsPrefix,
  className,
  style: style2,
  dialogClassName,
  contentClassName,
  children,
  dialogAs: Dialog = ModalDialog,
  "data-bs-theme": dataBsTheme,
  "aria-labelledby": ariaLabelledby,
  "aria-describedby": ariaDescribedby,
  "aria-label": ariaLabel,
  /* BaseModal props */
  show = false,
  animation = true,
  backdrop = true,
  keyboard = true,
  onEscapeKeyDown,
  onShow,
  onHide,
  container,
  autoFocus = true,
  enforceFocus = true,
  restoreFocus = true,
  restoreFocusOptions,
  onEntered,
  onExit,
  onExiting,
  onEnter,
  onEntering,
  onExited,
  backdropClassName,
  manager: propsManager,
  ...props
}, ref) => {
  const [modalStyle, setStyle] = reactExports.useState({});
  const [animateStaticModal, setAnimateStaticModal] = reactExports.useState(false);
  const waitingForMouseUpRef = reactExports.useRef(false);
  const ignoreBackdropClickRef = reactExports.useRef(false);
  const removeStaticModalAnimationRef = reactExports.useRef(null);
  const [modal, setModalRef] = useCallbackRef();
  const mergedRef = useMergedRefs$1(ref, setModalRef);
  const handleHide = useEventCallback$1(onHide);
  const isRTL = useIsRTL();
  bsPrefix = useBootstrapPrefix(bsPrefix, "modal");
  const modalContext = reactExports.useMemo(() => ({
    onHide: handleHide
  }), [handleHide]);
  function getModalManager() {
    if (propsManager) return propsManager;
    return getSharedManager({
      isRTL
    });
  }
  function updateDialogStyle(node) {
    if (!canUseDOM) return;
    const containerIsOverflowing = getModalManager().getScrollbarWidth() > 0;
    const modalIsOverflowing = node.scrollHeight > ownerDocument(node).documentElement.clientHeight;
    setStyle({
      paddingRight: containerIsOverflowing && !modalIsOverflowing ? scrollbarSize() : void 0,
      paddingLeft: !containerIsOverflowing && modalIsOverflowing ? scrollbarSize() : void 0
    });
  }
  const handleWindowResize = useEventCallback$1(() => {
    if (modal) {
      updateDialogStyle(modal.dialog);
    }
  });
  useWillUnmount$1(() => {
    removeEventListener(window, "resize", handleWindowResize);
    removeStaticModalAnimationRef.current == null || removeStaticModalAnimationRef.current();
  });
  const handleDialogMouseDown = () => {
    waitingForMouseUpRef.current = true;
  };
  const handleMouseUp = (e) => {
    if (waitingForMouseUpRef.current && modal && e.target === modal.dialog) {
      ignoreBackdropClickRef.current = true;
    }
    waitingForMouseUpRef.current = false;
  };
  const handleStaticModalAnimation = () => {
    setAnimateStaticModal(true);
    removeStaticModalAnimationRef.current = transitionEnd(modal.dialog, () => {
      setAnimateStaticModal(false);
    });
  };
  const handleStaticBackdropClick = (e) => {
    if (e.target !== e.currentTarget) {
      return;
    }
    handleStaticModalAnimation();
  };
  const handleClick = (e) => {
    if (backdrop === "static") {
      handleStaticBackdropClick(e);
      return;
    }
    if (ignoreBackdropClickRef.current || e.target !== e.currentTarget) {
      ignoreBackdropClickRef.current = false;
      return;
    }
    onHide == null || onHide();
  };
  const handleEscapeKeyDown = (e) => {
    if (keyboard) {
      onEscapeKeyDown == null || onEscapeKeyDown(e);
    } else {
      e.preventDefault();
      if (backdrop === "static") {
        handleStaticModalAnimation();
      }
    }
  };
  const handleEnter = (node, isAppearing) => {
    if (node) {
      updateDialogStyle(node);
    }
    onEnter == null || onEnter(node, isAppearing);
  };
  const handleExit = (node) => {
    removeStaticModalAnimationRef.current == null || removeStaticModalAnimationRef.current();
    onExit == null || onExit(node);
  };
  const handleEntering = (node, isAppearing) => {
    onEntering == null || onEntering(node, isAppearing);
    addEventListener(window, "resize", handleWindowResize);
  };
  const handleExited = (node) => {
    if (node) node.style.display = "";
    onExited == null || onExited(node);
    removeEventListener(window, "resize", handleWindowResize);
  };
  const renderBackdrop = reactExports.useCallback((backdropProps) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", {
    ...backdropProps,
    className: classNames(`${bsPrefix}-backdrop`, backdropClassName, !animation && "show")
  }), [animation, backdropClassName, bsPrefix]);
  const baseModalStyle = {
    ...style2,
    ...modalStyle
  };
  baseModalStyle.display = "block";
  const renderDialog = (dialogProps) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", {
    role: "dialog",
    ...dialogProps,
    style: baseModalStyle,
    className: classNames(className, bsPrefix, animateStaticModal && `${bsPrefix}-static`, !animation && "show"),
    onClick: backdrop ? handleClick : void 0,
    onMouseUp: handleMouseUp,
    "data-bs-theme": dataBsTheme,
    "aria-label": ariaLabel,
    "aria-labelledby": ariaLabelledby,
    "aria-describedby": ariaDescribedby,
    children: /* @__PURE__ */ jsxRuntimeExports.jsx(Dialog, {
      ...props,
      onMouseDown: handleDialogMouseDown,
      className: dialogClassName,
      contentClassName,
      children
    })
  });
  return /* @__PURE__ */ jsxRuntimeExports.jsx(ModalContext.Provider, {
    value: modalContext,
    children: /* @__PURE__ */ jsxRuntimeExports.jsx(BaseModal, {
      show,
      ref: mergedRef,
      backdrop,
      container,
      keyboard: true,
      autoFocus,
      enforceFocus,
      restoreFocus,
      restoreFocusOptions,
      onEscapeKeyDown: handleEscapeKeyDown,
      onShow,
      onHide,
      onEnter: handleEnter,
      onEntering: handleEntering,
      onEntered,
      onExit: handleExit,
      onExiting,
      onExited: handleExited,
      manager: getModalManager(),
      transition: animation ? DialogTransition : void 0,
      backdropTransition: animation ? BackdropTransition : void 0,
      renderBackdrop,
      renderDialog
    })
  });
});
Modal.displayName = "Modal";
const Modal$1 = Object.assign(Modal, {
  Body: ModalBody,
  Header: ModalHeader,
  Title: ModalTitle,
  Footer: ModalFooter,
  Dialog: ModalDialog,
  TRANSITION_DURATION: 300,
  BACKDROP_TRANSITION_DURATION: 150
});
const LEVEL_CONFIG = {
  debug: {
    icon: "bi-bug-fill",
    color: "success",
    bgColor: "rgba(25, 135, 84, 0.1)",
    label: "DEBUG"
  },
  info: {
    icon: "bi-info-circle-fill",
    color: "primary",
    bgColor: "rgba(13, 110, 253, 0.1)",
    label: "INFO"
  },
  warning: {
    icon: "bi-exclamation-triangle-fill",
    color: "warning",
    bgColor: "rgba(255, 193, 7, 0.1)",
    label: "WARNING"
  },
  error: {
    icon: "bi-x-circle-fill",
    color: "danger",
    bgColor: "rgba(220, 53, 69, 0.1)",
    label: "ERROR"
  },
  critical: {
    icon: "bi-fire",
    color: "danger",
    bgColor: "rgba(220, 53, 69, 0.2)",
    label: "CRITICAL"
  }
};
function StatCard({ level, count, isActive, onClick }) {
  const config2 = LEVEL_CONFIG[level.toLowerCase()] || LEVEL_CONFIG.info;
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
    "div",
    {
      className: `frosted-card h-100 ${isActive ? "border border-2" : ""}`,
      style: {
        backgroundColor: config2.bgColor,
        borderColor: isActive ? `var(--bs-${config2.color})` : "transparent",
        cursor: "pointer",
        transition: "all 0.2s ease",
        transform: isActive ? "scale(1.05)" : "scale(1)"
      },
      onClick,
      role: "button",
      tabIndex: 0,
      onKeyPress: (e) => {
        if (e.key === "Enter" || e.key === " ") {
          onClick();
        }
      },
      children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center p-4", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi ${config2.icon} text-${config2.color}`, style: { fontSize: "2.5rem" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3 mb-0", children: count.toLocaleString() }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: `text-${config2.color} fw-bold mb-0`, children: config2.label }),
        isActive && /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "d-block mt-2 text-muted", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-funnel-fill me-1" }),
          "Filtered"
        ] })
      ] })
    }
  ) });
}
function StatCards({ stats, selectedLevel, onLevelClick }) {
  const levels = ["debug", "info", "warning", "error", "critical"];
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: levels.map((level) => /* @__PURE__ */ jsxRuntimeExports.jsx(
    StatCard,
    {
      level,
      count: stats[level] || 0,
      isActive: selectedLevel === level,
      onClick: () => onLevelClick(level)
    },
    level
  )) });
}
const API_BASE_URL = "http://localhost:3000";
class ApiError extends Error {
  constructor(message, status) {
    super(message);
    this.name = "ApiError";
    this.status = status;
  }
}
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const { timeout, ...fetchOptions } = options;
  const defaultOptions = {
    headers: {
      "Content-Type": "application/json"
    },
    credentials: "include"
    // Include cookies for session auth
  };
  const controller = new AbortController();
  const signal = controller.signal;
  let timeoutId;
  try {
    if (timeout) {
      timeoutId = setTimeout(() => {
        controller.abort();
      }, timeout);
    }
    const response = await fetch(url, {
      ...defaultOptions,
      ...fetchOptions,
      signal
      // Pass abort signal to fetch
    });
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    if (!response.ok) {
      const errorText = await response.text();
      throw new ApiError(`HTTP ${response.status}: ${errorText}`, response.status);
    }
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      return response.json();
    }
    return response.text();
  } catch (error) {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    if (error.name === "AbortError") {
      throw new ApiError(`Request timeout after ${timeout}ms`, 408);
    }
    throw error;
  }
}
const reviewApi = {
  // Get available models
  getModels: () => apiRequest("/api/review/models"),
  // Create new review session
  createSession: (data) => apiRequest("/api/review/sessions", {
    method: "POST",
    body: JSON.stringify(data)
  }),
  // Run analysis in different modes
  // All AI analysis requests have 60-second timeout to prevent browser hangs
  runPreview: (sessionId, code, model, userMode = "intermediate", outputMode = "quick") => apiRequest("/api/review/modes/preview", {
    method: "POST",
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
    timeout: 6e4
    // 60 second timeout
  }),
  runSkim: (sessionId, code, model, userMode = "intermediate", outputMode = "quick") => apiRequest("/api/review/modes/skim", {
    method: "POST",
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
    timeout: 6e4
    // 60 second timeout
  }),
  runScan: (sessionId, code, model, query, userMode = "intermediate", outputMode = "quick") => apiRequest("/api/review/modes/scan", {
    method: "POST",
    body: JSON.stringify({ pasted_code: code, model, query, user_mode: userMode, output_mode: outputMode }),
    timeout: 6e4
    // 60 second timeout
  }),
  runDetailed: (sessionId, code, model, userMode = "intermediate", outputMode = "quick") => apiRequest("/api/review/modes/detailed", {
    method: "POST",
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
    timeout: 6e4
    // 60 second timeout
  }),
  runCritical: (sessionId, code, model, userMode = "intermediate", outputMode = "quick") => apiRequest("/api/review/modes/critical", {
    method: "POST",
    body: JSON.stringify({ pasted_code: code, model, user_mode: userMode, output_mode: outputMode }),
    timeout: 6e4
    // 60 second timeout
  }),
  // GitHub Integration API endpoints (Phase 1)
  // Fetch repository tree structure
  githubGetTree: (url, branch = "main") => {
    const params = new URLSearchParams({ url, branch });
    return apiRequest(`/api/review/github/tree?${params.toString()}`);
  },
  // Fetch individual file content
  githubGetFile: (url, path, branch = "main") => {
    const params = new URLSearchParams({ url, path, branch });
    return apiRequest(`/api/review/github/file?${params.toString()}`);
  },
  // Quick repo scan (fetches 5-8 core files)
  githubQuickScan: (url, branch = "main") => {
    const params = new URLSearchParams({ url, branch });
    return apiRequest(`/api/review/github/quick-scan?${params.toString()}`);
  },
  // Prompt Management API endpoints (Phase 4)
  // Get effective prompt (user custom or system default)
  getPrompt: (mode, userLevel = "intermediate", outputMode = "quick") => {
    const params = new URLSearchParams({ mode, user_level: userLevel, output_mode: outputMode });
    return apiRequest(`/api/review/prompts?${params.toString()}`);
  },
  // Save custom prompt
  savePrompt: (data) => apiRequest("/api/review/prompts", {
    method: "PUT",
    body: JSON.stringify(data)
  }),
  // Factory reset to system default
  resetPrompt: (mode, userLevel = "intermediate", outputMode = "quick") => {
    const params = new URLSearchParams({ mode, user_level: userLevel, output_mode: outputMode });
    return apiRequest(`/api/review/prompts?${params.toString()}`, {
      method: "DELETE"
    });
  },
  // Get prompt execution history
  getPromptHistory: (limit = 50) => {
    const params = new URLSearchParams({ limit: limit.toString() });
    return apiRequest(`/api/review/prompts/history?${params.toString()}`);
  },
  // Rate prompt execution
  rateExecution: (executionId, rating) => apiRequest(`/api/review/prompts/${executionId}/rate`, {
    method: "POST",
    body: JSON.stringify({ rating })
  })
};
function ModelSelector({ selectedModel, onModelSelect, disabled = false }) {
  const [models, setModels] = reactExports.useState([]);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  const { isDarkMode } = useTheme();
  reactExports.useEffect(() => {
    const loadModels = async () => {
      try {
        setLoading(true);
        const response = await apiRequest("/api/portal/llm-configs");
        console.log("LLM configs response:", response);
        let modelList = [];
        if (Array.isArray(response)) {
          modelList = response.map((config2) => ({
            name: config2.model || config2.model_name,
            // Backend uses "model", fallback to "model_name"
            displayName: config2.name || config2.display_name || config2.model || config2.model_name,
            // Backend computes "name" field
            provider: config2.provider,
            isDefault: config2.is_default
          }));
          modelList.sort((a, b) => {
            if (a.isDefault && !b.isDefault) return -1;
            if (!a.isDefault && b.isDefault) return 1;
            return a.displayName.localeCompare(b.displayName);
          });
        } else {
          console.warn("Unexpected LLM configs response format:", response);
        }
        setModels(modelList);
        console.log("Models loaded from Portal:", modelList);
        if (!selectedModel && modelList.length > 0) {
          const defaultModel = modelList.find((m2) => m2.isDefault);
          if (defaultModel) {
            console.log("Setting default model:", defaultModel.name);
            onModelSelect(defaultModel.name);
          } else {
            console.log("No default model found, using first:", modelList[0].name);
            onModelSelect(modelList[0].name);
          }
        }
      } catch (err) {
        console.error("Failed to load Portal LLM configs:", err);
        setError(err.message);
        setModels([]);
      } finally {
        setLoading(false);
      }
    };
    loadModels();
  }, [selectedModel, onModelSelect]);
  const handleModelChange = (e) => {
    console.log("Model selected:", e.target.value);
    if (onModelSelect) {
      onModelSelect(e.target.value);
    }
  };
  if (loading) {
    return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "model-selector mb-3", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "form-label", children: "AI Model:" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border spinner-border-sm me-2", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading models..." }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted", children: "Loading available models..." })
      ] })
    ] });
  }
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "model-selector mb-3", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "form-label", htmlFor: "model-select", children: "AI Model:" }),
    error && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "alert alert-warning py-1 mb-2", role: "alert", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Warning:" }),
      " ",
      error,
      ". Using fallback models."
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      "select",
      {
        id: "model-select",
        className: `form-select ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`,
        style: isDarkMode ? {
          backgroundColor: "#1a1d2e",
          color: "#e0e7ff",
          borderColor: "#4a5568"
        } : {},
        value: selectedModel || "",
        onChange: handleModelChange,
        disabled: disabled || models.length === 0,
        children: models.length === 0 ? /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "", children: "No models available" }) : models.map((model) => {
          const modelName = typeof model === "string" ? model : model.name;
          const displayName = typeof model === "object" && model.displayName ? model.displayName : modelName;
          const provider = typeof model === "object" && model.provider ? model.provider : "";
          return /* @__PURE__ */ jsxRuntimeExports.jsxs("option", { value: modelName, children: [
            provider,
            " - ",
            displayName
          ] }, modelName);
        })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "form-text text-muted mt-1", children: "Choose the AI model for code analysis. Larger models provide more detailed analysis." })
  ] });
}
const TagFilter = ({ availableTags, selectedTags, onTagToggle }) => {
  if (!availableTags || availableTags.length === 0) {
    return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-info", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-2" }),
      "No tags available. Tags are automatically generated from log content."
    ] });
  }
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "tag-filter", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center mb-2", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-tags me-2" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Filter by Tags:" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "ms-2 text-muted small", children: [
        "(",
        selectedTags.length,
        " selected)"
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex flex-wrap gap-2", children: availableTags.map((tag) => {
      const isSelected = selectedTags.includes(tag);
      return /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          type: "button",
          className: `btn btn-sm ${isSelected ? "btn-primary" : "btn-outline-secondary"}`,
          onClick: () => onTagToggle(tag),
          "aria-pressed": isSelected,
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi ${isSelected ? "bi-check-circle-fill" : "bi-tag"} me-1` }),
            tag
          ]
        },
        tag
      );
    }) }),
    selectedTags.length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-2", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
      "button",
      {
        type: "button",
        className: "btn btn-sm btn-outline-danger",
        onClick: () => selectedTags.forEach((tag) => onTagToggle(tag)),
        children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-x-circle me-1" }),
          "Clear all filters"
        ]
      }
    ) })
  ] });
};
TagFilter.propTypes = {
  availableTags: PropTypes.arrayOf(PropTypes.string).isRequired,
  selectedTags: PropTypes.arrayOf(PropTypes.string).isRequired,
  onTagToggle: PropTypes.func.isRequired
};
const LOGS_API_URL = "/api/logs";
const LogLevel = {
  INFO: "info",
  WARNING: "warning",
  ERROR: "error"
};
async function sendLog(level, message, metadata = {}, tags = []) {
  try {
    const logEntry = {
      service: "frontend",
      level,
      message,
      metadata: {
        ...metadata,
        url: window.location.href,
        userAgent: navigator.userAgent,
        timestamp: (/* @__PURE__ */ new Date()).toISOString()
      },
      tags: ["frontend", ...tags]
    };
    const isDebugEnabled = false;
    if (isDebugEnabled) ;
    fetch(LOGS_API_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json"
      },
      credentials: "include",
      body: JSON.stringify(logEntry)
    }).then((response) => {
      if (!response.ok) {
        if (isDebugEnabled) ;
        return response.text().then((text) => {
          if (isDebugEnabled) ;
        });
      }
      if (isDebugEnabled) ;
    }).catch((err) => {
      if (isDebugEnabled) ;
    });
    if (isDebugEnabled) ;
  } catch (err) {
    console.error("Logger error:", err);
  }
}
function logError(error, context = {}) {
  const metadata = {
    ...context,
    errorName: error.name,
    errorMessage: error.message,
    stack: error.stack
  };
  sendLog(LogLevel.ERROR, error.message, metadata, ["error", "uncaught"]);
}
function logWarning(message, context = {}) {
  sendLog(LogLevel.WARNING, message, context, ["warning"]);
}
function logInfo(message, context = {}) {
  sendLog(LogLevel.INFO, message, context, ["info"]);
}
function logDebug(message, context = {}) {
}
function setupGlobalErrorHandlers() {
  window.addEventListener("error", (event) => {
    logError(event.error || new Error(event.message), {
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      type: "unhandled_error"
    });
  });
  window.addEventListener("unhandledrejection", (event) => {
    const error = event.reason instanceof Error ? event.reason : new Error(String(event.reason));
    logError(error, {
      type: "unhandled_rejection",
      promise: event.promise
    });
  });
  const originalConsoleError = console.error;
  console.error = function(...args) {
    originalConsoleError.apply(console, args);
    const message = args.map(
      (arg) => typeof arg === "object" ? JSON.stringify(arg) : String(arg)
    ).join(" ");
    sendLog(LogLevel.ERROR, message, {
      type: "console_error",
      args
    }, ["console", "error"]);
  };
}
function HealthPage() {
  const { user, logout } = useAuth();
  const { theme, toggleTheme } = useTheme();
  const [activeTab, setActiveTab] = reactExports.useState("logs");
  const [unfilteredStats, setUnfilteredStats] = reactExports.useState({
    debug: 0,
    info: 0,
    warning: 0,
    error: 0,
    critical: 0
  });
  const [logs, setLogs] = reactExports.useState([]);
  const [filteredLogs, setFilteredLogs] = reactExports.useState([]);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  const [filters, setFilters] = reactExports.useState({
    level: "all",
    service: "all",
    project: "all",
    // Week 3: Add project filter
    search: ""
  });
  const [autoRefresh, setAutoRefresh] = reactExports.useState(false);
  const [selectedModel, setSelectedModel] = reactExports.useState("");
  const [showDetailModal, setShowDetailModal] = reactExports.useState(false);
  const [selectedLog, setSelectedLog] = reactExports.useState(null);
  const [aiInsights, setAiInsights] = reactExports.useState(null);
  const [loadingInsights, setLoadingInsights] = reactExports.useState(false);
  const [isGenerating, setIsGenerating] = reactExports.useState(false);
  const [wsConnected, setWsConnected] = reactExports.useState(false);
  const wsRef = reactExports.useRef(null);
  const reconnectTimeoutRef = reactExports.useRef(null);
  const [availableTags, setAvailableTags] = reactExports.useState([]);
  const [selectedTags, setSelectedTags] = reactExports.useState([]);
  const [newTagInput, setNewTagInput] = reactExports.useState("");
  const [projects, setProjects] = reactExports.useState([]);
  const [loadingProjects, setLoadingProjects] = reactExports.useState(false);
  const [addingTag, setAddingTag] = reactExports.useState(false);
  reactExports.useEffect(() => {
    const loadInitialData = async () => {
      try {
        setLoading(true);
        const [statsData, logsData, tagsData] = await Promise.all([
          apiRequest("/api/logs/v1/stats"),
          apiRequest("/api/logs?limit=100"),
          apiRequest("/api/logs/tags")
        ]);
        const entries = logsData.entries || [];
        setUnfilteredStats(statsData);
        setLogs(entries);
        setAvailableTags(tagsData.tags || []);
        setError(null);
      } catch (err) {
        logError(err, { context: "Health page initial data load failed" });
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };
    loadInitialData();
  }, [activeTab]);
  reactExports.useEffect(() => {
    if (!autoRefresh) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
        setWsConnected(false);
      }
      return;
    }
    const connectWebSocket = () => {
      const wsProtocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      const wsUrl = `${wsProtocol}//${window.location.host}/ws/logs`;
      const ws = new WebSocket(wsUrl);
      ws.onopen = () => {
        logInfo("WebSocket connection established", { autoRefresh });
        setWsConnected(true);
      };
      ws.onmessage = (event) => {
        try {
          const newLog = JSON.parse(event.data);
          logDebug("WebSocket received log", { logId: newLog.id, level: newLog.level });
          setLogs((prev) => [newLog, ...prev].slice(0, 100));
          setUnfilteredStats((prev) => ({
            ...prev,
            [newLog.level.toLowerCase()]: (prev[newLog.level.toLowerCase()] || 0) + 1
          }));
        } catch (error2) {
          logError(error2, { context: "WebSocket message parsing failed" });
        }
      };
      ws.onerror = (error2) => {
        logError(new Error("WebSocket error"), { errorEvent: error2.toString() });
        setWsConnected(false);
      };
      ws.onclose = () => {
        logInfo("WebSocket connection closed");
        setWsConnected(false);
        if (autoRefresh) {
          reconnectTimeoutRef.current = setTimeout(connectWebSocket, 5e3);
        }
      };
      wsRef.current = ws;
    };
    connectWebSocket();
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [autoRefresh]);
  const fetchData = reactExports.useCallback(async (isBackgroundRefresh = false) => {
    try {
      if (!isBackgroundRefresh) {
        setLoading(true);
      }
      let logsQuery = "/api/logs?limit=100";
      if (filters.level !== "all") {
        logsQuery += `&level=${filters.level}`;
      }
      if (filters.service !== "all") {
        logsQuery += `&service=${filters.service}`;
      }
      if (filters.project !== "all") {
        logsQuery += `&project_id=${filters.project}`;
      }
      const [statsData, logsData] = await Promise.all([
        apiRequest("/api/logs/v1/stats"),
        apiRequest(logsQuery)
      ]);
      const entries = logsData.entries || [];
      setUnfilteredStats(statsData);
      setLogs(entries);
      setError(null);
      setError(null);
    } catch (err) {
      logError(err, { context: "Health page data fetch failed" });
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [filters.level, filters.service]);
  reactExports.useEffect(() => {
    fetchData();
  }, [fetchData]);
  reactExports.useEffect(() => {
    applyFilters();
  }, [logs, filters, selectedTags]);
  const fetchAvailableTags = async () => {
    try {
      const data = await apiRequest("/api/logs/tags");
      setAvailableTags(data.tags || []);
    } catch (error2) {
      logWarning("Failed to fetch log tags", { error: error2.message });
    }
  };
  const fetchProjects = async () => {
    try {
      setLoadingProjects(true);
      const data = await apiRequest("/api/logs/projects");
      setProjects(Array.isArray(data) ? data : data.projects || []);
    } catch (error2) {
      logWarning("Failed to fetch projects", { error: error2.message });
      setProjects([]);
    } finally {
      setLoadingProjects(false);
    }
  };
  reactExports.useEffect(() => {
    fetchAvailableTags();
    fetchProjects();
  }, []);
  const handleTagToggle = (tag) => {
    setSelectedTags((prev) => {
      if (prev.includes(tag)) {
        return prev.filter((t2) => t2 !== tag);
      } else {
        return [...prev, tag];
      }
    });
  };
  const applyFilters = reactExports.useCallback(() => {
    let filtered = [...logs];
    if (filters.search) {
      const searchLower = filters.search.toLowerCase();
      filtered = filtered.filter(
        (log) => log.message.toLowerCase().includes(searchLower) || log.service.toLowerCase().includes(searchLower)
      );
    }
    if (selectedTags.length > 0) {
      filtered = filtered.filter((log) => {
        if (!log.tags || log.tags.length === 0) return false;
        return selectedTags.every((tag) => log.tags.includes(tag));
      });
    }
    setFilteredLogs(filtered);
  }, [logs, filters, selectedTags]);
  const getUniqueServices = () => {
    const logsToFilter = filters.project !== "all" ? logs.filter((log) => log.project_id === parseInt(filters.project)) : logs;
    const services = new Set(logsToFilter.map((log) => log.service));
    return Array.from(services).sort();
  };
  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    return date.toLocaleString("en-US", {
      month: "short",
      day: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit"
    });
  };
  const getLevelColor = (level) => {
    const levelLower = level.toLowerCase();
    switch (levelLower) {
      case "debug":
        return "secondary";
      case "info":
        return "info";
      case "warning":
        return "warning";
      case "error":
        return "danger";
      case "critical":
        return "danger";
      default:
        return "secondary";
    }
  };
  const openDetailModal = async (log) => {
    setSelectedLog(log);
    setAiInsights(null);
    setShowDetailModal(true);
    await fetchExistingInsights(log.id);
  };
  const fetchExistingInsights = async (logId) => {
    try {
      const data = await apiRequest(`/api/logs/${logId}/insights`);
      logDebug("Fetched existing AI insights", { logId, insightsCount: data ? 1 : 0 });
      setAiInsights(data);
      logInfo("Fetched existing AI insights", {
        log_id: logId,
        action: "fetch_insights_success"
      });
    } catch (error2) {
      if (error2.status === 404) {
        return;
      }
      logWarning("Failed to fetch existing insights", {
        log_id: logId,
        error: error2.message,
        action: "fetch_insights_error"
      });
    }
  };
  const generateAIInsights = async (logId) => {
    var _a;
    if (isGenerating) {
      return;
    }
    setLoadingInsights(true);
    setIsGenerating(true);
    try {
      logInfo("Generating AI insights", {
        log_id: logId,
        model: selectedModel,
        action: "generate_insights_start"
      });
      const data = await apiRequest(`/api/logs/${logId}/insights`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json"
        },
        body: JSON.stringify({
          model: selectedModel
        }),
        timeout: 6e4
        // 60 second timeout
      });
      logDebug("AI insights response received", { logId, hasAnalysis: !!(data == null ? void 0 : data.analysis) });
      setAiInsights(data);
      logInfo("AI insights generated successfully", {
        log_id: logId,
        model: selectedModel,
        action: "generate_insights_success"
      });
    } catch (error2) {
      if (error2.name === "AbortError" || ((_a = error2.message) == null ? void 0 : _a.includes("timeout"))) {
        const timeoutMsg = "AI analysis timed out after 60 seconds. Try a smaller/faster model or retry when server is less busy.";
        logError(new Error(timeoutMsg), {
          log_id: logId,
          model: selectedModel,
          action: "generate_insights_timeout"
        });
        setAiInsights({
          analysis: `⏱️ ${timeoutMsg}`,
          root_cause: "Request exceeded 60 second timeout limit",
          suggestions: [
            "Try a smaller model like qwen2.5-coder:7b-instruct-q4_K_M",
            "Check server logs for model loading issues",
            "Retry when server is less busy"
          ]
        });
      } else {
        logError(error2, {
          log_id: logId,
          model: selectedModel,
          status_code: error2.status,
          error_message: error2.message,
          action: "generate_insights_failed"
        });
        setAiInsights({
          analysis: `❌ Failed to generate insights: ${error2.message}`,
          root_cause: "AI service error",
          suggestions: [
            "Check that the AI model is running",
            "Verify model name is correct",
            "Check server logs for details"
          ]
        });
      }
    } finally {
      setLoadingInsights(false);
      setIsGenerating(false);
    }
  };
  const handleAddTag = async (logId, tag) => {
    if (!tag || tag.trim() === "") return;
    setAddingTag(true);
    try {
      await apiRequest(`/api/logs/${logId}/tags`, {
        method: "POST",
        body: { tag: tag.trim() }
      });
      const updatedLog = {
        ...selectedLog,
        tags: [...selectedLog.tags || [], tag.trim()]
      };
      setSelectedLog(updatedLog);
      setLogs(
        (prevLogs) => prevLogs.map(
          (log) => log.id === logId ? updatedLog : log
        )
      );
      await fetchAvailableTags();
      setNewTagInput("");
    } catch (error2) {
      logError(error2, {
        log_id: logId,
        tag,
        action: "add_tag_failed"
      });
      alert(`Failed to add tag: ${error2.message}`);
    } finally {
      setAddingTag(false);
    }
  };
  const handleRemoveTag = async (logId, tag) => {
    try {
      await apiRequest(`/api/logs/${logId}/tags/${encodeURIComponent(tag)}`, {
        method: "DELETE"
      });
      const updatedLog = {
        ...selectedLog,
        tags: (selectedLog.tags || []).filter((t2) => t2 !== tag)
      };
      setSelectedLog(updatedLog);
      setLogs(
        (prevLogs) => prevLogs.map(
          (log) => log.id === logId ? updatedLog : log
        )
      );
      await fetchAvailableTags();
    } catch (error2) {
      logError(error2, {
        log_id: logId,
        tag,
        action: "remove_tag_failed"
      });
      alert(`Failed to remove tag: ${error2.message}`);
    }
  };
  if (loading && logs.length === 0) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "container mt-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center py-5", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", style: { width: "3rem", height: "3rem" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mt-3 text-muted", children: "Loading health data..." })
    ] }) });
  }
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container mt-4", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("nav", { className: "navbar navbar-expand-lg navbar-light frosted-card mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: "/", className: "navbar-brand", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-left me-2" }),
        "Back to Dashboard"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(Link, { to: "/health", className: "navbar-brand", children: "Health" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "me-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
          ModelSelector,
          {
            selectedModel,
            onModelSelect: setSelectedModel
          }
        ) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-link p-2 me-3",
            onClick: toggleTheme,
            title: "Toggle Dark Mode",
            style: { fontSize: "1.25rem" },
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-${theme === "dark" ? "sun" : "moon"}-fill` })
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "me-3", children: [
          "Welcome, ",
          (user == null ? void 0 : user.username) || (user == null ? void 0 : user.name),
          "!"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-outline-danger btn-sm",
            onClick: () => logout(),
            children: "Logout"
          }
        )
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex justify-content-between align-items-center mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: "mb-0", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-heart-pulse text-primary me-2" }),
        "Health"
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("ul", { className: "nav nav-tabs mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { className: "nav-item", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            className: `nav-link ${activeTab === "logs" ? "active" : ""}`,
            onClick: () => setActiveTab("logs"),
            style: {
              backgroundColor: activeTab === "logs" ? "rgba(99, 102, 241, 0.2)" : "transparent",
              color: activeTab === "logs" ? "var(--bs-primary)" : "var(--bs-gray-300)",
              border: "none",
              borderBottom: activeTab === "logs" ? "2px solid var(--bs-primary)" : "none"
            },
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-list-ul me-2" }),
              "Logs"
            ]
          }
        ) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { className: "nav-item", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            className: `nav-link ${activeTab === "monitoring" ? "active" : ""}`,
            onClick: () => setActiveTab("monitoring"),
            style: {
              backgroundColor: activeTab === "monitoring" ? "rgba(99, 102, 241, 0.2)" : "transparent",
              color: activeTab === "monitoring" ? "var(--bs-primary)" : "var(--bs-gray-300)",
              border: "none",
              borderBottom: activeTab === "monitoring" ? "2px solid var(--bs-primary)" : "none"
            },
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-activity me-2" }),
              "Monitoring"
            ]
          }
        ) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { className: "nav-item", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            className: `nav-link ${activeTab === "analytics" ? "active" : ""}`,
            onClick: () => setActiveTab("analytics"),
            style: {
              backgroundColor: activeTab === "analytics" ? "rgba(99, 102, 241, 0.2)" : "transparent",
              color: activeTab === "analytics" ? "var(--bs-primary)" : "var(--bs-gray-300)",
              border: "none",
              borderBottom: activeTab === "analytics" ? "2px solid var(--bs-primary)" : "none"
            },
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-graph-up me-2" }),
              "Analytics"
            ]
          }
        ) })
      ] }),
      activeTab === "logs" && /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { color: "var(--bs-gray-200)" }, children: "Monitor your application logs in real-time" })
    ] }) }) }),
    activeTab === "logs" && /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
      loading && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "text-center my-5", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }) }),
      error && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger", role: "alert", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle-fill me-2" }),
        error
      ] }),
      !loading && !error && /* @__PURE__ */ jsxRuntimeExports.jsx(jsxRuntimeExports.Fragment, { children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row g-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col-lg-8", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
            StatCards,
            {
              stats: unfilteredStats,
              selectedLevel: filters.level === "all" ? null : filters.level,
              onLevelClick: (level) => {
                setFilters({
                  ...filters,
                  level: filters.level === level ? "all" : level
                });
              }
            }
          ) }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "mb-0", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-list-ul me-2" }),
                "Logs Feed (",
                filteredLogs.length,
                ")"
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-2 align-items-center", children: [
                autoRefresh && /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: `badge ${wsConnected ? "bg-success" : "bg-secondary"}`, children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-${wsConnected ? "check-circle" : "x-circle"} me-1` }),
                  wsConnected ? "Connected" : "Disconnected"
                ] }),
                /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "form-check form-switch", children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx(
                    "input",
                    {
                      className: "form-check-input",
                      type: "checkbox",
                      id: "autoRefresh",
                      checked: autoRefresh,
                      onChange: (e) => setAutoRefresh(e.target.checked)
                    }
                  ),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "form-check-label", htmlFor: "autoRefresh", children: "Auto-refresh" })
                ] }),
                /* @__PURE__ */ jsxRuntimeExports.jsxs(
                  "button",
                  {
                    className: "btn btn-sm btn-outline-primary",
                    onClick: () => fetchData(),
                    disabled: loading,
                    children: [
                      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-clockwise me-1" }),
                      "Refresh"
                    ]
                  }
                )
              ] })
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
                "select",
                {
                  className: "form-select form-select-sm bg-dark text-light border-secondary",
                  value: filters.project,
                  onChange: (e) => setFilters({ ...filters, project: e.target.value }),
                  disabled: loadingProjects,
                  children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "all", children: "All Projects" }),
                    projects.map((project) => /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: project.id, children: project.name }, project.id))
                  ]
                }
              ) }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
                "select",
                {
                  className: "form-select form-select-sm bg-dark text-light border-secondary",
                  value: filters.service,
                  onChange: (e) => setFilters({ ...filters, service: e.target.value }),
                  children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "all", children: "All Services" }),
                    getUniqueServices().map((service) => /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: service, children: service }, service))
                  ]
                }
              ) }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
                "input",
                {
                  type: "text",
                  className: "form-control form-control-sm bg-dark text-light border-secondary",
                  placeholder: "Search logs...",
                  value: filters.search,
                  onChange: (e) => setFilters({ ...filters, search: e.target.value })
                }
              ) })
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
              TagFilter,
              {
                availableTags,
                selectedTags,
                onTagToggle: handleTagToggle
              }
            ) }),
            filteredLogs.length === 0 ? /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center py-5", style: { color: "var(--bs-gray-400)" }, children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-inbox display-1 d-block mb-3" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", children: "No logs found matching your filters" })
            ] }) : /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "log-cards-container", style: {
              display: "flex",
              flexDirection: "column",
              gap: "0.5rem"
            }, children: filteredLogs.map((log) => /* @__PURE__ */ jsxRuntimeExports.jsx(
              "div",
              {
                className: "log-card",
                onClick: () => openDetailModal(log),
                style: {
                  background: theme === "dark" ? "rgba(255, 255, 255, 0.05)" : "var(--bs-card-bg)",
                  border: `1px solid ${theme === "dark" ? "rgba(255, 255, 255, 0.1)" : "var(--bs-border-color)"}`,
                  borderRadius: "0.5rem",
                  padding: "1rem",
                  cursor: "pointer",
                  transition: "all 0.2s"
                },
                onMouseEnter: (e) => {
                  e.currentTarget.style.borderColor = "var(--bs-primary)";
                  e.currentTarget.style.boxShadow = "0 2px 8px rgba(99, 102, 241, 0.15)";
                  e.currentTarget.style.transform = "translateY(-1px)";
                },
                onMouseLeave: (e) => {
                  e.currentTarget.style.borderColor = theme === "dark" ? "rgba(255, 255, 255, 0.1)" : "var(--bs-border-color)";
                  e.currentTarget.style.boxShadow = "none";
                  e.currentTarget.style.transform = "translateY(0)";
                },
                children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { style: {
                  display: "grid",
                  gridTemplateColumns: "100px 150px 180px 1fr",
                  gap: "1rem",
                  alignItems: "center"
                }, children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("div", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: `badge bg-${getLevelColor(log.level)}`, children: log.level.toUpperCase() }) }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap"
                  }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "text-primary", children: log.service }) }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap"
                  }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("small", { style: { color: "var(--bs-gray-400)" }, children: formatTimestamp(log.created_at) }) }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("div", { style: {
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap"
                  }, children: log.message })
                ] })
              },
              log.id
            )) })
          ] })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col-lg-4", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-3 mb-3", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-speedometer2 me-2" }),
              "Quick Stats"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex flex-column gap-2", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center p-2 rounded", style: { backgroundColor: "rgba(99, 102, 241, 0.05)" }, children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Total Logs" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: logs.length })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center p-2 rounded", style: { backgroundColor: "rgba(220, 38, 38, 0.05)" }, children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Errors" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { className: "text-danger", children: unfilteredStats.error + unfilteredStats.critical })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center p-2 rounded", style: { backgroundColor: "rgba(234, 179, 8, 0.05)" }, children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Warnings" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { className: "text-warning", children: unfilteredStats.warning })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center p-2 rounded", style: { backgroundColor: "rgba(59, 130, 246, 0.05)" }, children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Info" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { className: "text-info", children: unfilteredStats.info })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center p-2 rounded", style: { backgroundColor: "rgba(34, 197, 94, 0.05)" }, children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Debug" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { className: "text-success", children: unfilteredStats.debug })
              ] })
            ] })
          ] }),
          (filters.level !== "all" || filters.service !== "all" || filters.search || selectedTags.length > 0) && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-3 mb-3", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-funnel me-2" }),
              "Active Filters"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex flex-column gap-2", children: [
              filters.level !== "all" && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Level:" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: `badge bg-${getLevelColor(filters.level)}`, children: filters.level.toUpperCase() })
              ] }),
              filters.service !== "all" && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Service:" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "small text-primary", children: filters.service })
              ] }),
              filters.search && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Search:" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "small", children: filters.search })
              ] }),
              selectedTags.length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex flex-column gap-1", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "small", children: "Tags:" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex flex-wrap gap-1", children: selectedTags.map((tag) => /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-secondary small", children: tag }, tag)) })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs(
                "button",
                {
                  className: "btn btn-sm btn-outline-secondary mt-2",
                  onClick: () => {
                    setFilters({ level: "all", service: "all", search: "" });
                    setSelectedTags([]);
                  },
                  children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-x-circle me-1" }),
                    "Clear All"
                  ]
                }
              )
            ] })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-3", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle-fill text-danger me-2" }),
              "Critical Events"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex flex-column gap-2", children: [
              logs.filter((log) => log.level === "error" || log.level === "critical").slice(0, 5).map((log) => /* @__PURE__ */ jsxRuntimeExports.jsxs(
                "div",
                {
                  className: "p-2 rounded",
                  style: {
                    backgroundColor: "rgba(220, 38, 38, 0.05)",
                    borderLeft: "3px solid var(--bs-danger)",
                    cursor: "pointer"
                  },
                  onClick: () => handleViewDetails(log),
                  children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-start mb-1", children: [
                      /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted", children: formatTimestamp(log.created_at) }),
                      /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: `badge badge-sm bg-${getLevelColor(log.level)}`, children: log.level })
                    ] }),
                    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "small", style: {
                      overflow: "hidden",
                      textOverflow: "ellipsis",
                      display: "-webkit-box",
                      WebkitLineClamp: 2,
                      WebkitBoxOrient: "vertical"
                    }, children: log.message })
                  ]
                },
                log.id
              )),
              logs.filter((log) => log.level === "error" || log.level === "critical").length === 0 && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center py-3 text-muted small", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-check-circle me-1" }),
                "No critical events"
              ] })
            ] })
          ] })
        ] })
      ] }) })
    ] }),
    activeTab === "monitoring" && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "monitoring-tab text-center py-5", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-activity", style: { fontSize: "4rem", opacity: 0.3 } }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3 text-muted", children: "Monitoring Dashboard" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: "Coming Soon" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "small text-muted", children: "Real-time service health, request rates, and error tracking." })
    ] }),
    activeTab === "analytics" && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "analytics-tab text-center py-5", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-graph-up", style: { fontSize: "4rem", opacity: 0.3 } }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3 text-muted", children: "Analytics Dashboard" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: "Coming Soon" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "small text-muted", children: "Historical trends, error patterns, and AI-powered insights." })
    ] }),
    selectedLog && /* @__PURE__ */ jsxRuntimeExports.jsxs(Modal$1, { show: showDetailModal, onHide: () => setShowDetailModal(false), size: "lg", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx(Modal$1.Header, { closeButton: true, className: theme === "dark" ? "bg-dark text-light border-secondary" : "", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(Modal$1.Title, { children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: `badge bg-${getLevelColor(selectedLog.level)} me-2`, children: selectedLog.level.toUpperCase() }),
        "Log Details"
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs(Modal$1.Body, { className: theme === "dark" ? "bg-dark text-light" : "", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col-md-6", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Service:" }),
            " ",
            /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "text-primary", children: selectedLog.service })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col-md-6", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Timestamp:" }),
            " ",
            formatTimestamp(selectedLog.created_at)
          ] })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Message:" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-2 p-3 rounded", style: {
            backgroundColor: theme === "dark" ? "rgba(0,0,0,0.3)" : "rgba(0,0,0,0.05)",
            border: `1px solid ${theme === "dark" ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.1)"}`
          }, children: selectedLog.message })
        ] }),
        selectedLog.metadata && Object.keys(selectedLog.metadata).length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Metadata:" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("pre", { className: "mt-2 p-3 rounded", style: {
            backgroundColor: theme === "dark" ? "rgba(0,0,0,0.3)" : "rgba(0,0,0,0.05)",
            color: theme === "dark" ? "var(--bs-gray-300)" : "var(--bs-gray-800)",
            border: `1px solid ${theme === "dark" ? "rgba(255,255,255,0.1)" : "rgba(0,0,0,0.1)"}`
          }, children: JSON.stringify(selectedLog.metadata, null, 2) })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex justify-content-between align-items-center mb-2", children: /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Tags:" }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-2", children: selectedLog.tags && selectedLog.tags.length > 0 ? /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex flex-wrap gap-2", children: selectedLog.tags.map((tag, idx) => /* @__PURE__ */ jsxRuntimeExports.jsxs(
            "span",
            {
              className: "badge bg-secondary d-flex align-items-center",
              style: { fontSize: "0.9rem" },
              children: [
                tag,
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "button",
                  {
                    className: "btn-close btn-close-white ms-2",
                    style: { fontSize: "0.6rem" },
                    onClick: () => handleRemoveTag(selectedLog.id, tag),
                    title: "Remove tag",
                    "aria-label": `Remove ${tag} tag`
                  }
                )
              ]
            },
            idx
          )) }) : /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "text-muted small", children: "No tags yet" }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "input-group input-group-sm", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "input",
              {
                type: "text",
                className: `form-control ${theme === "dark" ? "bg-dark text-light border-secondary" : ""}`,
                placeholder: "Add a tag (e.g., investigated, resolved)",
                value: newTagInput,
                onChange: (e) => setNewTagInput(e.target.value),
                onKeyPress: (e) => {
                  if (e.key === "Enter") {
                    handleAddTag(selectedLog.id, newTagInput);
                  }
                },
                disabled: addingTag
              }
            ),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                className: "btn btn-outline-primary btn-sm",
                onClick: () => handleAddTag(selectedLog.id, newTagInput),
                disabled: addingTag || !newTagInput.trim(),
                children: addingTag ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-1" }),
                  "Adding..."
                ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle me-1" }),
                  "Add Tag"
                ] })
              }
            )
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted", children: "Tags help categorize and filter logs. Press Enter or click Add Tag." })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center mb-2", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "AI Insights:" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                className: "btn btn-primary btn-sm",
                onClick: () => generateAIInsights(selectedLog.id),
                disabled: loadingInsights || isGenerating,
                children: loadingInsights || isGenerating ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-2" }),
                  "Analyzing..."
                ] }) : aiInsights ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-clockwise me-2" }),
                  "Regenerate"
                ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-stars me-2" }),
                  "Generate Insights"
                ] })
              }
            )
          ] }),
          aiInsights && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "p-3 rounded", style: {
            backgroundColor: theme === "dark" ? "rgba(99,102,241,0.1)" : "rgba(99,102,241,0.05)",
            border: "1px solid rgba(99,102,241,0.3)"
          }, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-2", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Analysis:" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-2", children: aiInsights.analysis })
            ] }),
            aiInsights.root_cause && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-2", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Root Cause:" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-2", children: aiInsights.root_cause })
            ] }),
            aiInsights.suggestions && aiInsights.suggestions.length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Suggestions:" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("ul", { className: "mb-0", children: aiInsights.suggestions.map((suggestion, idx) => /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: suggestion }, idx)) })
            ] })
          ] })
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(Modal$1.Footer, { className: theme === "dark" ? "bg-dark border-secondary" : "", children: /* @__PURE__ */ jsxRuntimeExports.jsx("button", { className: "btn btn-secondary", onClick: () => setShowDetailModal(false), children: "Close" }) })
    ] })
  ] });
}
function _arrayLikeToArray(r2, a) {
  (null == a || a > r2.length) && (a = r2.length);
  for (var e = 0, n2 = Array(a); e < a; e++) n2[e] = r2[e];
  return n2;
}
function _arrayWithHoles(r2) {
  if (Array.isArray(r2)) return r2;
}
function _defineProperty$1(e, r2, t2) {
  return (r2 = _toPropertyKey(r2)) in e ? Object.defineProperty(e, r2, {
    value: t2,
    enumerable: true,
    configurable: true,
    writable: true
  }) : e[r2] = t2, e;
}
function _iterableToArrayLimit(r2, l2) {
  var t2 = null == r2 ? null : "undefined" != typeof Symbol && r2[Symbol.iterator] || r2["@@iterator"];
  if (null != t2) {
    var e, n2, i, u2, a = [], f2 = true, o = false;
    try {
      if (i = (t2 = t2.call(r2)).next, 0 === l2) ;
      else for (; !(f2 = (e = i.call(t2)).done) && (a.push(e.value), a.length !== l2); f2 = true) ;
    } catch (r3) {
      o = true, n2 = r3;
    } finally {
      try {
        if (!f2 && null != t2.return && (u2 = t2.return(), Object(u2) !== u2)) return;
      } finally {
        if (o) throw n2;
      }
    }
    return a;
  }
}
function _nonIterableRest() {
  throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function ownKeys$1(e, r2) {
  var t2 = Object.keys(e);
  if (Object.getOwnPropertySymbols) {
    var o = Object.getOwnPropertySymbols(e);
    r2 && (o = o.filter(function(r3) {
      return Object.getOwnPropertyDescriptor(e, r3).enumerable;
    })), t2.push.apply(t2, o);
  }
  return t2;
}
function _objectSpread2$1(e) {
  for (var r2 = 1; r2 < arguments.length; r2++) {
    var t2 = null != arguments[r2] ? arguments[r2] : {};
    r2 % 2 ? ownKeys$1(Object(t2), true).forEach(function(r3) {
      _defineProperty$1(e, r3, t2[r3]);
    }) : Object.getOwnPropertyDescriptors ? Object.defineProperties(e, Object.getOwnPropertyDescriptors(t2)) : ownKeys$1(Object(t2)).forEach(function(r3) {
      Object.defineProperty(e, r3, Object.getOwnPropertyDescriptor(t2, r3));
    });
  }
  return e;
}
function _objectWithoutProperties(e, t2) {
  if (null == e) return {};
  var o, r2, i = _objectWithoutPropertiesLoose(e, t2);
  if (Object.getOwnPropertySymbols) {
    var n2 = Object.getOwnPropertySymbols(e);
    for (r2 = 0; r2 < n2.length; r2++) o = n2[r2], -1 === t2.indexOf(o) && {}.propertyIsEnumerable.call(e, o) && (i[o] = e[o]);
  }
  return i;
}
function _objectWithoutPropertiesLoose(r2, e) {
  if (null == r2) return {};
  var t2 = {};
  for (var n2 in r2) if ({}.hasOwnProperty.call(r2, n2)) {
    if (-1 !== e.indexOf(n2)) continue;
    t2[n2] = r2[n2];
  }
  return t2;
}
function _slicedToArray(r2, e) {
  return _arrayWithHoles(r2) || _iterableToArrayLimit(r2, e) || _unsupportedIterableToArray(r2, e) || _nonIterableRest();
}
function _toPrimitive(t2, r2) {
  if ("object" != typeof t2 || !t2) return t2;
  var e = t2[Symbol.toPrimitive];
  if (void 0 !== e) {
    var i = e.call(t2, r2);
    if ("object" != typeof i) return i;
    throw new TypeError("@@toPrimitive must return a primitive value.");
  }
  return ("string" === r2 ? String : Number)(t2);
}
function _toPropertyKey(t2) {
  var i = _toPrimitive(t2, "string");
  return "symbol" == typeof i ? i : i + "";
}
function _unsupportedIterableToArray(r2, a) {
  if (r2) {
    if ("string" == typeof r2) return _arrayLikeToArray(r2, a);
    var t2 = {}.toString.call(r2).slice(8, -1);
    return "Object" === t2 && r2.constructor && (t2 = r2.constructor.name), "Map" === t2 || "Set" === t2 ? Array.from(r2) : "Arguments" === t2 || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t2) ? _arrayLikeToArray(r2, a) : void 0;
  }
}
function _defineProperty(obj, key, value) {
  if (key in obj) {
    Object.defineProperty(obj, key, {
      value,
      enumerable: true,
      configurable: true,
      writable: true
    });
  } else {
    obj[key] = value;
  }
  return obj;
}
function ownKeys(object, enumerableOnly) {
  var keys = Object.keys(object);
  if (Object.getOwnPropertySymbols) {
    var symbols = Object.getOwnPropertySymbols(object);
    if (enumerableOnly) symbols = symbols.filter(function(sym) {
      return Object.getOwnPropertyDescriptor(object, sym).enumerable;
    });
    keys.push.apply(keys, symbols);
  }
  return keys;
}
function _objectSpread2(target) {
  for (var i = 1; i < arguments.length; i++) {
    var source = arguments[i] != null ? arguments[i] : {};
    if (i % 2) {
      ownKeys(Object(source), true).forEach(function(key) {
        _defineProperty(target, key, source[key]);
      });
    } else if (Object.getOwnPropertyDescriptors) {
      Object.defineProperties(target, Object.getOwnPropertyDescriptors(source));
    } else {
      ownKeys(Object(source)).forEach(function(key) {
        Object.defineProperty(target, key, Object.getOwnPropertyDescriptor(source, key));
      });
    }
  }
  return target;
}
function compose$1() {
  for (var _len = arguments.length, fns = new Array(_len), _key = 0; _key < _len; _key++) {
    fns[_key] = arguments[_key];
  }
  return function(x2) {
    return fns.reduceRight(function(y2, f2) {
      return f2(y2);
    }, x2);
  };
}
function curry$1(fn) {
  return function curried() {
    var _this = this;
    for (var _len2 = arguments.length, args = new Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
      args[_key2] = arguments[_key2];
    }
    return args.length >= fn.length ? fn.apply(this, args) : function() {
      for (var _len3 = arguments.length, nextArgs = new Array(_len3), _key3 = 0; _key3 < _len3; _key3++) {
        nextArgs[_key3] = arguments[_key3];
      }
      return curried.apply(_this, [].concat(args, nextArgs));
    };
  };
}
function isObject$1(value) {
  return {}.toString.call(value).includes("Object");
}
function isEmpty(obj) {
  return !Object.keys(obj).length;
}
function isFunction(value) {
  return typeof value === "function";
}
function hasOwnProperty(object, property) {
  return Object.prototype.hasOwnProperty.call(object, property);
}
function validateChanges(initial, changes) {
  if (!isObject$1(changes)) errorHandler$1("changeType");
  if (Object.keys(changes).some(function(field) {
    return !hasOwnProperty(initial, field);
  })) errorHandler$1("changeField");
  return changes;
}
function validateSelector(selector) {
  if (!isFunction(selector)) errorHandler$1("selectorType");
}
function validateHandler(handler) {
  if (!(isFunction(handler) || isObject$1(handler))) errorHandler$1("handlerType");
  if (isObject$1(handler) && Object.values(handler).some(function(_handler) {
    return !isFunction(_handler);
  })) errorHandler$1("handlersType");
}
function validateInitial(initial) {
  if (!initial) errorHandler$1("initialIsRequired");
  if (!isObject$1(initial)) errorHandler$1("initialType");
  if (isEmpty(initial)) errorHandler$1("initialContent");
}
function throwError$1(errorMessages2, type) {
  throw new Error(errorMessages2[type] || errorMessages2["default"]);
}
var errorMessages$1 = {
  initialIsRequired: "initial state is required",
  initialType: "initial state should be an object",
  initialContent: "initial state shouldn't be an empty object",
  handlerType: "handler should be an object or a function",
  handlersType: "all handlers should be a functions",
  selectorType: "selector should be a function",
  changeType: "provided value of changes should be an object",
  changeField: 'it seams you want to change a field in the state which is not specified in the "initial" state',
  "default": "an unknown error accured in `state-local` package"
};
var errorHandler$1 = curry$1(throwError$1)(errorMessages$1);
var validators$1 = {
  changes: validateChanges,
  selector: validateSelector,
  handler: validateHandler,
  initial: validateInitial
};
function create(initial) {
  var handler = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
  validators$1.initial(initial);
  validators$1.handler(handler);
  var state = {
    current: initial
  };
  var didUpdate = curry$1(didStateUpdate)(state, handler);
  var update = curry$1(updateState)(state);
  var validate = curry$1(validators$1.changes)(initial);
  var getChanges = curry$1(extractChanges)(state);
  function getState2() {
    var selector = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : function(state2) {
      return state2;
    };
    validators$1.selector(selector);
    return selector(state.current);
  }
  function setState2(causedChanges) {
    compose$1(didUpdate, update, validate, getChanges)(causedChanges);
  }
  return [getState2, setState2];
}
function extractChanges(state, causedChanges) {
  return isFunction(causedChanges) ? causedChanges(state.current) : causedChanges;
}
function updateState(state, changes) {
  state.current = _objectSpread2(_objectSpread2({}, state.current), changes);
  return changes;
}
function didStateUpdate(state, handler, changes) {
  isFunction(handler) ? handler(state.current) : Object.keys(changes).forEach(function(field) {
    var _handler$field;
    return (_handler$field = handler[field]) === null || _handler$field === void 0 ? void 0 : _handler$field.call(handler, state.current[field]);
  });
  return changes;
}
var index = {
  create
};
var config$1 = {
  paths: {
    vs: "https://cdn.jsdelivr.net/npm/monaco-editor@0.54.0/min/vs"
  }
};
function curry(fn) {
  return function curried() {
    var _this = this;
    for (var _len = arguments.length, args = new Array(_len), _key = 0; _key < _len; _key++) {
      args[_key] = arguments[_key];
    }
    return args.length >= fn.length ? fn.apply(this, args) : function() {
      for (var _len2 = arguments.length, nextArgs = new Array(_len2), _key2 = 0; _key2 < _len2; _key2++) {
        nextArgs[_key2] = arguments[_key2];
      }
      return curried.apply(_this, [].concat(args, nextArgs));
    };
  };
}
function isObject(value) {
  return {}.toString.call(value).includes("Object");
}
function validateConfig(config2) {
  if (!config2) errorHandler("configIsRequired");
  if (!isObject(config2)) errorHandler("configType");
  if (config2.urls) {
    informAboutDeprecation();
    return {
      paths: {
        vs: config2.urls.monacoBase
      }
    };
  }
  return config2;
}
function informAboutDeprecation() {
  console.warn(errorMessages.deprecation);
}
function throwError(errorMessages2, type) {
  throw new Error(errorMessages2[type] || errorMessages2["default"]);
}
var errorMessages = {
  configIsRequired: "the configuration object is required",
  configType: "the configuration object should be an object",
  "default": "an unknown error accured in `@monaco-editor/loader` package",
  deprecation: "Deprecation warning!\n    You are using deprecated way of configuration.\n\n    Instead of using\n      monaco.config({ urls: { monacoBase: '...' } })\n    use\n      monaco.config({ paths: { vs: '...' } })\n\n    For more please check the link https://github.com/suren-atoyan/monaco-loader#config\n  "
};
var errorHandler = curry(throwError)(errorMessages);
var validators = {
  config: validateConfig
};
var compose = function compose2() {
  for (var _len = arguments.length, fns = new Array(_len), _key = 0; _key < _len; _key++) {
    fns[_key] = arguments[_key];
  }
  return function(x2) {
    return fns.reduceRight(function(y2, f2) {
      return f2(y2);
    }, x2);
  };
};
function merge(target, source) {
  Object.keys(source).forEach(function(key) {
    if (source[key] instanceof Object) {
      if (target[key]) {
        Object.assign(source[key], merge(target[key], source[key]));
      }
    }
  });
  return _objectSpread2$1(_objectSpread2$1({}, target), source);
}
var CANCELATION_MESSAGE = {
  type: "cancelation",
  msg: "operation is manually canceled"
};
function makeCancelable(promise) {
  var hasCanceled_ = false;
  var wrappedPromise = new Promise(function(resolve, reject) {
    promise.then(function(val) {
      return hasCanceled_ ? reject(CANCELATION_MESSAGE) : resolve(val);
    });
    promise["catch"](reject);
  });
  return wrappedPromise.cancel = function() {
    return hasCanceled_ = true;
  }, wrappedPromise;
}
var _excluded = ["monaco"];
var _state$create = index.create({
  config: config$1,
  isInitialized: false,
  resolve: null,
  reject: null,
  monaco: null
}), _state$create2 = _slicedToArray(_state$create, 2), getState = _state$create2[0], setState = _state$create2[1];
function config(globalConfig) {
  var _validators$config = validators.config(globalConfig), monaco = _validators$config.monaco, config2 = _objectWithoutProperties(_validators$config, _excluded);
  setState(function(state) {
    return {
      config: merge(state.config, config2),
      monaco
    };
  });
}
function init() {
  var state = getState(function(_ref) {
    var monaco = _ref.monaco, isInitialized = _ref.isInitialized, resolve = _ref.resolve;
    return {
      monaco,
      isInitialized,
      resolve
    };
  });
  if (!state.isInitialized) {
    setState({
      isInitialized: true
    });
    if (state.monaco) {
      state.resolve(state.monaco);
      return makeCancelable(wrapperPromise);
    }
    if (window.monaco && window.monaco.editor) {
      storeMonacoInstance(window.monaco);
      state.resolve(window.monaco);
      return makeCancelable(wrapperPromise);
    }
    compose(injectScripts, getMonacoLoaderScript)(configureLoader);
  }
  return makeCancelable(wrapperPromise);
}
function injectScripts(script) {
  return document.body.appendChild(script);
}
function createScript(src) {
  var script = document.createElement("script");
  return src && (script.src = src), script;
}
function getMonacoLoaderScript(configureLoader2) {
  var state = getState(function(_ref2) {
    var config2 = _ref2.config, reject = _ref2.reject;
    return {
      config: config2,
      reject
    };
  });
  var loaderScript = createScript("".concat(state.config.paths.vs, "/loader.js"));
  loaderScript.onload = function() {
    return configureLoader2();
  };
  loaderScript.onerror = state.reject;
  return loaderScript;
}
function configureLoader() {
  var state = getState(function(_ref3) {
    var config2 = _ref3.config, resolve = _ref3.resolve, reject = _ref3.reject;
    return {
      config: config2,
      resolve,
      reject
    };
  });
  var require2 = window.require;
  require2.config(state.config);
  require2(["vs/editor/editor.main"], function(_ref4) {
    var monaco = _ref4.m;
    storeMonacoInstance(monaco);
    state.resolve(monaco);
  }, function(error) {
    state.reject(error);
  });
}
function storeMonacoInstance(monaco) {
  if (!getState().monaco) {
    setState({
      monaco
    });
  }
}
function __getMonacoInstance() {
  return getState(function(_ref5) {
    var monaco = _ref5.monaco;
    return monaco;
  });
}
var wrapperPromise = new Promise(function(resolve, reject) {
  return setState({
    resolve,
    reject
  });
});
var loader = {
  config,
  init,
  __getMonacoInstance
};
var le = { wrapper: { display: "flex", position: "relative", textAlign: "initial" }, fullWidth: { width: "100%" }, hide: { display: "none" } }, v = le;
var ae = { container: { display: "flex", height: "100%", width: "100%", justifyContent: "center", alignItems: "center" } }, Y = ae;
function Me({ children: e }) {
  return We$1.createElement("div", { style: Y.container }, e);
}
var Z = Me;
var $ = Z;
function Ee({ width: e, height: r2, isEditorReady: n2, loading: t2, _ref: a, className: m2, wrapperProps: E2 }) {
  return We$1.createElement("section", { style: { ...v.wrapper, width: e, height: r2 }, ...E2 }, !n2 && We$1.createElement($, null, t2), We$1.createElement("div", { ref: a, style: { ...v.fullWidth, ...!n2 && v.hide }, className: m2 }));
}
var ee = Ee;
var H = reactExports.memo(ee);
function Ce(e) {
  reactExports.useEffect(e, []);
}
var k = Ce;
function he(e, r2, n2 = true) {
  let t2 = reactExports.useRef(true);
  reactExports.useEffect(t2.current || !n2 ? () => {
    t2.current = false;
  } : e, r2);
}
var l = he;
function D() {
}
function h(e, r2, n2, t2) {
  return De(e, t2) || be(e, r2, n2, t2);
}
function De(e, r2) {
  return e.editor.getModel(te(e, r2));
}
function be(e, r2, n2, t2) {
  return e.editor.createModel(r2, n2, t2 ? te(e, t2) : void 0);
}
function te(e, r2) {
  return e.Uri.parse(r2);
}
function Oe({ original: e, modified: r2, language: n2, originalLanguage: t2, modifiedLanguage: a, originalModelPath: m2, modifiedModelPath: E2, keepCurrentOriginalModel: g = false, keepCurrentModifiedModel: N2 = false, theme: x2 = "light", loading: P2 = "Loading...", options: y2 = {}, height: V2 = "100%", width: z2 = "100%", className: F2, wrapperProps: j = {}, beforeMount: A2 = D, onMount: q2 = D }) {
  let [M2, O2] = reactExports.useState(false), [T2, s] = reactExports.useState(true), u2 = reactExports.useRef(null), c = reactExports.useRef(null), w2 = reactExports.useRef(null), d = reactExports.useRef(q2), o = reactExports.useRef(A2), b = reactExports.useRef(false);
  k(() => {
    let i = loader.init();
    return i.then((f2) => (c.current = f2) && s(false)).catch((f2) => (f2 == null ? void 0 : f2.type) !== "cancelation" && console.error("Monaco initialization: error:", f2)), () => u2.current ? I2() : i.cancel();
  }), l(() => {
    if (u2.current && c.current) {
      let i = u2.current.getOriginalEditor(), f2 = h(c.current, e || "", t2 || n2 || "text", m2 || "");
      f2 !== i.getModel() && i.setModel(f2);
    }
  }, [m2], M2), l(() => {
    if (u2.current && c.current) {
      let i = u2.current.getModifiedEditor(), f2 = h(c.current, r2 || "", a || n2 || "text", E2 || "");
      f2 !== i.getModel() && i.setModel(f2);
    }
  }, [E2], M2), l(() => {
    let i = u2.current.getModifiedEditor();
    i.getOption(c.current.editor.EditorOption.readOnly) ? i.setValue(r2 || "") : r2 !== i.getValue() && (i.executeEdits("", [{ range: i.getModel().getFullModelRange(), text: r2 || "", forceMoveMarkers: true }]), i.pushUndoStop());
  }, [r2], M2), l(() => {
    var _a, _b;
    (_b = (_a = u2.current) == null ? void 0 : _a.getModel()) == null ? void 0 : _b.original.setValue(e || "");
  }, [e], M2), l(() => {
    let { original: i, modified: f2 } = u2.current.getModel();
    c.current.editor.setModelLanguage(i, t2 || n2 || "text"), c.current.editor.setModelLanguage(f2, a || n2 || "text");
  }, [n2, t2, a], M2), l(() => {
    var _a;
    (_a = c.current) == null ? void 0 : _a.editor.setTheme(x2);
  }, [x2], M2), l(() => {
    var _a;
    (_a = u2.current) == null ? void 0 : _a.updateOptions(y2);
  }, [y2], M2);
  let L2 = reactExports.useCallback(() => {
    var _a;
    if (!c.current) return;
    o.current(c.current);
    let i = h(c.current, e || "", t2 || n2 || "text", m2 || ""), f2 = h(c.current, r2 || "", a || n2 || "text", E2 || "");
    (_a = u2.current) == null ? void 0 : _a.setModel({ original: i, modified: f2 });
  }, [n2, r2, a, e, t2, m2, E2]), U2 = reactExports.useCallback(() => {
    var _a;
    !b.current && w2.current && (u2.current = c.current.editor.createDiffEditor(w2.current, { automaticLayout: true, ...y2 }), L2(), (_a = c.current) == null ? void 0 : _a.editor.setTheme(x2), O2(true), b.current = true);
  }, [y2, x2, L2]);
  reactExports.useEffect(() => {
    M2 && d.current(u2.current, c.current);
  }, [M2]), reactExports.useEffect(() => {
    !T2 && !M2 && U2();
  }, [T2, M2, U2]);
  function I2() {
    var _a, _b, _c, _d;
    let i = (_a = u2.current) == null ? void 0 : _a.getModel();
    g || ((_b = i == null ? void 0 : i.original) == null ? void 0 : _b.dispose()), N2 || ((_c = i == null ? void 0 : i.modified) == null ? void 0 : _c.dispose()), (_d = u2.current) == null ? void 0 : _d.dispose();
  }
  return We$1.createElement(H, { width: z2, height: V2, isEditorReady: M2, loading: P2, _ref: w2, className: F2, wrapperProps: j });
}
var ie = Oe;
reactExports.memo(ie);
function He(e) {
  let r2 = reactExports.useRef();
  return reactExports.useEffect(() => {
    r2.current = e;
  }, [e]), r2.current;
}
var se = He;
var _ = /* @__PURE__ */ new Map();
function Ve({ defaultValue: e, defaultLanguage: r2, defaultPath: n2, value: t2, language: a, path: m2, theme: E2 = "light", line: g, loading: N2 = "Loading...", options: x2 = {}, overrideServices: P2 = {}, saveViewState: y2 = true, keepCurrentModel: V2 = false, width: z2 = "100%", height: F2 = "100%", className: j, wrapperProps: A2 = {}, beforeMount: q2 = D, onMount: M2 = D, onChange: O2, onValidate: T2 = D }) {
  let [s, u2] = reactExports.useState(false), [c, w2] = reactExports.useState(true), d = reactExports.useRef(null), o = reactExports.useRef(null), b = reactExports.useRef(null), L2 = reactExports.useRef(M2), U2 = reactExports.useRef(q2), I2 = reactExports.useRef(), i = reactExports.useRef(t2), f2 = se(m2), Q2 = reactExports.useRef(false), B2 = reactExports.useRef(false);
  k(() => {
    let p2 = loader.init();
    return p2.then((R2) => (d.current = R2) && w2(false)).catch((R2) => (R2 == null ? void 0 : R2.type) !== "cancelation" && console.error("Monaco initialization: error:", R2)), () => o.current ? pe2() : p2.cancel();
  }), l(() => {
    var _a, _b, _c, _d;
    let p2 = h(d.current, e || t2 || "", r2 || a || "", m2 || n2 || "");
    p2 !== ((_a = o.current) == null ? void 0 : _a.getModel()) && (y2 && _.set(f2, (_b = o.current) == null ? void 0 : _b.saveViewState()), (_c = o.current) == null ? void 0 : _c.setModel(p2), y2 && ((_d = o.current) == null ? void 0 : _d.restoreViewState(_.get(m2))));
  }, [m2], s), l(() => {
    var _a;
    (_a = o.current) == null ? void 0 : _a.updateOptions(x2);
  }, [x2], s), l(() => {
    !o.current || t2 === void 0 || (o.current.getOption(d.current.editor.EditorOption.readOnly) ? o.current.setValue(t2) : t2 !== o.current.getValue() && (B2.current = true, o.current.executeEdits("", [{ range: o.current.getModel().getFullModelRange(), text: t2, forceMoveMarkers: true }]), o.current.pushUndoStop(), B2.current = false));
  }, [t2], s), l(() => {
    var _a, _b;
    let p2 = (_a = o.current) == null ? void 0 : _a.getModel();
    p2 && a && ((_b = d.current) == null ? void 0 : _b.editor.setModelLanguage(p2, a));
  }, [a], s), l(() => {
    var _a;
    g !== void 0 && ((_a = o.current) == null ? void 0 : _a.revealLine(g));
  }, [g], s), l(() => {
    var _a;
    (_a = d.current) == null ? void 0 : _a.editor.setTheme(E2);
  }, [E2], s);
  let X2 = reactExports.useCallback(() => {
    var _a;
    if (!(!b.current || !d.current) && !Q2.current) {
      U2.current(d.current);
      let p2 = m2 || n2, R2 = h(d.current, t2 || e || "", r2 || a || "", p2 || "");
      o.current = (_a = d.current) == null ? void 0 : _a.editor.create(b.current, { model: R2, automaticLayout: true, ...x2 }, P2), y2 && o.current.restoreViewState(_.get(p2)), d.current.editor.setTheme(E2), g !== void 0 && o.current.revealLine(g), u2(true), Q2.current = true;
    }
  }, [e, r2, n2, t2, a, m2, x2, P2, y2, E2, g]);
  reactExports.useEffect(() => {
    s && L2.current(o.current, d.current);
  }, [s]), reactExports.useEffect(() => {
    !c && !s && X2();
  }, [c, s, X2]), i.current = t2, reactExports.useEffect(() => {
    var _a, _b;
    s && O2 && ((_a = I2.current) == null ? void 0 : _a.dispose(), I2.current = (_b = o.current) == null ? void 0 : _b.onDidChangeModelContent((p2) => {
      B2.current || O2(o.current.getValue(), p2);
    }));
  }, [s, O2]), reactExports.useEffect(() => {
    if (s) {
      let p2 = d.current.editor.onDidChangeMarkers((R2) => {
        var _a;
        let G2 = (_a = o.current.getModel()) == null ? void 0 : _a.uri;
        if (G2 && R2.find((J2) => J2.path === G2.path)) {
          let J2 = d.current.editor.getModelMarkers({ resource: G2 });
          T2 == null ? void 0 : T2(J2);
        }
      });
      return () => {
        p2 == null ? void 0 : p2.dispose();
      };
    }
    return () => {
    };
  }, [s, T2]);
  function pe2() {
    var _a, _b;
    (_a = I2.current) == null ? void 0 : _a.dispose(), V2 ? y2 && _.set(m2, o.current.saveViewState()) : (_b = o.current.getModel()) == null ? void 0 : _b.dispose(), o.current.dispose();
  }
  return We$1.createElement(H, { width: z2, height: F2, isEditorReady: s, loading: N2, _ref: b, className: j, wrapperProps: A2 });
}
var fe = Ve;
var de = reactExports.memo(fe);
var Ft = de;
function CodeEditor({
  value = "",
  onChange,
  language = "javascript",
  placeholder = "Enter your code here...",
  readOnly = false,
  className = "",
  height = "600px"
}) {
  const { isDarkMode } = useTheme();
  const editorRef = reactExports.useRef(null);
  const [fontSize, setFontSize] = reactExports.useState("medium");
  const fontSizes = {
    xsmall: 12,
    // 12px
    small: 14,
    // 14px
    medium: 16,
    // 16px (default - middle size)
    large: 18,
    // 18px
    xlarge: 20
    // 20px
  };
  const getMonacoLanguage = (lang) => {
    const languageMap = {
      "js": "javascript",
      "jsx": "javascript",
      "ts": "typescript",
      "tsx": "typescript",
      "py": "python",
      "go": "go",
      "sql": "sql",
      "json": "json",
      "yaml": "yaml",
      "yml": "yaml",
      "md": "markdown",
      "html": "html",
      "css": "css",
      "sh": "shell",
      "bash": "shell"
    };
    return languageMap[lang.toLowerCase()] || lang.toLowerCase();
  };
  const handleEditorChange = (value2) => {
    if (onChange) {
      onChange(value2 || "");
    }
  };
  const handleEditorDidMount = (editor, monaco) => {
    editorRef.current = editor;
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, () => {
    });
  };
  const monacoLanguage = getMonacoLanguage(language);
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `code-editor-container ${className}`, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center justify-content-between mb-2 px-2", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { style: { color: "var(--bs-gray-200)" }, children: [
        "Language: ",
        /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge", style: {
          backgroundColor: "#6366f1",
          color: "white",
          fontWeight: "500"
        }, children: monacoLanguage })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center gap-2", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { style: {
          fontSize: "0.875rem",
          color: "var(--bs-gray-200)",
          opacity: 0.9
        }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-type me-1" }),
          "Editor Size:"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "btn-group btn-group-sm", role: "group", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn ${fontSize === "xsmall" ? "btn-primary" : "btn-outline-secondary"}`,
              onClick: () => setFontSize("xsmall"),
              title: "Extra Small (12px)",
              style: { fontSize: "0.7rem", padding: "0.25rem 0.5rem" },
              children: "A⁻⁻"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn ${fontSize === "small" ? "btn-primary" : "btn-outline-secondary"}`,
              onClick: () => setFontSize("small"),
              title: "Small (14px)",
              style: { fontSize: "0.8rem", padding: "0.25rem 0.5rem" },
              children: "A⁻"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn ${fontSize === "medium" ? "btn-primary" : "btn-outline-secondary"}`,
              onClick: () => setFontSize("medium"),
              title: "Medium (16px) - Default",
              style: { fontSize: "0.875rem", padding: "0.25rem 0.5rem" },
              children: "A"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn ${fontSize === "large" ? "btn-primary" : "btn-outline-secondary"}`,
              onClick: () => setFontSize("large"),
              title: "Large (18px)",
              style: { fontSize: "1rem", padding: "0.25rem 0.5rem" },
              children: "A⁺"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn ${fontSize === "xlarge" ? "btn-primary" : "btn-outline-secondary"}`,
              onClick: () => setFontSize("xlarge"),
              title: "Extra Large (20px)",
              style: { fontSize: "1.125rem", padding: "0.25rem 0.5rem" },
              children: "A⁺⁺"
            }
          )
        ] })
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "editor-wrapper", style: {
      border: "1px solid var(--border-color, #dee2e6)",
      borderRadius: "0.375rem",
      overflow: "hidden"
    }, children: /* @__PURE__ */ jsxRuntimeExports.jsx(
      Ft,
      {
        height,
        defaultLanguage: monacoLanguage,
        language: monacoLanguage,
        theme: isDarkMode ? "vs-dark" : "light",
        value,
        onChange: handleEditorChange,
        onMount: handleEditorDidMount,
        options: {
          // Editor behavior
          readOnly,
          automaticLayout: true,
          scrollBeyondLastLine: false,
          wordWrap: "on",
          wrappingIndent: "indent",
          // Line numbers and folding
          lineNumbers: "on",
          folding: true,
          foldingStrategy: "indentation",
          showFoldingControls: "always",
          // Minimap
          minimap: {
            enabled: true,
            maxColumn: 120,
            renderCharacters: true,
            showSlider: "mouseover",
            side: "right"
          },
          // Font and rendering
          fontSize: fontSizes[fontSize],
          fontFamily: "'Fira Code', 'Cascadia Code', 'Courier New', monospace",
          fontLigatures: true,
          lineHeight: 1.6,
          letterSpacing: 0.5,
          renderWhitespace: "selection",
          renderLineHighlight: "all",
          // Indentation
          tabSize: 2,
          insertSpaces: true,
          detectIndentation: true,
          // Scrolling
          scrollbar: {
            vertical: "visible",
            horizontal: "visible",
            verticalScrollbarSize: 10,
            horizontalScrollbarSize: 10
          },
          // Suggestions and IntelliSense
          quickSuggestions: {
            other: true,
            comments: false,
            strings: false
          },
          suggestOnTriggerCharacters: true,
          acceptSuggestionOnCommitCharacter: true,
          acceptSuggestionOnEnter: "on",
          // Brackets and matching
          bracketPairColorization: {
            enabled: true
          },
          matchBrackets: "always",
          autoClosingBrackets: "always",
          autoClosingQuotes: "always",
          // Other useful features
          contextmenu: true,
          mouseWheelZoom: true,
          smoothScrolling: true,
          cursorBlinking: "smooth",
          cursorSmoothCaretAnimation: "on",
          formatOnPaste: true,
          formatOnType: false
        },
        loading: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex justify-content-center align-items-center", style: { height }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading editor..." }) }) })
      }
    ) })
  ] });
}
const analysisModesConfig = {
  preview: {
    name: "Preview",
    description: "Quick structural assessment",
    icon: "👁️",
    color: "primary",
    purpose: "Get a high-level overview of code structure and organization"
  },
  skim: {
    name: "Skim",
    description: "Understand abstractions",
    icon: "📋",
    color: "info",
    purpose: "Focus on interfaces, function signatures, and key workflows"
  },
  scan: {
    name: "Scan",
    description: "Find specific information",
    icon: "🔍",
    color: "success",
    purpose: "Search for specific patterns, functions, or code elements"
  },
  detailed: {
    name: "Detailed",
    description: "Deep algorithm understanding",
    icon: "🔬",
    color: "warning",
    purpose: "Line-by-line analysis of complex logic and algorithms"
  },
  critical: {
    name: "Critical",
    description: "Quality evaluation",
    icon: "⚠️",
    color: "danger",
    purpose: "Identify issues, security concerns, and improvement opportunities"
  }
};
function AnalysisModeSelector({ selectedMode, onModeSelect, onDetailsClick, disabled = false }) {
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "analysis-mode-selector mb-3", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { className: "mb-2", children: "Analysis Mode:" }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row g-2", children: Object.entries(analysisModesConfig).map(([mode, config2]) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-2 col-sm-4 col-6", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
      "div",
      {
        className: `frosted-card h-100 mode-card ${mode} ${selectedMode === mode ? "border-primary border-3" : ""} ${disabled ? "opacity-50" : ""}`,
        style: {
          cursor: disabled ? "not-allowed" : "pointer",
          transition: "all 0.2s",
          padding: "0.75rem",
          position: "relative"
        },
        children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "div",
            {
              onClick: () => !disabled && onModeSelect(mode),
              style: { cursor: disabled ? "not-allowed" : "pointer" },
              children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "fs-4 mb-1", children: config2.icon }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { className: `mb-1 ${selectedMode === mode ? "text-primary" : ""}`, children: config2.name }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("small", { style: {
                  color: "var(--bs-gray-300)",
                  opacity: 0.9
                }, children: config2.description })
              ] })
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-2 text-center", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
            "button",
            {
              className: "btn btn-sm btn-outline-primary btn-details",
              onClick: (e) => {
                e.stopPropagation();
                onDetailsClick && onDetailsClick(mode);
              },
              disabled,
              title: `View and customize ${config2.name} mode prompt`,
              style: { fontSize: "0.75rem", padding: "0.25rem 0.5rem" },
              children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-file-text me-1" }),
                "Details"
              ]
            }
          ) })
        ]
      }
    ) }, mode)) }),
    selectedMode && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-2", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { style: {
      color: "var(--bs-gray-200)",
      opacity: 0.95
    }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("strong", { children: [
        analysisModesConfig[selectedMode].name,
        ":"
      ] }),
      " ",
      analysisModesConfig[selectedMode].purpose
    ] }) })
  ] });
}
function AnalysisOutput({
  result,
  loading = false,
  error = null,
  mode = "preview",
  onRetry = null
}) {
  var _a;
  const [fontSize, setFontSize] = reactExports.useState("medium");
  const fontSizes = {
    xsmall: "0.875rem",
    // 14px
    small: "1.0rem",
    // 16px
    medium: "1.125rem",
    // 18px (default - middle size)
    large: "1.25rem",
    // 20px
    xlarge: "1.375rem"
    // 22px
  };
  if (loading) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "analysis-output frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex justify-content-center align-items-center flex-grow-1", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary mb-3", role: "status", style: { width: "3rem", height: "3rem" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Analyzing code..." }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "text-muted mb-2", children: [
        "Running ",
        ((_a = analysisModesConfig[mode]) == null ? void 0 : _a.name) || mode,
        " analysis..."
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted", children: "This may take a few moments depending on code complexity and selected model." })
    ] }) }) });
  }
  if (error) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "analysis-output frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger m-3", role: "alert", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "alert-heading", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle-fill me-2" }),
        "Analysis Failed"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-2", children: error }),
      onRetry && /* @__PURE__ */ jsxRuntimeExports.jsxs("button", { className: "btn btn-outline-danger btn-sm", onClick: onRetry, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-clockwise me-1" }),
        "Try Again"
      ] })
    ] }) });
  }
  if (!result) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "analysis-output frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex justify-content-center align-items-center text-center flex-grow-1", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "fs-1 mb-3", children: "📝" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { style: { color: "var(--bs-gray-200)" }, className: "mb-2", children: "No Analysis Yet" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { style: { color: "var(--bs-gray-300)" }, className: "mb-0", children: 'Add some code to the left pane, select an analysis mode, and click "Analyze" to get started.' })
    ] }) }) });
  }
  const getRawText = () => {
    if (typeof result === "string") {
      if (result.includes("<div") || result.includes("<span") || result.includes("<p")) {
        try {
          const parser = new DOMParser();
          const doc = parser.parseFromString(result, "text/html");
          let extractedText = "";
          const processNode = (node) => {
            if (node.nodeType === Node.TEXT_NODE) {
              const text = node.textContent;
              if (text.trim()) {
                extractedText += text;
              }
            } else if (node.nodeType === Node.ELEMENT_NODE) {
              const tagName = node.tagName.toLowerCase();
              if (["p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li", "br"].includes(tagName)) {
                if (extractedText && !extractedText.endsWith("\n")) {
                  extractedText += "\n";
                }
              }
              if (tagName === "pre" || tagName === "code") {
                extractedText += node.textContent;
                if (tagName === "pre") extractedText += "\n";
              } else {
                node.childNodes.forEach(processNode);
              }
              if (["p", "div", "h1", "h2", "h3", "h4", "h5", "h6", "li"].includes(tagName)) {
                if (!extractedText.endsWith("\n")) {
                  extractedText += "\n";
                }
              }
            }
          };
          doc.body.childNodes.forEach(processNode);
          extractedText = extractedText.replace(/\n{3,}/g, "\n\n");
          return extractedText.trim() || result;
        } catch (e) {
          console.warn("Failed to parse HTML, returning raw result:", e);
          return result;
        }
      }
      return result;
    }
    if (typeof result === "object") {
      if (result.result) return String(result.result);
      if (result.text) return String(result.text);
      if (result.content) return String(result.content);
      return JSON.stringify(result, null, 2);
    }
    return String(result);
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "analysis-output frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "border-bottom d-flex align-items-center justify-content-end px-3 py-2", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center gap-2", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { style: {
        fontSize: "0.875rem",
        color: "var(--bs-gray-200)",
        opacity: 0.9
      }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-type me-1" }),
        "Text Size:"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "btn-group btn-group-sm", role: "group", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn ${fontSize === "xsmall" ? "btn-primary" : "btn-outline-secondary"}`,
            onClick: () => setFontSize("xsmall"),
            title: "Extra Small (14px)",
            style: { fontSize: "0.7rem", padding: "0.25rem 0.5rem" },
            children: "A⁻⁻"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn ${fontSize === "small" ? "btn-primary" : "btn-outline-secondary"}`,
            onClick: () => setFontSize("small"),
            title: "Small (16px)",
            style: { fontSize: "0.8rem", padding: "0.25rem 0.5rem" },
            children: "A⁻"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn ${fontSize === "medium" ? "btn-primary" : "btn-outline-secondary"}`,
            onClick: () => setFontSize("medium"),
            title: "Medium (18px) - Default",
            style: { fontSize: "0.875rem", padding: "0.25rem 0.5rem" },
            children: "A"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn ${fontSize === "large" ? "btn-primary" : "btn-outline-secondary"}`,
            onClick: () => setFontSize("large"),
            title: "Large (20px)",
            style: { fontSize: "1rem", padding: "0.25rem 0.5rem" },
            children: "A⁺"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn ${fontSize === "xlarge" ? "btn-primary" : "btn-outline-secondary"}`,
            onClick: () => setFontSize("xlarge"),
            title: "Extra Large (22px)",
            style: { fontSize: "1.125rem", padding: "0.25rem 0.5rem" },
            children: "A⁺⁺"
          }
        )
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "p-4 flex-grow-1", style: { overflowY: "auto", overflowX: "hidden" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("pre", { className: "mb-0", style: {
      whiteSpace: "pre-wrap",
      wordBreak: "break-word",
      lineHeight: "1.6",
      fontSize: fontSizes[fontSize],
      fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
      overflowX: "hidden",
      textAlign: "left",
      margin: 0,
      padding: 0
    }, children: getRawText() }) })
  ] });
}
function FileTabs({
  files = [],
  activeFileId,
  onFileSelect,
  onFileClose,
  onFileAdd,
  onFileRename
}) {
  const handleCloseTab = (e, fileId) => {
    e.stopPropagation();
    const file = files.find((f2) => f2.id === fileId);
    if (file == null ? void 0 : file.hasUnsavedChanges) {
      const confirmed = window.confirm(
        `"${file.name}" has unsaved changes. Do you want to close it anyway?`
      );
      if (!confirmed) return;
    }
    if (onFileClose) {
      onFileClose(fileId);
    }
  };
  const handleTabClick = (fileId) => {
    if (onFileSelect) {
      onFileSelect(fileId);
    }
  };
  const handleTabDoubleClick = (fileId) => {
    if (onFileRename) {
      const file = files.find((f2) => f2.id === fileId);
      const newName = prompt("Enter new file name:", file.name);
      if (newName && newName.trim() && newName !== file.name) {
        onFileRename(fileId, newName.trim());
      }
    }
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "file-tabs-container", style: {
    display: "flex",
    alignItems: "center",
    backgroundColor: (isDarkMode) => isDarkMode ? "rgba(30, 33, 48, 0.95)" : "rgba(250, 250, 255, 0.95)",
    borderBottom: "2px solid rgba(99, 102, 241, 0.2)",
    padding: "0.25rem 0.5rem",
    gap: "0.25rem",
    overflowX: "auto",
    overflowY: "hidden",
    whiteSpace: "nowrap",
    maxWidth: "100%"
  }, children: [
    files.map((file) => /* @__PURE__ */ jsxRuntimeExports.jsxs(
      "div",
      {
        onClick: () => handleTabClick(file.id),
        onDoubleClick: () => handleTabDoubleClick(file.id),
        className: `file-tab ${file.id === activeFileId ? "active" : ""}`,
        style: {
          display: "inline-flex",
          alignItems: "center",
          gap: "0.5rem",
          padding: "0.5rem 0.75rem",
          borderRadius: "8px 8px 0 0",
          cursor: "pointer",
          transition: "all 0.2s ease",
          position: "relative",
          backgroundColor: file.id === activeFileId ? "rgba(99, 102, 241, 0.15)" : "transparent",
          border: file.id === activeFileId ? "2px solid rgba(99, 102, 241, 0.4)" : "2px solid transparent",
          borderBottom: "none",
          minWidth: "fit-content"
        },
        title: file.path || file.name,
        children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
            fontSize: "1rem",
            opacity: 0.8
          }, children: getFileIcon$1(file.language) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
            fontSize: "0.875rem",
            fontWeight: file.id === activeFileId ? "600" : "400",
            color: file.id === activeFileId ? "#6366f1" : "inherit",
            maxWidth: "150px",
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap"
          }, children: file.name }),
          file.hasUnsavedChanges && /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
            width: "6px",
            height: "6px",
            borderRadius: "50%",
            backgroundColor: "#ec4899",
            flexShrink: 0
          }, title: "Unsaved changes" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              onClick: (e) => handleCloseTab(e, file.id),
              className: "btn-close-tab",
              style: {
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                width: "20px",
                height: "20px",
                border: "none",
                background: "transparent",
                cursor: "pointer",
                borderRadius: "4px",
                padding: 0,
                opacity: 0.6,
                transition: "all 0.2s ease",
                fontSize: "1rem",
                color: "inherit"
              },
              onMouseEnter: (e) => {
                e.currentTarget.style.opacity = "1";
                e.currentTarget.style.backgroundColor = "rgba(239, 68, 68, 0.2)";
              },
              onMouseLeave: (e) => {
                e.currentTarget.style.opacity = "0.6";
                e.currentTarget.style.backgroundColor = "transparent";
              },
              title: "Close (Ctrl+W)",
              children: "×"
            }
          )
        ]
      },
      file.id
    )),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      "button",
      {
        onClick: onFileAdd,
        className: "btn-add-file",
        style: {
          display: "inline-flex",
          alignItems: "center",
          justifyContent: "center",
          padding: "0.5rem",
          border: "none",
          background: "transparent",
          cursor: "pointer",
          borderRadius: "4px",
          transition: "all 0.2s ease",
          fontSize: "1.25rem",
          color: "#6366f1",
          opacity: 0.7,
          minWidth: "fit-content"
        },
        onMouseEnter: (e) => {
          e.currentTarget.style.opacity = "1";
          e.currentTarget.style.backgroundColor = "rgba(99, 102, 241, 0.1)";
        },
        onMouseLeave: (e) => {
          e.currentTarget.style.opacity = "0.7";
          e.currentTarget.style.backgroundColor = "transparent";
        },
        title: "Add new file (Ctrl+N)",
        children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle" })
      }
    ),
    files.length > 1 && /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { style: {
      fontSize: "0.75rem",
      color: "rgba(99, 102, 241, 0.6)",
      marginLeft: "auto",
      padding: "0.25rem 0.5rem",
      borderRadius: "12px",
      backgroundColor: "rgba(99, 102, 241, 0.1)",
      fontWeight: "500",
      minWidth: "fit-content"
    }, children: [
      files.length,
      " ",
      files.length === 1 ? "file" : "files"
    ] })
  ] });
}
function getFileIcon$1(language) {
  const iconMap = {
    "javascript": "📜",
    "typescript": "📘",
    "python": "🐍",
    "go": "🔷",
    "java": "☕",
    "rust": "🦀",
    "c": "⚙️",
    "cpp": "⚙️",
    "csharp": "#️⃣",
    "sql": "🗄️",
    "html": "🌐",
    "css": "🎨",
    "json": "📋",
    "yaml": "📝",
    "markdown": "📖",
    "shell": "💻",
    "bash": "💻",
    "php": "🐘",
    "ruby": "💎",
    "swift": "🦅",
    "kotlin": "🟣"
  };
  return iconMap[language == null ? void 0 : language.toLowerCase()] || "📄";
}
function FileTreeBrowser({
  treeData = [],
  selectedFiles = [],
  onFileSelect,
  onFilesAnalyze,
  loading = false
}) {
  const [expandedFolders, setExpandedFolders] = reactExports.useState(/* @__PURE__ */ new Set());
  const [searchQuery, setSearchQuery] = reactExports.useState("");
  const toggleFolder = (path) => {
    setExpandedFolders((prev) => {
      const newSet = new Set(prev);
      if (newSet.has(path)) {
        newSet.delete(path);
      } else {
        newSet.add(path);
      }
      return newSet;
    });
  };
  const handleFileClick = (file, event) => {
    if (onFileSelect) {
      const isMultiSelect = event.ctrlKey || event.metaKey;
      onFileSelect(file, isMultiSelect);
    }
  };
  const handleAnalyzeSelected = () => {
    if (onFilesAnalyze && selectedFiles.length > 0) {
      onFilesAnalyze(selectedFiles);
    }
  };
  const filterTree = (nodes, query) => {
    if (!query) return nodes;
    const lowerQuery = query.toLowerCase();
    return nodes.filter((node) => {
      if (node.type === "file") {
        return node.name.toLowerCase().includes(lowerQuery) || node.path.toLowerCase().includes(lowerQuery);
      } else {
        const filteredChildren = filterTree(node.children || [], query);
        return filteredChildren.length > 0 || node.name.toLowerCase().includes(lowerQuery);
      }
    }).map((node) => {
      if (node.type === "directory") {
        return {
          ...node,
          children: filterTree(node.children || [], query)
        };
      }
      return node;
    });
  };
  const filteredTree = filterTree(treeData, searchQuery);
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "file-tree-browser frosted-card", style: {
    height: "100%",
    display: "flex",
    flexDirection: "column",
    overflow: "hidden"
  }, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "p-3 border-bottom", style: {
      borderBottomColor: "rgba(99, 102, 241, 0.2)"
    }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center justify-content-between mb-2", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-0", style: { color: "#6366f1" }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder2-open me-2" }),
          "File Explorer"
        ] }),
        selectedFiles.length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "badge", style: {
          backgroundColor: "rgba(99, 102, 241, 0.2)",
          color: "#6366f1"
        }, children: [
          selectedFiles.length,
          " selected"
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "input-group input-group-sm", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "input-group-text", children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-search" }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "input",
          {
            type: "text",
            className: "form-control",
            placeholder: "Search files...",
            value: searchQuery,
            onChange: (e) => setSearchQuery(e.target.value)
          }
        ),
        searchQuery && /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-outline-secondary",
            onClick: () => setSearchQuery(""),
            title: "Clear search",
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-x" })
          }
        )
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "flex-grow-1", style: {
      overflowY: "auto",
      overflowX: "hidden",
      padding: "0.5rem"
    }, children: filteredTree.length === 0 ? /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center text-muted p-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder-x", style: { fontSize: "2rem" } }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mt-2 mb-0", children: searchQuery ? "No files match your search" : "No files found" })
    ] }) : /* @__PURE__ */ jsxRuntimeExports.jsx(
      TreeNodeRenderer,
      {
        nodes: filteredTree,
        expandedFolders,
        selectedFiles,
        onToggleFolder: toggleFolder,
        onFileClick: handleFileClick,
        level: 0
      }
    ) }),
    selectedFiles.length > 0 && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "p-3 border-top", style: {
      borderTopColor: "rgba(99, 102, 241, 0.2)",
      backgroundColor: "rgba(99, 102, 241, 0.05)"
    }, children: /* @__PURE__ */ jsxRuntimeExports.jsx(
      "button",
      {
        className: "btn btn-primary w-100",
        onClick: handleAnalyzeSelected,
        disabled: loading,
        children: loading ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-2" }),
          "Analyzing..."
        ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lightning-charge me-2" }),
          "Analyze ",
          selectedFiles.length,
          " File",
          selectedFiles.length !== 1 ? "s" : ""
        ] })
      }
    ) })
  ] });
}
function TreeNodeRenderer({
  nodes,
  expandedFolders,
  selectedFiles,
  onToggleFolder,
  onFileClick,
  level
}) {
  return /* @__PURE__ */ jsxRuntimeExports.jsx(jsxRuntimeExports.Fragment, { children: nodes.map((node, index2) => {
    const isExpanded = expandedFolders.has(node.path);
    const isSelected = selectedFiles.some((f2) => f2.path === node.path);
    const indent = level * 1.25;
    if (node.type === "directory") {
      return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "div",
          {
            onClick: () => onToggleFolder(node.path),
            style: {
              paddingLeft: `${indent}rem`,
              paddingTop: "0.375rem",
              paddingBottom: "0.375rem",
              cursor: "pointer",
              borderRadius: "0.25rem",
              transition: "all 0.15s ease",
              display: "flex",
              alignItems: "center",
              gap: "0.5rem"
            },
            onMouseEnter: (e) => {
              e.currentTarget.style.backgroundColor = "rgba(99, 102, 241, 0.08)";
            },
            onMouseLeave: (e) => {
              e.currentTarget.style.backgroundColor = "transparent";
            },
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "i",
                {
                  className: `bi bi-chevron-${isExpanded ? "down" : "right"}`,
                  style: { fontSize: "0.75rem", color: "#6366f1" }
                }
              ),
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "i",
                {
                  className: `bi bi-folder${isExpanded ? "-open" : ""}`,
                  style: { color: "#ec4899" }
                }
              ),
              /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { fontSize: "0.875rem", fontWeight: "500" }, children: node.name }),
              node.children && /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge badge-sm", style: {
                backgroundColor: "rgba(99, 102, 241, 0.15)",
                color: "#6366f1",
                fontSize: "0.7rem",
                marginLeft: "auto"
              }, children: node.children.length })
            ]
          }
        ),
        isExpanded && node.children && /* @__PURE__ */ jsxRuntimeExports.jsx(
          TreeNodeRenderer,
          {
            nodes: node.children,
            expandedFolders,
            selectedFiles,
            onToggleFolder,
            onFileClick,
            level: level + 1
          }
        )
      ] }, node.path || index2);
    } else {
      return /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "div",
        {
          onClick: (e) => onFileClick(node, e),
          style: {
            paddingLeft: `${indent + 0.5}rem`,
            paddingTop: "0.375rem",
            paddingBottom: "0.375rem",
            cursor: "pointer",
            borderRadius: "0.25rem",
            transition: "all 0.15s ease",
            display: "flex",
            alignItems: "center",
            gap: "0.5rem",
            backgroundColor: isSelected ? "rgba(99, 102, 241, 0.15)" : "transparent"
          },
          onMouseEnter: (e) => {
            if (!isSelected) {
              e.currentTarget.style.backgroundColor = "rgba(99, 102, 241, 0.08)";
            }
          },
          onMouseLeave: (e) => {
            if (!isSelected) {
              e.currentTarget.style.backgroundColor = "transparent";
            }
          },
          children: [
            isSelected && /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-check-circle-fill", style: { color: "#6366f1", fontSize: "0.875rem" } }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: { fontSize: "1rem" }, children: getFileIcon(node.name) }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { style: {
              fontSize: "0.875rem",
              color: isSelected ? "#6366f1" : "inherit",
              fontWeight: isSelected ? "500" : "400"
            }, children: node.name })
          ]
        },
        node.path || index2
      );
    }
  }) });
}
function getFileIcon(filename) {
  const ext = filename.split(".").pop().toLowerCase();
  const iconMap = {
    // Programming languages
    "js": "📜",
    "jsx": "⚛️",
    "ts": "📘",
    "tsx": "⚛️",
    "py": "🐍",
    "go": "🔷",
    "java": "☕",
    "rs": "🦀",
    "c": "⚙️",
    "cpp": "⚙️",
    "cs": "#️⃣",
    "php": "🐘",
    "rb": "💎",
    "swift": "🦅",
    "kt": "🟣",
    // Web
    "html": "🌐",
    "css": "🎨",
    "scss": "🎨",
    "vue": "💚",
    // Config/Data
    "json": "📋",
    "yaml": "📝",
    "yml": "📝",
    "xml": "📄",
    "toml": "⚙️",
    // Documentation
    "md": "📖",
    "txt": "📃",
    // Database
    "sql": "🗄️",
    // Shell
    "sh": "💻",
    "bash": "💻",
    // Other
    "gitignore": "🚫",
    "dockerignore": "🚫",
    "dockerfile": "🐳"
  };
  return iconMap[ext] || "📄";
}
function RepoImportModal({ show, onClose, onSuccess }) {
  const [githubUrl, setGithubUrl] = reactExports.useState("");
  const [branch, setBranch] = reactExports.useState("main");
  const [importMode, setImportMode] = reactExports.useState("quick");
  const [loading, setLoading] = reactExports.useState(false);
  const [error, setError] = reactExports.useState(null);
  const [validationError, setValidationError] = reactExports.useState(null);
  const validateGithubUrl = (url) => {
    const githubPattern = /^(https?:\/\/)?(www\.)?github\.com\/[\w-]+\/[\w.-]+\/?$/;
    if (!url.trim()) {
      return "GitHub URL is required";
    }
    if (!githubPattern.test(url.trim())) {
      return "Invalid GitHub URL format. Expected: github.com/owner/repo";
    }
    return null;
  };
  const parseGithubUrl = (url) => {
    const cleaned = url.trim().replace(/^(https?:\/\/)?(www\.)?/, "").replace(/\/$/, "");
    const parts = cleaned.split("/");
    if (parts.length >= 3 && parts[0] === "github.com") {
      return {
        owner: parts[1],
        repo: parts[2],
        fullUrl: cleaned
      };
    }
    return null;
  };
  const handleUrlChange = (e) => {
    const url = e.target.value;
    setGithubUrl(url);
    setValidationError(null);
    setError(null);
  };
  const handleSubmit = async (e) => {
    e.preventDefault();
    const urlError = validateGithubUrl(githubUrl);
    if (urlError) {
      setValidationError(urlError);
      return;
    }
    const parsed = parseGithubUrl(githubUrl);
    if (!parsed) {
      setValidationError("Could not parse GitHub URL");
      return;
    }
    setLoading(true);
    setError(null);
    setValidationError(null);
    try {
      let result;
      if (importMode === "quick") {
        result = await reviewApi.githubQuickScan(parsed.fullUrl, branch);
      } else {
        result = await reviewApi.githubGetTree(parsed.fullUrl, branch);
      }
      onSuccess({
        mode: importMode,
        data: result,
        repoInfo: {
          owner: parsed.owner,
          repo: parsed.repo,
          branch,
          url: parsed.fullUrl
        }
      });
      handleClose();
    } catch (err) {
      console.error("GitHub import failed:", err);
      if (err.message.includes("404")) {
        setError("Repository not found. Check the URL and branch name.");
      } else if (err.message.includes("403")) {
        setError("Access denied. You may need to authenticate for private repositories.");
      } else if (err.message.includes("Authentication required")) {
        setError("Please log in to access GitHub repositories.");
      } else {
        setError(err.message || "Failed to import repository. Please try again.");
      }
    } finally {
      setLoading(false);
    }
  };
  const handleClose = () => {
    if (!loading) {
      setGithubUrl("");
      setBranch("main");
      setImportMode("quick");
      setError(null);
      setValidationError(null);
      onClose();
    }
  };
  if (!show) return null;
  return /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      "div",
      {
        className: "modal-backdrop fade show",
        onClick: handleClose,
        style: { zIndex: 1040 }
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      "div",
      {
        className: "modal fade show d-block",
        tabIndex: "-1",
        role: "dialog",
        style: { zIndex: 1050 },
        children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-dialog-centered modal-lg", role: "document", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-content", style: {
          backgroundColor: "var(--bs-body-bg)",
          border: "2px solid rgba(99, 102, 241, 0.3)",
          borderRadius: "12px",
          boxShadow: "0 10px 40px rgba(0, 0, 0, 0.3)"
        }, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-header", style: {
            borderBottom: "1px solid rgba(99, 102, 241, 0.2)",
            padding: "1.5rem"
          }, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "modal-title", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-github me-2 text-primary" }),
              "Import from GitHub"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                type: "button",
                className: "btn-close",
                onClick: handleClose,
                disabled: loading,
                "aria-label": "Close"
              }
            )
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-body", style: { padding: "1.5rem" }, children: /* @__PURE__ */ jsxRuntimeExports.jsxs("form", { onSubmit: handleSubmit, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "githubUrl", className: "form-label", children: /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "GitHub Repository URL" }) }),
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "input",
                {
                  type: "text",
                  id: "githubUrl",
                  className: `form-control ${validationError ? "is-invalid" : ""}`,
                  placeholder: "github.com/owner/repo or https://github.com/owner/repo",
                  value: githubUrl,
                  onChange: handleUrlChange,
                  disabled: loading,
                  autoFocus: true,
                  style: {
                    padding: "0.75rem",
                    fontSize: "1rem",
                    borderRadius: "8px"
                  }
                }
              ),
              validationError && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "invalid-feedback d-block", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-circle me-1" }),
                validationError
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "form-text text-muted", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-1" }),
                "Enter the repository URL (e.g., github.com/golang/go)"
              ] })
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "branch", className: "form-label", children: /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Branch" }) }),
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "input",
                {
                  type: "text",
                  id: "branch",
                  className: "form-control",
                  placeholder: "main",
                  value: branch,
                  onChange: (e) => setBranch(e.target.value),
                  disabled: loading,
                  style: {
                    padding: "0.75rem",
                    fontSize: "1rem",
                    borderRadius: "8px"
                  }
                }
              ),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "form-text text-muted", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-1" }),
                "Default: main (or master for older repos)"
              ] })
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-4", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "form-label d-block", children: /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Import Mode" }) }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "form-check mb-2 p-3", style: {
                backgroundColor: importMode === "quick" ? "rgba(99, 102, 241, 0.1)" : "transparent",
                border: `2px solid ${importMode === "quick" ? "rgba(99, 102, 241, 0.4)" : "rgba(99, 102, 241, 0.2)"}`,
                borderRadius: "8px",
                cursor: "pointer",
                transition: "all 0.2s"
              }, onClick: () => !loading && setImportMode("quick"), children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "input",
                  {
                    className: "form-check-input",
                    type: "radio",
                    name: "importMode",
                    id: "quickScan",
                    value: "quick",
                    checked: importMode === "quick",
                    onChange: (e) => setImportMode(e.target.value),
                    disabled: loading,
                    style: { cursor: "pointer" }
                  }
                ),
                /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { className: "form-check-label ms-2", htmlFor: "quickScan", style: { cursor: "pointer" }, children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Quick Repo Scan" }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-success ms-2", children: "Fast" }),
                  /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-muted mt-1", style: { fontSize: "0.9rem" }, children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lightning-fill me-1" }),
                    "Fetches 5-8 core files (README, package files, entry points) in ~2 seconds. Best for quick project assessment."
                  ] })
                ] })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "form-check p-3", style: {
                backgroundColor: importMode === "full" ? "rgba(99, 102, 241, 0.1)" : "transparent",
                border: `2px solid ${importMode === "full" ? "rgba(99, 102, 241, 0.4)" : "rgba(99, 102, 241, 0.2)"}`,
                borderRadius: "8px",
                cursor: "pointer",
                transition: "all 0.2s"
              }, onClick: () => !loading && setImportMode("full"), children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "input",
                  {
                    className: "form-check-input",
                    type: "radio",
                    name: "importMode",
                    id: "fullBrowser",
                    value: "full",
                    checked: importMode === "full",
                    onChange: (e) => setImportMode(e.target.value),
                    disabled: loading,
                    style: { cursor: "pointer" }
                  }
                ),
                /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { className: "form-check-label ms-2", htmlFor: "fullBrowser", style: { cursor: "pointer" }, children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Full Repository Browser" }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-primary ms-2", children: "Complete" }),
                  /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-muted mt-1", style: { fontSize: "0.9rem" }, children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder-fill me-1" }),
                    "Fetches complete file tree structure. Explore all files and folders. Files loaded on-demand when selected."
                  ] })
                ] })
              ] })
            ] }),
            error && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger d-flex align-items-center", role: "alert", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle-fill me-2" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { children: error })
            ] }),
            loading && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-info d-flex align-items-center", role: "alert", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border spinner-border-sm me-2", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("div", { children: importMode === "quick" ? "Fetching core files from repository..." : "Fetching repository structure..." })
            ] })
          ] }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-footer", style: {
            borderTop: "1px solid rgba(99, 102, 241, 0.2)",
            padding: "1rem 1.5rem"
          }, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                type: "button",
                className: "btn btn-outline-secondary",
                onClick: handleClose,
                disabled: loading,
                children: "Cancel"
              }
            ),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                type: "button",
                className: "btn btn-primary",
                onClick: handleSubmit,
                disabled: loading || !githubUrl.trim(),
                children: loading ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-2", role: "status", "aria-hidden": "true" }),
                  "Importing..."
                ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-download me-2" }),
                  "Import Repository"
                ] })
              }
            )
          ] })
        ] }) })
      }
    )
  ] });
}
const MAX_PROMPT_LENGTH = 2e3;
const ERROR_MESSAGES = {
  LOAD_FAILED: "Failed to load prompt. Please try again.",
  SAVE_FAILED: "Failed to save prompt. Please try again.",
  RESET_FAILED: "Failed to reset prompt. Please try again.",
  VALIDATION_REQUIRED_VARS: "Prompt must contain all required variables",
  VALIDATION_MAX_LENGTH: `Prompt cannot exceed ${MAX_PROMPT_LENGTH} characters`
};
const MODE_VARIABLES = {
  preview: [
    { name: "{{code}}", description: "Code to analyze", required: true }
  ],
  skim: [
    { name: "{{code}}", description: "Code to analyze", required: true }
  ],
  scan: [
    { name: "{{code}}", description: "Code to analyze", required: true },
    { name: "{{query}}", description: "Search query", required: true }
  ],
  detailed: [
    { name: "{{code}}", description: "Code to analyze", required: true }
  ],
  critical: [
    { name: "{{code}}", description: "Code to analyze", required: true }
  ]
};
function PromptEditorModal({
  isOpen,
  onClose,
  mode,
  userLevel = "intermediate",
  outputMode = "quick"
}) {
  const { isDarkMode } = useTheme();
  const [promptText, setPromptText] = reactExports.useState("");
  const [originalPrompt, setOriginalPrompt] = reactExports.useState("");
  const [isCustom, setIsCustom] = reactExports.useState(false);
  const [canReset, setCanReset] = reactExports.useState(false);
  const [loading, setLoading] = reactExports.useState(false);
  const [error, setError] = reactExports.useState(null);
  const [validationError, setValidationError] = reactExports.useState(null);
  const [showResetConfirm, setShowResetConfirm] = reactExports.useState(false);
  const [variablesPanelExpanded, setVariablesPanelExpanded] = reactExports.useState(false);
  const variables = reactExports.useMemo(() => MODE_VARIABLES[mode] || MODE_VARIABLES.preview, [mode]);
  reactExports.useEffect(() => {
    if (isOpen) {
      loadPrompt();
    }
  }, [isOpen, mode, userLevel, outputMode]);
  const loadPrompt = async () => {
    setLoading(true);
    setError(null);
    setValidationError(null);
    try {
      const response = await reviewApi.getPrompt(mode, userLevel, outputMode);
      setPromptText(response.prompt_text);
      setOriginalPrompt(response.prompt_text);
      setIsCustom(response.is_custom || false);
      setCanReset(response.can_reset || false);
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.LOAD_FAILED);
    } finally {
      setLoading(false);
    }
  };
  const validatePrompt = (text) => {
    if (text.length > MAX_PROMPT_LENGTH) {
      return ERROR_MESSAGES.VALIDATION_MAX_LENGTH;
    }
    const requiredVars = variables.filter((v2) => v2.required);
    const missingVars = [];
    for (const variable of requiredVars) {
      if (!text.includes(variable.name)) {
        missingVars.push(variable.name);
      }
    }
    if (missingVars.length > 0) {
      return `${ERROR_MESSAGES.VALIDATION_REQUIRED_VARS}: ${missingVars.join(", ")}`;
    }
    return null;
  };
  const handleSave = async () => {
    setValidationError(null);
    const validationErr = validatePrompt(promptText);
    if (validationErr) {
      setValidationError(validationErr);
      return;
    }
    setLoading(true);
    setError(null);
    try {
      await reviewApi.savePrompt({
        mode,
        user_level: userLevel,
        output_mode: outputMode,
        prompt_text: promptText,
        variables: variables.map((v2) => v2.name)
      });
      await loadPrompt();
      onClose();
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.SAVE_FAILED);
    } finally {
      setLoading(false);
    }
  };
  const handleFactoryReset = async () => {
    setLoading(true);
    setError(null);
    setValidationError(null);
    try {
      await reviewApi.resetPrompt(mode, userLevel, outputMode);
      await loadPrompt();
      setShowResetConfirm(false);
    } catch (err) {
      setError(err.message || ERROR_MESSAGES.RESET_FAILED);
    } finally {
      setLoading(false);
    }
  };
  const handleCancel = () => {
    setPromptText(originalPrompt);
    setValidationError(null);
    setError(null);
    onClose();
  };
  const handlePromptChange = (e) => {
    setPromptText(e.target.value);
    setValidationError(null);
  };
  const highlightVariables = (text) => {
    const parts = [];
    let lastIndex = 0;
    const regex = /\{\{[\w_]+\}\}/g;
    let match;
    while ((match = regex.exec(text)) !== null) {
      if (match.index > lastIndex) {
        parts.push({
          type: "text",
          content: text.substring(lastIndex, match.index)
        });
      }
      parts.push({
        type: "variable",
        content: match[0]
      });
      lastIndex = match.index + match[0].length;
    }
    if (lastIndex < text.length) {
      parts.push({
        type: "text",
        content: text.substring(lastIndex)
      });
    }
    return parts;
  };
  if (!isOpen) return null;
  return /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal show d-block prompt-editor-modal", tabIndex: "-1", role: "dialog", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-lg", role: "document", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-content ${isDarkMode ? "bg-dark text-light" : ""}`, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-header ${isDarkMode ? "border-secondary" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "modal-title", children: [
          "Edit Prompt - ",
          mode.charAt(0).toUpperCase() + mode.slice(1),
          " Mode"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "ms-auto d-flex align-items-center gap-2", children: [
          isCustom ? /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-primary badge-custom", children: "Custom" }) : /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-secondary badge-default", children: "System Default" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn-close ${isDarkMode ? "btn-close-white" : ""}`,
              onClick: handleCancel,
              disabled: loading
            }
          )
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-body ${isDarkMode ? "bg-dark" : ""}`, children: [
        error && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger alert-dismissible fade show", role: "alert", children: [
          error,
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn-close",
              onClick: () => setError(null)
            }
          )
        ] }),
        validationError && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-warning validation-error", role: "alert", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle me-2" }),
          validationError
        ] }),
        loading ? /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "text-center py-5", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }) }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `card mb-3 variable-reference-panel ${isDarkMode ? "bg-dark border-secondary" : ""}`, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs(
              "div",
              {
                className: `card-header d-flex justify-content-between align-items-center ${isDarkMode ? "bg-dark border-secondary text-light" : ""}`,
                style: { cursor: "pointer" },
                onClick: () => setVariablesPanelExpanded(!variablesPanelExpanded),
                children: [
                  /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-code-square me-2" }),
                    "Available Variables"
                  ] }),
                  /* @__PURE__ */ jsxRuntimeExports.jsx(
                    "button",
                    {
                      className: `btn btn-sm btn-link expand-btn ${isDarkMode ? "text-light" : ""}`,
                      type: "button",
                      children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-chevron-${variablesPanelExpanded ? "up" : "down"}` })
                    }
                  )
                ]
              }
            ),
            variablesPanelExpanded && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: `card-body ${isDarkMode ? "bg-dark text-light" : ""}`, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: variables.map((variable, index2) => /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col-md-6 mb-2", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-start", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "text-primary me-2", children: variable.name }),
                variable.required && /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-danger badge-sm", children: "Required" })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted", children: variable.description })
            ] }, index2)) }) })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "prompt-text", className: "form-label", children: "Prompt Template" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "textarea",
              {
                id: "prompt-text",
                name: "prompt",
                className: `form-control font-monospace ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`,
                rows: "12",
                value: promptText,
                onChange: handlePromptChange,
                placeholder: "Enter your prompt template...",
                style: { fontSize: "14px" }
              }
            ),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between mt-1", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted", children: [
                "Use variables like ",
                /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "{{code}}" }),
                " in your prompt"
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted character-count", children: [
                promptText.length,
                " characters"
              ] })
            ] })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: `card ${isDarkMode ? "bg-dark border-secondary" : "bg-light"}`, children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `card-body ${isDarkMode ? "text-light" : ""}`, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { className: "card-title", children: "Preview" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "font-monospace", style: { fontSize: "13px", whiteSpace: "pre-wrap" }, children: highlightVariables(promptText).map((part, index2) => part.type === "variable" ? /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-primary fw-bold", children: part.content }, index2) : /* @__PURE__ */ jsxRuntimeExports.jsx("span", { children: part.content }, index2)) })
          ] }) })
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-footer d-flex justify-content-between ${isDarkMode ? "border-secondary" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { children: canReset && /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            type: "button",
            className: "btn btn-outline-danger btn-factory-reset",
            onClick: () => setShowResetConfirm(true),
            disabled: loading,
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-counterclockwise me-1" }),
              "Factory Reset"
            ]
          }
        ) }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-2", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn btn-secondary btn-cancel",
              onClick: handleCancel,
              disabled: loading,
              children: "Cancel"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn btn-primary btn-save",
              onClick: handleSave,
              disabled: loading || !promptText.trim(),
              children: loading ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-1", role: "status", "aria-hidden": "true" }),
                "Saving..."
              ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-save me-1" }),
                "Save Custom Prompt"
              ] })
            }
          )
        ] })
      ] })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-backdrop show" }),
    showResetConfirm && /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal show d-block", tabIndex: "-1", role: "dialog", style: { zIndex: 1060 }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-dialog-centered", role: "document", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-content ${isDarkMode ? "bg-dark text-light" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-header ${isDarkMode ? "border-secondary" : ""}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "modal-title", children: "Confirm Factory Reset" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: `btn-close ${isDarkMode ? "btn-close-white" : ""}`,
              onClick: () => setShowResetConfirm(false),
              disabled: loading
            }
          )
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-body ${isDarkMode ? "bg-dark" : ""}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("p", { children: "Are you sure you want to reset this prompt to the system default?" }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { className: "text-muted mb-0", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-1" }),
            "This will permanently delete your custom prompt for this mode."
          ] })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-footer ${isDarkMode ? "border-secondary" : ""}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn btn-secondary",
              onClick: () => setShowResetConfirm(false),
              disabled: loading,
              children: "Cancel"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn btn-danger confirm-reset-btn",
              onClick: handleFactoryReset,
              disabled: loading,
              children: loading ? "Resetting..." : "Yes, Reset to Default"
            }
          )
        ] })
      ] }) }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-backdrop show", style: { zIndex: 1055 } })
    ] })
  ] });
}
const defaultCode = `// Example JavaScript function to analyze
function fibonacci(n) {
  if (n <= 1) {
    return n;
  }
  return fibonacci(n - 1) + fibonacci(n - 2);
}

// TODO: Optimize for large values of n
// Consider using memoization or iterative approach
console.log(fibonacci(10));
`;
function ReviewPage() {
  const { user, logout } = useAuth();
  const { isDarkMode, toggleTheme } = useTheme();
  const [files, setFiles] = reactExports.useState([
    {
      id: "file_1",
      name: "example.js",
      language: "javascript",
      content: defaultCode,
      hasUnsavedChanges: false,
      path: null
    }
  ]);
  const [activeFileId, setActiveFileId] = reactExports.useState("file_1");
  const activeFile = files.find((f2) => f2.id === activeFileId);
  const code = (activeFile == null ? void 0 : activeFile.content) || "";
  const [selectedMode, setSelectedMode] = reactExports.useState("preview");
  const [selectedModel, setSelectedModel] = reactExports.useState("");
  const [scanQuery, setScanQuery] = reactExports.useState("");
  const [analysisResult, setAnalysisResult] = reactExports.useState(null);
  const [loading, setLoading] = reactExports.useState(false);
  const [error, setError] = reactExports.useState(null);
  const [sessionId] = reactExports.useState(() => `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`);
  const [userMode, setUserMode] = reactExports.useState("intermediate");
  const [outputMode, setOutputMode] = reactExports.useState("quick");
  const [showPromptEditor, setShowPromptEditor] = reactExports.useState(false);
  const [promptEditorMode, setPromptEditorMode] = reactExports.useState("preview");
  const [showImportModal, setShowImportModal] = reactExports.useState(false);
  const [repoInfo, setRepoInfo] = reactExports.useState(null);
  const [treeData, setTreeData] = reactExports.useState(null);
  const [showTree, setShowTree] = reactExports.useState(false);
  const [selectedTreeFiles, setSelectedTreeFiles] = reactExports.useState([]);
  const codeEditorRef = reactExports.useRef(null);
  const handleCodeChange = (newCode) => {
    setFiles((prevFiles) => prevFiles.map(
      (file) => file.id === activeFileId ? { ...file, content: newCode, hasUnsavedChanges: true } : file
    ));
  };
  const handleFileSelect = (fileId) => {
    setActiveFileId(fileId);
    setAnalysisResult(null);
    setError(null);
  };
  const handleFileClose = (fileId) => {
    const fileIndex = files.findIndex((f2) => f2.id === fileId);
    const newFiles = files.filter((f2) => f2.id !== fileId);
    if (newFiles.length === 0) {
      const newFileId = `file_${Date.now()}`;
      setFiles([{
        id: newFileId,
        name: "untitled.js",
        language: "javascript",
        content: "",
        hasUnsavedChanges: false,
        path: null
      }]);
      setActiveFileId(newFileId);
      return;
    }
    setFiles(newFiles);
    if (fileId === activeFileId) {
      const newActiveIndex = fileIndex > 0 ? fileIndex - 1 : 0;
      setActiveFileId(newFiles[newActiveIndex].id);
    }
  };
  const handleFileAdd = () => {
    const newFileId = `file_${Date.now()}`;
    const newFile = {
      id: newFileId,
      name: `untitled-${files.length + 1}.js`,
      language: "javascript",
      content: "// New file\n",
      hasUnsavedChanges: false,
      path: null
    };
    setFiles((prevFiles) => [...prevFiles, newFile]);
    setActiveFileId(newFileId);
    setAnalysisResult(null);
    setError(null);
  };
  const handleFileRename = (fileId, newName) => {
    const extension = newName.split(".").pop().toLowerCase();
    const languageMap = {
      "js": "javascript",
      "jsx": "javascript",
      "ts": "typescript",
      "tsx": "typescript",
      "py": "python",
      "go": "go",
      "rs": "rust",
      "java": "java",
      "c": "c",
      "cpp": "cpp",
      "cs": "csharp",
      "sql": "sql",
      "html": "html",
      "css": "css",
      "json": "json",
      "yaml": "yaml",
      "yml": "yaml",
      "md": "markdown",
      "sh": "shell",
      "bash": "shell"
    };
    const detectedLanguage = languageMap[extension] || "javascript";
    setFiles((prevFiles) => prevFiles.map(
      (file) => file.id === fileId ? { ...file, name: newName, language: detectedLanguage, hasUnsavedChanges: true } : file
    ));
  };
  const handleDetailsClick = (mode) => {
    setPromptEditorMode(mode);
    setShowPromptEditor(true);
  };
  const handlePromptEditorClose = () => {
    setShowPromptEditor(false);
  };
  const handleGitHubImportSuccess = (importData) => {
    const { mode, data, repoInfo: repo } = importData;
    setRepoInfo(repo);
    if (mode === "quick") {
      console.log("Quick scan data:", data);
      setFiles([]);
      const newFiles = [];
      if (data.readme) {
        newFiles.push({
          id: `file_readme_${Date.now()}`,
          name: "README.md",
          language: "markdown",
          content: data.readme,
          hasUnsavedChanges: false,
          path: "README.md",
          repoInfo: repo
        });
      }
      if (data.entry_points && Array.isArray(data.entry_points)) {
        data.entry_points.forEach((entry, idx) => {
          if (entry.content) {
            const fileName = entry.path.split("/").pop();
            const extension = fileName.split(".").pop().toLowerCase();
            const languageMap = {
              "js": "javascript",
              "jsx": "javascript",
              "ts": "typescript",
              "tsx": "typescript",
              "py": "python",
              "go": "go",
              "rs": "rust",
              "java": "java",
              "c": "c",
              "cpp": "cpp",
              "json": "json",
              "yaml": "yaml",
              "yml": "yaml"
            };
            newFiles.push({
              id: `file_entry_${idx}_${Date.now()}`,
              name: fileName,
              language: languageMap[extension] || "plaintext",
              content: entry.content,
              hasUnsavedChanges: false,
              path: entry.path,
              repoInfo: repo
            });
          }
        });
      }
      if (data.config_files && Array.isArray(data.config_files)) {
        data.config_files.forEach((config2, idx) => {
          if (config2.content) {
            const fileName = config2.path.split("/").pop();
            const extension = fileName.split(".").pop().toLowerCase();
            newFiles.push({
              id: `file_config_${idx}_${Date.now()}`,
              name: fileName,
              language: extension === "json" ? "json" : "yaml",
              content: config2.content,
              hasUnsavedChanges: false,
              path: config2.path,
              repoInfo: repo
            });
          }
        });
      }
      if (newFiles.length === 0) {
        newFiles.push({
          id: `file_${Date.now()}`,
          name: "info.txt",
          language: "plaintext",
          content: `Repository: ${repo.owner}/${repo.repo}
Branch: ${repo.branch}

No core files found.`,
          hasUnsavedChanges: false,
          path: null,
          repoInfo: repo
        });
      }
      setFiles(newFiles);
      setActiveFileId(newFiles[0].id);
    } else {
      console.log("Full tree data:", data);
      if (data.tree && Array.isArray(data.tree)) {
        setTreeData(data.tree);
        setShowTree(true);
        setRepoInfo(repo);
        setFiles([]);
        setActiveFileId(null);
      } else {
        console.error("Invalid tree data received:", data);
        setError("Failed to load repository tree");
      }
    }
    setShowImportModal(false);
    setError(null);
  };
  const handleTreeFileSelect = async (file, isMultiSelect) => {
    if (file.type === "directory") {
      return;
    }
    if (isMultiSelect) {
      setSelectedTreeFiles((prev) => {
        const isAlreadySelected = prev.some((f2) => f2.path === file.path);
        if (isAlreadySelected) {
          return prev.filter((f2) => f2.path !== file.path);
        } else {
          return [...prev, file];
        }
      });
    } else {
      await fetchAndOpenFile(file.path);
    }
  };
  const fetchAndOpenFile = async (filePath) => {
    const existingFile = files.find((f2) => f2.path === filePath);
    if (existingFile) {
      setActiveFileId(existingFile.id);
      return;
    }
    try {
      setLoading(true);
      const fileData = await reviewApi.githubGetFile(
        repoInfo.url,
        filePath,
        repoInfo.branch
      );
      const fileName = filePath.split("/").pop();
      const extension = fileName.split(".").pop().toLowerCase();
      const languageMap = {
        "js": "javascript",
        "jsx": "javascript",
        "ts": "typescript",
        "tsx": "typescript",
        "go": "go",
        "py": "python",
        "java": "java",
        "c": "c",
        "cpp": "cpp",
        "cs": "csharp",
        "html": "html",
        "css": "css",
        "scss": "scss",
        "json": "json",
        "xml": "xml",
        "yaml": "yaml",
        "yml": "yaml",
        "md": "markdown",
        "sql": "sql",
        "sh": "shell",
        "bash": "shell",
        "rs": "rust",
        "rb": "ruby",
        "php": "php"
      };
      const newFile = {
        id: `file_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`,
        name: fileName,
        language: fileData.language || languageMap[extension] || "plaintext",
        content: fileData.content,
        hasUnsavedChanges: false,
        path: filePath,
        repoInfo
      };
      setFiles((prev) => [...prev, newFile]);
      setActiveFileId(newFile.id);
    } catch (err) {
      console.error("Failed to fetch file:", err);
      setError(`Failed to load file: ${err.message || "Unknown error"}`);
    } finally {
      setLoading(false);
    }
  };
  const handleFilesAnalyze = async (selectedFiles) => {
    console.log("Analyzing selected files:", selectedFiles);
    for (const file of selectedFiles) {
      await fetchAndOpenFile(file.path);
    }
    setSelectedTreeFiles([]);
  };
  const handleAnalyze = async () => {
    if (!code.trim()) {
      setError("Please enter some code to analyze");
      return;
    }
    if (!selectedModel) {
      setError("Please select an AI model");
      return;
    }
    if (selectedMode === "scan" && !scanQuery.trim()) {
      setError("Please enter a search query for Scan mode");
      return;
    }
    try {
      setLoading(true);
      setError(null);
      setAnalysisResult(null);
      let result;
      switch (selectedMode) {
        case "preview":
          result = await reviewApi.runPreview(sessionId, code, selectedModel, userMode, outputMode);
          break;
        case "skim":
          result = await reviewApi.runSkim(sessionId, code, selectedModel, userMode, outputMode);
          break;
        case "scan":
          result = await reviewApi.runScan(sessionId, code, selectedModel, scanQuery, userMode, outputMode);
          break;
        case "detailed":
          result = await reviewApi.runDetailed(sessionId, code, selectedModel, userMode, outputMode);
          break;
        case "critical":
          result = await reviewApi.runCritical(sessionId, code, selectedModel, userMode, outputMode);
          break;
        default:
          throw new Error(`Unknown analysis mode: ${selectedMode}`);
      }
      setAnalysisResult(result);
    } catch (err) {
      console.error("Analysis failed:", err);
      setError(err.message || "Analysis failed. Please try again.");
    } finally {
      setLoading(false);
    }
  };
  const handleRetry = () => {
    setError(null);
    handleAnalyze();
  };
  const clearCode = () => {
    setFiles((prevFiles) => prevFiles.map(
      (file) => file.id === activeFileId ? { ...file, content: "", hasUnsavedChanges: false } : file
    ));
    setAnalysisResult(null);
    setError(null);
  };
  const resetToDefault = () => {
    const newFileId = `file_${Date.now()}`;
    setFiles([{
      id: newFileId,
      name: "info.txt",
      language: "plaintext",
      content: defaultCode,
      hasUnsavedChanges: false,
      path: null
    }]);
    setActiveFileId(newFileId);
    setAnalysisResult(null);
    setError(null);
    setTreeData(null);
    setShowTree(false);
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid py-3", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("nav", { className: "frosted-card mb-4 p-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: "/", className: "btn btn-outline-primary btn-sm", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-left me-2" }),
        "Back to Dashboard"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center gap-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("button", { onClick: toggleTheme, className: "theme-toggle", children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi ${isDarkMode ? "bi-sun-fill" : "bi-moon-fill"}` }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "me-2", children: [
          "Welcome, ",
          (user == null ? void 0 : user.username) || (user == null ? void 0 : user.name),
          "!"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-outline-danger btn-sm",
            onClick: () => logout(),
            children: "Logout"
          }
        )
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: "mb-1", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-code-square text-primary me-2" }),
          "Code Review"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted mb-0", children: "AI-powered code analysis with five distinct reading modes" })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-2 align-items-center", children: [
        repoInfo && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-end me-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted d-block", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-github me-1" }),
            repoInfo.owner,
            "/",
            repoInfo.repo
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-git me-1" }),
            repoInfo.branch
          ] })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted", children: [
          "Session: ",
          sessionId
        ] })
      ] })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
      AnalysisModeSelector,
      {
        selectedMode,
        onModeSelect: setSelectedMode,
        onDetailsClick: handleDetailsClick,
        disabled: loading
      }
    ) }) }),
    selectedMode === "scan" && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-3", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "scanQuery", className: "form-label mb-2", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-search me-2" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "What are you looking for?" })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "input",
        {
          type: "text",
          id: "scanQuery",
          className: "form-control",
          placeholder: 'Try "functions", "error handling", "database queries", "authentication logic", etc.',
          value: scanQuery,
          onChange: (e) => setScanQuery(e.target.value),
          disabled: loading,
          style: {
            backgroundColor: "rgba(255, 255, 255, 0.95)",
            border: "2px solid rgba(99, 102, 241, 0.2)",
            borderRadius: "8px",
            padding: "0.75rem"
          }
        }
      ),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted mt-2 d-block", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-1" }),
        "This is a ",
        /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "context-aware search" }),
        ' - ask for concepts like "functions" or "error handling" rather than exact text matches. The AI will find and analyze relevant code patterns.'
      ] })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row mb-3", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
        ModelSelector,
        {
          selectedModel,
          onModelSelect: setSelectedModel,
          disabled: loading
        }
      ) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { className: "form-label", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-person-circle me-2" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Experience Level" })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "select",
          {
            className: `form-select ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`,
            style: isDarkMode ? {
              backgroundColor: "#1a1d2e",
              color: "#e0e7ff",
              borderColor: "#4a5568"
            } : {},
            value: userMode,
            onChange: (e) => setUserMode(e.target.value),
            disabled: loading,
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "beginner", children: "🎓 Beginner (Detailed with analogies)" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "novice", children: "📚 Novice (<2 years)" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "intermediate", children: "⚡ Intermediate (3-5 years)" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "expert", children: "🚀 Expert (Concise bullets)" })
            ]
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted d-block mt-1", children: "Adjusts explanation depth and technical terminology" })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { className: "form-label", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lightbulb me-2" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Learning Style" })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "btn-group w-100", role: "group", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              type: "radio",
              className: "btn-check",
              name: "outputMode",
              id: "outputQuick",
              value: "quick",
              checked: outputMode === "quick",
              onChange: (e) => setOutputMode(e.target.value),
              disabled: loading
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "btn btn-outline-primary", htmlFor: "outputQuick", children: "Quick Learn" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              type: "radio",
              className: "btn-check",
              name: "outputMode",
              id: "outputDetailed",
              value: "detailed",
              checked: outputMode === "detailed",
              onChange: (e) => setOutputMode(e.target.value),
              disabled: loading
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "btn btn-outline-primary", htmlFor: "outputDetailed", children: "Full Learn" })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "text-muted d-block mt-1", children: outputMode === "detailed" ? "🧠 Shows AI reasoning process" : "⚡ Just the analysis" })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "d-flex gap-2 align-items-end h-100", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          className: "btn btn-primary flex-grow-1",
          onClick: handleAnalyze,
          disabled: loading || !code.trim() || !selectedModel,
          children: loading ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-2", role: "status", "aria-hidden": "true" }),
            "Analyzing..."
          ] }) : "Analyze Code"
        }
      ) }) })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-12", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-2", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          className: "btn btn-outline-secondary btn-sm",
          onClick: resetToDefault,
          disabled: loading,
          children: "Reset to Example"
        }
      ),
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          className: "btn btn-outline-danger btn-sm",
          onClick: clearCode,
          disabled: loading,
          children: "Clear"
        }
      ),
      /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          className: "btn btn-primary btn-sm",
          onClick: () => setShowImportModal(true),
          disabled: loading,
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-github me-2" }),
            "Import from GitHub"
          ]
        }
      )
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row g-3", children: [
      showTree && treeData && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "p-3 border-bottom d-flex justify-content-between align-items-center", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-0", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder-tree me-2" }),
            "Repository Files"
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              className: "btn btn-sm btn-outline-secondary",
              onClick: () => {
                setShowTree(false);
                setTreeData(null);
                setSelectedTreeFiles([]);
              },
              title: "Close file tree",
              children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-x-lg" })
            }
          )
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "flex-grow-1 overflow-auto", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
          FileTreeBrowser,
          {
            treeData,
            selectedFiles: selectedTreeFiles,
            onFileSelect: handleTreeFileSelect,
            onFilesAnalyze: handleFilesAnalyze,
            loading
          }
        ) })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: showTree ? "col-md-5" : "col-md-6", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card h-100", style: { display: "flex", flexDirection: "column" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "p-3 border-bottom d-flex justify-content-between align-items-center", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "mb-0", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-file-code me-2" }),
            "Code Input"
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-3", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { style: { color: "var(--bs-gray-200)" }, children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-type me-1" }),
              code.length,
              " chars"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { style: { color: "var(--bs-gray-200)" }, children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-list-ol me-1" }),
              code.split("\n").length,
              " lines"
            ] })
          ] })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          FileTabs,
          {
            files,
            activeFileId,
            onFileSelect: handleFileSelect,
            onFileClose: handleFileClose,
            onFileAdd: handleFileAdd,
            onFileRename: handleFileRename
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "p-0 flex-grow-1", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
          CodeEditor,
          {
            ref: codeEditorRef,
            value: code,
            onChange: handleCodeChange,
            language: (activeFile == null ? void 0 : activeFile.language) || "javascript",
            placeholder: "Enter your code here for analysis...",
            className: "h-100"
          }
        ) })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: showTree ? "col-md-4" : "col-md-6", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
        AnalysisOutput,
        {
          result: analysisResult,
          loading,
          error,
          mode: selectedMode,
          onRetry: handleRetry
        }
      ) })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row mt-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "frosted-card p-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { style: {
      color: "var(--bs-gray-200)",
      opacity: 0.95
    }, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lightbulb me-1" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Tips:" }),
      " Try different analysis modes to understand code from various perspectives. Preview for structure, Skim for abstractions, Scan for specific elements, Detailed for algorithms, and Critical for quality assessment."
    ] }) }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      RepoImportModal,
      {
        show: showImportModal,
        onClose: () => setShowImportModal(false),
        onSuccess: handleGitHubImportSuccess
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      PromptEditorModal,
      {
        isOpen: showPromptEditor,
        onClose: handlePromptEditorClose,
        mode: promptEditorMode,
        userLevel: userMode,
        outputMode
      }
    )
  ] });
}
function AnalyticsPage() {
  const { user, logout } = useAuth();
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container mt-4", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("nav", { className: "navbar navbar-expand-lg navbar-light frosted-card mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: "/", className: "navbar-brand", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-left me-2" }),
        "Back to Dashboard"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "me-3", children: [
          "Welcome, ",
          (user == null ? void 0 : user.username) || (user == null ? void 0 : user.name),
          "!"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-outline-danger btn-sm",
            onClick: () => logout(),
            children: "Logout"
          }
        )
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: "mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-graph-up text-info me-2" }),
        "Analytics Dashboard"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", children: "Analyze trends and patterns in your application data" })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-bar-chart-line text-primary", style: { fontSize: "3rem" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mt-3 mb-3", children: "Trends" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { color: "var(--bs-gray-200)" }, children: "Identify patterns and trends in your data" })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-diamond text-warning", style: { fontSize: "3rem" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mt-3 mb-3", children: "Anomalies" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { color: "var(--bs-gray-200)" }, children: "Detect unusual patterns and outliers" })
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4 text-center h-100", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-speedometer2 text-success", style: { fontSize: "3rem" } }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mt-3 mb-3", children: "Performance" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-0", style: { color: "var(--bs-gray-200)" }, children: "Monitor system performance metrics" })
      ] }) })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "frosted-card p-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-3", children: "Coming Soon" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "mb-3", style: { color: "var(--bs-gray-200)" }, children: "Analytics features are currently in development. Check back soon for:" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("ul", { style: { color: "var(--bs-gray-200)" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: "Real-time trend analysis" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: "Anomaly detection and alerts" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: "Performance monitoring dashboards" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: "Custom report generation" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("li", { children: "Data visualization and charts" })
      ] })
    ] }) }) })
  ] });
}
const PROVIDERS = [
  { value: "anthropic", label: "Anthropic (Claude)" },
  { value: "openai", label: "OpenAI (GPT)" },
  { value: "ollama", label: "Ollama (Local)" },
  { value: "deepseek", label: "DeepSeek" },
  { value: "mistral", label: "Mistral AI" }
];
const MODELS_BY_PROVIDER = {
  anthropic: [
    "claude-3-5-sonnet-20241022",
    "claude-3-5-sonnet-20240620",
    "claude-3-opus-20240229",
    "claude-3-5-haiku-20241022",
    "claude-3-sonnet-20240229",
    "claude-3-haiku-20240307"
  ],
  openai: [
    "gpt-4-turbo-preview",
    "gpt-4",
    "gpt-3.5-turbo",
    "gpt-4-32k"
  ],
  ollama: [
    "llama3.1:70b",
    "llama3.1:8b",
    "deepseek-coder-v2:16b",
    "deepseek-coder:6.7b",
    "qwen2.5-coder:7b",
    "codestral:22b"
  ],
  deepseek: [
    "deepseek-chat",
    "deepseek-coder"
  ],
  mistral: [
    "mistral-large",
    "mistral-medium",
    "mistral-small",
    "codestral-latest"
  ]
};
function AddLLMConfigModal({ isOpen, onClose, onSave, editingConfig }) {
  const [formData, setFormData] = reactExports.useState({
    name: "",
    provider: "anthropic",
    model: "",
    api_key: "",
    endpoint: "",
    is_default: false
  });
  const [availableModels, setAvailableModels] = reactExports.useState(MODELS_BY_PROVIDER.anthropic);
  const [testingConnection, setTestingConnection] = reactExports.useState(false);
  const [testResult, setTestResult] = reactExports.useState(null);
  const [errors, setErrors] = reactExports.useState({});
  const [showAdvanced, setShowAdvanced] = reactExports.useState(false);
  reactExports.useEffect(() => {
    if (editingConfig) {
      setFormData({
        name: editingConfig.name || "",
        provider: editingConfig.provider || "anthropic",
        model: editingConfig.model || "",
        api_key: "",
        // Never pre-fill API key for security
        endpoint: editingConfig.endpoint || "",
        is_default: editingConfig.is_default || false
      });
      setAvailableModels(MODELS_BY_PROVIDER[editingConfig.provider] || []);
    } else {
      setFormData({
        name: "",
        provider: "anthropic",
        model: "",
        api_key: "",
        endpoint: "",
        is_default: false
      });
      setAvailableModels(MODELS_BY_PROVIDER.anthropic);
    }
    setTestResult(null);
    setErrors({});
  }, [editingConfig, isOpen]);
  const handleProviderChange = (provider) => {
    setFormData((prev) => ({
      ...prev,
      provider,
      model: ""
      // Reset model when provider changes
    }));
    setAvailableModels(MODELS_BY_PROVIDER[provider] || []);
  };
  const handleInputChange = (e) => {
    const { name, value, type, checked } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value
    }));
    if (errors[name]) {
      setErrors((prev) => ({ ...prev, [name]: null }));
    }
  };
  const validateForm = () => {
    const newErrors = {};
    if (!formData.name.trim()) {
      newErrors.name = "Name is required";
    }
    if (!formData.model) {
      newErrors.model = "Model is required";
    }
    if (formData.provider !== "ollama" && !formData.api_key && !editingConfig) {
      newErrors.api_key = "API key is required";
    }
    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };
  const handleTestConnection = async () => {
    if (!validateForm()) {
      return;
    }
    setTestingConnection(true);
    setTestResult(null);
    try {
      const response = await apiRequest("/api/portal/llm-configs/test", {
        method: "POST",
        body: JSON.stringify({
          provider: formData.provider,
          model: formData.model,
          api_key: formData.api_key || void 0,
          endpoint: formData.endpoint || void 0
        })
      });
      setTestResult({
        success: true,
        message: (response == null ? void 0 : response.message) || "Connection successful!"
      });
    } catch (err) {
      setTestResult({
        success: false,
        message: err.message || "Connection failed"
      });
    } finally {
      setTestingConnection(false);
    }
  };
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!validateForm()) {
      return;
    }
    try {
      let response;
      if (editingConfig) {
        const updateData = { ...formData };
        if (!updateData.api_key) {
          delete updateData.api_key;
        }
        response = await apiRequest(`/api/portal/llm-configs/${editingConfig.id}`, {
          method: "PUT",
          body: JSON.stringify(updateData)
        });
      } else {
        response = await apiRequest("/api/portal/llm-configs", {
          method: "POST",
          body: JSON.stringify(formData)
        });
      }
      onSave(response);
      onClose();
    } catch (err) {
      console.error("Failed to save config:", err);
      alert("Failed to save configuration: " + (err.message || "Unknown error"));
    }
  };
  if (!isOpen) {
    return null;
  }
  const isFormValid = formData.name.trim() && formData.model && (formData.provider === "ollama" || formData.api_key || editingConfig);
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal show d-block", style: { backgroundColor: "rgba(0,0,0,0.5)" }, role: "dialog", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-lg", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-content", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-header", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "modal-title", children: [
        editingConfig ? "Edit" : "Add",
        " AI Model Configuration"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          type: "button",
          className: "btn-close",
          onClick: onClose,
          "aria-label": "Close"
        }
      )
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("form", { onSubmit: handleSubmit, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-body", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "config-name", className: "form-label", children: [
            "Configuration Name ",
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              type: "text",
              className: `form-control ${errors.name ? "is-invalid" : ""}`,
              id: "config-name",
              name: "name",
              value: formData.name,
              onChange: handleInputChange,
              placeholder: "e.g., Claude for Review",
              required: true
            }
          ),
          errors.name && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: errors.name })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "provider", className: "form-label", children: [
            "Provider ",
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "select",
            {
              className: "form-select",
              id: "provider",
              name: "provider",
              value: formData.provider,
              onChange: (e) => handleProviderChange(e.target.value),
              required: true,
              children: PROVIDERS.map((provider) => /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: provider.value, children: provider.label }, provider.value))
            }
          )
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "model", className: "form-label", children: [
            "Model ",
            /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs(
            "select",
            {
              className: `form-select ${errors.model ? "is-invalid" : ""}`,
              id: "model",
              name: "model",
              value: formData.model,
              onChange: handleInputChange,
              required: true,
              children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "", children: "Select a model" }),
                availableModels.map((model) => /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: model, children: model }, model))
              ]
            }
          ),
          errors.model && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: errors.model })
        ] }),
        formData.provider !== "ollama" && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "api-key", className: "form-label", children: [
            "API Key ",
            !editingConfig && /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              type: "password",
              className: `form-control ${errors.api_key ? "is-invalid" : ""}`,
              id: "api-key",
              name: "api_key",
              value: formData.api_key,
              onChange: handleInputChange,
              placeholder: editingConfig ? "Leave blank to keep existing key" : "Enter API key",
              required: !editingConfig
            }
          ),
          errors.api_key && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: errors.api_key }),
          editingConfig && /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "form-text text-muted", children: "Leave blank to keep the existing API key" })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "endpoint", className: "form-label", children: "Custom Endpoint (Optional)" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              type: "url",
              className: "form-control",
              id: "endpoint",
              name: "endpoint",
              value: formData.endpoint,
              onChange: handleInputChange,
              placeholder: "https://api.example.com/v1"
            }
          ),
          /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "form-text text-muted", children: "Override the default API endpoint" })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs(
            "button",
            {
              type: "button",
              className: "btn btn-link p-0 text-decoration-none d-flex align-items-center",
              onClick: () => setShowAdvanced(!showAdvanced),
              children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-chevron-${showAdvanced ? "down" : "right"} me-2` }),
                "Advanced Settings"
              ]
            }
          ),
          showAdvanced && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-3 ps-4 border-start", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "form-check", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "input",
              {
                type: "checkbox",
                className: "form-check-input",
                id: "is-default",
                name: "is_default",
                checked: formData.is_default,
                onChange: handleInputChange
              }
            ),
            /* @__PURE__ */ jsxRuntimeExports.jsx("label", { className: "form-check-label", htmlFor: "is-default", children: "Set as default configuration" })
          ] }) })
        ] }),
        testResult && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `alert ${testResult.success ? "alert-success" : "alert-danger"}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi ${testResult.success ? "bi-check-circle" : "bi-x-circle"} me-2` }),
          testResult.message
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-footer", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: "btn btn-secondary",
            onClick: handleTestConnection,
            disabled: testingConnection || !isFormValid,
            children: testingConnection ? /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "spinner-border spinner-border-sm me-2", role: "status" }),
              "Testing..."
            ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs(jsxRuntimeExports.Fragment, { children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plug me-1" }),
              "Test Connection"
            ] })
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: "btn btn-secondary",
            onClick: onClose,
            children: "Cancel"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            type: "submit",
            className: "btn btn-primary",
            disabled: !isFormValid,
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-save me-1" }),
              editingConfig ? "Update" : "Save"
            ]
          }
        )
      ] })
    ] })
  ] }) }) });
}
function LLMConfigPage() {
  var _a, _b, _c;
  const { isDarkMode, toggleTheme } = useTheme();
  const { user } = useAuth();
  const [configs, setConfigs] = reactExports.useState([]);
  const [appPreferences, setAppPreferences] = reactExports.useState({});
  const [usageSummary, setUsageSummary] = reactExports.useState(null);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  const [showAddModal, setShowAddModal] = reactExports.useState(false);
  const [editingConfig, setEditingConfig] = reactExports.useState(null);
  const [deletingConfigId, setDeletingConfigId] = reactExports.useState(null);
  reactExports.useEffect(() => {
    loadConfigs();
    loadAppPreferences();
    loadUsageSummary();
  }, []);
  const loadConfigs = async () => {
    try {
      setLoading(true);
      const data = await apiRequest("/api/portal/llm-configs");
      setConfigs(data || []);
      setError(null);
    } catch (err) {
      console.error("Failed to load configs:", err);
      setError("Failed to load AI model configurations");
      setConfigs([]);
    } finally {
      setLoading(false);
    }
  };
  const loadAppPreferences = async () => {
    try {
      const data = await apiRequest("/api/portal/app-llm-preferences");
      setAppPreferences(data || {});
    } catch (err) {
      console.error("Failed to load app preferences:", err);
      setAppPreferences({});
    }
  };
  const loadUsageSummary = async () => {
    try {
      const data = await apiRequest("/api/portal/llm-usage/summary?period=30d");
      setUsageSummary(data);
    } catch (err) {
      console.error("Failed to load usage summary:", err);
      setUsageSummary(null);
    }
  };
  const handleDeleteConfig = async (configId) => {
    try {
      await apiRequest(`/api/portal/llm-configs/${configId}`, { method: "DELETE" });
      await loadConfigs();
      setDeletingConfigId(null);
    } catch (err) {
      console.error("Failed to delete config:", err);
      alert("Failed to delete configuration: " + (err.message || "Unknown error"));
    }
  };
  const handleSetAppPreference = async (appName, configId) => {
    try {
      await apiRequest(`/api/portal/app-llm-preferences/${appName}`, {
        method: "PUT",
        body: JSON.stringify({ llm_config_id: configId })
      });
      await loadAppPreferences();
    } catch (err) {
      console.error("Failed to set app preference:", err);
      alert("Failed to set app preference");
    }
  };
  const handleToggleDefault = async (configId, currentDefault) => {
    try {
      await apiRequest(`/api/portal/llm-configs/${configId}/set-default`, {
        method: "PUT",
        body: JSON.stringify({ is_default: !currentDefault })
      });
      await loadConfigs();
    } catch (err) {
      console.error("Failed to set default:", err);
      alert("Failed to set default configuration: " + (err.message || "Unknown error"));
    }
  };
  const handleSaveConfig = async (configData) => {
    try {
      await loadConfigs();
      setShowAddModal(false);
      setEditingConfig(null);
    } catch (err) {
      console.error("Error refreshing configs:", err);
    }
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `container mt-4 ${isDarkMode ? "text-light" : ""}`, children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("nav", { className: `navbar navbar-expand-lg mb-4 frosted-card ${isDarkMode ? "navbar-dark bg-dark border-secondary" : "navbar-light"}`, children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "navbar-brand fw-bold", style: { fontSize: "1.5rem", color: isDarkMode ? "#e0e7ff" : "#1e293b" }, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-code-square me-2" }),
        "DevSmith Platform"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex align-items-center gap-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            onClick: toggleTheme,
            className: `btn btn-sm ${isDarkMode ? "btn-outline-light" : "btn-outline-dark"}`,
            title: isDarkMode ? "Switch to Light Mode" : "Switch to Dark Mode",
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-${isDarkMode ? "sun" : "moon"}-fill` })
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: "/portal", className: `btn btn-sm ${isDarkMode ? "btn-outline-light" : "btn-outline-secondary"}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-left me-1" }),
          "Back to Dashboard"
        ] })
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: `card shadow-sm ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`, children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: `card-title mb-3 ${isDarkMode ? "text-light" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-robot me-2", style: { fontSize: "1.5rem", verticalAlign: "middle" } }),
        "AI Factory"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: `card-text mb-0 ${isDarkMode ? "text-light" : "text-muted"}`, children: "Manage your AI model configurations, API keys, and app-specific preferences." })
    ] }) }) }) }),
    error && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle me-2" }),
      error
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `frosted-card p-4 ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("h4", { className: `mb-0 ${isDarkMode ? "text-light" : ""}`, children: "Your AI Models" }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            className: "btn btn-primary btn-sm",
            onClick: () => setShowAddModal(true),
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle me-1" }),
              "Add Model"
            ]
          }
        )
      ] }),
      loading ? /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "text-center py-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }) }) : configs.length === 0 ? /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `alert ${isDarkMode ? "alert-secondary" : "alert-info"} mb-0`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-2" }),
        'No AI models configured yet. Click "Add Model" to get started.'
      ] }) : /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "table-responsive", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("table", { className: `table table-hover ${isDarkMode ? "table-dark" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("thead", { children: /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Name" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Provider" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Model" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "API Key" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Default" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Actions" })
        ] }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("tbody", { children: configs.map((config2) => /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: config2.name }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-secondary", children: config2.provider }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: config2.model }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: config2.has_api_key ? /* @__PURE__ */ jsxRuntimeExports.jsxs("span", { className: "badge bg-success", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lock-fill me-1" }),
            "Set"
          ] }) : /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "badge bg-secondary", children: "None" }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "form-check form-switch", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
            "input",
            {
              className: "form-check-input",
              type: "checkbox",
              role: "switch",
              id: `default-toggle-${config2.id}`,
              checked: config2.is_default,
              onChange: () => handleToggleDefault(config2.id, config2.is_default),
              style: {
                cursor: "pointer",
                width: "3em",
                height: "1.5em"
              }
            }
          ) }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "btn-group btn-group-sm", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                className: "btn btn-outline-primary",
                title: "Edit",
                onClick: () => {
                  setEditingConfig(config2);
                  setShowAddModal(true);
                },
                children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-pencil" })
              }
            ),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                className: "btn btn-outline-danger",
                title: "Delete",
                onClick: () => setDeletingConfigId(config2.id),
                children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-trash" })
              }
            )
          ] }) })
        ] }, config2.id)) })
      ] }) })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `frosted-card p-4 ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h4", { className: `mb-3 ${isDarkMode ? "text-light" : ""}`, children: "App-Specific Preferences" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: `mb-3 ${isDarkMode ? "text-light" : "text-muted"}`, children: "Choose which AI model each app should use by default." }),
      ["review", "logs", "analytics"].map((appName) => {
        var _a2;
        return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row mb-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { className: `form-label text-capitalize ${isDarkMode ? "text-light" : ""}`, children: [
            appName,
            " App:"
          ] }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
            "select",
            {
              className: `form-select ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`,
              name: `${appName}-preference`,
              value: ((_a2 = appPreferences[appName]) == null ? void 0 : _a2.config_id) || "",
              onChange: (e) => handleSetAppPreference(appName, e.target.value),
              children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("option", { value: "", children: "Use Default" }),
                configs.map((config2) => /* @__PURE__ */ jsxRuntimeExports.jsxs("option", { value: config2.id, children: [
                  config2.name,
                  " (",
                  config2.provider_type,
                  ")"
                ] }, config2.id))
              ]
            }
          ) })
        ] }, appName);
      })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `frosted-card p-4 ${isDarkMode ? "bg-dark text-light border-secondary" : ""}`, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("h4", { className: `mb-3 ${isDarkMode ? "text-light" : ""}`, children: "Usage Summary (Last 30 Days)" }),
      usageSummary ? /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `text-center p-3 rounded ${isDarkMode ? "bg-secondary text-light" : "bg-light"}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: isDarkMode ? "text-light" : "text-muted", children: "Total Tokens" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: isDarkMode ? "text-light" : "", children: ((_a = usageSummary.total_tokens) == null ? void 0 : _a.toLocaleString()) || 0 })
        ] }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `text-center p-3 rounded ${isDarkMode ? "bg-secondary text-light" : "bg-light"}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: isDarkMode ? "text-light" : "text-muted", children: "Requests" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: isDarkMode ? "text-light" : "", children: ((_b = usageSummary.total_requests) == null ? void 0 : _b.toLocaleString()) || 0 })
        ] }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `text-center p-3 rounded ${isDarkMode ? "bg-secondary text-light" : "bg-light"}`, children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: isDarkMode ? "text-light" : "text-muted", children: "Estimated Cost" }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("h3", { className: isDarkMode ? "text-light" : "", children: [
            "$",
            ((_c = usageSummary.estimated_cost) == null ? void 0 : _c.toFixed(2)) || "0.00"
          ] })
        ] }) })
      ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `alert ${isDarkMode ? "alert-secondary" : "alert-info"} mb-0`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-info-circle me-2" }),
        "No usage data yet. Start using AI features to see statistics here."
      ] })
    ] }) }) }),
    deletingConfigId && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal show d-block", style: { backgroundColor: "rgba(0,0,0,0.5)" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-content ${isDarkMode ? "bg-dark text-light" : ""}`, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-header ${isDarkMode ? "border-secondary" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "modal-title", children: "Confirm Deletion" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: `btn-close ${isDarkMode ? "btn-close-white" : ""}`,
            onClick: () => setDeletingConfigId(null)
          }
        )
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-body ${isDarkMode ? "bg-dark text-light" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { children: "Are you sure you want to delete this AI model configuration?" }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { className: "text-danger mb-0", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle me-1" }),
          "This action cannot be undone."
        ] })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: `modal-footer ${isDarkMode ? "border-secondary" : ""}`, children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: "btn btn-secondary",
            onClick: () => setDeletingConfigId(null),
            children: "Cancel"
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            type: "button",
            className: "btn btn-danger",
            onClick: () => handleDeleteConfig(deletingConfigId),
            children: "Delete"
          }
        )
      ] })
    ] }) }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      AddLLMConfigModal,
      {
        isOpen: showAddModal,
        onClose: () => {
          setShowAddModal(false);
          setEditingConfig(null);
        },
        onSave: handleSaveConfig,
        editingConfig
      }
    )
  ] });
}
const projectsApi = {
  getAll: () => apiRequest("/api/logs/projects"),
  create: (data) => apiRequest("/api/logs/projects", {
    method: "POST",
    body: JSON.stringify(data)
  }),
  regenerateKey: (id2) => apiRequest(`/api/logs/projects/${id2}/regenerate-key`, {
    method: "POST"
  }),
  deactivate: (id2) => apiRequest(`/api/logs/projects/${id2}`, {
    method: "DELETE"
  })
};
function ProjectsPage() {
  const [projects, setProjects] = reactExports.useState([]);
  const [loading, setLoading] = reactExports.useState(true);
  const [error, setError] = reactExports.useState(null);
  const [showCreateModal, setShowCreateModal] = reactExports.useState(false);
  const [newApiKey, setNewApiKey] = reactExports.useState(null);
  const [regeneratedKey, setRegeneratedKey] = reactExports.useState(null);
  const [formData, setFormData] = reactExports.useState({
    name: "",
    slug: "",
    description: "",
    repository_url: ""
  });
  const [formErrors, setFormErrors] = reactExports.useState({});
  reactExports.useEffect(() => {
    fetchProjects();
  }, []);
  const fetchProjects = async () => {
    try {
      setLoading(true);
      const data = await projectsApi.getAll();
      setProjects(data || []);
      setError(null);
    } catch (err) {
      console.error("Failed to fetch projects:", err);
      setError(err.message || "Failed to load projects");
    } finally {
      setLoading(false);
    }
  };
  const validateForm = () => {
    const errors = {};
    if (!formData.name.trim()) {
      errors.name = "Project name is required";
    }
    if (!formData.slug.trim()) {
      errors.slug = "Project slug is required";
    } else if (!/^[a-z0-9-]+$/.test(formData.slug)) {
      errors.slug = "Slug must contain only lowercase letters, numbers, and hyphens";
    }
    if (formData.repository_url && !/^https?:\/\/.+/.test(formData.repository_url)) {
      errors.repository_url = "Repository URL must start with http:// or https://";
    }
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };
  const handleCreateProject = async (e) => {
    e.preventDefault();
    if (!validateForm()) {
      return;
    }
    try {
      const data = await projectsApi.create(formData);
      setNewApiKey(data.api_key);
      await fetchProjects();
      setFormData({
        name: "",
        slug: "",
        description: "",
        repository_url: ""
      });
      setFormErrors({});
    } catch (err) {
      console.error("Failed to create project:", err);
      setError(err.message || "Failed to create project");
    }
  };
  const handleRegenerateKey = async (projectId, projectName) => {
    if (!confirm(`Are you sure you want to regenerate the API key for "${projectName}"? The old key will stop working immediately.`)) {
      return;
    }
    try {
      const data = await projectsApi.regenerateKey(projectId);
      setRegeneratedKey({
        projectName,
        apiKey: data.api_key
      });
      await fetchProjects();
    } catch (err) {
      console.error("Failed to regenerate key:", err);
      setError(err.message || "Failed to regenerate API key");
    }
  };
  const handleDeactivateProject = async (projectId, projectName) => {
    if (!confirm(`Are you sure you want to deactivate "${projectName}"? This will stop accepting logs for this project.`)) {
      return;
    }
    try {
      await projectsApi.deactivate(projectId);
      await fetchProjects();
    } catch (err) {
      console.error("Failed to deactivate project:", err);
      setError(err.message || "Failed to deactivate project");
    }
  };
  const closeApiKeyModal = () => {
    setNewApiKey(null);
    setRegeneratedKey(null);
  };
  const copyToClipboard = (text) => {
    navigator.clipboard.writeText(text);
    alert("API key copied to clipboard!");
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid py-4", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "row mb-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "col", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("h1", { className: "h2 mb-1", children: "Cross-Repository Logging Projects" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: "Manage API keys for external applications to send logs to DevSmith" })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-auto", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          className: "btn btn-primary",
          onClick: () => setShowCreateModal(true),
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle me-2" }),
            "Create Project"
          ]
        }
      ) })
    ] }),
    error && /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-danger alert-dismissible fade show", role: "alert", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle me-2" }),
      error,
      /* @__PURE__ */ jsxRuntimeExports.jsx(
        "button",
        {
          type: "button",
          className: "btn-close",
          onClick: () => setError(null)
        }
      )
    ] }),
    loading ? /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "text-center py-5", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }) }) : projects.length === 0 ? /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body text-center py-5", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-inbox display-1 text-muted" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3", children: "No Projects Yet" }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted mb-4", children: "Create your first project to start sending logs from external applications" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          className: "btn btn-primary",
          onClick: () => setShowCreateModal(true),
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle me-2" }),
            "Create Your First Project"
          ]
        }
      )
    ] }) }) : /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row g-4", children: projects.map((project) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-12 col-md-6 col-xl-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card h-100", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-start mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "card-title mb-1", children: project.name }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("code", { className: "text-muted small", children: project.slug })
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "span",
          {
            className: `badge ${project.is_active ? "bg-success" : "bg-secondary"}`,
            children: project.is_active ? "Active" : "Inactive"
          }
        )
      ] }),
      project.description && /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "card-text text-muted small mb-3", children: project.description }),
      project.repository_url && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "a",
        {
          href: project.repository_url,
          target: "_blank",
          rel: "noopener noreferrer",
          className: "text-decoration-none small",
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-github me-1" }),
            "Repository"
          ]
        }
      ) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("small", { className: "text-muted", children: [
        "Created ",
        new Date(project.created_at).toLocaleDateString()
      ] }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex gap-2", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs(
          "button",
          {
            className: "btn btn-sm btn-outline-primary flex-grow-1",
            onClick: () => window.location.href = `/logs?project_id=${project.id}`,
            children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-list-ul me-1" }),
              "View Logs"
            ]
          }
        ),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-sm btn-outline-secondary",
            onClick: () => handleRegenerateKey(project.id, project.name),
            title: "Regenerate API Key",
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-clockwise" })
          }
        ),
        project.is_active && /* @__PURE__ */ jsxRuntimeExports.jsx(
          "button",
          {
            className: "btn btn-sm btn-outline-danger",
            onClick: () => handleDeactivateProject(project.id, project.name),
            title: "Deactivate Project",
            children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-x-circle" })
          }
        )
      ] })
    ] }) }) }, project.id)) }),
    showCreateModal && /* @__PURE__ */ jsxRuntimeExports.jsx(
      "div",
      {
        className: "modal show d-block",
        tabIndex: "-1",
        style: { backgroundColor: "rgba(0,0,0,0.5)" },
        children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-dialog-centered", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-content", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-header", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "modal-title", children: "Create New Project" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx(
              "button",
              {
                type: "button",
                className: "btn-close",
                onClick: () => {
                  setShowCreateModal(false);
                  setFormErrors({});
                }
              }
            )
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("form", { onSubmit: handleCreateProject, children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-body", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "name", className: "form-label", children: [
                  "Project Name ",
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
                ] }),
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "input",
                  {
                    type: "text",
                    className: `form-control ${formErrors.name ? "is-invalid" : ""}`,
                    id: "name",
                    value: formData.name,
                    onChange: (e) => setFormData({ ...formData, name: e.target.value }),
                    placeholder: "My Application"
                  }
                ),
                formErrors.name && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: formErrors.name })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsxs("label", { htmlFor: "slug", className: "form-label", children: [
                  "Project Slug ",
                  /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "text-danger", children: "*" })
                ] }),
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "input",
                  {
                    type: "text",
                    className: `form-control ${formErrors.slug ? "is-invalid" : ""}`,
                    id: "slug",
                    value: formData.slug,
                    onChange: (e) => setFormData({ ...formData, slug: e.target.value }),
                    placeholder: "my-application"
                  }
                ),
                /* @__PURE__ */ jsxRuntimeExports.jsx("small", { className: "form-text text-muted", children: "Used in API requests. Lowercase letters, numbers, and hyphens only." }),
                formErrors.slug && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: formErrors.slug })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "description", className: "form-label", children: "Description" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "textarea",
                  {
                    className: "form-control",
                    id: "description",
                    rows: "2",
                    value: formData.description,
                    onChange: (e) => setFormData({ ...formData, description: e.target.value }),
                    placeholder: "Brief description of this project..."
                  }
                )
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "repository_url", className: "form-label", children: "Repository URL" }),
                /* @__PURE__ */ jsxRuntimeExports.jsx(
                  "input",
                  {
                    type: "text",
                    className: `form-control ${formErrors.repository_url ? "is-invalid" : ""}`,
                    id: "repository_url",
                    value: formData.repository_url,
                    onChange: (e) => setFormData({ ...formData, repository_url: e.target.value }),
                    placeholder: "https://github.com/username/repo"
                  }
                ),
                formErrors.repository_url && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "invalid-feedback", children: formErrors.repository_url })
              ] })
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-footer", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "button",
                {
                  type: "button",
                  className: "btn btn-secondary",
                  onClick: () => {
                    setShowCreateModal(false);
                    setFormErrors({});
                  },
                  children: "Cancel"
                }
              ),
              /* @__PURE__ */ jsxRuntimeExports.jsxs("button", { type: "submit", className: "btn btn-primary", children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-plus-circle me-2" }),
                "Create Project"
              ] })
            ] })
          ] })
        ] }) })
      }
    ),
    (newApiKey || regeneratedKey) && /* @__PURE__ */ jsxRuntimeExports.jsx(
      "div",
      {
        className: "modal show d-block",
        tabIndex: "-1",
        style: { backgroundColor: "rgba(0,0,0,0.5)" },
        children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-dialog modal-dialog-centered", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-content", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-header bg-success text-white", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "modal-title", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-check-circle me-2" }),
            newApiKey ? "Project Created!" : "API Key Regenerated"
          ] }) }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "modal-body", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-warning", role: "alert", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle me-2" }),
              /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Important:" }),
              " Copy this API key now. You won't be able to see it again!"
            ] }),
            regeneratedKey && /* @__PURE__ */ jsxRuntimeExports.jsxs("p", { className: "mb-3", children: [
              "New API key for ",
              /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: regeneratedKey.projectName }),
              ":"
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "input-group mb-3", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "input",
                {
                  type: "text",
                  className: "form-control font-monospace",
                  value: newApiKey || (regeneratedKey == null ? void 0 : regeneratedKey.apiKey),
                  readOnly: true
                }
              ),
              /* @__PURE__ */ jsxRuntimeExports.jsx(
                "button",
                {
                  className: "btn btn-outline-secondary",
                  type: "button",
                  onClick: () => copyToClipboard(newApiKey || (regeneratedKey == null ? void 0 : regeneratedKey.apiKey)),
                  children: /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-clipboard" })
                }
              )
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { children: "Usage Example:" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("pre", { className: "bg-light p-3 rounded", children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: `curl -X POST http://your-devsmith-url/api/logs/batch \\
  -H "Authorization: Bearer ${newApiKey || (regeneratedKey == null ? void 0 : regeneratedKey.apiKey)}" \\
  -H "Content-Type: application/json" \\
  -d '{
    "project_slug": "${formData.slug || "your-project-slug"}",
    "logs": [{
      "timestamp": "2025-11-11T16:40:00Z",
      "level": "info",
      "message": "Application started",
      "service_name": "api-server",
      "context": {"version": "1.0.0"}
    }]
  }'` }) })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "modal-footer", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
            "button",
            {
              type: "button",
              className: "btn btn-primary",
              onClick: closeApiKeyModal,
              children: "I've Copied the Key"
            }
          ) })
        ] }) })
      }
    )
  ] });
}
function IntegrationDocsPage() {
  const [activeTab, setActiveTab] = reactExports.useState("javascript");
  const [copiedCode, setCopiedCode] = reactExports.useState(null);
  const handleCopyCode = (code, identifier) => {
    navigator.clipboard.writeText(code).then(() => {
      setCopiedCode(identifier);
      setTimeout(() => setCopiedCode(null), 2e3);
    });
  };
  const languages = {
    javascript: {
      label: "JavaScript / Node.js",
      icon: "bi-filetype-js",
      samples: [
        {
          title: "Basic Setup",
          description: "Minimal integration for Node.js applications",
          code: `// logs-client.js
const LOGS_API_URL = process.env.LOGS_API_URL || 'http://localhost:8082';
const LOGS_API_KEY = process.env.LOGS_API_KEY; // Get from Projects page

const logBatch = [];
const MAX_BATCH_SIZE = 1000;
const FLUSH_INTERVAL = 5000; // 5 seconds

async function sendBatch(entries) {
  if (entries.length === 0) return;
  
  try {
    const response = await fetch(\`\${LOGS_API_URL}/api/logs/batch\`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': \`Bearer \${LOGS_API_KEY}\`
      },
      body: JSON.stringify({ entries })
    });
    
    if (!response.ok) {
      console.error('Failed to send logs:', response.status);
    }
  } catch (error) {
    console.error('Error sending logs:', error);
  }
}

function log(level, message, metadata = {}) {
  logBatch.push({
    level,
    message,
    service: process.env.SERVICE_NAME || 'my-app',
    metadata,
    timestamp: new Date().toISOString()
  });
  
  if (logBatch.length >= MAX_BATCH_SIZE) {
    sendBatch([...logBatch]);
    logBatch.length = 0;
  }
}

// Flush logs periodically
setInterval(() => {
  if (logBatch.length > 0) {
    sendBatch([...logBatch]);
    logBatch.length = 0;
  }
}, FLUSH_INTERVAL);

// Export logging functions
module.exports = {
  debug: (msg, meta) => log('DEBUG', msg, meta),
  info: (msg, meta) => log('INFO', msg, meta),
  warn: (msg, meta) => log('WARNING', msg, meta),
  error: (msg, meta) => log('ERROR', msg, meta)
};`,
          language: "javascript"
        },
        {
          title: "Express.js Middleware",
          description: "Automatic request logging for Express applications",
          code: `// middleware/logging.js
const logger = require('./logs-client');

function loggingMiddleware(req, res, next) {
  const start = Date.now();
  
  // Log request
  logger.info('HTTP Request', {
    method: req.method,
    path: req.path,
    query: req.query,
    ip: req.ip
  });
  
  // Capture response
  res.on('finish', () => {
    const duration = Date.now() - start;
    const level = res.statusCode >= 500 ? 'ERROR' : 
                  res.statusCode >= 400 ? 'WARNING' : 'INFO';
    
    logger[level.toLowerCase()]('HTTP Response', {
      method: req.method,
      path: req.path,
      status: res.statusCode,
      duration_ms: duration
    });
  });
  
  next();
}

module.exports = loggingMiddleware;

// Usage in app.js:
// const loggingMiddleware = require('./middleware/logging');
// app.use(loggingMiddleware);`,
          language: "javascript"
        }
      ]
    },
    python: {
      label: "Python",
      icon: "bi-filetype-py",
      samples: [
        {
          title: "Basic Setup",
          description: "Minimal integration for Python applications",
          code: `# logs_client.py
import os
import json
import time
import requests
from threading import Thread, Lock
from datetime import datetime

LOGS_API_URL = os.getenv('LOGS_API_URL', 'http://localhost:8082')
LOGS_API_KEY = os.getenv('LOGS_API_KEY')  # Get from Projects page
SERVICE_NAME = os.getenv('SERVICE_NAME', 'my-app')

MAX_BATCH_SIZE = 1000
FLUSH_INTERVAL = 5  # seconds

class LogsClient:
    def __init__(self):
        self.batch = []
        self.lock = Lock()
        self.running = True
        self.thread = Thread(target=self._flush_worker, daemon=True)
        self.thread.start()
    
    def _send_batch(self, entries):
        if not entries:
            return
        
        try:
            response = requests.post(
                f'{LOGS_API_URL}/api/logs/batch',
                headers={
                    'Content-Type': 'application/json',
                    'Authorization': f'Bearer {LOGS_API_KEY}'
                },
                json={'entries': entries},
                timeout=10
            )
            response.raise_for_status()
        except Exception as e:
            print(f'Error sending logs: {e}')
    
    def _flush_worker(self):
        while self.running:
            time.sleep(FLUSH_INTERVAL)
            self.flush()
    
    def flush(self):
        with self.lock:
            if self.batch:
                self._send_batch(self.batch[:])
                self.batch.clear()
    
    def log(self, level, message, **metadata):
        entry = {
            'level': level,
            'message': message,
            'service': SERVICE_NAME,
            'metadata': metadata,
            'timestamp': datetime.utcnow().isoformat() + 'Z'
        }
        
        with self.lock:
            self.batch.append(entry)
            if len(self.batch) >= MAX_BATCH_SIZE:
                self._send_batch(self.batch[:])
                self.batch.clear()
    
    def debug(self, message, **metadata):
        self.log('DEBUG', message, **metadata)
    
    def info(self, message, **metadata):
        self.log('INFO', message, **metadata)
    
    def warn(self, message, **metadata):
        self.log('WARNING', message, **metadata)
    
    def error(self, message, **metadata):
        self.log('ERROR', message, **metadata)

# Create global logger instance
logger = LogsClient()`,
          language: "python"
        },
        {
          title: "Flask Integration",
          description: "Automatic request logging for Flask applications",
          code: `# app.py (Flask integration)
from flask import Flask, request, g
from logs_client import logger
import time

app = Flask(__name__)

@app.before_request
def log_request():
    g.start_time = time.time()
    logger.info('HTTP Request', 
        method=request.method,
        path=request.path,
        query=dict(request.args),
        ip=request.remote_addr
    )

@app.after_request
def log_response(response):
    duration = (time.time() - g.start_time) * 1000  # ms
    
    if response.status_code >= 500:
        level = 'error'
    elif response.status_code >= 400:
        level = 'warn'
    else:
        level = 'info'
    
    getattr(logger, level)('HTTP Response',
        method=request.method,
        path=request.path,
        status=response.status_code,
        duration_ms=duration
    )
    
    return response

# Your routes here
@app.route('/')
def index():
    logger.info('Index page accessed')
    return 'Hello World'

if __name__ == '__main__':
    app.run()`,
          language: "python"
        }
      ]
    },
    go: {
      label: "Go",
      icon: "bi-filetype-go",
      samples: [
        {
          title: "Basic Setup",
          description: "Minimal integration for Go applications",
          code: `// logs_client.go
package logs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	MaxBatchSize  = 1000
	FlushInterval = 5 * time.Second
)

var (
	logsAPIURL = os.Getenv("LOGS_API_URL")
	logsAPIKey = os.Getenv("LOGS_API_KEY")
	serviceName = os.Getenv("SERVICE_NAME")
)

type LogEntry struct {
	Level     string                 \`json:"level"\`
	Message   string                 \`json:"message"\`
	Service   string                 \`json:"service"\`
	Metadata  map[string]interface{} \`json:"metadata,omitempty"\`
	Timestamp string                 \`json:"timestamp"\`
}

type LogsClient struct {
	batch      []LogEntry
	mu         sync.Mutex
	httpClient *http.Client
}

func NewLogsClient() *LogsClient {
	if logsAPIURL == "" {
		logsAPIURL = "http://localhost:8082"
	}
	if serviceName == "" {
		serviceName = "my-app"
	}
	
	client := &LogsClient{
		batch:      make([]LogEntry, 0, MaxBatchSize),
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
	
	// Start flush worker
	go client.flushWorker()
	
	return client
}

func (c *LogsClient) sendBatch(entries []LogEntry) error {
	if len(entries) == 0 {
		return nil
	}
	
	body, err := json.Marshal(map[string]interface{}{
		"entries": entries,
	})
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}
	
	req, err := http.NewRequest("POST", logsAPIURL+"/api/logs/batch", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request creation error: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+logsAPIKey)
	
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	
	return nil
}

func (c *LogsClient) flushWorker() {
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		c.Flush()
	}
}

func (c *LogsClient) Flush() {
	c.mu.Lock()
	if len(c.batch) == 0 {
		c.mu.Unlock()
		return
	}
	
	toSend := make([]LogEntry, len(c.batch))
	copy(toSend, c.batch)
	c.batch = c.batch[:0]
	c.mu.Unlock()
	
	if err := c.sendBatch(toSend); err != nil {
		fmt.Printf("Error sending logs: %v\\n", err)
	}
}

func (c *LogsClient) log(level, message string, metadata map[string]interface{}) {
	entry := LogEntry{
		Level:     level,
		Message:   message,
		Service:   serviceName,
		Metadata:  metadata,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	
	c.mu.Lock()
	c.batch = append(c.batch, entry)
	shouldFlush := len(c.batch) >= MaxBatchSize
	c.mu.Unlock()
	
	if shouldFlush {
		c.Flush()
	}
}

func (c *LogsClient) Debug(message string, metadata map[string]interface{}) {
	c.log("DEBUG", message, metadata)
}

func (c *LogsClient) Info(message string, metadata map[string]interface{}) {
	c.log("INFO", message, metadata)
}

func (c *LogsClient) Warn(message string, metadata map[string]interface{}) {
	c.log("WARNING", message, metadata)
}

func (c *LogsClient) Error(message string, metadata map[string]interface{}) {
	c.log("ERROR", message, metadata)
}

// Global logger instance
var Logger = NewLogsClient()`,
          language: "go"
        },
        {
          title: "Gin Middleware",
          description: "Automatic request logging for Gin applications",
          code: `// middleware/logging.go
package middleware

import (
	"time"
	"your-project/logs"
	
	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// Log request
		logs.Logger.Info("HTTP Request", map[string]interface{}{
			"method": c.Request.Method,
			"path":   c.Request.URL.Path,
			"query":  c.Request.URL.RawQuery,
			"ip":     c.ClientIP(),
		})
		
		// Process request
		c.Next()
		
		// Log response
		duration := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		
		var level string
		switch {
		case status >= 500:
			level = "ERROR"
		case status >= 400:
			level = "WARNING"
		default:
			level = "INFO"
		}
		
		meta := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      status,
			"duration_ms": duration,
		}
		
		switch level {
		case "ERROR":
			logs.Logger.Error("HTTP Response", meta)
		case "WARNING":
			logs.Logger.Warn("HTTP Response", meta)
		default:
			logs.Logger.Info("HTTP Response", meta)
		}
	}
}

// Usage in main.go:
// r := gin.Default()
// r.Use(middleware.LoggingMiddleware())`,
          language: "go"
        }
      ]
    }
  };
  const setupSteps = [
    {
      number: 1,
      title: "Create a Project",
      description: "Go to the Projects page and create a new project for your application.",
      link: "/projects"
    },
    {
      number: 2,
      title: "Copy API Key",
      description: "Copy the generated API key (shown only once). Store it securely as an environment variable."
    },
    {
      number: 3,
      title: "Download Sample Code",
      description: "Select your language below and copy the sample code. Customize service name and metadata as needed."
    },
    {
      number: 4,
      title: "Configure Environment",
      description: "Set environment variables: LOGS_API_URL, LOGS_API_KEY, SERVICE_NAME"
    },
    {
      number: 5,
      title: "Deploy & Verify",
      description: "Deploy your application and check the Health Dashboard to see logs appearing in real-time."
    }
  ];
  return /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "container-fluid py-4", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center mb-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: "mb-1", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-book me-2" }),
          "Integration Documentation"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted mb-0", children: "Copy-paste examples for integrating external applications with the Logs service" })
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: "/projects", className: "btn btn-outline-primary", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-folder me-2" }),
        "Manage Projects"
      ] })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card mb-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "card-title mb-4", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-list-ol me-2" }),
        "Quick Setup Guide"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row", children: setupSteps.map((step) => /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-4 mb-3", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "flex-shrink-0", children: /* @__PURE__ */ jsxRuntimeExports.jsx(
          "div",
          {
            className: "rounded-circle bg-primary text-white d-flex align-items-center justify-content-center",
            style: { width: "32px", height: "32px", fontSize: "14px" },
            children: step.number
          }
        ) }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "flex-grow-1 ms-3", children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("h6", { className: "mb-1", children: step.title }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted small mb-0", children: step.description }),
          step.link && /* @__PURE__ */ jsxRuntimeExports.jsxs(Link, { to: step.link, className: "small", children: [
            "Go to Projects ",
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-arrow-right" })
          ] })
        ] })
      ] }) }, step.number)) })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card-header", children: /* @__PURE__ */ jsxRuntimeExports.jsx("ul", { className: "nav nav-tabs card-header-tabs", role: "tablist", children: Object.entries(languages).map(([key, lang]) => /* @__PURE__ */ jsxRuntimeExports.jsx("li", { className: "nav-item", role: "presentation", children: /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          className: `nav-link ${activeTab === key ? "active" : ""}`,
          onClick: () => setActiveTab(key),
          type: "button",
          role: "tab",
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `${lang.icon} me-2` }),
            lang.label
          ]
        }
      ) }, key)) }) }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card-body", children: Object.entries(languages).map(([key, lang]) => /* @__PURE__ */ jsxRuntimeExports.jsx(
        "div",
        {
          className: `tab-pane ${activeTab === key ? "active" : "d-none"}`,
          role: "tabpanel",
          children: lang.samples.map((sample, idx) => /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: idx > 0 ? "mt-4" : "", children: [
            /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "d-flex justify-content-between align-items-center mb-2", children: [
              /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
                /* @__PURE__ */ jsxRuntimeExports.jsx("h5", { className: "mb-0", children: sample.title }),
                /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted small mb-0", children: sample.description })
              ] }),
              /* @__PURE__ */ jsxRuntimeExports.jsxs(
                "button",
                {
                  className: "btn btn-sm btn-outline-secondary",
                  onClick: () => handleCopyCode(sample.code, `${key}-${idx}`),
                  children: [
                    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: `bi bi-${copiedCode === `${key}-${idx}` ? "check" : "clipboard"} me-1` }),
                    copiedCode === `${key}-${idx}` ? "Copied!" : "Copy"
                  ]
                }
              )
            ] }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("pre", { className: "bg-dark text-light p-3 rounded", style: { fontSize: "0.85rem", maxHeight: "500px", overflow: "auto" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: sample.code }) })
          ] }, idx))
        },
        key
      )) })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card mt-4", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h5", { className: "card-title mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-gear me-2" }),
        "Environment Variables"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("table", { className: "table table-sm", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("thead", { children: /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Variable" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Description" }),
          /* @__PURE__ */ jsxRuntimeExports.jsx("th", { children: "Example" })
        ] }) }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("tbody", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "LOGS_API_URL" }) }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: "URL of the Logs service API" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "http://localhost:8082" }) })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "LOGS_API_KEY" }) }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: "API key from Projects page (Bearer token)" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "proj_abc123..." }) })
          ] }),
          /* @__PURE__ */ jsxRuntimeExports.jsxs("tr", { children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "SERVICE_NAME" }) }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: "Name of your service (appears in logs)" }),
            /* @__PURE__ */ jsxRuntimeExports.jsx("td", { children: /* @__PURE__ */ jsxRuntimeExports.jsx("code", { children: "my-api" }) })
          ] })
        ] })
      ] })
    ] }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "alert alert-info mt-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("h6", { className: "alert-heading", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-lightning-charge me-2" }),
        "Performance Tips"
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("ul", { className: "mb-0", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsxs("li", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Batch size:" }),
          " Use batches of 100-1000 logs for optimal performance (14,000-33,000 logs/second)"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("li", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Flush interval:" }),
          " 5 seconds is recommended for balancing real-time visibility and throughput"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("li", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Rate limits:" }),
          " 1000 requests per minute per API key (contact admin if you need higher limits)"
        ] }),
        /* @__PURE__ */ jsxRuntimeExports.jsxs("li", { children: [
          /* @__PURE__ */ jsxRuntimeExports.jsx("strong", { children: "Metadata:" }),
          " Keep metadata concise to reduce payload size and improve query performance"
        ] })
      ] })
    ] })
  ] });
}
function generateCodeVerifier() {
  const array = new Uint8Array(32);
  window.crypto.getRandomValues(array);
  return base64URLEncode(array);
}
async function generateCodeChallenge(verifier) {
  const encoder = new TextEncoder();
  const data = encoder.encode(verifier);
  const digest = await window.crypto.subtle.digest("SHA-256", data);
  return base64URLEncode(new Uint8Array(digest));
}
function base64URLEncode(buffer) {
  const base64 = btoa(String.fromCharCode(...buffer));
  return base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}
function base64URLDecode(str) {
  const base64 = str.replace(/-/g, "+").replace(/_/g, "/");
  const binary = atob(base64);
  const bytes = new Uint8Array(binary.length);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return bytes;
}
async function openKeyDatabase() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open("devsmith-oauth", 1);
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    request.onupgradeneeded = (event) => {
      const db2 = event.target.result;
      if (!db2.objectStoreNames.contains("keys")) {
        db2.createObjectStore("keys");
      }
    };
  });
}
async function getOrCreateEncryptionKey() {
  const db2 = await openKeyDatabase();
  const transaction = db2.transaction("keys", "readonly");
  const store = transaction.objectStore("keys");
  const getRequest = store.get("encryption-key");
  const existingKey = await new Promise((resolve) => {
    getRequest.onsuccess = () => resolve(getRequest.result);
    getRequest.onerror = () => resolve(null);
  });
  if (existingKey) {
    db2.close();
    return existingKey;
  }
  const newKey = await window.crypto.subtle.generateKey(
    { name: "AES-GCM", length: 256 },
    false,
    // not extractable (cannot be exported)
    ["encrypt", "decrypt"]
  );
  const writeTransaction = db2.transaction("keys", "readwrite");
  const writeStore = writeTransaction.objectStore("keys");
  writeStore.put(newKey, "encryption-key");
  await new Promise((resolve, reject) => {
    writeTransaction.oncomplete = () => resolve();
    writeTransaction.onerror = () => reject(writeTransaction.error);
  });
  db2.close();
  return newKey;
}
async function encryptVerifier(verifier) {
  const encoder = new TextEncoder();
  const payload = {
    verifier,
    timestamp: Date.now(),
    nonce: base64URLEncode(window.crypto.getRandomValues(new Uint8Array(16)))
  };
  const data = encoder.encode(JSON.stringify(payload));
  const key = await getOrCreateEncryptionKey();
  const iv = window.crypto.getRandomValues(new Uint8Array(12));
  const encrypted = await window.crypto.subtle.encrypt(
    { name: "AES-GCM", iv },
    key,
    data
  );
  const combined = new Uint8Array(iv.length + encrypted.byteLength);
  combined.set(iv, 0);
  combined.set(new Uint8Array(encrypted), iv.length);
  return base64URLEncode(combined);
}
async function decryptVerifier(encryptedState) {
  try {
    const combined = base64URLDecode(encryptedState);
    const iv = combined.slice(0, 12);
    const ciphertext = combined.slice(12);
    const key = await getOrCreateEncryptionKey();
    const decrypted = await window.crypto.subtle.decrypt(
      { name: "AES-GCM", iv },
      key,
      ciphertext
    );
    const decoder = new TextDecoder();
    const payload = JSON.parse(decoder.decode(decrypted));
    const age = Date.now() - payload.timestamp;
    if (age > 6e5) {
      throw new Error("State expired (>10 minutes old)");
    }
    return payload.verifier;
  } catch (error) {
    if (error.message === "State expired (>10 minutes old)") {
      throw error;
    }
    throw new Error("Invalid or tampered state parameter");
  }
}
function LoginPage() {
  const [email, setEmail] = reactExports.useState("");
  const [password, setPassword] = reactExports.useState("");
  const [pkceError, setPkceError] = reactExports.useState(null);
  const { login, isAuthenticated, error } = useAuth();
  const navigate = useNavigate();
  reactExports.useEffect(() => {
    if (isAuthenticated) {
      navigate("/");
    }
  }, [isAuthenticated, navigate]);
  const handleSubmit = async (e) => {
    e.preventDefault();
    await login(email, password);
  };
  const handleGitHubLogin = async () => {
    try {
      const codeVerifier = generateCodeVerifier();
      const codeChallenge = await generateCodeChallenge(codeVerifier);
      const encryptedState = await encryptVerifier(codeVerifier);
      console.log("[PKCE] Generated encrypted state (verifier embedded)");
      const clientId = "Ov23liaV4He3p1k7VziT";
      if (!clientId) ;
      const params = new URLSearchParams({
        client_id: clientId,
        redirect_uri: window.location.origin + "/oauth/pkce-callback",
        scope: "user:email read:user",
        state: encryptedState,
        // Contains encrypted verifier + timestamp + nonce
        code_challenge: codeChallenge,
        code_challenge_method: "S256"
      });
      const authURL = `https://github.com/login/oauth/authorize?${params}`;
      console.log("[PKCE] Redirecting to GitHub with encrypted state and code_challenge");
      window.location.href = authURL;
    } catch (error2) {
      console.error("[PKCE] Failed to generate PKCE parameters:", error2);
      setPkceError("Failed to initiate login. Please try again.");
    }
  };
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "container", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row justify-content-center align-items-center", style: { minHeight: "100vh" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-4", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "card shadow", children: /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "card-body p-5", children: [
    /* @__PURE__ */ jsxRuntimeExports.jsxs("h2", { className: "text-center mb-4", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-code-square text-primary", style: { fontSize: "2.5rem" } }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mt-2", children: "DevSmith Platform" })
    ] }),
    (error || pkceError) && /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "alert alert-danger", role: "alert", children: error || pkceError }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("form", { onSubmit: handleSubmit, children: [
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "email", className: "form-label", children: "Email" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "input",
          {
            type: "email",
            className: "form-control",
            id: "email",
            value: email,
            onChange: (e) => setEmail(e.target.value),
            required: true
          }
        )
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "mb-3", children: [
        /* @__PURE__ */ jsxRuntimeExports.jsx("label", { htmlFor: "password", className: "form-label", children: "Password" }),
        /* @__PURE__ */ jsxRuntimeExports.jsx(
          "input",
          {
            type: "password",
            className: "form-control",
            id: "password",
            value: password,
            onChange: (e) => setPassword(e.target.value),
            required: true
          }
        )
      ] }),
      /* @__PURE__ */ jsxRuntimeExports.jsx("button", { type: "submit", className: "btn btn-primary w-100 mb-3", children: "Login" })
    ] }),
    /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { className: "text-center", children: [
      /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "mb-2", children: "or" }),
      /* @__PURE__ */ jsxRuntimeExports.jsxs(
        "button",
        {
          type: "button",
          className: "btn btn-dark w-100",
          onClick: handleGitHubLogin,
          children: [
            /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-github me-2" }),
            "Login with GitHub"
          ]
        }
      )
    ] })
  ] }) }) }) }) });
}
function OAuthCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [error, setError] = reactExports.useState(null);
  console.log("[OAuthCallback] Component mounted");
  console.log("[OAuthCallback] Search params:", searchParams.toString());
  reactExports.useEffect(() => {
    console.log("[OAuthCallback] useEffect triggered");
    const code = searchParams.get("code");
    const encryptedState = searchParams.get("state");
    const errorParam = searchParams.get("error");
    console.log("[OAuthCallback] Code from URL:", code);
    console.log("[OAuthCallback] Encrypted state from URL:", encryptedState);
    console.log("[OAuthCallback] Error from URL:", errorParam);
    if (errorParam) {
      console.log("[OAuthCallback] Error detected, showing error and redirecting");
      setError("GitHub authentication failed. Please try again.");
      setTimeout(() => navigate("/login"), 3e3);
      return;
    }
    if (!code) {
      console.log("[OAuthCallback] No authorization code detected, showing error and redirecting");
      setError("No authorization code received.");
      setTimeout(() => navigate("/login"), 3e3);
      return;
    }
    if (!encryptedState) {
      console.error("[PKCE] Missing encrypted state parameter");
      setError("Security validation failed. Please try again.");
      setTimeout(() => navigate("/login"), 3e3);
      return;
    }
    const exchangeCodeForToken = async () => {
      try {
        console.log("[PKCE] Decrypting verifier from state...");
        const codeVerifier = await decryptVerifier(encryptedState);
        console.log("[PKCE] Verifier decrypted successfully");
        console.log("[PKCE] Exchanging code for token...");
        const response = await fetch("/api/portal/auth/token", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            code,
            state: encryptedState,
            // Send for audit logging
            code_verifier: codeVerifier
          })
        });
        if (!response.ok) {
          const errorData = await response.json().catch(() => ({ error: "Unknown error" }));
          console.error("[PKCE] Token exchange failed:", errorData);
          throw new Error(errorData.error || "Token exchange failed");
        }
        const data = await response.json();
        console.log("[PKCE] Token exchange successful");
        localStorage.setItem("devsmith_token", data.token);
        console.log("[PKCE] Token stored in localStorage");
        console.log("[PKCE] Redirecting to dashboard");
        window.location.href = "/";
      } catch (err) {
        console.error("[PKCE] Token exchange error:", err);
        let errorMessage = "Failed to complete authentication";
        if (err.message.includes("expired")) {
          errorMessage = "Login session expired (>10 minutes). Please try again.";
        } else if (err.message.includes("Invalid or tampered")) {
          errorMessage = "Security validation failed. Please try again.";
        } else {
          errorMessage = `${errorMessage}: ${err.message}`;
        }
        setError(errorMessage);
        setTimeout(() => navigate("/login"), 3e3);
      }
    };
    exchangeCodeForToken();
  }, [searchParams, navigate]);
  return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "container", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "row justify-content-center align-items-center", style: { minHeight: "100vh" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "col-md-6 col-lg-4 text-center", children: error ? /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("i", { className: "bi bi-exclamation-triangle text-danger", style: { fontSize: "3rem" } }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3", children: "Authentication Error" }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: error }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: "Redirecting to login..." })
  ] }) : /* @__PURE__ */ jsxRuntimeExports.jsxs("div", { children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", style: { width: "3rem", height: "3rem" }, children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("h3", { className: "mt-3", children: "Authenticating with GitHub..." }),
    /* @__PURE__ */ jsxRuntimeExports.jsx("p", { className: "text-muted", children: "Please wait while we complete your login." })
  ] }) }) }) });
}
function ProtectedRoute({ children }) {
  const { isAuthenticated, loading } = useAuth();
  if (loading) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "container mt-5 text-center", children: /* @__PURE__ */ jsxRuntimeExports.jsx("div", { className: "spinner-border text-primary", role: "status", children: /* @__PURE__ */ jsxRuntimeExports.jsx("span", { className: "visually-hidden", children: "Loading..." }) }) });
  }
  if (!isAuthenticated) {
    return /* @__PURE__ */ jsxRuntimeExports.jsx(Navigate, { to: "/login", replace: true });
  }
  return children;
}
function App() {
  reactExports.useEffect(() => {
    setupGlobalErrorHandlers();
  }, []);
  return /* @__PURE__ */ jsxRuntimeExports.jsx(BrowserRouter, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(AuthProvider, { children: /* @__PURE__ */ jsxRuntimeExports.jsxs(Routes, { children: [
    /* @__PURE__ */ jsxRuntimeExports.jsx(Route, { path: "/login", element: /* @__PURE__ */ jsxRuntimeExports.jsx(LoginPage, {}) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(Route, { path: "/oauth/pkce-callback", element: /* @__PURE__ */ jsxRuntimeExports.jsx(OAuthCallback, {}) }),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(Dashboard, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/portal",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(Dashboard, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/health",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(HealthPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/review",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(ReviewPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/analytics",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(AnalyticsPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/llm-config",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(LLMConfigPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/projects",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(ProjectsPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(
      Route,
      {
        path: "/integration-docs",
        element: /* @__PURE__ */ jsxRuntimeExports.jsx(ProtectedRoute, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(IntegrationDocsPage, {}) })
      }
    ),
    /* @__PURE__ */ jsxRuntimeExports.jsx(Route, { path: "*", element: /* @__PURE__ */ jsxRuntimeExports.jsx(Navigate, { to: "/", replace: true }) })
  ] }) }) });
}
createRoot(document.getElementById("root")).render(
  /* @__PURE__ */ jsxRuntimeExports.jsx(reactExports.StrictMode, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(ThemeProvider, { children: /* @__PURE__ */ jsxRuntimeExports.jsx(App, {}) }) })
);
