/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['pkg/web/ui/**/*.templ'],
  theme: {
    extend: {},
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}

