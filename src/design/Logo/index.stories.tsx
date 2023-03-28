import type { Meta, StoryObj } from "@storybook/react";
import Logo from "./index";

const meta = {
  title: "Components/Logo",
  component: Logo,
} satisfies Meta<typeof Logo>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ ...args }) => <Logo {...args} />,
  argTypes: {
    className: {
      table: {
        disable: true,
      },
    },
    iconOnly: {
      description: "icon version of the logo",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const WithOrWithoutLogo = () => (
  <div className="flex space-x-5">
    <Logo />
    <Logo iconOnly />
  </div>
);

export const PassClassnames = () => (
  <div className="flex space-x-5">
    <Logo className="w-20" />
    <Logo className="w-24" />
    <Logo />
  </div>
);
