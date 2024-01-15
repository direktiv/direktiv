import { Drawer, DrawerContent, DrawerTrigger } from "./index";
import Button from "../Button";
import Logo from "../Logo";
import type { Meta } from "@storybook/react";

export default {
  title: "Components/Drawer",
  parameters: { layout: "fullscreen" },
} satisfies Meta<typeof Drawer>;

export const Default = () => (
  <Drawer>
    <div className="flex flex-col items-start space-y-5 p-10">
      <DrawerTrigger>
        <Button> Open Drawer</Button>
      </DrawerTrigger>
      <div>
        This is the <code>DrawerContent</code> component. Make sure to place it
        as a direct child of the <code>DrawerRoot</code> component and place the{" "}
        <code>DrawerMenu</code> directly after the <code>DrawerContent</code>.
      </div>
    </div>
    <DrawerContent>
      <Logo className="mx-2 mb-5 mt-1 h-8 w-auto" />
      This is the <code>DrawerMenu</code> component. Menu content goes here.
    </DrawerContent>
  </Drawer>
);
