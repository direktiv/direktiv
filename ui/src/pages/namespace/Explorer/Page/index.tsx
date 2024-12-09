import { Card } from "~/design/Card";
import { FC } from "react";
import { NoPermissions } from "~/design/Table";
import PageEditor from "./PageEditor";
import { PanelTop } from "lucide-react";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const UIPage: FC = () => {
  const pages = usePages();
  const { path } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: endpointData,
    isFetched: isPermissionCheckFetched,
  } = useFile({ path });

  if (!namespace) return null;
  if (!path) return null;
  if (endpointData?.type !== "endpoint") return null;
  if (!isPermissionCheckFetched) return null;

  if (isAllowed === false)
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  return (
    <>
      <div className="border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col gap-5 max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <PanelTop className="h-5" />
            {filename?.relative}
          </h3>
        </div>
      </div>

      <PageEditor data={endpointData} />
    </>
  );
};

export default UIPage;
