import {
  Box,
  Bug,
  Calendar,
  FolderTree,
  Layers,
  Network,
  Settings,
  Users,
} from "lucide-react";
import { Meta, StoryObj } from "@storybook/react";

import { Card } from "../Card";
import { NavigationLink } from "./index";

const meta = {
  title: "Components/NavigationLink",
  component: NavigationLink,
} satisfies Meta<typeof NavigationLink>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <NavigationLink {...args}>
      <FolderTree aria-hidden="true" />
      Some Menu Item
    </NavigationLink>
  ),
  args: {
    href: "https://direktiv.io",
  },
  argTypes: {
    children: {
      table: {
        disable: true,
      },
    },

    href: {
      description: "href link attribute",
      control: "text",
      type: { name: "string", required: true },
    },
    active: {
      description: "display as active",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

const navigation = [
  { name: "Explorer", href: "#", icon: FolderTree, current: true },
  { name: "Monitoring", href: "#", icon: Bug, current: false },
  { name: "Instances", href: "#", icon: Box, current: false },
  { name: "Events", href: "#", icon: Calendar, current: false },
  { name: "Gateway", href: "#", icon: Network, current: false },
  { name: "Permissions", href: "#", icon: Users, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

export const Navigation = () => (
  <Card className="m-5 w-44 p-3">
    {navigation.map((item) => (
      <NavigationLink key={item.name} href={item.href} active={item.current}>
        <item.icon aria-hidden="true" />
        {item.name}
      </NavigationLink>
    ))}
  </Card>
);
