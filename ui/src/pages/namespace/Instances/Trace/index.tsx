import { CircleX, SquareGanttChartIcon } from "lucide-react";
import { FC, useState } from "react";
import { SpanType, processSpans } from "./utils";
import {
  Table,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { TableBody } from "@tremor/react";
import mock from "./mock.json";
import moment from "moment";
import { useTranslation } from "react-i18next";

const TraceViewer: FC = () => {
  const { t } = useTranslation();
  const defaultSpans = mock.timeline;
  const [spans, setSpans] = useState<SpanType[]>(mock.timeline);
  const [isFiltered, setIsFiltered] = useState(false);
  const legendFormat = "YYYY-MM-DD, hh:mm:ss.SSS";

  const timelineStart = Math.min(
    ...spans.map((span) => Number(span.startTimeUnixNano))
  );
  const timelineEnd = Math.max(
    ...spans.map((span) => Number(span.endTimeUnixNano))
  );

  // Todo: when we have a real backend, it might be worth moving the spans
  // and filter functions to a hook or context provider, and also avoid
  // prop drilling the onFilter function.

  const applyFilter = (newSpans: SpanType[]) => {
    setSpans(newSpans);
    setIsFiltered(true);
  };

  const clearFilter = () => {
    setSpans(defaultSpans);
    setIsFiltered(false);
  };

  const spanElements = processSpans({
    spans,
    timelineStart,
    timelineEnd,
    onFilter: applyFilter,
  });

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
          <SquareGanttChartIcon className="h-5" />
          {t("pages.trace.title")}
        </h3>
      </div>
      <Card>
        <div className="flex gap-5 p-2 border-b border-gray-5 dark:border-gray-dark-5">
          {isFiltered ? (
            <Button type="button" variant="outline" icon onClick={clearFilter}>
              {t("pages.trace.filters.active")}
              <CircleX className="h-5" />
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              disabled
              className="font-normal"
            >
              {t("pages.trace.filters.inactive")}
            </Button>
          )}
        </div>

        <Table className="border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell className="w-56 border-r border-gray-3 dark:border-gray-dark-3">
                {t("pages.trace.tableHeader.spanId")}
              </TableHeaderCell>
              <TableHeaderCell className="flex flex-row justify-between">
                <span>
                  {moment(timelineStart / 1000000000).format(legendFormat)}
                </span>
                <span>
                  {moment(timelineEnd / 100000000).format(legendFormat)}
                </span>
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody className="divide-y-0">
            {spanElements.map((item) => (
              <TableRow key={item.id}>
                <TableCell className="border-b border-r border-gray-3 dark:border-gray-dark-3">
                  {item.tree}
                </TableCell>
                <TableCell>{item.timeline}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default TraceViewer;
