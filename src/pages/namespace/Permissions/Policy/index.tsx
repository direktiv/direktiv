import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { Save } from "lucide-react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const PolicyPage = () => {
  const theme = useTheme();
  const { t } = useTranslation();
  return (
    <div className="flex grow flex-col space-y-4 p-5">
      <Card className="grow p-4" data-testid="revisions-detail-editor">
        <Editor value="" theme={theme ?? undefined} language="plaintext" />
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <Button variant="outline">
          <Save />
          {t("pages.permissions.policy.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default PolicyPage;
