import { NoPermissions, NoResult, TableCell, TableRow } from "~/design/Table";
import {
  usePageSize,
  usePageSizeActions,
} from "~/util/store/pagesizes/pagesize";

import { Antenna } from "lucide-react";
import { Card } from "~/design/Card";
import ListenersTable from "./Table";
import { Pagination } from "~/components/Pagination";
import PaginationProvider from "~/components/PaginationProvider";
import Row from "./Row";
import { SelectPageSize } from "../History/components/SelectPageSize";
import { useEventListeners } from "~/api/eventListeners/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const ListenersList = () => {
  const pageSize = usePageSize("eventlisteners");
  const { setPageSize } = usePageSizeActions("eventlisteners");
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
      <PaginationProvider items={data.data} pageSize={parseInt(pageSize)}>
        {({ goToFirstPage }) => (
          <>
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
                  setPageSize(selectedSize);
                  goToFirstPage();
                }}
              />
              <Pagination
                itemsPerPage={parseInt(pageSize)}
                offset={offset}
                setOffset={(value) => setOffset(value)}
                totalItems={numberOfResults}
              />
            </div>
          </>
        )}
      </PaginationProvider>
    </div>
  );
};

export default ListenersList;
