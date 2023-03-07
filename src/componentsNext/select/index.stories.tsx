import type { Meta, StoryObj } from "@storybook/react";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
} from "./index";

const meta = {
  title: "Components (next)/Select",
  component: SelectTrigger,
} satisfies Meta<typeof SelectTrigger>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => {
    return (
      <Select>
        <SelectTrigger {...args}>
          <SelectValue placeholder="Select a fruit" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            <SelectLabel>Fruits</SelectLabel>
            <SelectItem value="apple">Apple</SelectItem>
            <SelectItem value="banana">Banana</SelectItem>
            <SelectItem value="blueberry">Blueberry</SelectItem>
            <SelectItem value="grapes">Grapes</SelectItem>
            <SelectItem value="pineapple">Pineapple</SelectItem>
          </SelectGroup>
          <SelectSeparator />
          <SelectGroup>
            <SelectLabel>Vegetables</SelectLabel>
            <SelectItem value="aubergine">Aubergine</SelectItem>
            <SelectItem value="broccoli">Broccoli</SelectItem>
            <SelectItem value="carrot" disabled>
              Carrot
            </SelectItem>
            <SelectItem value="courgette">Courgette</SelectItem>
            <SelectItem value="leek">Leek</SelectItem>
          </SelectGroup>
          <SelectSeparator />
          <SelectGroup>
            <SelectLabel>Meat</SelectLabel>
            <SelectItem value="beef">Beef</SelectItem>
            <SelectItem value="chicken">Chicken</SelectItem>
            <SelectItem value="lamb">Lamb</SelectItem>
            <SelectItem value="pork">Pork</SelectItem>
          </SelectGroup>
        </SelectContent>
      </Select>
    );
  },
  args: {},
  tags: ["autodocs"],
  argTypes: {
    size: {
      description: "select size",
      control: "select",
      options: ["xs", "sm", "lg"],
      type: { name: "string", required: false },
    },
    loading: {
      description: "select in loading state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    block: {
      description: "make select full width",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    ghost: {
      description: "ghost select",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    asChild: {
      table: {
        disable: true,
      },
    },
  },
};

export const SelectSizes = () => {
  return (
    <div className="flex flex-wrap gap-5">
      <Select>
        <SelectTrigger size="xs">
          <SelectValue placeholder="select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="1">Item 1</SelectItem>
          <SelectItem value="2">Item 2</SelectItem>
          <SelectItem value="3">Item 3</SelectItem>
        </SelectContent>
      </Select>
      <Select>
        <SelectTrigger size="sm">
          <SelectValue placeholder="select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="1">Item 1</SelectItem>
          <SelectItem value="2">Item 2</SelectItem>
          <SelectItem value="3">Item 3</SelectItem>
        </SelectContent>
      </Select>
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
      <Select>
        <SelectTrigger size="lg">
          <SelectValue placeholder="select" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="1">Item 1</SelectItem>
          <SelectItem value="2">Item 2</SelectItem>
          <SelectItem value="3">Item 3</SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
};

export const LoadingState = () => {
  return (
    <div className="flex flex-wrap gap-5">
      <Select>
        <SelectTrigger loading>
          <SelectValue placeholder="loading..." />
        </SelectTrigger>
      </Select>
    </div>
  );
};

export const Ghost = () => (
  <div className="bg-base-300 p-10">
    <Select>
      <SelectTrigger ghost>
        <SelectValue placeholder="ghost select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
  </div>
);

export const Block = () => (
  <Select>
    <SelectTrigger block>
      <SelectValue placeholder="block element" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="1">Item 1</SelectItem>
      <SelectItem value="2">Item 2</SelectItem>
      <SelectItem value="3">Item 3</SelectItem>
    </SelectContent>
  </Select>
);
