import "./style.css";

import { Link, matchPath, useLocation } from "react-router-dom";
import React, { useCallback, useState } from "react";
import {
  VscAdd,
  VscCloudUpload,
  VscFolderOpened,
  VscGraph,
  VscLayers,
  VscPlayCircle,
  VscSettingsGear,
  VscSymbolEvent,
  VscVmRunning,
} from "react-icons/vsc";

import FlexBox from "../flexbox";
import HelpIcon from "../help";
import HideShowButton from "../hide-show";
import Logo from "../../assets/nav-logo.png";
import Modal from "../modal";
import NamespaceSelector from "../namespace-selector";
import Tabs from "../tabs";
import Tippy from "@tippyjs/react";
import { mirrorSettingInfoMetaInfo } from "../../layouts/mirror/info";
import { useDropzone } from "react-dropzone";

function NavBar(props) {
  let { className } = props;
  const {
    onClick,
    style,
    footer,
    createNamespace,
    createMirrorNamespace,
    namespace,
    namespaces,
    createErr,
    toggleResponsive,
    setToggleResponsive,
    extraNavigation,
  } = props;

  if (!className) {
    className = "";
  }

  className = "navigation-master " + className;

  if (!namespace) {
    className += " loading";
  }

  if (toggleResponsive) {
    className += " toggled";
  }

  const { pathname } = useLocation();

  return (
    <>
      <ResponsiveNavbar
        toggled={toggleResponsive}
        setToggled={setToggleResponsive}
      />
      <FlexBox onClick={onClick} style={{ ...style }} className={className}>
        <FlexBox col tall style={{ gap: "12px" }}>
          <FlexBox className="navbar-logo">
            <img alt="logo" src={Logo} />
          </FlexBox>
          <div className="navbar-panel col shadow">
            <FlexBox>
              <NamespaceSelector
                pathname={pathname}
                toggleResponsive={setToggleResponsive}
                namespace={namespace}
                namespaces={namespaces}
              />
            </FlexBox>
            <FlexBox>
              <NewNamespaceBtn
                createErr={createErr}
                createNamespace={createNamespace}
                createMirrorNamespace={createMirrorNamespace}
              />
            </FlexBox>
            <NavItems
              extraNavigation={extraNavigation}
              pathname={pathname}
              toggleResponsive={setToggleResponsive}
              namespace={namespace}
              style={{ marginTop: "12px" }}
            />
          </div>

          <div className="navbar-panel col shadow">
            <GlobalNavItems namespace={namespace} />
          </div>

          {footer}
        </FlexBox>
      </FlexBox>
    </>
  );
}

export default NavBar;

export function ClientFileUpload(props) {
  const { setError, setFile, maxSize } = props;

  const onDropAccepted = useCallback(
    (acceptedFiles) => {
      acceptedFiles.forEach((file) => {
        const reader = new FileReader();

        reader.onabort = () => {
          setError("Failed to load file.");
        };
        reader.onerror = () => {
          setError("Failed to read file.");
        };
        reader.onload = () => {
          try {
            const binaryStr = reader.result;
            var enc = new TextDecoder("utf-8");
            const fileData = enc.decode(binaryStr);
            setFile(fileData);
          } catch (e) {
            setError("Failed to decode file: " + e.message);
          }
        };
        reader.readAsArrayBuffer(file);
      });
    },
    [setError, setFile]
  );

  const onDropRejected = useCallback(
    (rejectedFiles) => {
      setError(
        "Failed to upload file: " + rejectedFiles?.[0]?.errors?.[0]?.message
      );
    },
    [setError]
  );
  const { getRootProps, getInputProps } = useDropzone({
    onDropAccepted,
    onDropRejected,
    multiple: false,
    maxSize: maxSize,
  });

  return (
    <div {...getRootProps()} className="input-title-button">
      <input {...getInputProps()} />
      {props.children}
    </div>
  );
}

