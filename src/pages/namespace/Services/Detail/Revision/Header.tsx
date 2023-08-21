import { Diamond } from "lucide-react";
import moment from "moment";
import { useNodeContent } from "~/api/tree/query/node";
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
  const { t } = useTranslation();

  const created = useUpdatedAt(
    moment.unix(parseInt(revisionData?.created ?? "0"))
  );

  if (!revisionData) return null;

  const size = revisionData.size;
  const sizeLabel =
    size === 0 || size === 1 || size === 2
      ? t(`pages.services.create.sizeValues.${size}`)
      : ""; // TODO: move into a schema helper (also in src/pages/namespace/Services/List/Row.tsx, and src/pages/namespace/Services/List/Create.tsx)

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex gap-3">
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
          {revisionData.minScale} {/* TODO: show pods with state on hover*/}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.services.revision.detail.replicas")}
          </div>
          {revisionData.actualReplicas}/{revisionData.desiredReplicas}
        </div>
      </div>
    </div>
  );
};

export default Header;
