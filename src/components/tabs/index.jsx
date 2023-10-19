import "./style.css";

import React, { useEffect, useState } from "react";

import FlexBox from "../flexbox";

function Tabs(props) {
  const { callback, tabIndex, id } = props;
  const [activeTab, setActiveTab] = useState(0);
  const [init, setInit] = useState(false);

  useEffect(() => {
    if (!init) {
      return;
    }
    if (callback) {
      callback(activeTab);
    }
  }, [callback, activeTab, init]);

  useEffect(() => {
    if (tabIndex !== null && tabIndex !== undefined && tabIndex >= 0) {
      setActiveTab(tabIndex);
    }
    setInit(true);
  }, [tabIndex]);

  const { style, headers, tabs } = props;
  if (!headers || !tabs) {
    return <>Bad tabs definition (missing tabs or headers).</>;
  }

  if (headers.length !== tabs.length) {
    return (
      <>
        Bad tabs definition (headers array must be equal length to tabs array).
      </>
    );
  }

  const headerDOMs = [];
  for (let i = 0; i < headers.length; i++) {
    let classes = "tab-header center-align";
    if (i === activeTab) {
      classes += " active-tab";
    }

    headerDOMs.push(
      <FlexBox
        className={classes}
        onClick={() => {
          setActiveTab(i);
        }}
      >
        <span>{headers[i]}</span>
      </FlexBox>
    );
  }

  return (
    <FlexBox id={id} className="col gap" style={{ ...style }}>
      <div className="tabs-row">{headerDOMs}</div>
      <FlexBox>{tabs[activeTab]}</FlexBox>
    </FlexBox>
  );
}

export default Tabs;

export function Tab(props) {
  const { children, style } = props;
  return <FlexBox style={{ ...style }}>{children}</FlexBox>;
}
