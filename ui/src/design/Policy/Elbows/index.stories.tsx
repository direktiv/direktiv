import { Meta, StoryObj } from "@storybook/react-vite";
import { Elbows } from ".";

const meta = {
  title: "Components/Policy/Elbows",
  component: Elbows,
} satisfies Meta<typeof Elbows>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <div className="flex min-h-[360px] flex-row items-center">
      <Elbows {...args} />
    </div>
  ),
  args: {
    rowHeight: 96,
    targets: [1, 1],
    width: 64,
    r: 12,
    reverse: false,
  },
  argTypes: {
    rowHeight: {
      control: { type: "number" },
      description: "Height of one row in the OR group",
    },
    targets: {
      control: { type: "object" },
      description: "array with the height (in rows) of each target",
    },
    width: {
      control: { type: "number" },
      description: "Width of the elbows element",
    },
    r: {
      control: { type: "number" },
      description: "Radius of the rounded edges",
    },
    reverse: {
      control: { type: "boolean" },
      description: "Flip L/R",
    },
  },
};

const ElbowsLeft = () => (
  <svg
    viewBox="0 0 64 192"
    className="h-[192px] w-[64px] fill-none stroke-gray-400 stroke-2"
  >
    <path
      d="M0,96
    H 20
    A 12 12 0 0 0 32 84
    V 60
    A 12 12 0 0 1 44 48
    H 64"
    />
    <path
      d="M0,96
     H 20
     A 12 12 0 0 1 32 108
     V 132
     A 12 12 0 0 0 44 144
     H 64"
    />
  </svg>
);

const ElbowsRight = () => (
  <svg
    viewBox="0 0 64 192"
    className="h-[192px] w-[64px] fill-none stroke-gray-400 stroke-2"
  >
    <path
      d="M64,96
         H 44
         A 12 12 0 0 1 32 84
         V 60
         A 12 12 0 0 0 20 48
         H 0"
    />
    <path
      d="M64,96
         H 44
         A 12 12 0 0 0 32 108
         V 132
         A 12 12 0 0 1 20 144
         H 0"
    />
  </svg>
);

export const HardCodedElbowsLeft = {
  render: () => (
    <div className="flex min-h-[360px] flex-row items-center">
      <ElbowsLeft />
    </div>
  ),
};

export const HardCodedElbowsRight = {
  render: () => (
    <div className="flex min-h-[360px] flex-row items-center">
      <ElbowsRight />
    </div>
  ),
};
