### Web Server

Webpack is used to resolve CommonJS imports at compile-time.

#### Instructions

- [Install latest version of npm](https://docs.npmjs.com/cli/v7/commands/npm-install)
- `npm install` dependencies in *package.json*

#### Requirements

- [Protobuf compiler](https://developers.google.com/protocol-buffers/docs/reference/javascript-generated)
- [Webpack and webpack-cli](https://webpack.js.org/guides/getting-started/)

#### Testing commonjs imports for protobuf

- `npx webpack`
- open the `dist/index.html` file in browser and look at console to test imports
