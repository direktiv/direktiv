import "./style.css";

import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../components/content-panel";
import React, { useEffect, useRef, useState } from "react";
import { VscAdd, VscLock, VscSync, VscUnlock } from "react-icons/vsc";
import { useMirror, useNodes } from "../../hooks";
import { useNavigate, useParams } from "react-router";

import ActivityLogs from "./logs.jsx";
import ActivityTable from "./activities";
import Alert from "../../components/alert";
import Button from "../../components/button";
import { Config } from "../../util";
import FlexBox from "../../components/flexbox";
import Loader from "../../components/loader";
import MirrorInfoPanel from "./info";
import { ModalHeadless } from "../../components/modal";
import Tippy from "@tippyjs/react";
import { useApiKey } from "../../util/apiKeyProvider";
import { useTheme } from "@mui/material/styles";

export default function MirrorPage(props) {
  const { namespace, setBreadcrumbChildren } = props;
  const params = useParams();
  const navigate = useNavigate();
  const [activity, setActivity] = useState(null);
  const [currentlyLocking, setCurrentlyLocking] = useState(true);
  const [isReadOnly, setIsReadOnly] = useState(true);
  const [apiKey] = useApiKey();

  const [errorMsg, setErrorMsg] = useState(null);
  const [load, setLoad] = useState(true);
  const [syncVisible, setSyncVisible] = useState(false);
  const [ts, setTs] = useState(Date.now());
  const refetch = () => setTs(Date.now());

  let path = `/`;
  if (params["*"] !== undefined) {
    path = `/${params["*"]}`;
  }

  const {
    info,
    activities,
    err,
    setLock,
    updateSettings,
    cancelActivity,
    sync,
  } = useMirror(
    Config.url,
    false,
    namespace,
    path,
    apiKey,
    "limit=50",
    "order.field=CREATED",
    "order.direction=DESC",
    `timestamp=${ts}`
  );
  const {
    data,
    getNode,
    err: nodeErr,
  } = useNodes(Config.url, false, namespace, path, apiKey, `limit=1`);

  const setLockRef = useRef(setLock);
  const syncRef = useRef(sync);
  const getNodeRef = useRef(getNode);
  const setBreadcrumbChildrenRef = useRef(setBreadcrumbChildren);

  // Error Handling Non existent node and bad mirror
  useEffect(() => {
    if (err) {
      setErrorMsg("Error getting mirror info: " + err);
    } else if (nodeErr) {
      console.error("Error getting node: ", nodeErr);
      navigate(`/n/${namespace}/explorer${path}`);
    }
  }, [nodeErr, err, data, navigate, namespace, path]);

  // Error Handling bad node
  useEffect(() => {
    if (!getNodeRef.current) {
      return;
    }

    if (!load && data) {
      getNodeRef
        .current()
        .then((nodeData) => {
          if (nodeData.node.expandedType !== "git") {
            navigate(`/n/${namespace}/explorer${path}`);
          }
        })
        .catch(() => {
          navigate(`/n/${namespace}/explorer${path}`);
        });
    }
  }, [data, load, navigate, namespace, path]);

  // Keep track of getNodeRef
  useEffect(() => {
    getNodeRef.current = getNode;
  }, [getNode]);

  useEffect(() => {
    if (nodeErr) {
      setErrorMsg("Error getting node: " + nodeErr);
      return;
    }

    const handler = setTimeout(() => {
      if (currentlyLocking) {
        getNode()
          .then((nodeData) => {
            setIsReadOnly(nodeData.node.readOnly);
          })
          .catch((e) => {
            setErrorMsg("Error getting node: " + e.message);
          })
          .finally(() => {
            setCurrentlyLocking(false);
          });
      }
    }, 1000);

    return () => {
      clearTimeout(handler);
    };
  }, [currentlyLocking, getNode, nodeErr]);

  useEffect(() => {
    if (data && info) {
      setLoad(false);
    }
  }, [data, info, load]);

  useEffect(() => {
    if (!setBreadcrumbChildrenRef.current || !syncRef.current) {
      return;
    }

    setBreadcrumbChildrenRef.current(
      <FlexBox
        center
        row
        gap
        style={{ justifyContent: "flex-end", paddingRight: "6px" }}
      >
        <Button
          tooltip="Sync mirror with remote"
          variant="outlined"
          color="info"
          onClick={() => {
            setSyncVisible(!syncVisible);
          }}
        >
          <FlexBox center row gap="sm">
            <VscSync />
            Sync
          </FlexBox>
        </Button>
        <ModalHeadless
          visible={syncVisible}
          setVisible={setSyncVisible}
          escapeToCancel
          activeOverlay
          title="Sync Mirror"
          titleIcon={<VscSync />}
          style={{
            maxWidth: "68px",
          }}
          modalStyle={{
            width: "300px",
          }}
          actionButtons={[
            {
              label: "Sync",

              onClick: async () => {
                await syncRef.current(true);
                refetch();
                setTimeout(() => {
                  refetch();
                }, 2000);
              },

              buttonProps: { variant: "contained", color: "primary" },
              closesModal: true,
            },
            {
              label: "Cancel",
              closesModal: true,
            },
          ]}
        >
          <FlexBox col gap style={{ paddingTop: "8px" }}>
            <FlexBox className="col center info-update-label">
              Fetch and sync mirror with latest content from remote repository?
            </FlexBox>
          </FlexBox>
        </ModalHeadless>
      </FlexBox>
    );
  }, [currentlyLocking, isReadOnly, syncVisible]);

  // Keep Refs up to date
  useEffect(() => {
    setBreadcrumbChildrenRef.current = setBreadcrumbChildren;
    setLockRef.current = setLock;
    syncRef.current = sync;
  }, [setBreadcrumbChildren, setLock, sync]);

  // Unmount cleanup breadcrumb children
  useEffect(
    () => () => {
      if (setBreadcrumbChildrenRef.current) {
        setBreadcrumbChildrenRef.current(null);
      }
    },
    []
  );

  if (!namespace) {
    return null;
  }

  return (
    <>
      <Loader load={load} timer={1000}>
        {errorMsg ? (
          <FlexBox
            style={{
              maxHeight: "50px",
              paddingRight: "6px",
              paddingBottom: "8px",
            }}
          >
            <Alert
              severity="error"
              variant="filled"
              onClose={() => {
                setErrorMsg(null);
              }}
              grow
            >{`Error: ${errorMsg}`}</Alert>
          </FlexBox>
        ) : null}
        <FlexBox col gap style={{ paddingRight: "8px" }}>
          {/* <BreadcrumbCorner>
                    </BreadcrumbCorner> */}
          <FlexBox row gap wrap style={{ flex: "1 1 0%", maxHeight: "65vh" }}>
            <ContentPanel
              id="panel-activity-list"
              style={{
                flex: 2,
                width: "100%",
                minHeight: "60vh",
                maxHeight: "55vh",
              }}
            >
              <ContentPanelTitle>
                <ContentPanelTitleIcon>
                  <VscAdd />
                </ContentPanelTitleIcon>
                <FlexBox gap style={{ alignItems: "center" }}>
                  Activity List
                </FlexBox>
              </ContentPanelTitle>
              <ContentPanelBody style={{ overflow: "auto" }}>
                <FlexBox style={{ flexShrink: "1", height: "fit-content" }}>
                  <ActivityTable
                    activities={activities}
                    setActivity={setActivity}
                    cancelActivity={cancelActivity}
                    setErrorMsg={setErrorMsg}
                  />
                </FlexBox>
                <FlexBox style={{ flexGrow: "1" }}></FlexBox>
              </ContentPanelBody>
            </ContentPanel>
            <MirrorInfoPanel
              info={info}
              updateSettings={async (...props) => {
                await updateSettings(...props);
                refetch();
                setTimeout(() => {
                  refetch();
                }, 2000);
              }}
              namespace={namespace}
              style={{ width: "100%", height: "100%", flex: 1 }}
            />
          </FlexBox>
          <ContentPanel style={{ width: "100%", minHeight: "15vh", flex: 1 }}>
            <ContentPanelTitle>
              <ContentPanelTitleIcon>
                <VscAdd />
              </ContentPanelTitleIcon>
              <FlexBox gap style={{ alignItems: "center" }}>
                Activity Logs
              </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody>
              <ActivityLogs
                activity={activity}
                namespace={namespace}
                setErrorMsg={setErrorMsg}
              />
            </ContentPanelBody>
          </ContentPanel>
        </FlexBox>
      </Loader>
    </>
  );
}

export function MirrorReadOnlyBadge() {
  return (
    <Tippy
      content="This mirrors contents are currently read-only. This can be unlocked in mirror setttings"
      trigger="mouseenter focus"
      zIndex={10}
    >
      <div>
        <Button
          variant="contained"
          color="info"
          disabled
          style={{ borderRadius: "20px" }}
        >
          <FlexBox center row gap="sm">
            <VscLock />
            ReadOnly
          </FlexBox>
        </Button>
      </div>
    </Tippy>
  );
}

export function MirrorWritableBadge() {
  const theme = useTheme();
  return (
    <Tippy
      content="This mirrors contents are currently writable. This can be unlocked in mirror setttings"
      trigger="mouseenter focus"
      zIndex={10}
    >
      <div>
        <Button
          disabled
          variant="contained"
          color="secondary"
          sx={{
            "&:disabled": {
              backgroundColor: theme.palette.secondary.main,
              color: theme.palette.primary.main,
              fontWeight: "bold",
              borderRadius: "20px",
            },
          }}
        >
          <FlexBox center row gap="sm">
            <VscUnlock />
            Writable
          </FlexBox>
        </Button>
      </div>
    </Tippy>
  );
}
