import "./style.css";

import * as dayjs from "dayjs";

import { Config, GenerateRandomKey } from "../../util";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../components/content-panel";
import Pagination, { usePageHandler } from "../../components/pagination";
import {
  VscAdd,
  VscClose,
  VscCloudUpload,
  VscCode,
  VscEdit,
  VscFolderOpened,
  VscRepo,
  VscSearch,
  VscTrash,
} from "react-icons/vsc";
import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router";

import Alert from "../../components/alert";
import { AutoSizer } from "react-virtualized";
import Button from "../../components/button";
import CircularProgress from "@mui/material/CircularProgress";
import { ClientFileUpload } from "../../components/navbar";
import DirektivEditor from "../../components/editor";
import Fade from "@mui/material/Fade";
import { FcWorkflow } from "react-icons/fc";
import { FiFolder } from "react-icons/fi";
import FlexBox from "../../components/flexbox";
import HelpIcon from "../../components/help";
import { HiOutlineTrash } from "react-icons/hi";
import HideShowButton from "../../components/hide-show";
import Loader from "../../components/loader";
import { MirrorReadOnlyBadge } from "../mirror";
import Modal from "../../components/modal";
import NotFound from "../notfound";
import Tabs from "../../components/tabs";
import Tippy from "@tippyjs/react";
import WorkflowPage from "./workflow";
import WorkflowPod from "./workflow/pod";
import WorkflowRevisions from "./workflow/revision";
import { mirrorSettingInfoMetaInfo } from "../mirror/info";
import relativeTime from "dayjs/plugin/relativeTime";
import { useApiKey } from "../../util/apiKeyProvider";
import { useNodes } from "../../hooks";
import { useSearchParams } from "react-router-dom";
import utc from "dayjs/plugin/utc";

const PAGE_SIZE = 10;
const apiHelps = (namespace) => {
  const url = window.location.origin;
  return [
    {
      method: "GET",
      url: `${url}/api/namespaces/${namespace}/tree`,
      description: `List nodes`,
    },
    {
      method: "PUT",
      description: `Create a directory`,
      url: `${url}/api/namespaces/${namespace}/tree/NODE_NAME?op=create-directory`,
      body: `{
  "type": "directory"
}`,
      type: "json",
    },
    {
      method: "PUT",
      description: `Create a workflow `,
      url: `${url}/api/namespaces/${namespace}/tree/NODE_NAME?op=create-workflow`,
      body: `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
type: noop
transform:
    result: Hello world!`,
      type: "yaml",
    },
    {
      method: "POST",
      description: "Rename a node",
      url: `${url}/api/namespaces/${namespace}/tree/NODE_NAME?op=rename-node`,
      body: `{
  "new": "NEW_NODE_NAME"
}`,
      type: "json",
    },
    {
      method: "DEL",
      description: `Delete a node`,
      url: `${url}/api/namespaces/${namespace}/tree/NODE_NAME?op=delete-node`,
    },
  ];
};

function Explorer(props) {
  const params = useParams();
  const [searchParams] = useSearchParams(); // removed 'setSearchParams' from square brackets (this should not affect anything: search 'destructuring assignment')

  const { namespace, setBreadcrumbChildren } = props;
  let filepath = `/`;
  if (!namespace) {
    return null;
  }
  if (params["*"] !== undefined) {
    filepath = `/${params["*"]}`;
  }

  // pod revisions
  if (
    searchParams.get("function") &&
    searchParams.get("version") &&
    searchParams.get("revision")
  ) {
    return (
      <WorkflowPod
        filepath={filepath}
        namespace={namespace}
        service={searchParams.get("function")}
        version={searchParams.get("version")}
        revision={searchParams.get("revision")}
      />
    );
  }
  // service revisions
  if (searchParams.get("function") && searchParams.get("version")) {
    return (
      <WorkflowRevisions
        filepath={filepath}
        namespace={namespace}
        service={searchParams.get("function")}
        version={searchParams.get("version")}
      />
    );
  }

  return (
    <>
      <ExplorerList
        namespace={namespace}
        path={filepath}
        setBreadcrumbChildren={setBreadcrumbChildren}
      />
    </>
  );
}

