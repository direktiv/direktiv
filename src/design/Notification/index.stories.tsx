import {
  NotificationLoading,
  NotificationMessage,
  NotificationText,
} from "./NotificationModal";
import { Boxes } from "lucide-react";

// eslint-disable-next-line sort-imports
import type { Meta, StoryObj } from "@storybook/react";
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
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <NotificationLoading>Loading...</NotificationLoading>
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

export const NotificationHasMessage = () => (
  <div className="flex space-x-2">
    <Notification showIndicator={true}>
      <div>
        <NotificationMessage
          icon={Boxes}
          title="Critical Issue"
          text="An error occurred in one of your workflows."
        ></NotificationMessage>
      </div>
    </Notification>
  </div>
);
