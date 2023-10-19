import type { Meta, StoryObj } from "@storybook/react";

import { Checkbox } from "./index";

const meta = {
  title: "Components/Checkbox",
  component: Checkbox,
} satisfies Meta<typeof Checkbox>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <div className="flex space-x-2">
      <Checkbox id="terms1" {...args} />
    </div>
  ),
  argTypes: {
    size: {
      description: "select size",
      control: "select",
      options: [undefined, "sm", "lg"],
      type: { name: "string", required: false },
    },
    disabled: {
      description: "enable/disable the checkbox",
      type: { name: "boolean", required: false },
    },
  },
};
export function CheckboxWithText() {
  return (
    <div className="flex space-x-2 p-2 ">
      <Checkbox id="terms2" />
      <div className="grid gap-1.5 leading-none">
        <label
          htmlFor="terms2"
          className="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70"
        >
          Accept terms and conditions
        </label>
        <p className="text-sm">
          You agree to our Terms of Service and Privacy Policy.
        </p>
      </div>
    </div>
  );
}

export function DisabledCheckbox() {
  return (
    <div className="flex space-x-2 p-2 ">
      <Checkbox id="terms-disabled-1" disabled />
    </div>
  );
}

export function CheckboxSizes() {
  return (
    <div className="flex space-x-2 p-2 ">
      <Checkbox size="lg" />
      <Checkbox />
      <Checkbox size="sm" />
    </div>
  );
}
