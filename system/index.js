const { Message } = require('@projectriff/message');

const { USER_FUNCTION_URI, DEBUG } = process.env;

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
  if DEBUG {
    console.log('==System Function Start==');
    console.log('HEADERS:', headers);
    console.log('PAYLOAD:', payload);
  }
  const result = await userFn(payload);
  if DEBUG {
    console.log('RESULT', result);
    console.log('==System Function End==');
  }
  return result;
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;

