// Useful helper functions for the REPL

// prop :: String -> Object -> a
function prop(key) {
  return function (obj) {
    return obj[key];
  };
}

// is :: String -> a -> Object -> Bool
function is(key) {
  return function (val) {
    return function (obj) {
      return obj[key] === val;
    };
  };
}

// submissions :: () -> [Object]
function submissions() {
  return g.content.Submissions;
}
