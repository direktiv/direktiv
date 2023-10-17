import {
  NotificationHasresultsButton,
  NotificationHasresultsText,
  NotificationHasresultsTitle,
  NotificationLoading,
  NotificationNoresults,
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
  render: ({ ...args }) => <Notification {...args} />,
  argTypes: {
    showIndicator: {
      description: "Small red dot that indicates the existence of messages",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const NotificationIsLoading = () => (
  <div className="flex space-x-2">
    <Notification isLoading={true} showIndicator={true}>
      <NotificationLoading>Loading...</NotificationLoading>
    </Notification>
  </div>
);

export const NotificationHasMessage = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <NotificationHasresultsText>
        You have 142 unread messages!
      </NotificationHasresultsText>
    </Notification>
  </div>
);

export const NotificationNoMessage = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={false}>
      <NotificationNoresults>Everything is fine.</NotificationNoresults>
    </Notification>
  </div>
);

export const NotificationHasMessageComplexExample = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <div className="">
        <NotificationHasresultsTitle>Error Message</NotificationHasresultsTitle>
        <DropdownMenuSeparator className=""></DropdownMenuSeparator>
        <NotificationHasresultsText>
          Description of the issue...
        </NotificationHasresultsText>
      </div>
      <div className="flex justify-end">
        <NotificationHasresultsButton className="border border-gray-9">
          Go fix it
        </NotificationHasresultsButton>
      </div>
    </Notification>
  </div>
);
