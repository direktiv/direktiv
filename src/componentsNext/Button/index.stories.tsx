import type { Meta, StoryObj } from "@storybook/react";
import {
  MoreHorizontal,
  MoreVertical,
  Search,
  ZoomIn,
  ZoomOut,
} from "lucide-react";
import { useEffect, useState } from "react";
import Button from "./index";

const meta = {
  title: "Components (next)/Button",
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
    size: {
      description: "Button size",
      control: "select",
      options: ["xs", "sm", "lg"],
      type: { name: "string", required: false },
    },
    color: {
      description: "Button color",
      control: "select",
      options: [
        "primary",
        "secondary",
        "accent",
        "ghost",
        "link",
        "info",
        "success",
        "warning",
        "error",
      ],
      type: { name: "string", required: false },
    },
    outline: {
      description: "button with outline",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    active: {
      description: "button in active state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    loading: {
      description: "Button in loading state",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    circle: {
      description: "round variation of a button",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
    block: {
      description: "make button full width",
      control: "boolean",
      type: { name: "boolean", required: false },
    },
  },
};

export const ButtonSizes = () => {
  return (
    <div className="flex flex-wrap gap-5">
      <Button size="xs">xs Button</Button>
      <Button size="sm">sm Button</Button>
      <Button>Button</Button>
      <Button size="lg">lg Button</Button>
    </div>
  );
};

export const ButtonColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button>Default</Button>
    <Button color="primary">Primary</Button>
    <Button color="secondary">Secondary</Button>
    <Button color="accent">Accent</Button>
    <Button color="ghost">Ghost</Button>
    <Button color="link">Link</Button>
    <Button color="info">Info</Button>
    <Button color="success">Success</Button>
    <Button color="warning">Warning</Button>
    <Button color="error">Error</Button>
  </div>
);

export const ActiveButtonColors = () => (
  <div className="flex flex-wrap gap-5">
    <Button active>Default</Button>
    <Button active color="primary">
      Primary
    </Button>
    <Button active color="secondary">
      Secondary
    </Button>
    <Button active color="accent">
      Accent
    </Button>
    <Button active color="ghost">
      Ghost
    </Button>
    <Button active color="link">
      Link
    </Button>
    <Button active color="info">
      Info
    </Button>
    <Button active color="success">
      Success
    </Button>
    <Button active color="warning">
      Warning
    </Button>
    <Button active color="error">
      Error
    </Button>
  </div>
);

export const Outline = () => (
  <div className="flex flex-wrap gap-5">
    <Button outline>Default</Button>
    <Button outline color="primary">
      Primary
    </Button>
    <Button outline color="secondary">
      Secondary
    </Button>
    <Button outline color="accent">
      Accent
    </Button>
    <Button outline color="ghost">
      Ghost
    </Button>
    <Button outline color="link">
      Link
    </Button>
    <Button outline color="info">
      Info
    </Button>
    <Button outline color="success">
      Success
    </Button>
    <Button outline color="warning">
      Warning
    </Button>
    <Button outline color="error">
      Error
    </Button>
  </div>
);

export const Group = () => (
  <>
    <div className="mb-5">
      <div className="btn-group ">
        <Button color="primary">
          <ZoomIn />
          Button 1
        </Button>
        <Button>
          <ZoomOut />
          Button 2
        </Button>
        <Button>
          <Search />
          Button 3
        </Button>
      </div>
    </div>
    <div>
      <div className="btn-group btn-group-vertical">
        <Button color="primary">
          <ZoomIn />
          Button 1
        </Button>
        <Button>
          <ZoomOut />
          Button 2
        </Button>
        <Button>
          <Search />
          Button 3
        </Button>
      </div>
    </div>
  </>
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
      <Button outline loading>
        Loading
      </Button>
      <Button
        className="w-[300px]"
        loading={isLoading}
        color="primary"
        active
        onClick={() => {
          setIsLoading((old) => !old);
        }}
      >
        {isLoading
          ? "I'm loading and deactivated..."
          : "click me do start loading"}
      </Button>
    </div>
  );
};

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

export const CircleButton = () => (
  <div className="flex flex-wrap gap-5">
    <Button size="lg" color="primary" active circle>
      <ZoomIn />
    </Button>
    <Button outline circle>
      <ZoomIn />
    </Button>
    <Button size="sm" color="accent" circle>
      <ZoomIn />
    </Button>
    <Button size="xs" outline color="secondary" circle>
      <ZoomIn />
    </Button>
  </div>
);

export const Block = () => <Button block>Block Element Button</Button>;
