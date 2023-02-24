import {
  Bug,
  Calendar,
  FolderTree,
  Laptop,
  Layers,
  LogOut,
  Settings,
} from "lucide-react";

import Logo from "./Logo";
import clsx from "clsx";
import Button from "../button";

export default {
  title: "Design System/Compositions/App Shell",
};

const navigation = [
  { name: "Explorer", href: "#", icon: FolderTree, current: true },
  { name: "Monitoring", href: "#", icon: Bug, current: false },
  { name: "Instances", href: "#", icon: Laptop, current: false },
  { name: "Events", href: "#", icon: Calendar, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

export const Default = () => (
  <div className="min-h-full">
    <div className="lg:fixed lg:inset-y-0 lg:flex lg:w-64 lg:flex-col lg:border-r lg:border-gray-200 lg:bg-gray-100 lg:pt-5 lg:pb-4">
      <div className="flex flex-shrink-0 items-center px-6">
        <Logo className="h-8 w-auto" />
      </div>
      <div className="mt-5 flex h-0 flex-1 flex-col overflow-y-auto pt-1">
        <nav className="mt-6 px-3">
          <div className="space-y-1">
            {navigation.map((item) => (
              <a
                key={item.name}
                href={item.href}
                className={clsx(
                  item.current
                    ? "bg-gray-200 text-gray-900"
                    : "text-gray-700 hover:text-gray-900 hover:bg-gray-50",
                  "group flex items-center px-2 py-2 text-sm font-medium rounded-md"
                )}
                aria-current={item.current ? "page" : undefined}
              >
                <item.icon
                  className={clsx(
                    item.current
                      ? "text-gray-500"
                      : "text-gray-400 group-hover:text-gray-500",
                    "mr-3 flex-shrink-0 h-6 w-6"
                  )}
                  aria-hidden="true"
                />
                {item.name}
              </a>
            ))}
          </div>
        </nav>
      </div>
      <div className="flex flex-shrink-0 border-t border-gray-200 p-4">
        <div className="group w-full rounded-md bg-gray-100 px-3.5 py-2 text-left text-sm font-medium text-gray-700">
          <span className="flex w-full items-center justify-between">
            <span className="flex min-w-0 items-center justify-between space-x-3">
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
                  Version #78c688e
                </span>
              </span>
            </span>
          </span>
        </div>
      </div>
      <div className="flex flex-shrink-0 border-t border-gray-200 p-4">
        <Button color="ghost" block className="gap-2 justify-start">
          <LogOut /> Logout
        </Button>
      </div>
    </div>
  </div>
);
