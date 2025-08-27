import { Check, SquareArrowOutUpRight } from "lucide-react";
import {
  Command,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
} from "~/design/Command";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { EditorContent, useEditor } from "@tiptap/react";
import { FC, PropsWithChildren, useState } from "react";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { ContextVariables } from "../../../PageCompiler/primitives/Variable/VariableContext";
import Document from "@tiptap/extension-document";
import { InputWithButton } from "~/design/InputWithButton";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

type FakeInputProps = PropsWithChildren & {
  wrap?: boolean;
  sidebar?: boolean;
  className?: string;
};

const FakeInput: FC<FakeInputProps> = ({
  children,
  wrap,
  sidebar,
  className,
  ...props
}) => (
  <div
    className={twMergeClsx(
      // Todo:
      // - consolidate this with the input styling?
      // - focus ring does not work as expected
      sidebar && "min-w-[300px] max-w-[300px]",
      !wrap && "h-9",
      "rounded-md border bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
      "border-gray-4 placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
      "dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4",
      "dark:focus:ring-offset-gray-dark-1",
      className
    )}
    {...props}
  >
    {children}
  </div>
);

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

  type VariableBuilderState = null | {
    namespace: string;
    id?: string;
    idOptions?: string[];
  };

  const [variableBuilder, setVariableBuilder] =
    useState<VariableBuilderState>(null);

  return (
    <Dialog open={dialog} onOpenChange={setDialog}>
      <InputWithButton>
        <FakeInput sidebar>
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
                <Popover>
                  <PopoverTrigger asChild>
                    <Button variant="outline" type="button">
                      {t("direktivPage.blockEditor.smartInput.variableBtn")}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent align="start" container={dialogContainer}>
                    {!variableBuilder?.namespace && (
                      <Command>
                        <CommandInput placeholder="select variable namespace" />
                        <CommandList>
                          <CommandGroup heading="namespace">
                            {Object.entries(variables).map(
                              ([namespace, blockIds]) => (
                                <CommandItem
                                  key={namespace}
                                  onSelect={() =>
                                    setVariableBuilder({
                                      namespace,
                                      idOptions: Object.keys(blockIds),
                                    })
                                  }
                                >
                                  {namespace}
                                </CommandItem>
                              )
                            )}
                          </CommandGroup>
                        </CommandList>
                      </Command>
                    )}
                    {!!variableBuilder?.namespace &&
                      variableBuilder.idOptions && (
                        <Command>
                          <CommandInput placeholder="select block id" />
                          <CommandList>
                            <CommandGroup heading="block scope">
                              {variableBuilder.idOptions.map((id) => (
                                <CommandItem
                                  key={id}
                                  onSelect={() =>
                                    setVariableBuilder({
                                      ...variableBuilder,
                                      id,
                                      idOptions: [],
                                    })
                                  }
                                >
                                  {id}
                                </CommandItem>
                              ))}
                            </CommandGroup>
                          </CommandList>
                        </Command>
                      )}
                    {!!variableBuilder?.namespace && variableBuilder.id && (
                      <div>
                        <div>
                          {`Debug: you selected {{${variableBuilder.namespace}.${variableBuilder.id}}}`}
                        </div>
                        <Button
                          type="button"
                          onClick={() => setVariableBuilder(null)}
                        >
                          Reset
                        </Button>
                      </div>
                    )}
                  </PopoverContent>
                </Popover>
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
