const { Message } = require('@projectriff/message');
const Unmarshaller = require('../../node_modules/cloudevents-sdk/lib/bindings/http/unmarshaller.js');

const { USER_FUNCTION_URI } = process.env;

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
  console.log('HEADERS:', headers);
  console.log('PAYLOAD:', payload);
  let ce, data;
  try {
    ce = Unmarshaller.unmarhall(payload, headers);
    data = ce.getData();
  } catch (e) {
    console.log("Couldn't unmarshall message into a cloudevent:", e);
    data = payload
  }

  return await userFn(data);
};

module.exports.$argumentType = 'message';
module.exports.$init = userFn.$init;
module.exports.$destroy = userFn.$destroy;
