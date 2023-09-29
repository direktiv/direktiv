import { Bug, Save } from "lucide-react";
import { FC, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import PermissionsHint from "../components/PermissionsHint";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdatePolicy } from "~/api/enterprise/policy/mutate/update";

type PolicyEditorProps = {
  policyFromServer: string;
};

const PolicyEditor: FC<PolicyEditorProps> = ({ policyFromServer }) => {
  const theme = useTheme();
  const { t } = useTranslation();

  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanged, setHasUnsavedChanged] = useState(false);
  const [editorValue, setEditorValue] = useState(policyFromServer);

  const { mutate: updatePolicy, isLoading } = useUpdatePolicy({
    onSuccess: () => {
      setHasUnsavedChanged(false);
    },
    onError: (error) => {
      error && setError(error);
    },
  });

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updatePolicy({
        policyContent: toSave,
      });
    }
  };

  const onEditorValueChange = (newValue: string | undefined) => {
    setHasUnsavedChanged(newValue !== policyFromServer);
    setEditorValue(newValue ?? "");
  };

  return (
    <>
      <Card className="flex grow flex-col p-4">
        <div className="grow">
          <Editor
            value={editorValue}
            onMount={(editor) => {
              editor.focus();
            }}
            theme={theme ?? undefined}
            onChange={onEditorValueChange}
            language="plaintext"
            onSave={onSave}
          />
        </div>
        <div className="flex justify-between gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8">
          {error && (
            <Popover defaultOpen>
              <PopoverTrigger asChild>
                <span className="flex items-center gap-x-1 text-danger-11 dark:text-danger-dark-11">
                  <Bug className="h-5" />
                  {t("pages.permissions.policy.theresOneIssue")}
                </span>
              </PopoverTrigger>
              <PopoverContent asChild>
                <div className="flex p-4">
                  <div className="grow">{error}</div>
                </div>
              </PopoverContent>
            </Popover>
          )}

          {hasUnsavedChanged && (
            <span className="text-center">
              {t("pages.permissions.policy.unsavedNote")}
            </span>
          )}
        </div>
      </Card>
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <PermissionsHint />
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(editorValue);
          }}
        >
          <Save />
          {t("pages.permissions.policy.saveBtn")}
        </Button>
      </div>
    </>
  );
};

export default PolicyEditor;
