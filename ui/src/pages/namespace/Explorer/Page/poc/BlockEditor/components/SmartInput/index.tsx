import { Check, HelpCircleIcon, Maximize2 } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { EditorContent, useEditor } from "@tiptap/react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { Card } from "@tremor/react";
import Document from "@tiptap/extension-document";
import { FakeInput } from "~/design/FakeInput";
import { InputWithButton } from "~/design/InputWithButton";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { TreePicker } from "../TreePicker";
import { twMergeClsx } from "~/util/helpers";
import { usePageEditorPanel } from "../../EditorPanelProvider";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export const SmartInput = ({
  onChange,
  value,
  id,
  placeholder,
}: {
  onChange: (content: string) => void;
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

  const editor = useEditor({
    extensions: [
      Document,
      Text,
      Paragraph,
      Placeholder.configure({
        placeholder,
      }),
    ],
    content: value,
    onUpdate: ({ editor }) => {
      onChange(editor.getText());
    },
  });

  if (!panel) return null;

  const { variables } = panel;

  const insertText = (text: string) => {
    editor.chain().focus().insertContent(text).run();
  };

  const variableSegmentPlaceholders = [
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.namespace"),
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.id"),
    t("direktivPage.blockEditor.smartInput.templatePlaceholders.pointer"),
  ];

  return (
    <Dialog open={dialog} onOpenChange={setDialog}>
      <InputWithButton>
        <FakeInput narrow>
          {!dialog && (
            <EditorContent
              id={id}
              editor={editor}
              className={twMergeClsx(
                "max-w-full truncate",
                "min-h-9 text-sm [&>*]:outline-none",
                "[&_*.is-empty]:before:absolute",
                "[&_*.is-empty]:before:pointer-events-none",
                "[&_*.is-empty]:before:content-[attr(data-placeholder)]",
                "[&_*.is-empty]:before:text-gray-11"
              )}
            />
          )}
        </FakeInput>
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
                    onSubmit={insertText}
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
              </div>
              <EditorContent
                id={id}
                editor={editor}
                className={twMergeClsx(
                  "max-w-full",
                  "min-h-9 text-sm [&>*]:outline-none",
                  "[&_*.is-empty]:before:absolute",
                  "[&_*.is-empty]:before:pointer-events-none",
                  "[&_*.is-empty]:before:content-[attr(data-placeholder)]",
                  "[&_*.is-empty]:before:text-gray-11",
                  "dark:[&_*.is-empty]:before:text-gray-dark-11"
                )}
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
