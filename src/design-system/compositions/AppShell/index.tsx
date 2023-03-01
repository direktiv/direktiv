import { FC } from "react";
import Logo from "./Logo";

export const Root: FC = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

export const Sidebar: FC<{ version: string }> = ({ children, version }) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col lg:border-r lg:border-gray-gray5 dark:lg:border-grayDark-gray5 lg:pt-5">
    <div className="px-6">
      <Logo className="h-8 w-auto" />
    </div>
    <div className="mt-5 flex h-0 flex-1 flex-col overflow-y-auto pt-1">
      {children}
    </div>
    <div className="flex flex-shrink-0 p-5 text-left text-sm text-gray-gray8 dark:text-grayDark-gray8">
      {version}
    </div>
  </div>
);

export const SidebarMenu: FC = ({ children }) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col lg:border-r lg:border-gray-200 lg:pt-5">
    <div className="px-6">
      <Logo className="h-8 w-auto" />
    </div>
    {children}
  </div>
);

export const Main: FC = ({ children }) => (
  <div className="flex flex-col lg:pl-64">
    <main className="flex-1">{children}</main>
  </div>
);

export const MainTopBar: FC = ({ children }) => (
  <div className="border-b border-gray-gray5 dark:border-grayDark-gray5 px-4 py-4 sm:flex sm:items-center sm:justify-between">
    {children}
  </div>
);
