/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["pkg/web/ui/**/*.templ"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter Var", "sans-serif"],
      },
    },
  },
  plugins: [require("@tailwindcss/typography")],
};
