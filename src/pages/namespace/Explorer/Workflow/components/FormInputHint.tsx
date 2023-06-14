import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { Info } from "lucide-react";
import { useTranslation } from "react-i18next";

const FormInputHint = () => {
  const { t } = useTranslation();
  return (
    <Card className="flex h-96 w-full flex-col items-center justify-center gap-y-5 p-10 sm:h-[500px]">
      <span className="text-center text-sm">
        {t("pages.explorer.tree.workflow.runWorkflow.formInputHint")}
      </span>
      <Button variant="outline" asChild>
        <a
          href="https://docs.direktiv.io/spec/workflow-yaml/validate/"
          target="_blank"
          rel="noopener noreferrer"
        >
          <Info />
          {t("pages.explorer.tree.workflow.runWorkflow.learnMoreBtn")}
        </a>
      </Button>
    </Card>
  );
};

export default FormInputHint;
