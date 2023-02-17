const path = require("path")

var config = {
    module: {},
    mode: "development",
    devtool: "inline-source-map"
};

var exampleConfig = Object.assign({}, config, {
    name: "signing-agent-js-example",
    target: "node",
    entry: "./signingagent-client.js",
    output: {
       path: path.resolve(__dirname, 'dist'),
       filename: "signing-agent-js-example.js"
    },
});

module.exports = [
    exampleConfig
];
