import { BlockEditFormProps } from ".";
import { DialogFooter } from "./components/Footer";
import { DialogHeader } from "./components/Header";
import { TextType } from "../schema/blocks/text";
import { Textarea } from "~/design/TextArea";
import { useState } from "react";

/**
 * Please follow this pattern when adding editor components for new block types:
 * Omit the generic blocks type from BlockEditFormProps, and set it to the specific
 * block type this component is intended for.
 */
type TextBlockEditFormProps = Omit<BlockEditFormProps, "block"> & {
  block: TextType;
};

export const Text = ({
  action,
  block: propBlock,
  path,
  onSave,
}: TextBlockEditFormProps) => {
  const [block, setBlock] = useState<TextType>(structuredClone(propBlock));

  return (
    <>
      <DialogHeader action={action} path={path} type="text" />
      <Textarea
        value={block.content}
        onChange={(event) =>
          setBlock({
            ...block,
            content: event.target.value,
          })
        }
      />
      <div>Debug Info {JSON.stringify(block)}</div>

      <DialogFooter onSave={() => onSave(block)} />
    </>
  );
};
