import "../src/App.css";
import "reactflow/dist/base.css";
import "../src/design/WorkflowDiagram/style.css";

import React from "react";
import withTailwindThemeDecorator from "./withTailwindTheme.decorator";

export const parameters = {
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

export const decorators = [withTailwindThemeDecorator, (Story) => <Story />];
export const tags = ["autodocs"];
