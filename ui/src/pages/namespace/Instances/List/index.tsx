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

import { Boxes } from "lucide-react";
import { Card } from "~/design/Card";
import Filters from "../components/Filters";
import { FiltersObj } from "~/api/instances/query/utils";
import { Pagination } from "~/components/Pagination";
import Row from "./Row";
import { useInstances } from "~/api/instances/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const instancesPerPage = 15;

const InstancesListPage = () => {
  const [offset, setOffset] = useState(0);
  const [filters, setFilters] = useState<FiltersObj>({});
  const { t } = useTranslation();
  const { data, isSuccess, isAllowed, noPermissionMessage } = useInstances({
    limit: instancesPerPage,
    offset,
    filters,
  });

  const handleFilterChange = (filters: FiltersObj) => {
    setFilters(filters);
    setOffset(0);
  };

  const instances = data?.data ?? [];
  const numberOfInstances = data?.meta?.total ?? 0;
  const noResults = isSuccess && instances.length === 0;
  const showPagination = numberOfInstances > instancesPerPage;
  const hasFilters = !!Object.keys(filters).length;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <div className="flex flex-col gap-4 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 pb-1 font-bold">
          <Boxes className="h-5" />
          {t("pages.instances.list.title")}
        </h3>
      </div>
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
                {t("pages.instances.list.tableHeader.finishedAt")}
              </TableHeaderCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {isAllowed ? (
              <>
                {noResults ? (
                  <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                    <TableCell colSpan={6}>
                      <NoResult icon={Boxes}>
                        {hasFilters
                          ? t("pages.instances.list.empty.noFilterResults")
                          : t("pages.instances.list.empty.noInstances")}
                      </NoResult>
                    </TableCell>
                  </TableRow>
                ) : (
                  instances.map((instance) => (
                    <Row
                      instance={instance}
                      key={instance.id}
                      data-testid={`instance-row-${instance.id}`}
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
