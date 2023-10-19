/* eslint-disable tailwindcss/no-custom-classname */
import "./style.css";

import {
  CancelledState,
  FailState,
  InstancesTable,
  RunningState,
  SuccessState,
} from "../instances";
import { Config, GenerateRandomKey, createLogFilter } from "../../util";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../components/content-panel";
import { Link, useNavigate, useSearchParams } from "react-router-dom";
import Logs, { LogFooterButtons } from "../../components/logs/logs";
import React, { useCallback, useEffect, useState } from "react";
import {
  VscScreenFull,
  VscScreenNormal,
  VscSourceControl,
  VscTerminal,
} from "react-icons/vsc";
import { useInstance, useInstanceLogs, useWorkflow } from "../../hooks";

import Alert from "../../components/alert";
import { AutoSizer } from "react-virtualized";
import Button from "../../components/button";
import DirektivEditor from "../../components/editor";
import FlexBox from "../../components/flexbox";
import Loader from "../../components/loader";
import { Tooltip } from "@mui/material";
import WorkflowDiagram from "../../components/diagram";
import { useApiKey } from "../../util/apiKeyProvider";
import { useParams } from "react-router";

function InstancePageWrapper(props) {
  const { namespace } = props;
  if (!namespace) {
    return null;
  }

  return <InstancePage namespace={namespace} />;
}

export default InstancePageWrapper;

function TabbedButtons(props) {
  const { tabBtn, setTabBtn, setSearchParams } = props;

  const tabBtns = [];
  const tabBtnLabels = ["Flow Graph", "Child Instances"];

  for (let i = 0; i < tabBtnLabels.length; i++) {
    const key = GenerateRandomKey();
    let classes = "tab-btn";
    if (i === tabBtn) {
      classes += " active-tab-btn";
    }

    tabBtns.push(
      <FlexBox key={key} className={classes}>
        <div
          onClick={() => {
            setTabBtn(i);
            setSearchParams({
              tab: i,
            });
          }}
        >
          {tabBtnLabels[i]}
        </div>
      </FlexBox>
    );
  }

  return (
    <FlexBox
      className="tabbed-btns-container"
      style={{ flexShrink: "1", flexGrow: "0" }}
    >
      <FlexBox className="tabbed-btns">{tabBtns}</FlexBox>
    </FlexBox>
  );
}

