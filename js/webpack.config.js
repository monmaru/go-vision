var webpack = require('webpack');

module.exports = {
  entry: './index.js',
  output: {
    path: '../',
    filename: 'bundle.js',
  },
  devtool: "#source-map",
  plugins: [
    new webpack.optimize.UglifyJsPlugin({
      compress: {
        warnings: false,
      },
      output: {
        comments: false,
      },
    }),
  ]
}