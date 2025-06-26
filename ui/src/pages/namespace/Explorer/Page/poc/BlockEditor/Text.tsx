import { BlockEditFormProps } from ".";
import { DialogFooter } from "./components/Footer";
import { DialogHeader } from "./components/Header";
import { TextType } from "../schema/blocks/text";
import { Textarea } from "~/design/TextArea";
import { useState } from "react";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

const formId = "block-editor-text";

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
}: TextBlockEditFormProps) => {
  const [block, setBlock] = useState<TextType>(structuredClone(propBlock));

  return (
    <>
      <DialogHeader action={action} path={path} type={propBlock.type} />
      <Textarea
        value={block.content}
        onChange={(event) =>
          setBlock({
            ...block,
            content: event.target.value,
          })
        }
      />
      <DialogFooter formId={formId} />
    </>
  );
};
