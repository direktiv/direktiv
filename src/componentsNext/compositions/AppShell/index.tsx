import { FC, PropsWithChildren } from "react";

import Logo from "./Logo";
import LogoNoText from "./LogoNoText";
import clsx from "clsx";

export const Root: FC<PropsWithChildren> = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

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
    <nav className="menu p-4 w-52 bg-base-100 text-base-content">
      <div className="px-2">
        <Logo className="h-8 w-auto mb-5 mt-1" />
      </div>
      {children}
    </nav>
  </div>
);

export const Sidebar: FC<PropsWithChildren<{ version: string }>> = ({
  children,
  version,
}) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-52 lg:flex-col lg:border-r lg:border-gray-gray5 dark:lg:border-grayDark-gray5">
    {children}
    <div className="hidden lg:block flex-shrink-0 p-5 text-left text-sm text-gray-gray8 dark:text-grayDark-gray8">
      {version}
    </div>
  </div>
);

export const SidebarMenu: FC<PropsWithChildren> = ({ children }) => (
  <div className="hidden lg:block flex-1 overflow-y-auto">
    <nav className="mt-5 px-3 space-y-1">{children}</nav>
  </div>
);
export const SidebarLogo: FC<PropsWithChildren> = ({ children }) => (
  <div className="grid max-lg:space-x-5 max-lg:grid-cols-3 items-center lg:block border-b border-gray-gray5 dark:border-grayDark-gray5 px-6 py-5">
    {children}
    <LogoNoText className="sm:hidden h-8 w-auto justify-self-center" />
    <Logo className="hidden sm:block h-8 w-auto" />
  </div>
);

export const SidebarNavigationItem: FC<
  PropsWithChildren<{ href: string; active?: boolean }>
> = ({ children, href, active }) => (
  <a
    href={href}
    className={clsx(
      active
        ? "bg-primary50 dark:bg-primary700 text-gray-gray12 dark:text-grayDark-gray12"
        : "text-gray-gray11 dark:text-grayDark-gray11 hover:bg-gray-gray2 dark:hover:bg-grayDark-gray2",
      "group flex items-center p-2 text-sm font-medium rounded-md [&>svg]:mr-3 [&>svg]:group"
    )}
  >
    {children}
  </a>
);

export const Main: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex flex-col lg:pl-52">
    <main className="flex-1">{children}</main>
  </div>
);

export const MainTopBarRoot: FC<PropsWithChildren> = ({ children }) => (
  <div className="border-b border-gray-gray5 dark:border-grayDark-gray5 p-4 flex items-center justify-between">
    {children}
  </div>
);

export const MainTopBarLeft: FC<PropsWithChildren> = ({ children }) => (
  <div className="min-w-0 flex-1">
    <h1 className="text-lg font-medium leading-6 text-gray-gray12 dark:text-grayDark-gray12 sm:truncate">
      {children}
    </h1>
  </div>
);

export const MainTopBarRight: FC<PropsWithChildren> = ({ children }) => (
  <div className="mt-4 space-x-3 flex sm:mt-0 sm:ml-4">{children}</div>
);

export const MainContent: FC<PropsWithChildren> = ({ children }) => (
  <div>{children}</div>
);
