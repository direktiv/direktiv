import { FiltersObj, useInstances } from "~/api/instances_obsolete/query/get";
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
import { Pagination } from "~/components/Pagination";
import Row from "./Row";
import { useInstanceList } from "~/api/instances/query/get";
import { useNamespace } from "~/util/store/namespace";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const instancesPerPage = 15;

const InstancesListPage = () => {
  const [offset, setOffset] = useState(0);
  const [filters, setFilters] = useState<FiltersObj>({});
  const namespace = useNamespace();
  const { t } = useTranslation();
  const {
    data: instances,
    isSuccess,
    isAllowed,
    noPermissionMessage,
  } = useInstanceList({
    limit: instancesPerPage,
    offset,
  });

  const handleFilterChange = (filters: FiltersObj) => {
    setFilters(filters);
    setOffset(0);
  };

  const numberOfInstances = 0;
  const noResults = isSuccess && instances.length === 0;
  const showPagination = false; // TODO: numberOfInstances > instancesPerPage;
  const hasFilters = false; // TODO: !!Object.keys(filters).length;

  if (!namespace) return null;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Boxes className="h-5" />
        {t("pages.instances.list.title")}
      </h3>
      <Card>
        {/* TODO: */}
        {/* <Filters filters={filters} onUpdate={handleFilterChange} /> */}
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
                  instances?.map((instance) => (
                    <Row
                      instance={instance}
                      key={instance.id}
                      namespace={namespace}
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
