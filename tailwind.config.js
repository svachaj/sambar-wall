/** @type {import('tailwindcss').Config} */
const colors = require("tailwindcss/colors");

module.exports = {
  darkMode: "class",
  content: ["./modules/**/*.templ"],
  theme: {
    extend: {
      colors: {
        primary: colors.pink,
        secondary: colors.ambra,
        accent: colors.red,
        neutral: colors.gray,
      },
      animation: {
        "show-smooth-1s": "showSmooth 1s",
        "show-smooth-1/2s": "showSmooth 0.5s",
      },
      keyframes: {
        showSmooth: {
          "0%": { opacity: 0 },
          "100%": { opacity: 1 },
        },
      },
      screens: {
        extraSmall: "360px",
      },
    },
  },
  plugins: [require("@tailwindcss/typography"), require("@tailwindcss/forms")],
};
