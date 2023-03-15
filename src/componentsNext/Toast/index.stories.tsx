import type { Meta, StoryObj } from "@storybook/react";
import { Toast, ToastVariantsType, Toaster, useToast } from "./index";
import Button from "../../components/button";
import { FC } from "react";

const meta = {
  title: "Components (next)/Toast",
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
    });
  };
  return (
    <div className="flex flex-wrap gap-5">
      <Button
        color="error"
        onClick={() => {
          customToast("error");
        }}
      >
        Show Error Toast
      </Button>
      <Button
        color="success"
        onClick={() => {
          customToast("success");
        }}
      >
        Show Info Toast
      </Button>

      <Button
        color="warning"
        onClick={() => {
          customToast("warning");
        }}
      >
        Show Warning Toast
      </Button>
      <Button
        color="info"
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
  );
};
