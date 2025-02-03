import { FileSymlink, FileText } from "lucide-react";

// import BaseFileEditor from "./BaseFileEditor";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
// import ConsumerEditor from "./ConsumerEditor";
import { FC } from "react";
import { Link } from "react-router-dom";
import { NoPermissions } from "~/design/Table";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useNamespace } from "~/util/store/namespace";
import { usePages } from "~/util/router/pages";

// import { useTranslation } from "react-i18next";

const BaseFilePage: FC = () => {
  const pages = usePages();
  const { path } = pages.explorer.useParams();
  const namespace = useNamespace();
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  // const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: gatewayData,
    isFetched: isPermissionCheckFetched,
  } = useFile({ path });

  if (!namespace) return null;
  if (!path) return null;
  if (gatewayData?.type !== "gateway") return null;
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
            <FileText className="h-5" />
            {filename?.relative}
          </h3>
          <Button isAnchor asChild variant="primary">
            <Link
              to={pages.gateway.createHref({
                namespace,
                subpage: "info",
              })}
            >
              <FileSymlink />
              Go to Gateway Info
            </Link>
          </Button>
        </div>
      </div>
      {/* <BaseFileEditor data={gatewayData} /> */}
    </>
  );
};

export default BaseFilePage;