function InstancePage(props) {
  const { namespace } = props;
  const { id } = useParams();
  const [searchParams, setSearchParams] = useSearchParams();
  const navigate = useNavigate();

  const [apiKey] = useApiKey();
  const [load, setLoad] = useState(true);
  const [wfpath, setWFPath] = useState("");
  const [ref, setRef] = useState("");
  const [rev, setRev] = useState(null);
  const [follow, setFollow] = useState(true);
  const [width] = useState(window.innerWidth);
  const [clipData, setClipData] = useState(null);
  const [instanceID, setInstanceID] = useState(id);
  const [tabBtn, setTabBtn] = useState(
    searchParams.get("tab") !== null ? parseInt(searchParams.get("tab")) : 0
  );
  const [onlyShow, setOnlyShow] = useState("");

  const toggleFullscreen = useCallback(
    (targetWindow) => {
      if (onlyShow.length > 0) {
        setOnlyShow("");
        return;
      }

      setOnlyShow(targetWindow);
    },
    [onlyShow]
  );

  const getHideClass = useCallback(
    (targetWindow) => {
      if (onlyShow.length > 0 && onlyShow !== targetWindow) {
        return "hide";
      }

      return "";
    },
    [onlyShow]
  );

  const hideClassIf = useCallback(
    (targetWindow) => {
      if (onlyShow.length > 0 && onlyShow === targetWindow) {
        return "hide";
      }

      return "";
    },
    [onlyShow]
  );

  // let instanceID = params["id"];
  React.useEffect(() => {
    setLoad(true);
    setInstanceID(id);
  }, [id]);

  const {
    data,
    err,
    workflow,
    latestRevision,
    getInput,
    getOutput,
    cancelInstance,
  } = useInstance(Config.url, true, namespace, instanceID, apiKey);

  useEffect(() => {
    if (load && data !== null && workflow != null && latestRevision != null) {
      const split = data.as.split(":");
      setWFPath(split[0]);
      if (workflow.revision === latestRevision) {
        setRef("latest");
      } else if (split[1] !== "latest") {
        setRef(split[1]);
      }
      setRev(workflow.revision);
      setLoad(false);
    }
  }, [load, data, workflow, latestRevision]);

  useEffect(() => {
    if (
      err === "Not Found" ||
      (err !== null && err.indexOf("invalid UUID") >= 0)
    ) {
      navigate(`/not-found`);
    }
  }, [err, navigate]);

  if (data === null) {
    return null;
  }

  if (err !== null) {
    // TODO
    return null;
  }

  let label = null;
  if (data.status === "complete") {
    label = <SuccessState />;
  } else if (
    data.status === "cancelled" ||
    data.errorCode === "direktiv.cancels.api"
  ) {
    label = <CancelledState />;
  } else if (data.status === "failed" || data.status === "crashed") {
    label = <FailState />;
  } else if (data.status === "running") {
    label = <RunningState />;
  }

  const wfName = data.as.split(":")[0];

  return (
    <Loader load={load} timer={3000}>
      <FlexBox col gap style={{ paddingRight: "8px" }}>
        <FlexBox
          className={`gap wrap ${hideClassIf("output")}`}
          style={{ minHeight: "50%", flex: "1" }}
        >
          <FlexBox
            className={`${getHideClass("logs")}`}
            style={{ minWidth: "340px", flex: "5" }}
          >
            <ContentPanel style={{ width: "100%", minHeight: "40vh" }}>
              <ContentPanelTitle>
                <ContentPanelTitleIcon>
                  <VscTerminal />
                </ContentPanelTitleIcon>
                <FlexBox gap style={{ alignItems: "center" }}>
                  <div>Instance Details</div>
                  {label}
                  <FlexBox
                    row
                    gap
                    center="y"
                    style={{ justifyContent: "flex-end" }}
                  >
                    {data.status === "running" || data.status === "pending" ? (
                      <Button
                        color="info"
                        variant="outlined"
                        onClick={() => {
                          cancelInstance();
                          setLoad(true);
                        }}
                      >
                        <span className="red-text">Cancel</span>
                      </Button>
                    ) : null}
                    {rev === null || rev === "" ? null : (
                      <Link
                        to={`/n/${namespace}/explorer/${wfName}?${
                          ref === "latest"
                            ? `tab=2`
                            : `tab=1&revision=${rev}&revtab=0`
                        }`}
                      >
                        <Button color="info" variant="outlined">
                          <span className="hide-600">View&nbsp;</span>Workflow
                        </Button>
                      </Link>
                    )}
                    <Button
                      tooltip={onlyShow ? "Collapse Window" : "Expand Window"}
                      color="info"
                      variant="outlined"
                      onClick={() => toggleFullscreen("logs")}
                      style={{ marginTop: "2px" }}
                    >
                      {onlyShow ? <VscScreenNormal /> : <VscScreenFull />}
                    </Button>
                  </FlexBox>
                </FlexBox>
              </ContentPanelTitle>
              <InstanceLogs
                setClipData={setClipData}
                clipData={clipData}
                namespace={namespace}
                instanceID={instanceID}
                follow={follow}
                setFollow={setFollow}
                width={width}
              />
            </ContentPanel>
          </FlexBox>
          <FlexBox
            className={`gap wrap ${getHideClass("input")}`}
            style={{
              minIoCopyHeight: "40%",
              minWidth: "300px",
              flex: "2",
              flexWrap: "wrap-reverse",
            }}
          >
            <FlexBox style={{ minWidth: "300px" }}>
              <ContentPanel style={{ width: "100%", minHeight: "40vh" }}>
                <ContentPanelTitle>
                  <ContentPanelTitleIcon>
                    <VscTerminal />
                  </ContentPanelTitleIcon>
                  <FlexBox gap>
                    <div>Input</div>
                    <FlexBox
                      row
                      gap
                      center="y"
                      style={{ justifyContent: "flex-end" }}
                    >
                      <Button
                        tooltip={onlyShow ? "Collapse Window" : "Expand Window"}
                        color="info"
                        variant="outlined"
                        onClick={() => toggleFullscreen("input")}
                      >
                        <FlexBox col center style={{ fontSize: "15px" }}>
                          {onlyShow ? <VscScreenNormal /> : <VscScreenFull />}
                        </FlexBox>
                      </Button>
                    </FlexBox>
                  </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody>
                  <Input getInput={getInput} />
                </ContentPanelBody>
              </ContentPanel>
            </FlexBox>
          </FlexBox>
        </FlexBox>
        <FlexBox
          className={`gap wrap ${getHideClass("output")}`}
          style={{ minHeight: "40%", flex: "1" }}
        >
          <FlexBox
            className={hideClassIf("output")}
            style={{ minWidth: "300px", flex: "5" }}
          >
            <ContentPanel style={{ width: "100%", minHeight: "40vh" }}>
              <ContentPanelTitle>
                <ContentPanelTitleIcon>
                  <VscSourceControl />
                </ContentPanelTitleIcon>
                <FlexBox gap style={{ alignItems: "center" }}>
                  <div style={{ flex: "1", whiteSpace: "nowrap" }}>
                    {`${tabBtn === 0 ? "Flow Graph" : "Child Instances"}`}
                  </div>
                  {tabBtn === 1 && data.invoker.startsWith("instance:") ? (
                    <Link
                      to={`/n/${namespace}/instances/${data.invoker.replace(
                        "instance:",
                        ""
                      )}`}
                      reloadDocument
                    >
                      <Button color="info" variant="outlined">
                        <span className="hide-600">View</span> Parent
                      </Button>
                    </Link>
                  ) : null}
                  <TabbedButtons
                    setSearchParams={setSearchParams}
                    searchParams={searchParams}
                    tabBtn={tabBtn}
                    setTabBtn={setTabBtn}
                  />
                </FlexBox>
              </ContentPanelTitle>
              {tabBtn === 0 ? (
                <ContentPanelBody>
                  <InstanceDiagram
                    status={data.status}
                    namespace={namespace}
                    wfpath={wfpath}
                    rev={rev}
                    instRef={ref}
                    flow={data.flow}
                  />
                </ContentPanelBody>
              ) : null}
              {tabBtn === 1 ? (
                <InstancesTable
                  placeholder="No child instances have executed from this instance. Child instances will appear here."
                  namespace={namespace}
                  mini={true}
                  hideTitle={true}
                  panelStyle={{ border: "unset" }}
                  filter={[
                    `filter.field=TRIGGER&filter.type=MATCH&filter.val=instance:${instanceID}`,
                  ]}
                />
              ) : null}
            </ContentPanel>
          </FlexBox>
          <FlexBox style={{ minWidth: "300px", flex: "2" }}>
            <ContentPanel style={{ width: "100%", minHeight: "40vh" }}>
              <ContentPanelTitle>
                <ContentPanelTitleIcon>
                  <VscTerminal />
                </ContentPanelTitleIcon>
                <FlexBox gap>
                  <div>Output</div>
                  <FlexBox
                    row
                    gap
                    center="y"
                    style={{ justifyContent: "flex-end" }}
                  >
                    <Button
                      tooltip={onlyShow ? "Collapse Window" : "Expand Window"}
                      color="info"
                      variant="outlined"
                      onClick={() => toggleFullscreen("output")}
                    >
                      <FlexBox col center style={{ fontSize: "15px" }}>
                        {onlyShow ? <VscScreenNormal /> : <VscScreenFull />}
                      </FlexBox>
                    </Button>
                  </FlexBox>
                </FlexBox>
              </ContentPanelTitle>
              <ContentPanelBody>
                <Output getOutput={getOutput} status={data.status} />
              </ContentPanelBody>
            </ContentPanel>
          </FlexBox>
        </FlexBox>
      </FlexBox>
    </Loader>
  );
}

