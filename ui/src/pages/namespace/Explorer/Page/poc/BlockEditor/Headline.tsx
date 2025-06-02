import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { HeadlineType, headlineLevels } from "../schema/blocks/headline";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockEditFormProps } from ".";
import Button from "~/design/Button";
import Input from "~/design/Input";
import { useState } from "react";
import { useTranslation } from "react-i18next";

/**
 * Please follow this pattern when adding editor components for new block types:
 * Omit the generic blocks type from BlockEditFormProps, and set it to the specific
 * block type this component is intended for.
 */
type HeadlineEditFormProps = Omit<BlockEditFormProps, "block"> & {
  block: HeadlineType;
};

export const Headline = ({
  action,
  block: propBlock,
  path,
  onSave,
}: HeadlineEditFormProps) => {
  const { t } = useTranslation();

  const defaultLevel = headlineLevels[1];

  const [block, setBlock] = useState<HeadlineType>(structuredClone(propBlock));

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {t("direktivPage.blockEditor.dialogTitle", {
            path: path.join("."),
            action,
            type: "headline",
          })}
        </DialogTitle>
      </DialogHeader>
      <Input
        className="my-4"
        value={block.label}
        onChange={(e) => setBlock({ ...block, label: e.target.value })}
      />
      <Select
        value={block.level}
        onValueChange={(value: HeadlineType["level"]) =>
          setBlock({
            ...block,
            level: value,
          })
        }
        defaultValue={defaultLevel}
      >
        <SelectTrigger variant="outline">
          <SelectValue placeholder="something" />
        </SelectTrigger>
        <SelectContent>
          {headlineLevels.map((item) => (
            <SelectItem key={item} value={item}>
              <span>{item}</span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>

        <DialogClose asChild>
          <Button onClick={() => onSave(block)}>Save</Button>
        </DialogClose>
      </DialogFooter>
    </>
  );
};
