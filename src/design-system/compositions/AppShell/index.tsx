import Button from "../../button";
import { FC } from "react";
import Logo from "./Logo";
import { Menu } from "lucide-react";
import clsx from "clsx";

export const Root: FC = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

export const Sidebar: FC<{ version: string }> = ({ children, version }) => (
  <div className="lg:fixed lg:flex lg:w-52 lg:flex-col lg:border-r lg:border-gray-gray5 dark:lg:border-grayDark-gray5">
    <div className="grid max-lg:space-x-5 max-lg:grid-cols-3 items-center lg:block border-b border-gray-gray5 dark:border-grayDark-gray5 px-6 py-5">
      <Button color="ghost" className="lg:hidden px-1 justify-self-start">
        <Menu />
      </Button>
      <Logo className="h-8 justify-self-center  w-auto border-1 border-gray-800" />
    </div>
    <div className="hidden lg:block flex-1 overflow-y-auto">
      <nav className="mt-5 px-3 space-y-1">{children}</nav>
    </div>
    <div className="hidden lg:block flex-shrink-0 p-5 text-left text-sm text-gray-gray8 dark:text-grayDark-gray8">
      {version}
    </div>
  </div>
);

export const SidebarNavigationItem: FC<{ href: string; active?: boolean }> = ({
  children,
  href,
  active,
}) => (
  <a
    href={href}
    className={clsx(
      active
        ? "bg-primary50 dark:bg-primary700 text-gray-gray12 dark:text-grayDark-gray12"
        : "text-gray-gray11 dark:text-grayDark-gray11 hover:bg-gray-gray2 dark:hover:bg-grayDark-gray2",
      "group flex items-center px-2 py-2 text-sm font-medium rounded-md [&>svg]:mr-3 [&>svg]:group"
    )}
  >
    {children}
  </a>
);

export const Main: FC = ({ children }) => (
  <div className="flex flex-col lg:pl-52">
    <main className="flex-1">{children}</main>
  </div>
);

export const MainTopBar: FC = ({ children }) => (
  <div className="border-b border-gray-gray5 dark:border-grayDark-gray5 p-4 sm:flex sm:items-center sm:justify-between">
    {children}
  </div>
);

export const MainTopLeft: FC = ({ children }) => (
  <div className="min-w-0 flex-1">
    <h1 className="text-lg font-medium leading-6 text-gray-gray12 dark:text-grayDark-gray12 sm:truncate">
      {children}
    </h1>
  </div>
);

export const MainTopRight: FC = ({ children }) => (
  <div className="mt-4 space-x-3 flex sm:mt-0 sm:ml-4">{children}</div>
);

export const MainContent: FC = ({ children }) => <div>{children}</div>;
