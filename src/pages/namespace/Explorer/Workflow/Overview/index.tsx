import Badge from "~/design/Badge";
import { Card } from "~/design/Card";
import { FC } from "react";
import { pages } from "~/util/router/pages";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const ActiveWorkflowPage: FC = () => {
  const { t } = useTranslation();
  const { path } = pages.explorer.useParams();
  const { data } = useNodeContent({
    path,
  });

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>{t("pages.explorer.workflow.overview.title")}</h1>
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
    </div>
  );
};

export default ActiveWorkflowPage;
