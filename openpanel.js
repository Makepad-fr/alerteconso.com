(function () {
  var host = window.location.hostname.replace(/^www\./, "");
  if (host !== "alerteconso.com") {
    return;
  }
  if (typeof Proxy !== "function") {
    return;
  }

  window.op =
    window.op ||
    (function () {
      var queue = [];
      return new Proxy(
        function () {
          if (arguments.length) {
            queue.push([].slice.call(arguments));
          }
        },
        {
          get: function (_target, property) {
            if (property === "q") {
              return queue;
            }
            return function () {
              queue.push([property].concat([].slice.call(arguments)));
            };
          },
          has: function (_target, property) {
            return property === "q";
          },
        },
      );
    })();

  window.op("init", {
    apiUrl: "https://analytics.makepad.fr/api",
    clientId: "71eaa394-fd89-46bb-a39a-fac1845c0e02",
    trackScreenViews: true,
    trackOutgoingLinks: true,
    trackAttributes: true,
  });

  var script = document.createElement("script");
  script.src = "https://openpanel.dev/op1.js";
  script.async = true;
  document.head.appendChild(script);
})();
