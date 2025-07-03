import { HeadlineType, headlineLevels } from "../schema/blocks/headline";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockEditFormProps } from ".";
import { Footer } from "./components/Footer";
import { Header } from "./components/Header";
import Input from "~/design/Input";
import { useState } from "react";

type HeadlineEditFormProps = BlockEditFormProps<HeadlineType>;

export const Headline = ({
  action,
  block: propBlock,
  path,
  onSubmit,
  onCancel,
}: HeadlineEditFormProps) => {
  const defaultLevel = headlineLevels[1];

  const [block, setBlock] = useState<HeadlineType>(structuredClone(propBlock));

  return (
    <div className="flex flex-col gap-4">
      <Header action={action} path={path} block={propBlock} />
      <Input
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
      <Footer onSubmit={() => onSubmit(block)} onCancel={onCancel} />
    </div>
  );
};
