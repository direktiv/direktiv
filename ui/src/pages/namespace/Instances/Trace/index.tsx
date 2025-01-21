import { FC, ReactElement } from "react";

import mock from "./mock.json";

type SpanType = {
  spanId: string;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  children?: SpanType[];
};

type SpanRowProps = { span: SpanType; start: number; end: number };

const SpanRow: FC<SpanRowProps> = ({ span, start, end }) => (
  <div className="relative h-8">
    <div className="bg-blue-400 rounded h-5 absolute overflow-x-visible text-nowrap"></div>
    <div className="text-gray-500 absolute px-1">
      {span.spanId} {start} {end}{" "}
    </div>
  </div>
);

type TreeProps = { elements: ReactElement[] };

const Tree: FC<TreeProps> = ({ elements }) => <div>{...elements}</div>;

type TimelineProps = { elements: ReactElement[] };

const Timeline: FC<TimelineProps> = ({ elements }) => <div>{...elements}</div>;

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

  const processSpans = (spans: SpanType[]): SpanElement[] =>
    spans.reduce<SpanElement[]>((acc, span) => {
      const start = Math.round(
        (Number(span.startTimeUnixNano) / duration - timelineStart) * 100
      );
      const end = Math.round(
        (Number(span.endTimeUnixNano) / duration - timelineStart) * 100
      );

      acc.push({
        tree: <div>{span.spanId}</div>,
        timeline: <SpanRow span={span} start={start} end={end} />,
      });

      if (span.children && span.children.length > 0) {
        acc.push(...processSpans(span.children));
      }
      return acc;
    }, []);

  const spanElements = processSpans(spans);

  return (
    <div className="flex flex-row">
      <Tree elements={spanElements.map((item) => item.tree)} />
      <Timeline elements={spanElements.map((item) => item.timeline)} />
    </div>
  );
};

export default SpanViewer;
