import type { Meta, StoryObj } from "@storybook/react";
import OpenApiSpec from "./example.json";
import { RapiDoc } from "./index";

interface RapiDocProps {
  spec: typeof OpenApiSpec;
  className?: string;
}

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
    spec: OpenApiSpec as RapiDocProps["spec"],
    className: "h-[80vh]",
  },
};
