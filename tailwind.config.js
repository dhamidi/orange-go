/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./pages/**/*.go"],
  theme: {
    extend: {},
  },
  plugins: [require("@tailwindcss/forms"), require("@tailwindcss/typography")],
};
