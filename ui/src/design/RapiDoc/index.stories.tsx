import type { Meta, StoryObj } from "@storybook/react";
import OpenApiSpec from "./example.json";
import { RapiDoc } from "./index";

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
    spec: OpenApiSpec,
    className: "h-[80vh]",
  },
};
