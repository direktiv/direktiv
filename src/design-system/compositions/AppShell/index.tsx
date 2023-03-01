import Button from "../../button";
import { FC } from "react";
import { LogOut } from "lucide-react";
import Logo from "./Logo";

export const Root: FC = ({ children }) => (
  <div className="min-h-full">{children}</div>
);

export const Sidebar: FC = ({ children }) => (
  <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col lg:border-r lg:border-gray-200 lg:pt-5">
    <div className="px-6">
      <Logo className="h-8 w-auto" />
    </div>
    <div className="mt-5 flex h-0 flex-1 flex-col overflow-y-auto pt-1">
      {children}
    </div>
    <div className="flex flex-shrink-0 border-t border-gray-200 p-2 group w-full rounded-md py-5 text-left text-sm font-medium text-gray-700">
      <span className="flex w-full min-w-0 items-center justify-between space-x-3">
        <div className="avatar placeholder">
          <div className="bg-neutral-focus text-neutral-content rounded-full w-10">
            <span className="text-3xl">A</span>
          </div>
        </div>
        <span className="flex min-w-0 flex-1 flex-col">
          <span className="truncate text-sm font-medium text-gray-900">
            admin
          </span>
          <span className="truncate text-sm text-gray-400">
            Version: 78c688e
          </span>
        </span>
        <Button color="link" className="text-grayDark-gray9">
          <LogOut />
        </Button>
      </span>
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
