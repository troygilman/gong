/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./internal/**/*.templ"],
  theme: {
    extend: {
      transitionProperty: {
        width: "width, transform",
      },
    },
  },
  plugins: [require("daisyui")],
  daisyui: {
    themes: ["nord", "dark"],
  },
};
