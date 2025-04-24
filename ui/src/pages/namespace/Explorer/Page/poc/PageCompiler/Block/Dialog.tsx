import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  DialogClose,
  DialogContent,
  Dialog as DialogDesignComponent,
  DialogFooter,
  DialogTrigger,
} from "~/design/Dialog";

import { Block } from ".";
import { BlocksWrapper } from "./utils/BlocksWrapper";
import Button from "~/design/Button";
import { DialogType } from "../../schema/blocks/dialog";

/**
 *
 * TODO:
 * [] add title
 * [] add a concept for a submit and cancel as soon as the form is implemented
 * [] optionally there could always be an X icon at the top right corner
 * [] only render modal children when open
 */

type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPath;
};
export const Dialog = ({
  blockProps: { blocks, trigger: _trigger },
  blockPath,
}: DialogProps) => (
  <DialogDesignComponent>
    <DialogTrigger>Open</DialogTrigger>
    <DialogContent>
      <BlocksWrapper>
        {blocks.map((block, index) => (
          <Block
            key={index}
            block={block}
            blockPath={addSegmentsToPath(blockPath, index)}
          />
        ))}
      </BlocksWrapper>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </DialogDesignComponent>
);
