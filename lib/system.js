const { Message } = require('@projectriff/message');
const { USER_FUNCTION_URI } = process.env;

function getFunction(uri) {
  try {
    const mod = require(uri);
  } catch (e) {
    throw `Could not locate user function at ${USER_FUNCTION_URI}: ${e}`
  }
  if (mod.__esModule && typeof mod.default === 'function') {
      // transpiled ES Module interop
      return mod.default;
  }
  return mod;
}

const userFn = getFunction(USER_FUNCTION_URI);

module.exports = ({headers, payload}) => {
  return userFn(payload);
};

module.exports.$argumentType = 'message';

Message.install();
