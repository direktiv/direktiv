import type { Meta, StoryObj } from "@storybook/react";
import { Slider } from "./index";
import { useState } from "react";

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

export const BigStep = () => {
  const [value, setValue] = useState(5);
  return (
    <div className="flex w-40 flex-col items-center space-y-3">
      <Slider
        step={5}
        min={0}
        max={20}
        value={[value]}
        onValueChange={(e) => {
          setValue(e?.[0] || 0);
        }}
      />
      <div>Value: {value}</div>
    </div>
  );
};

export const Disabled = () => (
  <div className="flex w-40 flex-col items-center space-y-3">
    <Slider disabled />
  </div>
);

export const Inverted = () => (
  <div className="w-40">
    <Slider inverted />
  </div>
);
