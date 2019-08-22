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
    console.log(`HEADERS: ${headers}`);
    console.log(`ORIGINAL PAYLOAD: ${payload}`);
    console.log(`MIDDLEWARE_FUNCTION_URI: ${MIDDLEWARE_FUNCTION_URI}`);
    console.log('==Middleware Function(s) Start==');
  }

  const req = { headers, payload };
  let res = [];

  await Promise.all(middlewareFns.map(async (middleware) => {
        try {
          if (DEBUG) {
            console.log(`MIDDLEWARE REQ: ${req}`);
          }
          res = await middleware(req);
        } catch (err) {
          throw err
        }
      })
  );

  if (DEBUG) {
    console.log('==Middleware Function(s) End==');
    console.log(`USER FUNCTION RECEIVED: ${res}`);
  }
  if (!res || res.length === 0) {
    res.push(payload)
  }
  const result = await userFn(...res);

  if (DEBUG) {
    console.log('RESULT', result);
    console.log('==System Function End==');
  }
  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
