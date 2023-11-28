import { Drawer, DrawerContent, DrawerTrigger } from "~/design/Drawer";
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
} from "~/design/Appshell";
import { Outlet, useParams } from "react-router-dom";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";

import Breadcrumb from "~/componentsNext/Breadcrumb";
import Logo from "~/design/Logo";
import { Menu } from "lucide-react";
import Navigation from "~/componentsNext/Navigation";
import NotificationMenu from "~/componentsNext/NotificationMenu";
import UserMenu from "~/componentsNext/UserMenu";
import { useEffect } from "react";
import { useVersion } from "~/api/version/query/get";

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
            <div className="flex gap-3 justify-self-end lg:hidden">
              {/* error would be thrown if namespace is not yet defined */}
              {!!namespace && <NotificationMenu />}
              <UserMenu />
            </div>
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
              {/* error would be thrown if namespace is not yet defined */}
              {!!namespace && <NotificationMenu className="max-lg:hidden" />}
              <UserMenu className="max-lg:hidden" />
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
