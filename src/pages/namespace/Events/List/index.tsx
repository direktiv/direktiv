import { FiltersObj, useEvents } from "~/api/events/query/get";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Calendar } from "lucide-react";
import { Card } from "~/design/Card";
import Filters from "../components/Filters";
import NoResult from "./NoResult";
import Row from "./Row";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const itemsPerPage = 15;

const EventsPageList = () => {
  const [, setOffset] = useState(0);
  const [filters, setFilters] = useState<FiltersObj>({});

  const { data, isFetched } = useEvents({
    limit: itemsPerPage,
    offset: 0,
    filters,
  });

  const { t } = useTranslation();

  const handleFilterChange = (filters: FiltersObj) => {
    setFilters(filters);
    setOffset(0);
  };

  // const numberOfResults = data?.events?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.events.results.length === 0;
  // const showPagination = numberOfResults > itemsPerPage;
  const hasFilters = false; // !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Calendar className="h-5" />
        {t("pages.events.list.title")}
      </h3>
      <Card>
        <Filters filters={filters} onUpdate={handleFilterChange} />

        <Table className="border-t border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.events.list.tableHeader.type")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.list.tableHeader.id")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.list.tableHeader.source")}
              </TableHeaderCell>
              <TableHeaderCell>
                {t("pages.events.list.tableHeader.receivedAt")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {noResults ? (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoResult
                    message={
                      hasFilters
                        ? t("pages.events.list.empty.noFilterResults")
                        : t("pages.events.list.empty.noResults")
                    }
                  />
                </TableCell>
              </TableRow>
            ) : (
              data?.events.results.map((event) => (
                <Row
                  event={event}
                  key={event.id}
                  namespace={data.namespace}
                  data-testid={`event-row-${event.id}`}
                />
              ))
            )}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default EventsPageList;
