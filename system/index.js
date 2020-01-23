const { Message } = require('@projectriff/message');
const winston = require('winston');

const { DEBUG, MIDDLEWARE_FUNCTION_URI, USER_FUNCTION_URI } = process.env;

function initLogging(level, requestID) {
  let logger = winston.createLogger({
    level,
    transports: [
      new winston.transports.Console(),
    ],
  })

  if (requestID) {
    logger = logger.child({
      request_id: requestID,
    });
  }

  return logger
}

function getMiddlewareFunctions(uri) {
  const middlewareFunctions = [];
  if (uri) {
    uri.split(':').forEach(mw => middlewareFunctions.push(getFunction(mw)))
  }
  return middlewareFunctions;
}

function getFunction(uri) {
  let mod;
  try {
    mod = require(uri);
  } catch (e) {
    throw `Could not locate user function at ${uri}: ${e}`
  }
  if (mod.__esModule && typeof mod.default === 'function') {
    return mod.default;
  }
  return mod;
}

const middlewareFns = getMiddlewareFunctions(MIDDLEWARE_FUNCTION_URI);
const userFn = getFunction(USER_FUNCTION_URI);

module.exports = async (message) => {
  const payload = message.payload;
  // Remap headers to a standard JS object
  const headers = message.headers.toRiffHeaders();
  Object.keys(headers).map((key) => { headers[key] = message.headers.getValue(key) });

  const logLevel = DEBUG ? 'debug' : 'info';
  const log = initLogging(logLevel, headers['x-request-id']);

  log.debug({
    payload_length: JSON.stringify(payload).length,
    middware_function_uri: MIDDLEWARE_FUNCTION_URI,
    status: 'system function start',
  });

  log.debug({ status: 'middleware function(s) start' });

  const state = {};
  let middlewareResult = [payload];

  await Promise.all(middlewareFns.map(async (middleware) => {
    try {
      // input should be immutable
      const input = {
        payload: typeof payload == "object" ? Object.assign({}, payload) : payload,
        headers: Object.assign({}, headers)
      };

      log.debug({
        middleware_input: JSON.stringify(input),
        middleware_state: JSON.stringify(state),
        middleware_result: JSON.stringify(middlewareResult),
      });

      middlewareResult = await middleware(input, state, middlewareResult);

      log.debug({
        middleware_returned: JSON.stringify(middlewareResult),
      });

      if (!Array.isArray(middlewareResult)) {
        throw new Error('Invalid return type, middleware must return an array of arguments')
      }

    } catch (error) {
      log.error({ error });
      throw error;
    }
  }));

  log.debug({
    status: 'middleware function(s) end',
    user_func_args: JSON.stringify(middlewareResult),
  });

  middlewareResult.concat(log);

  let result;

  try {
    result = await userFn(...middlewareResult);

    log.debug({
      result,
      status: 'system function end',
    });

  } catch (error) {
    log.error({ error });
    throw error;
  }

  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
