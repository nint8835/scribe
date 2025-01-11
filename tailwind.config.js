/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["pkg/web/ui/**/*.templ"],
  theme: {
    fontFamily: {
      sans: ["Inter Var", "sans-serif"],
    },
    extend: {},
  },
  plugins: [require("@tailwindcss/typography")],
};
