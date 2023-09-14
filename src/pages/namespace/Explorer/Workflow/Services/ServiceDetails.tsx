import {
  NoResult,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "~/design/Table";

import { Card } from "~/design/Card";
import { Diamond } from "lucide-react";
import Row from "~/pages/namespace/Services/Detail/Row";
import { useServiceDetails } from "~/api/services/query/getDetails";
import { useTranslation } from "react-i18next";

const ServiceDetails = ({
  service,
  workflow,
  version,
}: {
  service: string;
  workflow: string;
  version: string;
}) => {
  const { t } = useTranslation();

  const { data, isSuccess } = useServiceDetails({
    service,
    workflow,
    version,
  });

  if (!data) return null;

  const showTable = (data.revisions.length ?? 0) > 0;
  const noResults = isSuccess && data.revisions.length === 0;

  return (
    <div>
      Revisions List
      <Card>
        <Table>
          <TableHead>
            <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
              <TableHeaderCell>
                {t("pages.services.revision.list.tableHeader.name")}
              </TableHeaderCell>
              <TableHeaderCell className="w-48">
                {t("pages.services.revision.list.tableHeader.image")}
              </TableHeaderCell>
              <TableHeaderCell className="w-16">
                {t("pages.services.revision.list.tableHeader.scale")}
              </TableHeaderCell>
              <TableHeaderCell className="w-20">
                {t("pages.services.revision.list.tableHeader.size")}
              </TableHeaderCell>
              <TableHeaderCell className="w-40">
                {t("pages.services.revision.list.tableHeader.createdAt")}
              </TableHeaderCell>
              <TableHeaderCell className="w-16" />
            </TableRow>
          </TableHead>
          <TableBody>
            {showTable &&
              data?.revisions?.map((revision, index) => (
                <Row
                  key={index}
                  revision={revision}
                  service={service}
                  workflow={workflow}
                  version={version}
                />
              ))}
            {noResults && (
              <TableRow className="hover:bg-inherit dark:hover:bg-inherit">
                <TableCell colSpan={6}>
                  <NoResult icon={Diamond}>
                    {t("pages.services.revision.list.empty.title")}
                  </NoResult>
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default ServiceDetails;
