import { FileSymlink, Users } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { Link } from "react-router-dom";
import { NoPermissions } from "~/design/Table";
import { analyzePath } from "~/util/router/utils";
import { decode } from "js-base64";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";
import { useTranslation } from "react-i18next";

const ApiPathPage: FC = () => {
  const pages = usePages();
  const { path } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: apiPathData,
    isFetched: isPermissionCheckFetched,
  } = useFile({ path });

  if (!namespace) return null;
  if (!path) return null;
  if (apiPathData?.type !== "api_path") return null;
  if (!isPermissionCheckFetched) return null;

  if (isAllowed === false) {
    return (
      <Card className="m-5 flex grow">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );
  }

  const fileContentFromServer = decode(apiPathData.data);

  return (
    <>
      <div className="border-b border-gray-5 bg-gray-1 p-5 dark:border-gray-dark-5 dark:bg-gray-dark-1">
        <div className="flex flex-col gap-5 max-sm:space-y-4 sm:flex-row sm:items-center sm:justify-between">
          <h3 className="flex items-center gap-x-2 font-bold text-primary-500">
            <Users className="h-5" />
            {filename?.relative}
          </h3>
          <Button isAnchor asChild variant="primary">
            <Link
              to={pages.gateway.createHref({
                namespace,
                subpage: "docs",
              })}
            >
              <FileSymlink />
              {t("pages.explorer.docs.goToDocs")}
            </Link>
          </Button>
        </div>
      </div>
      <div className="p-5">
        <Card className="overflow-auto p-4 max-h-[60vh] border border-gray-300 dark:border-gray-600">
          <pre className="whitespace-pre-wrap text-sm text-primary-500">
            {fileContentFromServer}
          </pre>
        </Card>
      </div>
    </>
  );
};

export default ApiPathPage;
