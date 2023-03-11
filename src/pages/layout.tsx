import { Link, Outlet } from "react-router-dom";
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
import { useApiActions, useApiKey } from "../util/store/apiKey";
import { useTheme, useThemeActions } from "../util/store/theme";

import { useEffect } from "react";
import { useNamespaces } from "../api/namespaces";
import { useVersion } from "../api/version";

const Layout = () => {
  const apiKey = useApiKey();
  const { setApiKey } = useApiActions();
  const theme = useTheme();
  const { setTheme } = useThemeActions();

  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } = useNamespaces();

  useEffect(() => {
    let applyTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light";
    if (theme) applyTheme = theme;
    document.querySelector("html")?.setAttribute("data-theme", applyTheme);
  }, [theme]);

  return (
    <Root>
      <div className="flex flex-col space-y-5 p-10">
        <div>
          <h1>
            theme <span className="font-bold">{theme}</span>
          </h1>
          <div className="flex space-x-5">
            <button
              className="btn-primary btn"
              onClick={() => setTheme("dark")}
            >
              darkmode
            </button>
            <button
              className="btn-primary btn"
              onClick={() => setTheme("light")}
            >
              lightmode
            </button>
            <button className="btn-error btn" onClick={() => setTheme(null)}>
              reset theme
            </button>
          </div>
        </div>
        <div>
          <h1>
            api key is <span className="font-bold">{apiKey}</span>
          </h1>
          <div className="flex space-x-5">
            <button
              className="btn-primary btn"
              onClick={() => setApiKey("password")}
            >
              set Api key to password
            </button>
            <button className="btn-error btn" onClick={() => setApiKey(null)}>
              reset api key
            </button>
          </div>
        </div>
        <div>
          <h1 className="font-bold">Version</h1>
          {isVersionLoading ? "Loading version...." : version?.api}
        </div>
        <div>
          <h1 className="font-bold">namespaces</h1>
          {isLoadingNamespaces
            ? "Loading namespaces"
            : namespaces?.results.map((namespace) => (
                <div key={namespace.name}>{namespace.name}</div>
              ))}
        </div>
        <div>
          <Link to="/">Start</Link> <Link to="/about">About</Link>
        </div>
        <div>
          <Outlet />
        </div>
      </div>

      <Sidebar version={version?.api ?? ""}>
        <SidebarTop /> {/* add drawer here */}
        <SidebarMain />
      </Sidebar>
      <Main>
        <MainTop>
          <MainTopLeft>1</MainTopLeft>
          <MainTopRight>2</MainTopRight>
        </MainTop>
        <MainContent>121</MainContent>
      </Main>
    </Root>
  );
};

export default Layout;
