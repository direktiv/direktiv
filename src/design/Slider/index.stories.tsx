import type { Meta, StoryObj } from "@storybook/react";
import { Slider } from "./index";

const meta = {
  title: "Components/Slider",
  component: Slider,
} satisfies Meta<typeof Slider>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Slider {...args} />,
  argTypes: {
    defaultValue: {
      description: "Default Value of the Slider",
      control: "number",
      type: { name: "number", required: false },
    },
    max: {
      description: "Max Value of the Slider",
      control: "number",
      type: { name: "number", required: false },
    },
    step: {
      description: "Step Value of the Slider",
      control: "number",
      type: { name: "number", required: false },
    },
    disabled: {
      description: "Disable/Enable the Slider",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    inverted: {
      description: "Invert the Slider",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const BigStep = () => <Slider step={5} />;

export const Disabled = () => <Slider disabled />;

export const Inverted = () => <Slider inverted />;
