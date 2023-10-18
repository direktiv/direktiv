import {
  NotificationButton,
  NotificationLoading,
  NotificationText,
  NotificationTitle,
} from "./NotificationModal";
// eslint-disable-next-line sort-imports
import type { Meta, StoryObj } from "@storybook/react";
import { DropdownMenuSeparator } from "../Dropdown";
import Notification from ".";

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
  <div className="">
    <Notification showIndicator={true}>
      <NotificationLoading>Loading...</NotificationLoading>
    </Notification>
  </div>
);

export const NotificationHasMessage = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <NotificationText>You have 142 unread messages!</NotificationText>
    </Notification>
  </div>
);

export const NotificationNoMessage = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={false}>
      <NotificationText>Everything is fine.</NotificationText>
    </Notification>
  </div>
);

export const NotificationHasMessageComplexExample = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <div>
        <NotificationTitle>Error Message</NotificationTitle>
        <DropdownMenuSeparator></DropdownMenuSeparator>
        <NotificationText>Description of the issue...</NotificationText>
      </div>
      <div className="flex justify-end">insert a button here</div>
    </Notification>
  </div>
);

export const ButtonTestWithLinkto = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <div className="flex justify-end">
        <NotificationButton>Go fix it</NotificationButton>
      </div>
    </Notification>
  </div>
);

export const ButtonTestEmpty = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <div className="flex justify-end">
        <NotificationButton>Go fix it</NotificationButton>
      </div>
    </Notification>
  </div>
);