function InstanceLogs(props) {
  const [apiKey] = useApiKey();
  const { noPadding, namespace, instanceID } = props;
  let paddingStyle = { padding: "12px" };
  if (noPadding) {
    paddingStyle = { padding: "0px" };
  }

  const [filterWorkflow, setFilterWorkflow] = useState("");
  const [filterStateId, setFilterStateId] = useState("");
  const [filterLoopIndex, setFilterLoopIndex] = useState("");
  const [buttonDisabled, setButtonDisabled] = useState(true);
  const [filterParams, setFilterParams] = useState([]);
  const { data } = useInstanceLogs(
    Config.url,
    true,
    namespace,
    instanceID,
    apiKey,
    ...filterParams
  );

  const example = data?.find(
    (x) => x?.tags && !!x?.tags?.workflow && !!x?.tags?.["state-id"]
  );

  const applyFilter = () => {
    setFilterParams(
      createLogFilter({
        workflow: filterWorkflow,
        stateId: filterStateId,
        loopIndex: filterLoopIndex,
      })
    );
  };

  const disableFilter = () => {
    setFilterWorkflow("");
    setFilterStateId("");
    setFilterParams([]);
  };

  const [wordWrap, setWordWrap] = useState(false);
  const [follow, setFollow] = useState(true);
  const [verbose, setVerbose] = useState(false);
  const [showFilterBar, setShowFilterbar] = useState(false);
  const [showTooltip, setShowTooltip] = useState(false);

  const displayValidationMsg = useCallback(() => {
    if (filterWorkflow === "" && filterStateId === "") {
      return null;
    }

    if (filterWorkflow === "") {
      return "please enter a workflow name";
    }

    if (filterStateId === "") {
      return "please enter a state name";
    }

    return null;
  }, [filterStateId, filterWorkflow]);

  useEffect(() => {
    if (displayValidationMsg() !== null) {
      setButtonDisabled(true);
      return;
    }

    const generateParams = JSON.stringify(
      createLogFilter({
        workflow: filterWorkflow,
        stateId: filterStateId,
        loopIndex: filterLoopIndex,
      })
    );

    if (generateParams === JSON.stringify(filterParams)) {
      setButtonDisabled(true);
      return;
    }

    setButtonDisabled(false);
  }, [
    displayValidationMsg,
    filterLoopIndex,
    filterParams,
    filterStateId,
    filterWorkflow,
  ]);

  return (
    <>
      <FlexBox col style={{ ...paddingStyle }}>
        <FlexBox className="logs">
          <Logs
            logItems={data}
            wordWrap={wordWrap}
            autoScroll={follow}
            verbose={verbose}
            setAutoScroll={setFollow}
            filterControls={{
              setFilterWorkflow,
              setFilterStateId,
              setFilterLoopIndex,
              setFilterParams,
              setShowFilterbar,
            }}
          />
        </FlexBox>
        <div
          className={`logs-footer ${showFilterBar && "logs-footer--two-lines"}`}
          style={{
            alignItems: "center",
            borderRadius: " 0px 0px 8px 8px",
            overflow: "hidden",
          }}
        >
          <form
            onSubmit={(e) => {
              e.preventDefault();
              applyFilter();
            }}
            style={{
              all: "unset",
            }}
          >
            <FlexBox
              gap
              style={{
                width: "100%",
                flexDirection: "row-reverse",
                height: "50%",
                alignItems: "center",
                ...(showFilterBar === false && { display: "none" }),
              }}
            >
              <Tooltip
                title={displayValidationMsg()}
                placement="top"
                arrow
                open={displayValidationMsg() && showTooltip ? true : false}
                onMouseEnter={() => {
                  setShowTooltip(true);
                }}
                onMouseLeave={() => {
                  setShowTooltip(false);
                }}
              >
                <div>
                  <Button
                    tabIndex="3"
                    color="terminal"
                    variant="contained"
                    type="submit"
                    disabled={buttonDisabled}
                  >
                    Update Filter
                  </Button>
                </div>
              </Tooltip>
              <label>
                state name{" "}
                <input
                  tabIndex="2"
                  placeholder={
                    example
                      ? `f.e. ${example?.tags?.["state-id"]}`
                      : "state name"
                  }
                  type="text"
                  value={filterStateId}
                  onChange={(e) => setFilterStateId(e.target.value)}
                  style={{
                    width: "100px",
                  }}
                />
              </label>
              <label>
                workflow name{" "}
                <input
                  tabIndex="1"
                  placeholder={
                    example
                      ? `f.e. ${example?.tags?.workflow}`
                      : "workflow name"
                  }
                  type="text"
                  value={filterWorkflow}
                  onChange={(e) => setFilterWorkflow(e.target.value)}
                  style={{
                    width: "200px",
                  }}
                />
              </label>
            </FlexBox>
          </form>
          <FlexBox
            gap
            style={{
              width: "100%",
              flexDirection: "row-reverse",
              ...(showFilterBar ? { height: "50%" } : { height: "100%" }),
              alignItems: "center",
            }}
          >
            <LogFooterButtons
              setFollow={setFollow}
              follow={follow}
              setVerbose={setVerbose}
              verbose={verbose}
              setFilter={(val) => {
                if (val === true) {
                  setShowFilterbar(true);
                } else {
                  disableFilter();
                  setShowFilterbar(false);
                }
              }}
              filter={showFilterBar}
              setWordWrap={setWordWrap}
              wordWrap={wordWrap}
              data={data}
            />
          </FlexBox>
        </div>
      </FlexBox>
    </>
  );
}

