import type { Meta, StoryObj } from "@storybook/react";
import { Popover, PopoverClose, PopoverContent, PopoverTrigger } from "./index";
import Button from "../Button";

const meta = {
  title: "Components/Popover",
  component: Popover,
} satisfies Meta<typeof Popover>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <Popover>
      <PopoverTrigger>Open</PopoverTrigger>
      <PopoverContent className="p-4">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
  ),
  tags: ["autodocs"],
  argTypes: {
    open: {
      description: "Is popover open",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    defaultOpen: {
      description: "Is popover open by default",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    modal: {
      description: "Is popover a modal",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const DefaultOpen = () => (
  <Popover defaultOpen>
    <PopoverTrigger>Open</PopoverTrigger>
    <PopoverContent className="flex flex-col gap-3 p-4">
      <div>Place content for the popover here.</div>
      <PopoverClose asChild>
        <Button>Close this popover</Button>
      </PopoverClose>
    </PopoverContent>
  </Popover>
);

export const AlignPopoverContent = () => (
  <div className="flex space-x-4">
    <Popover>
      <PopoverTrigger asChild>
        <Button>Align Start</Button>
      </PopoverTrigger>
      <PopoverContent align="start" className="p-4">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
    <Popover>
      <PopoverTrigger asChild>
        <Button>Align Center</Button>
      </PopoverTrigger>
      <PopoverContent align="center" className="p-4">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
    <Popover>
      <PopoverTrigger>
        <Button>Align End</Button>
      </PopoverTrigger>
      <PopoverContent align="end" className="p-4">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
  </div>
);
