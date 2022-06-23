import dayjs from "dayjs";
import { createContext, CSSProperties, forwardRef, useCallback, useContext, useEffect, useRef, useState } from "react";
import { VscInbox, VscLayers } from "react-icons/vsc";
import AutoSizer from "react-virtualized-auto-sizer";
import { VariableSizeList } from "react-window";
import FlexBox from "../flexbox";
import "./style.css"

export interface LogsProps {
    logItems?: LogItem[]
    wordWrap: boolean
    autoScroll: boolean
    setAutoScroll: React.Dispatch<React.SetStateAction<boolean>> | null
}

export interface LogItem {
    node: Node
}

export interface Node {
    t: string
    msg: string
}

export const DynamicListContext = createContext<
    Partial<{ setSize: (index: number, size: number) => void }>
>({});


export default function Logs({ logItems, wordWrap, autoScroll, setAutoScroll }: LogsProps) {
    const listRef = useRef<VariableSizeList | null>(null);

    const sizeMap = useRef<{ [key: string]: number }>({});

    const [scrollInit, setScrollInit] = useState(false)

    // AutoScroll to bottom when autoScroll is changed to true.
    // If listRef is not ready, scroll will be added to a que (This only happens the first time)
    useEffect(() => {
        if (autoScroll && listRef.current && sizeMap.current) {
            listRef.current.scrollToItem(Object.keys(sizeMap.current).length, "start");
            return
        } else if (scrollInit) {
            return
        }

        const interval = setInterval(() => {
            if (autoScroll && listRef.current && sizeMap.current) {
                listRef.current.scrollToItem(Object.keys(sizeMap.current).length - 1);
                setScrollInit(true)
            }
        }, 100);

        return () => clearInterval(interval);
    }, [autoScroll, scrollInit])

    // AutoScroll to bottom when logItems update
    const scrollToEnd = useCallback((props: { visibleStopIndex: number }) => {
        if (!autoScroll) {
            return
        }

        const finalIndex = logItems ? logItems.length - 1 : 0;
        if (props.visibleStopIndex < finalIndex && listRef.current) {
            listRef.current.scrollToItem(finalIndex);
        }
    }, [autoScroll, logItems])

    // Clear list cache when values change
    const setSize = useCallback((index: number, size: number) => {
        if (sizeMap.current[index] !== size) {
            sizeMap.current = { ...sizeMap.current, [index]: size };
            if (listRef.current) {
                listRef.current.resetAfterIndex(0);
            }
        }
    }, []);

    const getSize = useCallback((index: any) => {
        return sizeMap.current[index] || 100;
    }, []);

    const disableAutoScroll = useCallback((extraConditions: boolean) => {
        if (!extraConditions) {
            return
        }

        if (setAutoScroll && autoScroll) {
            setAutoScroll(false)
        }
    }, [setAutoScroll, autoScroll])

    return (
        <FlexBox className="log-window" onWheel={(e: any) => { disableAutoScroll(true) }} onMouseDown={(e: any) => {
            disableAutoScroll(true)
        }}>
            {
                logItems === null || logItems === undefined?
                    <FlexBox className="row center gap" style={{ fontSize: "18px" }}>
                        <VscLayers /> Loading Data
                    </FlexBox>
                    :
                    <>
                        {
                            logItems.length > 0 ?
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
                                                {({ ...props }) => <ListRow {...props} width={width} wordWrap={wordWrap} />}
                                            </VariableSizeList>
                                        )}
                                    </AutoSizer>
                                </DynamicListContext.Provider>
                                :
                                <FlexBox className="row center gap" style={{ fontSize: "18px" }}>
                                    <VscInbox /> No Data
                                </FlexBox>
                        }
                    </>
            }
        </FlexBox>
    )
}

interface ListRowProps {
    index: number;
    width: number;
    data: LogItem[];
    style: CSSProperties;
    wordWrap: boolean
}

const innerElementType = forwardRef(({ style, ...rest }: any, ref) => (
    <div
        ref={ref}
        style={{
            ...style,
            height: `${parseFloat(style.height) + 3 * 2}px`
        }}
        {...rest}
    />
));

const ListRow = ({ index, width, data, style, wordWrap }: ListRowProps) => {
    const { setSize } = useContext(DynamicListContext);
    const rowRoot = useRef<null | HTMLDivElement>(null);

    useEffect(() => {
        if (rowRoot.current) {
            setSize && setSize(index, rowRoot.current.getBoundingClientRect().height);
        }
    }, [index, setSize, width]);

    return (
        <div style={{
            ...style,
            top: `${parseFloat(style.top as string)}px`
        }}>
            <div className="log-row"
                ref={rowRoot}
            >
                <span className={wordWrap ? "word-wrap" : "whole-word"}>
                    <span className="timestamp">[{dayjs.utc(data[index].node.t).local().format("HH:mm:ss.SSS")}{`] `}</span>
                    <span className="msg">{data[index].node.msg}</span>
                </span>
            </div>
        </div>
    );
};


export function createClipboardData(data: Array<LogItem> | null) {
    if (!data) {
        return "";
    }

    let clipboardData = ""

    data.forEach(item => {
        clipboardData += `[${dayjs.utc(item.node.t).local().format("HH:mm:ss.SSS")}] ${item.node.msg}\n`
    });

    return clipboardData
}
