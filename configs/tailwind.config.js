/** @type {import('tailwindcss').Config} */
module.exports = {
	content: ["./internal/view/**/*.templ"],
	theme: {
		extend: {
			transitionProperty: {
				width: "width, transform",
			},
		},
	},
	plugins: [require("daisyui")],
	daisyui: {
		themes: ["nord"],
	},
};
