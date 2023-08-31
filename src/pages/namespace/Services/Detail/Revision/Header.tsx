import {
  HoverCard,
  HoverCardContent,
  HoverCardTrigger,
} from "~/design/HoverCard";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import { ConditionalWrapper } from "~/util/helpers";
import { Diamond } from "lucide-react";
import { SizeSchema } from "~/api/services/schema";
import { StatusBadge } from "../../components/StatusBadge";
import moment from "moment";
import { podStatusToBadgeVariant } from "../../components/utils";
import { usePods } from "~/api/services/query/revision/pods/getAll";
import { useServiceRevision } from "~/api/services/query/revision/getAll";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooksNext/useUpdatedAt";

const Header = ({
  service,
  revision,
}: {
  service: string;
  revision: string;
}) => {
  const { data: revisionData } = useServiceRevision({ service, revision });
  const { data: podsData } = usePods({ service, revision });
  const { t } = useTranslation();

  const created = useUpdatedAt(
    moment.unix(parseInt(revisionData?.created ?? "0"))
  );

  if (!revisionData) return null;

  const sizeParsed = SizeSchema.safeParse(revisionData.size);
  const sizeLabel = sizeParsed.success
    ? t(`pages.services.create.sizeValues.${sizeParsed.data}`)
    : "";

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-3 sm:flex-row">
        <h3 className="flex grow items-center gap-x-2 font-bold text-primary-500">
          <Diamond className="h-5" /> {service} / {revision}
        </h3>
        <div>{revisionData.image}</div>
      </div>
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-between">
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.created")}
          </div>
          {t("pages.services.revision.detail.realtiveTime", {
            relativeTime: created,
          })}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.size")}
          </div>
          {sizeLabel}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.Generation")}
          </div>
          {revisionData.generation}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.scale")}
          </div>
          <ConditionalWrapper
            condition={!!podsData?.pods?.length}
            wrapper={(children) => (
              <HoverCard>
                <HoverCardTrigger asChild>
                  <span className="cursor-pointer underline">{children}</span>
                </HoverCardTrigger>
                <HoverCardContent className="p-0">
                  <Table>
                    <TableBody>
                      {podsData?.pods.map((pod) => (
                        <TableRow key={pod.name}>
                          <TableCell>{pod.name.split("-").at(-1)}</TableCell>
                          <TableCell>
                            <Badge
                              variant={podStatusToBadgeVariant(pod.status)}
                            >
                              {pod.status}
                            </Badge>
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </HoverCardContent>
              </HoverCard>
            )}
          >
            <span>{revisionData.minScale}</span>
          </ConditionalWrapper>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.replicas")}
          </div>
          {revisionData.actualReplicas}/{revisionData.desiredReplicas}
        </div>
      </div>
      <div>
        <div className="flex flex-col gap-3 sm:flex-row">
          {revisionData.conditions.map((condition) => (
            <StatusBadge
              key={condition.name}
              status={condition.status}
              title={condition.reason}
              message={condition.message}
              className="self-start"
            >
              {condition.name}
            </StatusBadge>
          ))}
        </div>
      </div>
    </div>
  );
};

export default Header;
