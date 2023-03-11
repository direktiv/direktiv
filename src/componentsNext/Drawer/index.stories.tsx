import { DrawerContent, DrawerMenu, DrawerRoot } from "./index";

export default {
  title: "Components (next)/Drawer",
  parameters: { layout: "fullscreen" },
};

export const Default = () => (
  <DrawerRoot>
    <DrawerContent>
      {({ drawerLabelProps }) => (
        <div className="flex flex-col space-y-5 p-10">
          <label {...drawerLabelProps} className="btn w-40" role="button">
            Open Drawer
          </label>
          <div>
            This is the <code>DrawerContent</code> component. Make sure to place
            it as a direct child of the <code>DrawerRoot</code> component and
            place the <code>DrawerMenu</code> directly after the{" "}
            <code>DrawerContent</code>.
          </div>
        </div>
      )}
    </DrawerContent>
    <DrawerMenu>
      This is the <code>DrawerMenu</code> component. Menu content goes here.
    </DrawerMenu>
  </DrawerRoot>
);
