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
import {
  Main,
  MainTopBar,
  Root,
  Sidebar,
  SidebarNavigationItem,
} from "./AppShell";

import Button from "../button";

export default {
  title: "Design System/Compositions/App Shell",
  parameters: { layout: "fullscreen" },
};

const navigation = [
  { name: "Explorer", href: "#", icon: FolderTree, current: true },
  { name: "Monitoring", href: "#", icon: Bug, current: false },
  { name: "Instances", href: "#", icon: Box, current: false },
  { name: "Events", href: "#", icon: Calendar, current: false },
  { name: "Gateway", href: "#", icon: Network, current: false },
  { name: "Permissions", href: "#", icon: Users, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

export const Default = () => (
  <Root>
    <Sidebar version="Version: 78c688e">
      {navigation.map((item) => (
        <SidebarNavigationItem
          key={item.name}
          href={item.href}
          active={item.current}
        >
          <item.icon aria-hidden="true" />
          {item.name}
        </SidebarNavigationItem>
      ))}
    </Sidebar>
    <Main>
      <MainTopBar>
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
      </MainTopBar>
    </Main>
  </Root>
);
