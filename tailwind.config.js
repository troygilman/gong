/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./internal/view/**/*.templ"],
    plugins: [require("daisyui")],
    daisyui: {
        themes: ["nord"],
    },
};