function NewNamespaceBtn(props) {
  const { createNamespace, createMirrorNamespace } = props;

  const [tabIndex, setTabIndex] = useState(0);

  const [showPassphrase, setShowPassphrase] = useState(false);
  const [mirrorAuthMethod, setMirrorAuthMethod] = useState("none");
  const [mirrorSettings, setMirrorSettings] = useState({
    url: "",
    ref: "",
    passphrase: "",
    token: "",
    publicKey: "",
    privateKey: "",
  });

  const [mirrorErrors, setMirrorErrors] = useState({
    publicKey: null,
    privateKey: null,
  });

  const [ns, setNs] = useState("");

  return (
    <Modal
      title="New namespace"
      escapeToCancel
      modalStyle={{ width: "340px" }}
      button={
        <FlexBox className="gap center">
          <VscAdd />
          <span style={{ fontSize: "15px", fontWeight: "normal" }}>
            New namespace
          </span>
        </FlexBox>
      }
      buttonProps={{
        auto: true,
        size: "medium",
        fontSize: "large",
      }}
      titleIcon={<VscAdd />}
      onClose={() => {
        setNs("");
        setMirrorSettings({
          url: "",
          ref: "",
          passphrase: "",
          token: "",
          publicKey: "",
          privateKey: "",
        });
        setShowPassphrase(false);
        setMirrorAuthMethod("none");
        setTabIndex(0);
      }}
      keyDownActions={[
        {
          code: "Enter",

          fn: async () => {
            await createNamespace(ns);
            localStorage.setItem("namespace", ns);
            window.location = `/n/${ns}`;
            setNs("");
          },
          closeModal: true,
        },
      ]}
      actionButtons={[
        {
          label: "Add",

          onClick: async () => {
            if (tabIndex === 0) {
              await createNamespace(ns);
            } else {
              const processesMirrorSettings = JSON.parse(
                JSON.stringify(mirrorSettings)
              );
              if (mirrorAuthMethod === "token") {
                processesMirrorSettings["passphrase"] =
                  processesMirrorSettings["token"];
                processesMirrorSettings["privateKey"] = "";
                processesMirrorSettings["publicKey"] = "";
              } else if (mirrorAuthMethod === "none") {
                processesMirrorSettings["passphrase"] = "";
                processesMirrorSettings["privateKey"] = "";
                processesMirrorSettings["publicKey"] = "";
              }

              delete processesMirrorSettings["token"];
              await createMirrorNamespace(ns, processesMirrorSettings);
            }
            setTimeout(() => {
              localStorage.setItem("namespace", ns);
              window.location = `/n/${ns}`;
            }, 200);
            setNs("");
          },

          buttonProps: { variant: "contained", color: "primary" },
          closesModal: true,
          validate: true,
        },
        {
          label: "Cancel",
          onClick: () => {
            setNs("");
          },
          closesModal: true,
        },
      ]}
      requiredFields={[
        { tip: "namespace is required", value: ns },
        {
          tip: "mirror url is required",
          value: tabIndex === 0 ? true : mirrorSettings.url,
        },
        {
          tip: "mirror ref is required",
          value: tabIndex === 0 ? true : mirrorSettings.ref,
        },
        {
          tip: "public key is required",
          value:
            tabIndex === 0 ||
            mirrorAuthMethod === "none" ||
            mirrorAuthMethod === "token"
              ? true
              : mirrorSettings.publicKey,
        },
        {
          tip: "private key is required",
          value:
            tabIndex === 0 ||
            mirrorAuthMethod === "none" ||
            mirrorAuthMethod === "token"
              ? true
              : mirrorSettings.privateKey,
        },
        {
          tip: "token is required",
          value:
            tabIndex === 0 ||
            mirrorAuthMethod === "none" ||
            mirrorAuthMethod === "ssh"
              ? true
              : mirrorSettings.token,
        },
      ]}
    >
      <Tabs
        // TODO: change wf-execute-input => tabbed-form
        id="wf-execute-input"
        key="inputForm-ns"
        callback={setTabIndex}
        tabIndex={tabIndex}
        style={{ minWidth: "300px" }}
        headers={["Standard", "Mirror"]}
        tabs={[
          <FlexBox
            key="form-new-ns"
            className="col gap-md"
            style={{ paddingRight: "12px" }}
          >
            <FlexBox
              key="label-new-ns-name"
              className="row gap-sm"
              style={{ justifyContent: "flex-start" }}
            >
              <span className="input-title">Namespace*</span>
            </FlexBox>
            <input
              key="input-new-ns-name-input"
              autoFocus
              value={ns}
              onChange={(e) => setNs(e.target.value)}
              placeholder="Enter namespace name"
            />
          </FlexBox>,
          <FlexBox key="form-new-ns-mirror" className="col gap">
            <FlexBox
              key="input-new-ns-name"
              className="col gap-md"
              style={{ paddingRight: "12px" }}
            >
              <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                <span className="input-title">Namespace*</span>
              </FlexBox>
              <input
                autoFocus
                value={ns}
                onChange={(e) => setNs(e.target.value)}
                placeholder="Enter namespace name"
              />
            </FlexBox>
            <FlexBox key="input-new-ns-auth" className="col gap-md">
              <FlexBox row gap="sm" style={{ justifyContent: "flex-start" }}>
                <span className="input-title">Authentication Method</span>
              </FlexBox>
              <div
                style={{ width: "100%", paddingRight: "12px", display: "flex" }}
              >
                <select
                  style={{ width: "100%" }}
                  defaultValue={mirrorAuthMethod}
                  value={mirrorAuthMethod}
                  onChange={(e) => {
                    setMirrorAuthMethod(e.target.value);
                  }}
                >
                  <option value="none">None</option>
                  <option value="ssh">SSH Keys</option>
                  <option value="token">Access Token</option>
                </select>
              </div>
            </FlexBox>
            {Object.entries(mirrorSettings).map(([key, value]) => {
              if (
                (mirrorAuthMethod === "token" || mirrorAuthMethod === "none") &&
                (key === "publicKey" ||
                  key === "privateKey" ||
                  key === "passphrase")
              ) {
                return null;
              }

              if (
                (mirrorAuthMethod === "ssh" || mirrorAuthMethod === "none") &&
                key === "token"
              ) {
                return null;
              }

              return (
                <FlexBox
                  key={`input-new-ns-${key}`}
                  className="col gap-sm"
                  style={{ paddingRight: "12px" }}
                >
                  <FlexBox row style={{ justifyContent: "space-between" }}>
                    <FlexBox
                      row
                      gap="sm"
                      style={{ justifyContent: "flex-start" }}
                    >
                      <span className="input-title">
                        {mirrorSettingInfoMetaInfo[key].plainName}
                        {mirrorSettingInfoMetaInfo[key].required ? "*" : ""}
                      </span>
                      {mirrorSettingInfoMetaInfo[key].info ? (
                        <>
                          <HelpIcon
                            msg={mirrorSettingInfoMetaInfo[key].info}
                            zIndex={9999}
                          />
                          {key === "passphrase" ? (
                            <HideShowButton
                              show={showPassphrase}
                              setShow={setShowPassphrase}
                              field="Passphrase"
                            />
                          ) : null}
                        </>
                      ) : null}
                    </FlexBox>
                    {key === "publicKey" || key === "privateKey" ? (
                      <ClientFileUpload
                        setFile={(fileData) => {
                          const newSettings = mirrorSettings;
                          newSettings[key] = fileData;
                          setMirrorSettings({ ...newSettings });
                        }}
                        setError={(errorMsg) => {
                          const newErrors = mirrorErrors;
                          newErrors[key] = errorMsg;
                          setMirrorErrors({ ...newErrors });
                        }}
                        maxSize={40960}
                      >
                        <Tippy
                          content={
                            mirrorErrors[key]
                              ? mirrorErrors[key]
                              : `Upload key plaintext file content to ${mirrorSettingInfoMetaInfo[key].plainName} input. Warning: this will replace current ${mirrorSettingInfoMetaInfo[key].plainName} content`
                          }
                          trigger="click mouseenter focus"
                          onHide={() => {
                            const newErrors = mirrorErrors;
                            newErrors[key] = null;
                            setMirrorErrors({ ...newErrors });
                          }}
                          zIndex={9999}
                        >
                          <div className="input-title-button">
                            <FlexBox
                              center
                              row
                              gap="sm"
                              style={{
                                justifyContent: "flex-end",
                                marginRight: "-6px",
                              }}
                            >
                              <span>Upload</span>
                              <VscCloudUpload />
                            </FlexBox>
                          </div>
                        </Tippy>
                      </ClientFileUpload>
                    ) : null}
                  </FlexBox>
                  {key === "publicKey" ||
                  key === "privateKey" ||
                  key === "token" ? (
                    <textarea
                      style={{ width: "100%", resize: "none" }}
                      rows={5}
                      value={value}
                      spellCheck="false"
                      onChange={(e) => {
                        const newSettings = mirrorSettings;
                        newSettings[key] = e.target.value;
                        setMirrorSettings({ ...newSettings });
                      }}
                      autoFocus
                      placeholder={mirrorSettingInfoMetaInfo[key].placeholder}
                    />
                  ) : (
                    <input
                      type={
                        key === "passphrase" && !showPassphrase
                          ? "password"
                          : "text"
                      }
                      style={{ width: "100%" }}
                      value={value}
                      spellCheck="false"
                      onChange={(e) => {
                        const newSettings = mirrorSettings;
                        newSettings[key] = e.target.value;
                        setMirrorSettings({ ...newSettings });
                      }}
                      autoFocus
                      placeholder={mirrorSettingInfoMetaInfo[key].placeholder}
                    />
                  )}
                </FlexBox>
              );
            })}
          </FlexBox>,
        ]}
      />
    </Modal>
  );
}

