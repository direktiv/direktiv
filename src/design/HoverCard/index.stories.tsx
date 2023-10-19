import { HoverCard, HoverCardContent, HoverCardTrigger } from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import Alert from "../Alert";

const meta = {
  title: "Components/HoverCard",
  component: HoverCard,
} satisfies Meta<typeof HoverCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <HoverCard {...args}>
      <HoverCardTrigger>Hover Me</HoverCardTrigger>
      <HoverCardContent>Content goes here</HoverCardContent>
    </HoverCard>
  ),
};

export const ContentAlignment = () => (
  <div className=" flex w-full flex-row justify-around">
    <HoverCard>
      <HoverCardTrigger>Hover, Align Start</HoverCardTrigger>
      <HoverCardContent align="start">Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover, Align Center</HoverCardTrigger>
      <HoverCardContent align="center">Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover, Align End</HoverCardTrigger>
      <HoverCardContent align="end">Content goes here</HoverCardContent>
    </HoverCard>
  </div>
);

export const ContentSide = () => (
  <div className=" flex h-64 w-full flex-row items-center justify-around">
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see right</HoverCardTrigger>
      <HoverCardContent side="right">Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see left</HoverCardTrigger>
      <HoverCardContent side="left">Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see top</HoverCardTrigger>
      <HoverCardContent side="top">Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see bottom</HoverCardTrigger>
      <HoverCardContent side="bottom">Content goes here</HoverCardContent>
    </HoverCard>
  </div>
);

export const ContentSideOffset = () => (
  <div className="flex h-64 w-full flex-row items-center justify-around">
    <HoverCard>
      <HoverCardTrigger>
        Hover Me, to see 5px Margin of the Conntent
      </HoverCardTrigger>
      <HoverCardContent sideOffset={5}>Content goes here</HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>
        Hover Me, to see 30px Margin of the Conntent
      </HoverCardTrigger>
      <HoverCardContent sideOffset={30}>Content goes here</HoverCardContent>
    </HoverCard>
  </div>
);

export const OpenCloseDelay = () => (
  <div className=" flex h-64 w-full flex-row items-center justify-around">
    <HoverCard openDelay={3000} closeDelay={3000}>
      <HoverCardTrigger>Hover Me, to see in 3 secs</HoverCardTrigger>
      <HoverCardContent>Content goes here</HoverCardContent>
    </HoverCard>

    <HoverCard openDelay={1000} closeDelay={1000}>
      <HoverCardTrigger>Hover Me, to see in 1 sec</HoverCardTrigger>
      <HoverCardContent noBackground>Content goes here</HoverCardContent>
    </HoverCard>
  </div>
);

export const WithAlert = () => (
  <div className="flex h-64 w-full flex-row items-center justify-around">
    <HoverCard>
      <HoverCardTrigger>with alert</HoverCardTrigger>
      <HoverCardContent asChild noBackground>
        <Alert variant="error">Some error</Alert>
      </HoverCardContent>
    </HoverCard>
  </div>
);
