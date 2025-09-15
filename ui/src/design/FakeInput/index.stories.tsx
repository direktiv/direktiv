import type { Meta, StoryObj } from "@storybook/react";
import { FakeInput } from ".";
import Input from "../Input";

const meta = {
  title: "Components/FakeInput",
  component: FakeInput,
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <FakeInput {...args}>Not an input</FakeInput>,
  tags: ["autodocs"],
};

export const Comparison = () => (
  <div className="flex gap-3">
    <FakeInput className="w-64">
      This <i>looks</i> like an input
    </FakeInput>
    <Input className="w-64" defaultValue="This is an input" />
  </div>
);

export const Truncate = () => (
  <div className="flex gap-3">
    <FakeInput className="w-64">
      Long text will just be truncated like this, just watch
    </FakeInput>
  </div>
);

export const Wrap = () => (
  <div className="flex gap-3">
    <FakeInput wrap className="w-64">
      With the `wrap` prop, a long text in this fake input will be wrapped.
    </FakeInput>
  </div>
);
