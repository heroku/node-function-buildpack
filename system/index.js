const { Message } = require('@projectriff/message');
const SF = require('@salesforce/core');

const { DEBUG, MIDDLEWARE_FUNCTION_URI, USER_FUNCTION_URI } = process.env;

async function createLogger(level, requestID) {
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

const middlewareFns = getMiddlewareFunctions(MIDDLEWARE_FUNCTION_URI);
const userFn = getFunction(USER_FUNCTION_URI);

module.exports = async (message) => {
  const payload = message.payload;

  // Remap headers to a standard JS object
  const headers = message.headers.toRiffHeaders();
  Object.keys(headers).map((key) => { headers[key] = message.headers.getValue(key) });

  const logLevel = DEBUG ? SF.LoggerLevel.DEBUG : SF.LoggerLevel.INFO;
  const requestId = headers['ce-id'] || headers['x-request-id'];
  const logger = await createLogger(logLevel, requestId);

  const state = {};
  let middlewareResult = [payload, logger];

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
      logger.error({error});
      throw error;
    }
  }));

  middlewareResult = middlewareResult.concat(logger);

  let result;
  try {
    result = await userFn(...middlewareResult);
  } catch (error) {
    logger.error({error});
    throw error;
  }

  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
