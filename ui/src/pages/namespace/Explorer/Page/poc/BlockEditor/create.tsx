import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";

import { BlockPathType } from "../PageCompiler/Block";
import Button from "~/design/Button";
import { HeadlineType } from "../schema/blocks/headline";
import Input from "~/design/Input";
import { useBlock } from "../PageCompiler/context/pageCompilerContext";
import { useState } from "react";
import { useTranslation } from "react-i18next";

export type BlockFormProps = { path: BlockPathType };

export type BlockEditFormProps = { block: HeadlineType; path: BlockPathType };

const HeadlineLevelOne: HeadlineType["level"] = "h1";
const HeadlineLevelTwo: HeadlineType["level"] = "h2";
const HeadlineLevelThree: HeadlineType["level"] = "h3";

const HeadlineLevels: HeadlineType["level"][] = [
  HeadlineLevelOne,
  HeadlineLevelTwo,
  HeadlineLevelThree,
];

export const CreateBlockForm = ({
  path,
  setSelectedBlock,
}: {
  path: BlockPathType;
  setSelectedBlock: (block: HeadlineType) => void;
}) => {
  const { t } = useTranslation();
  const block = useBlock(path);

  const [label, setLabel] = useState("example");
  const [level, setLevel] = useState<HeadlineType["level"]>(HeadlineLevelOne);

  if (Array.isArray(block)) {
    throw Error("Can not load list into block editor");
  }

  const findLevel = (value: string) => {
    const levelResult =
      HeadlineLevels.find((level) => String(level) === value) ??
      HeadlineLevelOne;
    setLevel(levelResult);
    return levelResult;
  };

  const createBlock = () => {
    setSelectedBlock({ type: "headline", label, level });
  };

  return (
    <div>
      <DialogHeader>
        <DialogTitle>
          {t("direktivPage.blockEditor.dialogTitle", {
            path: path.join("."),
            action: "create",
            type: "headline",
          })}
        </DialogTitle>
      </DialogHeader>
      <Input
        className="my-4"
        value={label}
        onChange={(e) => setLabel(e.target.value)}
      />
      <Select
        value={level}
        onValueChange={(value) => findLevel(value)}
        defaultValue={HeadlineLevelThree}
      >
        <SelectTrigger variant="outline">
          <SelectValue placeholder="something" />
        </SelectTrigger>
        <SelectContent>
          {HeadlineLevels.map((item) => (
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
          <Button onClick={() => createBlock()} type="submit">
            Save
          </Button>
        </DialogClose>
      </DialogFooter>
    </div>
  );
};
