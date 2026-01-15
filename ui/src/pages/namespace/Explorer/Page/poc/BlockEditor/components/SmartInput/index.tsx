import { Check, HelpCircleIcon, Maximize2 } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { Fragment, useCallback, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import {
  VariableNamespace,
  localVariableNamespace,
} from "../../../schema/primitives/variable";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { Preview } from "./Preview";
import { Textarea } from "~/design/TextArea";
import { TreePicker } from "../TreePicker";
import { addSnippetToInputValue } from "./utils";
import { usePageEditorPanel } from "../../EditorPanelProvider";
import { useTranslation } from "react-i18next";

export const SmartInput = ({
  onUpdate,
  value,
  id,
  placeholder,
  blacklist,
}: {
  onUpdate: (value: string) => void;
  value: string;
  id?: string;
  placeholder: string;
  blacklist?: VariableNamespace[];
}) => {
  const { t } = useTranslation();
  const [dialog, setDialog] = useState(false);
  const [dialogContainer, setDialogContainer] = useState<HTMLDivElement | null>(
    null
  );
  const { panel } = usePageEditorPanel();
  const [textarea, setTextarea] = useState<HTMLTextAreaElement | null>();

  const preventSubmit = useCallback((path: string[]) => {
    if (path[0] === localVariableNamespace && path.length > 1) return false;
    if (path.length > 2) return false;
    return true;
  }, []);

  if (!panel) return null;

  const { variables: allVariables } = panel;

  const variables = Object.fromEntries(
    Object.entries(allVariables).filter(
      ([key]) => !(blacklist as string[])?.includes(key)
    )
  );

  return (
    <Dialog open={dialog} onOpenChange={setDialog}>
      <InputWithButton>
        <Input
          className="rounded-none"
          value={value}
          onChange={(event) => onUpdate(event.target.value)}
          placeholder={placeholder}
        />
        <DialogTrigger asChild>
          <Button icon variant="ghost" type="button">
            <Maximize2
              className="text-gray-11 dark:text-gray-dark-11"
              onClick={() => setDialog(true)}
            />
          </Button>
        </DialogTrigger>
      </InputWithButton>

      <DialogContent
        ref={setDialogContainer}
        className="min-w-[600px] max-w-[600px] p-4"
        onInteractOutside={(event) => {
          event.preventDefault();
        }}
      >
        {dialog && (
          <>
            <DialogHeader>
              <DialogTitle>
                {t("direktivPage.blockEditor.smartInput.dialogTitle")}
              </DialogTitle>
            </DialogHeader>
            <div>
              <div className="rounded-t-md border border-b-0 border-gray-4 p-2 dark:border-gray-dark-7">
                <ButtonBar>
                  <TreePicker
                    container={dialogContainer ?? undefined}
                    tree={variables}
                    onSubmit={(snippet) =>
                      textarea &&
                      addSnippetToInputValue({
                        element: textarea,
                        snippet: `{{${snippet}}}`,
                        value,
                        callback: onUpdate,
                      })
                    }
                    preview={(path) => <Preview path={path} />}
                    preventSubmit={preventSubmit}
                  >
                    <Button
                      variant="outline"
                      type="button"
                      className="dark:border-gray-dark-7"
                    >
                      {t("direktivPage.blockEditor.smartInput.variableBtn")}
                    </Button>
                  </TreePicker>
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button
                        variant="outline"
                        className="dark:border-gray-dark-7"
                      >
                        <HelpCircleIcon />
                      </Button>
                    </PopoverTrigger>
                    <PopoverContent>
                      <Card className="text-sm">
                        <p>
                          {t(
                            "direktivPage.blockEditor.smartInput.templateHelp.header"
                          )}
                        </p>
                        <ul className="ml-6 list-disc">
                          <li className="mt-1">
                            {t(
                              "direktivPage.blockEditor.smartInput.templateHelp.namespace"
                            )}
                          </li>
                          <li className="mt-1">
                            {t(
                              "direktivPage.blockEditor.smartInput.templateHelp.id"
                            )}
                          </li>
                          <li className="mt-1">
                            {t(
                              "direktivPage.blockEditor.smartInput.templateHelp.pointer"
                            )}
                          </li>
                        </ul>
                      </Card>
                    </PopoverContent>
                  </Popover>
                </ButtonBar>
              </div>
              <Textarea
                className="h-32 rounded-t-none border-gray-4 dark:border-gray-dark-7"
                ref={setTextarea}
                id={id}
                value={value}
                onChange={(event) => onUpdate(event.target.value)}
                placeholder={placeholder}
              />
            </div>
            <div className="flex justify-end">
              <Button
                className="dark:border-gray-dark-7"
                type="button"
                variant="outline"
                icon
                onClick={() => setDialog(false)}
              >
                <Check />
              </Button>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};
