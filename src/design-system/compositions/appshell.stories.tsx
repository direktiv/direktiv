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
  Moon,
  Network,
  Play,
  PlusCircle,
  Settings,
  Sun,
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
  MainTopLeft,
  MainTopRight,
  Root,
  Sidebar,
  SidebarNavigationItem,
} from "./AppShell";
import { useEffect, useState } from "react";

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

export const Default = () => {
  const [theme, setTheme] = useState<"light" | "dark" | undefined>();

  useEffect(() => {
    const html = document.documentElement;
    const theme = html.getAttribute("data-theme");
    if (theme === "dark") {
      setTheme("dark");
    } else {
      setTheme("light");
    }
  }, []);

  useEffect(() => {
    if (theme === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    } else {
      document.documentElement.setAttribute("data-theme", "light");
    }
  }, [theme]);

  return (
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
          <MainTopLeft>
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
          </MainTopLeft>
          <MainTopRight>
            <Button
              color="ghost"
              className="px-1"
              onClick={() =>
                setTheme((old) => (old === "light" ? "dark" : "light"))
              }
            >
              {theme === "dark" ? (
                <Sun className="w-5 h-auto" />
              ) : (
                <Moon className="w-5 h-auto" />
              )}
            </Button>
          </MainTopRight>
        </MainTopBar>
      </Main>
    </Root>
  );
};
