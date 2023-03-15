import {
  ChevronsUpDown,
  FolderOpen,
  Github,
  Home,
  Menu,
  Play,
  PlusCircle,
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

import Button from "../componentsNext/Button";
import Navigation from "../componentsNext/Navigation";
import { Outlet } from "react-router-dom";
import { useVersion } from "../api/version";

const Layout = () => {
  const { data: version } = useVersion();
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
                    <div className="breadcrumbs text-sm">
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
                  <MainTopRight>2</MainTopRight>
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
