import { EditorContent, useEditor } from "@tiptap/react";

import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Placeholder from "@tiptap/extension-placeholder";
import Text from "@tiptap/extension-text";
import { twMergeClsx } from "~/util/helpers";
import { useTranslation } from "react-i18next";

export const SmartInput = ({
  onChange,
  value,
  id,
}: {
  onChange: (content: string) => void;
  value: string;
  id: string;
}) => {
  const { t } = useTranslation();
  const editor = useEditor({
    extensions: [
      Document,
      Text,
      Paragraph,
      Placeholder.configure({
        placeholder: t(
          "direktivPage.blockEditor.blockForms.text.contentPlaceholder"
        ),
        showOnlyWhenEditable: true,
        showOnlyCurrent: false, // show even when not focused
      }),
    ],
    content: value,
    onUpdate: ({ editor }) => {
      onChange(editor.getText());
    },
  });

  return (
    <div
      className={twMergeClsx(
        "max-w-[300px]",
        "h-9 rounded-md border bg-transparent px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        "border-gray-4 placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
        "dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4 dark:focus:ring-offset-gray-dark-1"
      )}
      onClick={() => editor?.chain().focus().run()}
    >
      <EditorContent
        id={id}
        editor={editor}
        className={twMergeClsx(
          "truncate",
          "min-h-9 max-w-full text-sm [&>*]:outline-none",
          "[&_*.is-empty]:before:absolute",
          "[&_*.is-empty]:before:pointer-events-none",
          "[&_*.is-empty]:before:content-[attr(data-placeholder)]",
          "[&_*.is-empty]:before:text-gray-11"
        )}
      />
    </div>
  );
};
