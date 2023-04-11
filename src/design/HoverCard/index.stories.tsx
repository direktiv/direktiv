import { HoverCard, HoverCardContent, HoverCardTrigger } from "./index";

import type { Meta, StoryObj } from "@storybook/react";

const meta = {
  title: "Components/HoverCard",
  component: HoverCard,
} satisfies Meta<typeof HoverCard>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <HoverCard {...args}>
      <HoverCardTrigger>Hover</HoverCardTrigger>
      <HoverCardContent>
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  ),
};

export const ContentAlignment = () => (
  <div className=" flex w-full flex-row justify-around">
    <HoverCard>
      <HoverCardTrigger>Hover, Align Start</HoverCardTrigger>
      <HoverCardContent align="start">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover, Align Center</HoverCardTrigger>
      <HoverCardContent align="center">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover, Align End</HoverCardTrigger>
      <HoverCardContent align="end">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
);

export const ContentSide = () => (
  <div className=" flex h-64 w-full flex-row items-center justify-around">
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see right</HoverCardTrigger>
      <HoverCardContent side="right">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see left</HoverCardTrigger>
      <HoverCardContent side="left">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see top</HoverCardTrigger>
      <HoverCardContent side="top">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>Hover Me, to see bottom</HoverCardTrigger>
      <HoverCardContent side="bottom">
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
);

export const ContentSideOffset = () => (
  <div className="flex h-64 w-full flex-row items-center justify-around">
    <HoverCard>
      <HoverCardTrigger>
        Hover Me, to see 5px Margin of the Conntent
      </HoverCardTrigger>
      <HoverCardContent sideOffset={5}>
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
    <HoverCard>
      <HoverCardTrigger>
        Hover Me, to see 30px Margin of the Conntent
      </HoverCardTrigger>
      <HoverCardContent sideOffset={30}>
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
);

export const OpenCloseDelay = () => (
  <div className=" flex h-64 w-full flex-row items-center justify-around">
    <HoverCard openDelay={3000} closeDelay={3000}>
      <HoverCardTrigger>Hover Me, to see in 3 secs</HoverCardTrigger>
      <HoverCardContent>
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>

    <HoverCard openDelay={1000} closeDelay={1000}>
      <HoverCardTrigger>Hover Me, to see in 1 sec</HoverCardTrigger>
      <HoverCardContent>
        <div className="space-y-2">
          <h4 className="text-sm font-semibold">@nextjs</h4>
          <p className="text-sm">
            The React Framework – created and maintained by @vercel.
          </p>
        </div>
      </HoverCardContent>
    </HoverCard>
  </div>
);
