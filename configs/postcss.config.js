module.exports = {
  plugins: [
    require('tailwindcss'),
    require('autoprefixer'),
    require('postcss-short-classnames')({
      prefix: '_',
      minLength: 3,
      exclude: [
        'dark',
        'hidden',
        'active',  
      ],
      generate: 'base64',
    }),
  ]
}