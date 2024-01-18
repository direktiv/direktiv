import type { Meta, StoryObj } from "@storybook/react";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";
import Button from "../Button";
import Input from "./index";

const meta = {
  title: "Components/Input",
  component: Input,
} satisfies Meta<typeof Input>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => <Input placeholder="default" {...args} />,
  tags: ["autodocs"],
};

export const WithButton = () => (
  <div className="flex space-x-3">
    <Input value="Text" />
    <Select>
      <SelectTrigger>
        <SelectValue placeholder="select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Button>Button</Button>
  </div>
);

export const Disabled = () => (
  <div className="flex space-x-3">
    <Input value="not disabled" />
    <Input value="disabled" disabled />
  </div>
);

export const FileInput = () => (
  <div className="flex space-x-3">
    <Input type="file" />
  </div>
);
