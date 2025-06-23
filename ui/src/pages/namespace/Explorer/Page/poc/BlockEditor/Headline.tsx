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

type HeadlineEditFormProps = BlockEditFormProps<HeadlineType>;

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
      <DialogHeader action={action} path={path} type={propBlock.type} />
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
