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
import { Breadcrumb, BreadcrumbRoot } from "../Breadcump";
import { DrawerContent, DrawerMenu, DrawerRoot } from "../Drawer";
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "../Tabs";

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
          <div className="h-7 w-7 rounded-full bg-primary-500 text-neutral-content">
            <span className="text-xs">Ad</span>
          </div>
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
      <DrawerRoot>
        <DrawerContent>
          {({ drawerLabelProps }) => (
            <>
              <Sidebar version="Version: 78c688e">
                <SidebarTop>
                  <label
                    {...drawerLabelProps}
                    className="justify-self-start px-1 lg:hidden"
                    role="button"
                  >
                    <Menu />
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
                      setTheme((theme) =>
                        theme === "dark" ? "light" : "dark"
                      );
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
                      <Breadcrumb>
                        <a className="gap-2">
                          <Home className="h-4 w-auto" />
                          My-namespace
                        </a>
                        &nbsp;
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
                        <a className="gap-2">
                          <Github className="h-4 w-auto" />
                          Example Mirror
                        </a>
                      </Breadcrumb>
                      <Breadcrumb>
                        <a className="gap-2">
                          <FolderOpen className="h-4 w-auto" />
                          Folder
                        </a>
                      </Breadcrumb>
                      <Breadcrumb>
                        <a className="gap-2">
                          <Play className="h-4 w-auto" />
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
                        setTheme((theme) =>
                          theme === "dark" ? "light" : "dark"
                        );
                      }}
                    />
                  </MainTopRight>
                </MainTop>
                <MainContent>
                  <div className="space-y-5 border-b border-gray-5 bg-base-200 p-5 pb-0 dark:border-gray-dark-5">
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
                      <Tabs defaultValue="overview">
                        <TabsList>
                          <TabsTrigger value="overview">
                            <PieChart
                              aria-hidden="true"
                              className="h-4 w-auto"
                            />
                            Overview
                          </TabsTrigger>
                          <TabsTrigger value="active-rev">
                            <GitCommit
                              aria-hidden="true"
                              className="h-4 w-auto"
                            />
                            Active Revisions
                          </TabsTrigger>
                          <TabsTrigger value="revisions">
                            <GitMerge
                              aria-hidden="true"
                              className="h-4 w-auto"
                            />
                            Revisions
                          </TabsTrigger>
                          <TabsTrigger value="settings">
                            <Settings
                              aria-hidden="true"
                              className="h-4 w-auto"
                            />
                            Settings
                          </TabsTrigger>
                        </TabsList>
                        <TabsContent value="account">
                          <p className="text-sm text-gray-8 dark:text-gray-dark-8">
                            Make changes to your account here. Click save when
                            you&apos;re done.
                          </p>
                        </TabsContent>
                        <TabsContent value="password">
                          <p className="text-sm text-gray-8 dark:text-gray-dark-8">
                            Change your password here. After saving, you&apos;ll
                            be logged out.
                          </p>
                        </TabsContent>
                        <TabsContent value="third">
                          <p className="text-sm text-gray-8 dark:text-gray-dark-8">
                            Your third content here
                          </p>
                        </TabsContent>
                        <TabsContent value="fourth">
                          <p className="text-sm text-gray-8 dark:text-gray-dark-8">
                            The fourth content comes here
                          </p>
                        </TabsContent>
                      </Tabs>
                    </div>
                  </div>
                </MainContent>
              </Main>
            </>
          )}
        </DrawerContent>
        <DrawerMenu>
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
        </DrawerMenu>
      </DrawerRoot>
    </Root>
  );
};
