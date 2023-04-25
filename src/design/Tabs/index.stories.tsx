import { GitCommit, GitMerge, PieChart, Settings } from "lucide-react";
import type { Meta, StoryObj } from "@storybook/react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./index";
import clsx from "clsx";

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
);

export const TabsWithIcons = () => (
  <div>
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="overview" asChild>
          <a href="#">
            <PieChart aria-hidden="true" />
            Overview
          </a>
        </TabsTrigger>
        <TabsTrigger value="active-rev" asChild>
          <a href="#">
            <GitCommit aria-hidden="true" />
            Active Revisions
          </a>
        </TabsTrigger>
        <TabsTrigger value="revisions" asChild>
          <a href="#">
            <GitMerge aria-hidden="true" />
            Revisions
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

export const ContentNoBorder = () => (
  <Tabs defaultValue="account" className="w-[400px]">
    <TabsList>
      <TabsTrigger value="account">Account</TabsTrigger>
      <TabsTrigger value="password">Password</TabsTrigger>
      <TabsTrigger value="third">Third</TabsTrigger>
      <TabsTrigger value="fourth">Fourth</TabsTrigger>
    </TabsList>
    <TabsContent value="account" noBorder>
      <p className="text-sm text-gray-8 dark:text-gray-dark-8">
        Make changes to your account here. Click save when you&apos;re done.
      </p>
    </TabsContent>
    <TabsContent value="password" noBorder>
      <p className="text-sm text-gray-8 dark:text-gray-dark-8">
        Change your password here. After saving, you&apos;ll be logged out.
      </p>
    </TabsContent>
    <TabsContent value="third" noBorder>
      <p className="text-sm text-gray-8 dark:text-gray-dark-8">
        Your third content here
      </p>
    </TabsContent>
    <TabsContent value="fourth" noBorder>
      <p className="text-sm text-gray-8 dark:text-gray-dark-8">
        The fourth content comes here
      </p>
    </TabsContent>
  </Tabs>
);