function NavItems(props) {
  const { pathname, style, namespace, toggleResponsive, extraNavigation } =
    props;

  const explorer = matchPath("/n/:namespace", pathname);
  const wfexplorer = matchPath("/n/:namespace/explorer/*", pathname);
  const monitoring = matchPath("/n/:namespace/monitoring", pathname);
  // let builder = matchPath("/n/:namespace/builder", pathname)
  const events = matchPath("/n/:namespace/events", pathname);

  // instance path matching
  const instances = matchPath("/n/:namespace/instances", pathname);
  const instanceid = matchPath("/n/:namespace/instances/:id", pathname);

  const navItemMap = {};
  if (namespace !== null && namespace !== "") {
    if (extraNavigation) {
      for (let i = 0; i < extraNavigation.length; i++) {
        navItemMap[extraNavigation[i].path(namespace)] = matchPath(
          {
            path: extraNavigation[i].path(namespace),
            end: false,
          },
          pathname
        );
      }
    }
  }
  // let permissions = matchPath("/n/:namespace/permissions", pathname)

  // services pathname matching
  const services = matchPath("/n/:namespace/services", pathname);
  const service = matchPath("/n/:namespace/services/:service", pathname);
  const revision = matchPath(
    "/n/:namespace/services/:service/:revision",
    pathname
  );

  const settings = matchPath("/n/:namespace/settings", pathname);

  return (
    <FlexBox style={{ ...style }} className="nav-items">
      <ul>
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem
              className={explorer || wfexplorer ? "active" : ""}
              label="Explorer"
            >
              <VscFolderOpened />
            </NavItem>
          </Link>
        </li>
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}/monitoring`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem className={monitoring ? "active" : ""} label="Monitoring">
              <VscGraph />
            </NavItem>
          </Link>
        </li>
        {/* <li>
                    <Link to={`/n/${namespace}/builder`}>
                        <NavItem className={builder ? "active":""} label="Workflow Builder">
                            <IoGitNetworkOutline/>
                        </NavItem>
                    </Link>
                </li> */}
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}/instances`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem
              className={instances || instanceid ? "active" : ""}
              label="Instances"
            >
              <VscVmRunning />
            </NavItem>
          </Link>
        </li>
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}/events`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem className={events ? "active" : ""} label="Events">
              <VscSymbolEvent />
            </NavItem>
          </Link>
        </li>
        {namespace !== null && namespace !== ""
          ? extraNavigation?.map((obj) => {
              if (obj.hreflink) {
                return (
                  <li key={obj.title}>
                    <a href={obj.path(namespace)}>
                      <NavItem
                        className={
                          navItemMap[obj.path(namespace)] !== null
                            ? "active"
                            : ""
                        }
                        label={obj.title}
                      >
                        {obj.icon}
                      </NavItem>
                    </a>
                  </li>
                );
              } else {
                return (
                  <li key={obj.title}>
                    <Link
                      to={obj.path(namespace)}
                      onClick={() => {
                        toggleResponsive(false);
                      }}
                    >
                      <NavItem
                        className={
                          navItemMap[obj.path(namespace)] !== null
                            ? "active"
                            : ""
                        }
                        label={obj.title}
                      >
                        {obj.icon}
                      </NavItem>
                    </Link>
                  </li>
                );
              }
            })
          : ""}
        {/* <li>
                    <Link to={`/n/${namespace}/permissions`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={permissions ? "active":""} label="Permissions">
                            <VscLock/>
                        </NavItem>
                    </Link>
                </li> */}
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}/services`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem
              className={services || service || revision ? "active" : ""}
              label="Services"
            >
              <VscLayers />
            </NavItem>
          </Link>
        </li>
        <li className={`${!namespace ? "disabled-nav-item" : ""}`}>
          <Link
            to={!!namespace && `/n/${namespace}/settings`}
            onClick={() => {
              !!namespace && toggleResponsive(false);
            }}
          >
            <NavItem className={settings ? "active" : ""} label="Settings">
              <VscSettingsGear />
            </NavItem>
          </Link>
        </li>
      </ul>
    </FlexBox>
  );
}

