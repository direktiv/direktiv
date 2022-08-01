import dayjs from "dayjs";
import { createContext, CSSProperties, forwardRef, useCallback, useContext, useEffect, useRef, useState } from "react";
import { VscCopy, VscEye, VscEyeClosed, VscInbox, VscLayers, VscWholeWord, VscWordWrap } from "react-icons/vsc";
import AutoSizer from "react-virtualized-auto-sizer";
import { VariableSizeList } from "react-window";
import { copyTextToClipboard } from "../../util";
import Button from "../button";
import FlexBox from "../flexbox";
import "./style.css";

export interface LogsProps {
    logItems?: LogItem[]
    wordWrap: boolean
    autoScroll: boolean
    setAutoScroll: React.Dispatch<React.SetStateAction<boolean>> | null
    overrideNoDataMsg: string
    overrideLoadingMsg: string
}

export interface LogItem {
    t: string
    msg: string
}

export const DynamicListContext = createContext<
    Partial<{ setSize: (index: number, size: number) => void }>
>({});


export default function Logs({ logItems, wordWrap, autoScroll, setAutoScroll, overrideLoadingMsg, overrideNoDataMsg }: LogsProps) {
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
                logItems === null || logItems === undefined ?
                    <FlexBox className="row center gap" style={{ fontSize: "18px" }}>
                        <VscLayers /> {overrideLoadingMsg ? overrideLoadingMsg : "Loading Data"}
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
                                    <VscInbox />  {overrideNoDataMsg ? overrideNoDataMsg : "No Data"}
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
                    <span key={`log-timestamp-${index}`} className="timestamp">[{dayjs.utc(data[index].t).local().format("HH:mm:ss.SSS")}{`] `}</span>
                    {data[index].msg.match(/.{1,50}/g)?.map((mtkMsg, mtkIdx)=>{
                        return (
                            <span key={`log-msg-${mtkIdx}`} className="msg">{mtkMsg}</span>
                        )
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

    let clipboardData = ""

    data.forEach(item => {
        clipboardData += `[${dayjs.utc(item.t).local().format("HH:mm:ss.SSS")}] ${item.msg}\n`
    });

    return clipboardData
}

interface LogFooterButtonsProps {
    follow: boolean;
    setFollow: React.Dispatch<React.SetStateAction<boolean>>;
    wordWrap: boolean
    setWordWrap: React.Dispatch<React.SetStateAction<boolean>>;
    data: Array<LogItem> | null
    clipData: string
}

export function LogFooterButtons({ follow, setFollow, wordWrap, setWordWrap, data, clipData }: LogFooterButtonsProps) {
    return (
        <>

            <Button className="small terminal" onClick={() => {
                if (clipData) {
                    copyTextToClipboard(clipData)
                } else {
                    copyTextToClipboard(createClipboardData(data))
                }
            }}>
                <FlexBox className="row center gap-sm">
                    <VscCopy /> Copy <span className='hide-1000'>to Clipboard</span>
                </FlexBox>
            </Button>
            {
                follow !== undefined && setFollow !== undefined ?
                    <Button className="small terminal" onClick={() => setFollow(!follow)}>
                        <FlexBox className="row center gap-sm">
                            {follow ? <><VscEyeClosed /> Stop <span className='hide-1000'>watching</span></> :
                                <><VscEye /> Follow <span className='hide-1000'>logs</span></>}
                        </FlexBox>
                    </Button>
                    :
                    <></>
            }
            {
                wordWrap !== undefined && setWordWrap !== undefined ?
                    <Button className="small terminal" tip={wordWrap ? "Disable word wrapping" : "Enable word wrapping"} onClick={() => {
                        setWordWrap(!wordWrap)
                    }}>
                        <FlexBox className="row center gap-sm">
                            {wordWrap ? <><VscWholeWord /> Whole<span className='hide-1000'>{` Word`}</span></> :
                                <><VscWordWrap /> Wrap<span className='hide-1000'>{` Word`}</span></>}
                        </FlexBox>
                    </Button> :
                    <></>
            }
        </>
    )
}