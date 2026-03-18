import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { FC } from "react";
import { FileSchemaType } from "~/api/files/schema";
import { PolicyEditor } from "./PolicyEditor";
import { Save } from "lucide-react";
import { useTranslation } from "react-i18next";

type PolicyEditorProps = {
  data: FileSchemaType;
};

export const PageLayout: FC<PolicyEditorProps> = ({ data }) => {
  const { t } = useTranslation();

  const initialData: FileSchemaType = {
    ...data,
  };

  return (
    <div className="relative flex h-full min-h-0 flex-col space-y-4 p-5">
      <Card className="relative flex min-h-0 flex-1 flex-col">
        <div className="min-h-0 grow overflow-auto">
          <div className="h-full pt-5">
            <PolicyEditor data={initialData} />
          </div>
        </div>
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button variant="primary" type="button" onClick={() => {}}>
          <Save />
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </div>
    </div>
  );
};
