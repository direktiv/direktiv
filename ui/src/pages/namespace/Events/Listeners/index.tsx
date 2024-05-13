import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import ListenersTable from "./Table";
import { Pagination } from "~/components/Pagination";
import Row from "./Row";
import { useEventListeners } from "~/api/eventListenersv2/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const pageSize = 10;

const ListenersList = () => {
  const [offset, setOffset] = useState(0);
  const { data, isFetched, isAllowed, noPermissionMessage } = useEventListeners(
    { limit: pageSize, offset }
  );

  const { t } = useTranslation();

  if (!data?.data) return null;

  const numberOfResults = data?.meta?.total ?? 0;
  const noResults = isFetched && numberOfResults === 0;
  const showPagination = numberOfResults > pageSize;

  return (
    <div className="flex grow flex-col gap-y-3 p-5">
      <Card>
        <ListenersTable>
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
                data.data.map((listener, i) => (
                  <Row
                    listener={listener}
                    key={i}
                    namespace={listener.namespace}
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
        </ListenersTable>
      </Card>
      {showPagination && (
        <Pagination
          itemsPerPage={pageSize}
          offset={offset}
          setOffset={(value) => setOffset(value)}
          totalItems={numberOfResults}
        />
      )}
    </div>
  );
};

export default ListenersList;
