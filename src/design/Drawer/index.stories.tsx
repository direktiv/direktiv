import { Drawer, DrawerContent, DrawerMain, DrawerTrigger } from "./index";
import type { Meta, StoryObj } from "@storybook/react";
import Button from "../Button";
import Logo from "../Logo";

const meta = {
  title: "Components/Drawer",
  component: Drawer,
} satisfies Meta<typeof Drawer>;

export default meta;
type Story = StoryObj<typeof meta>;

const StoryCompontnt = () => (
  <Drawer>
    <DrawerMain>
      <div className="flex flex-col items-start space-y-5 p-10">
        <DrawerTrigger>
          <Button> Open Drawer</Button>
        </DrawerTrigger>
        <div>
          This is the <code>DrawerContent</code> component. Make sure to place
          it as a direct child of the <code>DrawerRoot</code> component and
          place the <code>DrawerMenu</code> directly after the{" "}
          <code>DrawerContent</code>.
        </div>
      </div>
    </DrawerMain>
    <DrawerContent>
      <div className="drawer-side flex w-full items-center justify-center">
        <Logo className="mb-5 mt-1 h-8 w-auto" />
      </div>
      This is the <code>DrawerMenu</code> component. Menu content goes here.
    </DrawerContent>
  </Drawer>
);

export const Default: Story = {
  render: () => <StoryCompontnt />,
};
