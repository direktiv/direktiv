import { FC } from "react";
import mock from "./mock.json";

type SpanType = {
  spanId: string;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  children?: SpanType[];
};

// const flattenSpans = (spans: SpanType[]): SpanType[] =>
//   spans.reduce<SpanType[]>((acc, span) => {
//     acc.push(span);
//     if (span.children) {
//       acc.push(...flattenSpans(span.children));
//     }
//     return acc;
//   }, []);

type SpanRowProps = { span: SpanType; scale: number; start: number };

const SpanRow: FC<SpanRowProps> = ({ span, scale, start }) => {
  const left = (Number(span.startTimeUnixNano) - start) * scale;
  const width =
    (Number(span.endTimeUnixNano) - Number(span.startTimeUnixNano)) * scale;

  return (
    <div className="relative h-8">
      <div
        className="bg-blue-400 rounded h-5 absolute overflow-x-visible text-nowrap"
        style={{ left: `${left}px`, width: `${width}px` }}
      ></div>
      <div
        className="text-gray-500 absolute px-1"
        style={{ left: `${left + width}px` }}
      >
        {span.spanId} {Math.round(width * 100) / 100}
      </div>
    </div>
  );
};

type SpanViewerProps = { spans: SpanType[]; timeLineWidth?: number };

const SpanViewer: FC<SpanViewerProps> = ({ spans, timeLineWidth = 700 }) => {
  const flattenedSpans: SpanType[] = [];

  const timelineStart = Math.min(
    ...spans.map((span) => Number(span.startTimeUnixNano))
  );
  const timelineEnd = Math.max(
    ...spans.map((span) => Number(span.endTimeUnixNano))
  );

  // Scale factor: pixels per millisecond
  const scale = timeLineWidth / (timelineEnd - timelineStart);

  const getSubTree = (spans: SpanType[]) =>
    spans.map((span) => (
      <div key={span.spanId} className="pl-2">
        <div className="flex flex-row">
          <div>{span.spanId}</div>
          <SpanRow span={span} scale={scale} start={timelineStart} />
        </div>
        {span.children && getSubTree(span.children)}
      </div>
    ));

  return (
    <div className="flex flex-row">
      <div>{getSubTree(spans)}</div>
      <TimeLine spans={flattenedSpans} />
    </div>
  );
};

type TimeLineProps = { spans: SpanType[]; containerWidth?: number };

const TimeLine: FC<TimeLineProps> = ({ spans, containerWidth = 1000 }) => {
  if (spans.length === 0) {
    return <div className="text-gray-500">No spans to display</div>;
  }

  // const allSpans = flattenSpans(spans);

  const timelineStart = Math.min(
    ...spans.map((span) => Number(span.startTimeUnixNano))
  );
  const timelineEnd = Math.max(
    ...spans.map((span) => Number(span.endTimeUnixNano))
  );

  // Scale factor: pixels per millisecond
  const scale = containerWidth / (timelineEnd - timelineStart);

  return (
    <div className="space-y-2 p-1">
      {/* Add vertical spacing between rows */}
      {spans.map((span) => (
        <SpanRow
          key={span.spanId}
          span={span}
          scale={scale}
          start={timelineStart}
        />
      ))}
    </div>
  );
};

// const Spans = ({ data }: { data: Array<SpanType> }) => <></>;

const TracePage: FC = () => (
  <div>
    <h1>Mock data:</h1>
    <SpanViewer spans={mock.timeline}></SpanViewer>
  </div>
);

export default TracePage;
