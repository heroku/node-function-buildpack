const { Message } = require('@projectriff/message');

const { DEBUG, MIDDLEWARE_FUNCTION_URI, USER_FUNCTION_URI } = process.env;

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

module.exports = async ({headers, payload}) => {
  if (DEBUG) {
    console.log('==System Function Start==');
    console.log(`HEADERS: ${JSON.stringify(headers)}`);
    console.log(`PAYLOAD: ${JSON.stringify(payload)}`);
    console.log(`MIDDLEWARE_FUNCTION_URI: ${MIDDLEWARE_FUNCTION_URI}`);
    console.log('==Middleware Function(s) Start==');
  }

  const state = {};
  let middlewareResult = [payload];

  await Promise.all(middlewareFns.map(async (middleware) => {
        try {
          // input should be immutable
          const input = {
            payload: typeof payload == "object" ? Object.assign({}, payload) : payload,
            headers: Object.assign({}, headers)
          };
          if (DEBUG) {
            console.log(`MIDDLEWARE INPUT: ${JSON.stringify(input)}`);
            console.log(`MIDDLEWARE STATE: ${JSON.stringify(state)}`);
            console.log(`MIDDLEWARE RESULT: ${JSON.stringify(middlewareResult)}`);
          }
          middlewareResult = await middleware(input, state, middlewareResult);
          if (DEBUG) {
            console.log(`MIDDLEWARE RETURNED: ${JSON.stringify(middlewareResult)}`);
          }
          if (!Array.isArray(middlewareResult)) {
            throw new Error('Invalid return type, middleware must return an array of arguments')
          }
        } catch (err) {
          throw err
        }
      })
  );

  if (DEBUG) {
    console.log('==Middleware Function(s) End==');
    console.log(`USER FUNCTION RECEIVED ARGS: ${JSON.stringify(middlewareResult)}`);
  }

  const result = await userFn(...middlewareResult);

  if (DEBUG) {
    console.log('RESULT', result);
    console.log('==System Function End==');
  }
  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
