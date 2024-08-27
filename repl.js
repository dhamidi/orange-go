// Useful helper functions for the REPL

// prop :: String -> Object -> a
function prop(key) {
  return function (obj) {
    return obj[key];
  };
}
prop.help =
  "prop(String) -> Object -> a - Returns the value of the key in the object";

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
submissions.help =
  "submissions() -> [Object] - Returns an array of all submissions";

// help :: () -> IO ()
function help() {
  let args = Array.prototype.slice.call(arguments);
  if (args.length > 0 && args[0].help) {
    print(args[0].help);
    return;
  }
  print("Available functions:");
  Object.keys(help.functions).forEach(function (fname) {
    print(help.functions[fname].help || key);
  });
}

help.functions = {
  submissions: submissions,
  prop: prop,
};
