const fn = (fn => {
    if (fn.__esModule && typeof fn.default === 'function') {
        // transpiled ES Module interop
        return fn.default;
    }
    return fn;
})(require(process.env.USER_FUNCTION_URI));


module.exports = (args) => {
  fn.apply(this, args)
}
