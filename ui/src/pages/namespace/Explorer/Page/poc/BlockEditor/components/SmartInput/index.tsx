import { Check, HelpCircleIcon, Maximize2 } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";
import { useRef, useState } from "react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "@tremor/react";
import { FakeInput } from "~/design/FakeInput";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { TreePicker } from "../TreePicker";
import { addSnippetToInputValue } from "../../utils";
import { usePageEditorPanel } from "../../EditorPanelProvider";
import { useTranslation } from "react-i18next";

export const SmartInput = ({
  onUpdate,
  value,
  id,
  placeholder,
}: {
  onUpdate: (value: string) => void;
  value: string;
  id: string;
  placeholder: string;
}) => {
  const { t } = useTranslation();
  const [dialog, setDialog] = useState(false);
  const [dialogContainer, setDialogContainer] = useState<HTMLDivElement | null>(
    null
  );
  const { panel } = usePageEditorPanel();
  const ref = useRef<HTMLInputElement>(null);

  if (!panel) return null;

  const { variables } = panel;

  const variableSegmentPlaceholders = [
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.namespace"),
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.id"),
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.pointer"),
  ];

  return (
    <Dialog open={dialog} onOpenChange={setDialog}>
      <InputWithButton>
        <Input
          ref={ref}
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
            <FakeInput wrap className="flex flex-col gap-2 p-2">
              <div className="border-b pb-2">
                <ButtonBar>
                  <TreePicker
                    label={t("direktivPage.blockEditor.smartInput.variableBtn")}
                    container={dialogContainer ?? undefined}
                    tree={variables}
                    onSubmit={(snippet) =>
                      ref.current &&
                      addSnippetToInputValue({
                        element: ref.current,
                        snippet,
                        value,
                        callback: onUpdate,
                      })
                    }
                    placeholders={variableSegmentPlaceholders}
                    minDepth={3}
                  />
                  <Popover>
                    <PopoverTrigger asChild>
                      <Button variant="outline">
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
                <TreePicker
                  label={t("direktivPage.blockEditor.smartInput.variableBtn")}
                  container={dialogContainer ?? undefined}
                  tree={variables}
                  onSubmit={(snippet) =>
                    ref.current &&
                    addSnippetToInputValue({
                      element: ref.current,
                      snippet,
                      value,
                      callback: onUpdate,
                    })
                  }
                  placeholders={variableSegmentPlaceholders}
                  minDepth={3}
                />
              </div>
              <Input
                ref={ref}
                id={id}
                value={value}
                onChange={(event) => onUpdate(event.target.value)}
                placeholder={placeholder}
              />
            </FakeInput>
            <div className="flex justify-end">
              <Button
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
