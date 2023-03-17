import { Breadcrumb, BreadcrumbRoot } from "../../componentsNext/Breadcump";
import {
  ChevronsUpDown,
  CurlyBraces,
  FolderOpen,
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
} from "../../componentsNext/Drawer";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../componentsNext/Dropdown";
import { FC, Fragment } from "react";
import { Link, Outlet, useNavigate } from "react-router-dom";
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
} from "../../componentsNext/Appshell";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";
import { useTheme, useThemeActions } from "../../util/store/theme";

import Button from "../../componentsNext/Button";
import Navigation from "../../componentsNext/Navigation";
import { RxChevronDown } from "react-icons/rx";
import clsx from "clsx";
import { pages } from "../../util/router/pages";
import { useNamespaces } from "../../api/namespaces";
import { useTree } from "../../api/tree";
import { useVersion } from "../../api/version";

const BreadcrumbComponent: FC<{ path: string }> = ({ path }) => {
  // split path string in to chunks, using the last / as the separator
  const segments = path.split("/");
  const namespace = useNamespace();

  const { data, isLoading } = useTree({
    directory: path,
  });

  if (!namespace) return null;

  let Icon = FolderOpen;

  if (data?.node.type === "directory") {
    Icon = FolderOpen;
  }

  if (data?.node.type === "workflow") {
    Icon = Play;
  }

  return (
    <Breadcrumb>
      <Link
        to={pages.explorer.createHref({ namespace, directory: path })}
        className="gap-2"
      >
        <Icon aria-hidden="true" className={clsx(isLoading && "invisible")} />
        {segments.slice(-1)}
      </Link>
    </Breadcrumb>
  );
};

const Layout = () => {
  const { data: version } = useVersion();
  const { setTheme } = useThemeActions();

  const { data: availableNamespaces, isLoading } = useNamespaces();
  const namespace = useNamespace();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  const theme = useTheme();

  if (!namespace) return null;

  const { directory } = pages.explorer.useParams();

  const onNameSpaceChange = (namespace: string) => {
    setNamespace(namespace);
    navigate(pages.explorer.createHref({ namespace }));
  };

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
                        <Link
                          to={pages.explorer.createHref({ namespace })}
                          className="gap-2"
                        >
                          <Home />
                          {namespace}
                        </Link>
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
                            <DropdownMenuRadioGroup
                              value={namespace}
                              onValueChange={onNameSpaceChange}
                            >
                              {availableNamespaces?.results.map((ns) => (
                                <DropdownMenuRadioItem
                                  key={ns.name}
                                  value={ns.name}
                                  textValue={ns.name}
                                >
                                  {ns.name}
                                </DropdownMenuRadioItem>
                              ))}
                            </DropdownMenuRadioGroup>
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
                      {/* TODO: extract this into a util and write some tests */}
                      {directory &&
                        directory?.split("/").map((segment, index, srcArr) => {
                          const absolutePath = srcArr
                            .slice(0, index + 1)
                            .join("/");
                          return (
                            <BreadcrumbComponent
                              key={absolutePath}
                              path={absolutePath}
                            />
                          );
                        })}
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
