import type { Meta, StoryObj } from "@storybook/react";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./index";

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

export const TabsPrimary = () => (
  <Tabs defaultValue="account" className="w-[400px]">
    <TabsList varient="primary">
      <TabsTrigger varient="primary" value="account">
        Account
      </TabsTrigger>
      <TabsTrigger varient="primary" value="password">
        Password
      </TabsTrigger>
      <TabsTrigger varient="primary" value="third">
        Third
      </TabsTrigger>
      <TabsTrigger varient="primary" value="fourth">
        Fourth
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
);

export const ContentNoBorder = () => (
  <Tabs defaultValue="account" className="w-[400px]">
    <TabsList varient="primary">
      <TabsTrigger varient="primary" value="account">
        Account
      </TabsTrigger>
      <TabsTrigger varient="primary" value="password">
        Password
      </TabsTrigger>
      <TabsTrigger varient="primary" value="third">
        Third
      </TabsTrigger>
      <TabsTrigger varient="primary" value="fourth">
        Fourth
      </TabsTrigger>
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
