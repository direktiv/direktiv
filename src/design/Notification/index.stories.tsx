import { BanIcon, Check, LucideActivity, Settings } from "lucide-react";

import {
  Notification,
  NotificationLoading,
  NotificationMessage,
  NotificationTitle,
} from "./";
import { DropdownMenuSeparator } from "../Dropdown";

// eslint-disable-next-line sort-imports
import type { Meta, StoryObj } from "@storybook/react";

const meta = {
  title: "Components/Notification",
  component: Notification,
} satisfies Meta<typeof Notification>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Notification {...args}>content goes here...</Notification>
  ),
  argTypes: {
    showIndicator: {
      description: "Small red dot that indicates the existence of messages",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const NotificationIsLoading = () => (
  <Notification showIndicator={true}>
    <NotificationTitle>Notifications</NotificationTitle>
    <DropdownMenuSeparator />
    <NotificationLoading>Loading...</NotificationLoading>
  </Notification>
);

export const NotificationNoMessage = () => (
  <Notification showIndicator={false}>
    <NotificationTitle>Notifications</NotificationTitle>
    <DropdownMenuSeparator />
    <NotificationMessage icon={Check} text="Everything is fine." />
  </Notification>
);

export const NotificationHasMessage = () => (
  <Notification showIndicator={true}>
    <NotificationTitle>Notifications</NotificationTitle>
    <DropdownMenuSeparator className="last:hidden" />
    <NotificationMessage
      icon={Settings}
      text="Settings for the current workflow are incomplete."
    />
    <DropdownMenuSeparator className="last:hidden" />
    <NotificationMessage
      icon={BanIcon}
      text="An error occurred in one of your workflows."
    />
    <DropdownMenuSeparator className="last:hidden" />
    <NotificationMessage
      icon={LucideActivity}
      text="Please check the Monitoring Logs."
    />
    <DropdownMenuSeparator className="last:hidden" />
  </Notification>
);
