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
import { Drawer, DrawerContent, DrawerTrigger } from "../../design/Drawer";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "../../design/Dropdown";
import { FC, useEffect } from "react";
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
} from "../../design/Appshell";
import { Outlet, useLocation, useParams } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "../../util/store/namespace";
import { useTheme, useThemeActions } from "../../util/store/theme";

import Avatar from "../../design/Avatar";
import Breadcrumb from "../../componentsNext/Breadcrumb";
import Button from "../../design/Button";
import Logo from "../../design/Logo";
import Navigation from "../../componentsNext/Navigation";
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
          <Button variant="ghost" icon data-testid="dropdown-trg-appearance">
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
            className="items-center px-1"
            role="button"
            icon
            data-testid="dropdown-trg-user"
          >
            <Avatar>Ad</Avatar>
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
  const namespace = useNamespace();
  const { setNamespace } = useNamespaceActions();
  const { namespace: namespaceFromUrl } = useParams();

  // when url with namespace is called directly, this updates ns in local store
  useEffect(() => {
    if (namespace === namespaceFromUrl) {
      return;
    }

    if (namespaceFromUrl) {
      setNamespace(namespaceFromUrl);
    }
  }, [namespace, setNamespace, namespaceFromUrl]);

  return (
    <Root>
      <Drawer>
        <Sidebar version={version?.api ?? ""}>
          <SidebarTop>
            <label className="justify-self-start px-1 lg:hidden" role="button">
              <DrawerTrigger asChild>
                <Menu />
              </DrawerTrigger>
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
            {/* error would be thrown if namespace is not yet defined */}
            {!!namespace && <Outlet />}
          </MainContent>
        </Main>
        <DrawerContent>
          <Logo className="mx-2 mb-5 mt-1 h-8 w-auto" />
          <Navigation />
        </DrawerContent>
      </Drawer>
    </Root>
  );
};

export default Layout;
