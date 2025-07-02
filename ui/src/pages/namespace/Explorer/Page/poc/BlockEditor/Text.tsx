import { BlockEditFormProps } from ".";
import { Footer } from "./components/Footer";
import { Header } from "./components/Header";
import { TextType } from "../schema/blocks/text";
import { Textarea } from "~/design/TextArea";
import { useState } from "react";

type TextBlockEditFormProps = BlockEditFormProps<TextType>;

export const Text = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: TextBlockEditFormProps) => {
  const [block, setBlock] = useState<TextType>(structuredClone(propBlock));

  return (
    <div className="flex flex-col gap-4">
      <Header action={action} path={path} block={propBlock} />
      <Textarea
        value={block.content}
        onChange={(event) =>
          setBlock({
            ...block,
            content: event.target.value,
          })
        }
      />

      <Footer
        onSubmit={() => {
          onSubmit(block);
        }}
        onCancel={onCancel}
      />
    </div>
  );
};
