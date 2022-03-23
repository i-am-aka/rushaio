var functionLogger = {};

functionLogger.log = true;//Set this to false to disable logging

/**
 * Gets a function that when called will log information about itself if logging is turned on.
 *
 * @param func The function to add logging to.
 * @param name The name of the function.
 *
 * @return A function that will perform logging and then call the function.
 */
functionLogger.getLoggableFunction = function(func, name) {
    return function() {
        if (functionLogger.log) {
            var logText = name + '(';

            for (var i = 0; i < arguments.length; i++) {
                if (i > 0) {
                    logText += ', ';
                }
                logText += JSON.stringify(arguments[i]);
            }
            logText += ');';

            console.trace();
            console.log(logText);
        }

        return func.apply(this, arguments);
    }
};

/**
 * After this is called, all direct children of the provided namespace object that are
 * functions will log their name as well as the values of the parameters passed in.
 *
 * @param namespaceObject The object whose child functions you'd like to add logging to.
 */
functionLogger.addLoggingToNamespace = function(namespaceObject){
    for(var name in namespaceObject){
        var potentialFunction = namespaceObject[name];

        if(Object.prototype.toString.call(potentialFunction) === '[object Function]'){
            Object.defineProperty(namespaceObject, name, {
                writable: true,
                value: functionLogger.getLoggableFunction(potentialFunction, name) });
        }
    }
};

functionLogger.accProxy = function(obj, name) {
    return new Proxy(obj, { get: function(obj, prop) {
        console.trace();
        if (typeof prop !== 'symbol') {
            console.log(`get ${name}.${prop}`)
        } else {
            console.log('get symbol');
        }
        return Reflect.get(...arguments);
    },
    set: function(obj, prop, value) {
            console.trace();
            if (typeof prop !== 'symbol'&&typeof value !== 'symbol') {
                console.log(`set ${name}.${prop}=${value}`)
            } else {
                console.log(`set ${name}.symbol=???`);
            }
            return Reflect.set(...arguments);
        },
     })
}

functionLogger.addAccessLoggingToNamespace = function(namespaceObject, namespace) {
    for(var name in namespaceObject){
        try {
            // console.log(Object.prototype.toString.call(namespaceObject[name]))
            if(Object.prototype.toString.call(namespaceObject[name]) !== '[object Function]'){
                Object.defineProperty(namespaceObject, name,
                    {
                        writable: true,
                        value: functionLogger.accProxy(namespaceObject[name], `${namespace}.${name}`)
                    }
                )
            }
        } catch{}
    }
};

export default functionLogger;

