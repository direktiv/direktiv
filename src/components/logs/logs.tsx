import dayjs from "dayjs";
import { createContext, CSSProperties, useCallback, useContext, useEffect, useRef } from "react";
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


export default function NewLogs({ logItems, wordWrap, autoScroll, setAutoScroll }: LogsProps) {
    const listRef = useRef<VariableSizeList | null>(null);

    const sizeMap = useRef<{ [key: string]: number }>({});

    useEffect(()=>{
        if (autoScroll && listRef.current && sizeMap.current){
            listRef.current.scrollToItem(Object.keys(sizeMap.current).length-1);
        }
    }, [autoScroll])

    // AutoScroll to bottom when logItems update
    const scrollToEnd = useCallback((props: {visibleStopIndex: number} ) =>{
        if (!autoScroll){
            return
        }

        const finalIndex = logItems ? logItems.length - 1 : 0;
        if (props.visibleStopIndex < finalIndex && listRef.current) {
            console.log(finalIndex)
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

        if(setAutoScroll && autoScroll){
            setAutoScroll(false)
        }
    }, [setAutoScroll, autoScroll])

    return (
        <FlexBox className="log-window" onWheel={(e: any)=>{disableAutoScroll(e.deltaY < 0)}} onMouseDown={(e: any)=>{disableAutoScroll(true)}}>
            {
                logItems ?
                    <DynamicListContext.Provider value={{ setSize }}>
                        <AutoSizer>
                            {({ height, width }) => (
                                <VariableSizeList
                                    onItemsRendered={scrollToEnd}
                                    ref={listRef}
                                    width={width}
                                    height={height}
                                    itemData={logItems}
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
                    <div>
                        No Data
                    </div>
            }
        </FlexBox>
    )
}

interface Props {
    index: number;
    width: number;
    data: LogItem[];
    style: CSSProperties;
    wordWrap: boolean
}

const ListRow = ({ index, width, data, style, wordWrap }: Props) => {
    const { setSize } = useContext(DynamicListContext);
    const rowRoot = useRef<null | HTMLDivElement>(null);

    useEffect(() => {
        if (rowRoot.current) {
            setSize && setSize(index, rowRoot.current.getBoundingClientRect().height);
        }
    }, [index, setSize, width]);

    return (
        <div style={style}>
            <div className="log-row"
                ref={rowRoot}
            >
                <p className={wordWrap ? "word-wrap" : ""}>
                    <span className="timestamp">[{dayjs.utc(data[index].node.t).local().format("HH:mm:ss.SSS")}{`] `}</span>
                    {data[index].node.msg}
                </p>
            </div>
        </div>
    );
};
