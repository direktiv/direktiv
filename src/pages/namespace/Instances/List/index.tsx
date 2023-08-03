import { FiltersObj, useInstances } from "~/api/instances/query/get";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import Filters from "../components/Filters";
import NoResult from "./NoResult";
import { Pagination } from "~/componentsNext/Pagination";
import Row from "./Row";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const instancesPerPage = 15;

const InstancesListPage = () => {
  const [offset, setOffset] = useState(0);
  const [filters, setFilters] = useState<FiltersObj>({});
  const { t } = useTranslation();
  const { data, isFetched } = useInstances({
    limit: instancesPerPage,
    offset,
    filters,
  });

  const handleFilterChange = (filters: FiltersObj) => {
    setFilters(filters);
    setOffset(0);
  };

  const numberOfInstances = data?.instances?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.instances.results.length === 0;
  const showPagination = numberOfInstances > instancesPerPage;
  const hasFilters = !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Boxes className="h-5" />
        {t("pages.instances.list.title")}
      </h3>
      <Card>
        <Filters filters={filters} onUpdate={handleFilterChange} />
        <Table className="border-t border-gray-5 dark:border-gray-dark-5">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.instances.list.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell className="w-32">
                {t("pages.instances.list.tableHeader.id")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.invoker")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.state")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.instances.list.tableHeader.startedAt")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.instances.list.tableHeader.updatedAt")}
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
                        ? t("pages.instances.list.empty.noFilterResults")
                        : t("pages.instances.list.empty.noInstances")
                    }
                  />
                </TableCell>
              </TableRow>
            ) : (
              data?.instances.results.map((instance) => (
                <Row
                  instance={instance}
                  key={instance.id}
                  namespace={data.namespace}
                  data-testid={`instance-row-${instance.id}`}
                />
              ))
            )}
          </TableBody>
        </Table>
      </Card>
      {showPagination && (
        <Pagination
          itemsPerPage={instancesPerPage}
          offset={offset}
          setOffset={setOffset}
          totalItems={numberOfInstances}
        />
      )}
    </div>
  );
};

export default InstancesListPage;
