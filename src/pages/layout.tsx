import { Breadcrumb, BreadcrumbRoot } from "../componentsNext/Breadcump";
import {
  ChevronsUpDown,
  CurlyBraces,
  FolderOpen,
  Github,
  Home,
  Loader2,
  LogOut,
  Menu,
  Moon,
  Play,
  PlusCircle,
  Settings2,
  Slack,
  Sun,
  Terminal,
} from "lucide-react";
import {
  DrawerContent,
  DrawerMenu,
  DrawerRoot,
} from "../componentsNext/Drawer";
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../componentsNext/Dropdown";
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
} from "../componentsNext/Appshell";
import { useTheme, useThemeActions } from "../util/store/theme";

import Button from "../componentsNext/Button";
import Navigation from "../componentsNext/Navigation";
import { Outlet } from "react-router-dom";
import { RxChevronDown } from "react-icons/rx";
import { useNamespace } from "../util/store/namespace";
import { useNamespaces } from "../api/namespaces";
import { useVersion } from "../api/version";

const Layout = () => {
  const { data: version } = useVersion();
  const { setTheme } = useThemeActions();

  const { data: availableNamespaces, isLoading } = useNamespaces();
  const activeNamespace = useNamespace();

  const theme = useTheme();
  return (
    <Root>
      <DrawerRoot>
        <DrawerContent>
          {({ drawerLabelProps }) => (
            <>
              <Sidebar version={version?.api ?? ""}>
                <SidebarTop>
                  <label
                    {...drawerLabelProps}
                    className="justify-self-start px-1 lg:hidden"
                    role="button"
                  >
                    <Menu />
                  </label>
                </SidebarTop>
                <SidebarMain>
                  <Navigation />
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
                            <Button size="xs" color="ghost" circle>
                              <ChevronsUpDown />
                            </Button>
                          </DropdownMenuTrigger>
                          <DropdownMenuContent className="w-56">
                            <DropdownMenuLabel>Namespaces</DropdownMenuLabel>
                            <DropdownMenuSeparator />
                            {availableNamespaces?.results.map((namespace) => (
                              <DropdownMenuCheckboxItem
                                key={namespace.name}
                                checked={activeNamespace === namespace.name}
                              >
                                {namespace.name}
                              </DropdownMenuCheckboxItem>
                            ))}
                            {isLoading && (
                              <DropdownMenuItem disabled>
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                loading...
                              </DropdownMenuItem>
                            )}
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
                            setTheme(theme === "dark" ? "light" : "dark")
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
                          <Terminal className="mr-2 h-4 w-4" /> Show API
                          Commands
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <CurlyBraces className="mr-2 h-4 w-4" /> Open JQ
                          Playground
                        </DropdownMenuItem>
                        <DropdownMenuItem>
                          <Slack className="mr-2 h-4 w-4" /> Support Channel on
                          Slack
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild>
                        <Button
                          color="ghost"
                          className="placeholder avatar items-center gap-1 px-1"
                          role="button"
                        >
                          <div className="h-7 w-7 rounded-full bg-primary500 text-neutral-content">
                            <span className="text-xs">Ad</span>
                          </div>
                          <RxChevronDown />
                        </Button>
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
                </MainTop>
                <MainContent>
                  <div className="flex flex-col space-y-5 p-10">
                    <Outlet />
                  </div>
                </MainContent>
              </Main>
            </>
          )}
        </DrawerContent>
        <DrawerMenu>
          <Navigation />
        </DrawerMenu>
      </DrawerRoot>
    </Root>
  );
};

export default Layout;
