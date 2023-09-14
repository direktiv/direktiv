const {
  gray,
  grayDark,
  green,
  greenDark,
  yellow,
  yellowDark,
  blue,
  blueDark,
  red,
  redDark,
  blackA,
  whiteA,
} = require("@radix-ui/colors");
const defaultTheme = require("tailwindcss/defaultTheme");

module.exports = {
  content: [
    "./src/**/*.{js,ts,jsx,tsx}",
    "./node_modules/@tremor/**/*.{js,ts,jsx,tsx}",
  ],
  plugins: [require("tailwindcss-animate"), require("@headlessui/tailwindcss")],
  darkMode: ["class", '[data-theme="dark"]'],
  theme: {
    transparent: "transparent",
    current: "currentColor",
    fontFamily: {
      sans: ["Inter", ...defaultTheme.fontFamily.sans],
      mono: ["RobotoMono", ...defaultTheme.fontFamily.mono],
      menlo: 'Menlo, Monaco, "Courier New", monospace',
    },
    extend: {
      colors: {
        primary: {
          50: "#EEEFFF",
          100: "#CBD1FF",
          200: "#A9B2FF",
          300: "#8793FF",
          400: "#6473FF",
          500: "#5364FF",
          600: "#4554DF",
          700: "#323C99",
          800: "#212866",
          900: "#4D5469",
        },
        "black-alpha": {
          1: blackA.blackA1,
          2: blackA.blackA2,
          3: blackA.blackA3,
          4: blackA.blackA4,
          5: blackA.blackA5,
          6: blackA.blackA6,
          7: blackA.blackA7,
          8: blackA.blackA8,
          9: blackA.blackA9,
          10: blackA.blackA10,
          11: blackA.blackA11,
          12: blackA.blackA12,
        },
        "white-alpha": {
          1: whiteA.whiteA1,
          2: whiteA.whiteA2,
          3: whiteA.whiteA3,
          4: whiteA.whiteA4,
          5: whiteA.whiteA5,
          6: whiteA.whiteA6,
          7: whiteA.whiteA7,
          8: whiteA.whiteA8,
          9: whiteA.whiteA9,
          10: whiteA.whiteA10,
          11: whiteA.whiteA11,
          12: whiteA.whiteA12,
        },
        gray: {
          1: gray.gray1,
          2: gray.gray2,
          3: gray.gray3,
          4: gray.gray4,
          5: gray.gray5,
          6: gray.gray6,
          7: gray.gray7,
          8: gray.gray8,
          9: gray.gray9,
          10: gray.gray10,
          11: gray.gray11,
          12: gray.gray12,
        },
        "gray-dark": {
          1: grayDark.gray1,
          2: grayDark.gray2,
          3: grayDark.gray3,
          4: grayDark.gray4,
          5: grayDark.gray5,
          6: grayDark.gray6,
          7: grayDark.gray7,
          8: grayDark.gray8,
          9: grayDark.gray9,
          10: grayDark.gray10,
          11: grayDark.gray11,
          12: grayDark.gray12,
        },
        danger: {
          1: red.red1,
          2: red.red2,
          3: red.red3,
          4: red.red4,
          5: red.red5,
          6: red.red6,
          7: red.red7,
          8: red.red8,
          9: red.red9,
          10: red.red10,
          11: red.red11,
          12: red.red12,
        },
        "danger-dark": {
          1: redDark.red1,
          2: redDark.red2,
          3: redDark.red3,
          4: redDark.red4,
          5: redDark.red5,
          6: redDark.red6,
          7: redDark.red7,
          8: redDark.red8,
          9: redDark.red9,
          10: redDark.red10,
          11: redDark.red11,
          12: redDark.red12,
        },
        info: {
          1: blue.blue1,
          2: blue.blue2,
          3: blue.blue3,
          4: blue.blue4,
          5: blue.blue5,
          6: blue.blue6,
          7: blue.blue7,
          8: blue.blue8,
          9: blue.blue9,
          10: blue.blue10,
          11: blue.blue11,
          12: blue.blue12,
        },
        "info-dark": {
          1: blueDark.blue1,
          2: blueDark.blue2,
          3: blueDark.blue3,
          4: blueDark.blue4,
          5: blueDark.blue5,
          6: blueDark.blue6,
          7: blueDark.blue7,
          8: blueDark.blue8,
          9: blueDark.blue9,
          10: blueDark.blue10,
          11: blueDark.blue11,
          12: blueDark.blue12,
        },
        warning: {
          1: yellow.yellow1,
          2: yellow.yellow2,
          3: yellow.yellow3,
          4: yellow.yellow4,
          5: yellow.yellow5,
          6: yellow.yellow6,
          7: yellow.yellow7,
          8: yellow.yellow8,
          9: yellow.yellow9,
          10: yellow.yellow10,
          11: yellow.yellow11,
          12: yellow.yellow12,
        },
        "warning-dark": {
          1: yellowDark.yellow1,
          2: yellowDark.yellow2,
          3: yellowDark.yellow3,
          4: yellowDark.yellow4,
          5: yellowDark.yellow5,
          6: yellowDark.yellow6,
          7: yellowDark.yellow7,
          8: yellowDark.yellow8,
          9: yellowDark.yellow9,
          10: yellowDark.yellow10,
          11: yellowDark.yellow11,
          12: yellowDark.yellow12,
        },
        success: {
          1: green.green1,
          2: green.green2,
          3: green.green3,
          4: green.green4,
          5: green.green5,
          6: green.green6,
          7: green.green7,
          8: green.green8,
          9: green.green9,
          10: green.green10,
          11: green.green11,
          12: green.green12,
        },
        "success-dark": {
          1: greenDark.green1,
          2: greenDark.green2,
          3: greenDark.green3,
          4: greenDark.green4,
          5: greenDark.green5,
          6: greenDark.green6,
          7: greenDark.green7,
          8: greenDark.green8,
          9: greenDark.green9,
          10: greenDark.green10,
          11: greenDark.green11,
          12: greenDark.green12,
        },
        tremor: {
          brand: {
            faint: "#eff6ff", // blue-50
            muted: "#bfdbfe", // blue-200
            subtle: "#60a5fa", // blue-400
            DEFAULT: "#3b82f6", // blue-500
            emphasis: "#1d4ed8", // blue-700
            inverted: "#ffffff", // white
          },
          background: {
            muted: "#f9fafb", // gray-50
            subtle: "#f3f4f6", // gray-100
            DEFAULT: "#ffffff", // white
            emphasis: "#374151", // gray-700
          },
          border: {
            DEFAULT: "#e5e7eb", // gray-200
          },
          ring: {
            DEFAULT: "#e5e7eb", // gray-200
          },
          content: {
            subtle: "#9ca3af", // gray-400
            DEFAULT: "#6b7280", // gray-500
            emphasis: "#374151", // gray-700
            strong: "#111827", // gray-900
            inverted: "#ffffff", // white
          },
        },
        "dark-tremor": {
          brand: {
            faint: "#0B1229", // custom
            muted: "#172554", // blue-950
            subtle: "#1e40af", // blue-800
            DEFAULT: "#3b82f6", // blue-500
            emphasis: "#60a5fa", // blue-400
            inverted: "#030712", // gray-950
          },
          background: {
            muted: "#131A2B", // custom
            subtle: "#1f2937", // gray-800
            DEFAULT: "#111827", // gray-900
            emphasis: "#d1d5db", // gray-300
          },
          border: {
            DEFAULT: "#1f2937", // gray-800
          },
          ring: {
            DEFAULT: "#1f2937", // gray-800
          },
          content: {
            subtle: "#4b5563", // gray-600
            DEFAULT: "#6b7280", // gray-600
            emphasis: "#e5e7eb", // gray-200
            strong: "#f9fafb", // gray-50
            inverted: "#000000", // black
          },
        },
      },
      boxShadow: {
        // light
        "tremor-input": "0 1px 2px 0 rgb(0 0 0 / 0.05)",
        "tremor-card":
          "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
        "tremor-dropdown":
          "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
        // dark
        "dark-tremor-input": "0 1px 2px 0 rgb(0 0 0 / 0.05)",
        "dark-tremor-card":
          "0 1px 3px 0 rgb(0 0 0 / 0.1), 0 1px 2px -1px rgb(0 0 0 / 0.1)",
        "dark-tremor-dropdown":
          "0 4px 6px -1px rgb(0 0 0 / 0.1), 0 2px 4px -2px rgb(0 0 0 / 0.1)",
      },
      borderRadius: {
        "tremor-small": "0.375rem",
        "tremor-default": "0.5rem",
        "tremor-full": "9999px",
      },
      fontSize: {
        "tremor-label": ["0.75rem"],
        "tremor-default": ["0.875rem", { lineHeight: "1.25rem" }],
        "tremor-title": ["1.125rem", { lineHeight: "1.75rem" }],
        "tremor-metric": ["1.875rem", { lineHeight: "2.25rem" }],
      },
      keyframes: {
        overlayShow: {
          from: { opacity: 0 },
          to: { opacity: 1 },
        },
        contentShow: {
          from: { opacity: 0, transform: "translate(-50%, -48%) scale(0.96)" },
          to: { opacity: 1, transform: "translate(-50%, -50%) scale(1)" },
        },
      },
      animation: {
        overlayShow: "overlayShow 150ms cubic-bezier(0.16, 1, 0.3, 1)",
        contentShow: "contentShow 150ms cubic-bezier(0.16, 1, 0.3, 1)",
      },
    },
  },
  safelist: [
    {
      pattern:
        /^(bg-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
      variants: ["hover", "ui-selected"],
    },
    {
      pattern:
        /^(text-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
      variants: ["hover", "ui-selected"],
    },
    {
      pattern:
        /^(border-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
      variants: ["hover", "ui-selected"],
    },
    {
      pattern:
        /^(ring-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
    },
    {
      pattern:
        /^(stroke-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
    },
    {
      pattern:
        /^(fill-(?:slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(?:50|100|200|300|400|500|600|700|800|900|950))$/,
    },
  ],
};
