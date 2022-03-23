var Color = {
    RESET: "\x1b[39;49;00m", Black: "0;01", Blue: "4;01", Cyan: "6;01", Gray: "7;11", Green: "2;01", Purple: "5;01", Red: "1;01", Yellow: "3;01",
    Light: {
        Black: "0;11", Blue: "4;11", Cyan: "6;11", Gray: "7;01", Green: "2;11", Purple: "5;11", Red: "1;11", Yellow: "3;11"
    }
};

/**
 *
 * @param input.
 *      If an object is passed it will print as json
 * @param kwargs  options map {
 *     -l level: string;   log/warn/error
 *     -i indent: boolean;     print JSON prettify
 *     -c color: @see ColorMap
 * }
 */
var LOG = function (input, kwargs) {
    kwargs = kwargs || {};
    var logLevel = kwargs['l'] || 'log', colorPrefix = '\x1b[3', colorSuffix = 'm';
    if (typeof input === 'object')
        input = JSON.stringify(input, null, kwargs['i'] ? 2 : null);
    if (kwargs['c'])
        input = colorPrefix + kwargs['c'] + colorSuffix + input + Color.RESET;
    console[logLevel](input);
    send(input);
};

var printBacktrace = function () {
    Java.perform(function() {
        var android_util_Log = Java.use('android.util.Log'), java_lang_Exception = Java.use('java.lang.Exception');
        // getting stacktrace by throwing an exception
        LOG(android_util_Log.getStackTraceString(java_lang_Exception.$new()), { c: Color.Gray });
    });
};

function traceClass(targetClass) {
    var hook;
    try {
        hook = Java.use(targetClass);
    } catch (e) {
        console.error("trace class failed", e);
        return;
    }

    var methods = hook.class.getDeclaredMethods();
    hook.$dispose();

    var parsedMethods = [];
    methods.forEach(function (method) {
        var methodStr = method.toString();
        var methodReplace = methodStr.replace(targetClass + ".", "TOKEN").match(/\sTOKEN(.*)\(/)[1];
         parsedMethods.push(methodReplace);
    });

    uniqBy(parsedMethods, JSON.stringify).forEach(function (targetMethod) {
        traceMethod(targetClass + '.' + targetMethod);
    });
}

function traceMethod(targetClassMethod) {
    var delim = targetClassMethod.lastIndexOf('.');
    if (delim === -1)
        return;

    var targetClass = targetClassMethod.slice(0, delim);
    var targetMethod = targetClassMethod.slice(delim + 1, targetClassMethod.length);

    var hook = Java.use(targetClass);
    var overloadCount = hook[targetMethod].overloads.length;

    LOG({ tracing: targetClassMethod, overloaded: overloadCount }, { c: Color.Green });

    for (var i = 0; i < overloadCount; i++) {
        hook[targetMethod].overloads[i].implementation = function () {
            var log = { '#': targetClassMethod, args: [] };

            for (var j = 0; j < arguments.length; j++) {
                var arg = arguments[j];
                // quick&dirty fix for java.io.StringWriter char[].toString() impl because frida prints [object Object]
                if (j === 0 && arguments[j]) {
                    if (arguments[j].toString() === '[object Object]') {
                        var s = [];
                        for (var k = 0, l = arguments[j].length; k < l; k++) {
                            s.push(arguments[j][k]);
                        }
                        arg = s.join('');
                    }
                }
                log.args.push({ i: j, o: arg, s: arg ? arg.toString(): 'null'});
            }

            var retval;
            try {
                retval = this[targetMethod].apply(this, arguments); // might crash (Frida bug?)
                log.returns = { val: retval, str: retval ? retval.toString() : null };
            } catch (e) {
                console.error(e);
            }
            LOG(log, { c: Color.Blue });
            return retval;
        }
    }
}

// remove duplicates from array
function uniqBy(array, key) {
    var seen = {};
    return array.filter(function (item) {
        var k = key(item);
        return seen.hasOwnProperty(k) ? false : (seen[k] = true);
    });
}

function traceStringBuilder() {
         var StringBuilder = Java.use('java.lang.StringBuilder');
//     StringBuilder.$init.overload('java.lang.String').implementation = function(str) {
//                         var bt = Java.use("android.util.Log").getStackTraceString(Java.use("java.lang.Exception").$new());
//                                 console.log("\nBacktrace:\n" + bt);

//         console.log("StringBuilder(" + str + ")")
//         return this.$init.call(this, str);
//     }
//     StringBuilder.append.overload('java.lang.String').implementation = function(str) {
//         console.log("StringBuilder.append(" + str + ")")

// //                      var bt = Java.use("android.util.Log").getStackTraceString(Java.use("java.lang.Exception").$new());
// //                              console.log("\nBacktrace:\n" + bt);

//         return this.append.call(this, str);

//     }
    StringBuilder.toString.implementation = function() {
        var result = this.toString.call(this);

        LOG(result);
        return result;
    }

}

var Main = function() {
    Java.perform(function () { // avoid java.lang.ClassNotFoundException
    traceStringBuilder();
    // traceMethod("com.akamai.botman.B")
        [
            // "java.io.File",
	    // 'android.support.v4.common.h84'
            //'com.akamai.botman.j',
            // 'com.akamai.botman.B',
	    // 'com.akamai.botman.f',
        // 'com.akamai.botman.F'
        ].forEach(traceClass);
//	    Java.use('java.lang.StringBuilder').toString().overload().implementation = function() {
//		LOG(arguments[0]);
//		    return
//	    }

//        Java.use('java.net.Socket').isConnected.overload().implementation = function () {
//            LOG('Socket.isConnected.overload', { c: Color.Light.Cyan });
//            printBacktrace();
//            return true;
//        }
    });
};

Java.perform(Main);
