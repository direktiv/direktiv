import type { Meta, StoryObj } from "@storybook/react";
import { RapiDoc } from "./index";
import exampleSpec from "./example.yaml";

const meta = {
  title: "Components/RapiDoc",
  component: RapiDoc,
  parameters: {
    layout: "fullscreen",
  },
} satisfies Meta<typeof RapiDoc>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  args: {
    spec: exampleSpec,
  },
};

export const DarkTheme: Story = {
  args: {
    spec: exampleSpec,
  },
  parameters: {
    themes: {
      default: "dark",
    },
  },
};

export const WithCustomStyles: Story = {
  args: {
    spec: exampleSpec,
    className: "custom-theme",
  },
};
