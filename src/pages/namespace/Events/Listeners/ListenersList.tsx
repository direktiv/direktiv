import { Dispatch, SetStateAction } from "react";
import {
  NoPermissions,
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import { Pagination } from "~/components/Pagination";
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
  const { data, isFetched, isAllowed, noPermissionMessage } = useEventListeners(
    {
      limit: itemsPerPage,
      offset,
    }
  );

  const { t } = useTranslation();

  const numberOfResults = data?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.results.length === 0;
  const showPagination = numberOfResults > itemsPerPage;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Card>
        <Table className="border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.type")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.listeners.tableHeader.target")}
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
            {isAllowed ? (
              <>
                {noResults ? (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={6}>
                      <NoResult icon={Antenna}>
                        {t("pages.events.listeners.empty.noResults")}
                      </NoResult>
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
              </>
            ) : (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoPermissions>{noPermissionMessage}</NoPermissions>
                </TableCell>
              </TableRow>
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
