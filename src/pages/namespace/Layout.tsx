import {
  CurlyBraces,
  LogOut,
  Menu,
  Moon,
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
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../componentsNext/Dropdown";
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
import { useTheme, useThemeActions } from "../../util/store/theme";

import Breadcrumb from "../../componentsStatefull/Breadcrumb";
import Button from "../../componentsNext/Button";
import { FC } from "react";
import Logo from "../../componentsNext/Logo";
import Navigation from "../../componentsNext/Navigation";
import { Outlet } from "react-router-dom";
import { RxChevronDown } from "react-icons/rx";
import clsx from "clsx";
import { useVersion } from "../../api/version";

// TODO: move to own file
const TopRightComponent: FC<{ className?: string }> = ({ className }) => {
  const { setTheme } = useThemeActions();
  const theme = useTheme();
  return (
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
          <DropdownMenuItem
            onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
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
          <Button
            variant="ghost"
            className="placeholder avatar items-center px-1"
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
};

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
                  <Logo
                    iconOnly
                    className="h-8 w-auto justify-self-center sm:hidden"
                  />
                  <Logo className="hidden h-8 w-auto justify-self-center sm:block" />
                  <TopRightComponent className="justify-self-end lg:hidden" />
                </SidebarTop>
                <SidebarMain>
                  <Navigation />
                </SidebarMain>
              </Sidebar>
              <Main>
                <MainTop>
                  <MainTopLeft>
                    <Breadcrumb />
                  </MainTopLeft>
                  <MainTopRight>
                    <TopRightComponent className="max-lg:hidden" />
                  </MainTopRight>
                </MainTop>
                <MainContent>
                  <Outlet />
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
