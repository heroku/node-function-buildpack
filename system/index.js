const { Message } = require('@projectriff/message');

const { DEBUG, MIDDLEWARE_FUNCTION_URI, USER_FUNCTION_URI } = process.env;

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

const userFn = getFunction(USER_FUNCTION_URI);

module.exports = async ({headers, payload}) => {
  if (DEBUG) {
    console.log('==System Function Start==');
    console.log(`HEADERS: ${JSON.stringify(headers)}`);
    console.log(`ORIGINAL PAYLOAD: ${payload}`);
    console.log(`MIDDLEWARE_FUNCTION_URI: ${MIDDLEWARE_FUNCTION_URI}`);
    console.log('==Middleware Function(s) Start==');
  }

  if (MIDDLEWARE_FUNCTION_URI) {
    await Promise.all(MIDDLEWARE_FUNCTION_URI.split(':').map(async (middleware) => {
          try {
            const middlewareFn = await getFunction(middleware);
            if (DEBUG) {
              console.log(`MIDDLEWARE RECEIVED PAYLOAD: ${payload}`);
            }
            payload = middlewareFn(payload);
          } catch (err) {
            throw err
          }
        })
    );
  }

  if (DEBUG) {
    console.log('==Middleware Function(s) End==');
    console.log(`USER FUNCTION RECEIVED PAYLOAD: ${payload}`);
  }
  const result = await userFn(payload);

  if (DEBUG) {
    console.log('RESULT', result);
    console.log('==System Function End==');
  }
  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