function InstanceDiagram(props) {
  const { wfpath, rev, instRef, flow, namespace, status } = props;

  const [load, setLoad] = useState(true);
  const [workflowMissing, setWorkflowMissing] = useState(false);
  const [wfdata, setWFData] = useState("");
  const [apiKey] = useApiKey();

  const { getWorkflowRevisionData } = useWorkflow(
    Config.url,
    false,
    namespace,
    wfpath,
    apiKey
  );

  useEffect(() => {
    const handler = setTimeout(() => {
      async function getwf() {
        if (
          wfpath !== "" &&
          instRef !== "" &&
          rev !== null &&
          rev !== "" &&
          load
        ) {
          const refWF = await getWorkflowRevisionData(
            instRef === "latest" ? instRef : rev
          );
          setWFData(atob(refWF.revision.source));
          setLoad(false);
        } else if (rev === "") {
          setWorkflowMissing(true);
        }
      }

      getwf();
    }, 200);

    return () => {
      clearTimeout(handler);
    };
  }, [wfpath, rev, load, instRef, getWorkflowRevisionData]);

  if (workflowMissing) {
    return (
      <div className="container-alert">
        Workflow revision that executed instance no longer exists
      </div>
    );
  }

  if (load) {
    return null;
  }

  return (
    <WorkflowDiagram
      instanceStatus={status}
      disabled={true}
      flow={flow}
      workflow={wfdata}
    />
  );
}