function GlobalNavItems({ namespace }) {
  const { pathname } = useLocation();

  const jq = matchPath("/jq", pathname);

  return (
    <FlexBox className="nav-items">
      <ul>
        <li
          className={`${!namespace ? "disabled-nav-item" : ""}`}
          style={{ marginTop: "0px" }}
        >
          <Link to={!!namespace && "/jq"}>
            <NavItem className={jq ? "active" : ""} label="jq Playground">
              <VscPlayCircle />
            </NavItem>
          </Link>
        </li>
      </ul>
    </FlexBox>
  );
}

export function NavItem(props) {
  let { className } = props;
  const { children, label } = props;
  if (!className) {
    className = "";
  }

  return (
    <FlexBox className={"nav-item " + className} style={{ gap: "8px" }}>
      <FlexBox style={{ maxWidth: "30px", width: "30px", margin: "auto" }}>
        {children}
      </FlexBox>
      <FlexBox style={{ textAlign: "left" }}>{label}</FlexBox>
    </FlexBox>
  );
}

function ResponsiveNavbar(props) {
  const { toggled, setToggled } = props;
  let panelClasses = "panel";
  let responsiveNavClasses = "responsive-nav hide-on-large";
  let responsiveNavOverlayClasses = "responsive-nav-overlay hide-on-large";

  if (toggled) {
    panelClasses += " toggled";
    responsiveNavClasses += " toggled";
    responsiveNavOverlayClasses += " toggled";
  } else {
    panelClasses += " disabled";
    responsiveNavClasses += " disabled";
    responsiveNavOverlayClasses += " disabled";
  }

  return (
    <>
      <div className={responsiveNavOverlayClasses}></div>
      <FlexBox
        id="clickme"
        className={responsiveNavClasses}
        onClick={(e) => {
          setToggled(false);
          e.stopPropagation();
        }}
      >
        <div className={panelClasses}></div>
      </FlexBox>
    </>
  );
}
