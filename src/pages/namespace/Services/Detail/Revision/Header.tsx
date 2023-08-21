import { Diamond } from "lucide-react";
import moment from "moment";
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

  const updatedAt = useUpdatedAt(
    moment.unix(parseInt(revisionData?.created ?? "0"))
  );

  if (!revisionData) return null;

  // TODO: update language keys
  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex gap-3">
        <h3 className="flex grow items-center gap-x-2 font-bold text-primary-500">
          <Diamond className="h-5" /> {service} / {revision}
        </h3>
        <div className="text-gray-10 dark:text-gray-dark-10">
          {revisionData.image}
        </div>
      </div>
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-between">
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">Created</div>
          {t("pages.instances.detail.header.realtiveTime", {
            relativeTime: updatedAt,
          })}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">Size</div>
          {revisionData.size}
          {/* TODO: */}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">Generation</div>
          {revisionData.generation}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">Scale</div>
          {revisionData.minScale} {/* TODO: show pods with state on hover*/}
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">Replicas</div>
          {revisionData.actualReplicas}/{revisionData.desiredReplicas}
        </div>
        {/* revisions page: service: {service} - revision: {revision}
        <div>
          <hr />
          {revisionData?.actualReplicas} / {revisionData?.desiredReplicas}
          <hr />
          {revisionData?.image}
          <hr />
          {revisionData?.generation}
        </div> */}
      </div>
    </div>
  );
};

export default Header;
