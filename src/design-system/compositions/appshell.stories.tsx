import {
  Box,
  Bug,
  Calendar,
  ChevronsUpDown,
  CurlyBraces,
  FolderOpen,
  FolderTree,
  GitCommit,
  GitMerge,
  Github,
  Home,
  Layers,
  LogOut,
  Moon,
  Network,
  PieChart,
  Play,
  PlusCircle,
  Settings,
  Settings2,
  Slack,
  Sun,
  Terminal,
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
  MainContent,
  MainTopBar,
  MainTopLeft,
  MainTopRight,
  Root,
  Sidebar,
  SidebarNavigationItem,
} from "./AppShell";
import { useEffect, useState } from "react";

import Button from "../button";
import { RxChevronDown } from "react-icons/rx";
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
  { name: "Gateway", href: "#", icon: Network, current: false },
  { name: "Permissions", href: "#", icon: Users, current: false },
  { name: "Services", href: "#", icon: Layers, current: false },
  { name: "Settings", href: "#", icon: Settings, current: false },
];

const tabs = [
  { name: "Overview", href: "#", icon: PieChart, current: true },
  { name: "Active Revisions", href: "#", icon: GitCommit, current: false },
  { name: "Revisions", href: "#", icon: GitMerge, current: false },
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
              // onClick={() =>
              //   setTheme((old) => (old === "light" ? "dark" : "light"))
              // }
            ></Button>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button color="ghost" className="px-1">
                  <Settings2 />
                  <RxChevronDown />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuLabel>Appearance</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem
                  onClick={() =>
                    setTheme((old) => (old === "light" ? "dark" : "light"))
                  }
                >
                  {theme === "dark" ? (
                    <>
                      <Sun className="mr-2 h-4 w-4" />
                      switch to Light mode
                    </>
                  ) : (
                    <>
                      <Moon className="mr-2 h-4 w-4" />
                      switch to dark mode
                    </>
                  )}
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>Help</DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <Terminal className="mr-2 h-4 w-4" /> Show API Commands
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <CurlyBraces className="mr-2 h-4 w-4" /> Open JQ Playground
                </DropdownMenuItem>
                <DropdownMenuItem>
                  <Slack className="mr-2 h-4 w-4" /> Support Channel on Slack
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <div
                  className="avatar placeholder items-center gap-1"
                  role="button"
                >
                  <div className="bg-primary500 text-neutral-content rounded-full w-7 h-7">
                    <span className="text-xs">Ad</span>
                  </div>
                  <RxChevronDown />
                </div>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56">
                <DropdownMenuLabel>
                  You are logged in as admin
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  <LogOut className="mr-2 h-4 w-4" />
                  <span>Logout</span>
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </MainTopRight>
        </MainTopBar>
        <MainContent>
          {/* <div className="flex ">
            <div className="tabs">
              <a className="tab tab-bordered tab-active">Overview</a>
              <a className="tab tab-bordered">Revisions</a>
              <a className="tab tab-bordered">Active Revision</a>
              <a className="tab tab-bordered">Settings</a>
            </div>
            <Button color="primary" className="ml-auto">
              wdede
            </Button>
          </div> */}

          <div className="p-5 space-y-5 border-b bg-base-200 border-gray-gray5 dark:border-grayDark-gray5 pb-5 sm:pb-0">
            <div className="md:flex md:items-center md:justify-between ">
              <h3 className="flex font-mono font-bold mono items-center  gap-x-2 text-primary500">
                <Play className="h-5" />
                workflow.yml
              </h3>
              <div className="mt-3 flex gap-3 md:mt-0">
                <Button color="primary" size="sm">
                  Actions <RxChevronDown />
                </Button>
              </div>
            </div>
            <div>
              <nav className="-mb-px flex space-x-8">
                {tabs.map((tab) => (
                  <a
                    key={tab.name}
                    href={tab.href}
                    className={clsx(
                      tab.current
                        ? "border-primary500 text-primary500"
                        : "border-transparent text-gray-gray11 hover:border-gray-gray8 hover:text-gray-gray12 dark:hover:border-grayDark-gray8 dark:hover:text-grayDark-gray12",
                      "flex items-center gap-x-2 whitespace-nowrap border-b-2 px-1 pb-4 text-sm font-medium"
                    )}
                    aria-current={tab.current ? "page" : undefined}
                  >
                    <tab.icon aria-hidden="true" className="h-4 w-auto" />{" "}
                    {tab.name}
                  </a>
                ))}
              </nav>
            </div>
          </div>
        </MainContent>
      </Main>
    </Root>
  );
};