export default Explorer;

export function SearchBar(props) {
  const { setSearch, style } = props;

  const [inputVal, setInputVal] = useState("");

  useEffect(() => {
    const handler = setTimeout(() => {
      setSearch(inputVal);
    }, 200);

    return () => {
      clearTimeout(handler);
    };
  }, [inputVal, setSearch]);

  return (
    <div className="explorer-searchbar" style={{ height: "29px", ...style }}>
      <FlexBox style={{ height: "100%" }}>
        <VscSearch className="auto-margin" />
        <input
          value={inputVal}
          onChange={(e) => {
            setInputVal(e.target.value);
          }}
          placeholder="Search items"
          style={{ boxSizing: "border-box" }}
        ></input>
      </FlexBox>
    </div>
  );
}

dayjs.extend(utc);
dayjs.extend(relativeTime);

const orderFieldDictionary = {
  Name: "NAME", // Default
  Created: "CREATED",
};

const orderFieldKeys = Object.keys(orderFieldDictionary);

function ExplorerList(props) {
  const { namespace, path, setBreadcrumbChildren } = props;
  const setBreadcrumbChildrenRef = useRef(setBreadcrumbChildren);
  const [apiKey] = useApiKey();

  const navigate = useNavigate();

  //api helper modal
  const [search, setSearch] = useState("");

  const [currPath, setCurrPath] = useState("");

  const [name, setName] = useState("");
  const [load, setLoad] = useState(true);
  const [tabIndex, setTabIndex] = useState(0);
  const [isReadOnly, setIsReadOnly] = useState(false);

  const [orderFieldKey, setOrderFieldKey] = useState(orderFieldKeys[0]);

  const [streamNodes, setStreamNodes] = useState(false);
  const [queryParams, setQueryParams] = useState([]);

  const pageHandler = usePageHandler(PAGE_SIZE);
  const goToFirstPage = pageHandler.goToFirstPage;

  const {
    data,
    err,
    templates,
    pageInfo,
    createNode,
    createMirrorNode,
    deleteNode,
    renameNode,
  } = useNodes(
    Config.url,
    streamNodes,
    namespace,
    path,
    apiKey,
    pageHandler.pageParams,
    ...queryParams,
    `order.field=${orderFieldDictionary[orderFieldKey]}`,
    `filter.field=NAME`,
    `filter.val=${search}`,
    `filter.type=CONTAINS`
  );

  const [wfData, setWfData] = useState(templates["noop"].data);
  const [wfTemplate, setWfTemplate] = useState("noop");

  // Reset Page to start when filters changes
  useEffect(() => {
    // TODO: This will interfere with page position if initPage > 1
    goToFirstPage();
  }, [search, goToFirstPage, path]);

  // Mirror
  const [showPassphrase, setShowPassphrase] = useState(false);
  const [mirrorAuthMethod, setMirrorAuthMethod] = useState("none");
  const [mirrorSettings, setMirrorSettings] = useState({
    url: "",
    ref: "",
    cron: "",
    passphrase: "",
    token: "",
    publicKey: "",
    privateKey: "",
  });

  const [mirrorErrors, setMirrorErrors] = useState({
    publicKey: null,
    privateKey: null,
  });

  useEffect(() => {
    setStreamNodes(false);
  }, [path]);

  useEffect(() => {
    if (data === null || !streamNodes) {
      return;
    }

    if (data?.node?.type === "workflow") {
      setStreamNodes(false);
    }
  }, [data, streamNodes]);

  function resetQueryParams() {
    setQueryParams([]);
    setSearch("");
  }

  // control loading icon todo work out how to display this error
  useEffect(() => {
    if (data !== null || err !== null) {
      setLoad(false);
      setIsReadOnly(data?.node?.readOnly);
    }
  }, [data, err]);

  useEffect(() => {
    if (!setBreadcrumbChildrenRef.current) {
      return;
    }

    setBreadcrumbChildrenRef.current(
      isReadOnly ? (
        <FlexBox
          center
          row
          gap
          style={{ justifyContent: "flex-end", paddingRight: "6px" }}
        >
          <MirrorReadOnlyBadge />
        </FlexBox>
      ) : null
    );
  }, [isReadOnly]);

  // Keep Refs up to date
  useEffect(() => {
    setBreadcrumbChildrenRef.current = setBreadcrumbChildren;
  }, [setBreadcrumbChildren]);

  // Unmount cleanup breadcrumb children
  useEffect(() => {
    return () => {
      if (setBreadcrumbChildrenRef.current) {
        setBreadcrumbChildrenRef.current(null);
      }
    };
  }, []);

  // Reset pagination queries when searching
  useEffect(() => {
    setQueryParams([]);
  }, [search]);

  // Reset pagination and search when namespace changes
  useEffect(() => {
    resetQueryParams();
  }, [namespace]);

  useEffect(() => {
    if (path !== currPath) {
      setCurrPath(path);
      setLoad(true);
    }
  }, [path, currPath]);

  if (err === "Not Found") {
    return <NotFound />;
  }

  return (
    <>
      {data !== null && data?.node?.type === "workflow" ? (
        <WorkflowPage namespace={namespace} />
      ) : (
        <FlexBox col gap style={{ paddingRight: "8px" }}>
          <Loader load={load} timer={1000}>
            <FlexBox gap style={{ maxHeight: "32px" }}>
              <FlexBox>
                <div>
                  <Modal
                    titleIcon={<VscCode />}
                    button={
                      <>
                        <VscCode
                          style={{ maxHeight: "12px", marginRight: "4px" }}
                        />
                        API Commands
                      </>
                    }
                    escapeToCancel
                    withCloseButton
                    maximised
                    title="Namespace API Interactions"
                  >
                    {apiHelps(namespace).map((help) => (
                      <ApiFragment
                        key={`${help.type}-key`}
                        description={help.description}
                        url={help.url}
                        method={help.method}
                        body={help.body}
                        type={help.type}
                      />
                    ))}
                  </Modal>
                </div>
              </FlexBox>
              <FlexBox style={{ flexDirection: "row-reverse" }}>
                <SearchBar setSearch={setSearch} />
              </FlexBox>
            </FlexBox>
            <ContentPanel>
              <ContentPanelTitle>
                <ContentPanelTitleIcon>
                  <VscFolderOpened />
                </ContentPanelTitleIcon>
                <FlexBox style={{ display: "flex", alignItems: "center" }} gap>
                  <div>Explorer</div>
                  <HelpIcon msg="Directory/workflow browser." />
                </FlexBox>
                <FlexBox
                  className="gap center-y"
                  style={{ flexDirection: "row-reverse" }}
                >
                  <Modal
                    title="New Workflow"
                    modalStyle={{ width: "600px" }}
                    escapeToCancel
                    button={
                      <>
                        <VscAdd />
                        <span className="hide-600">Workflow</span>
                        <span className="show-600">WF</span>
                      </>
                    }
                    onClose={() => {
                      setWfData(templates["noop"].data);
                      setWfTemplate("noop");
                      setName("");
                    }}
                    actionButtons={[
                      {
                        label: "Add",

                        onClick: async () => {
                          const result = await createNode(
                            name,
                            "workflow",
                            wfData
                          );
                          if (result.node && result.namespace) {
                            navigate(
                              `/n/${
                                result.namespace
                              }/explorer/${result.node.path.substring(1)}`
                            );
                          }
                        },

                        buttonProps: {
                          variant: "contained",
                          color: "primary",
                        },
                        closesModal: true,
                        validate: true,
                      },
                      {
                        label: "Cancel",
                        closesModal: true,
                      },
                    ]}
                    keyDownActions={[
                      {
                        code: "Enter",

                        fn: async () => {
                          if (name.trim()) {
                            await createNode(name, "workflow", wfData);
                          } else {
                            throw new Error("Please fill in name");
                          }
                        },
                        closeModal: true,
                        id: "workflow-name",
                      },
                    ]}
                    requiredFields={[
                      { tip: "workflow name is required", value: name },
                      { tip: "workflow cannot be empty", value: wfData },
                    ]}
                  >
                    <FlexBox
                      col
                      gap
                      style={{
                        fontSize: "12px",
                        minHeight: "500px",
                        minWidth: "550px",
                      }}
                    >
                      <div
                        style={{
                          width: "100%",
                          paddingRight: "12px",
                          display: "flex",
                        }}
                      >
                        <input
                          id="workflow-name"
                          value={name}
                          onChange={(e) => setName(e.target.value)}
                          autoFocus
                          placeholder="Enter workflow name"
                        />
                      </div>
                      <select
                        value={wfTemplate}
                        onChange={(e) => {
                          setWfTemplate(e.target.value);
                          // todo set wfdata to template on change
                          setWfData(templates[e.target.value].data);
                        }}
                      >
                        {Object.keys(templates).map((obj) => {
                          const key = GenerateRandomKey("");
                          return (
                            <option key={key} value={obj}>
                              {templates[obj].name}
                            </option>
                          );
                        })}
                      </select>
                      <FlexBox gap>
                        <FlexBox style={{ overflow: "hidden" }}>
                          <AutoSizer>
                            {({ height, width }) => (
                              <DirektivEditor
                                dlang="yaml"
                                width={width}
                                value={wfData}
                                setDValue={setWfData}
                                height={height}
                              />
                            )}
                          </AutoSizer>
                        </FlexBox>
                      </FlexBox>
                    </FlexBox>
                  </Modal>
                  <Modal
                    title="New Directory"
                    modalStyle={{ width: "340px" }}
                    escapeToCancel
                    button={
                      <>
                        <VscAdd />
                        <span className="hide-600">Directory</span>
                        <span className="show-600">Dir</span>
                      </>
                    }
                    onClose={() => {
                      setName("");
                      setMirrorSettings({
                        url: "",
                        ref: "",
                        cron: "",
                        passphrase: "",
                        token: "",
                        publicKey: "",
                        privateKey: "",
                      });
                      setShowPassphrase(false);
                      setMirrorAuthMethod("none");
                      setTabIndex(0);
                    }}
                    actionButtons={[
                      {
                        label: "Add",

                        onClick: async () => {
                          if (tabIndex === 0) {
                            await createNode(name, "directory");
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
                            await createMirrorNode(
                              name,
                              processesMirrorSettings
                            );
                          }
                        },

                        buttonProps: {
                          variant: "contained",
                          color: "primary",
                          disabled: name.trim().length === 0,
                        },
                        closesModal: true,
                        validate: true,
                      },
                      {
                        label: "Cancel",
                        closesModal: true,
                      },
                    ]}
                    keyDownActions={[
                      {
                        code: "Enter",

                        fn: async () => {
                          if (tabIndex === 0) {
                            await createNode(name, "directory");
                          } else {
                            await createMirrorNode(name, mirrorSettings);
                          }
                        },
                        closeModal: true,
                      },
                    ]}
                    requiredFields={[
                      { tip: "directory name is required", value: name },
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
                      key="inputForm"
                      callback={setTabIndex}
                      tabIndex={tabIndex}
                      style={{ minWidth: "300px" }}
                      headers={["Standard", "Mirror"]}
                      tabs={[
                        // no key needed for static array
                        // eslint-disable-next-line react/jsx-key
                        <FlexBox col gap="sm" style={{ paddingRight: "12px" }}>
                          <FlexBox
                            row
                            gap="sm"
                            style={{ justifyContent: "flex-start" }}
                          >
                            <span className="input-title">Directory*</span>
                          </FlexBox>
                          <input
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            autoFocus
                            placeholder="Enter a directory name"
                          />
                        </FlexBox>,
                        // no key needed for static array
                        // eslint-disable-next-line react/jsx-key
                        <FlexBox col gap>
                          <FlexBox
                            col
                            gap="sm"
                            style={{ paddingRight: "12px" }}
                          >
                            <FlexBox
                              row
                              gap="sm"
                              style={{ justifyContent: "flex-start" }}
                            >
                              <span className="input-title">Directory*</span>
                            </FlexBox>
                            <input
                              autoFocus
                              value={name}
                              onChange={(e) => setName(e.target.value)}
                              placeholder="Enter a directory name"
                            />
                          </FlexBox>
                          <FlexBox col gap="sm">
                            <FlexBox
                              row
                              gap="sm"
                              style={{ justifyContent: "flex-start" }}
                            >
                              <span className="input-title">
                                Authentication Method
                              </span>
                            </FlexBox>
                            <div
                              style={{
                                width: "100%",
                                paddingRight: "12px",
                                display: "flex",
                              }}
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
                          {Object.entries(mirrorSettings).map(
                            ([key, value]) => {
                              if (
                                (mirrorAuthMethod === "token" ||
                                  mirrorAuthMethod === "none") &&
                                (key === "publicKey" ||
                                  key === "privateKey" ||
                                  key === "passphrase")
                              ) {
                                return null;
                              }

                              if (
                                (mirrorAuthMethod === "ssh" ||
                                  mirrorAuthMethod === "none") &&
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
                                  <FlexBox
                                    row
                                    style={{
                                      justifyContent: "space-between",
                                    }}
                                  >
                                    <FlexBox
                                      row
                                      gap="sm"
                                      style={{ justifyContent: "flex-start" }}
                                    >
                                      <span className="input-title">
                                        {
                                          mirrorSettingInfoMetaInfo[key]
                                            .plainName
                                        }
                                        {mirrorSettingInfoMetaInfo[key].required
                                          ? "*"
                                          : ""}
                                      </span>
                                      {mirrorSettingInfoMetaInfo[key].info ? (
                                        <>
                                          <HelpIcon
                                            msg={
                                              mirrorSettingInfoMetaInfo[key]
                                                .info
                                            }
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
                                    {key === "publicKey" ||
                                    key === "privateKey" ? (
                                      <ClientFileUpload
                                        setFile={(fileData) => {
                                          const newSettings = mirrorSettings;
                                          newSettings[key] = fileData;
                                          setMirrorSettings({
                                            ...newSettings,
                                          });
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
                                          zIndex={10}
                                        >
                                          <div className="input-title-button">
                                            <FlexBox
                                              center
                                              row
                                              gap="sm"
                                              style={{
                                                justifyContent: "flex-end",
                                                marginRight: "-6px",
                                                fontWeight: "normal",
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
                                      style={{
                                        width: "100%",
                                        resize: "none",
                                      }}
                                      rows={5}
                                      value={value}
                                      onChange={(e) => {
                                        const newSettings = mirrorSettings;
                                        newSettings[key] = e.target.value;
                                        setMirrorSettings({ ...newSettings });
                                      }}
                                      autoFocus
                                      placeholder={
                                        mirrorSettingInfoMetaInfo[key]
                                          .placeholder
                                      }
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
                                      onChange={(e) => {
                                        const newSettings = mirrorSettings;
                                        newSettings[key] = e.target.value;
                                        setMirrorSettings({ ...newSettings });
                                      }}
                                      autoFocus
                                      placeholder={
                                        mirrorSettingInfoMetaInfo[key]
                                          .placeholder
                                      }
                                    />
                                  )}
                                </FlexBox>
                              );
                            }
                          )}
                        </FlexBox>,
                      ]}
                    />
                  </Modal>
                  {data && data?.node?.expandedType === "git" ? (
                    <Button
                      variant="outlined"
                      color="info"
                      href={`/n/${namespace}/mirror${path}`}
                    >
                      <VscRepo />
                      <span>Mirror Info</span>
                    </Button>
                  ) : null}
                  <div className="explorer-sort-by explorer-action-btn hide-600">
                    <FlexBox gap style={{ marginRight: "8px" }}>
                      <FlexBox center>Sort by:</FlexBox>
                      <FlexBox center>
                        <select
                          onChange={(e) => {
                            setOrderFieldKey(e.target.value);
                            setQueryParams([]);
                          }}
                          value={orderFieldKey}
                          className="dropdown-select"
                          style={{
                            paddingBottom: "0px",
                            paddingTop: "0px",
                            height: "27px",
                          }}
                        >
                          <option value="">{orderFieldKey}</option>
                          {orderFieldKeys.map((key) => {
                            if (key === orderFieldKey) {
                              return "";
                            }
                            return (
                              <option key={GenerateRandomKey()} value={key}>
                                {key}
                              </option>
                            );
                          })}
                        </select>
                      </FlexBox>
                    </FlexBox>
                  </div>
                </FlexBox>
              </ContentPanelTitle>
              <ContentPanelBody style={{ height: "100%" }}>
                <FlexBox col>
                  {data === null || pageInfo === null ? (
                    <div className="explorer-item">
                      <FlexBox className="explorer-item-container">
                        {err === "permission denied" ? (
                          <Alert severity="warning" variant="standard" grow>
                            Unfortunately, you do not have the necessary
                            permissions
                          </Alert>
                        ) : (
                          <>
                            <FlexBox
                              style={{ display: "flex", alignItems: "center" }}
                              className="explorer-item-icon"
                            >
                              <Fade
                                in={data === null || pageInfo === null}
                                style={{
                                  transitionDelay:
                                    data === null || pageInfo === null
                                      ? "200ms"
                                      : "0ms",
                                }}
                                unmountOnExit
                              >
                                <CircularProgress size="1rem" color="primary" />
                              </Fade>
                            </FlexBox>
                            <FlexBox
                              style={{ fontSize: "10pt" }}
                              className="explorer-item-name"
                            >
                              Loading results for &apos;{path}&apos;.
                            </FlexBox>
                          </>
                        )}
                      </FlexBox>
                    </div>
                  ) : (
                    <>
                      {data.children.results.length === 0 ? (
                        <div className="explorer-item">
                          <FlexBox className="explorer-item-container">
                            <FlexBox
                              style={{
                                display: "flex",
                                alignItems: "center",
                              }}
                              className="explorer-item-icon"
                            >
                              <VscSearch />
                            </FlexBox>
                            <FlexBox
                              style={{ fontSize: "10pt" }}
                              className="explorer-item-name"
                            >
                              No results found under &apos;{path}&apos;.
                            </FlexBox>
                          </FlexBox>
                        </div>
                      ) : (
                        <>
                          {data.children.results.map((obj) => {
                            if (obj.type === "directory") {
                              return (
                                <DirListItem
                                  isGit={data && obj.expandedType === "git"}
                                  namespace={namespace}
                                  renameNode={renameNode}
                                  deleteNode={deleteNode}
                                  path={obj.path}
                                  key={GenerateRandomKey("explorer-item-")}
                                  name={obj.name}
                                  resetQueryParams={resetQueryParams}
                                />
                              );
                            } else if (obj.type === "workflow") {
                              return (
                                <WorkflowListItem
                                  namespace={namespace}
                                  renameNode={renameNode}
                                  deleteNode={deleteNode}
                                  path={obj.path}
                                  key={GenerateRandomKey("explorer-item-")}
                                  name={obj.name}
                                />
                              );
                            }
                            return null;
                          })}
                        </>
                      )}
                    </>
                  )}
                </FlexBox>
              </ContentPanelBody>
            </ContentPanel>
            <FlexBox
              row
              style={{
                justifyContent: "flex-end",
                paddingBottom: "1em",
                flexGrow: 0,
              }}
            >
              <Pagination pageHandler={pageHandler} pageInfo={pageInfo} />
            </FlexBox>
          </Loader>
        </FlexBox>
      )}
    </>
  );
}

function DirListItem(props) {
  const {
    name,
    path,
    deleteNode,
    renameNode,
    namespace,
    resetQueryParams,
    className,
    isGit,
  } = props;

  const navigate = useNavigate();
  const [renameValue, setRenameValue] = useState(path);
  const [rename, setRename] = useState(false);
  const [err, setErr] = useState("");
  const [recursiveDelete, setRecursiveDelete] = useState(false);

  return (
    <div
      style={{ cursor: "pointer" }}
      onClick={() => {
        resetQueryParams();
        navigate(`/n/${namespace}/explorer/${path.substring(1)}`);
      }}
      className="explorer-item"
    >
      <FlexBox col>
        <FlexBox className="explorer-item-container gap wrap">
          <FlexBox className="explorer-item-icon">
            {isGit ? (
              <VscRepo className="auto-margin" />
            ) : (
              <FiFolder className="auto-margin" />
            )}
          </FlexBox>
          {rename ? (
            <FlexBox
              className="explorer-item-name"
              style={{ alignItems: "center" }}
            >
              <input
                style={{ width: "100%", height: "38px" }}
                onClick={(ev) => ev.stopPropagation()}
                type="text"
                value={renameValue}
                onKeyPress={async (e) => {
                  if (e.key === "Enter") {
                    try {
                      await renameNode("/", path, renameValue);
                      setRename(!rename);
                    } catch (err) {
                      setErr(err.message);
                    }
                  }
                }}
                onChange={(e) => setRenameValue(e.target.value)}
                autoFocus
              />
            </FlexBox>
          ) : (
            <FlexBox
              className={`explorer-item-name ${className ? className : ""}`}
            >
              {name}
            </FlexBox>
          )}
          <FlexBox>
            {err !== "" ? (
              <FlexBox>
                <Alert severity="error" variant="filled">
                  {err}
                </Alert>
              </FlexBox>
            ) : (
              <FlexBox />
            )}
            <FlexBox className="explorer-item-actions gap">
              {rename ? (
                <FlexBox
                  onClick={(ev) => {
                    setRename(!rename);
                    setErr("");
                    ev.stopPropagation();
                  }}
                >
                  <VscClose className="auto-margin" />
                </FlexBox>
              ) : (
                <FlexBox
                  onClick={(ev) => {
                    setRename(!rename);
                    setErr("");
                    ev.stopPropagation();
                  }}
                >
                  <VscEdit className="auto-margin" />
                </FlexBox>
              )}
              <FlexBox onClick={(ev) => ev.stopPropagation()}>
                <Modal
                  escapeToCancel
                  modalStyle={{ width: "240px" }}
                  style={{
                    flexDirection: "row-reverse",
                  }}
                  title="Delete a directory"
                  button={
                    <>
                      <VscTrash className="auto-margin red-text" />
                    </>
                  }
                  buttonProps={{
                    auto: true,
                    variant: "text",
                  }}
                  actionButtons={[
                    {
                      label: "Delete",

                      onClick: async () => {
                        const p = path.split("/", -1);
                        const pLast = p[p.length - 1];
                        await deleteNode(pLast, recursiveDelete);
                      },
                      buttonProps: { variant: "contained", color: "error" },
                      closesModal: true,
                    },
                    {
                      label: "Cancel",
                      closesModal: true,
                    },
                  ]}
                >
                  <FlexBox col gap>
                    <FlexBox gap center="y" style={{ fontWeight: "bold" }}>
                      Recursive Delete:
                      <label className="switch">
                        <input
                          onChange={() => {
                            setRecursiveDelete(!recursiveDelete);
                          }}
                          type="checkbox"
                          checked={recursiveDelete}
                        />
                        <span className="slider-broadcast"></span>
                      </label>
                    </FlexBox>
                    <FlexBox>
                      Are you sure you want to delete &apos;{name}&apos;?
                      <br />
                      This action cannot be undone.
                    </FlexBox>
                  </FlexBox>
                </Modal>
              </FlexBox>
            </FlexBox>
          </FlexBox>
        </FlexBox>
      </FlexBox>
    </div>
  );
}

function WorkflowListItem(props) {
  const { name, path, deleteNode, renameNode, namespace } = props;

  const navigate = useNavigate();
  const [renameValue, setRenameValue] = useState(path);
  const [rename, setRename] = useState(false);
  const [err, setErr] = useState("");

  return (
    <div
      style={{ cursor: "pointer" }}
      onClick={() => {
        navigate(`/n/${namespace}/explorer/${path.substring(1)}`);
      }}
      className="explorer-item"
    >
      <FlexBox col>
        <FlexBox className="explorer-item-container gap wrap">
          <FlexBox className="explorer-item-icon">
            <FcWorkflow className="auto-margin" />
          </FlexBox>
          {rename ? (
            <FlexBox
              className="explorer-item-name"
              style={{
                alignItems: "center",
                maxWidth: "300px",
                minWidth: "300px",
              }}
            >
              <input
                onClick={(ev) => ev.stopPropagation()}
                type="text"
                value={renameValue}
                onKeyPress={async (e) => {
                  if (e.key === "Enter") {
                    try {
                      await renameNode("/", path, renameValue);
                      setRename(!rename);
                    } catch (err) {
                      setErr(err.message);
                    }
                  }
                }}
                onChange={(e) => setRenameValue(e.target.value)}
                autoFocus
                style={{ maxWidth: "300px", height: "38px" }}
              />
            </FlexBox>
          ) : (
            <FlexBox className="explorer-item-name">{name}</FlexBox>
          )}
          <FlexBox>
            {err !== "" ? (
              <FlexBox>
                <Alert severity="error" variant="filled">
                  {err}
                </Alert>
              </FlexBox>
            ) : (
              <FlexBox />
            )}
            <FlexBox className="explorer-item-actions gap">
              {rename ? (
                <FlexBox
                  onClick={(ev) => {
                    setRename(!rename);
                    setErr("");
                    ev.stopPropagation();
                  }}
                >
                  <VscClose className="auto-margin" />
                </FlexBox>
              ) : (
                <FlexBox
                  onClick={(ev) => {
                    setRename(!rename);
                    setErr("");
                    ev.stopPropagation();
                  }}
                >
                  <VscEdit className="auto-margin" />
                </FlexBox>
              )}
              <FlexBox onClick={(ev) => ev.stopPropagation()}>
                <Modal
                  modalStyle={{ width: "400px" }}
                  escapeToCancel
                  style={{
                    flexDirection: "row-reverse",
                  }}
                  title="Delete a workflow"
                  button={
                    // <FlexBox style={{alignItems:'center'}}>
                    <HiOutlineTrash className="auto-margin red-text" />
                    // </FlexBox>
                  }
                  buttonProps={{
                    auto: true,
                    variant: "text",
                  }}
                  actionButtons={[
                    {
                      label: "Delete",

                      onClick: async () => {
                        const p = path.split("/", -1);
                        const pLast = p[p.length - 1];
                        await deleteNode(pLast, false);
                      },
                      buttonProps: { variant: "contained", color: "error" },
                      closesModal: true,
                    },
                    {
                      label: "Cancel",
                      closesModal: true,
                    },
                  ]}
                >
                  <FlexBox col gap>
                    <FlexBox>
                      Are you sure you want to delete &apos;{name}&apos;?
                      <br />
                      This action cannot be undone.
                    </FlexBox>
                  </FlexBox>
                </Modal>
              </FlexBox>
            </FlexBox>
          </FlexBox>
        </FlexBox>
      </FlexBox>
    </div>
  );
}

export function ApiFragment(props) {
  const { url, method, body, description } = props;
  return (
    <FlexBox className="helper-wrap col">
      <FlexBox className="helper-title row">
        <FlexBox className="row vertical-center">
          <Button className={`btn-method ${method}`}>{method}</Button>
          <div className="url">{url}</div>
        </FlexBox>
        <div className="description" style={{ textAlign: "right" }}>
          {description}
        </div>
      </FlexBox>
      {body ? (
        <FlexBox>
          <DirektivEditor
            height={150}
            value={props.body}
            readonly
            dlang={props.type}
          />
        </FlexBox>
      ) : (
        ""
      )}
    </FlexBox>
  );
}
