/** @type {import('tailwindcss').Config} */
const colors = require("tailwindcss/colors");

module.exports = {
  darkMode: ["class"],
  content: [
    "../../views/**/*.{templ}",
    "./src/**/*.{js,jsx,ts,tsx}",
    "./style/**/*.css",
    "./app/**/*.{js,jsx,ts,tsx}",
    "./app/*.{js,jsx,ts,tsx}",
  ],
  theme: {
    extend: {
      borderColor: {
        border: "var(--border)",
      },
      colors: {
        border: "var(--border)",
        input: "var(--input)",
        ring: "var(--ring)",
        background: "var(--background)",
        foreground: "var(--foreground)",
        primary: {
          DEFAULT: "var(--primary)",
          foreground: "var(--primary-foreground)",
        },
        secondary: {
          DEFAULT: "var(--secondary)",
          foreground: "var(--secondary-foreground)",
        },
        destructive: {
          DEFAULT: "var(--destructive)",
          foreground: "var(--destructive-foreground)",
        },
        muted: {
          DEFAULT: "var(--muted)",
          foreground: "var(--muted-foreground)",
        },
        accent: {
          DEFAULT: "var(--accent)",
          foreground: "var(--accent-foreground)",
        },
        popover: {
          DEFAULT: "var(--popover)",
          foreground: "var(--popover-foreground)",
        },
        card: {
          DEFAULT: "var(--card)",
          foreground: "var(--card-foreground)",
        },
      },
    },
  },
  plugins: [
    require("tailwind-scrollbar"),
    require("tailwindcss-animate"),
    require("@tailwindcss/typography"),
  ],
};
