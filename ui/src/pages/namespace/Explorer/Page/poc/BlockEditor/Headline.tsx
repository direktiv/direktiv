import { HeadlineType, headlineLevels } from "../schema/blocks/headline";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockEditFormProps } from ".";
import { DialogFooter } from "./components/Footer";
import { DialogHeader } from "./components/Header";
import Input from "~/design/Input";
import { useState } from "react";

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
  onSubmit,
}: HeadlineEditFormProps) => {
  const defaultLevel = headlineLevels[1];

  const [block, setBlock] = useState<HeadlineType>(structuredClone(propBlock));

  return (
    <>
      <DialogHeader action={action} path={path} type="headline" />
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
      <DialogFooter onSubmit={() => onSubmit(block)} />
    </>
  );
};
