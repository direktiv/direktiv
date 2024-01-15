import type { Meta, StoryObj } from "@storybook/react";
import {
  Toast,
  ToastAction,
  ToastVariantsType,
  Toaster,
  useToast,
} from "./index";
import Button from "../Button";
import { FC } from "react";

const meta = {
  title: "Components/Toast",
  component: Toast,
} satisfies Meta<typeof Toast>;

export default meta;
type Story = StoryObj<typeof meta>;

const StoryComponent: FC<Toast> = ({ ...args }) => {
  const { toast } = useToast();
  return (
    <>
      <Toaster />
      <Button
        onClick={() => {
          toast({
            title: "Scheduled: Catch up",
            description: "Friday, February 10, 2023 at 5:57 PM",
            variant: args.variant,
            action: <ToastAction altText="Try again">Try again</ToastAction>,
          });
        }}
      >
        Show Toast
      </Button>
    </>
  );
};

export const Default: Story = {
  render: ({ ...args }) => <StoryComponent {...args} />,
  tags: ["autodocs"],
  argTypes: {
    asChild: {
      table: {
        disable: true,
      },
    },
    variant: {
      description: "select variant",
      control: "select",
      options: ["info", "success", "warning", "error"],
      type: { name: "string", required: false },
    },
  },
};

export const ToastVariants = () => {
  const { toast } = useToast();

  const defContent = {
    title: "Scheduled: Catch up",
    description: "Friday, February 10, 2023 at 5:57 PM",
  };

  const customToast = (variant?: ToastVariantsType) => {
    toast({
      ...defContent,
      variant,
      action: <ToastAction altText="Try again">Try again</ToastAction>,
    });
  };

  return (
    <>
      <div className="flex flex-wrap gap-5">
        <Button
          onClick={() => {
            customToast("error");
          }}
        >
          Show Error Toast
        </Button>
        <Button
          onClick={() => {
            customToast("success");
          }}
        >
          Show Success Toast
        </Button>

        <Button
          onClick={() => {
            customToast("warning");
          }}
        >
          Show Warning Toast
        </Button>
        <Button
          onClick={() => {
            customToast("info");
          }}
        >
          Show Info Toast
        </Button>
        <Button
          onClick={() => {
            customToast();
          }}
        >
          Show Default Toast
        </Button>
      </div>
      <Toaster />
    </>
  );
};
