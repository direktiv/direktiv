import "./style.css";

import Logs, { LogFooterButtons } from "../../components/logs/logs";
import React, { useState } from "react";

import { Config } from "../../util";
import FlexBox from "../../components/flexbox";
import { useApiKey } from "../../util/apiKeyProvider";
import { useMirrorLogs } from "../../hooks";

export default function ActivityLogs(props) {
  const { activity, namespace } = props;
  const [apiKey] = useApiKey();

  const { data } = useMirrorLogs(Config.url, true, namespace, activity, apiKey);
  const [follow, setFollow] = useState(true);

  return (
    <>
      <FlexBox col>
        <FlexBox
          style={{
            backgroundColor: "#002240",
            color: "white",
            borderRadius: "8px 8px 0px 0px",
            overflow: "hidden",
            padding: "8px",
          }}
        >
          <Logs
            logItems={data}
            wordWrap={true}
            autoScroll={follow}
            setAutoScroll={setFollow}
            overrideLoadingMsg={
              activity === null ? "No Activity Selected" : null
            }
          />
        </FlexBox>
        <div
          style={{
            height: "40px",
            backgroundColor: "#223848",
            color: "white",
            maxHeight: "40px",
            minHeight: "40px",
            padding: "0px 10px 0px 10px",
            boxShadow: "0px 0px 3px 0px #fcfdfe",
            alignItems: "center",
            borderRadius: " 0px 0px 8px 8px",
            overflow: "hidden",
          }}
        >
          <FlexBox
            gap
            style={{
              width: "100%",
              flexDirection: "row-reverse",
              height: "100%",
              alignItems: "center",
            }}
          >
            <LogFooterButtons
              setFollow={setFollow}
              follow={follow}
              data={data}
            />
          </FlexBox>
        </div>
      </FlexBox>
    </>
  );
}
