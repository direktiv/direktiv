import "./style.css";

import {
  CSSProperties,
  createContext,
  forwardRef,
  useCallback,
  useContext,
  useEffect,
  useRef,
  useState,
} from "react";
import { TbBug, TbBugOff, TbFilter, TbFilterOff } from "react-icons/tb";
import {
  VscCopy,
  VscEye,
  VscEyeClosed,
  VscInbox,
  VscLayers,
  VscWholeWord,
  VscWordWrap,
} from "react-icons/vsc";
// @ts-ignore 🚧 ignore since this component will be removed after redesign
import { copyTextToClipboard, createLogFilter } from "../../util";

import AutoSizer from "react-virtualized-auto-sizer";
import Button from "../button";
import FlexBox from "../flexbox";
import Tippy from "@tippyjs/react";
import { VariableSizeList } from "react-window";
import dayjs from "dayjs";
import utc from "dayjs/plugin/utc";

dayjs.extend(utc);

type FilterControls = {
  setFilterWorkflow: React.Dispatch<React.SetStateAction<string>>;
  setFilterStateId: React.Dispatch<React.SetStateAction<string>>;
  setFilterLoopIndex: React.Dispatch<React.SetStateAction<string>>;
  setFilterParams: React.Dispatch<React.SetStateAction<string[]>>;
  setShowFilterbar: React.Dispatch<React.SetStateAction<boolean>>;
};

export interface LogsProps {
  logItems?: LogItem[];
  /**
   * Enables word wrapping on log lines
   */
  wordWrap: boolean;
  /**
   * If enabled, scrolls to bottom of list
   */
  autoScroll: boolean;
  /**
   * Show verbose logs
   */
  verbose?: boolean;
  /**
   * React Set State for autoscroll. This is used for setting autoScroll to false when user scrolls up.
   */
  setAutoScroll?: React.Dispatch<React.SetStateAction<boolean>>;
  /**
   * Message to display when logItems is length of 0
   */
  overrideNoDataMsg?: string;
  /**
   * Message to display when logItems is undefined
   */
  overrideLoadingMsg?: string;

  filterControls?: FilterControls;
}

export interface LogItem {
  t: string;
  level: "debug" | "info" | "error" | "panic";
  msg: string;
  tags: {
    workflow: string; // f.e. "somepath/to/someworkflow"
    "loop-index": string; // f.e. "1"
    // always there, except on the first step
    "state-id"?: string; // f.e. "getter"
  };
}

export const DynamicListContext = createContext<
  Partial<{ setSize: (index: number, size: number) => void }>
>({});

/**
 * Logs component that only renders the visible log items of a list.
 * Supports auto scrolling to bottom as logs change, and word wrapping.
 */
