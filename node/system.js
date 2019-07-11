const fn = (fn => {
    if (fn.__esModule && typeof fn.default === 'function') {
        // transpiled ES Module interop
        return fn.default;
    }
    return fn;
})(require(USER_FUNCTION_URI));

const middlewares = [];

module.exports = (args) => {
  for each middleware in middlewares {
    middleware.apply(args)
  }
  fn.apply(args)
}
