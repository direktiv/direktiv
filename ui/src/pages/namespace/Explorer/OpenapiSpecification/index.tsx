import { BookOpen, FileSymlink } from "lucide-react";
import { Link, useParams } from "@tanstack/react-router";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { NoPermissions } from "~/design/Table";
import OpenapiSpecificationEditor from "./OpenApiSpecificationEditor";
import { analyzePath } from "~/util/router/utils";
import { useFile } from "~/api/files/query/file";
import { useTranslation } from "react-i18next";

const OpenapiSpecificationPage: FC = () => {
  const { _splat: path } = useParams({ strict: false });
  const { segments } = analyzePath(path);
  const filename = segments[segments.length - 1];
  const { t } = useTranslation();

  const {
    isAllowed,
    noPermissionMessage,
    data: gatewayData,
    isFetched: isPermissionCheckFetched,
  } = useFile({ path });

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
            <BookOpen className="h-5" />
            {filename?.relative}
          </h3>
          <Button isAnchor asChild variant="primary">
            <Link to="/n/$namespace/gateway/gatewayInfo" from="/n/$namespace">
              <FileSymlink />
              {t("pages.explorer.tree.openapiSpecification.link")}
            </Link>
          </Button>
        </div>
      </div>
      <OpenapiSpecificationEditor data={gatewayData} />
    </>
  );
};

export default OpenapiSpecificationPage;
