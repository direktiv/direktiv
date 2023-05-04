import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "../Dialog";
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

import Button from "../Button";
import { Settings } from "lucide-react";

const meta = {
  title: "Components/Select",
  component: SelectTrigger,
} satisfies Meta<typeof SelectTrigger>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: (args) => (
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
  ),
  args: {},
  tags: ["autodocs"],
  argTypes: {
    variant: {
      description: "Select variant",
      control: "select",
      options: ["destructive", "outline", "primary", "ghost", "link"],
      type: { name: "string", required: false },
    },
    size: {
      description: "select size",
      control: "select",
      options: [undefined, "sm", "lg"],
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
    asChild: {
      table: {
        disable: true,
      },
    },
  },
};

export const SelectSizes = () => (
  <div className="flex flex-wrap gap-5">
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

export const LoadingState = () => (
  <div className="flex flex-wrap gap-5">
    <Select>
      <SelectTrigger loading>
        <SelectValue placeholder="loading..." />
      </SelectTrigger>
    </Select>
  </div>
);

export const TriggerVariants = () => (
  <div className="flex flex-wrap space-x-3">
    <Select>
      <SelectTrigger>
        <SelectValue placeholder="default select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Select>
      <SelectTrigger variant="primary">
        <SelectValue placeholder="primary select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Select>
      <SelectTrigger variant="outline">
        <SelectValue placeholder="outline select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Select>
      <SelectTrigger variant="ghost">
        <SelectValue placeholder="ghost select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Select>
      <SelectTrigger variant="link">
        <SelectValue placeholder="link select" />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value="1">Item 1</SelectItem>
        <SelectItem value="2">Item 2</SelectItem>
        <SelectItem value="3">Item 3</SelectItem>
      </SelectContent>
    </Select>
    <Select>
      <SelectTrigger variant="destructive">
        <SelectValue placeholder="destructive select" />
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
    <SelectTrigger className="w-full" block>
      <SelectValue placeholder="block element" />
    </SelectTrigger>
    <SelectContent>
      <SelectItem value="1">Item 1</SelectItem>
      <SelectItem value="2">Item 2</SelectItem>
      <SelectItem value="3">Item 3</SelectItem>
    </SelectContent>
  </Select>
);

export const SelectWithinADialog = () => (
  <Dialog>
    <DialogTrigger asChild>
      <Button>Open Dialog</Button>
    </DialogTrigger>
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Settings />
          Dialog Title
        </DialogTitle>
        <DialogDescription>
          This Demo shows that the select also works withing a modal. There have
          been some z-index conflicts before..
        </DialogDescription>
      </DialogHeader>
      <div>
        <Select>
          <SelectTrigger block>
            <SelectValue placeholder="select" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">Item 1</SelectItem>
            <SelectItem value="2">Item 2</SelectItem>
            <SelectItem value="3">Item 3</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button>Submit</Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);
