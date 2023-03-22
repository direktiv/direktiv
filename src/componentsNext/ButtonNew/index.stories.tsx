import type { Meta, StoryObj } from "@storybook/react";
import { MoreHorizontal, MoreVertical, ZoomIn } from "lucide-react";
import { useEffect, useState } from "react";
import Button from "./index";

const meta = {
  title: "Components (next)/Button New",
  component: Button,
} satisfies Meta<typeof Button>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: ({ children, ...args }) => <Button {...args}>{children}</Button>,
  args: {
    children: "Button",
  },
  argTypes: {
    children: {
      description: "Button text",
      control: {
        type: "text",
        defaultValue: "Button",
      },
      type: { name: "string", required: false },
    },
    variant: {
      description: "Button variant",
      control: "select",
      options: ["destructive", "outline", "primary", "ghost", "link"],
      type: { name: "string", required: false },
    },
    size: {
      description: "Button size",
      control: "select",
      options: ["xs", "sm", "lg"],
      type: { name: "string", required: false },
    },
    loading: {
      description: "button in loading state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    circle: {
      description: "circle button",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const ButtonVariants = () => (
  <div className="flex flex-wrap gap-5">
    <Button>Default Button</Button>
    <Button variant="primary">Primary Button</Button>
    <Button variant="outline">Outline Button</Button>
    <Button variant="ghost">Ghost Button</Button>
    <Button variant="link">Link Button</Button>
    <Button variant="destructive">Destructive Button</Button>
  </div>
);

export const ButtonSizes = () => (
  <div className="flex flex-wrap gap-5">
    <Button size="xs">Extra Small Button</Button>
    <Button size="sm">Small Button</Button>
    <Button>Default Button</Button>
    <Button size="lg">Large Button</Button>
  </div>
);

export const WithIcon = () => (
  <div className="space-y-5">
    <div className="flex gap-5">
      <Button size="xs">
        <ZoomIn /> xs Button
      </Button>
      <Button size="sm">
        <ZoomIn /> sm Button
      </Button>
      <Button>
        <ZoomIn /> Button
      </Button>
      <Button size="lg">
        <ZoomIn /> lg Button
      </Button>
    </div>
    <div className="flex gap-5">
      <Button size="xs">
        <MoreHorizontal /> xs Button
      </Button>
      <Button size="sm">
        <MoreHorizontal /> sm Button
      </Button>
      <Button>
        <MoreHorizontal /> Button
      </Button>
      <Button size="lg">
        <MoreHorizontal /> lg Button
      </Button>
    </div>
    <div className="flex gap-5">
      <Button size="xs">
        <MoreVertical /> xs Button
      </Button>
      <Button size="sm">
        <MoreVertical /> sm Button
      </Button>
      <Button>
        <MoreVertical /> Button
      </Button>
      <Button size="lg">
        <MoreVertical /> lg Button
      </Button>
    </div>
  </div>
);

export const Loading = () => {
  const [isLoading, setIsLoading] = useState(false);
  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (isLoading) {
      timeout = setTimeout((): void => {
        setIsLoading(false);
      }, 2000);
    }

    return () => {
      clearTimeout(timeout);
    };
  }, [isLoading]);

  return (
    <div className="flex space-x-2">
      <Button loading>Loading</Button>
      <Button
        className="w-[300px]"
        loading={isLoading}
        onClick={() => {
          setIsLoading((old) => !old);
        }}
      >
        {isLoading
          ? "I'm loading and deactivated..."
          : "click me to start loading"}
      </Button>
    </div>
  );
};

export const CircleButton = () => (
  <div className="flex flex-wrap gap-5">
    <Button size="xs" circle>
      <ZoomIn />
    </Button>
    <Button size="sm" circle>
      <ZoomIn />
    </Button>
    <Button circle>
      <ZoomIn />
    </Button>
    <Button size="lg" circle>
      <ZoomIn />
    </Button>
  </div>
);
