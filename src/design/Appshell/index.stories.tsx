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
  Menu,
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
import { Breadcrumb, BreadcrumbRoot } from "../Breadcrumbs";
import { Drawer, DrawerContent, DrawerMain, DrawerTrigger } from "../Drawer";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../Dropdown";
import { FC, useEffect, useState } from "react";
import {
  Main,
  MainContent,
  MainTop,
  MainTopLeft,
  MainTopRight,
  Root,
  Sidebar,
  SidebarMain,
  SidebarTop,
} from "./index";

import Avatar from "../Avatar";
import Button from "../Button";
import Logo from "../Logo";
import { NavigationLink } from "../NavigationLink";
import { RxChevronDown } from "react-icons/rx";
import clsx from "clsx";

export default {
  title: "Components/AppShell",
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

export const Default = () => (
  <Root>
    <Sidebar version="version">
      <SidebarTop>
        <div className="lg:hidden">Burger Menu Button</div>
      </SidebarTop>
      <SidebarMain>Sidebar</SidebarMain>
    </Sidebar>
    <Main>
      <MainTop>
        <MainTopLeft>Top Left</MainTopLeft>
        <MainTopRight>Top Right</MainTopRight>
      </MainTop>
      <MainContent>
        <div className="p-10">Main Content</div>
      </MainContent>
    </Main>
  </Root>
);

const TopRightComponent: FC<{
  className?: string;
  theme: "light" | "dark" | undefined;
  onThemeChange: () => void;
}> = ({ className, theme, onThemeChange }) => (
  <div className={clsx("flex space-x-2", className)}>
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" icon>
          <Settings2 />
          <RxChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuLabel>Appearance</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={onThemeChange}>
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
        <Button
          variant="ghost"
          className="placeholder avatar"
          role="button"
          icon
        >
          <Avatar>
            AD
          </Avatar>
          <RxChevronDown />
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56">
        <DropdownMenuLabel>You are logged in as admin</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem>
          <LogOut className="mr-2 h-4 w-4" />
          <span>Logout</span>
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  </div>
);

export const MoreDetailedShell = () => {
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
      <Drawer>
        <DrawerContent id="my-drawer">
          <div className="drawer-side">
            <label htmlFor="my-drawer" className="drawer-overlay"></label>
            <nav className="menu bg-gray-1 p-4 text-gray-11 dark:bg-gray-dark-1 ">
              <div className="px-2">
                <Logo className="mb-5 mt-1 h-8 w-auto" />
              </div>
              {navigation.map((item) => (
                <NavigationLink
                  key={item.name}
                  href={item.href}
                  active={item.current}
                >
                  <item.icon aria-hidden="true" />
                  {item.name}
                </NavigationLink>
              ))}
            </nav>
          </div>
        </DrawerContent>
        <DrawerMain>
          <Sidebar version="Version: 78c688e">
            <SidebarTop>
              <label
                className="justify-self-start px-1 lg:hidden"
                role="button"
              >
                <DrawerTrigger>
                  <div className="text-gray-10 dark:text-gray-dark-10">
                    <Menu />
                  </div>
                </DrawerTrigger>
              </label>
              <Logo
                iconOnly
                className="h-8 w-auto justify-self-center sm:hidden"
              />
              <Logo className="hidden h-8 w-auto justify-self-center sm:block" />
              <TopRightComponent
                className="justify-self-end lg:hidden"
                theme={theme}
                onThemeChange={() => {
                  setTheme((theme) => (theme === "dark" ? "light" : "dark"));
                }}
              />
            </SidebarTop>
            <SidebarMain>
              {navigation.map((item) => (
                <NavigationLink
                  key={item.name}
                  href={item.href}
                  active={item.current}
                >
                  <item.icon aria-hidden="true" />
                  {item.name}
                </NavigationLink>
              ))}
            </SidebarMain>
          </Sidebar>
          <Main>
            <MainTop>
              <MainTopLeft>
                <BreadcrumbRoot>
                  <Breadcrumb noArrow>
                    <a>
                      <Home />
                      My-namespace
                    </a>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button size="sm" variant="ghost" circle icon>
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
                  </Breadcrumb>
                  <Breadcrumb>
                    <a>
                      <Github />
                      Example Mirror
                    </a>
                  </Breadcrumb>
                  <Breadcrumb>
                    <a>
                      <FolderOpen />
                      Folder
                    </a>
                  </Breadcrumb>
                  <Breadcrumb>
                    <a>
                      <Play />
                      workflow.yml
                    </a>
                  </Breadcrumb>
                </BreadcrumbRoot>
              </MainTopLeft>
              <MainTopRight>
                <TopRightComponent
                  className="max-lg:hidden"
                  theme={theme}
                  onThemeChange={() => {
                    setTheme((theme) => (theme === "dark" ? "light" : "dark"));
                  }}
                />
              </MainTopRight>
            </MainTop>
            <MainContent>
              <div className="space-y-5 border-b border-gray-5 bg-gray-2 p-5 pb-0 dark:border-gray-dark-5 dark:bg-gray-dark-2">
                <div className="flex flex-col max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between ">
                  <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
                    <Play className="h-5" />
                    workflow.yml
                  </h3>
                  <Button variant="primary">
                    Actions <RxChevronDown />
                  </Button>
                </div>
                <div>
                  <nav className="-mb-px flex space-x-8">
                    {tabs.map((tab) => (
                      <a
                        key={tab.name}
                        href={tab.href}
                        className={clsx(
                          tab.current
                            ? "border-primary-500 text-primary-500"
                            : "border-transparent text-gray-11 hover:border-gray-8 hover:text-gray-12 dark:hover:border-gray-dark-8 dark:hover:text-gray-dark-12",
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
        </DrawerMain>
      </Drawer>
    </Root>
  );
};
