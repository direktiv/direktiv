import { FC } from "react";
import mock from "./mock.json";

type SpanType = { spanId: string; children?: SpanType[] };

const Spans = ({ data }: { data: Array<SpanType> }) => (
  <>
    {data.map((span) => (
      <div className="ml-2" key={span.spanId}>
        <div>{span.spanId}</div>
        {span.children && <Spans data={span.children} />}
      </div>
    ))}
  </>
);

const TracePage: FC = () => (
  <div>
    <h1>Mock data:</h1>
    <Spans data={mock.timeline} />
  </div>
);

export default TracePage;
