import { DialogFooter, DialogHeader, DialogTitle } from "~/design/Dialog";

import { BlockEditFormProps } from ".";
import Button from "~/design/Button";
import { TextType } from "../schema/blocks/text";
import { Textarea } from "~/design/TextArea";
import { useState } from "react";
import { useTranslation } from "react-i18next";

type TextBlockEditFormProps = Omit<BlockEditFormProps, "block"> & {
  block: TextType;
};

export const Text = ({
  block: propBlock,
  path,
  onSave,
}: TextBlockEditFormProps) => {
  const { t } = useTranslation();

  const [block, setBlock] = useState<TextType>(structuredClone(propBlock));

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {t("direktivPage.blockEditor.Text.modalTitle", {
            path: path.join("."),
          })}
        </DialogTitle>
      </DialogHeader>
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

      <DialogFooter>
        <Button variant="primary" onClick={() => onSave(block)}>
          {t("direktivPage.blockEditor.generic.saveButton")}
        </Button>
      </DialogFooter>
    </>
  );
};
