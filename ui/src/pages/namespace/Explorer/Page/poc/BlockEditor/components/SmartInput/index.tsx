import { EditorContent, useEditor } from "@tiptap/react";

import Document from "@tiptap/extension-document";
import Paragraph from "@tiptap/extension-paragraph";
import Text from "@tiptap/extension-text";
import { inputClasses } from "~/design/Input";

export const SmartInput = ({
  onChange,
  value,
  id,
}: {
  onChange: (content: string) => void;
  value: string;
  id: string;
}) => {
  const editor = useEditor({
    extensions: [Document, Text, Paragraph],
    content: value,
    onUpdate: () => {
      const text = editor.getText();
      onChange(text);
    },
  });

  return (
    <div className={inputClasses({})}>
      <EditorContent id={id} editor={editor} />
    </div>
  );
};
