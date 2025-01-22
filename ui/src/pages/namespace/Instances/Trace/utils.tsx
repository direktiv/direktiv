import { ReactElement } from "react";
import TimelineElement from "./components/TimelineElement";
import TreeElement from "./components/TreeElement";

type SpanType = {
  spanId: string;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  children?: SpanType[];
};

type SpanElement = {
  id: string;
  tree: ReactElement;
  timeline: ReactElement;
};

type Options = {
  spans: SpanType[];
  depth?: number;
  timelineStart: number;
  timelineEnd: number;
};

export const processSpans = ({
  spans,
  depth = 0,
  timelineStart,
  timelineEnd,
}: Options): SpanElement[] =>
  spans.reduce<SpanElement[]>((acc, span) => {
    const spanStart = Number(span.startTimeUnixNano);
    const spanEnd = Number(span.endTimeUnixNano);
    const timelineLength = timelineEnd - timelineStart;
    const spanLength = spanEnd - spanStart;

    const start = Math.round(
      (spanStart / timelineLength - timelineStart) * 100
    );
    const end = Math.round(
      (1 - (spanEnd - timelineStart) / timelineLength) * 100
    );

    // Todo: adjust to sensible values when we have real data
    const labelDivider = timelineLength < 100 ? 1000000 : 1000000000;
    const labelUnit = timelineLength < 100 ? "ms" : "s";
    const label = `${Math.round(spanLength / labelDivider) / 100} ${labelUnit}`;

    acc.push({
      id: span.spanId,
      tree: <TreeElement id={span.spanId} label={span.spanId} depth={depth} />,
      timeline: <TimelineElement start={start} end={end} label={label} />,
    });

    if (span.children && span.children.length > 0) {
      acc.push(
        ...processSpans({
          spans: span.children,
          depth: depth + 1,
          timelineStart,
          timelineEnd,
        })
      );
    }
    return acc;
  }, []);
