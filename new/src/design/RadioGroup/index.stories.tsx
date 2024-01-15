import type { Meta, StoryObj } from "@storybook/react";
import { RadioGroup, RadioGroupItem } from "./index";

const meta = {
  title: "Components/RadioGroup",
  component: RadioGroup,
} satisfies Meta<typeof RadioGroup>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <RadioGroup defaultValue="option-one" {...args}>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="option-one" id="option-one" />
        <label htmlFor="option-one">Option One</label>
      </div>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="option-two" id="option-two" />
        <label htmlFor="option-two">Option Two</label>
      </div>
      <div className="flex items-center space-x-2">
        <RadioGroupItem value="option-three" id="option-three" />
        <label htmlFor="option-three">Option Three</label>
      </div>
    </RadioGroup>
  ),
};

export const DisabledGroup = () => (
  <RadioGroup defaultValue="option-one" disabled>
    <div className="flex items-center space-x-2">
      <RadioGroupItem value="option-one" id="option-one" />
      <label htmlFor="option-one">Option One</label>
    </div>
    <div className="flex items-center space-x-2">
      <RadioGroupItem value="option-two" id="option-two" />
      <label htmlFor="option-two">Option Two</label>
    </div>
    <div className="flex items-center space-x-2">
      <RadioGroupItem value="option-three" id="option-three" />
      <label htmlFor="option-three">Option Three</label>
    </div>
  </RadioGroup>
);