function Logs({
  logItems,
  wordWrap = false,
  autoScroll = false,
  verbose,
  setAutoScroll,
  overrideLoadingMsg,
  overrideNoDataMsg,
  filterControls,
}: LogsProps) {
  const listRef = useRef<VariableSizeList | null>(null);

  const sizeMap = useRef<{ [key: string]: number }>({});

  const [scrollInit, setScrollInit] = useState(false);

  // AutoScroll to bottom when autoScroll is changed to true.
  // If listRef is not ready, scroll will be added to a que (This only happens the first time)
  useEffect(() => {
    if (autoScroll && listRef.current && sizeMap.current) {
      listRef.current.scrollToItem(
        Object.keys(sizeMap.current).length,
        "start"
      );
      return;
    } else if (scrollInit) {
      return;
    }

    const interval = setInterval(() => {
      if (autoScroll && listRef.current && sizeMap.current) {
        listRef.current.scrollToItem(Object.keys(sizeMap.current).length - 1);
        setScrollInit(true);
      }
    }, 100);

    return () => clearInterval(interval);
  }, [autoScroll, scrollInit]);

  // AutoScroll to bottom when logItems update
  const scrollToEnd = useCallback(
    (props: { visibleStopIndex: number }) => {
      if (!autoScroll) {
        return;
      }

      const finalIndex = logItems ? logItems.length - 1 : 0;
      if (props.visibleStopIndex < finalIndex && listRef.current) {
        listRef.current.scrollToItem(finalIndex);
      }
    },
    [autoScroll, logItems]
  );

  // Clear list cache when values change
  const setSize = useCallback((index: number, size: number) => {
    if (sizeMap.current[index] !== size) {
      sizeMap.current = { ...sizeMap.current, [index]: size };
      if (listRef.current) {
        listRef.current.resetAfterIndex(0);
      }
    }
  }, []);

  const getSize = useCallback((index: number) => {
    return sizeMap.current[index] || 100;
  }, []);

  const disableAutoScroll = useCallback(
    (extraConditions: boolean) => {
      if (!extraConditions) {
        return;
      }

      if (setAutoScroll && autoScroll) {
        setAutoScroll(false);
      }
    },
    [setAutoScroll, autoScroll]
  );

  return (
    <FlexBox
      className="log-window"
      onWheel={() => {
        disableAutoScroll(true);
      }}
      onClick={() => {
        disableAutoScroll(true);
      }}
    >
      {logItems === null || logItems === undefined ? (
        <FlexBox center row gap style={{ fontSize: "18px" }}>
          <VscLayers />{" "}
          {overrideLoadingMsg ? overrideLoadingMsg : "Loading Data"}
        </FlexBox>
      ) : (
        <>
          {logItems.length > 0 ? (
            <DynamicListContext.Provider value={{ setSize }}>
              <AutoSizer>
                {({ height, width }) => (
                  <VariableSizeList
                    onItemsRendered={scrollToEnd}
                    ref={listRef}
                    width={width}
                    height={height}
                    itemData={logItems}
                    innerElementType={innerElementType}
                    itemCount={logItems.length}
                    itemSize={getSize}
                    overscanCount={4}
                  >
                    {({ ...props }) => (
                      <ListRow
                        {...props}
                        width={width}
                        wordWrap={wordWrap}
                        verbose={verbose}
                        filterControls={filterControls}
                      />
                    )}
                  </VariableSizeList>
                )}
              </AutoSizer>
            </DynamicListContext.Provider>
          ) : (
            <FlexBox center row gap style={{ fontSize: "18px" }}>
              <VscInbox /> {overrideNoDataMsg ? overrideNoDataMsg : "No Data"}
            </FlexBox>
          )}
        </>
      )}
    </FlexBox>
  );
}

export default Logs;

interface ListRowProps {
  index: number;
  width: number;
  data: LogItem[];
  style: CSSProperties;
  wordWrap?: boolean;
  verbose?: boolean;
  filterControls?: FilterControls;
}

const innerElementType = forwardRef(({ style, ...rest }: any, ref) => (
  <div
    ref={ref}
    style={{
      ...style,
      height: `${parseFloat(style.height) + 3 * 2}px`,
    }}
    {...rest}
  />
));
innerElementType.displayName = "innerElementType";

