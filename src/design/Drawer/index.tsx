import { FC, PropsWithChildren } from "react";

import Logo from "../Logo";

export const DrawerRoot: FC<PropsWithChildren> = ({ children }) => (
  <div className="drawer">
    <input id="my-drawer" type="checkbox" className="drawer-toggle" />
    {children}
  </div>
);

export const DrawerContent: {
  ({
    children,
  }: {
    children: (props: {
      drawerLabelProps: React.HTMLProps<HTMLLabelElement>;
    }) => JSX.Element;
  }): JSX.Element;
} = ({ children }) => {
  const drawerLabelProps = {
    htmlFor: "my-drawer",
  };
  return <div className="drawer-content">{children({ drawerLabelProps })}</div>;
};

export const DrawerMenu: FC<PropsWithChildren> = ({ children }) => (
  <div className="drawer-side">
    <label htmlFor="my-drawer" className="drawer-overlay"></label>
    <nav className="menu w-52 bg-base-100 p-4 text-base-content">
      <div className="px-2">
        <Logo className="mb-5 mt-1 h-8 w-auto" />
      </div>
      {children}
    </nav>
  </div>
);
