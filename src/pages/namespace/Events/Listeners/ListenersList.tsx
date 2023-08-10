import { Dispatch, SetStateAction } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { ArrowDownToDot } from "lucide-react";
import { Card } from "~/design/Card";
import NoResult from "./NoResult";
import { Pagination } from "~/componentsNext/Pagination";
import Row from "./Row";
import { itemsPerPage } from ".";
import { useEventListeners } from "~/api/eventListeners/query/get";
import { useTranslation } from "react-i18next";

const ListenersList = ({
  offset,
  setOffset,
}: {
  offset: number;
  setOffset: Dispatch<SetStateAction<number>>;
}) => {
  const { data, isFetched } = useEventListeners({
    limit: itemsPerPage,
    offset,
  });

  const { t } = useTranslation();

  const numberOfResults = data?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.results.length === 0;
  const showPagination = numberOfResults > itemsPerPage;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <div className="flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 pb-2 pt-1 font-bold">
          <ArrowDownToDot className="h-5" />
          {t("pages.events.listeners.title")}
        </h3>
      </div>

      <Card>
        <Table className="border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.workflow")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.type")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.mode")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.createdAt")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.eventTypes")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {noResults ? (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoResult
                    message={t("pages.events.listeners.empty.noResults")}
                  />
                </TableCell>
              </TableRow>
            ) : (
              data?.results.map((listener, i) => (
                <Row
                  listener={listener}
                  key={i}
                  namespace={data.namespace}
                  data-testid={`listener-row-${i}`}
                />
              ))
            )}
          </TableBody>
        </Table>
      </Card>
      {showPagination && (
        <Pagination
          itemsPerPage={itemsPerPage}
          offset={offset}
          setOffset={(value) => setOffset(value)}
          totalItems={numberOfResults}
        />
      )}
    </div>
  );
};

export default ListenersList;
