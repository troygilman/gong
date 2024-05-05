/** @type {import('tailwindcss').Config} */
module.exports = {
    content: ["./view/**/*.templ"],
    plugins: [require("daisyui")],
    daisyui: {
        themes: ["dracula", "light"],
    },
};
