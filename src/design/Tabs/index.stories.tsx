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
  <div className="flex flex-col space-y-4">
    <Tabs defaultValue="overview">
      <TabsList>
        <TabsTrigger value="overview">
          <PieChart aria-hidden="true" />
          Overview
        </TabsTrigger>
        <TabsTrigger value="active-rev">
          <GitCommit aria-hidden="true" />
          Active Revisions
        </TabsTrigger>
        <TabsTrigger value="revisions">
          <GitMerge aria-hidden="true" />
          Revisions
        </TabsTrigger>
        <TabsTrigger value="settings">
          <Settings aria-hidden="true" />
          Settings
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
      <TabsContent value="third">
        <p className="text-sm text-gray-8 dark:text-gray-dark-8">
          Your third content here
        </p>
      </TabsContent>
      <TabsContent value="fourth">
        <p className="text-sm text-gray-8 dark:text-gray-dark-8">
          The fourth content comes here
        </p>
      </TabsContent>
    </Tabs>
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
