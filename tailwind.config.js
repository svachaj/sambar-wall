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
        "open-modal": "openModal 1s",
        "input-error": "openModal 0.5s",
      },
      keyframes: {
        openModal: {
          "0%": { opacity: 0 },
          "100%": { opacity: 1 },
        },
      },
    },
  },
  plugins: [require("@tailwindcss/typography"), require("@tailwindcss/forms")],
};
