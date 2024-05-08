import { Dispatch, SetStateAction } from "react";
import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import ListenersTable from "./Table";
import { Pagination } from "~/components/Pagination";
import Row from "./Row";
import { itemsPerPage } from ".";
import { useEventListeners } from "~/api/eventListenersv2/query/get";
import { useTranslation } from "react-i18next";

const ListenersList = ({
  offset,
  setOffset,
}: {
  offset: number;
  setOffset: Dispatch<SetStateAction<number>>;
}) => {
  const { data, isFetched, isAllowed, noPermissionMessage } =
    useEventListeners();

  const { t } = useTranslation();

  const numberOfResults = data?.data.length ?? 0;
  const noResults = isFetched && data?.data.length === 0;
  const showPagination = false; // numberOfResults > itemsPerPage;

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
                data?.data.map((listener, i) => (
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
