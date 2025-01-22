import { FC, ReactElement } from "react";

import mock from "./mock.json";
import { twMergeClsx } from "~/util/helpers";

type SpanType = {
  spanId: string;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  children?: SpanType[];
};

type TreeElementProps = { span: SpanType; depth?: number };

const TreeElement: FC<TreeElementProps> = ({ span, depth = 0 }) => (
  <div className="h-8 w-full" style={{ paddingLeft: `${depth * 12}px` }}>
    {span.spanId}
  </div>
);

type TimelineElementProps = { start: number; end: number };

const TimelineElement: FC<TimelineElementProps> = ({ start, end }) => (
  <div
    className="relative h-8 flex flex-row w-full"
    style={{ paddingLeft: `${start}%`, paddingRight: `${end}%` }}
  >
    <div
      className={twMergeClsx(
        "bg-primary-200 dark:bg-primary-700 rounded-sm h-5",
        "overflow-x-visible text-nowrap min-w-px w-full max-w-full"
      )}
      style={start > 50 ? { direction: "rtl" } : {}}
    >
      <div style={{ direction: "ltr", display: "inline-block" }}>
        {start} {end}
      </div>
    </div>
  </div>
  // <div className="relative w-full h-8 bg-gray-200">
  //   <div
  //     className="whitespace-nowrap absolute left-[48%] right-[48%] bg-blue-400 rounded h-5 text-white text-sm flex items-center overflow-visible"
  //     style={start > 50 ? { direction: "rtl" } : {}}
  //   >
  //     Some text
  //   </div>
  // </div>
);

type TreeProps = { elements: ReactElement[] };

const Tree: FC<TreeProps> = ({ elements }) => (
  <div className="w-3/12">{...elements}</div>
);

type TimelineProps = { elements: ReactElement[] };

const Timeline: FC<TimelineProps> = ({ elements }) => (
  <div className="w-9/12">{...elements}</div>
);

const SpanViewer: FC = () => {
  const { timeline: spans } = mock;

  const timelineStart = Math.min(
    ...spans.map((span) => Number(span.startTimeUnixNano))
  );
  const timelineEnd = Math.max(
    ...spans.map((span) => Number(span.endTimeUnixNano))
  );

  const duration = timelineEnd - timelineStart;

  type SpanElement = {
    tree: ReactElement;
    timeline: ReactElement;
  };

  const processSpans = (spans: SpanType[], depth = 0): SpanElement[] =>
    spans.reduce<SpanElement[]>((acc, span) => {
      const start = Math.round(
        (Number(span.startTimeUnixNano) / duration - timelineStart) * 100
      );
      const end = Math.round(
        (1 - (Number(span.endTimeUnixNano) - timelineStart) / duration) * 100
      );

      acc.push({
        tree: <TreeElement span={span} depth={depth} />,
        timeline: <TimelineElement start={start} end={end} />,
      });

      if (span.children && span.children.length > 0) {
        acc.push(...processSpans(span.children, depth + 1));
      }
      return acc;
    }, []);

  const spanElements = processSpans(spans);

  return (
    <div className="flex flex-row w-full">
      <Tree elements={spanElements.map((item) => item.tree)} />
      <Timeline elements={spanElements.map((item) => item.timeline)} />
    </div>
  );
};

export default SpanViewer;
