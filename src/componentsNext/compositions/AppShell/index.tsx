import { FC, PropsWithChildren } from "react";

import Logo from "./../../Logo";
import clsx from "clsx";

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

export const SidebarNavigationItem: FC<
  PropsWithChildren<{ href: string; active?: boolean }>
> = ({ children, href, active }) => (
  <a
    href={href}
    className={clsx(
      active
        ? "bg-primary50 text-gray-gray12 dark:bg-primary700 dark:text-grayDark-gray12"
        : "text-gray-gray11 hover:bg-gray-gray2 dark:text-grayDark-gray11 dark:hover:bg-grayDark-gray2",
      "[&>svg]:group group flex items-center rounded-md p-2 text-sm font-medium [&>svg]:mr-3"
    )}
  >
    {children}
  </a>
);
