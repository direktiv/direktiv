import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import {
  useEventListenersPageSize,
  usePageSizeActions,
} from "~/util/store/pagesize";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import ListenersTable from "./Table";
import { Pagination } from "~/components/Pagination";
import Row from "./Row";
import { SelectPageSize } from "../../../../components/SelectPageSize";
import { useEventListeners } from "~/api/eventListeners/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const ListenersList = () => {
  const pageSize = useEventListenersPageSize();
  const { setEventListenersPageSize } = usePageSizeActions();
  const [offset, setOffset] = useState(0);
  const { data, isFetched, isAllowed, noPermissionMessage } = useEventListeners(
    { limit: parseInt(pageSize), offset }
  );

  const { t } = useTranslation();

  if (!data?.data) return null;

  const numberOfResults = data?.meta?.total ?? 0;
  const noResults = isFetched && numberOfResults === 0;

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
      <div className="flex items-center justify-end gap-2">
        <SelectPageSize
          initialPageSize={pageSize}
          onSelect={(selectedSize) => {
            setEventListenersPageSize(selectedSize);
            setOffset(0);
          }}
        />
        <Pagination
          itemsPerPage={parseInt(pageSize)}
          value={offset}
          onChange={(value) => setOffset(value)}
          totalItems={numberOfResults}
        />
      </div>
    </div>
  );
};

export default ListenersList;
