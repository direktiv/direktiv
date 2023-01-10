import { addDecorator } from '@storybook/react';
import { ThemeProvider } from "@mui/material/styles";
import theme from "../src/theme/style"


addDecorator((story) => (
  <ThemeProvider theme={theme}>{story()}</ThemeProvider>
));

export const parameters = {
  actions: { argTypesRegex: "^on[A-Z].*" },
  controls: {
    matchers: {
      color: /(background|color)$/i,
      date: /Date$/,
    },
  },
}