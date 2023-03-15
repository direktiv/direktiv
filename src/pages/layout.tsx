import {
  DrawerContent,
  DrawerMenu,
  DrawerRoot,
} from "../componentsNext/Drawer";
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

import { Menu } from "lucide-react";
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
                  <MainTopLeft>1</MainTopLeft>
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