function Input(props) {
  const { getInput } = props;

  const [input, setInput] = useState(null);
  const [load, setLoad] = useState(true);

  useEffect(() => {
    async function get() {
      const data = await getInput();
      try {
        const x = JSON.stringify(JSON.parse(data), null, 2);
        setInput(x);
      } catch (e) {
        setInput(data);
      }
    }

    if (load && input === null && getInput) {
      setLoad(false);
      get();
    }
  }, [input, getInput, load]);

  return (
    <FlexBox style={{ flexDirection: "column" }}>
      {!input ? (
        <Alert severity="info">No input data was provided</Alert>
      ) : null}
      <FlexBox style={{ overflow: "hidden" }}>
        {/* <div style={{width: "100%", height: "100%"}}> */}
        <AutoSizer>
          {({ height, width }) => (
            <DirektivEditor
              height={height}
              width={width}
              dlang="json"
              value={input}
              readonly={true}
            />
          )}
        </AutoSizer>
        {/* </div> */}
      </FlexBox>
    </FlexBox>
  );
}

function Output(props) {
  const { getOutput, status } = props;

  const [load, setLoad] = useState(true);
  const [output, setOutput] = useState(null);

  useEffect(() => {
    async function get() {
      if (load && status !== "pending" && output === null && getOutput) {
        setLoad(false);
        try {
          const data = await getOutput();
          const x = JSON.stringify(JSON.parse(data), null, 2);
          setOutput(x);
        } catch (e) {
          console.error(e);
        }
      }
    }
    get();
  }, [output, load, getOutput, status]);

  return (
    <FlexBox col gap>
      {!output ? (
        <Alert severity="info">No output data was resolved</Alert>
      ) : null}
      <FlexBox style={{ padding: "0px", overflow: "hidden" }}>
        <AutoSizer>
          {({ height, width }) => (
            <DirektivEditor
              height={height}
              width={width}
              dlang="json"
              value={output}
              readonly={true}
            />
          )}
        </AutoSizer>
      </FlexBox>
    </FlexBox>
  );
}
