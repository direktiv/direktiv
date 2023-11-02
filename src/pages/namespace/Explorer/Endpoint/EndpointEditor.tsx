import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { Save } from "lucide-react";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const EndpointEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const { t } = useTranslation();

  const isLoading = false;
  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <Card className="flex grow flex-col p-4">test</Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button variant="outline" disabled={isLoading}>
          <Save />
          {t("pages.explorer.gateway.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default EndpointEditor;