const ListRow = ({
  index,
  width,
  data,
  style,
  wordWrap,
  verbose,
  filterControls,
}: ListRowProps) => {
  const { setSize } = useContext(DynamicListContext);
  const rowRoot = useRef<null | HTMLDivElement>(null);

  useEffect(() => {
    if (rowRoot.current) {
      setSize && setSize(index, rowRoot.current.getBoundingClientRect().height);
    }
  }, [index, setSize, width]);

  const {
    workflow,
    "state-id": stateId,
    "loop-index": loopIndex,
  } = data[index].tags;

  const isError = ["error", "panic"].includes(data[index].level);

  return (
    <div
      style={{
        ...style,
        top: `${parseFloat(style.top as string)}px`,
      }}
    >
      <div className={`log-row ${isError && "log-row--error"}`} ref={rowRoot}>
        <span className={wordWrap ? "word-wrap" : "whole-word"}>
          <span className="timestamp">
            [{dayjs.utc(data[index].t).local().format("HH:mm:ss")}
            {`] `}
          </span>
          {verbose ? (
            <Tippy
              content="filter logs by this tag"
              trigger="mouseenter focus"
              delay={100}
              zIndex={10}
            >
              <span
                role="button"
                style={{ cursor: "pointer" }}
                onClick={() => {
                  filterControls?.setFilterWorkflow(workflow ?? "");
                  filterControls?.setFilterStateId(stateId ?? "");
                  filterControls?.setFilterLoopIndex(loopIndex ?? "");
                  filterControls?.setShowFilterbar(true);
                  filterControls?.setFilterParams(
                    createLogFilter({
                      workflow: workflow ?? "",
                      stateId: stateId ?? "",
                      loopIndex: loopIndex ?? "",
                    })
                  );
                }}
              >
                {loopIndex && <span className="tag-name">({loopIndex}) </span>}
                {workflow && <span className="tag-name">{workflow}</span>}
                {stateId && <span className="tag-state">/{stateId}</span>}{" "}
              </span>
            </Tippy>
          ) : null}

          {data[index].msg.match(/.{1,50}/g)?.map((mtkMsg, mtkIdx) => {
            return (
              <span key={`log-msg-${mtkIdx}`} className="msg">
                {mtkMsg}
              </span>
            );
          })}
        </span>
      </div>
    </div>
  );
};

export function createClipboardData(data: Array<LogItem> | null) {
  if (!data) {
    return "";
  }

  let clipboardData = "";

  data.forEach((item) => {
    const {
      workflow,
      "state-id": stateId,
      "loop-index": loopIndex,
    } = item?.tags ?? {};

    const tags = [];

    loopIndex && tags.push(`(${loopIndex}) `);
    workflow && tags.push(workflow);
    stateId && tags.push(`/${stateId}`);

    clipboardData += `[${dayjs
      .utc(item.t)
      .local()
      .format("HH:mm:ss.SSS")}] ${tags.join("")} ${item.msg}\n`;
  });

  return clipboardData;
}

interface LogFooterButtonsProps {
  follow: boolean;
  setFollow: React.Dispatch<React.SetStateAction<boolean>>;
  verbose?: boolean;
  setVerbose?: React.Dispatch<React.SetStateAction<boolean>>;
  filter?: boolean;
  setFilter?: React.Dispatch<React.SetStateAction<boolean>>;
  wordWrap?: boolean;
  setWordWrap?: React.Dispatch<React.SetStateAction<boolean>>;
  data: Array<LogItem> | null;
  clipData?: string;
}

export function LogFooterButtons({
  follow,
  setFollow,
  verbose,
  setVerbose,
  filter,
  setFilter,
  wordWrap,
  setWordWrap,
  data,
  clipData,
}: LogFooterButtonsProps) {
  return (
    <>
      <Button
        color="terminal"
        variant="contained"
        onClick={() => {
          if (clipData) {
            copyTextToClipboard(clipData);
          } else {
            copyTextToClipboard(createClipboardData(data));
          }
        }}
      >
        <FlexBox center row gap="sm">
          <VscCopy /> Copy <span className="hide-1000">to Clipboard</span>
        </FlexBox>
      </Button>
      {filter !== undefined && setFilter && (
        <Button
          color="terminal"
          variant="contained"
          onClick={() => {
            setFilter(!filter);
          }}
        >
          <FlexBox center row gap="sm">
            {filter ? (
              <>
                <TbFilterOff />
                Disable Filter
              </>
            ) : (
              <>
                <TbFilter />
                Filter Logs
              </>
            )}
          </FlexBox>
        </Button>
      )}
      {verbose !== undefined && setVerbose && (
        <Button
          color="terminal"
          variant="contained"
          onClick={() => {
            setVerbose((old) => !old);
          }}
        >
          <FlexBox center row gap="sm">
            {verbose ? (
              <>
                <TbBugOff />
                Disable verbose logs
              </>
            ) : (
              <>
                <TbBug />
                Enable verbose logs
              </>
            )}
          </FlexBox>
        </Button>
      )}
      {follow !== undefined && setFollow !== undefined ? (
        <Button
          color="terminal"
          variant="contained"
          onClick={() => setFollow(!follow)}
        >
          <FlexBox center row gap="sm">
            {follow ? (
              <>
                <VscEyeClosed /> Stop{" "}
                <span className="hide-1000">watching</span>
              </>
            ) : (
              <>
                <VscEye /> Follow <span className="hide-1000">logs</span>
              </>
            )}
          </FlexBox>
        </Button>
      ) : null}
      {wordWrap !== undefined && setWordWrap !== undefined ? (
        <Button
          color="terminal"
          variant="contained"
          tooltip={wordWrap ? "Disable word wrapping" : "Enable word wrapping"}
          onClick={() => {
            setWordWrap(!wordWrap);
          }}
        >
          <FlexBox center row gap="sm">
            {wordWrap ? (
              <>
                <VscWholeWord /> Whole
                <span className="hide-1000">{` Word`}</span>
              </>
            ) : (
              <>
                <VscWordWrap /> Wrap<span className="hide-1000">{` Word`}</span>
              </>
            )}
          </FlexBox>
        </Button>
      ) : null}
    </>
  );
}
