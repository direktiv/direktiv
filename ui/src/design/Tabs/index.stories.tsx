import { Code2, Layers, PieChart, Settings } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./index";
import { Card } from "../Card";

const meta = {
  title: "Components/Tabs",
  component: Tabs,
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => (
    <Tabs defaultValue="account" className="w-[400px]" {...args}>
      <TabsList>
        <TabsTrigger value="account">Account</TabsTrigger>
        <TabsTrigger value="password">Password</TabsTrigger>
      </TabsList>
      <TabsContent value="account">
        <p className="text-sm text-gray-8 dark:text-gray-dark-8">
          Make changes to your account here. Click save when you&apos;re done.
        </p>
      </TabsContent>
      <TabsContent value="password">
        <p className="text-sm text-gray-8 dark:text-gray-dark-8">
          Change your password here. After saving, you&apos;ll be logged out.
        </p>
      </TabsContent>
    </Tabs>
  ),
};

export const Boxed = () => (
  <Tabs defaultValue="account" className="w-[400px]">
    <TabsList variant="boxed">
      <TabsTrigger variant="boxed" value="account">
        Account
      </TabsTrigger>
      <TabsTrigger variant="boxed" value="password">
        Password
      </TabsTrigger>
    </TabsList>
    <TabsContent value="account" asChild>
      <Card className="p-4 text-sm text-gray-8 dark:text-gray-dark-8" noShadow>
        Make changes to your account here. Click save when you&apos;re done.
      </Card>
    </TabsContent>
    <TabsContent value="password" asChild>
      <Card className="p-4 text-sm text-gray-8 dark:text-gray-dark-8" noShadow>
        Change your password here. After saving, you&apos;ll be logged out.
      </Card>
    </TabsContent>
  </Tabs>
);

export const TabsWithIcons = () => (
  <div>
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="active-rev" asChild>
          <a href="#">
            <Code2 aria-hidden="true" />
            File Content
          </a>
        </TabsTrigger>
        <TabsTrigger value="overview" asChild>
          <a href="#">
            <PieChart aria-hidden="true" />
            Overview
          </a>
        </TabsTrigger>
        <TabsTrigger value="services" asChild>
          <a href="#">
            <Layers aria-hidden="true" />
            Services
          </a>
        </TabsTrigger>
        <TabsTrigger value="settings" asChild>
          <a href="#">
            <Settings aria-hidden="true" />
            Settings
          </a>
        </TabsTrigger>
      </TabsList>
    </Tabs>
    <div className="py-4">
      This example also shows that you can use Links for the tabs.
    </div>
  </div>
);
