/* eslint-disable tailwindcss/no-custom-classname */
import "./AppLegacy.css";
import "./util/responsive.css";

import { useEffect, useState } from "react";

import { ApiKeyProvider } from "./util/apiKeyProvider";
import { Config } from "./util";
import FlexBox from "./components/flexbox";
import Login from "./layouts/login";
import MainLayout from "./layouts/main";
import { NavItem } from "./components/navbar";
import { ThemeProvider } from "@mui/material/styles";
import { VscSignOut } from "react-icons/vsc";
import theme from "./theme/style";

function App() {
  const [version, setVersion] = useState("");
  const [loadVersion, setLoadVersion] = useState(true);
  const [akey, setAKey] = useState(localStorage.getItem("apikey"));
  const [login, setLogin] = useState(false);
  const [akeyReq, setAKeyReq] = useState(false);
  // Todo find nice way to handle error
  const [, setErr] = useState("");

  useEffect(() => {
    async function fetchVersion() {
      try {
        const resp = await fetch(`${Config.url}version`, {
          method: "GET",
          headers: {
            "direktiv-token": akey,
          },
        });
        const respNoKey = await fetch(`${Config.url}version`, {
          method: "GET",
        });
        if (resp.ok) {
          const json = await resp.json();
          setLoadVersion(false);
          setVersion(json.api);

          // TODO if the akey is provided but not needed as authentication isn't required.
          // Might need to make an api to check if apikeys are required.
          if (akey !== "null" && respNoKey.status === 401) {
            setAKeyReq(true);
          }
        } else {
          if (resp.status === 401) {
            setLogin(true);
            setAKeyReq(true);
          }
        }
      } catch (e) {
        setErr(e);
      }
    }
    if (loadVersion) {
      fetchVersion();
    }
  }, [version, loadVersion, akey]);

  const f = (
    <>
      <FlexBox>
        {akeyReq ? (
          <FlexBox className="nav-items" style={{ paddingLeft: "10px" }}>
            <ul style={{ marginTop: "0px" }}>
              <li
                onClick={() => {
                  localStorage.setItem("apikey", null);
                  window.location.reload();
                }}
              >
                <NavItem className="red-text" label="Log Out">
                  <VscSignOut />
                </NavItem>
              </li>
            </ul>
          </FlexBox>
        ) : null}
      </FlexBox>
      <div>
        <FlexBox className="col navbar-userinfo">
          <FlexBox className="navbar-version">
            <b style={{ marginRight: "8px" }}>Version:</b> {version}
          </FlexBox>
        </FlexBox>
      </div>
    </>
  );
  return (
    <ThemeProvider theme={theme}>
      <ApiKeyProvider>
        <div className="App">
          {login ? (
            <Login setLogin={setLogin} setAKey={setAKey} />
          ) : (
            <MainLayout
              akey={akey}
              akeyReq={akeyReq}
              footer={f}
              extraRoutes={[]}
              extraNavigation={[]}
            />
          )}
        </div>
      </ApiKeyProvider>
    </ThemeProvider>
  );
}

export default App;
