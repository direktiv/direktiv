import { FC, PropsWithChildren } from "react";

export const Root: FC<PropsWithChildren> = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

export const Sidebar: FC<PropsWithChildren<{ version: string }>> = ({
  children,
  version,
}) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-52 lg:flex-col lg:border-r lg:border-gray-gray5 dark:lg:border-grayDark-gray5">
    {children}
    <div className="hidden shrink-0 p-5 text-left text-sm text-gray-gray8 dark:text-grayDark-gray8 lg:block">
      {version}
    </div>
  </div>
);

export const SidebarTop: FC<PropsWithChildren> = ({ children }) => (
  <div className="grid grid-cols-3 items-center border-b border-gray-gray5 px-6 py-5 dark:border-grayDark-gray5 lg:block lg:space-x-0">
    {children}
  </div>
);

export const SidebarMain: FC<PropsWithChildren> = ({ children }) => (
  <div className="hidden flex-1 overflow-y-auto lg:block">
    <nav className="mt-5 space-y-1 px-3">{children}</nav>
  </div>
);

export const Main: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex flex-col lg:pl-52">
    <main className="flex-1">{children}</main>
  </div>
);

export const MainTop: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex items-center justify-between border-b border-gray-gray5 p-4 dark:border-grayDark-gray5">
    {children}
  </div>
);

export const MainTopLeft: FC<PropsWithChildren> = ({ children }) => (
  <div className="min-w-0 flex-1">
    <h1 className="leading-6 text-gray-gray12 dark:text-grayDark-gray12 sm:truncate">
      {children}
    </h1>
  </div>
);

export const MainTopRight: FC<PropsWithChildren> = ({ children }) => (
  <div className="mt-4 flex space-x-3 sm:mt-0 sm:ml-4">{children}</div>
);

export const MainContent: FC<PropsWithChildren> = ({ children }) => (
  <div>{children}</div>
);
