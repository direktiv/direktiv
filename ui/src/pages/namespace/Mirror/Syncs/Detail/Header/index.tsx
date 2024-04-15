import Badge from "~/design/Badge";
import { GitCompare } from "lucide-react";
import { activityStatusToBadgeProps } from "../../utils";
import { useMirrorActivity } from "~/api/tree/query/mirrorInfo";
import { useTranslation } from "react-i18next";
import useUpdatedAt from "~/hooks/useUpdatedAt";

const Header = ({ activityId }: { activityId: string }) => {
  const { data } = useMirrorActivity({ id: activityId });
  const createdAt = useUpdatedAt(data?.createdAt);
  const { t } = useTranslation();

  if (!data) return null;

  const statusBadgeProps = activityStatusToBadgeProps(data.status);

  return (
    <div className="space-y-5 border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
      <div className="flex flex-col gap-x-7 max-md:space-y-4 md:flex-row md:items-center md:justify-start">
        <div className="flex flex-col items-start gap-2">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <GitCompare className="h-5" /> {data.id.slice(0, 8)}
          </h3>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.mirror.syncs.detail.header.status")}
          </div>
          <Badge
            variant={statusBadgeProps.variant}
            icon={statusBadgeProps.icon}
          >
            {data.status}
          </Badge>
        </div>
        <div className="text-sm">
          <div className="text-gray-10 dark:text-gray-dark-10">
            {t("pages.mirror.syncs.detail.header.createdAt")}
          </div>
          {t("pages.mirror.syncs.detail.header.relativeTime", {
            relativeTime: createdAt,
          })}
        </div>
      </div>
    </div>
  );
};

export default Header;
