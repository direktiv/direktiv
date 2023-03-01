import {
  Box,
  Bug,
  Calendar,
  ChevronsUpDown,
  FolderOpen,
  FolderTree,
  Github,
  Home,
  Layers,
  Network,
  Play,
  PlusCircle,
  Settings,
  Settings2,
  Users,
} from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../dropdown";
import { Root, Sidebar } from "./AppShell";

import Button from "../button";
import clsx from "clsx";

export default {
  title: "Design System/Compositions/App Shell",
  parameters: { layout: "fullscreen" },
};

const navigation = [
  { name: "Explorer", href: "#", icon: FolderTree, current: true },
  { name: "Monitoring", href: "#", icon: Bug, current: false },
  { name: "Instances", href: "#", icon: Box, current: false },
  { name: "Events", href: "#", icon: Calendar, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

const enterprise = [
  { name: "Gateway", href: "#", icon: Network, current: false },
  { name: "Permissions", href: "#", icon: Users, current: false },
];

export const Default = () => (
  <Root>
    <Sidebar>
      <nav className="mt-6 px-3 space-y-3">
        <div className="space-y-1">
          {navigation.map((item) => (
            <a
              key={item.name}
              href={item.href}
              className={clsx(
                item.current
                  ? "bg-primary50 text-gray-900"
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
        <div className="divider"></div>
        <div className="space-y-1">
          {enterprise.map((item) => (
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
    </Sidebar>
    <div className="flex flex-col lg:pl-64">
      <main className="flex-1">
        <div className="border-b border-gray-200 px-4 py-4 sm:flex sm:items-center sm:justify-between">
          <div className="min-w-0 flex-1">
            <h1 className="text-lg font-medium leading-6 text-gray-900 sm:truncate">
              <div className="text-sm breadcrumbs">
                <ul>
                  <li>
                    <a className="gap-2">
                      <Home className="h-4 w-auto" />
                      My-namespace
                    </a>
                    &nbsp;
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button size="xs" color="ghost" circle>
                          <ChevronsUpDown />
                        </Button>
                      </DropdownMenuTrigger>
                      <DropdownMenuContent className="w-56">
                        <DropdownMenuLabel>Namespaces</DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuCheckboxItem checked>
                          My-namespace
                        </DropdownMenuCheckboxItem>
                        <DropdownMenuCheckboxItem>
                          second-namespace
                        </DropdownMenuCheckboxItem>
                        <DropdownMenuCheckboxItem>
                          another-namespace
                        </DropdownMenuCheckboxItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem>
                          <PlusCircle className="mr-2 h-4 w-4" />
                          <span>Create new namespace</span>
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </li>
                  <li>
                    <a className="gap-2">
                      <Github className="h-4 w-auto" />
                      Example Mirror
                    </a>
                  </li>
                  <li>
                    <a className="gap-2">
                      <FolderOpen className="h-4 w-auto" />
                      Folder
                    </a>
                  </li>
                  <li>
                    <a className="gap-2">
                      <Play className="h-4 w-auto" />
                      workflow.yml
                    </a>
                  </li>
                </ul>
              </div>
            </h1>
          </div>
          <div className="mt-4 flex sm:mt-0 sm:ml-4">
            <Button color="ghost" className="px-1">
              <Settings2 className="w-5 h-auto" />
            </Button>
          </div>
        </div>
      </main>
    </div>
  </Root>
);
