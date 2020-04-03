const { Message } = require('@projectriff/message');
const SF = require('@salesforce/core');

const { DEBUG, MIDDLEWARE_FUNCTION_URI, USER_FUNCTION_URI } = process.env;

function createLogger(requestID) {
  const level = DEBUG ? SF.LoggerLevel.DEBUG : SF.LoggerLevel.INFO;
  const logger = new SF.Logger('Evergreen Logger');
  logger.addStream({stream: process.stderr});
  logger.setLevel(level);

  if (requestID) {
    logger.addField('request_id', requestID);
  }

  return logger;
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

const systemLogger = createLogger();

let middlewareFns;
let userFn;
try {
  middlewareFns = getMiddlewareFunctions(MIDDLEWARE_FUNCTION_URI);
  userFn = getFunction(USER_FUNCTION_URI);
} catch (error) {
  systemLogger.error(error.toString());
  throw error;
}

module.exports = async (message) => {
  const payload = message.payload;

  // Remap headers to a standard JS object
  const headers = message.headers.toRiffHeaders();
  Object.keys(headers).map((key) => { headers[key] = message.headers.getValue(key) });

  const requestId = headers['ce-id'] || headers['x-request-id'];
  const requestLogger = createLogger(requestId);

  const state = {};
  let middlewareResult = [payload, requestLogger];

  await Promise.all(middlewareFns.map(async (middleware) => {
    try {
      // input should be immutable
      const input = {
        payload: typeof payload == "object" ? Object.assign({}, payload) : payload,
        headers: Object.assign({}, headers)
      };

      middlewareResult = await middleware(input, state, middlewareResult);
      if (!Array.isArray(middlewareResult)) {
        throw new Error('Invalid return type, middleware must return an array of arguments')
      }
    } catch (error) {
      requestLogger.error(error.toString());
      throw error;
    }
  }));

  let result;
  try {
    result = await userFn(...middlewareResult);
  } catch (error) {
    requestLogger.error(error.toString());
    throw error;
  }

  //if userFn does not have explicit return, it would be undefined, when || with null, it would be null
  //for Accept header, riff node invoker's application/json marshaller Buffer.from(JSON.stringify(null))
  return result || null;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
