import type { Meta, StoryObj } from "@storybook/react";
import { Popover, PopoverContent, PopoverTrigger } from "./index";

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
      <PopoverContent>Place content for the popover here.</PopoverContent>
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
    <PopoverContent>Place content for the popover here.</PopoverContent>
  </Popover>
);

export const AlignPopoverContent = () => (
  <div>
    <Popover>
      <PopoverTrigger className="mx-2 w-48 bg-gray-5">
        Align Start
      </PopoverTrigger>
      <PopoverContent align="start">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
    <Popover>
      <PopoverTrigger className="mx-2 w-48 bg-gray-5">
        Align Center
      </PopoverTrigger>
      <PopoverContent align="center">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
    <Popover>
      <PopoverTrigger className="mx-2 w-48 bg-gray-5">Align End</PopoverTrigger>
      <PopoverContent align="end">
        Place content for the popover here.
      </PopoverContent>
    </Popover>
  </div>
);
