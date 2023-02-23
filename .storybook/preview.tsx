import { ThemeProvider } from "@mui/material/styles";
import withTailwindThemeDecorator from "./withTailwindTheme.decorator";
import React from "react";
import "../src/app.css";

import theme from "../src/theme/style";

export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
};

export const globalTypes = {
  theme: {
    name: "Theme",
    description: "Global theme for components",
    toolbar: {
      icon: "paintbrush",
      items: [
        { value: "light", title: "Light", left: "🌞" },
        { value: "dark", title: "Dark", left: "🌛" },
      ],
      dynamicTitle: true,
    },
  },
};

export const decorators = [
  withTailwindThemeDecorator,
  (Story) => (
    <ThemeProvider theme={theme}>
      <Story />
    </ThemeProvider>
  ),
];
