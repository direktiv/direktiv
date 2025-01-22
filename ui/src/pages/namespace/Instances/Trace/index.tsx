import { FC, ReactElement } from "react";
import {
  Table,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { SquareGanttChartIcon } from "lucide-react";
import { TableBody } from "@tremor/react";
import mock from "./mock.json";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type SpanType = {
  spanId: string;
  startTimeUnixNano: string;
  endTimeUnixNano: string;
  children?: SpanType[];
};

type TreeElementProps = { id: string; label: string; depth?: number };

const TreeElement: FC<TreeElementProps> = ({ label, depth = 0 }) => (
  <div className="w-full" style={{ paddingLeft: `${depth * 12}px` }}>
    {label}
  </div>
);

type TimelineElementProps = { start: number; end: number; label: string };

const TimelineElement: FC<TimelineElementProps> = ({ start, end, label }) => (
  <div
    className="relative flex flex-row w-full"
    style={{ paddingLeft: `${start}%`, paddingRight: `${end}%` }}
  >
    <div
      className={twMergeClsx(
        "bg-primary-200 dark:bg-primary-700 rounded-sm h-5",
        "overflow-x-visible text-nowrap min-w-px w-full max-w-full"
      )}
      style={start > 50 ? { direction: "rtl" } : {}}
    >
      <div style={{ direction: "ltr", display: "inline-block" }}>{label}</div>
    </div>
  </div>
);

const SpanViewer: FC = () => {
  const { t } = useTranslation();
  const { timeline: spans } = mock;

  const timelineStart = Math.min(
    ...spans.map((span) => Number(span.startTimeUnixNano))
  );
  const timelineEnd = Math.max(
    ...spans.map((span) => Number(span.endTimeUnixNano))
  );

  const duration = timelineEnd - timelineStart;

  type SpanElement = {
    id: string;
    tree: ReactElement;
    timeline: ReactElement;
  };

  const processSpans = (spans: SpanType[], depth = 0): SpanElement[] =>
    spans.reduce<SpanElement[]>((acc, span) => {
      const spanStart = Number(span.startTimeUnixNano);
      const spanEnd = Number(span.endTimeUnixNano);
      const spanLength = spanEnd - spanStart;

      const start = Math.round((spanStart / duration - timelineStart) * 100);
      const end = Math.round((1 - (spanEnd - timelineStart) / duration) * 100);

      // Todo: adjust to sensible values when we have real data
      const labelDivider = duration < 100 ? 1000000 : 1000000000;
      const labelUnit = duration < 100 ? "ms" : "s";
      const label = `${Math.round(spanLength / labelDivider) / 100} ${labelUnit}`;

      acc.push({
        id: span.spanId,
        tree: (
          <TreeElement id={span.spanId} label={span.spanId} depth={depth} />
        ),
        timeline: <TimelineElement start={start} end={end} label={label} />,
      });

      if (span.children && span.children.length > 0) {
        acc.push(...processSpans(span.children, depth + 1));
      }
      return acc;
    }, []);

  const spanElements = processSpans(spans);

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
          <SquareGanttChartIcon className="h-5" />
          {t("pages.trace.title")}
        </h3>
      </div>
      <Card>
        <Table className="border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell className="w-40">
                {t("pages.trace.tableHeader.spanId")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.trace.tableHeader.timeline")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {spanElements.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.tree}</TableCell>
                <TableCell>{item.timeline}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default SpanViewer;
