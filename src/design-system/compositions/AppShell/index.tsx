import { FC } from "react";
import Logo from "./Logo";
import clsx from "clsx";

export const Root: FC = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

export const Sidebar: FC<{ version: string }> = ({ children, version }) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col lg:border-r lg:border-gray-gray5 dark:lg:border-grayDark-gray5">
    <div className="border-b border-gray-gray5 dark:border-grayDark-gray5 px-6 py-5">
      <Logo className="h-8 w-auto border-1 border-gray-800" />
    </div>
    <div className="flex-1 overflow-y-auto">
      <nav className="mt-5 px-3 space-y-1">{children}</nav>
    </div>
    <div className="flex-shrink-0 p-5 text-left text-sm text-gray-gray8 dark:text-grayDark-gray8">
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
        : "text-gray-gray10 dark:text-grayDark-gray10 hover:bg-gray-gray2 dark:hover:bg-grayDark-gray2",
      "group flex items-center px-2 py-2 text-sm font-medium rounded-md [&>svg]:mr-3 [&>svg]:group"
    )}
  >
    {children}
  </a>
);

export const Main: FC = ({ children }) => (
  <div className="flex flex-col lg:pl-64">
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
  <div className="mt-4 flex sm:mt-0 sm:ml-4">{children}</div>
);
