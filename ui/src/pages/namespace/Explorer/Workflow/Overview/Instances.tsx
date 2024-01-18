import { NoResult, Table, TableBody } from "~/design/Table";

import { Boxes } from "lucide-react";
import { InstanceCard } from "~/pages/namespace/Monitoring/Instances/InstanceCard";
import { InstanceRow } from "~/pages/namespace/Monitoring/Instances/Row";
import RefreshButton from "~/design/RefreshButton";
import { ScrollArea } from "~/design/ScrollArea";
import { forceLeadingSlash } from "~/api/tree/utils";
import { useInstances } from "~/api/instances/query/get";
import { useTranslation } from "react-i18next";

const Instances = ({ workflow }: { workflow: string }) => {
  const { t } = useTranslation();

  const { data, isFetching, refetch } = useInstances({
    limit: 10,
    offset: 0,
    filters: { AS: { type: "WORKFLOW", value: forceLeadingSlash(workflow) } },
  });

  const instancesRefetchButton = (
    <RefreshButton
      icon
      size="sm"
      variant="ghost"
      disabled={isFetching}
      onClick={() => {
        refetch();
      }}
    />
  );

  return (
    <InstanceCard
      headline={t("pages.explorer.tree.workflow.overview.instances.header")}
      icon={Boxes}
      refetchButton={instancesRefetchButton}
    >
      {data?.instances?.results.length === 0 ? (
        <NoResult icon={Boxes}>
          {t("pages.explorer.tree.workflow.overview.instances.noResult")}
        </NoResult>
      ) : (
        <ScrollArea className="h-full">
          <Table>
            <TableBody>
              {data?.instances?.results.map((instance) => (
                <InstanceRow key={instance.id} instance={instance} />
              ))}
            </TableBody>
          </Table>
        </ScrollArea>
      )}
    </InstanceCard>
  );
};

export default Instances;
