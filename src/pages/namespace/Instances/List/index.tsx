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
import type { FiltersObj } from "../components/Filters";
import NoResult from "./NoResult";
import { Pagination } from "~/componentsNext/Pagination";
import Row from "./Row";
import { useInstances } from "~/api/instances/query/get";
import { useState } from "react";
import { useTranslation } from "react-i18next";

const instancesPerPage = 15;

const InstancesListPage = () => {
  const [offset, setOffset] = useState(0);
  const [filterQuery, setFilterQuery] = useState<string>("");
  const { t } = useTranslation();
  const { data, isFetched } = useInstances({
    limit: instancesPerPage,
    offset,
    filter: filterQuery,
  });

  const updateFilterQuery = (filters: FiltersObj) => {
    let query = "";
    const filterFields = Object.keys(filters) as Array<keyof FiltersObj>;

    filterFields.forEach((field) => {
      const filterItem = filters[field];
      // guard needed because TS thinks filterItem may be undefined
      if (!filterItem) {
        return;
      }

      // FilterFields uses BEFORE and AFTER as distinct types. In the query,
      // the format is filter.field=CREATED&filter.type=BEFORE|AFTER
      const queryField = ["AFTER", "BEFORE"].includes(field)
        ? "CREATED"
        : field;

      query = query.concat(
        `&filter.field=${queryField}&filter.type=${filterItem.type}&filter.val=${filterItem.value}`
      );
    });
    setFilterQuery(query);
  };

  const numberOfInstances = data?.instances?.pageInfo?.total ?? 0;
  const noResults = isFetched && data?.instances.results.length === 0;

  return (
    <div className="flex grow flex-col gap-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Boxes className="h-5" />
        {t("pages.instances.list.title")}
      </h3>
      <Card>
        <Filters onUpdate={updateFilterQuery} />
        <Table className="border-t">
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.instances.list.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell className="w-32">
                {t("pages.instances.list.tableHeader.id")}
              </TableHeaderCell>
              <TableHeaderCell className="w-28">
                {t("pages.instances.list.tableHeader.revisionId")}
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
                <TableCell colSpan={7}>
                  <NoResult />
                </TableCell>
              </TableRow>
            ) : (
              data?.instances.results.map((instance) => (
                <Row
                  instance={instance}
                  key={instance.id}
                  namespace={data.namespace}
                />
              ))
            )}
          </TableBody>
        </Table>
      </Card>
      <Pagination
        itemsPerPage={instancesPerPage}
        offset={offset}
        setOffset={setOffset}
        totalItems={numberOfInstances}
      />
    </div>
  );
};

export default InstancesListPage;
