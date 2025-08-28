import { Check, SquareArrowOutUpRight } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { EditorContent, useEditor } from "@tiptap/react";
import { FC, PropsWithChildren, useState } from "react";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { ContextVariables } from "../../../PageCompiler/primitives/Variable/VariableContext";
import Document from "@tiptap/extension-document";
import { FakeInput } from "~/design/FakeInput";
import { InputWithButton } from "~/design/InputWithButton";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { VariablePicker } from "../VariablePicker";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

const Toolbar: FC<PropsWithChildren> = ({ children }) => (
  <div className="border-b pb-2">{children}</div>
);

export const SmartInput = ({
  onChange,
  value,
  id,
  variables,
}: {
  onChange: (content: string) => void;
  value: string;
  id: string;
  variables: ContextVariables;
}) => {
  const { t } = useTranslation();
  const [dialog, setDialog] = useState(false);
  const [dialogContainer, setDialogContainer] = useState<HTMLDivElement | null>(
    null
  );

  const editor = useEditor({
    extensions: [
      Document,
      Text,
      Paragraph,
      Placeholder.configure({
        placeholder: t(
          "direktivPage.blockEditor.blockForms.text.contentPlaceholder"
        ),
      }),
    ],
    content: value,
    onUpdate: ({ editor }) => {
      onChange(editor.getText());
    },
  });

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
            <SquareArrowOutUpRight
              className="text-gray-11"
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
            <Alert variant="info" className="text-sm">
              {t("direktivPage.blockEditor.smartInput.templateHelp")}
            </Alert>
            <FakeInput wrap className="flex flex-col gap-2 p-2">
              <Toolbar>
                <VariablePicker
                  variables={variables}
                  container={dialogContainer ?? undefined}
                />
              </Toolbar>
              <EditorContent
                id={id}
                editor={editor}
                className={twMergeClsx(
                  "max-w-full",
                  "min-h-9 text-sm [&>*]:outline-none",
                  "[&_*.is-empty]:before:absolute",
                  "[&_*.is-empty]:before:pointer-events-none",
                  "[&_*.is-empty]:before:content-[attr(data-placeholder)]",
                  "[&_*.is-empty]:before:text-gray-11"
                )}
              />
            </FakeInput>
            <div className="flex justify-end">
              <Button
                type="button"
                variant="outline"
                className="h-8"
                icon
                onClick={() => setDialog(false)}
              >
                <Check size="12" />
              </Button>
            </div>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
};
