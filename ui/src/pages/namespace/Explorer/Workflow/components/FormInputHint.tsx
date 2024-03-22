import Button from "~/design/Button";
import { Info } from "lucide-react";
import { useTranslation } from "react-i18next";

const FormInputHint = () => {
  const { t } = useTranslation();
  return (
    <div className="flex h-full flex-col items-center justify-center gap-y-5 p-10">
      <span
        className="text-center text-sm"
        data-testid="run-workflow-form-input-hint"
      >
        {t("pages.explorer.tree.workflow.runWorkflow.formInputHint")}
      </span>
      <Button variant="outline" asChild isAnchor>
        <a
          href="https://docs.direktiv.io/spec/workflow-yaml/validate/"
          target="_blank"
          rel="noopener noreferrer"
        >
          <Info />
          {t("pages.explorer.tree.workflow.runWorkflow.learnMoreBtn")}
        </a>
      </Button>
    </div>
  );
};

export default FormInputHint;
